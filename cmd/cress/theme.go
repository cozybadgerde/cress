package main

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/urfave/cli/v3"

	"github.com/cozybadgerde/cress/internal/config"
	"github.com/cozybadgerde/cress/internal/scaffold"
)

// newThemeCommand groups the commands that act on a theme rather than on a
// site. The group exists so that later additions land under it instead of
// widening the top-level command surface once per theme-related idea.
func newThemeCommand() *cli.Command {
	return &cli.Command{
		Name:   "theme",
		Usage:  "Work with themes",
		Action: runTheme,
		Commands: []*cli.Command{
			newThemeInitCommand(),
		},
	}
}

// runTheme handles bare `cress theme`: it prints help, and rejects an unknown
// subcommand rather than silently doing nothing, mirroring the root command.
func runTheme(_ context.Context, cmd *cli.Command) error {
	if args := cmd.Args(); args.Len() > 0 {
		return fmt.Errorf("unknown theme command %q", args.First())
	}
	return cli.ShowSubcommandHelp(cmd)
}

// newThemeInitCommand scaffolds a starter theme under themes/.
func newThemeInitCommand() *cli.Command {
	return &cli.Command{
		Name:      "init",
		Usage:     "Scaffold a starter theme under themes/",
		ArgsUsage: "<name>",
		Flags: []cli.Flag{
			sourceFlag(),
			&cli.BoolFlag{
				Name:  flagAllowExisting,
				Usage: "scaffold into a theme directory that already has files",
			},
		},
		Action: runThemeInit,
	}
}

func runThemeInit(_ context.Context, cmd *cli.Command) error {
	name := cmd.Args().First()
	if name == "" {
		return errors.New("theme name required (cress theme init <name>)")
	}
	// The name is joined onto themes/, so it is checked before it reaches a
	// path, on the same rule that decides whether a configured theme can be
	// loaded at all.
	if !config.ValidThemeName(name) {
		return fmt.Errorf("invalid theme name %q (want a directory name under %s/, not a path)", name, config.ThemesDir)
	}

	dir := filepath.Join(cmd.String(flagSource), config.ThemesDir, name)
	skipped, err := scaffold.CreateTheme(dir, cmd.Bool(flagAllowExisting))
	if err != nil {
		if errors.Is(err, scaffold.ErrExists) {
			return fmt.Errorf("%w (use --%s to scaffold alongside them)", err, flagAllowExisting)
		}
		return err
	}

	w := cmd.Root().Writer
	if len(skipped) > 0 {
		fprintf(w, "scaffolded theme %q in %s (%d existing file(s) left untouched)\n", name, dir, len(skipped))
	} else {
		fprintf(w, "scaffolded theme %q in %s\n", name, dir)
	}
	fprintf(w, "next: set theme = %q in cress.toml, then cress serve\n", name)
	return nil
}
