package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/fr12k/go-mask/pkg/config"
)

// var execCommand = exec.Command

type CommandInterface interface {
	Command(name string, arg ...string) *exec.Cmd
}

type ExecCommand struct {}

func (m ExecCommand) Command(name string, arg ...string) *exec.Cmd {
	return exec.Command(name, arg...)
}

type Command struct {
	CommandInterface
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

func (c *Command) ExecuteCommand(cfg *config.Config, tmpfile string) error {
	args := []string{"go", cfg.Command.Name()}
	if cfg.Args != "" {
		args = append(args, strings.Fields(cfg.Args)...)
	}

	switch cfg.Command {
	case "test":
		args = append(args, tmpfile)
	case "build":
		args = append(args, "-o", cfg.Output, tmpfile)
	case "run":
		args = append(args, tmpfile)
	}

	cmdStr := strings.Join(args, " ")
	cmd := c.Command("sh", "-c", cmdStr)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error executing command: %v\n", err)
		return err
	}
	return nil
}
