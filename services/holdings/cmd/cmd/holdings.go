package cmd

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strings"
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
	cmd.AddCommand(positionsCommand())
	cmd.AddCommand(reportCommand())
	cmd.AddCommand(sharpeCommand())
	cmd.AddCommand(coreTiltCommand())

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

			conn, err := database.Open(ctx, cfg.DBPath)
			if err != nil {
				return err
			}
			defer conn.Close()

			queries := db.New(conn)

			if all {
				return listAllHoldings(ctx, queries, cfg.PortfolioURL, sortBy)
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

			cashBalanceVal, _ := queries.GetCashBalance(ctx, account.ID)
			cashBalance := toInt64Val(cashBalanceVal)

			symbols := make([]string, 0, len(holdings))
			bondSymbols := make(map[string]bool)
			for _, h := range holdings {
				symbols = append(symbols, h.Symbol)
				if h.SecurityType.String == "bond" {
					bondSymbols[h.Symbol] = true
				}
			}
			quotes := fetchQuotes(ctx, cfg.PortfolioURL, symbols, bondSymbols)

			fmt.Printf("\n=== %s: Current Holdings ===\n\n", account.Name)

			tbl := table.New("Symbol", "Quantity", "Cost Basis", "Price", "Mkt Value", "Gain", "Acquired")
			tbl.WithWriter(os.Stdout)

			var totalCostBasis int64
			var totalMktValue float64
			hasMktValue := false
			for _, h := range holdings {
				qty := nullFloat64ToFloat(h.QuantityMicros) / 1_000_000
				cost := nullFloat64ToFloat(h.CostBasisMicros) / 1_000_000
				totalCostBasis += int64(nullFloat64ToFloat(h.CostBasisMicros))

				acquired := ""
				if s, ok := h.EarliestAcquired.(string); ok {
					acquired = s
				}

				costStr := formatCurrency(cost)
				if interfaceToInt(h.EstimatedBasis) != 0 {
					costStr = "~" + costStr
				}

				priceStr, mktStr, gainStr := "-", "-", "-"
				if qd, ok := quotes[h.Symbol]; ok {
					mktValue := qty * qd.Price
					gain := mktValue - cost
					priceStr = formatCurrency(qd.Price)
					mktStr = formatCurrency(mktValue)
					gainStr = formatCurrency(gain)
					totalMktValue += mktValue
					hasMktValue = true
				} else if h.SecurityType.String == "bond" {
					priceStr = "par"
					mktStr = formatCurrency(qty)
					gainStr = formatCurrency(qty - cost)
					totalMktValue += qty
					hasMktValue = true
				}

				tbl.AddRow(h.Symbol, formatQty(qty), costStr, priceStr, mktStr, gainStr, acquired)
			}

			tbl.Print()

			fmt.Printf("\nPositions (cost basis):     %s\n", formatCurrency(float64(totalCostBasis)/1_000_000))
			if hasMktValue {
				fmt.Printf("Positions (est. value):     %s\n", formatCurrency(totalMktValue))
			}
			fmt.Printf("Cash:                       %s\n", formatCurrency(float64(cashBalance)/1_000_000))
			if hasMktValue {
				fmt.Printf("TOTAL (est. value + cash):  %s\n", formatCurrency(totalMktValue+float64(cashBalance)/1_000_000))
			}
			fmt.Printf("TOTAL (cost + cash):        %s\n", formatCurrency(float64(totalCostBasis+cashBalance)/1_000_000))

			return nil
		},
	}

	cmd.Flags().StringVar(&accountName, "account-name", "", "Account name")
	cmd.Flags().BoolVar(&all, "all", false, "Show holdings across all accounts")
	cmd.Flags().StringVar(&sortBy, "sort", "cost", "Sort by: cost, symbol, account")

	return cmd
}

func listAllHoldings(ctx context.Context, queries *db.Queries, portfolioURL string, sortBy string) error {
	holdings, err := queries.ListAllHoldings(ctx)
	if err != nil {
		return fmt.Errorf("failed to list holdings: %w", err)
	}

	accounts, err := queries.ListAccounts(ctx)
	if err != nil {
		return fmt.Errorf("failed to list accounts: %w", err)
	}

	var totalCash int64
	for _, a := range accounts {
		cashVal, _ := queries.GetCashBalance(ctx, a.ID)
		totalCash += toInt64Val(cashVal)
	}

	symbols := make([]string, 0, len(holdings))
	bondSymbols := make(map[string]bool)
	for _, h := range holdings {
		symbols = append(symbols, h.Symbol)
		if h.SecurityType.String == "bond" {
			bondSymbols[h.Symbol] = true
		}
	}
	quotes := fetchQuotes(ctx, portfolioURL, symbols, bondSymbols)

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
			return nullFloat64ToFloat(holdings[i].CostBasisMicros) > nullFloat64ToFloat(holdings[j].CostBasisMicros)
		})
	}

	fmt.Printf("\n=== All Holdings ===\n\n")

	tbl := table.New("Broker", "Account", "Symbol", "Quantity", "Cost Basis", "Price", "Mkt Value", "Gain", "Acquired")
	tbl.WithWriter(os.Stdout)

	var totalCostBasis int64
	var totalMktValue float64
	hasMktValue := false
	for _, h := range holdings {
		qty := nullFloat64ToFloat(h.QuantityMicros) / 1_000_000
		cost := nullFloat64ToFloat(h.CostBasisMicros) / 1_000_000
		totalCostBasis += int64(nullFloat64ToFloat(h.CostBasisMicros))

		acquired := ""
		if s, ok := h.EarliestAcquired.(string); ok {
			acquired = s
		}

		broker := h.Broker
		if broker == "" {
			broker = "-"
		}

		costStr := formatCurrency(cost)
		if interfaceToInt(h.EstimatedBasis) != 0 {
			costStr = "~" + costStr
		}

		priceStr, mktStr, gainStr := "-", "-", "-"
		if qd, ok := quotes[h.Symbol]; ok {
			mktValue := qty * qd.Price
			gain := mktValue - cost
			priceStr = formatCurrency(qd.Price)
			mktStr = formatCurrency(mktValue)
			gainStr = formatCurrency(gain)
			totalMktValue += mktValue
			hasMktValue = true
		} else if h.SecurityType.String == "bond" {
			priceStr = "par"
			mktStr = formatCurrency(qty)
			gainStr = formatCurrency(qty - cost)
			totalMktValue += qty
			hasMktValue = true
		}

		tbl.AddRow(broker, h.AccountName, h.Symbol, formatQty(qty), costStr, priceStr, mktStr, gainStr, acquired)
	}

	tbl.Print()

	fmt.Printf("\nPositions (cost basis):     %s\n", formatCurrency(float64(totalCostBasis)/1_000_000))
	if hasMktValue {
		fmt.Printf("Positions (est. value):     %s\n", formatCurrency(totalMktValue))
	}
	fmt.Printf("Cash:                       %s\n", formatCurrency(float64(totalCash)/1_000_000))
	if hasMktValue {
		fmt.Printf("TOTAL (est. value + cash):  %s\n", formatCurrency(totalMktValue+float64(totalCash)/1_000_000))
	}
	fmt.Printf("TOTAL (cost + cash):        %s\n", formatCurrency(float64(totalCostBasis+totalCash)/1_000_000))

	return nil
}

func positionsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "positions",
		Short: "List positions aggregated by symbol",
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

			positions, err := queries.ListPositions(ctx)
			if err != nil {
				return fmt.Errorf("failed to list positions: %w", err)
			}

			symbols := make([]string, 0, len(positions))
			bondSymbols := make(map[string]bool)
			for _, p := range positions {
				symbols = append(symbols, p.Symbol)
				if p.SecurityType.String == "bond" {
					bondSymbols[p.Symbol] = true
				}
			}
			quotes := fetchQuotes(ctx, cfg.PortfolioURL, symbols, bondSymbols)

			fmt.Printf("\n=== Positions (by Symbol) ===\n\n")

			tbl := table.New("Symbol", "Quantity", "Cost Basis", "Price", "Mkt Value", "Gain", "Accounts", "Acquired")
			tbl.WithWriter(os.Stdout)

			var totalCostBasis int64
			var totalMktValue float64
			hasMktValue := false
			for _, p := range positions {
				qty := nullFloat64ToFloat(p.QuantityMicros) / 1_000_000
				cost := nullFloat64ToFloat(p.CostBasisMicros) / 1_000_000
				totalCostBasis += int64(nullFloat64ToFloat(p.CostBasisMicros))

				acquired := ""
				if s, ok := p.EarliestAcquired.(string); ok {
					acquired = s
				}

				costStr := formatCurrency(cost)
				if interfaceToInt(p.EstimatedBasis) != 0 {
					costStr = "~" + costStr
				}

				priceStr, mktStr, gainStr := "-", "-", "-"
				if qd, ok := quotes[p.Symbol]; ok {
					mktValue := qty * qd.Price
					gain := mktValue - cost
					priceStr = formatCurrency(qd.Price)
					mktStr = formatCurrency(mktValue)
					gainStr = formatCurrency(gain)
					totalMktValue += mktValue
					hasMktValue = true
				} else if p.SecurityType.String == "bond" {
					priceStr = "par"
					mktStr = formatCurrency(qty)
					gainStr = formatCurrency(qty - cost)
					totalMktValue += qty
					hasMktValue = true
				}

				tbl.AddRow(p.Symbol, formatQty(qty), costStr, priceStr, mktStr, gainStr, p.AccountCount, acquired)
			}

			tbl.Print()

			fmt.Printf("\nTOTAL (cost basis):  %s\n", formatCurrency(float64(totalCostBasis)/1_000_000))
			if hasMktValue {
				fmt.Printf("TOTAL (est. value):  %s\n", formatCurrency(totalMktValue))
			}

			return nil
		},
	}

	return cmd
}

func formatCurrency(amount float64) string {
	p := message.NewPrinter(language.English)
	return p.Sprintf("$%.2f", amount)
}

func formatQty(qty float64) string {
	p := message.NewPrinter(language.English)
	return p.Sprintf("%.2f", qty)
}

func nullFloat64ToFloat(nf interface{}) float64 {
	switch v := nf.(type) {
	case sql.NullFloat64:
		return v.Float64
	case float64:
		return v
	case int64:
		return float64(v)
	default:
		return 0
	}
}

func toInt64Val(v interface{}) int64 {
	switch val := v.(type) {
	case int64:
		return val
	case int:
		return int64(val)
	case float64:
		return int64(val)
	default:
		return 0
	}
}

func interfaceToInt(v interface{}) int64 {
	switch val := v.(type) {
	case int64:
		return val
	case float64:
		return int64(val)
	default:
		return 0
	}
}

type quoteData struct {
	Price           float64
	Name            string
	Sector          string
	Industry        string
	Category        string
	DividendRate    float64
	DividendYield   float64
	YieldPct        float64
	NetExpenseRatio float64
	FundFamily      string
}

func fetchQuotes(ctx context.Context, portfolioURL string, symbols []string, skipSymbols map[string]bool) map[string]quoteData {
	result := make(map[string]quoteData)

	var filtered []string
	for _, s := range symbols {
		if !skipSymbols[s] {
			filtered = append(filtered, s)
		}
	}
	if len(filtered) == 0 {
		return result
	}

	seen := make(map[string]bool, len(filtered))
	var unique []string
	for _, s := range filtered {
		if !seen[s] {
			seen[s] = true
			unique = append(unique, s)
		}
	}

	url := fmt.Sprintf("%s/api/v1/quotes?symbols=%s", portfolioURL, strings.Join(unique, ","))

	client := &http.Client{Timeout: 120 * time.Second}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return result
	}

	resp, err := client.Do(req)
	if err != nil {
		return result
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return result
	}

	var body struct {
		Quotes []struct {
			Symbol        string   `json:"symbol"`
			Name          string   `json:"name"`
			Price         *float64 `json:"price"`
			Sector        string   `json:"sector"`
			Industry      string   `json:"industry"`
			Category      string   `json:"category"`
			DividendRate    *float64 `json:"dividend_rate"`
			DividendYield   *float64 `json:"dividend_yield"`
			YieldPct        *float64 `json:"yield_pct"`
			NetExpenseRatio *float64 `json:"net_expense_ratio"`
			FundFamily      string   `json:"fund_family"`
		} `json:"quotes"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return result
	}

	for _, q := range body.Quotes {
		if q.Price != nil {
			qd := quoteData{
				Price:    *q.Price,
				Name:     q.Name,
				Sector:   q.Sector,
				Industry: q.Industry,
				Category: q.Category,
			}
			if q.DividendRate != nil {
				qd.DividendRate = *q.DividendRate
			}
			if q.DividendYield != nil {
				qd.DividendYield = *q.DividendYield
			}
			if q.YieldPct != nil {
				qd.YieldPct = *q.YieldPct
			}
			if q.NetExpenseRatio != nil {
				qd.NetExpenseRatio = *q.NetExpenseRatio
			}
			qd.FundFamily = q.FundFamily
			result[q.Symbol] = qd
		}
	}
	return result
}

