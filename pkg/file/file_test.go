package file

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestReadFileSuccess verifies that the file content is correctly read when the file exists.
func TestReadFileSuccess(t *testing.T) {
	// Create a File instance
	file := NewFileReader(io.NopCloser(strings.NewReader("Hello, World!")))

	// Read the file content
	content, err := file.Read()
	require.NoError(t, err)
	assert.Equal(t, "Hello, World!", string(content))

	// Close the file
	err = file.Close()
	require.NoError(t, err)
}

// TestReadFileNotExist verifies that an error is returned when the file does not exist.
func TestReadFileNotExist(t *testing.T) {
	// Create a File instance with a non-existent file
	file := NewFile("nonexistent.txt")

	// Try to read the file
	_, err := file.Read()
	assert.Error(t, err)
	assert.True(t, os.IsNotExist(err))
	err = file.Close()
	require.NoError(t, err)
}

// TestLazyLoadErrorOnLoad simulates an error during the lazy load process.
func TestLazyLoadErrorOnLoad(t *testing.T) {
	// Create a File instance with a custom loader that fails
	file := &File{
		FilePath: "fakefile",
		reader: func() (io.Reader, error) {
			return nil, io.EOF // Simulate a load error
		},
	}

	// Attempt to read, expecting an error
	_, err := file.Read()
	require.Error(t, err)
	assert.Equal(t, io.EOF, err)
}

// TestCustomReader verifies that a custom reader can be used to test without file system access.
func TestCustomReader(t *testing.T) {
	// Use a mock reader to test without accessing the file system
	mockReader := NewMockReader("Mock Data")

	file := &File{
		FilePath: "mockpath",
		Reader:   mockReader,
	}

	content, err := file.Read()
	require.NoError(t, err)
	assert.Equal(t, "Mock Data", string(content))
}

func TestFileExist(t *testing.T) {
	tmpFile, closeFnc := createFile(t, "Hello, World!")
	defer closeFnc()
	// Create a File instance with a non-existent file
	file := NewFile(tmpFile)

	// Try to read the file
	exists, err := file.Exists()
	require.NoError(t, err)
	assert.True(t, exists)
}

func TestFileExistFalse(t *testing.T) {
	// Create a File instance with a non-existent file
	file := NewFile("nonexistent.txt")

	// Try to read the file
	exists, err := file.Exists()
	require.NoError(t, err)
	assert.False(t, exists)
}

func TestFileExistError(t *testing.T) {
	// Create a File instance with a non-existent file
	file := NewFileReaderError(os.ErrClosed)

	// Try to read the file
	exists, err := file.Exists()
	assert.Error(t, err)
	assert.False(t, exists)
}

func TestNewFile(t *testing.T) {
	// Define a test file path
	filePath := "./testFile.txt"

	// Clean up the file after the test
	defer os.Remove(filePath)

	// Create a new file
	file := NewFile(filePath)
	assert.NotNil(t, file, "Expected a non-nil file")

	// Test that the file path matches the expected one
	assert.Equal(t, filePath, file.FilePath, "Expected file path to match the input path")
}

func TestNewFileWriter(t *testing.T) {
	// Test directory structure
	baseDir := "testdata"
	testFilePath := filepath.Join(baseDir, "logs", "output.log")

	t.Run("CreatesFileWriterWhenDirectoryDoesNotExist", func(t *testing.T) {
		// Clean up after the tests
		defer os.RemoveAll(baseDir)
		file := NewFileWriter(testFilePath)
		writer, err := file.writer()()
		require.NoError(t, err)

		assert.Equal(t, filepath.Dir(testFilePath), writer.Directory)
		assert.Equal(t, filepath.Base(testFilePath), writer.FileName)

		_, err = file.Write([]byte("Hello, World!"))
		require.NoError(t, err)

		_, err = file.Write([]byte("Hello, World!"))
		require.NoError(t, err)

		// Verify the directory was created
		_, err = os.Stat(writer.Directory)
		require.NoError(t, err)

		// Verify the file was created
		_, err = os.Stat(filepath.Join(writer.Directory, writer.FileName))
		require.NoError(t, err)

		// Test Close
		err = file.Close()
		require.NoError(t, err)
	})

	t.Run("CreatesFileWriterWhenDirectoryExists", func(t *testing.T) {
		// Clean up after the tests
		defer os.RemoveAll(baseDir)
		// Ensure the directory exists
		err := os.MkdirAll(filepath.Dir(testFilePath), os.ModePerm)
		require.NoError(t, err)

		file := NewFileWriter(testFilePath)
		writer, err := file.writer()()
		require.NoError(t, err)

		assert.Equal(t, filepath.Dir(testFilePath), writer.Directory)
		assert.Equal(t, filepath.Base(testFilePath), writer.FileName)

		// Verify the file was created
		_, err = os.Stat(filepath.Join(writer.Directory, writer.FileName))
		require.NoError(t, err)
	})

	t.Run("FailsToCreateDirectory", func(t *testing.T) {
		// Clean up after the tests
		defer os.RemoveAll(baseDir)
		// Create a file at the directory path to cause MkdirAll to fail
		err := os.MkdirAll(baseDir, os.ModePerm)
		require.NoError(t, err)
		dir := filepath.Join(baseDir, "logs")
		err = os.WriteFile(dir, []byte{}, os.ModePerm) // Create a file where the directory should be
		require.NoError(t, err)
		defer os.Remove(dir)

		file := NewFileWriter(dir + "/")
		_, err = file.writer()()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create directory")
	})

	t.Run("FailsToCreateFile", func(t *testing.T) {
		// Create a temporary directory
		baseDir, err := os.MkdirTemp("", "readonly-test")
		assert.NoError(t, err)
		defer os.RemoveAll(baseDir)

		testFilePath := filepath.Join(baseDir, "output.log")

		file := NewFileWriter(testFilePath)
		fnc := file.writer()

		os.RemoveAll(baseDir)
		_, err = fnc()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create file")

		_, err = file.Write([]byte("Hello, World!"))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create file")
	})
}

func TestNewFileWriterBuffer(t *testing.T) {
	// Test directory structure
	baseDir := "."
	testFilePath := filepath.Join(baseDir, "output.log")

	// Clean up after the tests
	defer os.RemoveAll(baseDir)
	var buf bytes.Buffer
	file := NewFileWriterBuffer(&buf, testFilePath)
	writer, err := file.writer()()
	require.NotNil(t, writer)
	require.NoError(t, err)

	assert.Equal(t, filepath.Dir(testFilePath), writer.Directory)
	assert.Equal(t, filepath.Base(testFilePath), writer.FileName)
}

func TestNewFileWriterError(t *testing.T) {
	file := NewFileWriterError(os.ErrClosed)
	_, err := file.writer()()
	assert.Error(t, err)
}

// Test Utility

type MockReader struct {
	reader func(p []byte) (int, error)
}

func (m *MockReader) Read(p []byte) (int, error) {
	return m.reader(p)
}

func NewMockReader(data string) *MockReader {
	return &MockReader{reader: strings.NewReader(data).Read}
}

func createFile(t *testing.T, cnt string) (string, func()) {
	// Create a temporary file with test content
	tmpFile, err := os.CreateTemp("", "testfile")
	require.NoError(t, err)

	_, err = tmpFile.WriteString(cnt)
	require.NoError(t, err)

	require.NoError(t, tmpFile.Close())
	return tmpFile.Name(), func() {
		os.Remove(tmpFile.Name())
	}
}
