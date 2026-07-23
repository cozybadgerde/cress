package main

import (
	"context"

	"github.com/urfave/cli/v3"

	"github.com/cozybadgerde/cress/internal/build"
)

// newBuildCommand renders the site into the output directory.
func newBuildCommand() *cli.Command {
	return &cli.Command{
		Name:  "build",
		Usage: "Render the site into the output directory",
		Flags: []cli.Flag{
			sourceFlag(),
			&cli.StringFlag{
				Name:    flagOutput,
				Aliases: []string{"o"},
				Usage:   "output directory",
				Value:   build.OutputDir,
			},
			draftsFlag(),
		},
		Action: runBuild,
	}
}

func runBuild(_ context.Context, cmd *cli.Command) error {
	res, err := build.Build(build.Options{
		Root:   cmd.String(flagSource),
		Output: cmd.String(flagOutput),
		Drafts: cmd.Bool(flagDrafts),
	})
	if err != nil {
		return err
	}

	w := cmd.Root().Writer
	for _, warning := range res.Warnings {
		fprintf(w, "warning: %s\n", warning)
	}
	fprintf(w, "built %d page(s) into %s\n", res.Pages, res.Output)
	return nil
}
