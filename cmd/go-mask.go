package cmd

import (
	// "bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fr12k/go-mask/pkg/cmd"
	"github.com/fr12k/go-mask/pkg/code"
	"github.com/fr12k/go-mask/pkg/config"
	"github.com/fr12k/go-mask/pkg/file"
	"gopkg.in/yaml.v3"
)

type (
	GoMask struct {
		loader  *config.ConfigLoader
		reader  func(cfg *config.Config) *code.CodeReader
		writer  func(cfg *config.Config) *file.File
		command *cmd.Command
	}

	Option func(*GoMask)

	Result struct {
		Stdout string
		Stderr string
	}
)

func WithConfig(cfg *config.Config) Option {
	return func(g *GoMask) {
		//ignore error because we are sure that the config is valid
		b, _ := yaml.Marshal(cfg)
		g.loader = config.NewConfigLoaderBuffer(string(b))
	}
}

func NewGoMask(opts ...Option) *GoMask {
	goMask := &GoMask{
		loader: config.NewConfigLoader(".go-mask.yml"),
		reader: func(cfg *config.Config) *code.CodeReader {
			return code.NewCodeReader(strings.NewReader(cfg.Code))
		},
		writer: func(cfg *config.Config) *file.File {
			return file.NewFileWriter(filepath.Join(cfg.Directory, cfg.SaveAs()))
		},
		command: cmd.NewCommand(),
	}

	for _, opt := range opts {
		opt(goMask)
	}

	return goMask
}

func (g *GoMask) Run() (*Result, error) {
	cfg, err := g.loader.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		return nil, err
	}

	// Parse flags from config
	config.ApplyFlags(cfg)

	// Read the input code
	reader := g.reader(cfg)

	// Generate the Go code
	generatedCode, err := reader.GenerateGoCode(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating Go code: %v\n", err)
		return nil, err
	}

	// Debug mode: print generated code
	if cfg.Debug {
		return &Result{
			Stdout: generatedCode,
		}, nil
	}

	// Write the generated code to a file
	writer := g.writer(cfg)

	_, err = writer.Write([]byte(generatedCode))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing to file: %v\n", err)
		return nil, err
	}

	// Run the build/run/test command
	res, err := g.command.ExecuteCommand(cfg, writer.Writer.FilePath)
	if err != nil {
		return toResult(res), err
	}
	return toResult(res), nil
}

func toResult(res *cmd.CommandResult) *Result {
	return &Result{
		Stdout: res.Stdout,
		Stderr: res.Stderr,
	}
}
