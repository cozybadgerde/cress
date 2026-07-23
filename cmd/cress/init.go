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
				Name:    flagForce,
				Aliases: []string{"f"},
				Usage:   "write into a non-empty directory",
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

	if err := scaffold.Create(dir, cmd.Bool(flagForce)); err != nil {
		if errors.Is(err, scaffold.ErrExists) {
			return fmt.Errorf("%w (use --force to write anyway)", err)
		}
		return err
	}

	w := cmd.Root().Writer
	fprintf(w, "scaffolded a new cress site in %s\n", dir)
	fprintln(w, "next: cress serve")
	return nil
}
