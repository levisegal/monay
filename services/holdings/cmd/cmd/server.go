package cmd

import (
	"fmt"
	"log/slog"

	"github.com/levisegal/monay/services/holdings/config"
	"github.com/levisegal/monay/services/holdings/database"
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

			slog.Info("database config",
				"host", cfg.Database.Host,
				"port", cfg.Database.Port,
				"database", cfg.Database.Name,
				"user", cfg.Database.User,
			)

			if err := database.MigrateUp(ctx, cfg.Database.ConnString()); err != nil {
				return fmt.Errorf("database migrations failed: %w", err)
			}

			return server.Start(ctx, cfg)
		},
	}

	return cmd
}
