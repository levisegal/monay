package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/levisegal/monay/services/holdings/cmd/cmd"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	defer func() {
		signal.Stop(sigChan)
		close(sigChan)
	}()

	go func() {
		sig := <-sigChan
		slog.InfoContext(ctx, "Received signal, initiating shutdown", "signal", sig.String())
		cancel()
	}()

	if err := cmd.Execute(ctx); err != nil {
		os.Exit(1)
	}
}


