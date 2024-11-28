package config

import (
	"os"
	"strings"
	"testing"
	"testing/iotest"

	"github.com/fr12k/go-mask/pkg/file"
	"github.com/stretchr/testify/assert"
)

func TestName(t *testing.T) {
	cmd := Command("build")
	assert.Equal(t, "build", cmd.Name(), "Expected command name to be 'build'")
}

func TestFileName(t *testing.T) {
	test := []struct {
		Command  Command
		FileName string
	}{{
		Command("test"),
		"go-mask_test.go",
	},
		{
			Command("build"),
			"go-mask.go",
		},
		{
			Command("run"),
			"go-mask.go",
		},
		{
			Command("invalid"),
			"go-mask.go",
		},
	}
	for _, tt := range test {
		t.Run(string(tt.Command), func(t *testing.T) {
			assert.Equal(t, tt.FileName, tt.Command.FileName(), "Expected filename to be 'go-mask.go'")
		})
	}
}

func TestLoadConfigError(t *testing.T) {
	t.Run("FileDoesNotExist", func(t *testing.T) {
		loader := &ConfigLoader{file.NewFileReaderError(os.ErrNotExist)}
		cfg, err := loader.LoadConfig()
		assert.NoError(t, err, "Failed to load config")
		assert.Equal(t, "run", string(cfg.Command), "Expected default command to be 'build'")
		assert.Equal(t, ".", cfg.Directory, "Expected default directory to be '.'")
	})

	t.Run("FileExistButIsNotAccessabile", func(t *testing.T) {
		loader := &ConfigLoader{file.NewFileReaderError(os.ErrPermission)}
		_, err := loader.LoadConfig()
		assert.Error(t, err, "Failed to load config")
	})

	t.Run("ErrorReadingFile", func(t *testing.T) {
		loader := &ConfigLoader{file.NewFileReader(iotest.ErrReader(os.ErrPermission))}
		_, err := loader.LoadConfig()
		assert.Error(t, err, "Expected an error when reading file")
	})

	t.Run("ErrorUnmarshallingYAML", func(t *testing.T) {
		loader := &ConfigLoader{file.NewFileReader(strings.NewReader("invalid_yaml: : :"))}
		_, err := loader.LoadConfig()
		assert.Error(t, err, "Expected an error when unmarshalling YAML")
		// assert.True(t, exitCalled, "os.Exit should have been called")
	})

	t.Run("ValidConfigFile", func(t *testing.T) {
		// Write valid YAML content
		yamlContent := `
args: "-v"
command: "run"
debug: true
directory: "./testdir"
imports:
  - "fmt"
  - "os"
mainfunc: true
package: "main"
output: "output.bin"
`
		// Check LoadConfig behavior
		loader := &ConfigLoader{file.NewFileReader(strings.NewReader(yamlContent))}
		cfg, err := loader.LoadConfig()
		assert.NoError(t, err, "Failed to load config")
		assert.Equal(t, "-v", cfg.Args, "Expected args to be '-v'")
		assert.Equal(t, "run", string(cfg.Command), "Expected command to be 'run'")
		assert.True(t, cfg.Debug, "Expected debug to be true")
		assert.Equal(t, "./testdir", cfg.Directory, "Expected directory to be './testdir'")
		assert.Equal(t, stringArray{"fmt", "os"}, cfg.Imports, "Expected imports to match")
		assert.True(t, cfg.MainFunc, "Expected mainfunc to be true")
		assert.Equal(t, "main", cfg.Package, "Expected package to be 'main'")
		assert.Equal(t, "output.bin", cfg.Output, "Expected output to be 'output.bin'")
	})
}

func TestLoadConfig(t *testing.T) {
	// Case 2: Valid YAML file, should load config correctly
	t.Run("ValidYAML", func(t *testing.T) {
		// Prepare a temporary YAML file
		yamlData := []byte(`
command: "run"
directory: "./tmp"
args: "-v"
debug: true
package: "main"
mainfunc: true
output: "out.go"
imports:
  - "fmt"
`)
		tmpFile := "./test-yml"
		defer os.Remove(tmpFile) // Clean up the temporary file
		if err := os.WriteFile(tmpFile, yamlData, 0644); err != nil {
			t.Fatal(err)
		}

		loader := NewConfigLoader(tmpFile)
		cfg, err := loader.LoadConfig()
		assert.NoError(t, err, "Failed to load config")
		assert.Equal(t, "run", string(cfg.Command), "Expected command to be 'run'")
		assert.Equal(t, "./tmp", cfg.Directory, "Expected directory to be './tmp'")
		assert.True(t, cfg.Debug, "Expected debug to be true")
		assert.Equal(t, "main", cfg.Package, "Expected package to be 'main'")
		assert.True(t, cfg.MainFunc, "Expected mainfunc to be true")
		assert.Equal(t, "out.go", cfg.Output, "Expected output to be 'out.go'")
		assert.Contains(t, cfg.Imports, "fmt", "Expected imports to contain 'fmt'")
	})
}

func TestApplyFlags(t *testing.T) {
	// Mock command-line flags
	os.Args = []string{
		"test",               // Program name
		"-command=test",      // Command flag
		"-debug=true",        // Debug flag
		"-directory=./temp",  // Directory flag
		"-package=mypackage", // Package flag
		"-mainfunc=false",    // Mainfunc flag
		"-output=build.go",   // Output flag
	}
	// Set up a test case for flags
	t.Run("FlagParsing", func(t *testing.T) {

		// Create an empty config struct to apply flags
		cfg := Config{}
		ApplyFlags(&cfg)

		// Validate that the flags are correctly applied
		assert.Equal(t, "test", string(cfg.Command), "Expected command to be 'test'")
		assert.True(t, cfg.Debug, "Expected debug to be true")
		assert.Equal(t, "./temp", cfg.Directory, "Expected directory to be './temp'")
		assert.Equal(t, "mypackage", cfg.Package, "Expected package to be 'mypackage'")
		assert.False(t, cfg.MainFunc, "Expected mainfunc to be false")
		assert.Equal(t, "build.go", cfg.Output, "Expected output to be 'build.go'")
	})
}

func TestStringArray(t *testing.T) {
	// Create a string array
	sa := stringArray{"a", "b", "c"}
	assert.Equal(t, "a,b,c", sa.String(), "Expected string representation to be 'a,b,c'")

	err := sa.Set("d,e,f")
	assert.NoError(t, err, "Expected no error when setting string array")
	assert.Equal(t, stringArray{"a", "b", "c", "d", "e", "f"}, sa, "Expected string array to be 'd,e,f'")

	sa = stringArray{}
	_ = sa.Set("d,e,f")
	assert.Equal(t, stringArray{"d", "e", "f"}, sa, "Expected string array to be 'd,e,f'")
}
