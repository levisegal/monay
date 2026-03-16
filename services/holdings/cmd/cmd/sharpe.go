package cmd

import (
	"context"
	"fmt"
	"math"
	"sort"

	"github.com/rodaine/table"
	"github.com/spf13/cobra"

	"github.com/levisegal/monay/services/holdings/config"
	"github.com/levisegal/monay/services/holdings/database"
	"github.com/levisegal/monay/services/holdings/gen/db"
)

const tradingDaysPerYear = 252

func sharpeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sharpe",
		Short: "Compute Sharpe ratios for total portfolio and each account",
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
			return runSharpe(ctx, queries, cfg.PortfolioURL)
		},
	}
	return cmd
}

type sharpeRow struct {
	Name       string
	AcctType   string
	Value      float64
	NumSymbols int
	AnnReturn  *float64
	AnnVol     *float64
	Sharpe     *float64
	MaxDD      *float64
}

func runSharpe(ctx context.Context, queries *db.Queries, portfolioURL string) error {
	allHoldings, err := queries.ListAllHoldings(ctx)
	if err != nil {
		return fmt.Errorf("failed to list holdings: %w", err)
	}

	accounts, err := queries.ListAccounts(ctx)
	if err != nil {
		return fmt.Errorf("failed to list accounts: %w", err)
	}

	acctMap := make(map[string]db.Account)
	cashByAcct := make(map[string]int64)
	for _, a := range accounts {
		acctMap[a.ID] = a
		cashVal, _ := queries.GetCashBalance(ctx, a.ID)
		cashByAcct[a.ID] = toInt64Val(cashVal)
	}

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
	quotes := fetchQuotes(ctx, portfolioURL, symbols, bondSymbols)

	holdingsByAcct := groupHoldingsByAccount(allHoldings)

	allInputs := buildAnalyzeInputs(allHoldings, acctMap, quotes)
	totalValue := computeTotalValue(positions, accounts, cashByAcct, quotes, bondSymbols)
	totalReturns, riskFreeRate := fetchPortfolioReturnsWithRate(ctx, portfolioURL, allInputs, totalValue)
	totalMetrics := computeMetricsFromReturns(totalReturns, riskFreeRate)

	var rows []sharpeRow

	for _, acct := range accounts {
		holdings := holdingsByAcct[acct.ID]
		if len(holdings) == 0 {
			continue
		}

		acctValue := accountMarketValue(holdings, cashByAcct[acct.ID], quotes)
		if acctValue <= 0 {
			continue
		}

		inputs := buildAnalyzeInputs(holdings, acctMap, quotes)
		returns := fetchPortfolioReturns(ctx, portfolioURL, inputs, acctValue)
		metrics := computeMetricsFromReturns(returns, riskFreeRate)

		rows = append(rows, sharpeRow{
			Name:       fmt.Sprintf("%s %s", acct.InstitutionName, acct.Name),
			AcctType:   acct.AccountType,
			Value:      acctValue,
			NumSymbols: countSymbols(holdings),
			AnnReturn:  metrics.AnnReturn,
			AnnVol:     metrics.AnnVol,
			Sharpe:     metrics.Sharpe,
			MaxDD:      metrics.MaxDD,
		})
	}

	sort.Slice(rows, func(i, j int) bool { return rows[i].Value > rows[j].Value })

	printSharpeTable(rows, totalValue, totalMetrics, riskFreeRate)
	return nil
}

type computedMetrics struct {
	AnnReturn *float64
	AnnVol    *float64
	Sharpe    *float64
	MaxDD     *float64
}

func computeMetricsFromReturns(returns []dailyReturn, riskFreeRate float64) computedMetrics {
	if len(returns) < 30 {
		return computedMetrics{}
	}

	values := make([]float64, len(returns))
	sum := 0.0
	for i, r := range returns {
		values[i] = r.Value
		sum += r.Value
	}

	n := float64(len(values))
	mean := sum / n
	annReturn := mean * tradingDaysPerYear

	variance := 0.0
	for _, v := range values {
		d := v - mean
		variance += d * d
	}
	variance /= (n - 1)
	dailyVol := math.Sqrt(variance)
	annVol := dailyVol * math.Sqrt(tradingDaysPerYear)

	var sharpe *float64
	if annVol > 0 {
		s := (annReturn - riskFreeRate) / annVol
		sharpe = &s
	}

	cumulative := make([]float64, len(values))
	cumulative[0] = 1 + values[0]
	for i := 1; i < len(values); i++ {
		cumulative[i] = cumulative[i-1] * (1 + values[i])
	}
	runningMax := cumulative[0]
	maxDD := 0.0
	for _, c := range cumulative {
		if c > runningMax {
			runningMax = c
		}
		dd := (c - runningMax) / runningMax
		if dd < maxDD {
			maxDD = dd
		}
	}

	return computedMetrics{
		AnnReturn: &annReturn,
		AnnVol:    &annVol,
		Sharpe:    sharpe,
		MaxDD:     &maxDD,
	}
}

func computeCVaR(returns []dailyReturn) float64 {
	values := make([]float64, len(returns))
	for i, r := range returns {
		values[i] = r.Value
	}
	sort.Float64s(values)
	cutoff := len(values) * 5 / 100
	if cutoff < 1 {
		cutoff = 1
	}
	sum := 0.0
	for i := 0; i < cutoff; i++ {
		sum += values[i]
	}
	return (sum / float64(cutoff)) * math.Sqrt(tradingDaysPerYear)
}

func groupHoldingsByAccount(holdings []db.ListAllHoldingsRow) map[string][]db.ListAllHoldingsRow {
	result := make(map[string][]db.ListAllHoldingsRow)
	for _, h := range holdings {
		result[h.AccountID] = append(result[h.AccountID], h)
	}
	return result
}

func accountMarketValue(holdings []db.ListAllHoldingsRow, cashMicros int64, quotes map[string]quoteData) float64 {
	total := float64(cashMicros) / 1_000_000
	for _, h := range holdings {
		qty := nullFloat64ToFloat(h.QuantityMicros) / 1_000_000
		cost := nullFloat64ToFloat(h.CostBasisMicros) / 1_000_000
		total += marketValue(h.Symbol, h.SecurityType.String, qty, cost, quotes)
	}
	return total
}

func countSymbols(holdings []db.ListAllHoldingsRow) int {
	seen := make(map[string]bool)
	for _, h := range holdings {
		seen[h.Symbol] = true
	}
	return len(seen)
}

func printSharpeTable(rows []sharpeRow, totalValue float64, totalMetrics computedMetrics, riskFreeRate float64) {
	fmt.Printf("\nRisk-free rate (10Y Treasury): %.2f%%\n\n", riskFreeRate*100)

	tbl := table.New("ACCOUNT", "TYPE", "VALUE", "SYMBOLS", "SHARPE", "ANN. RETURN", "ANN. VOL", "MAX DD")

	for _, r := range rows {
		tbl.AddRow(
			r.Name,
			formatAccountType(r.AcctType),
			formatCurrency(r.Value),
			r.NumSymbols,
			fmtRatio(r.Sharpe),
			fmtPct(r.AnnReturn),
			fmtPct(r.AnnVol),
			fmtPct(r.MaxDD),
		)
	}

	tbl.AddRow("", "", "", "", "", "", "", "")
	tbl.AddRow(
		"TOTAL PORTFOLIO",
		"",
		formatCurrency(totalValue),
		"",
		fmtRatio(totalMetrics.Sharpe),
		fmtPct(totalMetrics.AnnReturn),
		fmtPct(totalMetrics.AnnVol),
		fmtPct(totalMetrics.MaxDD),
	)

	tbl.Print()
}

func fmtRatio(v *float64) string {
	if v == nil {
		return "-"
	}
	return fmt.Sprintf("%.2f", *v)
}

func fmtPct(v *float64) string {
	if v == nil {
		return "-"
	}
	return fmt.Sprintf("%.1f%%", *v*100)
}
