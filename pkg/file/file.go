package file

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
)

type Writer struct {
	Directory string
	FileName  string
	FilePath  string
	io.Writer
}

type File struct {
	FilePath string
	Reader   io.Reader
	Writer   *Writer

	reader func() (io.Reader, error)
	writer func() func() (*Writer, error)
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

func NewReader(reader io.Reader) *File {
	load := func() (io.Reader, error) {
		return reader, nil
	}
	return &File{
		reader: sync.OnceValues(load),
	}
}

func NewReaderError(err error) *File {
	load := func() (io.Reader, error) {
		return nil, err
	}
	return &File{
		reader: sync.OnceValues(load),
	}
}

func writerFunc(filePath string) func() func() (*Writer, error) {
	return func() func() (*Writer, error) {
		// Ensure the directory exists
		dir := filepath.Dir(filePath)
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			return func() (*Writer, error) {
				return nil, fmt.Errorf("failed to create directory %q: %w", dir, err)
			}
		}
		fileName := filepath.Base(filePath)
		return func() (*Writer, error) {
			file, err := os.Create(filePath)
			if err != nil {
				return nil, fmt.Errorf("failed to create file: %w", err)
			}
			return &Writer{Directory: dir, FileName: fileName, FilePath: filePath, Writer: file}, nil
		}
	}
}

func NewWriter(filePath string) *File {
	return &File{writer: sync.OnceValue(writerFunc(filePath))}
}

func NewWriterBuffer(w io.Writer, filePath string) *File {
	//nolint:unparam // the param error is only needed to satisfy the func interface
	writer := func() (*Writer, error) {
		dir := filepath.Dir(filePath)
		filename := filepath.Base(filePath)
		return &Writer{Writer: w, Directory: dir, FileName: filename, FilePath: filePath}, nil
	}
	return &File{writer: sync.OnceValue(func() func() (*Writer, error) {
		return writer
	})}
}

func NewWriterError(err error) *File {
	//nolint:unparam // the param *Writer is only needed to satisfy the func interface
	writer := func() (*Writer, error) {
		return nil, err
	}
	return &File{writer: sync.OnceValue(func() func() (*Writer, error) {
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
			return -1, errors.New("unexpected Writer is nil")
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
