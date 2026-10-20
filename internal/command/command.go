package command

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/urfave/cli/v3"

	"github.com/cozybadgerde/cress/internal/generator"
	"github.com/cozybadgerde/cress/internal/theme"
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
						Usage:     "Build pages into public",
						ArgsUsage: "[directory]",
						Action:    buildSite,
					},
				},
			},
			{
				Name:   "theme",
				Usage:  "Theme development",
				Action: showScopeHelp,
				Commands: []*cli.Command{{
					Name: "init", Usage: "Create a new local theme from the embedded skeleton",
					ArgsUsage: "<name> [site-directory]", Action: initTheme,
				}},
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

func initTheme(ctx context.Context, command *cli.Command) error {
	if command.Args().Len() < 1 || command.Args().Len() > 2 {
		return fmt.Errorf("theme init requires a name and accepts an optional site directory")
	}
	directory := "."
	if command.Args().Len() == 2 {
		directory = command.Args().Get(1)
	}
	name := command.Args().First()
	if err := theme.Init(ctx, directory, name); err != nil {
		return err
	}
	_, err := fmt.Fprintf(command.Root().Writer, "Created themes/%s/. Set [theme] name = %q in cress.toml to select it.\n", name, name)
	if err != nil {
		return fmt.Errorf("report theme initialization: %w", err)
	}
	return nil
}

func buildSite(ctx context.Context, command *cli.Command) error {
	if command.Args().Len() > 1 {
		return fmt.Errorf("site build accepts at most one directory")
	}
	directory := "."
	if command.Args().Present() {
		directory = command.Args().First()
	}
	result, buildErr := generator.BuildVersion(ctx, directory, command.Root().Version)
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
