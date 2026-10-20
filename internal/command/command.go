package command

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/urfave/cli/v3"

	"github.com/cozybadgerde/cress/internal/generator"
)

// Deps supplies version information and command output destinations.
type Deps struct {
	Version string
	Output  io.Writer
	Errors  io.Writer
}

// NewCommand creates the scope-first Cress CLI.
func NewCommand(deps Deps) *cli.Command {
	return &cli.Command{
		Name:      "cress",
		Usage:     "Generate small static sites offline",
		Version:   deps.Version,
		Writer:    deps.Output,
		ErrWriter: deps.Errors,
		Action:    showScopeHelp,
		Commands: []*cli.Command{
			{
				Name:   "site",
				Usage:  "Build a site",
				Action: showScopeHelp,
				Commands: []*cli.Command{
					{
						Name:      "build",
						Usage:     "Build ordinary pages into a new or empty public directory",
						ArgsUsage: "[directory]",
						Action:    buildSite,
					},
				},
			},
			{
				Name:   "theme",
				Usage:  "Theme development (commands arrive in a later phase)",
				Action: showScopeHelp,
			},
		},
	}
}

func showScopeHelp(ctx context.Context, command *cli.Command) error {
	if command.Args().Len() != 0 {
		return fmt.Errorf("unknown command %q for %s", command.Args().First(), command.FullName())
	}
	return cli.ShowSubcommandHelp(command)
}

func buildSite(ctx context.Context, command *cli.Command) error {
	if command.Args().Len() > 1 {
		return fmt.Errorf("site build accepts at most one directory")
	}
	directory := "."
	if command.Args().Present() {
		directory = command.Args().First()
	}
	result, buildErr := generator.Build(ctx, directory)
	for _, warning := range result.Warnings {
		if _, err := fmt.Fprintf(command.Root().ErrWriter, "warning: %s\n", warning); err != nil {
			return errors.Join(buildErr, fmt.Errorf("report build warning: %w", err))
		}
	}
	if buildErr != nil {
		return buildErr
	}
	if _, err := fmt.Fprintf(command.Root().Writer, "Built %d page(s) into public/.\n", result.Pages); err != nil {
		return fmt.Errorf("report build result: %w", err)
	}
	return nil
}
