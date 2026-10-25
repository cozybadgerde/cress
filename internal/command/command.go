package command

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/urfave/cli/v3"

	"github.com/cozybadgerde/cress/internal/generator"
	"github.com/cozybadgerde/cress/internal/preview"
	"github.com/cozybadgerde/cress/internal/scaffold"
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
				Usage:  "Initialize, build, and preview a site",
				Action: showScopeHelp,
				Commands: []*cli.Command{
					{
						Name: "init", Usage: "Create a new sample site in a new or empty directory",
						ArgsUsage: "[directory]", Action: initSite,
						Flags: []cli.Flag{&cli.StringFlag{Name: "theme", Value: theme.DefaultName, Usage: "Built-in theme: poppy or clover"}},
					},
					{
						Name:      "build",
						Usage:     "Build pages into public",
						ArgsUsage: "[directory]",
						Action:    buildSite,
					},
					{
						Name: "preview", Usage: "Watch and preview an isolated local site",
						ArgsUsage: "[directory]", Action: previewSite,
						Flags: []cli.Flag{
							&cli.BoolFlag{Name: "drafts", Usage: "Include drafts in this preview session"},
							&cli.IntFlag{Name: "port", Value: preview.DefaultPort, Usage: "Loopback port (0 chooses a free port)"},
						},
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

func initSite(ctx context.Context, command *cli.Command) error {
	if command.Args().Len() > 1 {
		return fmt.Errorf("site init accepts at most one directory")
	}
	directory := "."
	if command.Args().Present() {
		directory = command.Args().First()
	}
	if err := scaffold.Init(ctx, directory, command.String("theme")); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(command.Root().Writer, "Created site in %s.\nEdit its cress.toml and content/, then run: cress site build %q\n", directory, directory); err != nil {
		return fmt.Errorf("report site initialization: %w", err)
	}
	return nil
}

func previewSite(ctx context.Context, command *cli.Command) error {
	if command.Args().Len() > 1 {
		return fmt.Errorf("site preview accepts at most one directory")
	}
	directory := "."
	if command.Args().Present() {
		directory = command.Args().First()
	}
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()
	return preview.Run(ctx, preview.Options{
		Directory: directory, Version: command.Root().Version, Drafts: command.Bool("drafts"),
		Port: command.Int("port"), Output: command.Root().Writer, Errors: command.Root().ErrWriter,
	})
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
