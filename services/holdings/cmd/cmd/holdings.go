package cmd

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/text/language"
	"golang.org/x/text/message"

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
	var all bool
	var sortBy string

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

			if all {
				return listAllHoldings(queries, ctx, sortBy)
			}

			if accountName == "" {
				return fmt.Errorf("--account-name required (or use --all)")
			}

			account, err := queries.GetAccountByName(ctx, accountName)
			if err != nil {
				return fmt.Errorf("account not found: %s", accountName)
			}

			holdings, err := queries.ListHoldingsByAccount(ctx, account.ID)
			if err != nil {
				return fmt.Errorf("failed to list holdings: %w", err)
			}

			p := message.NewPrinter(language.English)

			fmt.Printf("\n=== %s: Current Holdings ===\n\n", account.Name)
			fmt.Printf("  %-8s %14s %16s %12s\n", "Symbol", "Quantity", "Cost Basis", "Acquired")
			fmt.Printf("  %-8s %14s %16s %12s\n", "------", "--------", "----------", "--------")

			var totalCostBasis int64
			for _, h := range holdings {
				qty := float64(h.QuantityMicros) / 1_000_000
				cost := float64(h.CostBasisMicros) / 1_000_000
				totalCostBasis += h.CostBasisMicros

				acquired := ""
				if t, ok := h.EarliestAcquired.(time.Time); ok {
					acquired = t.Format("2006-01-02")
				}

				p.Printf("  %-8s %14.2f %16s %12s\n", h.Symbol, qty, formatCurrency(cost), acquired)
			}

			p.Printf("\n  %-8s %14s %16s\n", "TOTAL", "", formatCurrency(float64(totalCostBasis)/1_000_000))

			return nil
		},
	}

	cmd.Flags().StringVar(&accountName, "account-name", "", "Account name")
	cmd.Flags().BoolVar(&all, "all", false, "Show holdings across all accounts")
	cmd.Flags().StringVar(&sortBy, "sort", "cost", "Sort by: cost, symbol, account")

	return cmd
}

func listAllHoldings(queries *db.Queries, ctx context.Context, sortBy string) error {
	holdings, err := queries.ListAllHoldings(ctx)
	if err != nil {
		return fmt.Errorf("failed to list holdings: %w", err)
	}

	// Sort based on flag
	switch sortBy {
	case "symbol":
		sort.Slice(holdings, func(i, j int) bool {
			return holdings[i].Symbol < holdings[j].Symbol
		})
	case "account":
		sort.Slice(holdings, func(i, j int) bool {
			return holdings[i].AccountName < holdings[j].AccountName
		})
	default: // cost (descending)
		sort.Slice(holdings, func(i, j int) bool {
			return holdings[i].CostBasisMicros > holdings[j].CostBasisMicros
		})
	}

	p := message.NewPrinter(language.English)

	fmt.Printf("\n=== All Holdings ===\n\n")
	fmt.Printf("  %-10s %-12s %-8s %14s %16s %12s\n", "Broker", "Account", "Symbol", "Quantity", "Cost Basis", "Acquired")
	fmt.Printf("  %-10s %-12s %-8s %14s %16s %12s\n", "------", "-------", "------", "--------", "----------", "--------")

	var totalCostBasis int64
	for _, h := range holdings {
		qty := float64(h.QuantityMicros) / 1_000_000
		cost := float64(h.CostBasisMicros) / 1_000_000
		totalCostBasis += h.CostBasisMicros

		acquired := ""
		if t, ok := h.EarliestAcquired.(time.Time); ok {
			acquired = t.Format("2006-01-02")
		}

		broker := h.Broker
		if broker == "" {
			broker = "-"
		}

		p.Printf("  %-10s %-12s %-8s %14.2f %16s %12s\n", broker, h.AccountName, h.Symbol, qty, formatCurrency(cost), acquired)
	}

	p.Printf("\n  %-10s %-12s %-8s %14s %16s\n", "", "TOTAL", "", "", formatCurrency(float64(totalCostBasis)/1_000_000))

	return nil
}

func formatCurrency(amount float64) string {
	p := message.NewPrinter(language.English)
	return p.Sprintf("$%.2f", amount)
}
