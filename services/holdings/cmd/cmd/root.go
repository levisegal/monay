package cmd

import (
	"context"

	"github.com/spf13/cobra"
)

func Execute(ctx context.Context) error {
	cmd := Command()
	cmd.SetContext(ctx)
	return cmd.Execute()
}

func Command() *cobra.Command {
	command := &cobra.Command{
		Use:   "holdings",
		Short: "Monay holdings service",
	}

	command.AddCommand(serverCommand())
	command.AddCommand(plaidCommand())
	command.AddCommand(importCommand())
	command.AddCommand(lotsCommand())
	command.AddCommand(holdingsCommand())

	return command
}

