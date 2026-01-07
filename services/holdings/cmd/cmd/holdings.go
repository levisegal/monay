package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/levisegal/monay/services/holdings/config"
	"github.com/levisegal/monay/services/holdings/database"
	"github.com/levisegal/monay/services/holdings/gen/db"
)

func holdingsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "holdings",
		Short: "View current holdings",
	}

	cmd.AddCommand(listHoldingsCommand())

	return cmd
}

func listHoldingsCommand() *cobra.Command {
	var accountName string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List current holdings with cost basis",
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

			holdings, err := queries.ListHoldingsByAccount(ctx, account.ID)
			if err != nil {
				return fmt.Errorf("failed to list holdings: %w", err)
			}

			fmt.Printf("\n=== %s: Current Holdings ===\n\n", account.Name)
			fmt.Printf("  %-10s %12s %15s %12s\n", "Symbol", "Quantity", "Cost Basis", "Acquired")
			fmt.Printf("  %-10s %12s %15s %12s\n", "------", "--------", "----------", "--------")

			var totalCostBasis int64
			for _, h := range holdings {
				qty := float64(h.QuantityMicros) / 1_000_000
				cost := float64(h.CostBasisMicros) / 1_000_000
				totalCostBasis += h.CostBasisMicros

				acquired := ""
				if t, ok := h.EarliestAcquired.(time.Time); ok {
					acquired = t.Format("2006-01-02")
				}

				fmt.Printf("  %-10s %12.2f %15.2f %12s\n", h.Symbol, qty, cost, acquired)
			}

			fmt.Printf("\n  %-10s %12s %15.2f\n", "TOTAL", "", float64(totalCostBasis)/1_000_000)

			return nil
		},
	}

	cmd.Flags().StringVar(&accountName, "account-name", "", "Account name")
	cmd.MarkFlagRequired("account-name")

	return cmd
}

