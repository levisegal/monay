package cmd

import (
	"context"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/rodaine/table"
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

			fmt.Printf("\n=== %s: Current Holdings ===\n\n", account.Name)

			tbl := table.New("Symbol", "Quantity", "Cost Basis", "Acquired")
			tbl.WithWriter(os.Stdout)

			var totalCostBasis int64
			for _, h := range holdings {
				qty := float64(h.QuantityMicros) / 1_000_000
				cost := float64(h.CostBasisMicros) / 1_000_000
				totalCostBasis += h.CostBasisMicros

				acquired := ""
				if t, ok := h.EarliestAcquired.(time.Time); ok {
					acquired = t.Format("2006-01-02")
				}

				tbl.AddRow(h.Symbol, formatQty(qty), formatCurrency(cost), acquired)
			}

			tbl.Print()

			fmt.Printf("\nTOTAL: %s\n", formatCurrency(float64(totalCostBasis)/1_000_000))

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

	switch sortBy {
	case "symbol":
		sort.Slice(holdings, func(i, j int) bool {
			return holdings[i].Symbol < holdings[j].Symbol
		})
	case "account":
		sort.Slice(holdings, func(i, j int) bool {
			return holdings[i].AccountName < holdings[j].AccountName
		})
	default:
		sort.Slice(holdings, func(i, j int) bool {
			return holdings[i].CostBasisMicros > holdings[j].CostBasisMicros
		})
	}

	fmt.Printf("\n=== All Holdings ===\n\n")

	tbl := table.New("Broker", "Account", "Symbol", "Quantity", "Cost Basis", "Acquired")
	tbl.WithWriter(os.Stdout)

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

		tbl.AddRow(broker, h.AccountName, h.Symbol, formatQty(qty), formatCurrency(cost), acquired)
	}

	tbl.Print()

	fmt.Printf("\nTOTAL: %s\n", formatCurrency(float64(totalCostBasis)/1_000_000))

	return nil
}

func formatCurrency(amount float64) string {
	p := message.NewPrinter(language.English)
	return p.Sprintf("$%.2f", amount)
}

func formatQty(qty float64) string {
	p := message.NewPrinter(language.English)
	return p.Sprintf("%.2f", qty)
}
