package main

import (
	"fmt"
	"os"

	"github.com/Nivl/git-go"

	"github.com/spf13/cobra"
)

type config struct {
	C *string
}

func main() {
	root := newRootCmd()
	err := root.Execute()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "git-go",
		Short:         "git implementation in pure Go",
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	cfg := &config{}
	cfg.C = cmd.PersistentFlags().StringP("C", "C", "", "Run as if git was started in the provided path instead of the current working directory.")

	// porcelain
	cmd.AddCommand(newInitCmd())

	// plumbing
	cmd.AddCommand(newCatFileCmd(cfg))
	cmd.AddCommand(newHashObjectCmd())

	return cmd
}

func newInitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "init a new git repository",
	}

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		return initCmd()
	}

	return cmd
}

func initCmd() error {
	pwd, err := os.Getwd()
	if err != nil {
		return err
	}
	_, err = git.InitRepository(pwd)
	return err
}
