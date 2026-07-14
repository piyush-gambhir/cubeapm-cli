package main

import (
	"os"

	"github.com/piyush-gambhir/cubeapm-cli/cli-go/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
