package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fr12k/go-mask/pkg/cmd"
	"github.com/fr12k/go-mask/pkg/code"
	"github.com/fr12k/go-mask/pkg/config"
	"github.com/fr12k/go-mask/pkg/file"
)

type GoMask struct {
	loader *config.ConfigLoader
	reader func(cfg *config.Config) *code.CodeReader
	writer func(cfg *config.Config) *file.File
	command *cmd.Command
}

func NewGoMask() *GoMask {
	return &GoMask{
		loader: config.NewConfigLoader(".go-mask.yml"),
		reader: func(cfg *config.Config) *code.CodeReader {
			return code.NewCodeReader(strings.NewReader(cfg.Code))
		},
		writer: func(cfg *config.Config) *file.File {
			return file.NewFileWriter(filepath.Join(cfg.Directory, cfg.Command.FileName()))
		},
		command : cmd.NewCommand(),
	}
}

func (g *GoMask) Run() error {
	cfg, err := g.loader.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		return err
	}

	// Parse flags from config
	config.ApplyFlags(cfg)

	// Read the input code
	reader := g.reader(cfg)

	// Generate the Go code
	generatedCode, err := reader.GenerateGoCode(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating Go code: %v\n", err)
		return err
	}

	// Debug mode: print generated code
	if cfg.Debug {
		fmt.Println(generatedCode)
		return nil
	}

	writer := g.writer(cfg)

	_, err = writer.Write([]byte(generatedCode))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing to file: %v\n", err)
		return err
	}

	// Run the build/test command
	return g.command.ExecuteCommand(cfg, writer.Writer.FilePath)
}
