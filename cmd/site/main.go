package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/f110/site/pkg/cmd/site"
)

func command(args []string) error {
	rootCmd := &cobra.Command{Use: "site"}
	site.Update(rootCmd)

	rootCmd.SetArgs(args)
	return rootCmd.Execute()
}

func main() {
	if err := command(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "%+v", err)
		os.Exit(1)
	}
}
