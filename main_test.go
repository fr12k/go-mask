//go:build testrunmain

package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
)

var appName = "go-mask2"

func TestMain(m *testing.M) {
	fmt.Println("-> Building...")

	build := exec.Command("go", "build", "-cover", "-o", appName)
	build.Stdout = os.Stdout
	build.Stderr = os.Stderr
	if err := build.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error building %s: %s", appName, err)
		os.Exit(1)
	}
	fmt.Println("-> Running...")
	err := os.MkdirAll(".coverdata", 0o600)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating directory: %s", err)
		os.Exit(1)
	}
	// Running the
	result := m.Run()
	fmt.Println("-> Getting coverage...")
	cmd := exec.Command("go", "tool", "covdata", "textfmt", "-i=.coverdata/,./.coverdata/unit", "-o", "coverage.txt")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error running command: %s", err)
		os.Exit(1)
	}

	os.RemoveAll(".coverdata")
	os.Exit(result)
}

func TestCallMain(t *testing.T) {
	var buf bytes.Buffer
	cmd := exec.Command("./"+appName, "--debug", "-c", "fmt.Println(\"Hello, World!\")")
	cmd.Env = append(os.Environ(), "GOCOVERDIR=.coverdata")
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	err := cmd.Run()
	assert.NoError(t, err)
	assert.Equal(t, "fmt.Println(\"Hello, World!\")\n", buf.String())
}

func TestCallMainError(t *testing.T) {
	var buf bytes.Buffer
	cmd := exec.Command("./"+appName, "-c", "fmt.Println(\"Hello, World!\")")
	cmd.Env = append(os.Environ(), "GOCOVERDIR=.coverdata")
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	err := cmd.Run()
	assert.Error(t, err)
	assert.Equal(t, "Error executing command: exit status 1\nFAIL\tcommand-line-arguments [setup failed]\nFAIL\n# command-line-arguments\n.go-mask/go-mask_test.go:1:1: expected 'package', found fmt\nexit status 1\n", buf.String())
}
