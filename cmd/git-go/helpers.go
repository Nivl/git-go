package main

import (
	"fmt"
	"io"

	git "github.com/Nivl/git-go"
	"github.com/Nivl/git-go/ginternals/config"
)

func loadRepository(cfg *globalFlags) (*git.Repository, error) {
	p, err := config.LoadConfig(cfg.env, config.LoadConfigOptions{
		WorkingDirectory: cfg.C.String(),
		GitDirPath:       cfg.GitDir,
		WorkTreePath:     cfg.WorkTree,
		IsBare:           cfg.Bare,
	})
	if err != nil {
		return nil, fmt.Errorf("could not create param: %w", err)
	}

	// run the command
	return git.OpenRepositoryWithParams(p, git.OpenOptions{
		IsBare: cfg.Bare,
	})
}

func fprintln(quiet bool, out io.Writer, msg ...interface{}) {
	if !quiet {
		fmt.Fprintln(out, msg...)
	}
}

func fprintf(quiet bool, out io.Writer, format string, a ...interface{}) {
	if !quiet {
		fmt.Fprintf(out, format, a...)
	}
}
