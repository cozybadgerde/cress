package main

import (
	"context"
	"os"
	"os/signal"

	"github.com/urfave/cli/v3"

	"github.com/cozybadgerde/cress/internal/serve"
)

// newServeCommand builds the site and serves it locally, rebuilding on change.
func newServeCommand() *cli.Command {
	return &cli.Command{
		Name:  "serve",
		Usage: "Preview the site locally, rebuilding on change",
		Flags: []cli.Flag{
			sourceFlag(),
			&cli.StringFlag{
				Name:  flagAddr,
				Usage: "listen address",
				Value: serve.DefaultAddr,
			},
			draftsFlag(),
		},
		Action: runServe,
	}
}

func runServe(ctx context.Context, cmd *cli.Command) error {
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt)
	defer stop()

	w := cmd.Root().Writer
	return serve.Run(ctx, serve.Options{
		Root:   cmd.String(flagSource),
		Addr:   cmd.String(flagAddr),
		Drafts: cmd.Bool(flagDrafts),
		Logf: func(format string, args ...any) {
			fprintf(w, format+"\n", args...)
		},
	})
}
