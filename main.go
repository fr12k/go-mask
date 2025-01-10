package main

import (
	"fmt"
	"os"

	"github.com/fr12k/go-mask/cmd"
)

func main() {
	res, err := cmd.NewGoMask().Run()
	fmt.Print(res.Stdout)
	fmt.Print(res.Stderr)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
