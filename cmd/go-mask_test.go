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

	"github.com/fr12k/go-mask/pkg/cmd"
	"github.com/fr12k/go-mask/pkg/code"
	"github.com/fr12k/go-mask/pkg/config"
	"github.com/fr12k/go-mask/pkg/file"
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
		debugMode     bool
	}{
		{
			name:      "Run success with debug",
			debugMode: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			yamlConfig := fmt.Sprintf(yamlConfig, tt.debugMode, t.TempDir())
			cfgLoader := &config.ConfigLoader{
				File: file.NewFileReader(strings.NewReader(yamlConfig)),
			}
			gomask := NewGoMask()
			gomask.loader = cfgLoader
			err := gomask.Run()
			assert.NoError(t, err)
		})
	}
}

func TestRunErrorLoadConfig(t *testing.T) {
	cfgLoader := &config.ConfigLoader{
		File: file.NewFileReaderError(os.ErrClosed),
	}
	gomask := GoMask{cfgLoader, nil, nil, nil}
	err := gomask.Run()
	assert.Error(t, err)
}

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
			cfgLoader := &config.ConfigLoader{
				File: file.NewFileReader(strings.NewReader(yamlConfig)),
			}
			gomask := GoMask{
				cfgLoader,
				func(cfg *config.Config) *code.CodeReader {
					return code.NewCodeReader(&errorReader{limit: 0})
				},
				nil,
				nil,
			}
			err := gomask.Run()
			assert.Error(t, err)
		})
	}
}

func TestRunWriteCode(t *testing.T) {
	os.Args = []string{"go-mask"}

	yamlConfig := fmt.Sprintf(yamlConfig, false, t.TempDir())
	cfgLoader := &config.ConfigLoader{
		File: file.NewFileReader(strings.NewReader(yamlConfig)),
	}

	var buf bytes.Buffer
	gomask := GoMask{
		loader: cfgLoader,
		reader: func(cfg *config.Config) *code.CodeReader {
			return code.NewCodeReader(strings.NewReader("fmt.Println(\"Hello World\")"))
		},
		writer: func(cfg *config.Config) *file.File {
			return file.NewFileWriterBuffer(&buf, filepath.Join(t.TempDir(), cfg.Command.FileName()))
		},
		command: NewMockCommand(),
	}
	err := gomask.Run()
	assert.NoError(t, err)
	assert.Equal(t, "package main\n\nimport \"fmt\"\nimport \"os\"\n\nfunc main() {\nfmt.Println(\"Hello World\")\n}\n", buf.String())
}

func TestRunWriteCodeError(t *testing.T) {
	os.Args = []string{"go-mask"}

	yamlConfig := fmt.Sprintf(yamlConfig, false, t.TempDir())
	cfgLoader := &config.ConfigLoader{
		File: file.NewFileReader(strings.NewReader(yamlConfig)),
	}

	gomask := GoMask{
		cfgLoader,
		func(cfg *config.Config) *code.CodeReader {
			return code.NewCodeReader(strings.NewReader("fmt.Println(\"Hello World\")"))
		},
		func(cfg *config.Config) *file.File {
			return file.NewFileWriterError(os.ErrClosed)
		},
		nil,
	}
	err := gomask.Run()
	assert.Error(t, err)
	assert.Equal(t, os.ErrClosed, err)
}

func NewMockCommand() *cmd.Command {
	return &cmd.Command{
		CommandInterface: MockCommand{
			command: func(name string, arg ...string) *exec.Cmd {
				return exec.Command("echo")
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
