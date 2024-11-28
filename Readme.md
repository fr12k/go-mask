# go-mask

`go-mask` is a Go-based tool that helps you generate, build, and run Go code directly from the terminal. It provides an easy-to-use interface for creating Go files with optional imports, the inclusion of a `main` function, and direct code input either from stdin or a command-line flag. The tool simplifies the process of compiling and executing Go code with flexible options for testing and debugging.

This project is part of the larger [mask](https://github.com/jacobdeichert/mask) project by Jacob Deichert, which integrates the `go-mask` program to automate the process of compiling, running, and testing Go code as part of an extended Go code management framework.

## Features

- **Generate Go Code**:
  - Use the `-i` flag to specify packages to import.
  - Include or exclude the `package main` declaration with the `-p` flag.
  - Wrap your code in a `main()` function using the `-m` flag.
  - Optionally pass Go code directly using the `-c` flag instead of reading from stdin.

- **Build Go Code**:
  - Generate a `.go` file, write it to a temporary directory, and compile it with `go build`.
  - Creates an executable file in the `tmp` directory of your current working directory.

- **Debug Mode**:
  - Use the `-d` flag to print the generated Go code to the console instead of building and running it.

## Installation

### With Go

To install `go-mask` using Go, run the following command:

```bash
go get -u github.com/fr12k/go-mask
```

### From Github

To install `go-mask`, clone this repository and build the Go program:

Clone the Repository first then build the program
```
git clone https://github.com/your-username/go-mask.git
cd go-mask
```

```bash
go build -o go-mask
```

Alternatively, you can download the precompiled binaries from the releases section.

## Usage

You can use `go-mask` from the command line to generate Go code, build it, and run it.

### Command-Line Flags

- `-i` or `--import`: Comma-separated list of Go packages to import (e.g., `fmt,os`).
- `-p` or `--no-package`: Excludes `package main` from the generated code (default is enabled).
- `-m` or `--main`: Wraps the input code in a `main()` function block (default is disabled).
- `-c` or `--code`: Pass Go code directly as a string. This overrides stdin input.
- `-d` or `--debug`: Prints the generated Go code instead of building and running it.

### Examples

1. **Generate Go code and run it:**

   ```bash
   go run go-mask.go -i fmt -m
   ```

   This will prompt you to enter Go code via stdin. Once you finish, the code will be compiled and run.

2. **Pass Go code directly with the `-c` flag:**

   ```bash
   go run go-mask.go -i fmt -c "fmt.Println(\"Hello, Go!\")" -m
   ```

   This will run the Go code `fmt.Println("Hello, Go!")` directly without needing stdin input.

3. **Enable Debug Mode to print the generated code:**

   ```bash
   go run go-mask.go -i fmt -c "fmt.Println(\"Hello, Debug!\")" -d
   ```

   This will print the generated Go code instead of building and running it.

4. **Compile and run the generated code (without stdin):**

   ```bash
   go run go-mask.go -i fmt -c "fmt.Println(\"Hello, Go from go-mask!\")"
   ```

   This will compile and execute the code after generating the Go file.

### Generated Executable

By default, `go-mask` generates a `.go` file in a temporary directory (`./tmp`) and compiles it into an executable named `script` (or `script.exe` on Windows). The executable will be placed in the same directory.

After successful execution, you can find the output file in `./tmp/script` or `./tmp/script.exe` depending on your operating system.

## Integration with [mask](https://github.com/jacobdeichert/mask)

`go-mask` is used by the [mask](https://github.com/jacobdeichert/mask) project to manage the process of compiling, running, and testing Go code. The `mask` project provides a flexible framework for Go code testing and automation. By integrating `go-mask`, `mask` makes it easy to run Go code snippets, apply automated testing strategies, and ensure that your Go code is functioning as expected.

In the `mask` project, `go-mask` is utilized to:

- Generate Go files dynamically with the required imports.
- Test Go code by running it as part of a larger testing pipeline.
- Handle the automation of compiling and running Go code as part of a CI/CD workflow.

For more information on how `go-mask` integrates into the `mask` project, visit the [mask repository](https://github.com/jacobdeichert/mask).

## Contributing

Contributions to the `go-mask` project are welcome! Feel free to open issues or submit pull requests. To contribute:

1. Fork the repository.
2. Create a new branch (`git checkout -b feature-branch`).
3. Commit your changes (`git commit -am 'Add new feature'`).
4. Push to the branch (`git push origin feature-branch`).
5. Open a pull request.

## License

`go-mask` is licensed under the MIT License. See the [LICENSE](LICENSE) file for more details.
