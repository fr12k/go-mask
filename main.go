package main

import (
	"fmt"
	"os"

	"github.com/fr12k/go-mask/cmd"
)

func main() {
	err := cmd.NewGoMask().Run()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
