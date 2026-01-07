package cmd

import (
	"fmt"
	"log/slog"

	"github.com/spf13/cobra"

	"github.com/levisegal/monay/services/holdings/config"
	"github.com/levisegal/monay/services/holdings/database"
	"github.com/levisegal/monay/services/holdings/gen/db"
	"github.com/levisegal/monay/services/holdings/taxlots"
)

func lotsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "lots",
		Short: "Tax lot management commands",
	}

	cmd.AddCommand(processLotsCommand())

	return cmd
}

func processLotsCommand() *cobra.Command {
	var accountName string

	cmd := &cobra.Command{
		Use:   "process",
		Short: "Process tax lots from transactions (FIFO matching)",
		Long: `Process transactions to create tax lots and match sells to buys using FIFO.
		
This should be run after importing all transaction history for an account.
It will clear existing lots and recompute from scratch.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			cfg, err := config.Load()
			if err != nil {
				return err
			}

			conn, err := database.Open(ctx, cfg.Database.ConnString())
			if err != nil {
				return err
			}
			defer conn.Close()

			queries := db.New(conn)

			account, err := queries.GetAccountByName(ctx, accountName)
			if err != nil {
				return fmt.Errorf("account not found: %s", accountName)
			}

			slog.Info("processing lots", "account", account.Name, "account_id", account.ID)

			processor := taxlots.NewProcessor(queries)
			if err := processor.ProcessTransactions(ctx, account.ID); err != nil {
				return fmt.Errorf("failed to process tax lots: %w", err)
			}

			slog.Info("lot processing complete", "account", account.Name)

			return nil
		},
	}

	cmd.Flags().StringVar(&accountName, "account-name", "", "Account name to process")
	cmd.MarkFlagRequired("account-name")

	return cmd
}


