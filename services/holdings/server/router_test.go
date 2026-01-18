package server_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/levisegal/monay/services/holdings/database"
	"github.com/levisegal/monay/services/holdings/gen/db"
	"github.com/levisegal/monay/services/holdings/server"
)

func setupTestDB(t *testing.T) (*sql.DB, *db.Queries, func()) {
	t.Helper()
	ctx := context.Background()

	tmpFile, err := os.CreateTemp("", "router-test-*.db")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	tmpFile.Close()

	conn, err := database.Open(ctx, tmpFile.Name())
	if err != nil {
		os.Remove(tmpFile.Name())
		t.Fatalf("failed to open database: %v", err)
	}

	queries := db.New(conn)
	cleanup := func() {
		conn.Close()
		os.Remove(tmpFile.Name())
	}

	return conn, queries, cleanup
}

func TestHealthEndpoint(t *testing.T) {
	_, queries, cleanup := setupTestDB(t)
	defer cleanup()

	handler := server.NewRouter(queries)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var resp struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Status != "ok" {
		t.Errorf("expected status 'ok', got %q", resp.Status)
	}
}

func TestVersionEndpoint(t *testing.T) {
	_, queries, cleanup := setupTestDB(t)
	defer cleanup()

	handler := server.NewRouter(queries)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/version", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var resp struct {
		Version  string `json:"version"`
		Revision string `json:"revision"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Version == "" {
		t.Error("expected non-empty version")
	}
}

func TestListAccountsEndpoint(t *testing.T) {
	ctx := context.Background()
	_, queries, cleanup := setupTestDB(t)
	defer cleanup()

	_, err := queries.CreateAccount(ctx, db.CreateAccountParams{
		ID:              database.NewID(database.PrefixAccount),
		Name:            "Brokerage Account",
		InstitutionName: "etrade",
		AccountType:     "brokerage",
	})
	if err != nil {
		t.Fatalf("failed to create account: %v", err)
	}

	_, err = queries.CreateAccount(ctx, db.CreateAccountParams{
		ID:              database.NewID(database.PrefixAccount),
		Name:            "Retirement Account",
		InstitutionName: "fidelity",
		AccountType:     "ira",
	})
	if err != nil {
		t.Fatalf("failed to create account: %v", err)
	}

	handler := server.NewRouter(queries)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/accounts", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var resp struct {
		Accounts []struct {
			ID              string `json:"id"`
			Name            string `json:"name"`
			InstitutionName string `json:"institution_name"`
			AccountType     string `json:"account_type"`
		} `json:"accounts"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(resp.Accounts) != 2 {
		t.Errorf("expected 2 accounts, got %d", len(resp.Accounts))
	}
	if resp.Accounts[0].Name != "Brokerage Account" {
		t.Errorf("expected first account 'Brokerage Account', got %q", resp.Accounts[0].Name)
	}
}

func TestGetAccountEndpoint(t *testing.T) {
	ctx := context.Background()
	_, queries, cleanup := setupTestDB(t)
	defer cleanup()

	acct, err := queries.CreateAccount(ctx, db.CreateAccountParams{
		ID:              database.NewID(database.PrefixAccount),
		Name:            "Test Account",
		InstitutionName: "schwab",
		AccountType:     "brokerage",
	})
	if err != nil {
		t.Fatalf("failed to create account: %v", err)
	}

	handler := server.NewRouter(queries)

	t.Run("existing account", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/accounts/"+acct.ID, nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
		}

		var resp struct {
			ID              string `json:"id"`
			Name            string `json:"name"`
			InstitutionName string `json:"institution_name"`
		}
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if resp.ID != acct.ID {
			t.Errorf("expected id %q, got %q", acct.ID, resp.ID)
		}
		if resp.Name != "Test Account" {
			t.Errorf("expected name 'Test Account', got %q", resp.Name)
		}
	})

	t.Run("non-existent account", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/accounts/acct_nonexistent", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Errorf("expected status %d, got %d", http.StatusNotFound, rec.Code)
		}
	})
}

func TestListHoldingsEndpoint(t *testing.T) {
	ctx := context.Background()
	_, queries, cleanup := setupTestDB(t)
	defer cleanup()

	acct, err := queries.CreateAccount(ctx, db.CreateAccountParams{
		ID:              database.NewID(database.PrefixAccount),
		Name:            "Investment Account",
		InstitutionName: "etrade",
		AccountType:     "brokerage",
	})
	if err != nil {
		t.Fatalf("failed to create account: %v", err)
	}

	sec, err := queries.UpsertSecurity(ctx, db.UpsertSecurityParams{
		ID:     database.NewID(database.PrefixSecurity),
		Symbol: "AAPL",
		Name:   sql.NullString{String: "Apple Inc.", Valid: true},
	})
	if err != nil {
		t.Fatalf("failed to create security: %v", err)
	}

	_, err = queries.UpsertPosition(ctx, db.UpsertPositionParams{
		ID:                database.NewID(database.PrefixPosition),
		AccountID:         acct.ID,
		SecurityID:        sec.ID,
		QuantityMicros:    100_000_000,
		CostBasisMicros:   sql.NullInt64{Int64: 15000_000_000, Valid: true},
		MarketValueMicros: sql.NullInt64{Int64: 17500_000_000, Valid: true},
		AsOfDate:          "2024-01-15",
	})
	if err != nil {
		t.Fatalf("failed to create position: %v", err)
	}

	handler := server.NewRouter(queries)

	t.Run("list all holdings", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/holdings", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
		}

		var resp struct {
			Holdings []struct {
				ID              string   `json:"id"`
				AccountID       string   `json:"account_id"`
				AccountName     string   `json:"account_name"`
				Symbol          string   `json:"symbol"`
				SecurityName    *string  `json:"security_name"`
				Quantity        float64  `json:"quantity"`
				CostBasisMicros *int64   `json:"cost_basis_micros"`
			} `json:"holdings"`
		}
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if len(resp.Holdings) != 1 {
			t.Fatalf("expected 1 holding, got %d", len(resp.Holdings))
		}
		h := resp.Holdings[0]
		if h.Symbol != "AAPL" {
			t.Errorf("expected symbol 'AAPL', got %q", h.Symbol)
		}
		if h.Quantity != 100.0 {
			t.Errorf("expected quantity 100.0, got %f", h.Quantity)
		}
		if h.AccountName != "Investment Account" {
			t.Errorf("expected account name 'Investment Account', got %q", h.AccountName)
		}
		if h.SecurityName == nil || *h.SecurityName != "Apple Inc." {
			t.Errorf("expected security name 'Apple Inc.', got %v", h.SecurityName)
		}
		if h.CostBasisMicros == nil || *h.CostBasisMicros != 15000_000_000 {
			t.Errorf("expected cost basis 15000_000_000, got %v", h.CostBasisMicros)
		}
	})

	t.Run("filter by account", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/holdings?account_id="+acct.ID, nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
		}

		var resp struct {
			Holdings []struct {
				AccountID string `json:"account_id"`
			} `json:"holdings"`
		}
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if len(resp.Holdings) != 1 {
			t.Fatalf("expected 1 holding, got %d", len(resp.Holdings))
		}
		if resp.Holdings[0].AccountID != acct.ID {
			t.Errorf("expected account_id %q, got %q", acct.ID, resp.Holdings[0].AccountID)
		}
	})

	t.Run("filter by non-existent account", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/holdings?account_id=acct_nonexistent", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusInternalServerError {
			t.Errorf("expected status %d, got %d", http.StatusInternalServerError, rec.Code)
		}
	})
}

func TestListHoldingsMultiplePositions(t *testing.T) {
	ctx := context.Background()
	_, queries, cleanup := setupTestDB(t)
	defer cleanup()

	acct1, _ := queries.CreateAccount(ctx, db.CreateAccountParams{
		ID:              database.NewID(database.PrefixAccount),
		Name:            "Account A",
		InstitutionName: "etrade",
		AccountType:     "brokerage",
	})
	acct2, _ := queries.CreateAccount(ctx, db.CreateAccountParams{
		ID:              database.NewID(database.PrefixAccount),
		Name:            "Account B",
		InstitutionName: "fidelity",
		AccountType:     "ira",
	})

	secAAPL, _ := queries.UpsertSecurity(ctx, db.UpsertSecurityParams{
		ID:     database.NewID(database.PrefixSecurity),
		Symbol: "AAPL",
		Name:   sql.NullString{String: "Apple Inc.", Valid: true},
	})
	secMSFT, _ := queries.UpsertSecurity(ctx, db.UpsertSecurityParams{
		ID:     database.NewID(database.PrefixSecurity),
		Symbol: "MSFT",
		Name:   sql.NullString{String: "Microsoft Corp", Valid: true},
	})

	queries.UpsertPosition(ctx, db.UpsertPositionParams{
		ID:             database.NewID(database.PrefixPosition),
		AccountID:      acct1.ID,
		SecurityID:     secAAPL.ID,
		QuantityMicros: 50_000_000,
		AsOfDate:       "2024-01-15",
	})
	queries.UpsertPosition(ctx, db.UpsertPositionParams{
		ID:             database.NewID(database.PrefixPosition),
		AccountID:      acct1.ID,
		SecurityID:     secMSFT.ID,
		QuantityMicros: 25_000_000,
		AsOfDate:       "2024-01-15",
	})
	queries.UpsertPosition(ctx, db.UpsertPositionParams{
		ID:             database.NewID(database.PrefixPosition),
		AccountID:      acct2.ID,
		SecurityID:     secAAPL.ID,
		QuantityMicros: 75_000_000,
		AsOfDate:       "2024-01-15",
	})

	handler := server.NewRouter(queries)

	t.Run("list all returns all positions", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/holdings", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		var resp struct {
			Holdings []struct {
				Symbol      string `json:"symbol"`
				AccountName string `json:"account_name"`
			} `json:"holdings"`
		}
		json.NewDecoder(rec.Body).Decode(&resp)

		if len(resp.Holdings) != 3 {
			t.Errorf("expected 3 holdings, got %d", len(resp.Holdings))
		}
	})

	t.Run("filter returns only matching account", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/holdings?account_id="+acct1.ID, nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		var resp struct {
			Holdings []struct {
				Symbol string `json:"symbol"`
			} `json:"holdings"`
		}
		json.NewDecoder(rec.Body).Decode(&resp)

		if len(resp.Holdings) != 2 {
			t.Errorf("expected 2 holdings for account A, got %d", len(resp.Holdings))
		}
	})
}

func TestEmptyResponses(t *testing.T) {
	_, queries, cleanup := setupTestDB(t)
	defer cleanup()

	handler := server.NewRouter(queries)

	t.Run("empty accounts list", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/accounts", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
		}

		var resp struct {
			Accounts []any `json:"accounts"`
		}
		json.NewDecoder(rec.Body).Decode(&resp)

		if len(resp.Accounts) != 0 {
			t.Errorf("expected empty accounts, got %d", len(resp.Accounts))
		}
	})

	t.Run("empty holdings list", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/holdings", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
		}

		var resp struct {
			Holdings []any `json:"holdings"`
		}
		json.NewDecoder(rec.Body).Decode(&resp)

		if len(resp.Holdings) != 0 {
			t.Errorf("expected empty holdings, got %d", len(resp.Holdings))
		}
	})
}

func TestContentType(t *testing.T) {
	_, queries, cleanup := setupTestDB(t)
	defer cleanup()

	handler := server.NewRouter(queries)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	contentType := rec.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("expected Content-Type 'application/json', got %q", contentType)
	}
}
