package database_test

import (
	"context"
	"database/sql"
	"os"
	"testing"

	"github.com/levisegal/monay/services/holdings/database"
	"github.com/levisegal/monay/services/holdings/gen/db"
)

func TestOpen(t *testing.T) {
	ctx := context.Background()

	tmpFile, err := os.CreateTemp("", "test-*.db")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	conn, err := database.Open(ctx, tmpFile.Name())
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer conn.Close()

	var count int
	err = conn.QueryRowContext(ctx, "SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='accounts'").Scan(&count)
	if err != nil {
		t.Fatalf("failed to query tables: %v", err)
	}
	if count != 1 {
		t.Errorf("expected accounts table to exist, got count=%d", count)
	}
}

func TestAccountsCRUD(t *testing.T) {
	ctx := context.Background()

	tmpFile, err := os.CreateTemp("", "test-*.db")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	conn, err := database.Open(ctx, tmpFile.Name())
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer conn.Close()

	queries := db.New(conn)

	t.Run("create account", func(t *testing.T) {
		acct, err := queries.CreateAccount(ctx, db.CreateAccountParams{
			ID:              database.NewID(database.PrefixAccount),
			Name:            "Test Brokerage",
			InstitutionName: "etrade",
			AccountType:     "brokerage",
		})
		if err != nil {
			t.Fatalf("failed to create account: %v", err)
		}
		if acct.Name != "Test Brokerage" {
			t.Errorf("expected name 'Test Brokerage', got %q", acct.Name)
		}
		if acct.InstitutionName != "etrade" {
			t.Errorf("expected institution 'etrade', got %q", acct.InstitutionName)
		}
	})

	t.Run("get account by name", func(t *testing.T) {
		acct, err := queries.GetAccountByName(ctx, "Test Brokerage")
		if err != nil {
			t.Fatalf("failed to get account: %v", err)
		}
		if acct.Name != "Test Brokerage" {
			t.Errorf("expected name 'Test Brokerage', got %q", acct.Name)
		}
	})

	t.Run("list accounts", func(t *testing.T) {
		accounts, err := queries.ListAccounts(ctx)
		if err != nil {
			t.Fatalf("failed to list accounts: %v", err)
		}
		if len(accounts) != 1 {
			t.Errorf("expected 1 account, got %d", len(accounts))
		}
	})

	t.Run("update account", func(t *testing.T) {
		acct, err := queries.GetAccountByName(ctx, "Test Brokerage")
		if err != nil {
			t.Fatalf("failed to get account: %v", err)
		}

		updated, err := queries.UpdateAccount(ctx, db.UpdateAccountParams{
			ID:   acct.ID,
			Name: "Updated Brokerage",
		})
		if err != nil {
			t.Fatalf("failed to update account: %v", err)
		}
		if updated.Name != "Updated Brokerage" {
			t.Errorf("expected name 'Updated Brokerage', got %q", updated.Name)
		}
	})

	t.Run("delete account", func(t *testing.T) {
		acct, err := queries.GetAccountByName(ctx, "Updated Brokerage")
		if err != nil {
			t.Fatalf("failed to get account: %v", err)
		}

		err = queries.DeleteAccount(ctx, acct.ID)
		if err != nil {
			t.Fatalf("failed to delete account: %v", err)
		}

		accounts, err := queries.ListAccounts(ctx)
		if err != nil {
			t.Fatalf("failed to list accounts: %v", err)
		}
		if len(accounts) != 0 {
			t.Errorf("expected 0 accounts after delete, got %d", len(accounts))
		}
	})
}

func TestSecuritiesCRUD(t *testing.T) {
	ctx := context.Background()

	tmpFile, err := os.CreateTemp("", "test-*.db")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	conn, err := database.Open(ctx, tmpFile.Name())
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer conn.Close()

	queries := db.New(conn)

	t.Run("upsert security", func(t *testing.T) {
		sec, err := queries.UpsertSecurity(ctx, db.UpsertSecurityParams{
			ID:           database.NewID(database.PrefixSecurity),
			Symbol:       "AAPL",
			Name:         sql.NullString{String: "Apple Inc.", Valid: true},
			SecurityType: sql.NullString{String: "equity", Valid: true},
		})
		if err != nil {
			t.Fatalf("failed to upsert security: %v", err)
		}
		if sec.Symbol != "AAPL" {
			t.Errorf("expected symbol 'AAPL', got %q", sec.Symbol)
		}
	})

	t.Run("get security by symbol", func(t *testing.T) {
		sec, err := queries.GetSecurityBySymbol(ctx, "AAPL")
		if err != nil {
			t.Fatalf("failed to get security: %v", err)
		}
		if sec.Name.String != "Apple Inc." {
			t.Errorf("expected name 'Apple Inc.', got %q", sec.Name.String)
		}
	})

	t.Run("upsert updates existing", func(t *testing.T) {
		sec, err := queries.UpsertSecurity(ctx, db.UpsertSecurityParams{
			ID:           database.NewID(database.PrefixSecurity),
			Symbol:       "AAPL",
			Name:         sql.NullString{String: "Apple Inc. (Updated)", Valid: true},
			SecurityType: sql.NullString{String: "equity", Valid: true},
		})
		if err != nil {
			t.Fatalf("failed to upsert security: %v", err)
		}
		if sec.Name.String != "Apple Inc. (Updated)" {
			t.Errorf("expected updated name, got %q", sec.Name.String)
		}

		securities, err := queries.ListSecurities(ctx)
		if err != nil {
			t.Fatalf("failed to list securities: %v", err)
		}
		if len(securities) != 1 {
			t.Errorf("expected 1 security after upsert, got %d", len(securities))
		}
	})
}

func TestTransactionsAndLots(t *testing.T) {
	ctx := context.Background()

	tmpFile, err := os.CreateTemp("", "test-*.db")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	conn, err := database.Open(ctx, tmpFile.Name())
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer conn.Close()

	queries := db.New(conn)

	acct, err := queries.CreateAccount(ctx, db.CreateAccountParams{
		ID:              database.NewID(database.PrefixAccount),
		Name:            "Test Account",
		InstitutionName: "test",
		AccountType:     "brokerage",
	})
	if err != nil {
		t.Fatalf("failed to create account: %v", err)
	}

	sec, err := queries.UpsertSecurity(ctx, db.UpsertSecurityParams{
		ID:     database.NewID(database.PrefixSecurity),
		Symbol: "MSFT",
		Name:   sql.NullString{String: "Microsoft Corp", Valid: true},
	})
	if err != nil {
		t.Fatalf("failed to create security: %v", err)
	}

	t.Run("create transaction", func(t *testing.T) {
		err := queries.CreateTransaction(ctx, db.CreateTransactionParams{
			ID:              database.NewID(database.PrefixTransaction),
			AccountID:       acct.ID,
			SecurityID:      sql.NullString{String: sec.ID, Valid: true},
			TransactionType: "buy",
			TransactionDate: "2024-01-15",
			QuantityMicros:  sql.NullInt64{Int64: 10_000_000, Valid: true},
			PriceMicros:     sql.NullInt64{Int64: 350_000_000, Valid: true},
			AmountMicros:    3500_000_000,
		})
		if err != nil {
			t.Fatalf("failed to create transaction: %v", err)
		}
	})

	t.Run("list transactions", func(t *testing.T) {
		txns, err := queries.ListTransactionsByAccount(ctx, acct.ID)
		if err != nil {
			t.Fatalf("failed to list transactions: %v", err)
		}
		if len(txns) != 1 {
			t.Errorf("expected 1 transaction, got %d", len(txns))
		}
		if txns[0].Symbol.String != "MSFT" {
			t.Errorf("expected symbol 'MSFT', got %q", txns[0].Symbol.String)
		}
	})

	t.Run("create and list lots", func(t *testing.T) {
		txns, _ := queries.ListTransactionsByAccount(ctx, acct.ID)
		txnID := txns[0].ID

		lot, err := queries.CreateLot(ctx, db.CreateLotParams{
			ID:              database.NewID(database.PrefixLot),
			AccountID:       acct.ID,
			SecurityID:      sec.ID,
			TransactionID:   txnID,
			AcquiredDate:    "2024-01-15",
			QuantityMicros:  10_000_000,
			RemainingMicros: 10_000_000,
			CostBasisMicros: 3500_000_000,
		})
		if err != nil {
			t.Fatalf("failed to create lot: %v", err)
		}
		if lot.QuantityMicros != 10_000_000 {
			t.Errorf("expected quantity 10_000_000, got %d", lot.QuantityMicros)
		}

		lots, err := queries.ListLotsByAccount(ctx, acct.ID)
		if err != nil {
			t.Fatalf("failed to list lots: %v", err)
		}
		if len(lots) != 1 {
			t.Errorf("expected 1 lot, got %d", len(lots))
		}
	})
}

func TestCashTransactions(t *testing.T) {
	ctx := context.Background()

	tmpFile, err := os.CreateTemp("", "test-*.db")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	conn, err := database.Open(ctx, tmpFile.Name())
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer conn.Close()

	queries := db.New(conn)

	acct, err := queries.CreateAccount(ctx, db.CreateAccountParams{
		ID:              database.NewID(database.PrefixAccount),
		Name:            "Cash Test Account",
		InstitutionName: "test",
		AccountType:     "brokerage",
	})
	if err != nil {
		t.Fatalf("failed to create account: %v", err)
	}

	t.Run("create cash transaction", func(t *testing.T) {
		err := queries.CreateCashTransaction(ctx, db.CreateCashTransactionParams{
			ID:              database.NewID(database.PrefixCashTxn),
			AccountID:       acct.ID,
			TransactionDate: "2024-01-01",
			CashType:        "opening",
			AmountMicros:    5000_000_000,
			Description:     sql.NullString{String: "Opening balance", Valid: true},
		})
		if err != nil {
			t.Fatalf("failed to create cash transaction: %v", err)
		}
	})

	t.Run("get cash balance", func(t *testing.T) {
		balance, err := queries.GetCashBalance(ctx, acct.ID)
		if err != nil {
			t.Fatalf("failed to get cash balance: %v", err)
		}
		expected := int64(5000_000_000)
		if balance != expected {
			t.Errorf("expected balance %d, got %d", expected, balance)
		}
	})

	t.Run("add dividend and check balance", func(t *testing.T) {
		err := queries.CreateCashTransaction(ctx, db.CreateCashTransactionParams{
			ID:              database.NewID(database.PrefixCashTxn),
			AccountID:       acct.ID,
			TransactionDate: "2024-03-15",
			CashType:        "dividend",
			AmountMicros:    50_000_000,
			Description:     sql.NullString{String: "AAPL dividend", Valid: true},
		})
		if err != nil {
			t.Fatalf("failed to create dividend: %v", err)
		}

		balance, err := queries.GetCashBalance(ctx, acct.ID)
		if err != nil {
			t.Fatalf("failed to get balance: %v", err)
		}
		expected := int64(5050_000_000)
		if balance != expected {
			t.Errorf("expected balance %d, got %d", expected, balance)
		}
	})
}

func TestPositions(t *testing.T) {
	ctx := context.Background()

	tmpFile, err := os.CreateTemp("", "test-*.db")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	conn, err := database.Open(ctx, tmpFile.Name())
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer conn.Close()

	queries := db.New(conn)

	acct, err := queries.CreateAccount(ctx, db.CreateAccountParams{
		ID:              database.NewID(database.PrefixAccount),
		Name:            "Position Test",
		InstitutionName: "test",
		AccountType:     "brokerage",
	})
	if err != nil {
		t.Fatalf("failed to create account: %v", err)
	}

	sec, err := queries.UpsertSecurity(ctx, db.UpsertSecurityParams{
		ID:     database.NewID(database.PrefixSecurity),
		Symbol: "GOOGL",
	})
	if err != nil {
		t.Fatalf("failed to create security: %v", err)
	}

	t.Run("upsert position", func(t *testing.T) {
		pos, err := queries.UpsertPosition(ctx, db.UpsertPositionParams{
			ID:                database.NewID(database.PrefixPosition),
			AccountID:         acct.ID,
			SecurityID:        sec.ID,
			QuantityMicros:    5_000_000,
			CostBasisMicros:   sql.NullInt64{Int64: 500_000_000, Valid: true},
			MarketValueMicros: sql.NullInt64{Int64: 550_000_000, Valid: true},
			AsOfDate:          "2024-01-15",
		})
		if err != nil {
			t.Fatalf("failed to upsert position: %v", err)
		}
		if pos.QuantityMicros != 5_000_000 {
			t.Errorf("expected quantity 5_000_000, got %d", pos.QuantityMicros)
		}
	})

	t.Run("upsert updates existing position", func(t *testing.T) {
		pos, err := queries.UpsertPosition(ctx, db.UpsertPositionParams{
			ID:                database.NewID(database.PrefixPosition),
			AccountID:         acct.ID,
			SecurityID:        sec.ID,
			QuantityMicros:    10_000_000,
			CostBasisMicros:   sql.NullInt64{Int64: 1000_000_000, Valid: true},
			MarketValueMicros: sql.NullInt64{Int64: 1100_000_000, Valid: true},
			AsOfDate:          "2024-01-15",
		})
		if err != nil {
			t.Fatalf("failed to upsert position: %v", err)
		}
		if pos.QuantityMicros != 10_000_000 {
			t.Errorf("expected updated quantity 10_000_000, got %d", pos.QuantityMicros)
		}

		positions, err := queries.ListPositionsByAccount(ctx, acct.ID)
		if err != nil {
			t.Fatalf("failed to list positions: %v", err)
		}
		if len(positions) != 1 {
			t.Errorf("expected 1 position after upsert, got %d", len(positions))
		}
	})
}
