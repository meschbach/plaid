package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/meschbach/plaid/controllers/filewatch/fsn"
	"github.com/spf13/cobra"
	"github.com/thejerf/suture/v4"
	"golang.org/x/sys/unix"
	"os"
	"os/signal"
)

func main() {
	root := &cobra.Command{
		Use: "fswatch",
		RunE: func(cmd *cobra.Command, args []string) error {
			rootContext, done := signal.NotifyContext(cmd.Context(), unix.SIGINT, unix.SIGTERM, unix.SIGINT)
			defer done()

			var watchDir string
			if len(args) > 0 {
				watchDir = args[0]
			} else {
				wd, err := os.Getwd()
				if err != nil {
					return err
				}
				watchDir = wd
			}

			core := fsn.NewCore()

			supervisionTree := suture.NewSimple("root")
			supervisionTree.Add(core)
			treeErr := supervisionTree.ServeBackground(rootContext)

			fmt.Printf("Starting fswatch of %s\n", watchDir)
			err := core.Watch2(rootContext, fsn.WatchConfig{
				Path:          watchDir,
				ExcludeSuffix: []string{".git", ".idea", "~"},
			})
			if err != nil {
				return err
			}

			for {
				select {
				case change := <-core.Output:
					fmt.Printf("Change: %s\n", change)
				case <-rootContext.Done():
					err := rootContext.Err()

					if errors.Is(err, context.Canceled) {
						err = nil
					}
					return err
				case err := <-treeErr:
					if errors.Is(err, context.Canceled) {
						err = nil
					}
					return err
				}
			}
		},
	}

	if err := root.Execute(); err != nil {
		if _, err := fmt.Fprintf(os.Stderr, "Failed %s\n", err.Error()); err != nil {
			panic(err)
		}
	}
}
