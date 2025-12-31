package cmd

import (
	"github.com/levisegal/monay/services/holdings/config"
	"github.com/levisegal/monay/services/holdings/server"
	"github.com/spf13/cobra"
)

func serverCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "server",
		Short: "Start the holdings API server",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			cfg, err := config.Load()
			if err != nil {
				return err
			}

			return server.Start(ctx, cfg)
		},
	}

	return cmd
}
