package cmd

import (
	"bufio"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shopspring/decimal"
	"github.com/spf13/cobra"

	"github.com/levisegal/monay/services/holdings/config"
	"github.com/levisegal/monay/services/holdings/database"
	"github.com/levisegal/monay/services/holdings/gen/db"
	"github.com/levisegal/monay/services/holdings/importer"
	"github.com/levisegal/monay/services/holdings/taxlots"
)

func lotsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "lots",
		Short: "Tax lot management commands",
	}

	cmd.AddCommand(processLotsCommand())
	cmd.AddCommand(createLotsCommand())
	cmd.AddCommand(checkLotsCommand())
	cmd.AddCommand(clearLotsCommand())

	return cmd
}

func clearLotsCommand() *cobra.Command {
	var accountName string

	cmd := &cobra.Command{
		Use:   "clear",
		Short: "Clear all transactions and lots for an account",
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

			if err := queries.DeleteLotsByAccount(ctx, account.ID); err != nil {
				return fmt.Errorf("failed to delete lots: %w", err)
			}

			if err := queries.DeleteTransactionsByAccount(ctx, account.ID); err != nil {
				return fmt.Errorf("failed to delete transactions: %w", err)
			}

			slog.Info("cleared account data", "account", account.Name)
			return nil
		},
	}

	cmd.Flags().StringVar(&accountName, "account-name", "", "Account name to clear")
	cmd.MarkFlagRequired("account-name")

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

func createLotsCommand() *cobra.Command {
	var accountName string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Interactively create opening balance lots",
		Long: `Create opening balance transactions for positions acquired before your transaction history.
		
This will prompt you for symbol, quantity, cost basis, and acquisition date for each lot.`,
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

			scanner := bufio.NewScanner(os.Stdin)

			for {
				lot, err := promptForLot(scanner)
				if err != nil {
					return err
				}
				if lot == nil {
					break
				}

				sec, err := queries.UpsertSecurity(ctx, db.UpsertSecurityParams{
					ID:     database.NewID(database.PrefixSecurity),
					Symbol: lot.symbol,
					Name:   pgtype.Text{},
				})
				if err != nil {
					return fmt.Errorf("failed to upsert security: %w", err)
				}

				txnID := database.NewID(database.PrefixTransaction)
				err = queries.CreateTransaction(ctx, db.CreateTransactionParams{
					ID:              txnID,
					AccountID:       account.ID,
					SecurityID:      pgtype.Text{String: sec.ID, Valid: true},
					TransactionType: string(importer.TransactionTypeOpeningBalance),
					TransactionDate: pgtype.Date{Time: lot.acquiredDate, Valid: true},
					QuantityMicros:  pgtype.Int8{Int64: lot.quantityMicros, Valid: true},
					PriceMicros:     pgtype.Int8{Int64: lot.costBasisMicros / lot.quantityMicros, Valid: true},
					AmountMicros:    lot.costBasisMicros,
					FeesMicros:      pgtype.Int8{Int64: 0, Valid: true},
					Description:     pgtype.Text{String: "Opening balance - manual entry", Valid: true},
				})
				if err != nil {
					return fmt.Errorf("failed to create transaction: %w", err)
				}

				slog.Info("created opening balance",
					"symbol", lot.symbol,
					"quantity", float64(lot.quantityMicros)/1_000_000,
					"cost_basis", float64(lot.costBasisMicros)/1_000_000,
					"acquired", lot.acquiredDate.Format("2006-01-02"),
				)

				fmt.Print("\nAdd another lot? (y/n): ")
				if !scanner.Scan() {
					fmt.Println("\nCancelled.")
					break
				}
				if strings.ToLower(strings.TrimSpace(scanner.Text())) != "y" {
					break
				}
				fmt.Println()
			}

			fmt.Println("\nDone. Run 'lots process' to rebuild tax lots.")
			return nil
		},
	}

	cmd.Flags().StringVar(&accountName, "account-name", "", "Account name")
	cmd.MarkFlagRequired("account-name")

	return cmd
}

type lotInput struct {
	symbol          string
	quantityMicros  int64
	costBasisMicros int64
	acquiredDate    time.Time
}

func checkLotsCommand() *cobra.Command {
	var accountName string
	var fix bool

	cmd := &cobra.Command{
		Use:   "check",
		Short: "Analyze lot gaps and identify positions needing attention",
		Long: `Analyze transactions to find sells without matching buys.

Categorizes gaps as:
- SAFE TO IGNORE: fully sold positions, no current holdings affected
- NEEDS REVIEW: still held positions with missing cost basis

Use --fix to interactively add opening balances for positions needing review.`,
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

			analyzer := taxlots.NewAnalyzer(queries)
			result, err := analyzer.Analyze(ctx, account.ID)
			if err != nil {
				return fmt.Errorf("analysis failed: %w", err)
			}

			var safeToIgnore, needsReview []taxlots.SymbolGap
			for _, gap := range result.Gaps {
				if gap.NeedsOpeningBalance {
					needsReview = append(needsReview, gap)
				} else {
					safeToIgnore = append(safeToIgnore, gap)
				}
			}

			fmt.Printf("\n=== %s: Lot Analysis ===\n", account.Name)

			if len(safeToIgnore) > 0 {
				fmt.Println("\nSAFE TO IGNORE (old positions fully closed, or new buys after old sells):")
				fmt.Printf("  %-10s %15s %15s\n", "Symbol", "Unmatched Qty", "Current Qty")
				for _, gap := range safeToIgnore {
					fmt.Printf("  %-10s %15.2f %15.2f\n", gap.Symbol,
						float64(gap.UnmatchedMicros)/1_000_000,
						float64(gap.RemainingMicros)/1_000_000)
				}
			}

			if len(needsReview) > 0 {
				fmt.Println("\nNEEDS OPENING BALANCE (current holdings affected by missing buys):")
				fmt.Printf("  %-10s %15s %15s\n", "Symbol", "Unmatched Qty", "Current Qty")
				for _, gap := range needsReview {
					fmt.Printf("  %-10s %15.2f %15.2f\n", gap.Symbol,
						float64(gap.UnmatchedMicros)/1_000_000,
						float64(gap.RemainingMicros)/1_000_000)
				}
			}

			fmt.Printf("\nSummary: %d historical gaps (ignorable), %d need opening balances\n",
				len(safeToIgnore), len(needsReview))

			if len(needsReview) == 0 {
				fmt.Println("No action needed.")
				return nil
			}

			if !fix {
				fmt.Println("Run with --fix to add opening balances interactively.")
				return nil
			}

			scanner := bufio.NewScanner(os.Stdin)
			fmt.Print("\nFix these now? (y/n): ")
			if !scanner.Scan() {
				fmt.Println("\nCancelled.")
				return nil
			}
			if strings.ToLower(strings.TrimSpace(scanner.Text())) != "y" {
				return nil
			}

			for _, gap := range needsReview {
				fmt.Printf("\n--- %s (%.2f shares missing) ---\n", gap.Symbol, float64(gap.UnmatchedMicros)/1_000_000)
				fmt.Println("Look up in E*Trade: Portfolios > Positions > " + gap.Symbol + " > Date Acquired, Total Cost")

				lot, err := promptForLotWithSymbol(scanner, gap.Symbol, gap.UnmatchedMicros)
				if err != nil {
					return err
				}
				if lot == nil {
					fmt.Println("Skipped.")
					continue
				}

				sec, err := queries.UpsertSecurity(ctx, db.UpsertSecurityParams{
					ID:     database.NewID(database.PrefixSecurity),
					Symbol: lot.symbol,
					Name:   pgtype.Text{},
				})
				if err != nil {
					return fmt.Errorf("failed to upsert security: %w", err)
				}

				err = queries.CreateTransaction(ctx, db.CreateTransactionParams{
					ID:              database.NewID(database.PrefixTransaction),
					AccountID:       account.ID,
					SecurityID:      pgtype.Text{String: sec.ID, Valid: true},
					TransactionType: string(importer.TransactionTypeOpeningBalance),
					TransactionDate: pgtype.Date{Time: lot.acquiredDate, Valid: true},
					QuantityMicros:  pgtype.Int8{Int64: lot.quantityMicros, Valid: true},
					PriceMicros:     pgtype.Int8{Int64: lot.costBasisMicros / lot.quantityMicros, Valid: true},
					AmountMicros:    lot.costBasisMicros,
					FeesMicros:      pgtype.Int8{Int64: 0, Valid: true},
					Description:     pgtype.Text{String: "Opening balance - manual entry", Valid: true},
				})
				if err != nil {
					return fmt.Errorf("failed to create transaction: %w", err)
				}

				fmt.Printf("Created opening balance for %s.\n", lot.symbol)
			}

			fmt.Printf("\nDone. Run 'lots process --account-name %s' to rebuild lots.\n", account.Name)
			return nil
		},
	}

	cmd.Flags().StringVar(&accountName, "account-name", "", "Account name to check")
	cmd.Flags().BoolVar(&fix, "fix", false, "Interactively add opening balances for positions needing review")
	cmd.MarkFlagRequired("account-name")

	return cmd
}

func promptForLotWithSymbol(scanner *bufio.Scanner, symbol string, suggestedQtyMicros int64) (*lotInput, error) {
	fmt.Printf("Quantity (shares) [%.2f]: ", float64(suggestedQtyMicros)/1_000_000)
	if !scanner.Scan() {
		return nil, scanner.Err()
	}
	qtyStr := strings.TrimSpace(scanner.Text())
	var quantityMicros int64
	if qtyStr == "" {
		quantityMicros = suggestedQtyMicros
	} else {
		qtyDec, err := decimal.NewFromString(qtyStr)
		if err != nil {
			return nil, fmt.Errorf("invalid quantity: %w", err)
		}
		quantityMicros = qtyDec.Mul(decimal.NewFromInt(1_000_000)).IntPart()
	}

	fmt.Print("Total cost basis ($): ")
	if !scanner.Scan() {
		return nil, scanner.Err()
	}
	costStr := strings.TrimSpace(scanner.Text())
	if costStr == "" {
		return nil, nil
	}
	costDec, err := decimal.NewFromString(costStr)
	if err != nil {
		return nil, fmt.Errorf("invalid cost basis: %w", err)
	}
	costBasisMicros := costDec.Mul(decimal.NewFromInt(1_000_000)).IntPart()

	fmt.Print("Acquisition date (YYYY-MM-DD): ")
	if !scanner.Scan() {
		return nil, scanner.Err()
	}
	dateStr := strings.TrimSpace(scanner.Text())
	if dateStr == "" {
		return nil, nil
	}
	acquiredDate, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return nil, fmt.Errorf("invalid date (use YYYY-MM-DD): %w", err)
	}

	return &lotInput{
		symbol:          symbol,
		quantityMicros:  quantityMicros,
		costBasisMicros: costBasisMicros,
		acquiredDate:    acquiredDate,
	}, nil
}

func promptForLot(scanner *bufio.Scanner) (*lotInput, error) {
	fmt.Print("Symbol (or 'done' to finish): ")
	if !scanner.Scan() {
		return nil, scanner.Err()
	}
	symbol := strings.ToUpper(strings.TrimSpace(scanner.Text()))
	if symbol == "" || symbol == "DONE" {
		return nil, nil
	}

	fmt.Print("Quantity (shares): ")
	if !scanner.Scan() {
		return nil, scanner.Err()
	}
	qtyDec, err := decimal.NewFromString(strings.TrimSpace(scanner.Text()))
	if err != nil {
		return nil, fmt.Errorf("invalid quantity: %w", err)
	}
	quantityMicros := qtyDec.Mul(decimal.NewFromInt(1_000_000)).IntPart()

	fmt.Print("Total cost basis ($): ")
	if !scanner.Scan() {
		return nil, scanner.Err()
	}
	costDec, err := decimal.NewFromString(strings.TrimSpace(scanner.Text()))
	if err != nil {
		return nil, fmt.Errorf("invalid cost basis: %w", err)
	}
	costBasisMicros := costDec.Mul(decimal.NewFromInt(1_000_000)).IntPart()

	fmt.Print("Acquisition date (YYYY-MM-DD): ")
	if !scanner.Scan() {
		return nil, scanner.Err()
	}
	dateStr := strings.TrimSpace(scanner.Text())
	acquiredDate, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return nil, fmt.Errorf("invalid date (use YYYY-MM-DD): %w", err)
	}

	return &lotInput{
		symbol:          symbol,
		quantityMicros:  quantityMicros,
		costBasisMicros: costBasisMicros,
		acquiredDate:    acquiredDate,
	}, nil
}
