package cmd

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"

	"github.com/spf13/cobra"

	"github.com/levisegal/monay/services/holdings/config"
	"github.com/levisegal/monay/services/holdings/database"
	"github.com/levisegal/monay/services/holdings/gen/db"
	"github.com/levisegal/monay/services/holdings/importer"
)

func importCommand() *cobra.Command {
	var (
		broker      string
		files       []string
		accountName string
	)

	cmd := &cobra.Command{
		Use:   "import",
		Short: "Import transactions from CSV files",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			cfg, err := config.Load()
			if err != nil {
				return err
			}

			for _, file := range files {
				if err := runImport(ctx, cfg, broker, file, accountName); err != nil {
					return err
				}
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&broker, "broker", "", "Broker name (etrade, schwab, fidelity, vanguard, lpl)")
	cmd.Flags().StringArrayVar(&files, "file", nil, "Path to CSV file(s) - can be repeated")
	cmd.Flags().StringVar(&accountName, "account-name", "", "Account name for imported data")

	cmd.MarkFlagRequired("broker")
	cmd.MarkFlagRequired("file")
	cmd.MarkFlagRequired("account-name")

	return cmd
}

func runImport(ctx context.Context, cfg *config.Config, brokerName, filePath, accountName string) error {
	parser, err := importer.GetParser(importer.Broker(brokerName))
	if err != nil {
		return err
	}

	f, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	result, err := parser.Parse(ctx, f)
	if err != nil {
		return fmt.Errorf("failed to parse CSV: %w", err)
	}

	slog.Info("parsed CSV",
		"transactions", len(result.Transactions),
		"positions", len(result.Positions),
	)

	conn, err := database.Open(ctx, cfg.DBPath)
	if err != nil {
		return err
	}
	defer conn.Close()

	queries := db.New(conn)

	account, err := queries.GetAccountByName(ctx, accountName)
	if err != nil {
		account, err = queries.CreateAccount(ctx, db.CreateAccountParams{
			ID:              database.NewID(database.PrefixAccount),
			Name:            accountName,
			InstitutionName: brokerName,
			AccountType:     "brokerage",
		})
		if err != nil {
			return fmt.Errorf("failed to create account: %w", err)
		}
		slog.Info("created account", "id", account.ID, "name", account.Name)
	}

	for _, txn := range result.Transactions {
		var securityID sql.NullString

		if txn.Symbol != "" {
			sec, err := queries.UpsertSecurity(ctx, db.UpsertSecurityParams{
				ID:     database.NewID(database.PrefixSecurity),
				Symbol: txn.Symbol,
				Name:   sql.NullString{String: txn.SecurityName, Valid: txn.SecurityName != ""},
			})
			if err != nil {
				return fmt.Errorf("failed to upsert security %s: %w", txn.Symbol, err)
			}
			securityID = sql.NullString{String: sec.ID, Valid: true}
		}

		err = queries.CreateTransaction(ctx, db.CreateTransactionParams{
			ID:              database.NewID(database.PrefixTransaction),
			AccountID:       account.ID,
			SecurityID:      securityID,
			TransactionType: string(txn.TransactionType),
			TransactionDate: txn.TransactionDate.Format("2006-01-02"),
			QuantityMicros:  sql.NullInt64{Int64: txn.QuantityMicros, Valid: true},
			PriceMicros:     sql.NullInt64{Int64: txn.PriceMicros, Valid: true},
			AmountMicros:    txn.AmountMicros,
			FeesMicros:      sql.NullInt64{Int64: txn.FeesMicros, Valid: true},
			Description:     sql.NullString{String: txn.Description, Valid: txn.Description != ""},
		})
		if err != nil {
			return fmt.Errorf("failed to create transaction: %w", err)
		}
	}

	for _, pos := range result.Positions {
		sec, err := queries.UpsertSecurity(ctx, db.UpsertSecurityParams{
			ID:     database.NewID(database.PrefixSecurity),
			Symbol: pos.Symbol,
			Name:   sql.NullString{String: pos.SecurityName, Valid: pos.SecurityName != ""},
		})
		if err != nil {
			return fmt.Errorf("failed to upsert security %s: %w", pos.Symbol, err)
		}

		_, err = queries.UpsertPosition(ctx, db.UpsertPositionParams{
			ID:                database.NewID(database.PrefixPosition),
			AccountID:         account.ID,
			SecurityID:        sec.ID,
			QuantityMicros:    pos.QuantityMicros,
			CostBasisMicros:   sql.NullInt64{Int64: pos.CostBasisMicros, Valid: true},
			MarketValueMicros: sql.NullInt64{Int64: pos.MarketValueMicros, Valid: true},
			AsOfDate:          pos.AsOfDate.Format("2006-01-02"),
		})
		if err != nil {
			return fmt.Errorf("failed to upsert position: %w", err)
		}
	}

	slog.Info("import complete",
		"account", accountName,
		"transactions", len(result.Transactions),
		"positions", len(result.Positions),
	)

	return nil
}
