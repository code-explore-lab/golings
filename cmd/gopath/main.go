package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"gopath/exercises"
	"gopath/internal/cli"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	if err := cli.Run(ctx, os.Args[1:], os.Stdout, os.Stderr, os.Stdin, exercises.FS); err != nil {
		fmt.Fprintln(os.Stderr, "gopath:", err)
		os.Exit(1)
	}
}
