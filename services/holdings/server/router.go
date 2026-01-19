package server

import (
	"database/sql"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/levisegal/monay/services/holdings/gen/db"
	"github.com/levisegal/monay/services/holdings/version"
)

type Router struct {
	queries *db.Queries
}

func NewRouter(queries *db.Queries) http.Handler {
	r := &Router{queries: queries}

	mux := chi.NewRouter()
	mux.Use(middleware.RequestID)
	mux.Use(middleware.RealIP)
	mux.Use(requestLogger)
	mux.Use(middleware.Recoverer)
	mux.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Content-Type"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	mux.Route("/api/v1", func(api chi.Router) {
		api.Get("/health", r.getHealth)
		api.Get("/version", r.getVersion)
		api.Get("/accounts", r.listAccounts)
		api.Get("/accounts/{id}", r.getAccount)
		api.Get("/holdings", r.listHoldings)
	})

	return mux
}

func requestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		slog.Info("request",
			"method", r.Method,
			"path", r.URL.Path,
			"remote_addr", r.RemoteAddr,
		)
		next.ServeHTTP(w, r)
	})
}

type HealthResponse struct {
	Status string `json:"status"`
}

func (rt *Router) getHealth(w http.ResponseWriter, r *http.Request) {
	respond(w, http.StatusOK, HealthResponse{Status: "ok"})
}

type VersionResponse struct {
	Version  string `json:"version"`
	Revision string `json:"revision"`
}

func (rt *Router) getVersion(w http.ResponseWriter, r *http.Request) {
	ver, rev := version.GetReleaseInfo()
	respond(w, http.StatusOK, VersionResponse{Version: ver, Revision: rev})
}

type AccountResponse struct {
	ID              string  `json:"id"`
	Name            string  `json:"name"`
	InstitutionName string  `json:"institution_name"`
	AccountNumber   *string `json:"account_number,omitempty"`
	AccountType     string  `json:"account_type"`
	CreatedAt       string  `json:"created_at"`
}

type AccountsListResponse struct {
	Accounts []AccountResponse `json:"accounts"`
}

func (rt *Router) listAccounts(w http.ResponseWriter, r *http.Request) {
	accounts, err := rt.queries.ListAccounts(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to list accounts")
		slog.Error("failed to list accounts", "error", err)
		return
	}

	resp := AccountsListResponse{Accounts: make([]AccountResponse, len(accounts))}
	for i, a := range accounts {
		resp.Accounts[i] = accountToResponse(a)
	}
	respond(w, http.StatusOK, resp)
}

func (rt *Router) getAccount(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		respondError(w, http.StatusBadRequest, "missing account id")
		return
	}

	account, err := rt.queries.GetAccount(r.Context(), id)
	if err == sql.ErrNoRows {
		respondError(w, http.StatusNotFound, "account not found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to get account")
		slog.Error("failed to get account", "error", err)
		return
	}

	respond(w, http.StatusOK, accountToResponse(account))
}

type HoldingResponse struct {
	AccountName  string  `json:"account_name"`
	Symbol       string  `json:"symbol"`
	SecurityName *string `json:"security_name,omitempty"`
	Quantity     float64 `json:"quantity"`
	CostBasis    *int64  `json:"cost_basis_micros,omitempty"`
}

type HoldingsListResponse struct {
	Holdings []HoldingResponse `json:"holdings"`
}

func (rt *Router) listHoldings(w http.ResponseWriter, r *http.Request) {
	accountID := r.URL.Query().Get("account_id")

	slog.Info("listHoldings", "account_id", accountID)

	var holdings []HoldingResponse
	if accountID != "" {
		rows, err := rt.queries.ListHoldingsByAccount(r.Context(), accountID)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "failed to list holdings")
			slog.Error("failed to list holdings", "error", err)
			return
		}

		account, err := rt.queries.GetAccount(r.Context(), accountID)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "failed to get account")
			slog.Error("failed to get account", "error", err)
			return
		}

		slog.Info("listHoldingsByAccount", "rows", len(rows))
		holdings = make([]HoldingResponse, len(rows))
		for i, h := range rows {
			holdings[i] = holdingByAccountToResponse(h, account.Name)
		}
	} else {
		rows, err := rt.queries.ListAllHoldings(r.Context())
		if err != nil {
			respondError(w, http.StatusInternalServerError, "failed to list holdings")
			slog.Error("failed to list holdings", "error", err)
			return
		}

		slog.Info("listAllHoldings", "rows", len(rows))
		holdings = make([]HoldingResponse, len(rows))
		for i, h := range rows {
			holdings[i] = holdingToResponse(h)
		}
	}

	respond(w, http.StatusOK, HoldingsListResponse{Holdings: holdings})
}

func accountToResponse(a db.Account) AccountResponse {
	resp := AccountResponse{
		ID:              a.ID,
		Name:            a.Name,
		InstitutionName: a.InstitutionName,
		AccountType:     a.AccountType,
		CreatedAt:       a.CreatedAt,
	}
	if a.ExternalAccountNumber.Valid {
		resp.AccountNumber = &a.ExternalAccountNumber.String
	}
	return resp
}

func holdingToResponse(h db.ListAllHoldingsRow) HoldingResponse {
	resp := HoldingResponse{
		AccountName: h.AccountName,
		Symbol:      h.Symbol,
		Quantity:    h.QuantityMicros.Float64 / 1_000_000,
	}
	if h.SecurityName.Valid {
		resp.SecurityName = &h.SecurityName.String
	}
	if h.CostBasisMicros.Valid {
		costBasis := int64(h.CostBasisMicros.Float64)
		resp.CostBasis = &costBasis
	}
	return resp
}

func holdingByAccountToResponse(h db.ListHoldingsByAccountRow, accountName string) HoldingResponse {
	resp := HoldingResponse{
		AccountName: accountName,
		Symbol:      h.Symbol,
		Quantity:    h.QuantityMicros.Float64 / 1_000_000,
	}
	if h.SecurityName.Valid {
		resp.SecurityName = &h.SecurityName.String
	}
	if h.CostBasisMicros.Valid {
		costBasis := int64(h.CostBasisMicros.Float64)
		resp.CostBasis = &costBasis
	}
	return resp
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func respond(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, message string) {
	respond(w, status, ErrorResponse{Error: message})
}
