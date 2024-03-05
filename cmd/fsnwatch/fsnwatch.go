package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

func main() {
	console := &cobra.Command{
		Use: "console [dir]", Short: "watches a directory and prints the changes to stdout",
		RunE: runConsole,
	}

	root := &cobra.Command{
		Use: "fswatch",
	}
	root.AddCommand(hostedCommand())
	root.AddCommand(console)

	if err := root.Execute(); err != nil {
		if _, err := fmt.Fprintf(os.Stderr, "Failed %s\n", err.Error()); err != nil {
			panic(err)
		}
	}
}
