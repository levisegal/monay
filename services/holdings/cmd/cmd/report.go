package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/spf13/cobra"
	"github.com/xuri/excelize/v2"
	"gopkg.in/yaml.v3"

	"github.com/levisegal/monay/services/holdings/config"
	"github.com/levisegal/monay/services/holdings/database"
	"github.com/levisegal/monay/services/holdings/gen/db"
)

func reportCommand() *cobra.Command {
	var output string
	var convictionsFlag string

	cmd := &cobra.Command{
		Use:   "report",
		Short: "Generate XLSX portfolio report",
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

			convPath := convictionsFlag
			if convPath == "" {
				convPath = cfg.ConvictionsPath
			}

			queries := db.New(conn)
			return generateReport(ctx, queries, cfg.PortfolioURL, output, convPath)
		},
	}

	cmd.Flags().StringVarP(&output, "output", "o", "", "Output file path (required)")
	cmd.MarkFlagRequired("output")
	cmd.Flags().StringVar(&convictionsFlag, "convictions", "", "Path to convictions YAML file")

	return cmd
}

type reportData struct {
	allHoldings       []db.ListAllHoldingsRow
	positions         []db.ListPositionsRow
	accounts          []db.Account
	cashByAcct        map[string]int64
	quotes            map[string]quoteData
	totalValue        float64
	taxableLots       []db.ListOpenLotsByAccountTypeRow
	realizedGL        db.SumRealizedGainLossByYearRow
	dispositions      []db.ListDispositionsByYearRow
	analysis          *analyzeResponse
	actualPerformance *actualPerformanceReport
	convictions       *convictionsConfig
}

type actualPerformanceReport struct {
	StartDate     string
	AsOf          string
	Portfolio     performanceRow
	Accounts      []performanceRow
	Monthly       []performanceMonthlyRow
	CoverageNotes []string
}

type performanceRow struct {
	Name             string
	AccountType      string
	StartValue       float64
	NetExternalFlows float64
	EndValue         float64
	OneMonth         *float64
	ThreeMonth       *float64
	SixMonth         *float64
	YTD              *float64
	OneYear          *float64
}

type performanceMonthlyRow struct {
	Period    string
	ReturnPct float64
}

type dailySeriesPoint struct {
	Date         string
	Value        float64
	ExternalFlow float64
	Return       *float64
}

type accountPerformanceSeries struct {
	Row    performanceRow
	Series []dailySeriesPoint
}

type performanceChartPoint struct {
	Timestamp string   `json:"timestamp"`
	Close     *float64 `json:"close"`
}

type analyzeRequest struct {
	Holdings    []analyzeHolding    `json:"holdings"`
	TotalValue  float64             `json:"total_value"`
	Constraints analyzeConstraints  `json:"constraints"`
	Convictions *analyzeConvictions `json:"convictions,omitempty"`
}

type analyzeHolding struct {
	Symbol        string  `json:"symbol"`
	Value         float64 `json:"value"`
	Sector        string  `json:"sector"`
	AssetClass    string  `json:"asset_class"`
	AccountType   string  `json:"account_type"`
	IsMuni        bool    `json:"is_muni"`
	SecurityType  string  `json:"security_type"`
	DividendYield float64 `json:"dividend_yield"`
}

type analyzeConstraints struct {
	MaxPositionWeight      float64    `json:"max_position_weight"`
	MaxSectorWeight        float64    `json:"max_sector_weight"`
	TargetEquityRange      [2]float64 `json:"target_equity_range"`
	TargetFixedIncomeRange [2]float64 `json:"target_fixed_income_range"`
	TargetCashRange        [2]float64 `json:"target_cash_range"`
}

type analyzeResponse struct {
	RiskMetrics          riskMetrics           `json:"risk_metrics"`
	TargetRiskMetrics    riskMetrics           `json:"target_risk_metrics"`
	Performance          *performanceSummary   `json:"performance"`
	CorrelationClusters  []correlationCluster  `json:"correlation_clusters"`
	ConstraintViolations []constraintViolation `json:"constraint_violations"`
	LocationIssues       []locationIssue       `json:"location_issues"`
	TargetWeights        map[string]float64    `json:"target_weights"`
	RebalanceDeltas      []rebalanceDelta      `json:"rebalance_deltas"`
	ConvictionStatuses   []convictionStatus    `json:"conviction_statuses"`
	ConsolidationRecs    []consolidationRec    `json:"consolidation_recs"`
}

type riskMetrics struct {
	AnnualizedVolatility *float64 `json:"annualized_volatility"`
	CVaR95               *float64 `json:"cvar_95"`
	MaxDrawdown          *float64 `json:"max_drawdown"`
	SharpeRatio          *float64 `json:"sharpe_ratio"`
}

type performanceSummary struct {
	AsOf           string              `json:"as_of"`
	HistoryStart   string              `json:"history_start"`
	Current        trailingReturns     `json:"current"`
	Target         trailingReturns     `json:"target"`
	MonthlyReturns []monthlyReturn     `json:"monthly_returns"`
	Symbols        []symbolPerformance `json:"symbols"`
}

type trailingReturns struct {
	OneMonth          *float64 `json:"one_month"`
	ThreeMonth        *float64 `json:"three_month"`
	SixMonth          *float64 `json:"six_month"`
	YTD               *float64 `json:"ytd"`
	OneYear           *float64 `json:"one_year"`
	TwoYearAnnualized *float64 `json:"two_year_annualized"`
}

type monthlyReturn struct {
	Period    string  `json:"period"`
	ReturnPct float64 `json:"return_pct"`
}

type symbolPerformance struct {
	Symbol              string   `json:"symbol"`
	Weight              float64  `json:"weight"`
	OneMonth            *float64 `json:"one_month"`
	ThreeMonth          *float64 `json:"three_month"`
	YTD                 *float64 `json:"ytd"`
	OneYear             *float64 `json:"one_year"`
	ContributionOneYear *float64 `json:"contribution_one_year"`
}

type correlationCluster struct {
	Symbols        []string `json:"symbols"`
	AvgCorrelation float64  `json:"avg_correlation"`
	CombinedWeight float64  `json:"combined_weight"`
}

type constraintViolation struct {
	ViolationType string  `json:"violation_type"`
	Name          string  `json:"name"`
	Current       float64 `json:"current"`
	Limit         float64 `json:"limit"`
	Excess        float64 `json:"excess"`
}

type locationIssue struct {
	Symbol                 string  `json:"symbol"`
	Name                   string  `json:"name"`
	CurrentAccountType     string  `json:"current_account_type"`
	RecommendedAccountType string  `json:"recommended_account_type"`
	Reason                 string  `json:"reason"`
	Value                  float64 `json:"value"`
}

type rebalanceDelta struct {
	Symbol        string  `json:"symbol"`
	Name          string  `json:"name"`
	Sector        string  `json:"sector"`
	CurrentWeight float64 `json:"current_weight"`
	TargetWeight  float64 `json:"target_weight"`
	DeltaDollars  float64 `json:"delta_dollars"`
	Action        string  `json:"action"`
	BestAccount   string  `json:"best_account"`
	Note          string  `json:"note"`
}

type convictionStatus struct {
	Type          string  `json:"type"`
	Name          string  `json:"name"`
	Thesis        string  `json:"thesis"`
	TargetWeight  float64 `json:"target_weight"`
	CurrentWeight float64 `json:"current_weight"`
	Status        string  `json:"status"`
}

type consolidationRec struct {
	Symbol        string  `json:"symbol"`
	Name          string  `json:"name"`
	CurrentWeight float64 `json:"current_weight"`
	Reason        string  `json:"reason"`
	MergeInto     string  `json:"merge_into"`
}

type convictionsConfig struct {
	Positions    map[string]positionConviction `yaml:"positions"`
	Sectors      map[string]sectorConviction   `yaml:"sectors"`
	Strategies   map[string]strategyConviction `yaml:"strategies"`
	AssetClasses map[string]assetClassOverride `yaml:"asset_classes"`
}

type positionConviction struct {
	Thesis string     `yaml:"thesis"`
	Range  [2]float64 `yaml:"range"`
}

type sectorConviction struct {
	Thesis       string   `yaml:"thesis"`
	TargetWeight float64  `yaml:"target_weight"`
	Instruments  []string `yaml:"instruments"`
}

type strategyConviction struct {
	Thesis       string   `yaml:"thesis"`
	TargetWeight float64  `yaml:"target_weight"`
	MinYield     float64  `yaml:"min_yield"`
	Instruments  []string `yaml:"instruments"`
	Location     string   `yaml:"location"`
}

type assetClassOverride struct {
	Thesis      string     `yaml:"thesis"`
	TargetRange [2]float64 `yaml:"target_range"`
}

type analyzeConvictions struct {
	Positions    map[string]analyzePositionConviction `json:"positions,omitempty"`
	Sectors      map[string]analyzeSectorConviction   `json:"sectors,omitempty"`
	Strategies   map[string]analyzeStrategyConviction `json:"strategies,omitempty"`
	AssetClasses map[string]analyzeAssetClassOverride `json:"asset_classes,omitempty"`
}

type analyzePositionConviction struct {
	Thesis string     `json:"thesis"`
	Range  [2]float64 `json:"range"`
}

type analyzeSectorConviction struct {
	Thesis       string   `json:"thesis"`
	TargetWeight float64  `json:"target_weight"`
	Instruments  []string `json:"instruments,omitempty"`
}

type analyzeStrategyConviction struct {
	Thesis       string   `json:"thesis"`
	TargetWeight float64  `json:"target_weight"`
	MinYield     float64  `json:"min_yield,omitempty"`
	Instruments  []string `json:"instruments,omitempty"`
	Location     string   `json:"location,omitempty"`
}

type analyzeAssetClassOverride struct {
	Thesis      string     `json:"thesis"`
	TargetRange [2]float64 `json:"target_range"`
}

func generateReport(ctx context.Context, queries *db.Queries, portfolioURL, output, convictionsPath string) error {
	var conv *convictionsConfig
	if convictionsPath != "" {
		var err error
		conv, err = loadConvictions(convictionsPath)
		if err != nil {
			return fmt.Errorf("failed to load convictions: %w", err)
		}
		slog.Info("loaded convictions", "path", convictionsPath,
			"positions", len(conv.Positions),
			"sectors", len(conv.Sectors),
			"strategies", len(conv.Strategies),
			"asset_classes", len(conv.AssetClasses),
		)
	}

	data, err := loadReportData(ctx, queries, portfolioURL, conv)
	if err != nil {
		return err
	}
	if data.analysis == nil {
		return fmt.Errorf("portfolio analysis unavailable from %s", portfolioURL)
	}

	f := excelize.NewFile()
	defer f.Close()

	currencyFmt, _ := f.NewStyle(&excelize.Style{NumFmt: 4})
	pctFmt, _ := f.NewStyle(&excelize.Style{NumFmt: 10})
	headerStyle, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true}})

	writeHoldingsSheet(f, data, currencyFmt, pctFmt, headerStyle)
	writeHoldingsByAccountSheet(f, data, currencyFmt, pctFmt, headerStyle)
	writeAccountsSheet(f, data, currencyFmt, pctFmt, headerStyle)
	writeAssetAllocationSheet(f, data, currencyFmt, pctFmt, headerStyle)
	writeSectorAllocationSheet(f, data, currencyFmt, pctFmt, headerStyle)
	writeConcentrationSheet(f, data, currencyFmt, pctFmt, headerStyle)
	writePerformanceSheet(f, data, currencyFmt, pctFmt, headerStyle)
	writePlaybookSheet(f, data, currencyFmt, pctFmt, headerStyle)
	writeTaxHarvestingSheet(f, data, currencyFmt, pctFmt, headerStyle)
	writePlaybookDetailSheet(f, data, currencyFmt, pctFmt, headerStyle)

	f.DeleteSheet("Sheet1")

	if err := f.SaveAs(output); err != nil {
		return fmt.Errorf("failed to save report: %w", err)
	}

	slog.Info("report generated", "output", output)
	return nil
}

func loadReportData(ctx context.Context, queries *db.Queries, portfolioURL string, conv *convictionsConfig) (*reportData, error) {
	allHoldings, err := queries.ListAllHoldings(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list holdings: %w", err)
	}

	positions, err := queries.ListPositions(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list positions: %w", err)
	}

	accounts, err := queries.ListAccounts(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list accounts: %w", err)
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

	taxableLots, err := queries.ListOpenLotsByAccountType(ctx, "taxable")
	if err != nil {
		return nil, fmt.Errorf("failed to list taxable lots: %w", err)
	}

	year := fmt.Sprintf("%d", time.Now().Year())
	realizedGL, err := queries.SumRealizedGainLossByYear(ctx, year)
	if err != nil {
		return nil, fmt.Errorf("failed to sum realized gains: %w", err)
	}

	dispositions, err := queries.ListDispositionsByYear(ctx, year)
	if err != nil {
		return nil, fmt.Errorf("failed to list dispositions: %w", err)
	}

	acctMap := make(map[string]db.Account)
	for _, a := range accounts {
		acctMap[a.ID] = a
	}
	analysis := callPortfolioAnalyze(ctx, portfolioURL, allHoldings, acctMap, cashByAcct, quotes, totalValue, conv)
	actualPerformance, err := buildActualPerformance(ctx, queries, portfolioURL, allHoldings, accounts, cashByAcct, quotes)
	if err != nil {
		return nil, fmt.Errorf("failed to build performance report: %w", err)
	}

	return &reportData{
		allHoldings:       allHoldings,
		positions:         positions,
		accounts:          accounts,
		cashByAcct:        cashByAcct,
		quotes:            quotes,
		totalValue:        totalValue,
		taxableLots:       taxableLots,
		realizedGL:        realizedGL,
		dispositions:      dispositions,
		analysis:          analysis,
		actualPerformance: actualPerformance,
		convictions:       conv,
	}, nil
}

func callPortfolioAnalyze(ctx context.Context, portfolioURL string, holdings []db.ListAllHoldingsRow, acctMap map[string]db.Account, cashByAcct map[string]int64, quotes map[string]quoteData, totalValue float64, conv *convictionsConfig) *analyzeResponse {
	if totalValue <= 0 {
		return nil
	}

	var holdingInputs []analyzeHolding
	for _, h := range holdings {
		acct := acctMap[h.AccountID]
		qty := nullFloat64ToFloat(h.QuantityMicros) / 1_000_000
		cost := nullFloat64ToFloat(h.CostBasisMicros) / 1_000_000
		mv := marketValue(h.Symbol, h.SecurityType.String, qty, cost, quotes)
		qd := quotes[h.Symbol]
		name := bestName(h.SecurityName.String, qd.Name)
		sector := resolveSector(qd, h.SecurityType.String, h.Symbol, name)
		bucket := sectorToAssetBucket(sector, h.SecurityType.String)

		isMuni := sector == "Municipal Bonds"
		assetClass := "equity"
		switch bucket {
		case "Fixed Income":
			assetClass = "fixed_income"
		case "Cash & Equivalents":
			assetClass = "cash"
		case "Alternatives":
			assetClass = "alternatives"
		case "Real Estate":
			assetClass = "real_estate"
		case "Crypto":
			assetClass = "crypto"
		}

		holdingInputs = append(holdingInputs, analyzeHolding{
			Symbol:        h.Symbol,
			Value:         mv,
			Sector:        sector,
			AssetClass:    assetClass,
			AccountType:   acct.AccountType,
			IsMuni:        isMuni,
			SecurityType:  h.SecurityType.String,
			DividendYield: qd.DividendYield,
		})
	}

	req := analyzeRequest{
		Holdings:   holdingInputs,
		TotalValue: totalValue,
		Constraints: analyzeConstraints{
			MaxPositionWeight:      positionConcentrationLimit,
			MaxSectorWeight:        0.15,
			TargetEquityRange:      [2]float64{0.30, 0.40},
			TargetFixedIncomeRange: [2]float64{0.40, 0.55},
			TargetCashRange:        [2]float64{0.05, 0.15},
		},
		Convictions: toAnalyzeConvictions(conv),
	}

	body, err := json.Marshal(req)
	if err != nil {
		slog.Warn("failed to marshal analyze request", "error", err)
		return nil
	}

	url := fmt.Sprintf("%s/api/v1/portfolio/analyze", portfolioURL)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		slog.Warn("failed to create analyze request", "error", err)
		return nil
	}
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		slog.Warn("portfolio analyze unavailable", "error", err)
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		slog.Warn("portfolio analyze failed", "status", resp.StatusCode)
		return nil
	}

	var result analyzeResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		slog.Warn("failed to decode analyze response", "error", err)
		return nil
	}

	slog.Info("portfolio analysis complete",
		"violations", len(result.ConstraintViolations),
		"location_issues", len(result.LocationIssues),
		"rebalance_deltas", len(result.RebalanceDeltas),
		"conviction_statuses", len(result.ConvictionStatuses),
		"consolidation_recs", len(result.ConsolidationRecs),
	)
	return &result
}

func loadConvictions(path string) (*convictionsConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg convictionsConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func toAnalyzeConvictions(cfg *convictionsConfig) *analyzeConvictions {
	if cfg == nil {
		return nil
	}

	ac := &analyzeConvictions{
		Positions:    make(map[string]analyzePositionConviction, len(cfg.Positions)),
		Sectors:      make(map[string]analyzeSectorConviction, len(cfg.Sectors)),
		Strategies:   make(map[string]analyzeStrategyConviction, len(cfg.Strategies)),
		AssetClasses: make(map[string]analyzeAssetClassOverride, len(cfg.AssetClasses)),
	}

	for sym, pc := range cfg.Positions {
		ac.Positions[sym] = analyzePositionConviction{
			Thesis: pc.Thesis,
			Range:  pc.Range,
		}
	}
	for name, sc := range cfg.Sectors {
		ac.Sectors[name] = analyzeSectorConviction{
			Thesis:       sc.Thesis,
			TargetWeight: sc.TargetWeight,
			Instruments:  sc.Instruments,
		}
	}
	for name, st := range cfg.Strategies {
		ac.Strategies[name] = analyzeStrategyConviction{
			Thesis:       st.Thesis,
			TargetWeight: st.TargetWeight,
			MinYield:     st.MinYield,
			Instruments:  st.Instruments,
			Location:     st.Location,
		}
	}
	for name, aco := range cfg.AssetClasses {
		ac.AssetClasses[name] = analyzeAssetClassOverride{
			Thesis:      aco.Thesis,
			TargetRange: aco.TargetRange,
		}
	}

	return ac
}

func computeTotalValue(positions []db.ListPositionsRow, accounts []db.Account, cashByAcct map[string]int64, quotes map[string]quoteData, bondSymbols map[string]bool) float64 {
	var total float64
	for _, p := range positions {
		total += marketValue(p.Symbol, p.SecurityType.String, nullFloat64ToFloat(p.QuantityMicros)/1_000_000, nullFloat64ToFloat(p.CostBasisMicros)/1_000_000, quotes)
	}
	for _, a := range accounts {
		total += float64(cashByAcct[a.ID]) / 1_000_000
	}
	return total
}

func marketValue(symbol, secType string, qty, costBasis float64, quotes map[string]quoteData) float64 {
	if qd, ok := quotes[symbol]; ok {
		return qty * qd.Price
	}
	if secType == "bond" {
		return qty
	}
	return costBasis
}

func bestName(dbName, quoteName string) string {
	if dbName != "" {
		return dbName
	}
	return quoteName
}

var symbolSectorOverrides = map[string][2]string{
	"FSELX":     {"Technology", "Semiconductors"},
	"MANLX":     {"Municipal Bonds", "National Intermediate"},
	"PDBZX":     {"Corporate Bonds", "Core-Plus Bond"},
	"MDIJX":     {"International Equity", "Diversified International"},
	"GSFTX":     {"Large Cap Equity", "Dividend Income"},
	"MBXIX":     {"Alternatives", "Systematic Macro"},
	"FGSAX":     {"Mid Cap Equity", "Mid-Cap Growth"},
	"MIVIX":     {"Large Cap Equity", "Large-Cap Value"},
	"IBDW":      {"Corporate Bonds", "Target Maturity"},
	"XLY":       {"Consumer Discretionary", "Consumer Discretionary Select"},
	"EKWAX":     {"Basic Materials", "Precious Metals"},
	"AVALX":     {"Small Cap Equity", "Small-Cap Value"},
	"FSPCX":     {"Financial Services", "Insurance"},
	"FXAIX":     {"Large Cap Equity", "Large Blend"},
	"TMCXX":     {"Cash & Equivalents", "Money Market"},
	"FTXSX":     {"Small Cap Equity", "Small-Cap Growth"},
	"PMJPX":     {"Small Cap Equity", "Small-Cap Blend"},
	"437355100": {"Large Cap Equity", "Home Improvement Retail"},
	"FSPTX":     {"Technology", "Technology"},
	"FDGRX":     {"Large Cap Equity", "Large-Cap Growth"},
}

func resolveSector(qd quoteData, secType, symbol, name string) string {
	if override, ok := symbolSectorOverrides[symbol]; ok {
		return override[0]
	}

	if qd.Sector != "" {
		s := qd.Sector
		if s == "Consumer Cyclical" {
			return "Consumer Discretionary"
		}
		if s == "Consumer Defensive" {
			return "Consumer Staples"
		}
		return s
	}

	if secType == "bond" {
		upper := strings.ToUpper(name)
		if strings.Contains(upper, "MUNI") || strings.Contains(upper, "MUNICIPAL") {
			return "Municipal Bonds"
		}
		return "Government Bonds"
	}
	if secType == "crypto" {
		return "Crypto"
	}
	if secType == "reit" {
		return "Real Estate"
	}
	if secType == "cash" {
		return "Cash & Equivalents"
	}

	if qd.Category != "" {
		return categoryToSector(qd.Category)
	}

	upper := strings.ToUpper(name)
	if strings.Contains(upper, "MONEY MARKET") || strings.Contains(upper, "MMKT") ||
		strings.Contains(upper, "CASH RESERVE") || name == "CASH" {
		return "Cash & Equivalents"
	}
	if strings.Contains(upper, "MUNI") || strings.Contains(upper, "MUNICIPAL") {
		return "Municipal Bonds"
	}
	if strings.Contains(upper, "BOND") || strings.Contains(upper, "FIXED INCOME") ||
		strings.Contains(upper, "TOTAL RETURN") || strings.Contains(upper, "IBONDS") {
		return "Corporate Bonds"
	}

	return "Other"
}

func resolveIndustry(qd quoteData, secType, symbol, name string) string {
	if override, ok := symbolSectorOverrides[symbol]; ok {
		return override[1]
	}
	if qd.Industry != "" {
		return qd.Industry
	}
	if qd.Category != "" {
		return qd.Category
	}
	if secType == "bond" {
		upper := strings.ToUpper(name)
		if strings.Contains(upper, "MUNI") || strings.Contains(upper, "MUNICIPAL") {
			return "Municipal Bond"
		}
		return "Fixed Income"
	}
	if secType == "crypto" {
		return "Digital Assets"
	}
	if secType == "reit" {
		return "Real Estate Investment Trust"
	}
	upper := strings.ToUpper(name)
	if strings.Contains(upper, "MONEY MARKET") || strings.Contains(upper, "MMKT") ||
		strings.Contains(upper, "CASH RESERVE") || strings.Contains(upper, "BANK DEPOSIT") ||
		strings.Contains(upper, "LIQUIDITY") {
		return "Money Market"
	}
	if symbol == "CASH" || symbol == "FCASH" || upper == "CASH" {
		return "Cash"
	}
	return ""
}

func categoryToSector(category string) string {
	upper := strings.ToUpper(category)

	if strings.Contains(upper, "MONEY MARKET") {
		return "Cash & Equivalents"
	}

	if strings.Contains(upper, "MUNI") {
		return "Municipal Bonds"
	}
	if strings.Contains(upper, "CORPORATE") || strings.Contains(upper, "HIGH YIELD") ||
		strings.Contains(upper, "TARGET MATURITY") || strings.Contains(upper, "CORE-PLUS") ||
		strings.Contains(upper, "CORE PLUS") {
		return "Corporate Bonds"
	}
	if strings.Contains(upper, "GOVERNMENT") || strings.Contains(upper, "TREASURY") ||
		strings.Contains(upper, "INFLATION") {
		return "Government Bonds"
	}
	if strings.Contains(upper, "BOND") || strings.Contains(upper, "FIXED") ||
		strings.Contains(upper, "INCOME") {
		return "Corporate Bonds"
	}

	if strings.Contains(upper, "TECHNOLOGY") || strings.Contains(upper, "SEMICONDUCTOR") {
		return "Technology"
	}
	if strings.Contains(upper, "HEALTH") {
		return "Healthcare"
	}
	if strings.Contains(upper, "FINANCIAL") || strings.Contains(upper, "INSURANCE") {
		return "Financial Services"
	}
	if strings.Contains(upper, "REAL ESTATE") {
		return "Real Estate"
	}
	if strings.Contains(upper, "ENERGY") {
		return "Energy"
	}
	if strings.Contains(upper, "PRECIOUS METAL") || strings.Contains(upper, "NATURAL RES") {
		return "Basic Materials"
	}
	if strings.Contains(upper, "UTILITIES") {
		return "Utilities"
	}
	if strings.Contains(upper, "INDUSTRIALS") || strings.Contains(upper, "INFRASTRUCTURE") {
		return "Industrials"
	}
	if strings.Contains(upper, "CONSUMER DEFENSIVE") || strings.Contains(upper, "CONSUMER STAPLE") {
		return "Consumer Staples"
	}
	if strings.Contains(upper, "CONSUMER") {
		return "Consumer Discretionary"
	}
	if strings.Contains(upper, "COMMUNICATION") {
		return "Communication Services"
	}

	if strings.Contains(upper, "ALTERNATIVE") || strings.Contains(upper, "HEDGE") ||
		strings.Contains(upper, "LONG-SHORT") || strings.Contains(upper, "MANAGED FUTURES") ||
		strings.Contains(upper, "MACRO") || strings.Contains(upper, "MARKET NEUTRAL") {
		return "Alternatives"
	}

	if strings.Contains(upper, "INTERNATIONAL") || strings.Contains(upper, "FOREIGN") ||
		strings.Contains(upper, "WORLD") || strings.Contains(upper, "GLOBAL") {
		return "International Equity"
	}
	if strings.Contains(upper, "SMALL") {
		return "Small Cap Equity"
	}
	if strings.Contains(upper, "MID") {
		return "Mid Cap Equity"
	}
	if strings.Contains(upper, "LARGE") {
		return "Large Cap Equity"
	}
	if strings.Contains(upper, "BLEND") || strings.Contains(upper, "GROWTH") ||
		strings.Contains(upper, "VALUE") || strings.Contains(upper, "EQUITY") ||
		strings.Contains(upper, "STOCK") || strings.Contains(upper, "DIVERSIFIED") {
		return "Diversified Equity"
	}

	return "Other"
}

func writeHoldingsSheet(f *excelize.File, data *reportData, currencyFmt, pctFmt, headerStyle int) {
	sheet := "Holdings"
	f.NewSheet(sheet)

	headers := []string{"Symbol", "Name", "Sector", "Industry", "Qty", "Cost Basis", "Mkt Value", "Gain/Loss", "% of Portfolio"}
	writeHeaderRow(f, sheet, headers, headerStyle)

	type holdingRow struct {
		symbol   string
		name     string
		sector   string
		industry string
		qty      float64
		cost     float64
		mktVal   float64
	}

	var allRows []holdingRow
	for _, p := range data.positions {
		qty := nullFloat64ToFloat(p.QuantityMicros) / 1_000_000
		cost := nullFloat64ToFloat(p.CostBasisMicros) / 1_000_000
		mktVal := marketValue(p.Symbol, p.SecurityType.String, qty, cost, data.quotes)
		qd := data.quotes[p.Symbol]
		name := bestName(p.SecurityName.String, qd.Name)
		allRows = append(allRows, holdingRow{
			symbol:   p.Symbol,
			name:     name,
			sector:   resolveSector(qd, p.SecurityType.String, p.Symbol, name),
			industry: resolveIndustry(qd, p.SecurityType.String, p.Symbol, name),
			qty:      qty,
			cost:     cost,
			mktVal:   mktVal,
		})
	}

	var totalCash float64
	for _, a := range data.accounts {
		totalCash += float64(data.cashByAcct[a.ID]) / 1_000_000
	}
	allRows = append(allRows, holdingRow{
		symbol: "CASH",
		name:   "Account Cash Balances",
		sector: "Cash & Equivalents",
		mktVal: totalCash,
	})

	sort.Slice(allRows, func(i, j int) bool {
		return allRows[i].mktVal > allRows[j].mktVal
	})

	totalRow := 2 + len(allRows)

	row := 2
	for _, hr := range allRows {
		f.SetCellValue(sheet, cell("A", row), hr.symbol)
		f.SetCellValue(sheet, cell("B", row), hr.name)
		f.SetCellValue(sheet, cell("C", row), hr.sector)
		f.SetCellValue(sheet, cell("D", row), hr.industry)
		if hr.qty != 0 {
			f.SetCellValue(sheet, cell("E", row), hr.qty)
		}
		if hr.cost != 0 {
			setCurrency(f, sheet, "F", row, hr.cost, currencyFmt)
		}
		setCurrency(f, sheet, "G", row, hr.mktVal, currencyFmt)
		if hr.cost != 0 {
			setCurrencyFormula(f, sheet, "H", row, fmt.Sprintf("G%d-F%d", row, row), currencyFmt)
		}
		setPercentFormula(f, sheet, "I", row, fmt.Sprintf("G%d/G%d", row, totalRow), pctFmt)
		row++
	}

	last := totalRow - 1
	f.SetCellValue(sheet, cell("A", totalRow), "TOTAL")
	f.SetCellStyle(sheet, cell("A", totalRow), cell("A", totalRow), headerStyle)
	setCurrencyFormula(f, sheet, "F", totalRow, fmt.Sprintf("SUM(F2:F%d)", last), currencyFmt)
	setCurrencyFormula(f, sheet, "G", totalRow, fmt.Sprintf("SUM(G2:G%d)", last), currencyFmt)
	setCurrencyFormula(f, sheet, "H", totalRow, fmt.Sprintf("G%d-F%d", totalRow, totalRow), currencyFmt)
	setPercent(f, sheet, "I", totalRow, 1.0, pctFmt)
}

func writeHoldingsByAccountSheet(f *excelize.File, data *reportData, currencyFmt, pctFmt, headerStyle int) {
	sheet := "Holdings by Account"
	f.NewSheet(sheet)

	headers := []string{"Account", "Broker", "Symbol", "Name", "Qty", "Cost Basis", "Mkt Value", "Gain/Loss", "% of Portfolio"}
	writeHeaderRow(f, sheet, headers, headerStyle)

	cashAcctCount := 0
	for _, a := range data.accounts {
		if data.cashByAcct[a.ID] != 0 {
			cashAcctCount++
		}
	}
	totalRow := 2 + len(data.allHoldings) + cashAcctCount

	row := 2
	for _, h := range data.allHoldings {
		qty := nullFloat64ToFloat(h.QuantityMicros) / 1_000_000
		cost := nullFloat64ToFloat(h.CostBasisMicros) / 1_000_000
		mktVal := marketValue(h.Symbol, h.SecurityType.String, qty, cost, data.quotes)

		qd := data.quotes[h.Symbol]
		name := bestName(h.SecurityName.String, qd.Name)
		f.SetCellValue(sheet, cell("A", row), h.AccountName)
		f.SetCellValue(sheet, cell("B", row), h.Broker)
		f.SetCellValue(sheet, cell("C", row), h.Symbol)
		f.SetCellValue(sheet, cell("D", row), name)
		f.SetCellValue(sheet, cell("E", row), qty)
		setCurrency(f, sheet, "F", row, cost, currencyFmt)
		setCurrency(f, sheet, "G", row, mktVal, currencyFmt)
		setCurrencyFormula(f, sheet, "H", row, fmt.Sprintf("G%d-F%d", row, row), currencyFmt)
		setPercentFormula(f, sheet, "I", row, fmt.Sprintf("G%d/G%d", row, totalRow), pctFmt)
		row++
	}

	for _, a := range data.accounts {
		cash := float64(data.cashByAcct[a.ID]) / 1_000_000
		if cash == 0 {
			continue
		}
		f.SetCellValue(sheet, cell("A", row), a.Name)
		f.SetCellValue(sheet, cell("B", row), a.InstitutionName)
		f.SetCellValue(sheet, cell("C", row), "CASH")
		f.SetCellValue(sheet, cell("D", row), "Cash Balance")
		setCurrency(f, sheet, "G", row, cash, currencyFmt)
		setPercentFormula(f, sheet, "I", row, fmt.Sprintf("G%d/G%d", row, totalRow), pctFmt)
		row++
	}

	last := totalRow - 1
	f.SetCellValue(sheet, cell("A", totalRow), "TOTAL")
	f.SetCellStyle(sheet, cell("A", totalRow), cell("A", totalRow), headerStyle)
	setCurrencyFormula(f, sheet, "F", totalRow, fmt.Sprintf("SUM(F2:F%d)", last), currencyFmt)
	setCurrencyFormula(f, sheet, "G", totalRow, fmt.Sprintf("SUM(G2:G%d)", last), currencyFmt)
	setCurrencyFormula(f, sheet, "H", totalRow, fmt.Sprintf("G%d-F%d", totalRow, totalRow), currencyFmt)
	setPercent(f, sheet, "I", totalRow, 1.0, pctFmt)
}

func writeAccountsSheet(f *excelize.File, data *reportData, currencyFmt, pctFmt, headerStyle int) {
	sheet := "Accounts"
	f.NewSheet(sheet)

	headers := []string{"Account", "Broker", "Positions Value", "Cash", "Total", "% of Portfolio"}
	writeHeaderRow(f, sheet, headers, headerStyle)

	posValueByAcct := make(map[string]float64)
	for _, h := range data.allHoldings {
		qty := nullFloat64ToFloat(h.QuantityMicros) / 1_000_000
		cost := nullFloat64ToFloat(h.CostBasisMicros) / 1_000_000
		posValueByAcct[h.AccountID] += marketValue(h.Symbol, h.SecurityType.String, qty, cost, data.quotes)
	}

	totalRow := 2 + len(data.accounts)

	row := 2
	for _, a := range data.accounts {
		posVal := posValueByAcct[a.ID]
		cash := float64(data.cashByAcct[a.ID]) / 1_000_000

		f.SetCellValue(sheet, cell("A", row), a.Name)
		f.SetCellValue(sheet, cell("B", row), a.InstitutionName)
		setCurrency(f, sheet, "C", row, posVal, currencyFmt)
		setCurrency(f, sheet, "D", row, cash, currencyFmt)
		setCurrencyFormula(f, sheet, "E", row, fmt.Sprintf("C%d+D%d", row, row), currencyFmt)
		setPercentFormula(f, sheet, "F", row, fmt.Sprintf("E%d/E%d", row, totalRow), pctFmt)
		row++
	}

	last := totalRow - 1
	f.SetCellValue(sheet, cell("A", totalRow), "TOTAL")
	f.SetCellStyle(sheet, cell("A", totalRow), cell("A", totalRow), headerStyle)
	setCurrencyFormula(f, sheet, "C", totalRow, fmt.Sprintf("SUM(C2:C%d)", last), currencyFmt)
	setCurrencyFormula(f, sheet, "D", totalRow, fmt.Sprintf("SUM(D2:D%d)", last), currencyFmt)
	setCurrencyFormula(f, sheet, "E", totalRow, fmt.Sprintf("C%d+D%d", totalRow, totalRow), currencyFmt)
	setPercent(f, sheet, "F", totalRow, 1.0, pctFmt)
}

func writeAssetAllocationSheet(f *excelize.File, data *reportData, currencyFmt, pctFmt, headerStyle int) {
	sheet := "Asset Allocation"
	f.NewSheet(sheet)

	headers := []string{"Type", "Market Value", "% of Portfolio"}
	writeHeaderRow(f, sheet, headers, headerStyle)

	alloc := make(map[string]float64)
	for _, p := range data.positions {
		qty := nullFloat64ToFloat(p.QuantityMicros) / 1_000_000
		cost := nullFloat64ToFloat(p.CostBasisMicros) / 1_000_000
		mktVal := marketValue(p.Symbol, p.SecurityType.String, qty, cost, data.quotes)
		qd := data.quotes[p.Symbol]
		name := bestName(p.SecurityName.String, qd.Name)
		sector := resolveSector(qd, p.SecurityType.String, p.Symbol, name)
		bucket := sectorToAssetBucket(sector, p.SecurityType.String)
		alloc[bucket] += mktVal
	}

	var totalCash float64
	for _, a := range data.accounts {
		totalCash += float64(data.cashByAcct[a.ID]) / 1_000_000
	}
	alloc["Cash"] += totalCash

	buckets := sortedKeys(alloc)
	totalRow := 2 + len(buckets)

	row := 2
	for _, bucket := range buckets {
		val := alloc[bucket]
		f.SetCellValue(sheet, cell("A", row), bucket)
		setCurrency(f, sheet, "B", row, val, currencyFmt)
		setPercentFormula(f, sheet, "C", row, fmt.Sprintf("B%d/B%d", row, totalRow), pctFmt)
		row++
	}

	last := totalRow - 1
	f.SetCellValue(sheet, cell("A", totalRow), "TOTAL")
	f.SetCellStyle(sheet, cell("A", totalRow), cell("A", totalRow), headerStyle)
	setCurrencyFormula(f, sheet, "B", totalRow, fmt.Sprintf("SUM(B2:B%d)", last), currencyFmt)
	setPercent(f, sheet, "C", totalRow, 1.0, pctFmt)
}

func writeSectorAllocationSheet(f *excelize.File, data *reportData, currencyFmt, pctFmt, headerStyle int) {
	sheet := "Sector Allocation"
	f.NewSheet(sheet)

	headers := []string{"Sector", "Market Value", "% of Portfolio"}
	writeHeaderRow(f, sheet, headers, headerStyle)

	sectors := make(map[string]float64)
	for _, p := range data.positions {
		qd := data.quotes[p.Symbol]
		qty := nullFloat64ToFloat(p.QuantityMicros) / 1_000_000
		cost := nullFloat64ToFloat(p.CostBasisMicros) / 1_000_000
		mktVal := marketValue(p.Symbol, p.SecurityType.String, qty, cost, data.quotes)
		name := bestName(p.SecurityName.String, qd.Name)
		sector := resolveSector(qd, p.SecurityType.String, p.Symbol, name)
		sectors[sector] += mktVal
	}

	var accountCash float64
	for _, a := range data.accounts {
		accountCash += float64(data.cashByAcct[a.ID]) / 1_000_000
	}
	sectors["Cash & Equivalents"] += accountCash

	type sectorEntry struct {
		name  string
		value float64
	}
	var entries []sectorEntry
	for name, val := range sectors {
		entries = append(entries, sectorEntry{name, val})
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].value > entries[j].value
	})

	totalRow := 2 + len(entries)

	row := 2
	for _, e := range entries {
		f.SetCellValue(sheet, cell("A", row), e.name)
		setCurrency(f, sheet, "B", row, e.value, currencyFmt)
		setPercentFormula(f, sheet, "C", row, fmt.Sprintf("B%d/B%d", row, totalRow), pctFmt)
		row++
	}

	last := totalRow - 1
	f.SetCellValue(sheet, cell("A", totalRow), "TOTAL")
	f.SetCellStyle(sheet, cell("A", totalRow), cell("A", totalRow), headerStyle)
	setCurrencyFormula(f, sheet, "B", totalRow, fmt.Sprintf("SUM(B2:B%d)", last), currencyFmt)
	setPercent(f, sheet, "C", totalRow, 1.0, pctFmt)
}

func writeConcentrationSheet(f *excelize.File, data *reportData, currencyFmt, pctFmt, headerStyle int) {
	sheet := "Concentration"
	f.NewSheet(sheet)

	acctMap := make(map[string]db.Account)
	for _, a := range data.accounts {
		acctMap[a.ID] = a
	}

	// Build position-level data with sector/asset bucket
	type posInfo struct {
		symbol      string
		name        string
		sector      string
		assetBucket string
		secType     string
		value       float64
	}
	var positions []posInfo
	for _, p := range data.positions {
		qty := nullFloat64ToFloat(p.QuantityMicros) / 1_000_000
		cost := nullFloat64ToFloat(p.CostBasisMicros) / 1_000_000
		mv := marketValue(p.Symbol, p.SecurityType.String, qty, cost, data.quotes)
		qd := data.quotes[p.Symbol]
		n := bestName(p.SecurityName.String, qd.Name)
		sector := resolveSector(qd, p.SecurityType.String, p.Symbol, n)
		bucket := sectorToAssetBucket(sector, p.SecurityType.String)
		positions = append(positions, posInfo{
			symbol:      p.Symbol,
			name:        n,
			sector:      sector,
			assetBucket: bucket,
			secType:     p.SecurityType.String,
			value:       mv,
		})
	}

	var totalCash float64
	for _, a := range data.accounts {
		totalCash += float64(data.cashByAcct[a.ID]) / 1_000_000
	}

	sort.Slice(positions, func(i, j int) bool { return positions[i].value > positions[j].value })

	row := 1

	// Section 1: Asset Allocation
	f.SetCellValue(sheet, cell("A", row), "Asset Allocation")
	f.SetCellStyle(sheet, cell("A", row), cell("A", row), headerStyle)
	row++

	writeSubHeader(f, sheet, row, []string{"Asset Class", "Market Value", "% of Portfolio", "# Positions"}, headerStyle)
	row++

	bucketValues := make(map[string]float64)
	bucketCounts := make(map[string]int)
	for _, p := range positions {
		bucketValues[p.assetBucket] += p.value
		bucketCounts[p.assetBucket]++
	}
	bucketValues["Cash & Equivalents"] += totalCash

	type weightEntry struct {
		name  string
		val   float64
		count int
	}

	bucketOrder := []string{"Equity", "Fixed Income", "Alternatives", "Real Estate", "Crypto", "Cash & Equivalents", "Other"}
	for _, name := range bucketOrder {
		val := bucketValues[name]
		if val == 0 {
			continue
		}
		f.SetCellValue(sheet, cell("A", row), name)
		setCurrency(f, sheet, "B", row, val, currencyFmt)
		setPercent(f, sheet, "C", row, val/data.totalValue, pctFmt)
		f.SetCellValue(sheet, cell("D", row), bucketCounts[name])
		row++
	}

	row++
	equityPct := bucketValues["Equity"] / data.totalValue
	fixedPct := (bucketValues["Fixed Income"] + bucketValues["Cash & Equivalents"]) / data.totalValue
	f.SetCellValue(sheet, cell("A", row), "Equity / Fixed+Cash Split")
	f.SetCellStyle(sheet, cell("A", row), cell("A", row), headerStyle)
	f.SetCellValue(sheet, cell("B", row), fmt.Sprintf("%.0f / %.0f", equityPct*100, fixedPct*100))
	row++

	ageBondRule := float64(birthYear+66) / 100.0
	_ = ageBondRule
	f.SetCellValue(sheet, cell("A", row), "Age-based guideline: ~66% bonds+cash for age 66")
	row++
	f.SetCellValue(sheet, cell("A", row), fmt.Sprintf("Current fixed+cash: %.0f%% — gap of %.0f%% to guideline", fixedPct*100, (0.66-fixedPct)*100))
	row++
	row++

	// Section 2: Account Type Distribution
	f.SetCellValue(sheet, cell("A", row), "Account Type Distribution")
	f.SetCellStyle(sheet, cell("A", row), cell("A", row), headerStyle)
	row++

	writeSubHeader(f, sheet, row, []string{"Account Type", "Market Value", "% of Portfolio"}, headerStyle)
	row++

	typeValues := make(map[string]float64)
	for _, h := range data.allHoldings {
		acct := acctMap[h.AccountID]
		qty := nullFloat64ToFloat(h.QuantityMicros) / 1_000_000
		cost := nullFloat64ToFloat(h.CostBasisMicros) / 1_000_000
		mv := marketValue(h.Symbol, h.SecurityType.String, qty, cost, data.quotes)
		typeValues[acct.AccountType] += mv
	}
	for _, a := range data.accounts {
		cash := float64(data.cashByAcct[a.ID]) / 1_000_000
		typeValues[a.AccountType] += cash
	}

	typeLabels := map[string]string{
		"taxable":         "Taxable",
		"traditional_ira": "Traditional IRA",
		"sep_ira":         "SEP IRA",
		"roth_ira":        "Roth IRA",
	}
	typeOrder := []string{"taxable", "traditional_ira", "sep_ira", "roth_ira"}
	for _, t := range typeOrder {
		val := typeValues[t]
		if val == 0 {
			continue
		}
		label := typeLabels[t]
		if label == "" {
			label = t
		}
		f.SetCellValue(sheet, cell("A", row), label)
		setCurrency(f, sheet, "B", row, val, currencyFmt)
		setPercent(f, sheet, "C", row, val/data.totalValue, pctFmt)
		row++
	}

	// Combine trad + SEP for total tax-deferred
	tradSEP := typeValues["traditional_ira"] + typeValues["sep_ira"]
	if tradSEP > 0 {
		row++
		f.SetCellValue(sheet, cell("A", row), "Total Tax-Deferred (Trad + SEP)")
		f.SetCellStyle(sheet, cell("A", row), cell("A", row), headerStyle)
		setCurrency(f, sheet, "B", row, tradSEP, currencyFmt)
		setPercent(f, sheet, "C", row, tradSEP/data.totalValue, pctFmt)
		row++
	}
	row++

	// Section 3: Sector Concentration
	f.SetCellValue(sheet, cell("A", row), "Sector Concentration")
	f.SetCellStyle(sheet, cell("A", row), cell("A", row), headerStyle)
	row++

	writeSubHeader(f, sheet, row, []string{"Sector", "Market Value", "% of Portfolio", "# Positions"}, headerStyle)
	row++

	sectorValues := make(map[string]float64)
	sectorCounts := make(map[string]int)
	for _, p := range positions {
		sectorValues[p.sector] += p.value
		sectorCounts[p.sector]++
	}
	sectorValues["Cash & Equivalents"] += totalCash

	type sectorEntry struct {
		name  string
		val   float64
		count int
	}
	var sectorEntries []sectorEntry
	for name, val := range sectorValues {
		sectorEntries = append(sectorEntries, sectorEntry{name, val, sectorCounts[name]})
	}
	sort.Slice(sectorEntries, func(i, j int) bool { return sectorEntries[i].val > sectorEntries[j].val })

	sectorFirstRow := row
	for _, se := range sectorEntries {
		f.SetCellValue(sheet, cell("A", row), se.name)
		setCurrency(f, sheet, "B", row, se.val, currencyFmt)
		setPercent(f, sheet, "C", row, se.val/data.totalValue, pctFmt)
		f.SetCellValue(sheet, cell("D", row), se.count)
		row++
	}
	sectorLastRow := row - 1

	f.SetCellValue(sheet, cell("A", row), "TOTAL")
	f.SetCellStyle(sheet, cell("A", row), cell("A", row), headerStyle)
	setCurrencyFormula(f, sheet, "B", row, fmt.Sprintf("SUM(B%d:B%d)", sectorFirstRow, sectorLastRow), currencyFmt)
	setPercent(f, sheet, "C", row, 1.0, pctFmt)
	row++
	row++

	// Section 4: Top Individual Positions
	f.SetCellValue(sheet, cell("A", row), "Top Individual Positions")
	f.SetCellStyle(sheet, cell("A", row), cell("A", row), headerStyle)
	row++

	writeSubHeader(f, sheet, row, []string{"Symbol", "Name", "Sector", "Market Value", "% of Portfolio"}, headerStyle)
	row++

	topN := 15
	if len(positions) < topN {
		topN = len(positions)
	}
	topFirstRow := row
	for i := 0; i < topN; i++ {
		p := positions[i]
		f.SetCellValue(sheet, cell("A", row), p.symbol)
		f.SetCellValue(sheet, cell("B", row), p.name)
		f.SetCellValue(sheet, cell("C", row), p.sector)
		setCurrency(f, sheet, "D", row, p.value, currencyFmt)
		setPercent(f, sheet, "E", row, p.value/data.totalValue, pctFmt)
		row++
	}
	topLastRow := row - 1

	f.SetCellValue(sheet, cell("A", row), fmt.Sprintf("Top %d Total", topN))
	f.SetCellStyle(sheet, cell("A", row), cell("A", row), headerStyle)
	setCurrencyFormula(f, sheet, "D", row, fmt.Sprintf("SUM(D%d:D%d)", topFirstRow, topLastRow), currencyFmt)
	setPercentFormula(f, sheet, "E", row, fmt.Sprintf("D%d/%g", row, data.totalValue), pctFmt)
	row++
	row++

	// Section 5: Correlated Exposure
	f.SetCellValue(sheet, cell("A", row), "Correlated Exposure")
	f.SetCellStyle(sheet, cell("A", row), cell("A", row), headerStyle)
	row++
	f.SetCellValue(sheet, cell("A", row), "Positions in the same sector move together — individual weights understate true concentration")
	row++
	row++

	writeSubHeader(f, sheet, row, []string{"Sector", "Positions", "Combined Value", "% of Portfolio"}, headerStyle)
	row++

	for _, se := range sectorEntries {
		if se.count < 2 {
			continue
		}
		f.SetCellValue(sheet, cell("A", row), se.name)

		var symbols []string
		for _, p := range positions {
			if p.sector == se.name {
				symbols = append(symbols, p.symbol)
			}
		}
		maxShow := 6
		label := strings.Join(symbols, ", ")
		if len(symbols) > maxShow {
			label = strings.Join(symbols[:maxShow], ", ") + fmt.Sprintf(" +%d more", len(symbols)-maxShow)
		}
		f.SetCellValue(sheet, cell("B", row), label)
		setCurrency(f, sheet, "C", row, se.val, currencyFmt)
		setPercent(f, sheet, "D", row, se.val/data.totalValue, pctFmt)
		row++
	}
	row++

	// Section 6: Asset Location Efficiency
	f.SetCellValue(sheet, cell("A", row), "Asset Location")
	f.SetCellStyle(sheet, cell("A", row), cell("A", row), headerStyle)
	row++
	f.SetCellValue(sheet, cell("A", row), "Tax-efficient placement: growth assets in Roth, bonds/income in tax-deferred, tax-efficient index in taxable")
	row++
	row++

	writeSubHeader(f, sheet, row, []string{"Asset Class", "Taxable", "Tax-Deferred", "Roth"}, headerStyle)
	row++

	locationBuckets := make(map[string]map[string]float64)
	for _, h := range data.allHoldings {
		acct := acctMap[h.AccountID]
		qty := nullFloat64ToFloat(h.QuantityMicros) / 1_000_000
		cost := nullFloat64ToFloat(h.CostBasisMicros) / 1_000_000
		mv := marketValue(h.Symbol, h.SecurityType.String, qty, cost, data.quotes)
		qd := data.quotes[h.Symbol]
		n := bestName(h.SecurityName.String, qd.Name)
		sector := resolveSector(qd, h.SecurityType.String, h.Symbol, n)
		bucket := sectorToAssetBucket(sector, h.SecurityType.String)

		loc := "taxable"
		switch acct.AccountType {
		case "traditional_ira", "sep_ira":
			loc = "deferred"
		case "roth_ira":
			loc = "roth"
		}

		if locationBuckets[bucket] == nil {
			locationBuckets[bucket] = make(map[string]float64)
		}
		locationBuckets[bucket][loc] += mv
	}
	for _, a := range data.accounts {
		cash := float64(data.cashByAcct[a.ID]) / 1_000_000
		if cash == 0 {
			continue
		}
		loc := "taxable"
		switch a.AccountType {
		case "traditional_ira", "sep_ira":
			loc = "deferred"
		case "roth_ira":
			loc = "roth"
		}
		if locationBuckets["Cash & Equivalents"] == nil {
			locationBuckets["Cash & Equivalents"] = make(map[string]float64)
		}
		locationBuckets["Cash & Equivalents"][loc] += cash
	}

	for _, bucket := range bucketOrder {
		locs := locationBuckets[bucket]
		if locs == nil {
			continue
		}
		total := locs["taxable"] + locs["deferred"] + locs["roth"]
		if total == 0 {
			continue
		}
		f.SetCellValue(sheet, cell("A", row), bucket)
		setCurrency(f, sheet, "B", row, locs["taxable"], currencyFmt)
		setCurrency(f, sheet, "C", row, locs["deferred"], currencyFmt)
		setCurrency(f, sheet, "D", row, locs["roth"], currencyFmt)
		row++
	}
}

func writePerformanceSheet(f *excelize.File, data *reportData, currencyFmt, pctFmt, headerStyle int) {
	sheet := "Performance"
	f.NewSheet(sheet)

	row := 1
	f.SetCellValue(sheet, cell("A", row), "Performance")
	f.SetCellStyle(sheet, cell("A", row), cell("A", row), headerStyle)
	row++

	if data.actualPerformance == nil {
		f.SetCellValue(sheet, cell("A", row), "Performance data unavailable.")
		return
	}

	perf := data.actualPerformance
	f.SetCellValue(sheet, cell("A", row), fmt.Sprintf("Actual time-weighted returns from reconstructed daily account values, %s through %s.", perf.StartDate, perf.AsOf))
	row++
	f.SetCellValue(sheet, cell("A", row), "External cash flows include opening balances and transfers. Buys, sells, dividends, interest, and fees remain inside performance.")
	row++
	row++

	f.SetCellValue(sheet, cell("A", row), "Summary")
	f.SetCellStyle(sheet, cell("A", row), cell("A", row), headerStyle)
	row++
	writeSubHeader(f, sheet, row, []string{"Account", "Type", "Start Value", "Net Flows", "End Value", "1M", "3M", "6M", "YTD", "1Y"}, headerStyle)
	row++

	rows := append([]performanceRow{perf.Portfolio}, perf.Accounts...)
	for _, pr := range rows {
		f.SetCellValue(sheet, cell("A", row), pr.Name)
		f.SetCellValue(sheet, cell("B", row), pr.AccountType)
		setCurrency(f, sheet, "C", row, pr.StartValue, currencyFmt)
		setCurrency(f, sheet, "D", row, pr.NetExternalFlows, currencyFmt)
		setCurrency(f, sheet, "E", row, pr.EndValue, currencyFmt)
		setPerformancePctCell(f, sheet, "F", row, pr.OneMonth, pctFmt)
		setPerformancePctCell(f, sheet, "G", row, pr.ThreeMonth, pctFmt)
		setPerformancePctCell(f, sheet, "H", row, pr.SixMonth, pctFmt)
		setPerformancePctCell(f, sheet, "I", row, pr.YTD, pctFmt)
		setPerformancePctCell(f, sheet, "J", row, pr.OneYear, pctFmt)
		row++
	}
	row++

	f.SetCellValue(sheet, cell("A", row), "Monthly Returns (Portfolio)")
	f.SetCellStyle(sheet, cell("A", row), cell("A", row), headerStyle)
	row++
	writeSubHeader(f, sheet, row, []string{"Month", "Return"}, headerStyle)
	row++
	for _, mr := range perf.Monthly {
		f.SetCellValue(sheet, cell("A", row), mr.Period)
		setPercent(f, sheet, "B", row, mr.ReturnPct, pctFmt)
		row++
	}

	if len(perf.CoverageNotes) > 0 {
		row++
		f.SetCellValue(sheet, cell("A", row), "Coverage Notes")
		f.SetCellStyle(sheet, cell("A", row), cell("A", row), headerStyle)
		row++
		for _, note := range perf.CoverageNotes {
			f.SetCellValue(sheet, cell("A", row), note)
			row++
		}
	}
}

func buildActualPerformance(
	ctx context.Context,
	queries *db.Queries,
	portfolioURL string,
	allHoldings []db.ListAllHoldingsRow,
	accounts []db.Account,
	cashByAcct map[string]int64,
	quotes map[string]quoteData,
) (*actualPerformanceReport, error) {
	endDate := time.Now()
	startDate := endDate.AddDate(-1, 0, 0)
	startDateStr := startDate.Format("2006-01-02")
	endDateStr := endDate.Format("2006-01-02")

	currentQtyByAccount := make(map[string]map[string]int64)
	securityTypeBySymbol := make(map[string]string)
	fallbackPriceBySymbol := make(map[string]float64)
	symbolSet := make(map[string]bool)

	for _, h := range allHoldings {
		qtyMicros := int64(math.Round(nullFloat64ToFloat(h.QuantityMicros)))
		if _, ok := currentQtyByAccount[h.AccountID]; !ok {
			currentQtyByAccount[h.AccountID] = make(map[string]int64)
		}
		currentQtyByAccount[h.AccountID][h.Symbol] = qtyMicros
		securityTypeBySymbol[h.Symbol] = h.SecurityType.String
		symbolSet[h.Symbol] = true

		qty := float64(qtyMicros) / 1_000_000
		if qty > 0 {
			if qd, ok := quotes[h.Symbol]; ok && qd.Price > 0 {
				fallbackPriceBySymbol[h.Symbol] = qd.Price
			} else {
				cost := nullFloat64ToFloat(h.CostBasisMicros) / 1_000_000
				if cost > 0 {
					fallbackPriceBySymbol[h.Symbol] = cost / qty
				}
			}
		}
	}

	transactionsByAccount := make(map[string][]db.ListTransactionsByAccountRow)
	cashTxnsByAccount := make(map[string][]db.ListCashTransactionsByDateRangeRow)
	startCashByAccount := make(map[string]float64)
	lotsByAccount := make(map[string][]db.ListLotsForPerformanceRow)
	dispositionsByLot := make(map[string][]db.ListLotDispositionsForPerformanceRow)

	lots, err := queries.ListLotsForPerformance(ctx)
	if err != nil {
		return nil, fmt.Errorf("list lots for performance: %w", err)
	}
	for _, lot := range lots {
		lotsByAccount[lot.AccountID] = append(lotsByAccount[lot.AccountID], lot)
		symbolSet[lot.Symbol] = true
		if _, ok := securityTypeBySymbol[lot.Symbol]; !ok {
			securityTypeBySymbol[lot.Symbol] = lot.SecurityType.String
		}
	}

	dispositions, err := queries.ListLotDispositionsForPerformance(ctx)
	if err != nil {
		return nil, fmt.Errorf("list lot dispositions for performance: %w", err)
	}
	for _, disposition := range dispositions {
		dispositionsByLot[disposition.LotID] = append(dispositionsByLot[disposition.LotID], disposition)
	}

	for _, account := range accounts {
		txns, err := queries.ListTransactionsByAccount(ctx, account.ID)
		if err != nil {
			return nil, fmt.Errorf("list transactions for %s: %w", account.Name, err)
		}
		transactionsByAccount[account.ID] = txns

		for _, txn := range txns {
			if txn.Symbol.Valid {
				symbolSet[txn.Symbol.String] = true
			}
			if txn.Symbol.Valid && txn.PriceMicros.Valid && txn.PriceMicros.Int64 > 0 {
				if _, ok := fallbackPriceBySymbol[txn.Symbol.String]; !ok {
					fallbackPriceBySymbol[txn.Symbol.String] = float64(txn.PriceMicros.Int64) / 1_000_000
				}
			}
		}

		cashStart, err := queries.GetCashBalanceAsOfDate(ctx, db.GetCashBalanceAsOfDateParams{
			AccountID: account.ID,
			AsOfDate:  startDateStr,
		})
		if err != nil {
			return nil, fmt.Errorf("get opening cash for %s: %w", account.Name, err)
		}
		startCashByAccount[account.ID] = float64(toInt64Val(cashStart)) / 1_000_000

		cashTxns, err := queries.ListCashTransactionsByDateRange(ctx, db.ListCashTransactionsByDateRangeParams{
			AccountID: account.ID,
			StartDate: startDateStr,
			EndDate:   endDateStr,
		})
		if err != nil {
			return nil, fmt.Errorf("list cash transactions for %s: %w", account.Name, err)
		}
		cashTxnsByAccount[account.ID] = cashTxns
	}

	var symbols []string
	for symbol := range symbolSet {
		symbols = append(symbols, symbol)
	}
	sort.Strings(symbols)

	priceHistory, coverageNotes, err := fetchPerformancePriceHistory(ctx, portfolioURL, symbols, startDate, endDate, fallbackPriceBySymbol, securityTypeBySymbol)
	if err != nil {
		return nil, err
	}

	var accountSeries []accountPerformanceSeries
	for _, account := range accounts {
		series, err := buildAccountPerformanceSeries(
			account,
			lotsByAccount[account.ID],
			dispositionsByLot,
			transactionsByAccount[account.ID],
			cashTxnsByAccount[account.ID],
			startCashByAccount[account.ID],
			priceHistory,
			startDate,
			endDate,
		)
		if err != nil {
			return nil, fmt.Errorf("build performance for %s: %w", account.Name, err)
		}
		accountSeries = append(accountSeries, series)
	}

	portfolioSeries := combinePortfolioSeries(accountSeries)
	return &actualPerformanceReport{
		StartDate:     startDateStr,
		AsOf:          endDateStr,
		Portfolio:     summarizePerformanceRow("Total Portfolio", "All Accounts", portfolioSeries),
		Accounts:      summarizeAccountRows(accountSeries),
		Monthly:       monthlyReturns(portfolioSeries),
		CoverageNotes: coverageNotes,
	}, nil
}

func fetchPerformancePriceHistory(
	ctx context.Context,
	portfolioURL string,
	symbols []string,
	startDate, endDate time.Time,
	fallbackPriceBySymbol map[string]float64,
	securityTypeBySymbol map[string]string,
) (map[string]map[string]float64, []string, error) {
	result := make(map[string]map[string]float64, len(symbols))
	coverageNotes := make([]string, 0)
	client := &http.Client{Timeout: 120 * time.Second}

	type chartResponse struct {
		Symbol string                  `json:"symbol"`
		Points []performanceChartPoint `json:"points"`
	}

	var mu sync.Mutex
	var wg sync.WaitGroup
	sem := make(chan struct{}, 8)
	var firstErr error

	for _, symbol := range symbols {
		symbol := symbol
		wg.Add(1)
		go func() {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			url := fmt.Sprintf("%s/api/v1/chart/%s?range=2y&interval=1d", portfolioURL, url.PathEscape(symbol))
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
			if err != nil {
				mu.Lock()
				if firstErr == nil {
					firstErr = err
				}
				mu.Unlock()
				return
			}

			resp, err := client.Do(req)
			if err != nil {
				mu.Lock()
				if firstErr == nil {
					firstErr = err
				}
				mu.Unlock()
				return
			}
			defer resp.Body.Close()

			var body chartResponse
			if resp.StatusCode == http.StatusOK {
				if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
					mu.Lock()
					if firstErr == nil {
						firstErr = err
					}
					mu.Unlock()
					return
				}
			}

			prices, note := expandPriceHistory(symbol, body.Points, startDate, endDate, fallbackPriceBySymbol[symbol], securityTypeBySymbol[symbol])
			mu.Lock()
			result[symbol] = prices
			if note != "" {
				coverageNotes = append(coverageNotes, note)
			}
			mu.Unlock()
		}()
	}

	wg.Wait()
	if firstErr != nil {
		return nil, nil, firstErr
	}
	sort.Strings(coverageNotes)
	return result, coverageNotes, nil
}

func expandPriceHistory(symbol string, points []performanceChartPoint, startDate, endDate time.Time, fallbackPrice float64, securityType string) (map[string]float64, string) {
	prices := make(map[string]float64)
	if fallbackPrice <= 0 && securityType == "bond" {
		fallbackPrice = 1
	}

	if len(points) == 0 {
		if fallbackPrice <= 0 {
			fallbackPrice = 0
		}
		for d := startDate; !d.After(endDate); d = d.AddDate(0, 0, 1) {
			prices[d.Format("2006-01-02")] = fallbackPrice
		}
		return prices, fmt.Sprintf("Used flat fallback pricing for %s due to missing chart history.", symbol)
	}

	type pricePoint struct {
		Date  string
		Close float64
	}
	var usable []pricePoint
	for _, point := range points {
		if point.Close == nil || *point.Close <= 0 || len(point.Timestamp) < 10 {
			continue
		}
		usable = append(usable, pricePoint{Date: point.Timestamp[:10], Close: *point.Close})
	}
	if len(usable) == 0 {
		if fallbackPrice <= 0 {
			fallbackPrice = 0
		}
		for d := startDate; !d.After(endDate); d = d.AddDate(0, 0, 1) {
			prices[d.Format("2006-01-02")] = fallbackPrice
		}
		return prices, fmt.Sprintf("Used flat fallback pricing for %s due to missing close prices.", symbol)
	}

	sort.Slice(usable, func(i, j int) bool { return usable[i].Date < usable[j].Date })
	idx := 0
	lastPrice := 0.0
	havePrice := false
	for d := startDate; !d.After(endDate); d = d.AddDate(0, 0, 1) {
		dateStr := d.Format("2006-01-02")
		for idx < len(usable) && usable[idx].Date <= dateStr {
			lastPrice = usable[idx].Close
			havePrice = true
			idx++
		}
		if havePrice {
			prices[dateStr] = lastPrice
			continue
		}
		if fallbackPrice > 0 {
			prices[dateStr] = fallbackPrice
		}
	}

	if !havePrice && fallbackPrice > 0 {
		return prices, fmt.Sprintf("Used flat fallback pricing for %s before market history begins.", symbol)
	}
	if fallbackPrice > 0 {
		firstDate := usable[0].Date
		if firstDate > startDate.Format("2006-01-02") {
			return prices, fmt.Sprintf("Used fallback pricing for %s until %s when market history begins.", symbol, firstDate)
		}
	}
	return prices, ""
}

func buildAccountPerformanceSeries(
	account db.Account,
	lots []db.ListLotsForPerformanceRow,
	dispositionsByLot map[string][]db.ListLotDispositionsForPerformanceRow,
	transactions []db.ListTransactionsByAccountRow,
	cashTxns []db.ListCashTransactionsByDateRangeRow,
	startCash float64,
	priceHistory map[string]map[string]float64,
	startDate, endDate time.Time,
) (accountPerformanceSeries, error) {
	startDateStr := startDate.Format("2006-01-02")
	endDateStr := endDate.Format("2006-01-02")

	qtyBySymbol := make(map[string]int64)

	dailyQtyDelta := make(map[string]map[string]int64)
	for _, lot := range lots {
		if lot.AcquiredDate <= startDateStr {
			qtyBySymbol[lot.Symbol] += lot.QuantityMicros
		} else if lot.AcquiredDate <= endDateStr {
			if _, ok := dailyQtyDelta[lot.AcquiredDate]; !ok {
				dailyQtyDelta[lot.AcquiredDate] = make(map[string]int64)
			}
			dailyQtyDelta[lot.AcquiredDate][lot.Symbol] += lot.QuantityMicros
		}

		for _, disposition := range dispositionsByLot[lot.ID] {
			if disposition.DisposedDate <= startDateStr {
				qtyBySymbol[lot.Symbol] -= disposition.QuantityMicros
			} else if disposition.DisposedDate <= endDateStr {
				if _, ok := dailyQtyDelta[disposition.DisposedDate]; !ok {
					dailyQtyDelta[disposition.DisposedDate] = make(map[string]int64)
				}
				dailyQtyDelta[disposition.DisposedDate][lot.Symbol] -= disposition.QuantityMicros
			}
		}
		if qtyBySymbol[lot.Symbol] == 0 {
			delete(qtyBySymbol, lot.Symbol)
		}
	}

	cashDeltaByDate := make(map[string]float64)
	externalFlowByDate := make(map[string]float64)
	for _, cashTxn := range cashTxns {
		if cashTxn.TransactionDate <= startDateStr || cashTxn.TransactionDate > endDateStr {
			continue
		}
		amount := float64(cashTxn.AmountMicros) / 1_000_000
		cashDeltaByDate[cashTxn.TransactionDate] += amount
		if isExternalCashFlow(cashTxn.CashType) {
			externalFlowByDate[cashTxn.TransactionDate] += amount
		}
	}
	for _, txn := range transactions {
		if !txn.Symbol.Valid || !txn.QuantityMicros.Valid {
			continue
		}
		if txn.TransactionDate <= startDateStr || txn.TransactionDate > endDateStr {
			continue
		}
		if !isExternalSecurityFlow(txn.TransactionType) {
			continue
		}
		unitPrice := externalSecurityUnitPrice(txn, priceHistory)
		if unitPrice <= 0 {
			continue
		}
		deltaQty := float64(quantityDeltaMicros(txn.TransactionType, txn.QuantityMicros.Int64)) / 1_000_000
		externalFlowByDate[txn.TransactionDate] += deltaQty * unitPrice
	}

	series := make([]dailySeriesPoint, 0, 370)
	cashBalance := startCash
	startValue, err := holdingsMarketValue(qtyBySymbol, startDateStr, priceHistory)
	if err != nil {
		return accountPerformanceSeries{}, err
	}
	series = append(series, dailySeriesPoint{Date: startDateStr, Value: cashBalance + startValue})
	previousValue := series[0].Value

	for d := startDate.AddDate(0, 0, 1); !d.After(endDate); d = d.AddDate(0, 0, 1) {
		dateStr := d.Format("2006-01-02")
		cashBalance += cashDeltaByDate[dateStr]
		for symbol, delta := range dailyQtyDelta[dateStr] {
			qtyBySymbol[symbol] += delta
			if qtyBySymbol[symbol] == 0 {
				delete(qtyBySymbol, symbol)
			}
		}

		holdingsValue, err := holdingsMarketValue(qtyBySymbol, dateStr, priceHistory)
		if err != nil {
			return accountPerformanceSeries{}, err
		}
		point := dailySeriesPoint{
			Date:         dateStr,
			Value:        cashBalance + holdingsValue,
			ExternalFlow: externalFlowByDate[dateStr],
		}
		if previousValue > 0 {
			ret := (point.Value - point.ExternalFlow - previousValue) / previousValue
			if !math.IsNaN(ret) && !math.IsInf(ret, 0) {
				point.Return = &ret
			}
		}
		series = append(series, point)
		previousValue = point.Value
	}

	return accountPerformanceSeries{
		Row:    summarizePerformanceRow(account.Name, formatAccountType(account.AccountType), series),
		Series: series,
	}, nil
}

func holdingsMarketValue(qtyBySymbol map[string]int64, dateStr string, priceHistory map[string]map[string]float64) (float64, error) {
	value := 0.0
	for symbol, qtyMicros := range qtyBySymbol {
		if qtyMicros == 0 {
			continue
		}
		prices, ok := priceHistory[symbol]
		if !ok {
			return 0, fmt.Errorf("missing price history for %s", symbol)
		}
		price, ok := prices[dateStr]
		if !ok {
			return 0, fmt.Errorf("missing price for %s on %s", symbol, dateStr)
		}
		value += float64(qtyMicros) / 1_000_000 * price
	}
	return value, nil
}

func quantityDeltaMicros(transactionType string, qtyMicros int64) int64 {
	absQty := qtyMicros
	if absQty < 0 {
		absQty = -absQty
	}

	switch transactionType {
	case "buy", "opening_balance", "security_transfer", "reorg_in", "split", "transfer_in":
		return absQty
	case "sell", "reorg_out", "transfer_out":
		return -absQty
	default:
		return 0
	}
}

func isExternalCashFlow(cashType string) bool {
	switch cashType {
	case "opening", "transfer_in", "transfer_out":
		return true
	default:
		return false
	}
}

func isExternalSecurityFlow(transactionType string) bool {
	switch transactionType {
	case "opening_balance", "security_transfer", "transfer_in", "transfer_out":
		return true
	default:
		return false
	}
}

func externalSecurityUnitPrice(txn db.ListTransactionsByAccountRow, priceHistory map[string]map[string]float64) float64 {
	prices, ok := priceHistory[txn.Symbol.String]
	if ok {
		if price, ok := prices[txn.TransactionDate]; ok && price > 0 {
			return price
		}
	}
	if txn.QuantityMicros.Valid && txn.QuantityMicros.Int64 != 0 && txn.AmountMicros != 0 {
		qty := math.Abs(float64(txn.QuantityMicros.Int64)) / 1_000_000
		if qty > 0 {
			return math.Abs(float64(txn.AmountMicros)/1_000_000) / qty
		}
	}
	if txn.PriceMicros.Valid && txn.PriceMicros.Int64 > 0 {
		return float64(txn.PriceMicros.Int64) / 1_000_000
	}
	return 0
}

func combinePortfolioSeries(accounts []accountPerformanceSeries) []dailySeriesPoint {
	if len(accounts) == 0 {
		return nil
	}
	series := make([]dailySeriesPoint, 0, len(accounts[0].Series))
	for i := range accounts[0].Series {
		point := dailySeriesPoint{Date: accounts[0].Series[i].Date}
		for _, account := range accounts {
			if i >= len(account.Series) {
				continue
			}
			point.Value += account.Series[i].Value
			point.ExternalFlow += account.Series[i].ExternalFlow
		}
		if i > 0 {
			prev := series[i-1].Value
			if prev > 0 {
				ret := (point.Value - point.ExternalFlow - prev) / prev
				if !math.IsNaN(ret) && !math.IsInf(ret, 0) {
					point.Return = &ret
				}
			}
		}
		series = append(series, point)
	}
	return series
}

func summarizeAccountRows(accounts []accountPerformanceSeries) []performanceRow {
	rows := make([]performanceRow, 0, len(accounts))
	for _, account := range accounts {
		rows = append(rows, account.Row)
	}
	sort.Slice(rows, func(i, j int) bool {
		return rows[i].EndValue > rows[j].EndValue
	})
	return rows
}

func summarizePerformanceRow(name, accountType string, series []dailySeriesPoint) performanceRow {
	row := performanceRow{Name: name, AccountType: accountType}
	if len(series) == 0 {
		return row
	}
	row.StartValue = series[0].Value
	row.EndValue = series[len(series)-1].Value
	for _, point := range series {
		row.NetExternalFlows += point.ExternalFlow
	}

	asOf := mustParseDate(series[len(series)-1].Date)
	inception := firstPositiveValueDate(series)
	row.OneMonth = periodReturnSince(series, asOf.AddDate(0, -1, 0), inception)
	row.ThreeMonth = periodReturnSince(series, asOf.AddDate(0, -3, 0), inception)
	row.SixMonth = periodReturnSince(series, asOf.AddDate(0, -6, 0), inception)
	row.YTD = periodReturnSince(series, time.Date(asOf.Year(), 1, 1, 0, 0, 0, 0, asOf.Location()), inception)
	row.OneYear = periodReturnSince(series, asOf.AddDate(-1, 0, 0), inception)
	return row
}

func periodReturnSince(series []dailySeriesPoint, start, inception time.Time) *float64 {
	if inception.After(start) {
		return nil
	}
	result := 1.0
	haveReturn := false
	for _, point := range series {
		if point.Return == nil {
			continue
		}
		pointDate := mustParseDate(point.Date)
		if pointDate.Before(start) {
			continue
		}
		result *= 1 + *point.Return
		haveReturn = true
	}
	if !haveReturn {
		return nil
	}
	value := result - 1
	return &value
}

func firstPositiveValueDate(series []dailySeriesPoint) time.Time {
	for _, point := range series {
		if point.Value > 0 {
			return mustParseDate(point.Date)
		}
	}
	return mustParseDate(series[0].Date)
}

func monthlyReturns(series []dailySeriesPoint) []performanceMonthlyRow {
	monthlyGrowth := make(map[string]float64)
	var periods []string
	seen := make(map[string]bool)
	for _, point := range series {
		if point.Return == nil || len(point.Date) < 7 {
			continue
		}
		period := point.Date[:7]
		if !seen[period] {
			seen[period] = true
			periods = append(periods, period)
			monthlyGrowth[period] = 1
		}
		monthlyGrowth[period] *= 1 + *point.Return
	}
	sort.Strings(periods)
	if len(periods) > 12 {
		periods = periods[len(periods)-12:]
	}
	rows := make([]performanceMonthlyRow, 0, len(periods))
	for _, period := range periods {
		rows = append(rows, performanceMonthlyRow{
			Period:    period,
			ReturnPct: monthlyGrowth[period] - 1,
		})
	}
	return rows
}

func setPerformancePctCell(f *excelize.File, sheet, col string, row int, val *float64, style int) {
	if val == nil {
		return
	}
	setPercent(f, sheet, col, row, *val, style)
}

func mustParseDate(dateStr string) time.Time {
	t, _ := time.Parse("2006-01-02", dateStr)
	return t
}

func writeHeaderRow(f *excelize.File, sheet string, headers []string, style int) {
	for i, h := range headers {
		c := cell(string(rune('A'+i)), 1)
		f.SetCellValue(sheet, c, h)
		f.SetCellStyle(sheet, c, c, style)
	}
}

func cell(col string, row int) string {
	return fmt.Sprintf("%s%d", col, row)
}

func setCurrency(f *excelize.File, sheet, col string, row int, val float64, style int) {
	c := cell(col, row)
	f.SetCellValue(sheet, c, val)
	f.SetCellStyle(sheet, c, c, style)
}

func setPercent(f *excelize.File, sheet, col string, row int, val float64, style int) {
	c := cell(col, row)
	f.SetCellValue(sheet, c, val)
	f.SetCellStyle(sheet, c, c, style)
}

func setCurrencyFormula(f *excelize.File, sheet, col string, row int, formula string, style int) {
	c := cell(col, row)
	f.SetCellFormula(sheet, c, formula)
	f.SetCellStyle(sheet, c, c, style)
}

func setPercentFormula(f *excelize.File, sheet, col string, row int, formula string, style int) {
	c := cell(col, row)
	f.SetCellFormula(sheet, c, formula)
	f.SetCellStyle(sheet, c, c, style)
}

func sectorToAssetBucket(sector, secType string) string {
	switch sector {
	case "Municipal Bonds", "Corporate Bonds", "Government Bonds":
		return "Fixed Income"
	case "Cash & Equivalents":
		return "Cash & Equivalents"
	case "Crypto":
		return "Crypto"
	case "Real Estate":
		return "Real Estate"
	case "Alternatives":
		return "Alternatives"
	}
	if secType == "bond" {
		return "Fixed Income"
	}
	return "Equity"
}

func sortedKeys(m map[string]float64) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		return m[keys[i]] > m[keys[j]]
	})
	return keys
}
