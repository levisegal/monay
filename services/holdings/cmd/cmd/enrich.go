package cmd

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/spf13/cobra"

	"github.com/levisegal/monay/services/holdings/config"
	"github.com/levisegal/monay/services/holdings/database"
	"github.com/levisegal/monay/services/holdings/gen/db"
)

func enrichCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "enrich",
		Short: "Enrich securities with expense ratios and fund metadata from yfinance",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			cfg, err := config.Load()
			if err != nil {
				return err
			}

			conn, err := database.Open(ctx, cfg.DBPath)
			if err != nil {
				return err
			}
			defer conn.Close()

			queries := db.New(conn)
			return runEnrich(ctx, queries, cfg.PortfolioURL)
		},
	}
	return cmd
}

func runEnrich(ctx context.Context, queries *db.Queries, portfolioURL string) error {
	securities, err := queries.ListSecuritiesWithOpenLots(ctx)
	if err != nil {
		return fmt.Errorf("failed to list securities: %w", err)
	}

	symbols := make([]string, 0, len(securities))
	for _, s := range securities {
		symbols = append(symbols, s.Symbol)
	}

	slog.Info("fetching metadata", "symbols", len(symbols))
	quotes := fetchQuotes(ctx, portfolioURL, symbols, nil)

	updated := 0
	for _, s := range securities {
		qd, ok := quotes[s.Symbol]
		if !ok {
			continue
		}

		var erBps sql.NullInt64
		if qd.NetExpenseRatio > 0 {
			erBps = sql.NullInt64{Int64: int64(qd.NetExpenseRatio * 100), Valid: true}
		}

		var family sql.NullString
		if qd.FundFamily != "" {
			family = sql.NullString{String: qd.FundFamily, Valid: true}
		}

		var category sql.NullString
		if qd.Category != "" {
			category = sql.NullString{String: qd.Category, Valid: true}
		}

		if !erBps.Valid && !family.Valid && !category.Valid {
			continue
		}

		err := queries.UpdateSecurityMetadata(ctx, db.UpdateSecurityMetadataParams{
			ExpenseRatioBps: erBps,
			FundFamily:      family,
			FundCategory:    category,
			Symbol:          s.Symbol,
		})
		if err != nil {
			slog.Warn("failed to update", "symbol", s.Symbol, "error", err)
			continue
		}

		bps := ""
		if erBps.Valid {
			bps = fmt.Sprintf("%d bps", erBps.Int64)
		}
		fam := family.String
		cat := category.String
		slog.Info("enriched", "symbol", s.Symbol, "expense_ratio", bps, "fund_family", fam, "category", cat)
		updated++
	}

	slog.Info("enrichment complete", "updated", updated, "total", len(securities))
	return nil
}
