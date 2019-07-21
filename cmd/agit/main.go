package main

import (
	"fmt"
	"os"

	"github.com/goabstract/git"

	"github.com/spf13/cobra"
)

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
		Use:   "agit",
		Short: "abstract's git implementation in pure Go",
	}

	cmd.AddCommand(newInitCmd())

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
		panic(err)
	}
	_, err = git.InitRepository(pwd)
	return err
}
