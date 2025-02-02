package cmd

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/fr12k/go-mask/pkg/config"
)

type (
	CommandInterface interface {
		Command(name string, arg ...string) *exec.Cmd
	}

	ExecCommand struct{}

	Command struct {
		CommandInterface
	}

	CommandResult struct {
		Stdout string
		Stderr string
	}
)

func (m ExecCommand) Command(name string, arg ...string) *exec.Cmd {
	return exec.Command(name, arg...)
}

func NewCommand() *Command {
	cmd := &Command{
		CommandInterface: ExecCommand{},
	}
	return cmd
}

func (c Command) Command(name string, arg ...string) *exec.Cmd {
	return c.CommandInterface.Command(name, arg...)
}

func (c *Command) ExecuteCommand(cfg *config.Config, tmpfile string) (*CommandResult, error) {
	args := []string{"go", cfg.Command.Name()}
	if cfg.Args != "" {
		args = append(args, strings.Fields(cfg.Args)...)
	}

	switch cfg.Command {
	case "test":
		dir := filepath.Dir(tmpfile)
		files, err := listFilesWithSuffix(dir, ".go", "test.go")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error listing files: %v\n", err)
			return nil, err
		}
		args = append(args, tmpfile)
		args = append(args, files...)
	case "build":
		args = append(args, "-o", cfg.Output, tmpfile)
	case "run":
		args = append(args, tmpfile)
	}

	cmdStr := strings.Join(args, " ")
	cmd := c.Command("sh", "-c", cmdStr)

	stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}

	cmd.Stdout = stdout
	cmd.Stderr = stderr

	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error executing command: %v\n", err)
		return &CommandResult{
			Stdout: stdout.String(),
			Stderr: stderr.String(),
		}, err
	}

	return &CommandResult{
		Stdout: stdout.String(),
		Stderr: stderr.String(),
	}, nil
}

func listFilesWithSuffix(dir, suffix, excludeSuffix string) ([]string, error) {
	var files []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Check if it's a file and ends with the desired suffix
		if !info.IsDir() &&
			strings.HasSuffix(info.Name(), suffix) &&
			!strings.HasSuffix(info.Name(), excludeSuffix) {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}
