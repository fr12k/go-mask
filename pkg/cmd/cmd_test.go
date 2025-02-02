package cmd

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fr12k/go-mask/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCommand(t *testing.T) {
	cmd := NewCommand()
	assert.NotNil(t, cmd)

	c := cmd.Command("echo")
	assert.NotNil(t, c)
}

func TestExecuteCommand(t *testing.T) {
	tests := []struct {
		name          string
		cfg           *config.Config
		tmpfile       string
		expectedError bool
		mockError     error
	}{
		{
			name: "ExecuteCommand_Test",
			cfg: &config.Config{
				Command: "test",
				Args:    "arg1 arg2",
			},
			tmpfile:       "testfile.go",
			expectedError: false,
		},
		{
			name: "ExecuteCommand_Build",
			cfg: &config.Config{
				Command: "build",
				Args:    "arg1 arg2",
				Output:  "outputfile",
			},
			tmpfile:       "testfile.go",
			expectedError: false,
		},
		{
			name: "ExecuteCommand_Run",
			cfg: &config.Config{
				Command: "run",
				Args:    "arg1 arg2",
			},
			tmpfile:       "testfile.go",
			expectedError: false,
		},
		{
			name: "ExecuteCommand_Error",
			cfg: &config.Config{
				Command: "error",
				Args:    "arg1 arg2",
				Output:  "outputfile",
			},
			tmpfile:       "testfile.go",
			expectedError: true,
			// mockError:     fmt.Errorf("command failed"),
		},
		{
			name: "ExecuteCommand_Error",
			cfg: &config.Config{
				Command: "test",
				Args:    "arg1 arg2",
				Output:  "outputfile",
			},
			tmpfile:       "invalidDir/testfile.go",
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			execCommand := func(name string, args ...string) *exec.Cmd {
				assert.Equal(t, "sh", name)
				if tt.cfg.Command == "test" {
					assert.Equal(t, "-c go test arg1 arg2 testfile.go cmd.go", strings.Join(args, " "))
				}
				if tt.cfg.Command == "build" {
					assert.Equal(t, "-c go build arg1 arg2 -o outputfile testfile.go", strings.Join(args, " "))
				}
				if tt.cfg.Command == "run" {
					assert.Equal(t, "-c go run arg1 arg2 testfile.go", strings.Join(args, " "))
				}
				if tt.cfg.Command == "error" {
					return exec.Command("")
				}
				return exec.Command("echo")
			}
			// Execute the command
			cmd := Command{
				MockCommand{
					command: execCommand,
				},
			}
			_, err := cmd.ExecuteCommand(tt.cfg, tt.tmpfile)

			// Assert if the error matches the expectation
			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestListFilesWithSuffix(t *testing.T) {
	tests := []struct {
		name           string
		suffix         string
		excludeSuffix  string
		expectedFiles  []string
		setup          func(t *testing.T) (string, func())
		expectedErrMsg string
	}{
		{
			name:          "No matching files",
			suffix:        ".txt",
			excludeSuffix: ".log",
			expectedFiles: []string{},
			setup: func(t *testing.T) (string, func()) {
				t.Helper()
				tempDir := t.TempDir()
				return tempDir, func() { os.RemoveAll(tempDir) }
			},
		},
		{
			name:          "Matching files found",
			suffix:        ".txt",
			excludeSuffix: ".log",
			expectedFiles: []string{"file1.txt", "file2.txt"},
			setup: func(t *testing.T) (string, func()) {
				t.Helper()
				// Create a temporary directory with files
				tempDir := t.TempDir()
				WriteTestFile(t, tempDir, "file1.txt", "content")
				WriteTestFile(t, tempDir, "file2.txt", "content")
				WriteTestFile(t, tempDir, "file3.log", "content")
				return tempDir, func() { os.RemoveAll(tempDir) }
			},
		},
		{
			name:          "Error during filepath.Walk",
			suffix:        ".txt",
			excludeSuffix: ".log",
			expectedFiles: nil,
			setup: func(_ *testing.T) (string, func()) {
				return "invalidDir", func() {}
			},
			expectedErrMsg: "no such file or directory",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Set up temporary test data
			dir, cleanup := tc.setup(t)
			defer cleanup()

			// Call the function
			files, err := listFilesWithSuffix(dir, tc.suffix, tc.excludeSuffix)

			_files := []string{}
			for _, file := range files {
				_files = append(_files, filepath.Base(file))
			}
			// Verify the results
			assert.ElementsMatch(t, tc.expectedFiles, _files)
			if tc.expectedErrMsg != "" {
				assert.ErrorContains(t, err, tc.expectedErrMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// test utility

type MockCommand struct {
	command func(name string, arg ...string) *exec.Cmd
}

func (m MockCommand) Command(name string, arg ...string) *exec.Cmd {
	return m.command(name, arg...)
}

func WriteTestFile(t *testing.T, dir, file, content string) {
	t.Helper()
	err := os.WriteFile(filepath.Join(dir, file), []byte(content), 0o600)
	require.NoError(t, err)
}
