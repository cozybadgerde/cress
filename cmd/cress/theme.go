package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"path/filepath"

	"github.com/urfave/cli/v3"

	"github.com/cozybadgerde/cress/internal/config"
	"github.com/cozybadgerde/cress/internal/scaffold"
	"github.com/cozybadgerde/cress/internal/theme"
	"github.com/cozybadgerde/cress/internal/version"
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
			newThemeValidateCommand(),
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

// newThemeValidateCommand checks a theme against the template contract.
func newThemeValidateCommand() *cli.Command {
	return &cli.Command{
		Name:      "validate",
		Usage:     "Check a theme against the template contract",
		ArgsUsage: "[name]",
		Flags: []cli.Flag{
			sourceFlag(),
			&cli.BoolFlag{
				Name:    flagQuiet,
				Aliases: []string{"q"},
				Usage:   "print nothing and report the result in the exit status",
			},
		},
		Action: runThemeValidate,
	}
}

// runThemeValidate validates one theme and exits non-zero if anything is wrong,
// so a theme author can put it in their own pipeline.
//
// Every problem it prints is one it is sure about. That is what makes failing
// the command the right response to any of them, and it is why a check that
// could not be made certain was left out rather than softened into a warning
// nobody would act on.
func runThemeValidate(_ context.Context, cmd *cli.Command) error {
	root := cmd.String(flagSource)
	name, err := themeToValidate(root, cmd.Args().First())
	if err != nil {
		return err
	}

	report, err := theme.Validate(theme.ValidateOptions{
		SiteRoot:     root,
		ThemesDir:    config.ThemesDir,
		Name:         name,
		CressVersion: version.Version,
	})
	if err != nil {
		return err
	}

	if !cmd.Bool(flagQuiet) {
		printReport(cmd.Root().Writer, report)
	}
	if !report.OK() {
		return errFailed
	}
	return nil
}

// themeToValidate settles which theme was meant: the one named on the command
// line, or failing that the one the site already renders with, since that is
// the theme somebody standing in a site is working on.
func themeToValidate(root, named string) (string, error) {
	name := named
	if name == "" {
		cfg, err := config.Load(filepath.Join(root, config.FileName))
		if err != nil {
			if errors.Is(err, config.ErrNotFound) {
				return "", fmt.Errorf("no %s in %s, so there is no theme to infer: name one (cress theme validate <name>)", config.FileName, root)
			}
			return "", err
		}
		name = cfg.Site.Theme
	}
	// The name is joined onto themes/, so it is checked on the same rule that
	// decides whether a configured theme can be loaded at all.
	if !config.ValidThemeName(name) {
		return "", fmt.Errorf("invalid theme name %q (want a directory name under %s/, not a path)", name, config.ThemesDir)
	}
	return name, nil
}

// printReport writes a report the way somebody reads a compiler's: a heading
// that says how much there is, then one problem per line with its place first,
// so the locations line up and the eye runs down them.
func printReport(w io.Writer, report *theme.Report) {
	if report.OK() {
		fprintf(w, "%s: no problems found\n", report.Path)
		return
	}

	fprintf(w, "%s: %d problem(s)\n", report.Path, len(report.Findings))
	places := make([]string, len(report.Findings))
	width := 0
	for i, finding := range report.Findings {
		places[i] = place(finding)
		width = max(width, len(places[i]))
	}
	for i, finding := range report.Findings {
		fprintf(w, "  %-*s  %s\n", width, places[i], finding.Message)
	}
}

// place names where a finding is. A finding with no line is about the file as a
// whole, and one with no file is about the theme as a whole.
func place(finding theme.Finding) string {
	switch {
	case finding.File == "":
		return "."
	case finding.Line == 0:
		return finding.File
	default:
		return fmt.Sprintf("%s:%d", finding.File, finding.Line)
	}
}
