package main

import (
	"context"
	"os"
	"os/signal"

	"github.com/gemaraproj/gemara-mcp/internal/cli"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	if err := cli.New().ExecuteContext(ctx); err != nil {
		os.Exit(1)
	}
}
