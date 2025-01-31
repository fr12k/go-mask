package file

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
)

type FileWriter struct {
	Directory string
	FileName  string
	FilePath  string
	io.Writer
}

type File struct {
	FilePath string
	Reader   io.Reader
	Writer   *FileWriter

	reader func() (io.Reader, error)
	writer func() func() (*FileWriter, error)
}

func readerFunc(filePath string) func() (io.Reader, error) {
	return func() (io.Reader, error) {
		file, err := os.Open(filePath)
		return file, err
	}
}

func NewFile(filePath string) *File {
	return &File{
		FilePath: filePath,
		reader:   sync.OnceValues(readerFunc(filePath)),
		writer:   sync.OnceValue(writerFunc(filePath)),
	}
}

func NewFileReader(reader io.Reader) *File {
	load := func() (io.Reader, error) {
		return reader, nil
	}
	return &File{
		reader: sync.OnceValues(load),
	}
}

func NewFileReaderError(err error) *File {
	load := func() (io.Reader, error) {
		return nil, err
	}
	return &File{
		reader: sync.OnceValues(load),
	}
}

func writerFunc(filePath string) func() func() (*FileWriter, error) {
	return func() func() (*FileWriter, error) {
		// Ensure the directory exists
		dir := filepath.Dir(filePath)
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			return func() (*FileWriter, error) {
				return nil, fmt.Errorf("failed to create directory %q: %w", dir, err)
			}
		}
		fileName := filepath.Base(filePath)
		return func() (*FileWriter, error) {
			file, err := os.Create(filePath)
			if err != nil {
				return nil, fmt.Errorf("failed to create file: %w", err)
			}
			return &FileWriter{Directory: dir, FileName: fileName, FilePath: filePath, Writer: file}, nil
		}
	}
}

func NewFileWriter(filePath string) *File {
	return &File{writer: sync.OnceValue(writerFunc(filePath))}
}

func NewFileWriterBuffer(w io.Writer, filePath string) *File {
	writer := func() (*FileWriter, error) {
		dir := filepath.Dir(filePath)
		filename := filepath.Base(filePath)
		return &FileWriter{Writer: w, Directory: dir, FileName: filename, FilePath: filePath}, nil
	}
	return &File{writer: sync.OnceValue(func() func() (*FileWriter, error) {
		return writer
	})}
}

func NewFileWriterError(err error) *File {
	writer := func() (*FileWriter, error) {
		return nil, err
	}
	return &File{writer: sync.OnceValue(func() func() (*FileWriter, error) {
		return writer
	})}
}

func (f *File) Exists() (bool, error) {
	if f.Reader == nil {
		reader, err := f.reader()
		if err != nil {
			if os.IsNotExist(err) {
				f.reader = sync.OnceValues(readerFunc(f.FilePath))
				return false, nil
			}
			return false, err
		}
		f.Reader = reader
	}
	return true, nil
}

func (f *File) Read() ([]byte, error) {
	if f.Reader == nil {
		reader, err := f.reader()
		if err != nil {
			return nil, err
		}
		f.Reader = reader
	}
	return io.ReadAll(f.Reader)
}

// Write implements the io.Writer interface.
func (f *File) Write(p []byte) (n int, err error) {
	if f.Writer == nil {
		fw, err := f.writer()()
		if err != nil {
			return 0, err
		}
		if fw == nil {
			return -1, errors.New("unexpected FileWriter is nil")
		}
		f.Writer = fw
		return fw.Write(p)
	}
	return f.Writer.Write(p)
}

func (f *File) Close() error {
	if f.Reader != nil {
		if closer, ok := f.Reader.(io.Closer); ok {
			return closer.Close()
		}
	}
	if f.Writer != nil {
		if closer, ok := f.Writer.Writer.(io.Closer); ok {
			return closer.Close()
		}
	}
	return nil
}
