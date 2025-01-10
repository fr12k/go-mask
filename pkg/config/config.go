package config

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/fr12k/go-mask/pkg/file"

	"gopkg.in/yaml.v3"
)

type (
	stringArray []string

	Command string

	ConfigLoader struct {
		File *file.File
	}

	Config struct {
		Args      string      `yaml:"args"`
		Command   Command     `yaml:"command"`
		FileName  string      `yaml:"filename"`
		Debug     bool        `yaml:"debug"`
		Directory string      `yaml:"directory"`
		Imports   stringArray `yaml:"imports"`
		MainFunc  bool        `yaml:"mainfunc"`
		Package   string      `yaml:"package"`
		Output    string      `yaml:"output"`

		// Internal fields
		Code string
	}
)

func (s *stringArray) String() string {
	return strings.Join(*s, ",")
}

func (s *stringArray) Set(value string) error {
	*s = append(*s, strings.Split(value, ",")...)
	return nil
}

func (c *Command) Name() string {
	return string(*c)
}

func (c *Config) SaveAs() string {
	if c.FileName != "" {
		return c.FileName
	}
	switch c.Command {
	case "test":
		c.FileName = "go-mask_test.go"
	case "build", "run":
		c.FileName = "go-mask.go"
	default:
		c.FileName = "go-mask.go"
	}
	return c.FileName
}

func NewConfigLoaderBuffer(content string) *ConfigLoader {
	return &ConfigLoader{file.NewFileReader(strings.NewReader(content))}
}

func NewConfigLoader(filename string) *ConfigLoader {
	return &ConfigLoader{file.NewFile(filename)}
}

func (c *ConfigLoader) LoadConfig() (*Config, error) {
	defer c.File.Close()
	exist, err := c.File.Exists()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error checking file existence: %v", err)
		return nil, err
	}
	if !exist {
		return &Config{Command: "run", Directory: "."}, nil
	}
	data, err := c.File.Read()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file: %v", err)
		return nil, err
	}
	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error unmarshalling YAML: %v", err)
		return nil, err
	}
	return &config, nil
}

func ApplyFlags(cfg *Config) {
	fs := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	flag.CommandLine = fs
	// flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	fs.Var(&cfg.Imports, "i", "Comma-separated list of strings")
	fs.StringVar(&cfg.Args, "args", cfg.Args, "Arguments to pass to the go command")
	fs.StringVar((*string)(&cfg.Command), "command", (string)(cfg.Command), "Command to run (build, run, test)")
	fs.BoolVar(&cfg.Debug, "debug", cfg.Debug, "Enable debug mode")
	fs.StringVar(&cfg.Directory, "directory", cfg.Directory, "Directory for temporary files")
	fs.StringVar(&cfg.Package, "package", cfg.Package, "Go package name")
	fs.BoolVar(&cfg.MainFunc, "mainfunc", cfg.MainFunc, "Wrap code in main function")
	fs.StringVar(&cfg.Output, "output", cfg.Output, "Output file name for build command")
	fs.StringVar(&cfg.Code, "c", cfg.Code, "Go code to run")
	_ = fs.Parse(os.Args[1:])
}
