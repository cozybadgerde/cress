package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/urfave/cli/v3"

	"github.com/cozybadgerde/cress/internal/scaffold"
)

// newInitCommand scaffolds a new site in the given directory (default ".").
func newInitCommand() *cli.Command {
	return &cli.Command{
		Name:      "init",
		Usage:     "Scaffold a new site (default: current directory)",
		ArgsUsage: "[dir]",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  flagAllowExisting,
				Usage: "scaffold into a directory that already has files",
			},
		},
		Action: runInit,
	}
}

func runInit(_ context.Context, cmd *cli.Command) error {
	dir := cmd.Args().First()
	if dir == "" {
		dir = "."
	}

	skipped, err := scaffold.Create(dir, cmd.Bool(flagAllowExisting))
	if err != nil {
		if errors.Is(err, scaffold.ErrExists) {
			return fmt.Errorf("%w (use --%s to scaffold alongside them)", err, flagAllowExisting)
		}
		return err
	}

	w := cmd.Root().Writer
	if len(skipped) > 0 {
		fprintf(w, "scaffolded a new cress site in %s (%d existing file(s) left untouched)\n", dir, len(skipped))
	} else {
		fprintf(w, "scaffolded a new cress site in %s\n", dir)
	}
	fprintln(w, "next: cress serve")
	return nil
}
