// Command cress is a cozy static site generator: Markdown plus config plus a
// theme, rendered into a small, fast website.
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/urfave/cli/v3"

	"github.com/cozybadgerde/cress/internal/version"
)

// Flag names shared between a flag's declaration and the handlers that read it.
const (
	flagSource = "source"
	flagOutput = "output"
	flagDrafts = "drafts"
	flagAddr   = "addr"
	flagForce  = "force"
	flagDryRun = "dry-run"
)

func main() {
	cli.VersionPrinter = func(cmd *cli.Command) {
		fprintln(cmd.Root().Writer, version.Info())
	}

	if err := newRootCommand().Run(context.Background(), os.Args); err != nil {
		fprintln(os.Stderr, "cress:", err)
		os.Exit(1)
	}
}

// newRootCommand assembles the command tree. Bare `cress` prints help.
func newRootCommand() *cli.Command {
	return &cli.Command{
		Name:                  version.Name,
		Usage:                 "A cozy static site generator. Fresh little sites, fast.",
		Version:               version.Version,
		EnableShellCompletion: true,
		Commands: []*cli.Command{
			newBuildCommand(),
			newCleanCommand(),
			newInitCommand(),
			newServeCommand(),
			newVersionCommand(),
		},
		Action: runRoot,
	}
}

// runRoot handles bare `cress`: it prints help, and rejects an unknown
// subcommand rather than silently doing nothing.
func runRoot(_ context.Context, cmd *cli.Command) error {
	if args := cmd.Args(); args.Len() > 0 {
		return fmt.Errorf("unknown command %q", args.First())
	}
	return cli.ShowRootCommandHelp(cmd)
}

// newVersionCommand prints the version, mirroring `cress --version`.
func newVersionCommand() *cli.Command {
	return &cli.Command{
		Name:  "version",
		Usage: "Print the version and exit",
		Action: func(_ context.Context, cmd *cli.Command) error {
			fprintln(cmd.Root().Writer, version.Info())
			return nil
		},
	}
}

// sourceFlag is the shared --source/-s flag naming the site root directory.
func sourceFlag() cli.Flag {
	return &cli.StringFlag{
		Name:    flagSource,
		Aliases: []string{"s"},
		Usage:   "site root directory",
		Value:   ".",
	}
}

// draftsFlag is the shared --drafts flag that includes draft pages.
func draftsFlag() cli.Flag {
	return &cli.BoolFlag{
		Name:  flagDrafts,
		Usage: "include pages marked draft",
	}
}
