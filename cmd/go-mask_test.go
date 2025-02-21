package cmd

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fr12k/go-file"
	"github.com/fr12k/go-mask/pkg/cmd"
	"github.com/fr12k/go-mask/pkg/code"
	"github.com/fr12k/go-mask/pkg/config"

	"github.com/stretchr/testify/assert"
)

const yamlConfig = `
command: build
debug: %t
directory: %s
output: test
imports:
  - "fmt"
  - "os"
mainfunc: true
package: "main"
`

func TestNewGoMask(t *testing.T) {
	gomask := NewGoMask()

	assert.NotNil(t, gomask.loader)
	assert.NotNil(t, gomask.reader)
	assert.NotNil(t, gomask.writer)
	assert.NotNil(t, gomask.command)

	cfg, err := gomask.loader.LoadConfig()
	assert.NoError(t, err)
	reader := gomask.reader(cfg)
	assert.NotNil(t, reader)

	writer := gomask.writer(cfg)
	assert.NotNil(t, writer)
}

func TestRun(t *testing.T) {
	os.Args = []string{"go-mask"}
	tests := []struct {
		name          string
		cfg           *config.Config
		mockError     error
		expectedError bool
	}{
		{
			name: "Run success with debug",
			cfg: &config.Config{
				Command: "build",
				Debug:   true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.cfg.Directory = t.TempDir()
			gomask := NewGoMask(WithConfig(
				tt.cfg,
			))
			_, err := gomask.Run()
			assert.NoError(t, err)
		})
	}
}

func TestToResult(t *testing.T) {
	res := toResult(nil)
	assert.Equal(t, Result{}, res)
}

func TestRunErrorLoadConfig(t *testing.T) {
	os.Args = []string{"go-mask", "--invalid args"}
	gomask := NewGoMask(WithConfig(
		&config.Config{},
	))
	_, err := gomask.Run()
	assert.Error(t, err)
}

func TestRunErrorApplyFlags(t *testing.T) {
	cfgLoader := &config.Loader{
		File: file.NewReaderError(os.ErrClosed),
	}
	gomask := GoMask{cfgLoader, nil, nil, nil}
	_, err := gomask.Run()
	assert.Error(t, err)
}

func TestRunWithCommandError(t *testing.T) {
	os.Args = []string{"go-mask"}
	gomask := NewGoMask(WithConfig(
		&config.Config{
			Command:   "test",
			Directory: t.TempDir(),
			Code:      "fmt.Println(\"Hello World\")",
		},
	))
	gomask.command = NewMockCommandWithError(assert.AnError)
	_, err := gomask.Run()
	assert.ErrorIs(t, err, assert.AnError)
}

func TestRunErrorReadCode(t *testing.T) {
	os.Args = []string{"go-mask", "--debug"}

	tests := []struct {
		name string
	}{
		{
			name: "Run error generate code",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			yamlConfig := fmt.Sprintf(yamlConfig, false, t.TempDir())
			cfgLoader := &config.Loader{
				File: file.NewReader(strings.NewReader(yamlConfig)),
			}
			gomask := GoMask{
				cfgLoader,
				func(_ *config.Config) *code.Reader {
					return code.NewReader(&errorReader{limit: 0})
				},
				nil,
				nil,
			}
			_, err := gomask.Run()
			assert.Error(t, err)
		})
	}
}

func TestRunWriteCode(t *testing.T) {
	os.Args = []string{"go-mask"}

	yamlConfig := fmt.Sprintf(yamlConfig, false, t.TempDir())
	cfgLoader := &config.Loader{
		File: file.NewReader(strings.NewReader(yamlConfig)),
	}

	var buf bytes.Buffer
	gomask := GoMask{
		loader: cfgLoader,
		reader: func(_ *config.Config) *code.Reader {
			return code.NewReader(strings.NewReader("fmt.Println(\"Hello World\")"))
		},
		writer: func(cfg *config.Config) *file.File {
			return file.NewWriterBuffer(&buf, filepath.Join(t.TempDir(), cfg.SaveAs()))
		},
		command: NewMockCommand(),
	}
	_, err := gomask.Run()
	assert.NoError(t, err)
	assert.Equal(t, "package main\n\nimport \"fmt\"\nimport \"os\"\n\nfunc main() {\nfmt.Println(\"Hello World\")\n}\n", buf.String())
}

func TestRunWriteCodeError(t *testing.T) {
	os.Args = []string{"go-mask"}

	yamlConfig := fmt.Sprintf(yamlConfig, false, t.TempDir())
	cfgLoader := &config.Loader{
		File: file.NewReader(strings.NewReader(yamlConfig)),
	}

	gomask := GoMask{
		cfgLoader,
		func(_ *config.Config) *code.Reader {
			return code.NewReader(strings.NewReader("fmt.Println(\"Hello World\")"))
		},
		func(_ *config.Config) *file.File {
			return file.NewWriterError(os.ErrClosed)
		},
		nil,
	}
	_, err := gomask.Run()
	assert.Error(t, err)
	assert.Equal(t, os.ErrClosed, err)
}

// test utilities

type errorReader struct {
	limit int
	count int
}

func (e *errorReader) Read(p []byte) (n int, err error) {
	if e.count >= e.limit {
		return 0, os.ErrClosed
	}
	e.count++
	return copy(p, "fmt.Println(\"Hello World\")\n"), io.EOF
}

func NewMockCommand() *cmd.Command {
	return &cmd.Command{
		CommandInterface: MockCommand{
			command: func(_ string, _ ...string) *exec.Cmd {
				return exec.Command("echo")
			},
		},
	}
}

func NewMockCommandWithError(err error) *cmd.Command {
	return &cmd.Command{
		CommandInterface: MockCommand{
			command: func(_ string, _ ...string) *exec.Cmd {
				cmd := exec.Command("echo")
				cmd.Err = err
				return cmd
			},
		},
	}
}

type MockCommand struct {
	command func(name string, arg ...string) *exec.Cmd
}

func (m MockCommand) Command(name string, arg ...string) *exec.Cmd {
	return m.command(name, arg...)
}
