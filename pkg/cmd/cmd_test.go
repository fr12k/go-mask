package cmd

import (
	"fmt"
	"os/exec"
	"strings"
	"testing"

	"github.com/fr12k/go-mask/pkg/config"
	"github.com/stretchr/testify/assert"
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
			mockError:     nil,
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
			mockError:     nil,
		},
		{
			name: "ExecuteCommand_Run",
			cfg: &config.Config{
				Command: "run",
				Args:    "arg1 arg2",
			},
			tmpfile:       "testfile.go",
			expectedError: false,
			mockError:     nil,
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
			mockError:     fmt.Errorf("command failed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			execCommand := func(name string, args ...string) *exec.Cmd {
				assert.Equal(t, "sh", name)
				if tt.cfg.Command == "test" {
					assert.Equal(t, "-c go test arg1 arg2 testfile.go", strings.Join(args, " "))
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
			err := cmd.ExecuteCommand(tt.cfg, tt.tmpfile)

			// Assert if the error matches the expectation
			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

type MockCommand struct {
	command func(name string, arg ...string) *exec.Cmd
}

func (m MockCommand) Command(name string, arg ...string) *exec.Cmd {
	return m.command(name, arg...)
}

