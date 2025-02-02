package config

import (
	"os"
	"strings"
	"testing"
	"testing/iotest"

	"github.com/fr12k/go-mask/pkg/file"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestName(t *testing.T) {
	cmd := Command("build")
	assert.Equal(t, "build", cmd.Name(), "Expected command name to be 'build'")
}

func TestFileName(t *testing.T) {
	test := []struct {
		Config   Config
		FileName string
	}{
		{
			Config{Command: "test", FileName: "test.go"},
			"test.go",
		},
		{
			Config{Command: "test"},
			"go-mask_test.go",
		},
		{
			Config{Command: "build"},
			"go-mask.go",
		},
		{
			Config{Command: "run"},
			"go-mask.go",
		},
		{
			Config{Command: "invalid"},
			"go-mask.go",
		},
	}
	for _, tt := range test {
		t.Run(string(tt.Config.Command), func(t *testing.T) {
			assert.Equal(t, tt.FileName, tt.Config.SaveAs(), "Expected filename to be 'go-mask.go'")
		})
	}
}

func TestLoadConfigError(t *testing.T) {
	t.Run("FileDoesNotExist", func(t *testing.T) {
		loader := &Loader{file.NewReaderError(os.ErrNotExist)}
		cfg, err := loader.LoadConfig()
		require.NoError(t, err, "Failed to load config")
		assert.Equal(t, "run", string(cfg.Command), "Expected default command to be 'build'")
		assert.Equal(t, ".", cfg.Directory, "Expected default directory to be '.'")
	})

	t.Run("FileExistButIsNotAccessabile", func(t *testing.T) {
		loader := &Loader{file.NewReaderError(os.ErrPermission)}
		_, err := loader.LoadConfig()
		assert.Error(t, err, "Failed to load config")
	})

	t.Run("ErrorReadingFile", func(t *testing.T) {
		loader := &Loader{file.NewReader(iotest.ErrReader(os.ErrPermission))}
		_, err := loader.LoadConfig()
		assert.Error(t, err, "Expected an error when reading file")
	})

	t.Run("ErrorUnmarshallingYAML", func(t *testing.T) {
		loader := &Loader{file.NewReader(strings.NewReader("invalid_yaml: : :"))}
		_, err := loader.LoadConfig()
		assert.Error(t, err, "Expected an error when unmarshalling YAML")
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
		loader := &Loader{file.NewReader(strings.NewReader(yamlContent))}
		cfg, err := loader.LoadConfig()
		require.NoError(t, err, "Failed to load config")
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
		err := os.WriteFile(tmpFile, yamlData, 0o600)
		require.NoError(t, err)

		loader := NewLoader(tmpFile)
		cfg, err := loader.LoadConfig()
		require.NoError(t, err, "Failed to load config")
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
		err := ApplyFlags(&cfg)
		require.NoError(t, err)

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
	require.NoError(t, err, "Expected no error when setting string array")
	assert.Equal(t, stringArray{"a", "b", "c", "d", "e", "f"}, sa, "Expected string array to be 'd,e,f'")

	sa = stringArray{}
	err = sa.Set("d,e,f")
	require.NoError(t, err, "Expected no error when setting string array")
	assert.Equal(t, stringArray{"d", "e", "f"}, sa, "Expected string array to be 'd,e,f'")
}

func TestNewLoaderBuffer(t *testing.T) {
	tests := []struct {
		name   string
		data   string
		config Config
	}{
		{
			name:   "EmptyConfig",
			data:   "",
			config: Config{},
		},
		{
			name: "ValidConfig",
			data: `
command: "run"
directory: "./tmp"
args: "-v"
debug: true
package: "main"
mainfunc: true
output: "out.go"`,
			config: Config{
				Command:   "run",
				Directory: "./tmp",
				Args:      "-v",
				Debug:     true,
				Package:   "main",
				MainFunc:  true,
				Output:    "out.go",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loader := NewLoaderBuffer(tt.data)
			cfg, err := loader.LoadConfig()
			require.NoError(t, err, "Failed to load config")
			assert.Equal(t, tt.config, *cfg, "Expected config to match")
		})
	}
}
