package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/xuri/excelize/v2"

	"github.com/levisegal/monay/services/holdings/config"
	"github.com/levisegal/monay/services/holdings/database"
	"github.com/levisegal/monay/services/holdings/gen/db"
)

const (
	birthYear = 1960
	rmdStartAge = 75
	rmdEndAge   = 90
	growthRate  = 0.05

	// Federal MFJ 2025 standard deduction (both 65+: base $30,000 + $1,600×2 for 65+)
	federalStdDeduction = 32600.0
	// CA MFJ standard deduction
	caStdDeduction = 10726.0
	// Qualified dividend 0% threshold MFJ 2025
	qualDivZeroPctThreshold = 94050.0
)

func incomeCommand() *cobra.Command {
	var output string

	cmd := &cobra.Command{
		Use:   "income",
		Short: "Generate projected income report",
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
			return generateIncomeReport(ctx, queries, cfg, output)
		},
	}

	cmd.Flags().StringVarP(&output, "output", "o", "", "Output XLSX file path")
	cmd.MarkFlagRequired("output")

	return cmd
}

type incomePosition struct {
	Symbol            string
	Name              string
	AccountName       string
	AccountID         string
	AccountType       string
	SecurityType      string
	Qty               float64
	CostBasis         float64
	MktValue          float64
	DivRate           float64
	YieldPct          float64
	CouponPct         float64
	ParValue          float64
	AnnualIncome      float64
	TaxTreatment      string
	Frequency         string
	State             string
	Maturity          string
	NextCallDate      string
	Callable          bool
	LikelyRepayment   string
	PricedToCall      bool
}

type bondInfo struct {
	CouponPct      float64 `json:"coupon_pct"`
	MaturityDate   string  `json:"maturity_date"`
	NextCallDate   string  `json:"next_call_date"`
	Callable       bool    `json:"callable"`
	State          string  `json:"state"`
	YieldToWorst   float64 `json:"yield_to_worst"`
	YieldToMaturity float64 `json:"yield_to_maturity"`
}

// IRS Uniform Lifetime Table (SECURE 2.0)
var uniformLifetimeTable = map[int]float64{
	75: 24.6, 76: 23.7, 77: 22.9, 78: 22.0, 79: 21.1,
	80: 20.2, 81: 19.4, 82: 18.5, 83: 17.7, 84: 16.8,
	85: 16.0, 86: 15.2, 87: 14.4, 88: 13.7, 89: 12.9,
	90: 12.2,
}

func generateIncomeReport(ctx context.Context, queries *db.Queries, cfg *config.Config, output string) error {
	holdings, err := queries.ListAllHoldings(ctx)
	if err != nil {
		return fmt.Errorf("list holdings: %w", err)
	}

	accounts, err := queries.ListAccounts(ctx)
	if err != nil {
		return fmt.Errorf("list accounts: %w", err)
	}

	acctMap := make(map[string]db.Account)
	for _, a := range accounts {
		acctMap[a.ID] = a
	}

	symbols := make([]string, 0, len(holdings))
	bondSymbols := make(map[string]bool)
	for _, h := range holdings {
		symbols = append(symbols, h.Symbol)
		if h.SecurityType.String == "bond" && h.SecurityType.Valid {
			bondSymbols[h.Symbol] = true
		}
	}
	quotes := fetchQuotes(ctx, cfg.PortfolioURL, symbols, bondSymbols)

	tradeableCount := len(symbols) - len(bondSymbols)
	if tradeableCount > 0 && len(quotes) == 0 {
		return fmt.Errorf("portfolio service returned no quotes — is it running at %s?", cfg.PortfolioURL)
	}

	bonds := fetchBondData(ctx, cfg.BondServiceURL, bondSymbols)

	trailing := fetchTrailingIncome(ctx, cfg.DBPath)

	totalQtyBySymbol := make(map[string]float64)
	for _, h := range holdings {
		totalQtyBySymbol[h.Symbol] += nullFloat64ToFloat(h.QuantityMicros) / 1_000_000
	}

	var positions []incomePosition
	for _, h := range holdings {
		if h.SecurityType.Valid && h.SecurityType.String == "cash" {
			continue
		}

		acct := acctMap[h.AccountID]
		qty := nullFloat64ToFloat(h.QuantityMicros) / 1_000_000
		qd := quotes[h.Symbol]
		name := bestName(h.SecurityName.String, qd.Name)

		mktVal := qty * qd.Price
		if h.SecurityType.String == "bond" && h.SecurityType.Valid {
			mktVal = qty
		}
		if qd.Price == 0 && h.SecurityType.String != "bond" {
			mktVal = nullFloat64ToFloat(h.CostBasisMicros) / 1_000_000
		}

		pos := incomePosition{
			Symbol:       h.Symbol,
			Name:         name,
			AccountName:  h.AccountName,
			AccountID:    h.AccountID,
			AccountType:  acct.AccountType,
			SecurityType: h.SecurityType.String,
			Qty:          qty,
			CostBasis:    nullFloat64ToFloat(h.CostBasisMicros) / 1_000_000,
			MktValue:     mktVal,
			DivRate:      qd.DividendRate,
			YieldPct:     qd.YieldPct,
		}

		if bi, ok := bonds[h.Symbol]; ok {
			pos.CouponPct = bi.CouponPct
			pos.ParValue = qty
			pos.Maturity = bi.MaturityDate
			pos.NextCallDate = bi.NextCallDate
			pos.Callable = bi.Callable
			pos.State = bi.State

			if bi.Callable && bi.YieldToWorst < bi.YieldToMaturity && bi.NextCallDate != "" {
				pos.LikelyRepayment = bi.NextCallDate
				pos.PricedToCall = true
			} else {
				pos.LikelyRepayment = bi.MaturityDate
			}
		}

		if pos.SecurityType == "bond" && pos.State == "" {
			pos.State = inferBondState(name)
		}

		qtyShare := 1.0
		if total := totalQtyBySymbol[h.Symbol]; total > 0 {
			qtyShare = qty / total
		}
		pos.AnnualIncome = computeAnnualIncome(pos, trailing, qtyShare)
		pos.TaxTreatment = determineTaxTreatment(pos)
		pos.Frequency = determineFrequency(pos)

		positions = append(positions, pos)
	}

	cashPositions := buildCashPositions(ctx, queries, acctMap, trailing)
	positions = append(positions, cashPositions...)

	f := excelize.NewFile()
	defer f.Close()

	currencyFmt, _ := f.NewStyle(&excelize.Style{NumFmt: 4})
	pctFmt, _ := f.NewStyle(&excelize.Style{NumFmt: 10})
	headerStyle, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true}})

	writeIncomeSummarySheet(f, positions, currencyFmt, pctFmt, headerStyle)
	writePositionIncomeSheet(f, positions, currencyFmt, pctFmt, headerStyle)
	writeRothConversionSheet(f, positions, acctMap, currencyFmt, pctFmt, headerStyle)
	writeInheritanceSheet(f, positions, acctMap, currencyFmt, pctFmt, headerStyle)
	writeTaxWaterfallSheet(f, positions, currencyFmt, pctFmt, headerStyle)

	f.DeleteSheet("Sheet1")

	if err := f.SaveAs(output); err != nil {
		return fmt.Errorf("save report: %w", err)
	}

	slog.Info("income report generated", "output", output)
	return nil
}

func computeAnnualIncome(pos incomePosition, trailing map[string]float64, qtyShare float64) float64 {
	if pos.SecurityType == "bond" && pos.CouponPct > 0 {
		return pos.ParValue * pos.CouponPct / 100
	}

	if pos.DivRate > 0 {
		return pos.Qty * pos.DivRate
	}

	if pos.YieldPct > 0 {
		return pos.MktValue * pos.YieldPct
	}

	if t, ok := trailing[pos.Symbol]; ok && t > 0 {
		return t * qtyShare
	}

	return 0
}

func determineTaxTreatment(pos incomePosition) string {
	switch pos.AccountType {
	case "roth_ira":
		return "Tax-Free (Roth)"
	case "traditional_ira", "sep_ira":
		return "Tax-Deferred (IRA)"
	}

	if pos.SecurityType == "bond" {
		if pos.State == "CA" || isCASecurity(pos.Symbol) {
			return "Tax-Exempt (CA Muni)"
		}
		return "Tax-Exempt (Non-CA Muni)"
	}

	if isNationalMuniFund(pos.Symbol) {
		return "Tax-Exempt (National Muni Fund)"
	}

	if isMoneyMarket(pos.Symbol, pos.Name) {
		return "Ordinary Income"
	}

	if isQualifiedDividendPayer(pos.SecurityType) {
		return "Qualified Dividend"
	}

	return "Ordinary Income"
}

func determineFrequency(pos incomePosition) string {
	if pos.SecurityType == "bond" {
		return "Semi-Annual"
	}
	if isMoneyMarket(pos.Symbol, pos.Name) {
		return "Monthly"
	}
	if pos.SecurityType == "" || pos.SecurityType == "equity" || pos.SecurityType == "etf" {
		return "Quarterly"
	}
	return "Varies"
}

func isMoneyMarket(symbol, name string) bool {
	mm := map[string]bool{
		"VMFXX": true, "TMCXX": true, "FCASH": true,
		"SPAXX": true, "FDRXX": true, "SWVXX": true,
	}
	if mm[symbol] {
		return true
	}
	upper := strings.ToUpper(name)
	return strings.Contains(upper, "MONEY MARKET") || strings.Contains(upper, "MMKT") ||
		strings.Contains(upper, "CASH RESERVE") || strings.Contains(upper, "BANK DEPOSIT")
}

func isNationalMuniFund(symbol string) bool {
	return symbol == "MANLX" || symbol == "ITM"
}

func isCASecurity(symbol string) bool {
	return false
}

func inferBondState(name string) string {
	upper := strings.ToUpper(name)
	caPatterns := []string{
		"CALIFORNIA", "CALIF", " CA ", "CA ST", "CA CITY",
		"LOS ANGELES", "SAN FRANCISCO", "SAN DIEGO",
		"BAY AREA", "BAY TOLL",
		"LONG BCH CA", "PLEASANTON CA", "SWEETWATER CA",
		"NORTHERN CA", "METROPOLITAN WTR DIST STHN CA",
	}
	for _, p := range caPatterns {
		if strings.Contains(upper, p) {
			return "CA"
		}
	}
	return ""
}

func isQualifiedDividendPayer(secType string) bool {
	return secType == "" || secType == "equity" || secType == "etf" || secType == "reit"
}

const bondCacheFile = "bond_cache.json"

func fetchBondData(ctx context.Context, bondServiceURL string, bondSymbols map[string]bool) map[string]bondInfo {
	result := make(map[string]bondInfo)
	if len(bondSymbols) == 0 {
		return result
	}

	cached := loadBondCache()

	var missing []string
	for sym := range bondSymbols {
		if bi, ok := cached[sym]; ok {
			result[sym] = bi
		} else {
			missing = append(missing, sym)
		}
	}

	if len(missing) == 0 {
		return result
	}

	slog.Info("fetching bond details", "cached", len(result), "missing", len(missing))

	for _, cusip := range missing {
		bi, err := fetchSingleBondDetails(ctx, bondServiceURL, cusip)
		if err != nil {
			slog.Warn("bond details failed", "cusip", cusip, "error", err)
			continue
		}
		result[cusip] = bi
		cached[cusip] = bi
	}

	saveBondCache(cached)
	return result
}

func fetchSingleBondDetails(ctx context.Context, bondServiceURL, cusip string) (bondInfo, error) {
	client := &http.Client{Timeout: 15 * time.Second}
	payload, _ := json.Marshal(map[string]interface{}{"cusip": cusip})

	bi, err := fetchBondDetailsRPC(ctx, client, bondServiceURL, payload)
	if err != nil {
		return bondInfo{}, err
	}

	fetchBondYields(ctx, client, bondServiceURL, payload, &bi)

	return bi, nil
}

func fetchBondDetailsRPC(ctx context.Context, client *http.Client, bondServiceURL string, payload []byte) (bondInfo, error) {
	url := bondServiceURL + "/bond.v1beta1.BondsService/GetBondDetails"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return bondInfo{}, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return bondInfo{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return bondInfo{}, fmt.Errorf("status %d", resp.StatusCode)
	}

	var r struct {
		CouponPct    struct{ Value string } `json:"couponPct"`
		MaturityDate struct {
			Year  int `json:"year"`
			Month int `json:"month"`
			Day   int `json:"day"`
		} `json:"maturityDate"`
		NextCallDate *struct {
			Year  int `json:"year"`
			Month int `json:"month"`
			Day   int `json:"day"`
		} `json:"nextCallDate"`
		StateAbbrev string `json:"stateAbbrev"`
		Callable    bool   `json:"callable"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return bondInfo{}, err
	}

	coupon := 0.0
	if r.CouponPct.Value != "" {
		fmt.Sscanf(r.CouponPct.Value, "%f", &coupon)
	}

	bi := bondInfo{
		CouponPct:    coupon,
		MaturityDate: fmt.Sprintf("%d-%02d-%02d", r.MaturityDate.Year, r.MaturityDate.Month, r.MaturityDate.Day),
		State:        r.StateAbbrev,
		Callable:     r.Callable,
	}
	if r.NextCallDate != nil {
		bi.NextCallDate = fmt.Sprintf("%d-%02d-%02d", r.NextCallDate.Year, r.NextCallDate.Month, r.NextCallDate.Day)
	}

	return bi, nil
}

func fetchBondYields(ctx context.Context, client *http.Client, bondServiceURL string, payload []byte, bi *bondInfo) {
	url := bondServiceURL + "/bond.v1beta1.BondsService/GetInstrument"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return
	}

	var r struct {
		CloseYtw struct{ Value string } `json:"closeYtw"`
		CloseYtm struct{ Value string } `json:"closeYtm"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return
	}

	fmt.Sscanf(r.CloseYtw.Value, "%f", &bi.YieldToWorst)
	fmt.Sscanf(r.CloseYtm.Value, "%f", &bi.YieldToMaturity)
}

func loadBondCache() map[string]bondInfo {
	result := make(map[string]bondInfo)
	data, err := os.ReadFile(bondCacheFile)
	if err != nil {
		return result
	}
	json.Unmarshal(data, &result)
	return result
}

func saveBondCache(cache map[string]bondInfo) {
	data, err := json.MarshalIndent(cache, "", "  ")
	if err != nil {
		return
	}
	os.WriteFile(bondCacheFile, data, 0644)
}

func fetchTrailingIncome(ctx context.Context, dbPath string) map[string]float64 {
	result := make(map[string]float64)

	conn, err := database.Open(ctx, dbPath)
	if err != nil {
		return result
	}
	defer conn.Close()

	cutoff := time.Now().AddDate(-1, 0, 0).Format("2006-01-02")
	rows, err := conn.QueryContext(ctx, `
		select s.symbol, sum(ct.amount_micros) as total
		from cash_transactions ct
		join securities s on s.id = ct.security_id
		where ct.cash_type in ('dividend', 'interest')
		  and ct.transaction_date >= ?
		group by s.symbol
	`, cutoff)
	if err != nil {
		return result
	}
	defer rows.Close()

	for rows.Next() {
		var symbol string
		var total int64
		if err := rows.Scan(&symbol, &total); err != nil {
			continue
		}
		result[symbol] = float64(total) / 1_000_000
	}

	return result
}

func buildCashPositions(ctx context.Context, queries *db.Queries, acctMap map[string]db.Account, trailing map[string]float64) []incomePosition {
	var positions []incomePosition

	for _, acct := range acctMap {
		cashVal, _ := queries.GetCashBalance(ctx, acct.ID)
		balance := float64(toInt64Val(cashVal)) / 1_000_000
		if balance <= 0 {
			continue
		}

		pos := incomePosition{
			Symbol:      "CASH",
			Name:        "Cash Balance",
			AccountName: acct.Name,
			AccountID:   acct.ID,
			AccountType: acct.AccountType,
			Qty:         balance,
			MktValue:    balance,
		}

		switch acct.AccountType {
		case "roth_ira":
			pos.TaxTreatment = "Tax-Free (Roth)"
		case "traditional_ira", "sep_ira":
			pos.TaxTreatment = "Tax-Deferred (IRA)"
		default:
			pos.TaxTreatment = "Ordinary Income"
		}

		pos.Frequency = "Monthly"
		positions = append(positions, pos)
	}

	return positions
}

func writeIncomeSummarySheet(f *excelize.File, positions []incomePosition, currencyFmt, pctFmt, headerStyle int) {
	sheet := "Income Summary"
	f.NewSheet(sheet)

	headers := []string{"Category", "Annual Income", "Tax Treatment", "% of Total"}
	writeHeaderRow(f, sheet, headers, headerStyle)

	categories := map[string]float64{}
	catTreatment := map[string]string{}
	for _, p := range positions {
		cat := incomeCategory(p)
		categories[cat] += p.AnnualIncome
		catTreatment[cat] = p.TaxTreatment
	}

	catOrder := []string{
		"Qualified Dividends",
		"Ordinary Dividends",
		"Tax-Exempt Interest (CA Muni)",
		"Tax-Exempt Interest (National Muni Fund)",
		"Bond Coupons (Non-CA Muni)",
		"Taxable Interest",
		"Tax-Deferred Income (IRA)",
		"Tax-Free Income (Roth)",
	}

	var totalIncome float64
	for _, v := range categories {
		totalIncome += v
	}

	firstDataRow := 2
	row := 2
	for _, cat := range catOrder {
		val, ok := categories[cat]
		if !ok || val == 0 {
			continue
		}
		f.SetCellValue(sheet, cell("A", row), cat)
		setCurrency(f, sheet, "B", row, val, currencyFmt)
		f.SetCellValue(sheet, cell("C", row), catTreatment[cat])
		row++
		delete(categories, cat)
	}

	for cat, val := range categories {
		if val == 0 {
			continue
		}
		f.SetCellValue(sheet, cell("A", row), cat)
		setCurrency(f, sheet, "B", row, val, currencyFmt)
		f.SetCellValue(sheet, cell("C", row), catTreatment[cat])
		row++
	}

	lastDataRow := row - 1
	row++
	totalRow := row
	f.SetCellValue(sheet, cell("A", totalRow), "TOTAL")
	f.SetCellStyle(sheet, cell("A", totalRow), cell("A", totalRow), headerStyle)
	setCurrencyFormula(f, sheet, "B", totalRow, fmt.Sprintf("SUM(B%d:B%d)", firstDataRow, lastDataRow), currencyFmt)
	setPercent(f, sheet, "D", totalRow, 1.0, pctFmt)

	for r := firstDataRow; r <= lastDataRow; r++ {
		setPercentFormula(f, sheet, "D", r, fmt.Sprintf("B%d/B%d", r, totalRow), pctFmt)
	}
}

func incomeCategory(p incomePosition) string {
	switch {
	case p.AccountType == "roth_ira":
		return "Tax-Free Income (Roth)"
	case p.AccountType == "traditional_ira" || p.AccountType == "sep_ira":
		return "Tax-Deferred Income (IRA)"
	case p.SecurityType == "bond":
		if p.State == "CA" || isCASecurity(p.Symbol) {
			return "Tax-Exempt Interest (CA Muni)"
		}
		return "Bond Coupons (Non-CA Muni)"
	case isNationalMuniFund(p.Symbol):
		return "Tax-Exempt Interest (National Muni Fund)"
	case isMoneyMarket(p.Symbol, p.Name) || p.Symbol == "CASH":
		return "Taxable Interest"
	case isQualifiedDividendPayer(p.SecurityType):
		return "Qualified Dividends"
	default:
		return "Ordinary Dividends"
	}
}

func writePositionIncomeSheet(f *excelize.File, positions []incomePosition, currencyFmt, pctFmt, headerStyle int) {
	sheet := "Position Income Detail"
	f.NewSheet(sheet)

	headers := []string{"Symbol", "Name", "Account", "Qty", "Mkt Value", "Div Rate", "Annual Income", "Yield %", "Frequency", "Tax Treatment", "Maturity", "Next Call", "Callable", "Likely Repay"}
	writeHeaderRow(f, sheet, headers, headerStyle)

	sort.Slice(positions, func(i, j int) bool {
		return positions[i].AnnualIncome > positions[j].AnnualIncome
	})

	row := 2
	for _, p := range positions {
		f.SetCellValue(sheet, cell("A", row), p.Symbol)
		f.SetCellValue(sheet, cell("B", row), p.Name)
		f.SetCellValue(sheet, cell("C", row), p.AccountName)
		if p.Qty != 0 {
			f.SetCellValue(sheet, cell("D", row), p.Qty)
		}
		setCurrency(f, sheet, "E", row, p.MktValue, currencyFmt)
		if p.DivRate > 0 {
			setCurrency(f, sheet, "F", row, p.DivRate, currencyFmt)
		} else if p.CouponPct > 0 {
			setPercent(f, sheet, "F", row, p.CouponPct/100, pctFmt)
		}
		setCurrency(f, sheet, "G", row, p.AnnualIncome, currencyFmt)
		if p.MktValue > 0 {
			setPercentFormula(f, sheet, "H", row, fmt.Sprintf("IF(E%d>0,G%d/E%d,0)", row, row, row), pctFmt)
		}
		f.SetCellValue(sheet, cell("I", row), p.Frequency)
		f.SetCellValue(sheet, cell("J", row), p.TaxTreatment)
		if p.Maturity != "" {
			f.SetCellValue(sheet, cell("K", row), p.Maturity)
		}
		if p.NextCallDate != "" {
			f.SetCellValue(sheet, cell("L", row), p.NextCallDate)
		}
		if p.Callable {
			f.SetCellValue(sheet, cell("M", row), "Yes")
		}
		if p.LikelyRepayment != "" {
			label := p.LikelyRepayment
			if p.PricedToCall {
				label += " (call)"
			}
			f.SetCellValue(sheet, cell("N", row), label)
		}
		row++
	}

	lastDataRow := row - 1
	row++
	totalRow := row
	f.SetCellValue(sheet, cell("A", totalRow), "TOTAL")
	f.SetCellStyle(sheet, cell("A", totalRow), cell("A", totalRow), headerStyle)
	setCurrencyFormula(f, sheet, "E", totalRow, fmt.Sprintf("SUM(E2:E%d)", lastDataRow), currencyFmt)
	setCurrencyFormula(f, sheet, "G", totalRow, fmt.Sprintf("SUM(G2:G%d)", lastDataRow), currencyFmt)
	setPercentFormula(f, sheet, "H", totalRow, fmt.Sprintf("IF(E%d>0,G%d/E%d,0)", totalRow, totalRow, totalRow), pctFmt)
}

func writeRothConversionSheet(f *excelize.File, positions []incomePosition, acctMap map[string]db.Account, currencyFmt, pctFmt, headerStyle int) {
	sheet := "Roth Conversion"
	f.NewSheet(sheet)

	var iraBalance float64
	for _, p := range positions {
		acct := acctMap[p.AccountID]
		if acct.AccountType == "traditional_ira" || acct.AccountType == "sep_ira" {
			iraBalance += p.MktValue
		}
	}

	var taxableOrdinary float64
	for _, p := range positions {
		cat := incomeCategory(p)
		switch cat {
		case "Taxable Interest", "Ordinary Dividends", "Bond Coupons (Non-CA Muni)":
			taxableOrdinary += p.AnnualIncome
		}
	}
	fedBaseTaxable := math.Max(taxableOrdinary-federalStdDeduction, 0)
	caBaseTaxable := math.Max(taxableOrdinary-caStdDeduction, 0)

	topOf12 := 94_300.0
	topOf22 := 201_050.0
	spaceToTop12 := math.Max(topOf12-fedBaseTaxable, 0)
	spaceToTop22 := math.Max(topOf22-fedBaseTaxable, 0)

	currentYear := time.Now().Year()
	currentAge := currentYear - birthYear

	// Simulate to compute per-year taxes (bracket math stays in Go)
	type yearTax struct {
		age  int
		aTax float64
		bTax float64
		cTax float64
	}
	var yearTaxes []yearTax
	aIRA, bIRA, cIRA := iraBalance, iraBalance, iraBalance

	for age := currentAge; age <= rmdEndAge; age++ {
		var yt yearTax
		yt.age = age

		if age < rmdStartAge {
			bConvert := math.Min(spaceToTop12, bIRA)
			yt.bTax = computeFedTax(fedBaseTaxable+bConvert) - computeFedTax(fedBaseTaxable) +
				computeCATax(caBaseTaxable+bConvert) - computeCATax(caBaseTaxable)
			bIRA -= bConvert

			cConvert := math.Min(spaceToTop22, cIRA)
			yt.cTax = computeFedTax(fedBaseTaxable+cConvert) - computeFedTax(fedBaseTaxable) +
				computeCATax(caBaseTaxable+cConvert) - computeCATax(caBaseTaxable)
			cIRA -= cConvert
		} else {
			divisor := uniformLifetimeTable[age]
			if aIRA > 0 {
				aRMD := aIRA / divisor
				yt.aTax = computeFedTax(fedBaseTaxable+aRMD) - computeFedTax(fedBaseTaxable) +
					computeCATax(caBaseTaxable+aRMD) - computeCATax(caBaseTaxable)
				aIRA -= aRMD
			}
			if bIRA > 0 {
				bRMD := bIRA / divisor
				yt.bTax = computeFedTax(fedBaseTaxable+bRMD) - computeFedTax(fedBaseTaxable) +
					computeCATax(caBaseTaxable+bRMD) - computeCATax(caBaseTaxable)
				bIRA -= bRMD
			}
			if cIRA > 0 {
				cRMD := cIRA / divisor
				yt.cTax = computeFedTax(fedBaseTaxable+cRMD) - computeFedTax(fedBaseTaxable) +
					computeCATax(caBaseTaxable+cRMD) - computeCATax(caBaseTaxable)
				cIRA -= cRMD
			}
		}

		yearTaxes = append(yearTaxes, yt)
		aIRA *= (1 + growthRate)
		bIRA *= (1 + growthRate)
		cIRA *= (1 + growthRate)
	}

	// Assumptions block (B3-B9 are the editable cells)
	row := 1
	f.SetCellValue(sheet, cell("A", row), "Roth Conversion Strategy: Path A (None) vs B (12%) vs C (22%)")
	f.SetCellStyle(sheet, cell("A", row), cell("A", row), headerStyle)
	row++

	f.SetCellValue(sheet, cell("A", row), "Assumptions")
	f.SetCellStyle(sheet, cell("A", row), cell("A", row), headerStyle)
	row++

	// B3=iraBalance, B4=growthRate, B5=spaceToTop12, B6=spaceToTop22
	// B7=federalStdDeduction, B8=caStdDeduction, B9=taxableOrdinary
	assumptions := []struct {
		label string
		value float64
		style int
	}{
		{"IRA Balance", iraBalance, currencyFmt},
		{"Growth Rate", growthRate, pctFmt},
		{"12% Bracket Space (/yr)", spaceToTop12, currencyFmt},
		{"22% Bracket Space (/yr)", spaceToTop22, currencyFmt},
		{"Fed Std Deduction", federalStdDeduction, currencyFmt},
		{"CA Std Deduction", caStdDeduction, currencyFmt},
		{"Portfolio Ordinary Income", taxableOrdinary, currencyFmt},
	}
	for _, a := range assumptions {
		f.SetCellValue(sheet, cell("A", row), a.label)
		c := cell("B", row)
		f.SetCellValue(sheet, c, a.value)
		f.SetCellStyle(sheet, c, c, a.style)
		row++
	}
	row++

	// Column headers
	// D-G: Path A (IRA, Action, Tax/yr, Cumul Tax)
	// H-L: Path B (IRA, Action, Tax/yr, Cumul Tax, Roth)
	// M-Q: Path C (IRA, Action, Tax/yr, Cumul Tax, Roth)
	headers := []string{
		"Age", "Year", "Divisor",
		"A: IRA", "A: Action", "A: Tax/yr", "A: Cumul Tax",
		"B: IRA", "B: Action", "B: Tax/yr", "B: Cumul Tax", "B: Roth",
		"C: IRA", "C: Action", "C: Tax/yr", "C: Cumul Tax", "C: Roth",
	}
	writeSubHeader(f, sheet, row, headers, headerStyle)
	row++

	firstDataRow := row

	for i, yt := range yearTaxes {
		age := yt.age
		year := birthYear + age

		f.SetCellValue(sheet, cell("A", row), age)
		f.SetCellValue(sheet, cell("B", row), year)

		if age >= rmdStartAge {
			f.SetCellValue(sheet, cell("C", row), uniformLifetimeTable[age])
		}

		if i == 0 {
			// First row: reference starting IRA directly
			setCurrencyFormula(f, sheet, "D", row, "$B$3", currencyFmt)
			// Path B: convert MIN(bracket_space, starting_ira)
			setCurrencyFormula(f, sheet, "I", row, "MIN($B$5,$B$3)", currencyFmt)
			setCurrencyFormula(f, sheet, "H", row, fmt.Sprintf("$B$3-I%d", row), currencyFmt)
			setCurrencyFormula(f, sheet, "L", row, fmt.Sprintf("I%d", row), currencyFmt)
			// Path C
			setCurrencyFormula(f, sheet, "N", row, "MIN($B$6,$B$3)", currencyFmt)
			setCurrencyFormula(f, sheet, "M", row, fmt.Sprintf("$B$3-N%d", row), currencyFmt)
			setCurrencyFormula(f, sheet, "Q", row, fmt.Sprintf("N%d", row), currencyFmt)
		} else {
			prev := row - 1
			// Path A: RMD when divisor present, else no action
			setCurrencyFormula(f, sheet, "E", row,
				fmt.Sprintf("IF(C%d>0,D%d*(1+$B$4)/C%d,0)", row, prev, row), currencyFmt)
			setCurrencyFormula(f, sheet, "D", row,
				fmt.Sprintf("D%d*(1+$B$4)-E%d", prev, row), currencyFmt)
			// Path B: convert pre-RMD, RMD after
			setCurrencyFormula(f, sheet, "I", row,
				fmt.Sprintf("IF(C%d>0,H%d*(1+$B$4)/C%d,MIN($B$5,H%d*(1+$B$4)))", row, prev, row, prev), currencyFmt)
			setCurrencyFormula(f, sheet, "H", row,
				fmt.Sprintf("H%d*(1+$B$4)-I%d", prev, row), currencyFmt)
			setCurrencyFormula(f, sheet, "L", row,
				fmt.Sprintf("L%d*(1+$B$4)+IF(C%d=0,I%d,0)", prev, row, row), currencyFmt)
			// Path C
			setCurrencyFormula(f, sheet, "N", row,
				fmt.Sprintf("IF(C%d>0,M%d*(1+$B$4)/C%d,MIN($B$6,M%d*(1+$B$4)))", row, prev, row, prev), currencyFmt)
			setCurrencyFormula(f, sheet, "M", row,
				fmt.Sprintf("M%d*(1+$B$4)-N%d", prev, row), currencyFmt)
			setCurrencyFormula(f, sheet, "Q", row,
				fmt.Sprintf("Q%d*(1+$B$4)+IF(C%d=0,N%d,0)", prev, row, row), currencyFmt)
		}

		// Tax per year (computed — bracket math is unwieldy in Excel)
		setCurrency(f, sheet, "F", row, yt.aTax, currencyFmt)
		setCurrency(f, sheet, "J", row, yt.bTax, currencyFmt)
		setCurrency(f, sheet, "O", row, yt.cTax, currencyFmt)

		// Cumulative tax (SUM formula)
		setCurrencyFormula(f, sheet, "G", row,
			fmt.Sprintf("SUM(F$%d:F%d)", firstDataRow, row), currencyFmt)
		setCurrencyFormula(f, sheet, "K", row,
			fmt.Sprintf("SUM(J$%d:J%d)", firstDataRow, row), currencyFmt)
		setCurrencyFormula(f, sheet, "P", row,
			fmt.Sprintf("SUM(O$%d:O%d)", firstDataRow, row), currencyFmt)

		row++
	}

	lastDataRow := row - 1

	// Summary
	row++
	f.SetCellValue(sheet, cell("A", row), "TOTALS")
	f.SetCellStyle(sheet, cell("A", row), cell("A", row), headerStyle)
	row++

	writeSubHeader(f, sheet, row, []string{"", "Path A", "Path B", "Path C"}, headerStyle)
	row++

	taxRow := row
	f.SetCellValue(sheet, cell("A", row), "Cumulative Tax Paid")
	f.SetCellStyle(sheet, cell("A", row), cell("A", row), headerStyle)
	setCurrencyFormula(f, sheet, "B", row, fmt.Sprintf("G%d", lastDataRow), currencyFmt)
	setCurrencyFormula(f, sheet, "C", row, fmt.Sprintf("K%d", lastDataRow), currencyFmt)
	setCurrencyFormula(f, sheet, "D", row, fmt.Sprintf("P%d", lastDataRow), currencyFmt)
	row++

	f.SetCellValue(sheet, cell("A", row), "IRA Remaining at 90")
	setCurrencyFormula(f, sheet, "B", row, fmt.Sprintf("D%d*(1+$B$4)", lastDataRow), currencyFmt)
	setCurrencyFormula(f, sheet, "C", row, fmt.Sprintf("H%d*(1+$B$4)", lastDataRow), currencyFmt)
	setCurrencyFormula(f, sheet, "D", row, fmt.Sprintf("M%d*(1+$B$4)", lastDataRow), currencyFmt)
	row++

	f.SetCellValue(sheet, cell("A", row), "Roth Balance at 90")
	setCurrency(f, sheet, "B", row, 0, currencyFmt)
	setCurrencyFormula(f, sheet, "C", row, fmt.Sprintf("L%d*(1+$B$4)", lastDataRow), currencyFmt)
	setCurrencyFormula(f, sheet, "D", row, fmt.Sprintf("Q%d*(1+$B$4)", lastDataRow), currencyFmt)
	row++

	f.SetCellValue(sheet, cell("A", row), "Tax Savings vs Path A")
	f.SetCellStyle(sheet, cell("A", row), cell("A", row), headerStyle)
	setCurrency(f, sheet, "B", row, 0, currencyFmt)
	setCurrencyFormula(f, sheet, "C", row, fmt.Sprintf("B%d-C%d", taxRow, taxRow), currencyFmt)
	setCurrencyFormula(f, sheet, "D", row, fmt.Sprintf("B%d-D%d", taxRow, taxRow), currencyFmt)
}


func writeInheritanceSheet(f *excelize.File, positions []incomePosition, acctMap map[string]db.Account, currencyFmt, pctFmt, headerStyle int) {
	sheet := "Inheritance"
	f.NewSheet(sheet)

	var iraBalance float64
	for _, p := range positions {
		acct := acctMap[p.AccountID]
		if acct.AccountType == "traditional_ira" || acct.AccountType == "sep_ira" {
			iraBalance += p.MktValue
		}
	}

	var taxableOrdinary float64
	for _, p := range positions {
		cat := incomeCategory(p)
		switch cat {
		case "Taxable Interest", "Ordinary Dividends", "Bond Coupons (Non-CA Muni)":
			taxableOrdinary += p.AnnualIncome
		}
	}
	fedBaseTaxable := math.Max(taxableOrdinary-federalStdDeduction, 0)
	caBaseTaxable := math.Max(taxableOrdinary-caStdDeduction, 0)

	topOf12 := 94_300.0
	topOf22 := 201_050.0
	spaceToTop12 := math.Max(topOf12-fedBaseTaxable, 0)
	spaceToTop22 := math.Max(topOf22-fedBaseTaxable, 0)

	currentYear := time.Now().Year()
	currentAge := currentYear - birthYear
	yearsToRMD := rmdStartAge - currentAge

	_, pathBTax := simulateConversions(iraBalance, spaceToTop12, yearsToRMD, fedBaseTaxable, caBaseTaxable)
	_, pathCTax := simulateConversions(iraBalance, spaceToTop22, yearsToRMD, fedBaseTaxable, caBaseTaxable)

	// Pre-compute for heir tax (bracket math stays in Go)
	iraNoConvert := iraBalance
	for i := 0; i < yearsToRMD; i++ {
		iraNoConvert *= (1 + growthRate)
	}
	pathBConverted, _ := simulateConversions(iraBalance, spaceToTop12, yearsToRMD, fedBaseTaxable, caBaseTaxable)
	pathCConverted, _ := simulateConversions(iraBalance, spaceToTop22, yearsToRMD, fedBaseTaxable, caBaseTaxable)
	iraRemainingB := (iraBalance - pathBConverted)
	iraRemainingC := (iraBalance - pathCConverted)
	rothBalB := pathBConverted
	rothBalC := pathCConverted
	for i := 0; i < yearsToRMD; i++ {
		iraRemainingB *= (1 + growthRate)
		iraRemainingC *= (1 + growthRate)
		rothBalB *= (1 + growthRate)
		rothBalC *= (1 + growthRate)
	}

	// Row 1: Title
	row := 1
	f.SetCellValue(sheet, cell("A", row), "Inheritance Impact: Traditional IRA vs Roth")
	f.SetCellStyle(sheet, cell("A", row), cell("A", row), headerStyle)
	row++
	f.SetCellValue(sheet, cell("A", row), "SECURE Act: non-spouse heirs must drain inherited IRA within 10 years.")
	row++
	f.SetCellValue(sheet, cell("A", row), "Traditional IRA withdrawals taxed as ordinary income at heir's rate.")
	row++
	f.SetCellValue(sheet, cell("A", row), "Roth IRA withdrawals are tax-free.")
	row++
	row++

	// Assumptions block: B7=IRA, B8=growth, B9=years, B10=space12, B11=space22
	f.SetCellValue(sheet, cell("A", row), "Assumptions")
	f.SetCellStyle(sheet, cell("A", row), cell("A", row), headerStyle)
	row++

	assumptions := []struct {
		label string
		value float64
		style int
	}{
		{"IRA Balance", iraBalance, currencyFmt},
		{"Growth Rate", growthRate, pctFmt},
		{"Years to RMD", float64(yearsToRMD), 0},
		{"12% Bracket Space (/yr)", spaceToTop12, currencyFmt},
		{"22% Bracket Space (/yr)", spaceToTop22, currencyFmt},
	}
	for _, a := range assumptions {
		f.SetCellValue(sheet, cell("A", row), a.label)
		c := cell("B", row)
		f.SetCellValue(sheet, c, a.value)
		if a.style != 0 {
			f.SetCellStyle(sheet, c, c, a.style)
		}
		row++
	}
	row++

	// "What You Leave Behind" with formulas
	f.SetCellValue(sheet, cell("A", row), "What You Leave Behind (at age 75)")
	f.SetCellStyle(sheet, cell("A", row), cell("A", row), headerStyle)
	row++
	row++

	writeSubHeader(f, sheet, row, []string{"", "Path A: No Convert", "Path B: 12% Convert", "Path C: 22% Convert"}, headerStyle)
	row++

	// Traditional IRA: Path A = IRA * (1+g)^yrs, Path B/C = MAX(IRA-space*yrs, 0) * (1+g)^yrs
	iraRow := row
	f.SetCellValue(sheet, cell("A", row), "Traditional IRA")
	setCurrencyFormula(f, sheet, "B", row, "$B$7*POWER(1+$B$8,$B$9)", currencyFmt)
	setCurrencyFormula(f, sheet, "C", row, "MAX($B$7-$B$10*$B$9,0)*POWER(1+$B$8,$B$9)", currencyFmt)
	setCurrencyFormula(f, sheet, "D", row, "MAX($B$7-$B$11*$B$9,0)*POWER(1+$B$8,$B$9)", currencyFmt)
	row++

	// Roth: Path A = 0, Path B/C = MIN(space*yrs, IRA) * (1+g)^yrs
	rothRow := row
	f.SetCellValue(sheet, cell("A", row), "Roth IRA")
	setCurrency(f, sheet, "B", row, 0, currencyFmt)
	setCurrencyFormula(f, sheet, "C", row, "MIN($B$10*$B$9,$B$7)*POWER(1+$B$8,$B$9)", currencyFmt)
	setCurrencyFormula(f, sheet, "D", row, "MIN($B$11*$B$9,$B$7)*POWER(1+$B$8,$B$9)", currencyFmt)
	row++

	f.SetCellValue(sheet, cell("A", row), "Conversion Tax Paid")
	setCurrency(f, sheet, "B", row, 0, currencyFmt)
	setCurrency(f, sheet, "C", row, pathBTax, currencyFmt)
	setCurrency(f, sheet, "D", row, pathCTax, currencyFmt)
	row++

	f.SetCellValue(sheet, cell("A", row), "Total Estate Value (IRA + Roth)")
	f.SetCellStyle(sheet, cell("A", row), cell("A", row), headerStyle)
	setCurrencyFormula(f, sheet, "B", row, fmt.Sprintf("B%d+B%d", iraRow, rothRow), currencyFmt)
	setCurrencyFormula(f, sheet, "C", row, fmt.Sprintf("C%d+C%d", iraRow, rothRow), currencyFmt)
	setCurrencyFormula(f, sheet, "D", row, fmt.Sprintf("D%d+D%d", iraRow, rothRow), currencyFmt)
	row++
	row++

	// Heir Tax Impact
	f.SetCellValue(sheet, cell("A", row), "Heir Tax Impact (10-Year Drain, SECURE Act)")
	f.SetCellStyle(sheet, cell("A", row), cell("A", row), headerStyle)
	row++
	row++

	heirScenarios := []struct {
		label  string
		salary float64
	}{
		{"Heir at $300K (32% bracket)", 300_000},
		{"Heir at $500K (35% bracket)", 500_000},
		{"Heir at $750K+ (37% bracket)", 750_000},
	}

	heirStdDed := 30_750.0

	for _, heir := range heirScenarios {
		f.SetCellValue(sheet, cell("A", row), heir.label)
		f.SetCellStyle(sheet, cell("A", row), cell("A", row), headerStyle)
		row++

		writeSubHeader(f, sheet, row, []string{"", "Path A", "Path B", "Path C"}, headerStyle)
		row++

		heirBaseTaxable := math.Max(heir.salary-heirStdDed, 0)

		pathIRAs := []float64{iraNoConvert, iraRemainingB, iraRemainingC}

		// Inherited IRA (reference the "What You Leave Behind" rows)
		inhIRARow := row
		f.SetCellValue(sheet, cell("A", row), "Inherited IRA")
		setCurrencyFormula(f, sheet, "B", row, fmt.Sprintf("B%d", iraRow), currencyFmt)
		setCurrencyFormula(f, sheet, "C", row, fmt.Sprintf("C%d", iraRow), currencyFmt)
		setCurrencyFormula(f, sheet, "D", row, fmt.Sprintf("D%d", iraRow), currencyFmt)
		row++

		inhRothRow := row
		f.SetCellValue(sheet, cell("A", row), "Inherited Roth")
		setCurrencyFormula(f, sheet, "B", row, fmt.Sprintf("B%d", rothRow), currencyFmt)
		setCurrencyFormula(f, sheet, "C", row, fmt.Sprintf("C%d", rothRow), currencyFmt)
		setCurrencyFormula(f, sheet, "D", row, fmt.Sprintf("D%d", rothRow), currencyFmt)
		row++

		f.SetCellValue(sheet, cell("A", row), "IRA Drain: /yr × 10yr")
		setCurrencyFormula(f, sheet, "B", row, fmt.Sprintf("B%d/10", inhIRARow), currencyFmt)
		setCurrencyFormula(f, sheet, "C", row, fmt.Sprintf("C%d/10", inhIRARow), currencyFmt)
		setCurrencyFormula(f, sheet, "D", row, fmt.Sprintf("D%d/10", inhIRARow), currencyFmt)
		row++

		// Tax on drain (computed — bracket math)
		taxRow := row
		f.SetCellValue(sheet, cell("A", row), "Tax on IRA Drain (10yr)")
		for i, col := range []string{"B", "C", "D"} {
			annual := pathIRAs[i] / 10
			fedTax := computeFedTax(heirBaseTaxable+annual) - computeFedTax(heirBaseTaxable)
			caTax := annual * 0.123
			setCurrency(f, sheet, col, row, (fedTax+caTax)*10, currencyFmt)
		}
		row++

		// Heir Keeps = IRA - Tax + Roth
		keepsRow := row
		f.SetCellValue(sheet, cell("A", row), "Heir Keeps (after tax)")
		f.SetCellStyle(sheet, cell("A", row), cell("A", row), headerStyle)
		for _, col := range []string{"B", "C", "D"} {
			setCurrencyFormula(f, sheet, col, row,
				fmt.Sprintf("%s%d-%s%d+%s%d", col, inhIRARow, col, taxRow, col, inhRothRow), currencyFmt)
		}
		row++

		// vs Path A
		f.SetCellValue(sheet, cell("A", row), "vs Path A")
		f.SetCellStyle(sheet, cell("A", row), cell("A", row), headerStyle)
		setCurrency(f, sheet, "B", row, 0, currencyFmt)
		setCurrencyFormula(f, sheet, "C", row, fmt.Sprintf("C%d-B%d", keepsRow, keepsRow), currencyFmt)
		setCurrencyFormula(f, sheet, "D", row, fmt.Sprintf("D%d-B%d", keepsRow, keepsRow), currencyFmt)
		row++
		row++
	}

	row++
	f.SetCellValue(sheet, cell("A", row), "ROI on Conversion Tax")
	f.SetCellStyle(sheet, cell("A", row), cell("A", row), headerStyle)
	row++
	row++

	writeSubHeader(f, sheet, row, []string{"", "Path B", "Path C"}, headerStyle)
	row++

	f.SetCellValue(sheet, cell("A", row), "Your Conversion Tax")
	setCurrency(f, sheet, "B", row, pathBTax, currencyFmt)
	setCurrency(f, sheet, "C", row, pathCTax, currencyFmt)
	row++

	for _, heir := range heirScenarios {
		heirBaseTaxable := math.Max(heir.salary-heirStdDed, 0)

		pathIRAs := []float64{iraNoConvert, iraRemainingB, iraRemainingC}
		pathRoths := []float64{0, rothBalB, rothBalC}

		var heirKeepsA, heirKeepsB, heirKeepsC float64
		for i, ira := range pathIRAs {
			annual := ira / 10
			fedTax := computeFedTax(heirBaseTaxable+annual) - computeFedTax(heirBaseTaxable)
			caTax := annual * 0.123
			keeps := ira - (fedTax+caTax)*10 + pathRoths[i]
			switch i {
			case 0:
				heirKeepsA = keeps
			case 1:
				heirKeepsB = keeps
			case 2:
				heirKeepsC = keeps
			}
		}

		f.SetCellValue(sheet, cell("A", row), fmt.Sprintf("Heir benefit (%s)", heir.label))
		setCurrency(f, sheet, "B", row, heirKeepsB-heirKeepsA, currencyFmt)
		setCurrency(f, sheet, "C", row, heirKeepsC-heirKeepsA, currencyFmt)
		row++
	}
}

func simulateConversions(iraBalance, annualConversion float64, years int, fedBase, caBase float64) (totalConverted, totalTax float64) {
	balance := iraBalance
	for i := 0; i < years; i++ {
		convert := math.Min(annualConversion, balance)
		fedTax := computeFedTax(fedBase+convert) - computeFedTax(fedBase)
		caTax := computeCATax(caBase+convert) - computeCATax(caBase)
		totalConverted += convert
		totalTax += fedTax + caTax
		balance -= convert
	}
	return
}

func computeFedTax(taxable float64) float64 {
	if taxable <= 0 {
		return 0
	}

	brackets := []struct {
		rate float64
		low  float64
		high float64
	}{
		{0.10, 0, 23200},
		{0.12, 23200, 94300},
		{0.22, 94300, 201050},
		{0.24, 201050, 383900},
		{0.32, 383900, 487450},
		{0.35, 487450, 731200},
		{0.37, 731200, math.MaxFloat64},
	}

	var tax float64
	for _, b := range brackets {
		if taxable <= b.low {
			break
		}
		width := math.Min(taxable, b.high) - b.low
		tax += width * b.rate
	}

	return tax
}

func writeTaxWaterfallSheet(f *excelize.File, positions []incomePosition, currencyFmt, pctFmt, headerStyle int) {
	sheet := "Tax Waterfall"
	f.NewSheet(sheet)

	var caMuniIncome, nationalMuniFundIncome float64
	var ordinaryIncome, qualDivIncome float64
	var taxDeferredIncome, rothIncome float64

	for _, p := range positions {
		cat := incomeCategory(p)
		switch cat {
		case "Tax-Exempt Interest (CA Muni)":
			caMuniIncome += p.AnnualIncome
		case "Tax-Exempt Interest (National Muni Fund)":
			nationalMuniFundIncome += p.AnnualIncome
		case "Qualified Dividends":
			qualDivIncome += p.AnnualIncome
		case "Tax-Deferred Income (IRA)":
			taxDeferredIncome += p.AnnualIncome
		case "Tax-Free Income (Roth)":
			rothIncome += p.AnnualIncome
		default:
			ordinaryIncome += p.AnnualIncome
		}
	}

	// Assumptions block: B3=fedStdDed, B4=caStdDed, B5=qdThreshold
	row := 1
	f.SetCellValue(sheet, cell("A", row), "Tax Waterfall")
	f.SetCellStyle(sheet, cell("A", row), cell("A", row), headerStyle)
	row++

	f.SetCellValue(sheet, cell("A", row), "Assumptions")
	f.SetCellStyle(sheet, cell("A", row), cell("A", row), headerStyle)
	row++

	f.SetCellValue(sheet, cell("A", row), "Fed Std Deduction (MFJ, both 65+)")
	setCurrency(f, sheet, "B", row, federalStdDeduction, currencyFmt)
	row++

	f.SetCellValue(sheet, cell("A", row), "CA Std Deduction")
	setCurrency(f, sheet, "B", row, caStdDeduction, currencyFmt)
	row++

	f.SetCellValue(sheet, cell("A", row), "QD 0% Threshold (MFJ)")
	setCurrency(f, sheet, "B", row, qualDivZeroPctThreshold, currencyFmt)
	row++
	row++

	// Section 1: Tax-Exempt
	f.SetCellValue(sheet, cell("A", row), "Section 1: Tax-Exempt Income (Outside AGI)")
	f.SetCellStyle(sheet, cell("A", row), cell("A", row), headerStyle)
	row++
	row++

	writeSubHeader(f, sheet, row, []string{"Source", "Annual Income", "Federal", "California"}, headerStyle)
	row++

	f.SetCellValue(sheet, cell("A", row), "CA Muni Bonds (Individual)")
	setCurrency(f, sheet, "B", row, caMuniIncome, currencyFmt)
	f.SetCellValue(sheet, cell("C", row), "Exempt")
	f.SetCellValue(sheet, cell("D", row), "Exempt")
	row++

	f.SetCellValue(sheet, cell("A", row), "National Muni Funds (MANLX, ITM)")
	setCurrency(f, sheet, "B", row, nationalMuniFundIncome, currencyFmt)
	f.SetCellValue(sheet, cell("C", row), "Exempt")
	f.SetCellValue(sheet, cell("D", row), "Taxable (~non-CA portion)")
	row++
	row++

	// Section 2: Ordinary Income Stacking
	f.SetCellValue(sheet, cell("A", row), "Section 2: Ordinary Income Stacking")
	f.SetCellStyle(sheet, cell("A", row), cell("A", row), headerStyle)
	row++
	row++

	writeSubHeader(f, sheet, row, []string{"Source", "Amount"}, headerStyle)
	row++

	// Track income source rows for SUM formula
	var ordinarySourceRows []int

	f.SetCellValue(sheet, cell("A", row), "Money Market / Savings Interest")
	var mmIncome float64
	for _, p := range positions {
		if incomeCategory(p) == "Taxable Interest" {
			mmIncome += p.AnnualIncome
		}
	}
	setCurrency(f, sheet, "B", row, mmIncome, currencyFmt)
	ordinarySourceRows = append(ordinarySourceRows, row)
	row++

	f.SetCellValue(sheet, cell("A", row), "Ordinary Dividends")
	var ordDivs float64
	for _, p := range positions {
		cat := incomeCategory(p)
		if cat == "Ordinary Dividends" {
			ordDivs += p.AnnualIncome
		}
	}
	setCurrency(f, sheet, "B", row, ordDivs, currencyFmt)
	ordinarySourceRows = append(ordinarySourceRows, row)
	row++

	var nonCABondIncome float64
	for _, p := range positions {
		if incomeCategory(p) == "Bond Coupons (Non-CA Muni)" {
			nonCABondIncome += p.AnnualIncome
		}
	}
	if nonCABondIncome > 0 {
		f.SetCellValue(sheet, cell("A", row), "Non-CA Muni Bond Coupons")
		setCurrency(f, sheet, "B", row, nonCABondIncome, currencyFmt)
		ordinarySourceRows = append(ordinarySourceRows, row)
		row++
	}

	caNonCAPortion := nationalMuniFundIncome * 0.5
	if caNonCAPortion > 0 {
		f.SetCellValue(sheet, cell("A", row), "National Muni Fund (CA taxable portion, ~50%)")
		setCurrency(f, sheet, "B", row, caNonCAPortion, currencyFmt)
		row++
	}

	// Gross Ordinary = SUM of source rows
	grossRow := row
	f.SetCellValue(sheet, cell("A", row), "Gross Ordinary Income")
	f.SetCellStyle(sheet, cell("A", row), cell("A", row), headerStyle)
	sumParts := make([]string, len(ordinarySourceRows))
	for i, r := range ordinarySourceRows {
		sumParts[i] = fmt.Sprintf("B%d", r)
	}
	setCurrencyFormula(f, sheet, "B", row, strings.Join(sumParts, "+"), currencyFmt)
	row++

	// Deduction references assumption cell $B$3
	deductionRow := row
	f.SetCellValue(sheet, cell("A", row), "Less: Standard Deduction (MFJ, both 65+)")
	setCurrencyFormula(f, sheet, "B", row, "-$B$3", currencyFmt)
	row++

	// Taxable Ordinary = MAX(gross - deduction, 0)
	taxableRow := row
	f.SetCellValue(sheet, cell("A", row), "Taxable Ordinary Income (Federal)")
	f.SetCellStyle(sheet, cell("A", row), cell("A", row), headerStyle)
	setCurrencyFormula(f, sheet, "B", row,
		fmt.Sprintf("MAX(B%d+B%d,0)", grossRow, deductionRow), currencyFmt)
	row++
	row++

	// Federal bracket fill with formulas
	f.SetCellValue(sheet, cell("A", row), "Federal Bracket Fill (MFJ 2025)")
	f.SetCellStyle(sheet, cell("A", row), cell("A", row), headerStyle)
	row++
	row++

	writeSubHeader(f, sheet, row, []string{"Bracket", "Range", "Income in Bracket", "Tax"}, headerStyle)
	row++

	fedBrackets := []struct {
		rate  float64
		low   float64
		high  float64
		label string
	}{
		{0.10, 0, 23200, "10%"},
		{0.12, 23200, 94300, "12%"},
		{0.22, 94300, 201050, "22%"},
		{0.24, 201050, 383900, "24%"},
	}

	var bracketTaxRows []int
	for _, b := range fedBrackets {
		width := b.high - b.low

		f.SetCellValue(sheet, cell("A", row), b.label)
		f.SetCellValue(sheet, cell("B", row), fmt.Sprintf("$%s – $%s", formatCompact(b.low), formatCompact(b.high)))

		// Income in bracket: MIN(MAX(taxable-low, 0), width)
		setCurrencyFormula(f, sheet, "C", row,
			fmt.Sprintf("MIN(MAX(B%d-%g,0),%g)", taxableRow, b.low, width), currencyFmt)
		// Tax = income_in_bracket * rate
		setCurrencyFormula(f, sheet, "D", row,
			fmt.Sprintf("C%d*%g", row, b.rate), currencyFmt)

		bracketTaxRows = append(bracketTaxRows, row)
		row++
	}
	row++

	// Section 3: Qualified Dividends
	f.SetCellValue(sheet, cell("A", row), "Section 3: Qualified Dividends (Stack on Top)")
	f.SetCellStyle(sheet, cell("A", row), cell("A", row), headerStyle)
	row++
	row++

	qdRow := row
	f.SetCellValue(sheet, cell("A", row), "Qualified Dividends")
	setCurrency(f, sheet, "B", row, qualDivIncome, currencyFmt)
	row++

	// QD tax: =MIN(MAX(taxable+QD-threshold, 0), QD) * 0.15
	taxableOrdinary := math.Max(mmIncome+ordDivs+nonCABondIncome-federalStdDeduction, 0)
	combinedForQD := taxableOrdinary + qualDivIncome
	if combinedForQD <= qualDivZeroPctThreshold {
		f.SetCellValue(sheet, cell("A", row), "Combined taxable income below $94,050 → 0% rate")
	} else {
		above := combinedForQD - qualDivZeroPctThreshold
		qdAt15 := math.Min(above, qualDivIncome)
		qdAt0 := qualDivIncome - qdAt15
		f.SetCellValue(sheet, cell("A", row), fmt.Sprintf("$%s at 0%%, $%s at 15%%", formatCompact(qdAt0), formatCompact(qdAt15)))
	}
	row++

	fedQDTaxRow := row
	f.SetCellValue(sheet, cell("A", row), "Federal Tax on Qualified Dividends")
	setCurrencyFormula(f, sheet, "B", row,
		fmt.Sprintf("MIN(MAX(B%d+B%d-$B$5,0),B%d)*0.15", taxableRow, qdRow, qdRow), currencyFmt)
	row++
	row++

	// CA taxes (computed — 9 brackets unwieldy in formulas)
	grossOrdinary := mmIncome + ordDivs + nonCABondIncome
	caGrossOrdinary := grossOrdinary + caNonCAPortion
	caTaxableOrdinary := math.Max(caGrossOrdinary-caStdDeduction, 0)
	caTaxOnOrdinary := computeCATax(caTaxableOrdinary)
	caTaxOnQD := computeCATax(caTaxableOrdinary+qualDivIncome) - caTaxOnOrdinary

	// Section 4: Summary
	f.SetCellValue(sheet, cell("A", row), "Section 4: Tax Summary")
	f.SetCellStyle(sheet, cell("A", row), cell("A", row), headerStyle)
	row++
	row++

	writeSubHeader(f, sheet, row, []string{"", "Federal", "California", "Combined"}, headerStyle)
	row++

	// Tax on Ordinary: Fed = SUM of bracket taxes, CA = computed, Combined = formula
	ordTaxRow := row
	f.SetCellValue(sheet, cell("A", row), "Tax on Ordinary Income")
	taxSumParts := make([]string, len(bracketTaxRows))
	for i, r := range bracketTaxRows {
		taxSumParts[i] = fmt.Sprintf("D%d", r)
	}
	setCurrencyFormula(f, sheet, "B", row, strings.Join(taxSumParts, "+"), currencyFmt)
	setCurrency(f, sheet, "C", row, caTaxOnOrdinary, currencyFmt)
	setCurrencyFormula(f, sheet, "D", row, fmt.Sprintf("B%d+C%d", row, row), currencyFmt)
	row++

	qdTaxRow := row
	f.SetCellValue(sheet, cell("A", row), "Tax on Qualified Dividends")
	setCurrencyFormula(f, sheet, "B", row, fmt.Sprintf("B%d", fedQDTaxRow), currencyFmt)
	setCurrency(f, sheet, "C", row, caTaxOnQD, currencyFmt)
	setCurrencyFormula(f, sheet, "D", row, fmt.Sprintf("B%d+C%d", row, row), currencyFmt)
	row++

	totalTaxRow := row
	f.SetCellValue(sheet, cell("A", row), "Total Tax")
	f.SetCellStyle(sheet, cell("A", row), cell("A", row), headerStyle)
	setCurrencyFormula(f, sheet, "B", row, fmt.Sprintf("B%d+B%d", ordTaxRow, qdTaxRow), currencyFmt)
	setCurrencyFormula(f, sheet, "C", row, fmt.Sprintf("C%d+C%d", ordTaxRow, qdTaxRow), currencyFmt)
	setCurrencyFormula(f, sheet, "D", row, fmt.Sprintf("B%d+C%d", row, row), currencyFmt)
	row++
	row++

	var totalGrossIncome float64
	for _, p := range positions {
		totalGrossIncome += p.AnnualIncome
	}

	grossIncomeRow := row
	f.SetCellValue(sheet, cell("A", row), "Total Gross Income")
	setCurrency(f, sheet, "B", row, totalGrossIncome, currencyFmt)
	row++

	f.SetCellValue(sheet, cell("A", row), "Total Net Income (After Tax)")
	f.SetCellStyle(sheet, cell("A", row), cell("A", row), headerStyle)
	setCurrencyFormula(f, sheet, "B", row,
		fmt.Sprintf("B%d-D%d", grossIncomeRow, totalTaxRow), currencyFmt)
	row++

	if totalGrossIncome > 0 {
		f.SetCellValue(sheet, cell("A", row), "Effective Federal Rate")
		setPercentFormula(f, sheet, "B", row,
			fmt.Sprintf("B%d/B%d", totalTaxRow, grossIncomeRow), pctFmt)
		row++
		f.SetCellValue(sheet, cell("A", row), "Effective CA Rate")
		setPercentFormula(f, sheet, "B", row,
			fmt.Sprintf("C%d/B%d", totalTaxRow, grossIncomeRow), pctFmt)
		row++
		f.SetCellValue(sheet, cell("A", row), "Effective Combined Rate")
		setPercentFormula(f, sheet, "B", row,
			fmt.Sprintf("D%d/B%d", totalTaxRow, grossIncomeRow), pctFmt)
	}
}

// CA MFJ brackets 2025
func computeCATax(taxable float64) float64 {
	if taxable <= 0 {
		return 0
	}

	brackets := []struct {
		rate float64
		low  float64
		high float64
	}{
		{0.01, 0, 20824},
		{0.02, 20824, 49368},
		{0.04, 49368, 77918},
		{0.06, 77918, 108162},
		{0.08, 108162, 136700},
		{0.093, 136700, 698274},
		{0.103, 698274, 837922},
		{0.113, 837922, 1_000_000},
		{0.123, 1_000_000, math.MaxFloat64},
	}

	var tax float64
	for _, b := range brackets {
		if taxable <= b.low {
			break
		}
		width := math.Min(taxable, b.high) - b.low
		tax += width * b.rate
	}

	return tax
}

func writeSubHeader(f *excelize.File, sheet string, row int, headers []string, style int) {
	for i, h := range headers {
		c := cell(string(rune('A'+i)), row)
		f.SetCellValue(sheet, c, h)
		f.SetCellStyle(sheet, c, c, style)
	}
}

func formatCompact(v float64) string {
	if v >= 1_000_000 {
		return fmt.Sprintf("%.1fM", v/1_000_000)
	}
	if v >= 1000 {
		return fmt.Sprintf("%.0fK", v/1000)
	}
	return fmt.Sprintf("%.0f", v)
}
