package main

import (
	"context"
	"log"
	"os"

	"github.com/cozybadgerde/cress/internal/command"
)

var version = "dev"

func main() {
	cmd := command.NewCommand(command.Deps{
		Version: version,
		Output:  os.Stdout,
		Errors:  os.Stderr,
	})
	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.New(os.Stderr, "cress: ", 0).Print(err)
		os.Exit(1)
	}
}
