package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/urfave/cli/v3"

	"github.com/cozybadgerde/cress/internal/build"
	"github.com/cozybadgerde/cress/internal/clean"
)

// newCleanCommand empties the output directory.
func newCleanCommand() *cli.Command {
	return &cli.Command{
		Name:  "clean",
		Usage: "Remove everything in the output directory",
		Flags: []cli.Flag{
			sourceFlag(),
			&cli.StringFlag{
				Name:    flagOutput,
				Aliases: []string{"o"},
				Usage:   "output directory",
				Value:   build.OutputDir,
			},
			&cli.BoolFlag{
				Name:    flagForce,
				Aliases: []string{"f"},
				Usage:   "remove without asking for confirmation",
			},
			&cli.BoolFlag{
				Name:  flagDryRun,
				Usage: "list what would be removed, and remove nothing",
			},
		},
		Action: runClean,
	}
}

func runClean(_ context.Context, cmd *cli.Command) error {
	plan, err := clean.Prepare(clean.Options{
		Root:   cmd.String(flagSource),
		Output: cmd.String(flagOutput),
	})
	if err != nil {
		return fmt.Errorf("refusing to clean: %w", err)
	}

	w := cmd.Root().Writer
	if plan.Empty() {
		fprintf(w, "nothing to remove in %s\n", plan.Output)
		return nil
	}

	if cmd.Bool(flagDryRun) {
		for _, file := range plan.Files {
			fprintf(w, "would remove %s\n", file)
		}
		fprintf(w, "%d file(s) in %s, nothing removed\n", len(plan.Files), plan.Output)
		return nil
	}

	if !cmd.Bool(flagForce) {
		ok, err := confirmClean(w, cmd.Root().Reader, plan)
		if err != nil {
			return err
		}
		if !ok {
			fprintln(w, "nothing removed")
			return nil
		}
	}

	if err := plan.Execute(); err != nil {
		return err
	}
	fprintf(w, "removed %d file(s) from %s\n", len(plan.Files), plan.Output)
	return nil
}

// confirmClean asks before deleting. The question is not the point: the
// resolved absolute path printed with it is, since --output makes the target
// something the user has to recognize rather than assume.
func confirmClean(w io.Writer, r io.Reader, plan *clean.Plan) (bool, error) {
	if !interactive(r) {
		return false, unanswered(plan)
	}

	fprintf(w, "remove %d file(s) from %s? [y/N]: ", len(plan.Files), plan.Output)
	answer, err := bufio.NewReader(r).ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return false, fmt.Errorf("reading confirmation: %w", err)
	}
	// Reaching the end of the input without an answer means nobody was there to
	// give one. /dev/null arrives here rather than at the check above, because it
	// is a character device and so indistinguishable from a terminal without
	// asking the kernel. Refusing is what keeps `cress clean </dev/null` in a CI
	// job from reporting success for work it did not do.
	if errors.Is(err, io.EOF) && strings.TrimSpace(answer) == "" {
		fprintln(w)
		return false, unanswered(plan)
	}
	return affirmative(answer), nil
}

// unanswered reports that the confirmation went unanswered, and names the two
// flags that do not need one.
func unanswered(plan *clean.Plan) error {
	return fmt.Errorf(
		"nothing answered the confirmation: pass --%s to remove %d file(s) from %s, or --%s to see them",
		flagForce, len(plan.Files), plan.Output, flagDryRun)
}

// affirmative reports whether an answer to the confirmation is a yes. Anything
// else is a no, including an empty line: the default has to be the outcome that
// destroys nothing, since that is the one a mistaken keystroke can be undone
// from.
func affirmative(answer string) bool {
	switch strings.ToLower(strings.TrimSpace(answer)) {
	case "y", "yes":
		return true
	default:
		return false
	}
}

// interactive reports whether r is a terminal somebody can answer from. A pipe,
// a file, or a test buffer is not: prompting there would hang a CI job or read
// end-of-input as a decision nobody made.
func interactive(r io.Reader) bool {
	f, ok := r.(*os.File)
	if !ok {
		return false
	}
	info, err := f.Stat()
	return err == nil && info.Mode()&os.ModeCharDevice != 0
}
