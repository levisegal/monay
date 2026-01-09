package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shopspring/decimal"
	"github.com/spf13/cobra"

	"github.com/levisegal/monay/services/holdings/config"
	"github.com/levisegal/monay/services/holdings/database"
	"github.com/levisegal/monay/services/holdings/gen/db"
)

func cashCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cash",
		Short: "Cash balance tracking commands",
	}

	cmd.AddCommand(cashSetCommand())
	cmd.AddCommand(cashBalanceCommand())
	cmd.AddCommand(cashLedgerCommand())
	cmd.AddCommand(cashGenerateCommand())

	return cmd
}

func cashSetCommand() *cobra.Command {
	var (
		accountName string
		dateStr     string
		balance     string
	)

	cmd := &cobra.Command{
		Use:   "set",
		Short: "Set opening cash balance for an account",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			cfg, err := config.Load()
			if err != nil {
				return err
			}

			date, err := time.Parse("2006-01-02", dateStr)
			if err != nil {
				return fmt.Errorf("invalid date format, use YYYY-MM-DD: %w", err)
			}

			balanceDec, err := decimal.NewFromString(balance)
			if err != nil {
				return fmt.Errorf("invalid balance: %w", err)
			}
			balanceMicros := balanceDec.Mul(decimal.NewFromInt(1_000_000)).IntPart()

			return setCashOpening(ctx, cfg, accountName, date, balanceMicros)
		},
	}

	cmd.Flags().StringVar(&accountName, "account-name", "", "Account name")
	cmd.Flags().StringVar(&dateStr, "date", "", "Opening balance date (YYYY-MM-DD)")
	cmd.Flags().StringVar(&balance, "balance", "", "Opening cash balance (e.g., 5000.00)")

	cmd.MarkFlagRequired("account-name")
	cmd.MarkFlagRequired("date")
	cmd.MarkFlagRequired("balance")

	return cmd
}

func cashBalanceCommand() *cobra.Command {
	var accountName string

	cmd := &cobra.Command{
		Use:   "balance",
		Short: "Show current cash balance for an account",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			cfg, err := config.Load()
			if err != nil {
				return err
			}

			return showCashBalance(ctx, cfg, accountName)
		},
	}

	cmd.Flags().StringVar(&accountName, "account-name", "", "Account name")
	cmd.MarkFlagRequired("account-name")

	return cmd
}

func cashLedgerCommand() *cobra.Command {
	var (
		accountName string
		year        int
	)

	cmd := &cobra.Command{
		Use:   "ledger",
		Short: "Show cash ledger for an account",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			cfg, err := config.Load()
			if err != nil {
				return err
			}

			return showCashLedger(ctx, cfg, accountName, year)
		},
	}

	cmd.Flags().StringVar(&accountName, "account-name", "", "Account name")
	cmd.Flags().IntVar(&year, "year", 0, "Filter by year (optional)")
	cmd.MarkFlagRequired("account-name")

	return cmd
}

func cashGenerateCommand() *cobra.Command {
	var accountName string

	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate cash transactions from existing transactions",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			cfg, err := config.Load()
			if err != nil {
				return err
			}

			return generateCashTransactions(ctx, cfg, accountName)
		},
	}

	cmd.Flags().StringVar(&accountName, "account-name", "", "Account name")
	cmd.MarkFlagRequired("account-name")

	return cmd
}

func setCashOpening(ctx context.Context, cfg *config.Config, accountName string, date time.Time, balanceMicros int64) error {
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

	existing, err := queries.GetOpeningCashBalance(ctx, account.ID)
	if err == nil {
		slog.Info("replacing existing opening balance",
			"old_date", existing.TransactionDate.Time.Format("2006-01-02"),
			"old_balance", formatMicros(existing.AmountMicros),
		)
		if err := queries.DeleteCashTransactionsByAccount(ctx, account.ID); err != nil {
			return fmt.Errorf("failed to delete existing cash transactions: %w", err)
		}
	}

	err = queries.CreateCashTransaction(ctx, db.CreateCashTransactionParams{
		ID:              database.NewID(database.PrefixCashTxn),
		AccountID:       account.ID,
		TransactionDate: pgtype.Date{Time: date, Valid: true},
		CashType:        "opening",
		AmountMicros:    balanceMicros,
		Description:     pgtype.Text{String: "Opening cash balance", Valid: true},
	})
	if err != nil {
		return fmt.Errorf("failed to create opening balance: %w", err)
	}

	slog.Info("set opening cash balance",
		"account", accountName,
		"date", date.Format("2006-01-02"),
		"balance", formatMicros(balanceMicros),
	)

	return nil
}

func showCashBalance(ctx context.Context, cfg *config.Config, accountName string) error {
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

	balanceMicros, err := queries.GetCashBalance(ctx, account.ID)
	if err != nil {
		return fmt.Errorf("failed to get cash balance: %w", err)
	}

	fmt.Printf("Account: %s\n", accountName)
	fmt.Printf("Cash Balance: %s\n", formatMicros(balanceMicros))

	return nil
}

func showCashLedger(ctx context.Context, cfg *config.Config, accountName string, year int) error {
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

	var rows []db.ListCashTransactionsRow

	if year > 0 {
		startDate := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
		endDate := time.Date(year, 12, 31, 0, 0, 0, 0, time.UTC)

		rangeRows, err := queries.ListCashTransactionsByDateRange(ctx, db.ListCashTransactionsByDateRangeParams{
			AccountID: account.ID,
			StartDate: pgtype.Date{Time: startDate, Valid: true},
			EndDate:   pgtype.Date{Time: endDate, Valid: true},
		})
		if err != nil {
			return fmt.Errorf("failed to list cash transactions: %w", err)
		}
		for _, r := range rangeRows {
			rows = append(rows, db.ListCashTransactionsRow(r))
		}
	} else {
		rows, err = queries.ListCashTransactions(ctx, account.ID)
		if err != nil {
			return fmt.Errorf("failed to list cash transactions: %w", err)
		}
	}

	if len(rows) == 0 {
		fmt.Printf("No cash transactions found for %s\n", accountName)
		return nil
	}

	fmt.Printf("=== %s: Cash Ledger ===\n\n", accountName)
	fmt.Printf("%-12s %-15s %15s  %s\n", "Date", "Type", "Amount", "Description")
	fmt.Println("─────────────────────────────────────────────────────────────────────────")

	var totalIncome, totalExpenses int64
	for _, row := range rows {
		symbol := ""
		if row.Symbol.Valid {
			symbol = row.Symbol.String
		}
		desc := row.Description.String
		if symbol != "" && desc == "" {
			desc = symbol
		} else if symbol != "" {
			desc = symbol + ": " + desc
		}
		if len(desc) > 40 {
			desc = desc[:37] + "..."
		}

		fmt.Printf("%-12s %-15s %15s  %s\n",
			row.TransactionDate.Time.Format("2006-01-02"),
			row.CashType,
			formatMicros(row.AmountMicros),
			desc,
		)

		if row.AmountMicros > 0 && row.CashType != "opening" {
			totalIncome += row.AmountMicros
		} else if row.AmountMicros < 0 {
			totalExpenses += row.AmountMicros
		}
	}

	fmt.Println("─────────────────────────────────────────────────────────────────────────")
	fmt.Printf("Total Income:   %s\n", formatMicros(totalIncome))
	fmt.Printf("Total Expenses: %s\n", formatMicros(totalExpenses))

	balanceMicros, _ := queries.GetCashBalance(ctx, account.ID)
	fmt.Printf("Current Balance: %s\n", formatMicros(balanceMicros))

	return nil
}

func generateCashTransactions(ctx context.Context, cfg *config.Config, accountName string) error {
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

	if err := queries.DeleteNonOpeningCashTransactionsByAccount(ctx, account.ID); err != nil {
		return fmt.Errorf("failed to clear existing cash transactions: %w", err)
	}

	transactions, err := queries.ListTransactionsByAccount(ctx, account.ID)
	if err != nil {
		return fmt.Errorf("failed to list transactions: %w", err)
	}

	var created int
	for _, txn := range transactions {
		cashType, hasCashImpact := mapTransactionTypeToCashType(txn.TransactionType)
		if !hasCashImpact {
			continue
		}

		amountMicros := normalizeCashAmount(cashType, txn.AmountMicros)

		err := queries.CreateCashTransaction(ctx, db.CreateCashTransactionParams{
			ID:              database.NewID(database.PrefixCashTxn),
			AccountID:       account.ID,
			TransactionID:   pgtype.Text{String: txn.ID, Valid: true},
			TransactionDate: txn.TransactionDate,
			CashType:        cashType,
			AmountMicros:    amountMicros,
			SecurityID:      txn.SecurityID,
			Description:     txn.Description,
		})
		if err != nil {
			return fmt.Errorf("failed to create cash transaction: %w", err)
		}
		created++
	}

	slog.Info("generated cash transactions",
		"account", accountName,
		"transactions", len(transactions),
		"cash_records", created,
	)

	return nil
}

func mapTransactionTypeToCashType(txnType string) (cashType string, hasCashImpact bool) {
	switch txnType {
	case "buy":
		return "purchase", true
	case "sell":
		return "proceeds", true
	case "dividend":
		return "dividend", true
	case "interest":
		return "interest", true
	case "cap_gain":
		return "cap_gain", true
	case "fee":
		return "fee", true
	case "transfer_in":
		return "transfer_in", true
	case "transfer_out":
		return "transfer_out", true
	default:
		return "", false
	}
}

func normalizeCashAmount(cashType string, amountMicros int64) int64 {
	switch cashType {
	case "purchase":
		if amountMicros > 0 {
			return -amountMicros
		}
		return amountMicros
	case "fee":
		if amountMicros > 0 {
			return -amountMicros
		}
		return amountMicros
	case "transfer_out":
		if amountMicros > 0 {
			return -amountMicros
		}
		return amountMicros
	case "proceeds", "dividend", "interest", "cap_gain", "transfer_in":
		if amountMicros < 0 {
			return -amountMicros
		}
		return amountMicros
	default:
		return amountMicros
	}
}

func formatMicros(micros int64) string {
	d := decimal.NewFromInt(micros).Div(decimal.NewFromInt(1_000_000))
	if micros >= 0 {
		return fmt.Sprintf("$%s", d.StringFixed(2))
	}
	return fmt.Sprintf("-$%s", d.Abs().StringFixed(2))
}

