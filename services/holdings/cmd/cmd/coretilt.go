package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/levisegal/monay/services/holdings/config"
	"github.com/levisegal/monay/services/holdings/database"
	"github.com/levisegal/monay/services/holdings/gen/db"
)

func coreTiltCommand() *cobra.Command {
	var configFlag string

	cmd := &cobra.Command{
		Use:   "core-tilt",
		Short: "Analyze portfolio core holdings vs thematic tilts with fee breakdown",
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

			ctPath := configFlag
			if ctPath == "" {
				ctPath = cfg.CoreTiltsPath
			}
			if ctPath == "" {
				return fmt.Errorf("core-tilts config required: use --config or set MONAY_HOLDINGS_CORE_TILTS_PATH")
			}

			queries := db.New(conn)
			return runCoreTilt(ctx, queries, cfg.PortfolioURL, ctPath)
		},
	}

	cmd.Flags().StringVar(&configFlag, "config", "", "Path to core-tilts YAML config")
	return cmd
}

type coreTiltConfig struct {
	Core  map[string]string          `yaml:"core"`
	Tilts map[string]tiltDefinition  `yaml:"tilts"`
}

type tiltDefinition struct {
	Thesis  string   `yaml:"thesis"`
	Symbols []string `yaml:"symbols"`
}

type classifiedPosition struct {
	Symbol       string
	Name         string
	Value        float64
	Weight       float64
	ExpenseBps   int
	FundFamily   string
	AnnualFee    float64
	SecurityType string
}

type tiltGroup struct {
	Name      string
	Thesis    string
	Positions []classifiedPosition
	Value     float64
	Weight    float64
	AvgER     float64
	AnnFee    float64
}

func runCoreTilt(ctx context.Context, queries *db.Queries, portfolioURL, configPath string) error {
	ctConfig, err := loadCoreTiltConfig(configPath)
	if err != nil {
		return err
	}

	positions, err := queries.ListPositions(ctx)
	if err != nil {
		return fmt.Errorf("failed to list positions: %w", err)
	}

	accounts, err := queries.ListAccounts(ctx)
	if err != nil {
		return fmt.Errorf("failed to list accounts: %w", err)
	}

	cashByAcct := make(map[string]int64)
	for _, a := range accounts {
		cashVal, _ := queries.GetCashBalance(ctx, a.ID)
		cashByAcct[a.ID] = toInt64Val(cashVal)
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

	totalValue := computeTotalValue(positions, accounts, cashByAcct, quotes, bondSymbols)

	securities, err := queries.ListSecuritiesWithOpenLots(ctx)
	if err != nil {
		return fmt.Errorf("failed to list securities: %w", err)
	}
	secMap := make(map[string]db.Security)
	for _, s := range securities {
		secMap[s.Symbol] = s
	}

	symbolToTilt := make(map[string]string)
	for tiltName, td := range ctConfig.Tilts {
		for _, sym := range td.Symbols {
			symbolToTilt[sym] = tiltName
		}
	}

	positionsBySymbol := aggregatePositions(positions, quotes, totalValue)

	var corePositions []classifiedPosition
	tiltGroups := make(map[string]*tiltGroup)
	var unclassified []classifiedPosition

	for sym, cp := range positionsBySymbol {
		sec := secMap[sym]
		cp.ExpenseBps = expenseBps(sec, quotes[sym])
		cp.FundFamily = fundFamily(sec, quotes[sym])
		cp.AnnualFee = float64(cp.ExpenseBps) / 10000.0 * cp.Value
		positionsBySymbol[sym] = cp

		if _, isCore := ctConfig.Core[sym]; isCore {
			corePositions = append(corePositions, cp)
		} else if tiltName, isTilt := symbolToTilt[sym]; isTilt {
			tg, ok := tiltGroups[tiltName]
			if !ok {
				td := ctConfig.Tilts[tiltName]
				tg = &tiltGroup{Name: tiltName, Thesis: td.Thesis}
				tiltGroups[tiltName] = tg
			}
			tg.Positions = append(tg.Positions, cp)
			tg.Value += cp.Value
			tg.AnnFee += cp.AnnualFee
		} else {
			unclassified = append(unclassified, cp)
		}
	}

	for _, tg := range tiltGroups {
		tg.Weight = tg.Value / totalValue
		if tg.Value > 0 {
			tg.AvgER = tg.AnnFee / tg.Value * 10000
		}
	}

	sort.Slice(corePositions, func(i, j int) bool { return corePositions[i].Value > corePositions[j].Value })
	sort.Slice(unclassified, func(i, j int) bool { return unclassified[i].Value > unclassified[j].Value })

	allClassified := make([]string, 0, len(corePositions))
	for _, cp := range corePositions {
		allClassified = append(allClassified, cp.Symbol)
	}
	for _, tiltName := range sortedTiltNames(tiltGroups) {
		for _, cp := range tiltGroups[tiltName].Positions {
			if cp.SecurityType == "etf" || cp.SecurityType == "" {
				allClassified = append(allClassified, cp.Symbol)
			}
		}
	}

	overlaps := computeOverlaps(ctx, portfolioURL, allClassified, positionsBySymbol)

	printCoreTiltReport(corePositions, tiltGroups, unclassified, totalValue, ctConfig)
	printOverlapAnalysis(overlaps)
	return nil
}

func sortedTiltNames(tilts map[string]*tiltGroup) []string {
	names := make([]string, 0, len(tilts))
	for n := range tilts {
		names = append(names, n)
	}
	sort.Slice(names, func(i, j int) bool { return tilts[names[i]].Value > tilts[names[j]].Value })
	return names
}

func aggregatePositions(positions []db.ListPositionsRow, quotes map[string]quoteData, totalValue float64) map[string]classifiedPosition {
	symbolValues := make(map[string]float64)
	symbolTypes := make(map[string]string)
	for _, p := range positions {
		qty := nullFloat64ToFloat(p.QuantityMicros) / 1_000_000
		cost := nullFloat64ToFloat(p.CostBasisMicros) / 1_000_000
		mv := marketValue(p.Symbol, p.SecurityType.String, qty, cost, quotes)
		symbolValues[p.Symbol] += mv
		symbolTypes[p.Symbol] = p.SecurityType.String
	}

	result := make(map[string]classifiedPosition)
	for sym, val := range symbolValues {
		qd := quotes[sym]
		name := qd.Name
		if name == "" {
			name = sym
		}
		result[sym] = classifiedPosition{
			Symbol:       sym,
			Name:         name,
			Value:        val,
			Weight:       val / totalValue,
			SecurityType: symbolTypes[sym],
		}
	}
	return result
}

func expenseBps(sec db.Security, qd quoteData) int {
	if sec.ExpenseRatioBps.Valid {
		return int(sec.ExpenseRatioBps.Int64)
	}
	if qd.NetExpenseRatio > 0 {
		return int(qd.NetExpenseRatio * 100)
	}
	return 0
}

func fundFamily(sec db.Security, qd quoteData) string {
	if sec.FundFamily.Valid {
		return sec.FundFamily.String
	}
	return qd.FundFamily
}

func printCoreTiltReport(core []classifiedPosition, tilts map[string]*tiltGroup, unclassified []classifiedPosition, totalValue float64, cfg *coreTiltConfig) {
	coreValue, coreFee := 0.0, 0.0
	tiltValue, tiltFee := 0.0, 0.0
	unclValue, unclFee := 0.0, 0.0

	fmt.Println()
	fmt.Println("═══ CORE HOLDINGS ═══")
	fmt.Printf("%-8s %-30s %12s %7s %6s %10s %s\n", "SYMBOL", "NAME", "VALUE", "WEIGHT", "ER", "ANN FEE", "MANAGER")
	fmt.Println(strings.Repeat("─", 95))
	for _, p := range core {
		role := cfg.Core[p.Symbol]
		name := truncate(p.Name, 25)
		if role != "" {
			name = truncate(role, 25)
		}
		fmt.Printf("%-8s %-30s %12s %6.1f%% %4dbps %10s %s\n",
			p.Symbol, name, formatCurrency(p.Value), p.Weight*100, p.ExpenseBps, formatCurrency(p.AnnualFee), p.FundFamily)
		coreValue += p.Value
		coreFee += p.AnnualFee
	}
	coreWeight := coreValue / totalValue * 100
	coreAvgER := 0.0
	if coreValue > 0 {
		coreAvgER = coreFee / coreValue * 10000
	}
	fmt.Println(strings.Repeat("─", 95))
	fmt.Printf("%-39s %12s %6.1f%% %4.0fbps %10s\n", "Core Total", formatCurrency(coreValue), coreWeight, coreAvgER, formatCurrency(coreFee))

	sortedTilts := make([]string, 0, len(tilts))
	for name := range tilts {
		sortedTilts = append(sortedTilts, name)
	}
	sort.Slice(sortedTilts, func(i, j int) bool { return tilts[sortedTilts[i]].Value > tilts[sortedTilts[j]].Value })

	fmt.Println()
	fmt.Println("═══ THEMATIC TILTS ═══")
	for _, name := range sortedTilts {
		tg := tilts[name]
		tiltValue += tg.Value
		tiltFee += tg.AnnFee

		fmt.Printf("\n[%s] %s (%.1f%%)\n", strings.ToUpper(name), tg.Thesis, tg.Weight*100)
		sort.Slice(tg.Positions, func(i, j int) bool { return tg.Positions[i].Value > tg.Positions[j].Value })
		for _, p := range tg.Positions {
			fmt.Printf("  %-8s %-28s %12s %6.1f%% %4dbps %10s %s\n",
				p.Symbol, truncate(p.Name, 24), formatCurrency(p.Value), p.Weight*100, p.ExpenseBps, formatCurrency(p.AnnualFee), p.FundFamily)
		}
	}
	tiltWeight := tiltValue / totalValue * 100
	tiltAvgER := 0.0
	if tiltValue > 0 {
		tiltAvgER = tiltFee / tiltValue * 10000
	}
	fmt.Println()
	fmt.Println(strings.Repeat("─", 95))
	fmt.Printf("%-39s %12s %6.1f%% %4.0fbps %10s\n", "Tilts Total", formatCurrency(tiltValue), tiltWeight, tiltAvgER, formatCurrency(tiltFee))

	if len(unclassified) > 0 {
		fmt.Println()
		fmt.Println("═══ UNCLASSIFIED ═══")
		for _, p := range unclassified {
			fmt.Printf("  %-8s %-28s %12s %6.1f%% %4dbps %10s %s\n",
				p.Symbol, truncate(p.Name, 24), formatCurrency(p.Value), p.Weight*100, p.ExpenseBps, formatCurrency(p.AnnualFee), p.FundFamily)
			unclValue += p.Value
			unclFee += p.AnnualFee
		}
		unclWeight := unclValue / totalValue * 100
		fmt.Println(strings.Repeat("─", 95))
		fmt.Printf("%-39s %12s %6.1f%%        %10s\n", "Unclassified Total", formatCurrency(unclValue), unclWeight, formatCurrency(unclFee))
	}

	totalFee := coreFee + tiltFee + unclFee
	blendedER := 0.0
	if totalValue > 0 {
		blendedER = totalFee / totalValue * 10000
	}

	fmt.Println()
	fmt.Println("═══ FEE SUMMARY ═══")
	fmt.Printf("Portfolio value:    %s\n", formatCurrency(totalValue))
	fmt.Printf("Core fees:          %s  (%.0f bps weighted avg)\n", formatCurrency(coreFee), coreAvgER)
	fmt.Printf("Tilt fees:          %s  (%.0f bps weighted avg)\n", formatCurrency(tiltFee), tiltAvgER)
	fmt.Printf("Total annual fees:  %s  (%.0f bps blended)\n", formatCurrency(totalFee), blendedER)
	fmt.Println()

	familyFees := make(map[string]float64)
	familyValues := make(map[string]float64)
	allPositions := append(append(core, unclassified...), collectTiltPositions(tilts)...)
	for _, p := range allPositions {
		fam := p.FundFamily
		if fam == "" {
			fam = "(individual stocks)"
		}
		familyFees[fam] += p.AnnualFee
		familyValues[fam] += p.Value
	}

	type famRow struct {
		Name  string
		Value float64
		Fee   float64
	}
	var famRows []famRow
	for name, fee := range familyFees {
		famRows = append(famRows, famRow{name, familyValues[name], fee})
	}
	sort.Slice(famRows, func(i, j int) bool { return famRows[i].Value > famRows[j].Value })

	fmt.Println("═══ BY FUND MANAGER ═══")
	fmt.Printf("%-25s %12s %7s %10s\n", "MANAGER", "VALUE", "WEIGHT", "ANN FEE")
	fmt.Println(strings.Repeat("─", 60))
	for _, fr := range famRows {
		w := fr.Value / totalValue * 100
		fmt.Printf("%-25s %12s %6.1f%% %10s\n", truncate(fr.Name, 24), formatCurrency(fr.Value), w, formatCurrency(fr.Fee))
	}
	fmt.Println()
}

func collectTiltPositions(tilts map[string]*tiltGroup) []classifiedPosition {
	var all []classifiedPosition
	for _, tg := range tilts {
		all = append(all, tg.Positions...)
	}
	return all
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-1] + "…"
}

type overlapPair struct {
	SymA, SymB  string
	Correlation float64
	Note        string
	CombinedFee float64
	Actionable  bool
}

func computeOverlaps(ctx context.Context, portfolioURL string, symbols []string, posMap map[string]classifiedPosition) []overlapPair {
	returnsBySymbol := make(map[string][]float64)
	datesBySymbol := make(map[string][]string)

	client := &http.Client{Timeout: 30 * time.Second}
	for _, sym := range symbols {
		url := fmt.Sprintf("%s/api/v1/chart/%s?range=1y", portfolioURL, sym)
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			continue
		}
		resp, err := client.Do(req)
		if err != nil {
			continue
		}
		var chart struct {
			Points []struct {
				Timestamp string   `json:"timestamp"`
				Close     *float64 `json:"close"`
			} `json:"points"`
		}
		json.NewDecoder(resp.Body).Decode(&chart)
		resp.Body.Close()

		if len(chart.Points) < 30 {
			continue
		}

		var dates []string
		var returns []float64
		for i := 1; i < len(chart.Points); i++ {
			if chart.Points[i].Close == nil || chart.Points[i-1].Close == nil || *chart.Points[i-1].Close == 0 {
				continue
			}
			ret := (*chart.Points[i].Close - *chart.Points[i-1].Close) / *chart.Points[i-1].Close
			returns = append(returns, ret)
			dates = append(dates, chart.Points[i].Timestamp)
		}
		returnsBySymbol[sym] = returns
		datesBySymbol[sym] = dates
	}

	var pairs []overlapPair
	for i := 0; i < len(symbols); i++ {
		for j := i + 1; j < len(symbols); j++ {
			a, b := symbols[i], symbols[j]
			rA, okA := returnsBySymbol[a]
			rB, okB := returnsBySymbol[b]
			if !okA || !okB {
				continue
			}

			corr := pearsonCorrelation(rA, rB, datesBySymbol[a], datesBySymbol[b])
			if corr < 0.80 {
				continue
			}

			erA := posMap[a].ExpenseBps
			erB := posMap[b].ExpenseBps
			valA := posMap[a].Value
			valB := posMap[b].Value
			bothHaveFees := erA > 0 && erB > 0
			eitherHasFees := erA > 0 || erB > 0

			keep, sell := a, b
			if erA > erB || (erA == erB && valA < valB) {
				keep, sell = b, a
			}
			keepER := posMap[keep].ExpenseBps
			sellER := posMap[sell].ExpenseBps
			sellVal := posMap[sell].Value
			savings := sellVal * float64(sellER-keepER) / 10000.0

			note := ""
			actionable := false
			if corr >= 0.95 && bothHaveFees && savings >= 5 {
				note = fmt.Sprintf("REDUNDANT — sell %s, keep %s → save $%.0f/yr", sell, keep, savings)
				actionable = true
			} else if corr >= 0.95 {
				note = "identical exposure, no fee impact"
			} else if corr >= 0.85 && eitherHasFees && savings >= 10 {
				note = fmt.Sprintf("HIGH OVERLAP — sell %s, keep %s → save $%.0f/yr", sell, keep, savings)
				actionable = true
			} else if corr >= 0.85 {
				note = "high overlap, no fee impact"
			} else if eitherHasFees {
				note = "moderate overlap with fee drag"
			} else {
				note = "moderate overlap, no fee impact"
			}

			pairs = append(pairs, overlapPair{
				SymA:        a,
				SymB:        b,
				Correlation: corr,
				Note:        note,
				CombinedFee: savings,
				Actionable:  actionable,
			})
		}
	}

	sort.Slice(pairs, func(i, j int) bool { return pairs[i].Correlation > pairs[j].Correlation })
	return pairs
}

func pearsonCorrelation(a, b []float64, datesA, datesB []string) float64 {
	dateMapA := make(map[string]float64, len(datesA))
	for i, d := range datesA {
		dateMapA[d] = a[i]
	}

	var xVals, yVals []float64
	for i, d := range datesB {
		if va, ok := dateMapA[d]; ok {
			xVals = append(xVals, va)
			yVals = append(yVals, b[i])
		}
	}

	n := float64(len(xVals))
	if n < 30 {
		return 0
	}

	sumX, sumY, sumXY, sumX2, sumY2 := 0.0, 0.0, 0.0, 0.0, 0.0
	for i := range xVals {
		sumX += xVals[i]
		sumY += yVals[i]
		sumXY += xVals[i] * yVals[i]
		sumX2 += xVals[i] * xVals[i]
		sumY2 += yVals[i] * yVals[i]
	}

	denom := math.Sqrt((n*sumX2 - sumX*sumX) * (n*sumY2 - sumY*sumY))
	if denom == 0 {
		return 0
	}
	return (n*sumXY - sumX*sumY) / denom
}

func printOverlapAnalysis(overlaps []overlapPair) {
	if len(overlaps) == 0 {
		return
	}

	var actionable, informational []overlapPair
	for _, o := range overlaps {
		if o.Actionable {
			actionable = append(actionable, o)
		} else {
			informational = append(informational, o)
		}
	}

	if len(actionable) > 0 {
		fmt.Println("═══ OVERLAP — ACTIONABLE (fee savings) ═══")
		fmt.Printf("%-8s %-8s %8s  %s\n", "SYMBOL", "SYMBOL", "CORR", "NOTE")
		fmt.Println(strings.Repeat("─", 80))
		for _, o := range actionable {
			fmt.Printf("%-8s %-8s %7.1f%%  %s\n", o.SymA, o.SymB, o.Correlation*100, o.Note)
		}
		fmt.Println()
	}

	if len(informational) > 0 {
		fmt.Println("═══ OVERLAP — INFORMATIONAL (no fee impact) ═══")
		fmt.Printf("%-8s %-8s %8s  %s\n", "SYMBOL", "SYMBOL", "CORR", "NOTE")
		fmt.Println(strings.Repeat("─", 80))
		for _, o := range informational {
			fmt.Printf("%-8s %-8s %7.1f%%  %s\n", o.SymA, o.SymB, o.Correlation*100, o.Note)
		}
		fmt.Println()
	}
}

func loadCoreTiltConfig(path string) (*coreTiltConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read core-tilts config: %w", err)
	}
	var cfg coreTiltConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse core-tilts config: %w", err)
	}
	return &cfg, nil
}
