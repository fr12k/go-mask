package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fr12k/go-file"
	"github.com/fr12k/go-mask/pkg/cmd"
	"github.com/fr12k/go-mask/pkg/code"
	"github.com/fr12k/go-mask/pkg/config"

	"gopkg.in/yaml.v3"
)

type (
	GoMask struct {
		loader  *config.Loader
		reader  func(cfg *config.Config) *code.Reader
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
		//nolint:errcheck // ignore error because we are sure that the config is valid
		b, _ := yaml.Marshal(cfg)
		g.loader = config.NewLoaderBuffer(string(b))
	}
}

func NewGoMask(opts ...Option) *GoMask {
	goMask := &GoMask{
		loader: config.NewLoader(".go-mask.yml"),
		reader: func(cfg *config.Config) *code.Reader {
			return code.NewReader(strings.NewReader(cfg.Code))
		},
		writer: func(cfg *config.Config) *file.File {
			return file.NewWriter(filepath.Join(cfg.Directory, cfg.SaveAs()))
		},
		command: cmd.NewCommand(),
	}

	for _, opt := range opts {
		opt(goMask)
	}

	return goMask
}

func (g *GoMask) Run() (Result, error) {
	cfg, err := g.loader.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		return Result{}, err
	}

	// Parse flags from config
	err = config.ApplyFlags(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating parsing flags from commandline: %v\n", err)
		return Result{}, err
	}

	// Read the input code
	reader := g.reader(cfg)

	// Generate the Go code
	generatedCode, err := reader.GenerateGoCode(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating Go code: %v\n", err)
		return Result{}, err
	}

	// Debug mode: print generated code
	if cfg.Debug {
		return Result{
			Stdout: generatedCode,
		}, nil
	}

	// Write the generated code to a file
	writer := g.writer(cfg)

	_, err = writer.Write([]byte(generatedCode))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing to file: %v\n", err)
		return Result{}, err
	}

	// Run the build/run/test command
	res, err := g.command.ExecuteCommand(cfg, writer.Writer.FilePath)
	if err != nil {
		return toResult(res), err
	}
	return toResult(res), nil
}

func toResult(res *cmd.CommandResult) Result {
	if res == nil {
		return Result{}
	}
	return Result{
		Stdout: res.Stdout,
		Stderr: res.Stderr,
	}
}
