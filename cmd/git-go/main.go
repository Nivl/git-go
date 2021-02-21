package main

import (
	"fmt"
	"os"
)

func exitError(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}

func main() {
	root, err := newRootCmd()
	if err != nil {
		exitError(err)
	}

	if err = root.Execute(); err != nil {
		exitError(err)
	}
}
