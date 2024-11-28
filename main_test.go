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
	err := os.MkdirAll(".coverdata", 0755)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating directory: %s", err)
		os.Exit(1)
	}
	// Running the
	result := m.Run()
	fmt.Println("-> Getting done...")
	//go tool covdata textfmt -i=coverdata/ -o system.out
	// os.Remove(appName)

	cmd := exec.Command("go", "tool", "covdata", "textfmt", "-i=.coverdata/", "-o", "system.out")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error running command: %s", err)
		os.Exit(1)
	}

	var buf bytes.Buffer
	//gocovmerge  system.out coverage.out > merged.out
	cmd = exec.Command("gocovmerge", "system.out", "coverage.txt")
	cmd.Stdout = &buf
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error running command: %s", err)
		os.Exit(1)
	}
	err = os.WriteFile("coverage.txt", buf.Bytes(), 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing file: %s", err)
		os.Exit(1)
	}
	os.Remove(appName)
	os.Remove("system.out")
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
	assert.Equal(t, "Code: fmt.Println(\"Hello, World!\")\nfmt.Println(\"Hello, World!\")\n\n", buf.String())
}

func TestCallMainError(t *testing.T) {
	var buf bytes.Buffer
	cmd := exec.Command("./"+appName, "-c", "fmt.Println(\"Hello, World!\")")
	cmd.Env = append(os.Environ(), "GOCOVERDIR=.coverdata")
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	err := cmd.Run()
	assert.Error(t, err)
	assert.Equal(t, "Code: fmt.Println(\"Hello, World!\")\n.go-mask/go-mask_test.go:1:1: expected 'package', found fmt\nError executing command: exit status 1\nexit status 1\n", buf.String())
}
