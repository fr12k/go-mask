package code

import (
	"os"
	"strings"
	"testing"
	"testing/iotest"

	"github.com/fr12k/go-mask/pkg/config"
	"github.com/stretchr/testify/assert"
)

func TestReadCode_FlagCodeProvided(t *testing.T) {
	// Simulate flag input
	c :=  "package main\nfunc main() {}"
	reader := NewCodeReader(strings.NewReader(c))
	code, err := reader.ReadCode()
	assert.NoError(t, err, "No error should be returned when reading code from flag")
	assert.Equal(t, "package main\nfunc main() {}", code, "Code should match the flag input")
}

func TestReadCode_ReadFromStdin(t *testing.T) {
	// Simulate stdin input
	mockInput := "package main\nfunc main() {}\n"
	stdin := os.Stdin
	defer func() { os.Stdin = stdin }()
	r, w, _ := os.Pipe()
	os.Stdin = r

	_, err := w.WriteString(mockInput)
	assert.NoError(t, err)
	w.Close()

	// Read from stdin
	reader := NewCodeReader(nil)
	code, err := reader.ReadCode()
	assert.NoError(t, err, "No error should be returned when reading code from stdin")
	assert.Equal(t, mockInput, code, "Code should match the stdin input")
}

func TestReadCode_ReadFromStdinWithError(t *testing.T) {
	// Simulate an error in stdin
	stdin := os.Stdin
	defer func() { os.Stdin = stdin }()
	r, _, _ := os.Pipe()
	os.Stdin = r
	r.Close() // Immediately close the read pipe to cause an error

	reader := NewCodeReader(nil)
	_, err := reader.ReadCode()
	assert.Error(t, err, "An error should be returned when reading from stdin")
}

func TestGenerateGoCode(t *testing.T) {
	tests := []struct {
		name     string
		cfg      *config.Config
		code    string
		expected string
	}{
		{
			name: "GenerateWithPackageAndImports",
			cfg: &config.Config{
				Package:  "mypackage",
				Imports:  []string{"fmt", "os"},
				MainFunc: false,
			},
			code: `fmt.Println("Hello, World!")`,
			expected: `package mypackage

import "fmt"
import "os"

fmt.Println("Hello, World!")
`,
		},
		{
			name: "GenerateWithMainFunction",
			cfg: &config.Config{
				Package:  "mypackage",
				Imports:  []string{"fmt"},
				MainFunc: true,
			},
			code: `fmt.Println("Hello, World!")`,
			expected: `package mypackage

import "fmt"

func main() {
fmt.Println("Hello, World!")
}
`,
		},
		{
			name: "GenerateWithoutPackageAndImports",
			cfg: &config.Config{
				Package:  "",
				Imports:  []string{},
				MainFunc: false,
			},
			code: `fmt.Println("Hello, World!")`,
			expected: `fmt.Println("Hello, World!")
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := NewCodeReader(strings.NewReader(tt.code))
			output, _ := reader.GenerateGoCode(tt.cfg)
			assert.Equal(t, tt.expected, output, "Generated code should match the expected output")
		})
	}
}

func TestReadCodeErrorReadingCode(t *testing.T) {
	// Simulate an error while reading code
	reader := NewCodeReader(iotest.ErrReader(os.ErrClosed))
	output, err := reader.ReadCode()
	assert.Error(t, err, "An error should be returned when reading code")
	assert.Empty(t, output, "Generated code should be empty when an error occurs")
}

func TestGenerateGoCodeErrorReadingCode(t *testing.T) {
	// Simulate an error while reading code
	reader := NewCodeReader(iotest.ErrReader(os.ErrClosed))
	output, err := reader.GenerateGoCode(&config.Config{})
	assert.Error(t, err, "An error should be returned when reading code")
	assert.Empty(t, output, "Generated code should be empty when an error occurs")
}
