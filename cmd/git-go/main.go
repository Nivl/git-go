package main

import (
	"fmt"
	"os"

	"github.com/Nivl/git-go/internal/env"
)

func exitError(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}

func main() {
	cwd, err := os.Getwd()
	if err != nil {
		exitError(err)
	}

	root := newRootCmd(cwd, env.NewFromOs())
	if err = root.Execute(); err != nil {
		exitError(err)
	}
}
