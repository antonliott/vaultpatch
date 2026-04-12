package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/your-org/vaultpatch/internal/diff"
	"github.com/your-org/vaultpatch/internal/vault"
	"github.com/your-org/vaultpatch/internal/watch"
)

varcobra.ExactArgs(1),
	RunE:  runWatch,
}

var watchInterval time.Duration

func init() {
	watchCmd.Flags().DurationVar(&watchInterval, "interval", 30*time.Second,
		"Polling interval (e.g. 10s, 1m)")
	rootCmd.AddCommand(watchCmd)
}

func runWatch(cmd *cobra.Command, args []string) error {
	path := args[0]

	client, err := vault.NewClient(vault.Config{
		Address:   os.Getenv("VAULT_ADDR"),
		Token:     os.Getenv("VAULT_TOKEN"),
		Namespace: os.Getenv("VAULT_NAMESPACE"),
	})
	if err != nil {
		return fmt.Errorf("vault client: %w", err)
	}

	w := watch.NewWatcher(client, path, watchInterval)

	ctx, cancel := signal.NotifyContext(context.Background(),
		os.Interrupt, syscall.SIGTERM)
	defer cancel()

	fmt.Fprintf(cmd.OutOrStdout(), "Watching %s every %s — press Ctrl+C to stop\n",
		path, watchInterval)

	go func() {
		if err := w.Run(ctx); err != nil && err != context.Canceled {
			fmt.Fprintf(cmd.ErrOrStderr(), "watcher error: %v\n", err)
		}
	}()

	for {
		select {
		case ev, ok := <-w.Changes:
			if !ok {
				return nil
			}
			fmt.Fprintf(cmd.OutOrStdout(), "\n[%s] changes detected at %s:\n",
				ev.SeenAt.Format(time.RFC3339), ev.Path)
			fmt.Fprint(cmd.OutOrStdout(), diff.Format(ev.Deltas))
		case <-ctx.Done():
			return nil
		}
	}
}
