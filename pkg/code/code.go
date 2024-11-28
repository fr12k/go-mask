package code

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/fr12k/go-mask/pkg/config"
)

type CodeReader struct {
	code io.Reader
}

func NewCodeReader(code io.Reader) *CodeReader {
	return &CodeReader{code}
}

func (c *CodeReader) ReadCode() (string, error) {
	str, err := readAllCode(c.code)
	if err != nil {
		return "", err
	}
	if str == "" {
		var input strings.Builder
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			input.WriteString(scanner.Text() + "\n")
		}
		if err := scanner.Err(); err != nil {
			fmt.Fprintf(os.Stderr, "Error reading from stdin: %v\n", err)
			return "", err
		}
		c.code = strings.NewReader(input.String())
		return input.String(), nil
	}

	c.code = strings.NewReader(str)
	fmt.Printf("Code: %s\n", str)
	return str, nil
}

func (c *CodeReader) GenerateGoCode(cfg *config.Config) (string, error) {
	code, err := c.ReadCode()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading code: %v\n", err)
		return "", err
	}
	var out strings.Builder
	if cfg.Package != "" {
		out.WriteString(fmt.Sprintf("package %s\n\n", cfg.Package))
	}

	for _, pkg := range cfg.Imports {
		out.WriteString(fmt.Sprintf("import \"%s\"\n", strings.TrimSpace(pkg)))
	}
	if len(cfg.Imports) > 0 {
		out.WriteString("\n")
	}

	if cfg.MainFunc {
		out.WriteString("func main() {\n")
		out.WriteString(code)
		out.WriteString("\n")
		out.WriteString("}\n")
	} else {
		out.WriteString(code)
		out.WriteString("\n")
	}

	return out.String(), nil
}

func readAllCode(code io.Reader) (string, error) {
	if code == nil {
		return "", nil
	}
	b, err := io.ReadAll(code)
	return string(b), err
}
