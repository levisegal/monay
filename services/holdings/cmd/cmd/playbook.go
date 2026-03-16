package cmd

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"

	"github.com/levisegal/monay/services/holdings/gen/db"
)

// MFJ 2026 bracket ceilings (taxable income, not AGI)
var bracketCeilings = map[string]float64{
	"10%": 23200,
	"12%": 94300,
	"22%": 201050,
	"24%": 394600,
	"32%": 501050,
	"35%": 751600,
}

const (
	defaultTargetBracket       = "22%"
	defaultOtherIncome         = 0.0
	positionConcentrationLimit = 0.03
	defaultSectorCap           = 0.15
	defaultTargetEquityPct     = 0.34
	defaultTargetFixedPct      = 0.50
	defaultTargetCashPct       = 0.16

	fedStdDed2026 = 35400.0
	caStdDed2026  = 10726.0
	niitThreshold = 250000.0
	niitRate      = 0.038
)

type irmaaThreshold struct {
	magi     float64
	monthlyB float64
	monthlyD float64
}

var irmaaThresholds = []irmaaThreshold{
	{206000, 0, 0},
	{258000, 70.00, 13.70},
	{322000, 175.00, 35.50},
	{386000, 280.00, 57.30},
	{750000, 384.00, 79.10},
}

func writePlaybookSheet(f *excelize.File, data *reportData, currencyFmt, pctFmt, headerStyle int) {
	sheet := "Playbook"
	f.NewSheet(sheet)

	yellowFill, _ := f.NewStyle(&excelize.Style{
		Fill: excelize.Fill{Type: "pattern", Color: []string{"FFFF00"}, Pattern: 1},
	})
	yellowCurrency, _ := f.NewStyle(&excelize.Style{
		NumFmt: 4,
		Fill:   excelize.Fill{Type: "pattern", Color: []string{"FFFF00"}, Pattern: 1},
	})
	yellowPct, _ := f.NewStyle(&excelize.Style{
		NumFmt: 10,
		Fill:   excelize.Fill{Type: "pattern", Color: []string{"FFFF00"}, Pattern: 1},
	})
	portfolioIncome := computePortfolioOrdinaryIncome(data)
	bracketCeiling := bracketCeilings[defaultTargetBracket]

	row := 1

	f.SetCellValue(sheet, cell("A", row), "Control Panel")
	f.SetCellStyle(sheet, cell("A", row), cell("A", row), headerStyle)
	row++

	f.SetCellValue(sheet, cell("A", row), "Portfolio Ordinary Income")
	setCurrency(f, sheet, "B", row, portfolioIncome, currencyFmt)
	portfolioIncomeRow := row
	row++

	f.SetCellValue(sheet, cell("A", row), "Other Income (SS, pensions, W-2)")
	setCurrency(f, sheet, "B", row, defaultOtherIncome, yellowCurrency)
	otherIncomeRow := row
	row++

	f.SetCellValue(sheet, cell("A", row), "Target Bracket")
	f.SetCellValue(sheet, cell("B", row), defaultTargetBracket)
	f.SetCellStyle(sheet, cell("B", row), cell("B", row), yellowFill)
	row++

	f.SetCellValue(sheet, cell("A", row), "Position Cap")
	setPercent(f, sheet, "B", row, positionConcentrationLimit, yellowPct)
	row++

	f.SetCellValue(sheet, cell("A", row), "Sector Cap")
	setPercent(f, sheet, "B", row, defaultSectorCap, yellowPct)
	row++

	f.SetCellValue(sheet, cell("A", row), "Target Equity %")
	setPercent(f, sheet, "B", row, defaultTargetEquityPct, yellowPct)
	row++

	f.SetCellValue(sheet, cell("A", row), "Target Fixed Income %")
	setPercent(f, sheet, "B", row, defaultTargetFixedPct, yellowPct)
	row++

	f.SetCellValue(sheet, cell("A", row), "Target Cash/Other %")
	setPercent(f, sheet, "B", row, defaultTargetCashPct, yellowPct)
	row++
	row++

	// Risk Dashboard
	if data.analysis != nil && data.analysis.RiskMetrics.AnnualizedVolatility != nil {
		f.SetCellValue(sheet, cell("A", row), "Risk Dashboard")
		f.SetCellStyle(sheet, cell("A", row), cell("A", row), headerStyle)
		row++

		writeSubHeader(f, sheet, row, []string{"Metric", "Current", "After Rebalance"}, headerStyle)
		row++

		rm := data.analysis.RiskMetrics
		trm := data.analysis.TargetRiskMetrics

		writeSubHeader(f, sheet, row, []string{"Metric", "Current", "After Rebalance", "What It Means"}, headerStyle)
		row++

		riskRows := []struct {
			label   string
			current *float64
			target  *float64
			explain string
		}{
			{"Annualized Volatility", rm.AnnualizedVolatility, trm.AnnualizedVolatility,
				"How much the portfolio swings in a year. Lower = smoother ride."},
			{"CVaR 95%", rm.CVaR95, trm.CVaR95,
				"Average loss in the worst 5% of days, annualized. Your realistic bad-case scenario."},
			{"Max Drawdown", rm.MaxDrawdown, trm.MaxDrawdown,
				"Largest peak-to-trough drop over 2 years. How much you'd have lost at the worst point."},
			{"Sharpe Ratio", rm.SharpeRatio, trm.SharpeRatio,
				"Return per unit of risk (vs T-bills). Above 1.0 is good, above 2.0 is excellent."},
		}

		for _, r := range riskRows {
			f.SetCellValue(sheet, cell("A", row), r.label)
			if r.current != nil {
				if r.label == "Sharpe Ratio" {
					f.SetCellValue(sheet, cell("B", row), fmt.Sprintf("%.2f", *r.current))
					if r.target != nil {
						f.SetCellValue(sheet, cell("C", row), fmt.Sprintf("%.2f", *r.target))
					}
				} else {
					setPercent(f, sheet, "B", row, *r.current, pctFmt)
					if r.target != nil {
						setPercent(f, sheet, "C", row, *r.target, pctFmt)
					}
				}
			}
			f.SetCellValue(sheet, cell("D", row), r.explain)
			row++
		}

		if rm.AnnualizedVolatility != nil && trm.AnnualizedVolatility != nil {
			f.SetCellValue(sheet, cell("A", row),
				fmt.Sprintf("Rebalancing reduces annual volatility from %.1f%% to %.1f%%",
					*rm.AnnualizedVolatility*100, *trm.AnnualizedVolatility*100))
		}
		row++
		row++
	}

	// Investment Theses
	if data.analysis != nil && len(data.analysis.ConvictionStatuses) > 0 {
		f.SetCellValue(sheet, cell("A", row), "Investment Theses")
		f.SetCellStyle(sheet, cell("A", row), cell("A", row), headerStyle)
		row++

		writeSubHeader(f, sheet, row, []string{"Type", "Name", "Thesis", "Target", "Current", "Status"}, headerStyle)
		row++

		for _, cs := range data.analysis.ConvictionStatuses {
			f.SetCellValue(sheet, cell("A", row), cs.Type)
			f.SetCellValue(sheet, cell("B", row), cs.Name)
			f.SetCellValue(sheet, cell("C", row), cs.Thesis)
			setPercent(f, sheet, "D", row, cs.TargetWeight, pctFmt)
			setPercent(f, sheet, "E", row, cs.CurrentWeight, pctFmt)
			f.SetCellValue(sheet, cell("F", row), cs.Status)
			row++
		}

		if data.convictions != nil {
			for name, strat := range data.convictions.Strategies {
				income := computeStrategyIncome(data, strat.Instruments)
				if income > 0 {
					f.SetCellValue(sheet, cell("A", row),
						fmt.Sprintf("%s — estimated annual income: %s from qualifying holdings", name, formatCurrency(income)))
					row++
				}
			}
		}

		row++
	}

	// Correlation Clusters
	if data.analysis != nil && len(data.analysis.CorrelationClusters) > 0 {
		f.SetCellValue(sheet, cell("A", row), "Correlation Clusters")
		f.SetCellStyle(sheet, cell("A", row), cell("A", row), headerStyle)
		row++

		f.SetCellValue(sheet, cell("A", row), "These positions move together — individual weights understate concentration")
		row++

		writeSubHeader(f, sheet, row, []string{"Symbols", "Avg Correlation", "Combined Weight"}, headerStyle)
		row++

		for _, c := range data.analysis.CorrelationClusters {
			f.SetCellValue(sheet, cell("A", row), strings.Join(c.Symbols, ", "))
			f.SetCellValue(sheet, cell("B", row), fmt.Sprintf("%.2f", c.AvgCorrelation))
			setPercent(f, sheet, "C", row, c.CombinedWeight, pctFmt)
			row++
		}
		row++
	}

	// Constraint Violations
	if data.analysis != nil && len(data.analysis.ConstraintViolations) > 0 {
		f.SetCellValue(sheet, cell("A", row), "Constraint Violations")
		f.SetCellStyle(sheet, cell("A", row), cell("A", row), headerStyle)
		row++

		writeSubHeader(f, sheet, row, []string{"Type", "Name", "Current", "Limit", "Excess"}, headerStyle)
		row++

		for _, v := range data.analysis.ConstraintViolations {
			f.SetCellValue(sheet, cell("A", row), v.ViolationType)
			f.SetCellValue(sheet, cell("B", row), v.Name)
			setPercent(f, sheet, "C", row, v.Current, pctFmt)
			setPercent(f, sheet, "D", row, v.Limit, pctFmt)
			setPercent(f, sheet, "E", row, v.Excess, pctFmt)
			row++
		}
		row++
	}

	// Asset Location Issues
	if data.analysis != nil && len(data.analysis.LocationIssues) > 0 {
		f.SetCellValue(sheet, cell("A", row), "Asset Location Issues")
		f.SetCellStyle(sheet, cell("A", row), cell("A", row), headerStyle)
		row++

		f.SetCellValue(sheet, cell("A", row), "Fix these before converting — don't pay conversion tax on misallocated assets")
		row++

		writeSubHeader(f, sheet, row, []string{"Symbol", "Current Account", "Recommended Account", "Reason", "Value"}, headerStyle)
		row++

		for _, li := range data.analysis.LocationIssues {
			f.SetCellValue(sheet, cell("A", row), li.Symbol)
			f.SetCellValue(sheet, cell("B", row), formatAccountType(li.CurrentAccountType))
			f.SetCellValue(sheet, cell("C", row), formatAccountType(li.RecommendedAccountType))
			f.SetCellValue(sheet, cell("D", row), li.Reason)
			setCurrency(f, sheet, "E", row, li.Value, currencyFmt)
			row++
		}
		row++
	}

	// Your Situation
	f.SetCellValue(sheet, cell("A", row), "Your Situation")
	f.SetCellStyle(sheet, cell("A", row), cell("A", row), headerStyle)
	row++

	f.SetCellValue(sheet, cell("A", row), "Total Portfolio Value")
	setCurrency(f, sheet, "B", row, data.totalValue, currencyFmt)
	row++

	taxableValue, iraValue, rothValue := computeAccountTypeValues(data)

	f.SetCellValue(sheet, cell("A", row), "Taxable Accounts")
	setCurrency(f, sheet, "B", row, taxableValue, currencyFmt)
	row++

	f.SetCellValue(sheet, cell("A", row), "Traditional IRA + SEP IRA")
	setCurrency(f, sheet, "B", row, iraValue, currencyFmt)
	row++

	f.SetCellValue(sheet, cell("A", row), "Roth IRA")
	setCurrency(f, sheet, "B", row, rothValue, currencyFmt)
	row++
	row++

	f.SetCellValue(sheet, cell("A", row), "Total Ordinary Income")
	setCurrencyFormula(f, sheet, "B", row,
		fmt.Sprintf("B%d+B%d", portfolioIncomeRow, otherIncomeRow), currencyFmt)
	row++

	totalOrdinary := portfolioIncome + defaultOtherIncome
	fedTaxableIncome := math.Max(totalOrdinary-fedStdDed2026, 0)

	f.SetCellValue(sheet, cell("A", row), "Fed Taxable Income (before conversion)")
	setCurrency(f, sheet, "B", row, fedTaxableIncome, currencyFmt)
	fedTaxableRow := row
	row++

	f.SetCellValue(sheet, cell("A", row), "Bracket Ceiling (MFJ 2026)")
	setCurrency(f, sheet, "B", row, bracketCeiling, currencyFmt)
	bracketCeilingRow := row
	row++

	f.SetCellValue(sheet, cell("A", row), "Bracket Space (available for conversion)")
	setCurrencyFormula(f, sheet, "B", row,
		fmt.Sprintf("MAX(B%d-B%d,0)", bracketCeilingRow, fedTaxableRow), currencyFmt)
	bracketSpaceRow := row
	row++
	row++

	// MAGI & Conversion Impact
	conversionAmt := math.Max(bracketCeiling-fedTaxableIncome, 0)
	projectedMAGI := totalOrdinary + conversionAmt

	f.SetCellValue(sheet, cell("A", row), "MAGI & Conversion Impact")
	f.SetCellStyle(sheet, cell("A", row), cell("A", row), headerStyle)
	row++

	f.SetCellValue(sheet, cell("A", row), "Projected MAGI (income + conversion)")
	setCurrency(f, sheet, "B", row, projectedMAGI, currencyFmt)
	row++

	currentIRMAA, nextCliff := findIRMAABracket(projectedMAGI)
	f.SetCellValue(sheet, cell("A", row), "IRMAA Bracket")
	if currentIRMAA.monthlyB > 0 {
		annualSurcharge := (currentIRMAA.monthlyB + currentIRMAA.monthlyD) * 12 * 2
		f.SetCellValue(sheet, cell("B", row), fmt.Sprintf("$%.0f/yr surcharge (2 people)", annualSurcharge))
	} else {
		f.SetCellValue(sheet, cell("B", row), "Base — no surcharge")
	}
	row++

	if nextCliff > 0 {
		distance := nextCliff - projectedMAGI
		f.SetCellValue(sheet, cell("A", row), "Distance to Next IRMAA Cliff")
		setCurrency(f, sheet, "B", row, distance, currencyFmt)
		if distance < 0 {
			f.SetCellValue(sheet, cell("C", row), "WARNING: conversion crosses IRMAA threshold")
		}
		row++
	}

	if projectedMAGI > niitThreshold {
		niitCost := (projectedMAGI - niitThreshold) * niitRate
		f.SetCellValue(sheet, cell("A", row), "NIIT (3.8% on investment income above $250k)")
		setCurrency(f, sheet, "B", row, niitCost, currencyFmt)
		row++
	}
	row++

	// Roth Conversion Plan
	f.SetCellValue(sheet, cell("A", row), "Roth Conversion Plan")
	f.SetCellStyle(sheet, cell("A", row), cell("A", row), headerStyle)
	row++

	f.SetCellValue(sheet, cell("A", row), "Conversion Amount")
	setCurrencyFormula(f, sheet, "B", row, fmt.Sprintf("B%d", bracketSpaceRow), currencyFmt)
	conversionAmtRow := row
	row++

	caBaseTaxable := math.Max(totalOrdinary-caStdDed2026, 0)
	fedTaxOnConv := computeFedTax(fedTaxableIncome+conversionAmt) - computeFedTax(fedTaxableIncome)
	caTaxOnConv := computeCATax(caBaseTaxable+conversionAmt) - computeCATax(caBaseTaxable)

	f.SetCellValue(sheet, cell("A", row), "Fed Tax on Conversion")
	setCurrency(f, sheet, "B", row, fedTaxOnConv, currencyFmt)
	fedTaxRow := row
	row++

	f.SetCellValue(sheet, cell("A", row), "CA Tax on Conversion")
	setCurrency(f, sheet, "B", row, caTaxOnConv, currencyFmt)
	caTaxRow := row
	row++

	f.SetCellValue(sheet, cell("A", row), "Total Conversion Tax")
	setCurrencyFormula(f, sheet, "B", row, fmt.Sprintf("B%d+B%d", fedTaxRow, caTaxRow), currencyFmt)
	totalConvTaxRow := row
	row++

	f.SetCellValue(sheet, cell("A", row), "Effective Rate")
	setPercentFormula(f, sheet, "B", row,
		fmt.Sprintf("IF(B%d>0,B%d/B%d,0)", conversionAmtRow, totalConvTaxRow, conversionAmtRow), pctFmt)
	row++

	unrealizedST, unrealizedLT := computeUnrealizedByPeriod(data)
	harvestableTotal := -(unrealizedST.losses + unrealizedLT.losses)
	if harvestableTotal > 0 {
		fedLTCGRate := 0.15
		caCGRate := 0.093
		fedSTRate := 0.24
		savings := -unrealizedST.losses*(fedSTRate+caCGRate) + -unrealizedLT.losses*(fedLTCGRate+caCGRate)
		f.SetCellValue(sheet, cell("A", row), fmt.Sprintf("Harvest %s in losses first to reduce conversion tax by %s", formatCurrency(harvestableTotal), formatCurrency(savings)))
	}
	row++
	row++

	// 2026 Tax Bill — Realized Gains
	year := time.Now().Year()
	f.SetCellValue(sheet, cell("A", row), fmt.Sprintf("%d Capital Gains Tax Exposure", year))
	f.SetCellStyle(sheet, cell("A", row), cell("A", row), headerStyle)
	row++
	f.SetCellValue(sheet, cell("A", row), "Realized gains from sales this year create a tax bill due Apr 2027. Harvesting losses NOW reduces it.")
	row++
	row++

	stGains := float64(toInt64Val(data.realizedGL.StGains)) / 1_000_000
	stLosses := float64(toInt64Val(data.realizedGL.StLosses)) / 1_000_000
	ltGains := float64(toInt64Val(data.realizedGL.LtGains)) / 1_000_000
	ltLosses := float64(toInt64Val(data.realizedGL.LtLosses)) / 1_000_000

	fedLTCGRate := 0.15
	caCGRate := 0.093
	fedSTRate := 0.24

	netST := stGains + stLosses
	netLT := ltGains + ltLosses

	writeSubHeader(f, sheet, row, []string{"", "Short-Term", "Long-Term", "Total"}, headerStyle)
	row++

	f.SetCellValue(sheet, cell("A", row), fmt.Sprintf("%d Realized Gains", year))
	setCurrency(f, sheet, "B", row, stGains, currencyFmt)
	setCurrency(f, sheet, "C", row, ltGains, currencyFmt)
	setCurrencyFormula(f, sheet, "D", row, fmt.Sprintf("B%d+C%d", row, row), currencyFmt)
	row++

	f.SetCellValue(sheet, cell("A", row), fmt.Sprintf("%d Realized Losses", year))
	setCurrency(f, sheet, "B", row, stLosses, currencyFmt)
	setCurrency(f, sheet, "C", row, ltLosses, currencyFmt)
	setCurrencyFormula(f, sheet, "D", row, fmt.Sprintf("B%d+C%d", row, row), currencyFmt)
	row++

	f.SetCellValue(sheet, cell("A", row), "Net Realized")
	f.SetCellStyle(sheet, cell("A", row), cell("A", row), headerStyle)
	setCurrency(f, sheet, "B", row, netST, currencyFmt)
	setCurrency(f, sheet, "C", row, netLT, currencyFmt)
	setCurrency(f, sheet, "D", row, netST+netLT, currencyFmt)
	row++
	row++

	// Tax owed on realized gains
	stTax := 0.0
	if netST > 0 {
		stTax = netST * (fedSTRate + caCGRate)
	}
	ltTax := 0.0
	if netLT > 0 {
		ltTax = netLT * (fedLTCGRate + caCGRate)
	}
	totalTax := stTax + ltTax

	f.SetCellValue(sheet, cell("A", row), fmt.Sprintf("Tax Owed on %d Gains (Fed + CA)", year))
	f.SetCellStyle(sheet, cell("A", row), cell("A", row), headerStyle)
	row++

	f.SetCellValue(sheet, cell("A", row), fmt.Sprintf("ST gains taxed at %.1f%% (%.0f%% fed ordinary + %.1f%% CA)", (fedSTRate+caCGRate)*100, fedSTRate*100, caCGRate*100))
	setCurrency(f, sheet, "D", row, stTax, currencyFmt)
	row++
	f.SetCellValue(sheet, cell("A", row), fmt.Sprintf("LT gains taxed at %.1f%% (%.0f%% fed LTCG + %.1f%% CA)", (fedLTCGRate+caCGRate)*100, fedLTCGRate*100, caCGRate*100))
	setCurrency(f, sheet, "D", row, ltTax, currencyFmt)
	row++
	f.SetCellValue(sheet, cell("A", row), fmt.Sprintf("Total %d Capital Gains Tax", year))
	f.SetCellStyle(sheet, cell("A", row), cell("A", row), headerStyle)
	setCurrency(f, sheet, "D", row, totalTax, currencyFmt)
	row++
	row++

	// Harvesting opportunity
	f.SetCellValue(sheet, cell("A", row), "Tax Loss Harvesting (excludes bonds and non-traded REITs)")
	f.SetCellStyle(sheet, cell("A", row), cell("A", row), headerStyle)
	row++
	f.SetCellValue(sheet, cell("A", row), "Sell losing positions now. Losses offset gains dollar-for-dollar on your annual return.")
	row++
	row++

	writeSubHeader(f, sheet, row, []string{"", "Short-Term", "Long-Term", "Total"}, headerStyle)
	row++

	pbImpact := computeHarvestingImpact(netST, netLT, unrealizedST, unrealizedLT, fedSTRate, fedLTCGRate, caCGRate)

	f.SetCellValue(sheet, cell("A", row), "Harvestable Losses (unrealized, taxable)")
	setCurrency(f, sheet, "B", row, unrealizedST.losses, currencyFmt)
	setCurrency(f, sheet, "C", row, unrealizedLT.losses, currencyFmt)
	setCurrencyFormula(f, sheet, "D", row, fmt.Sprintf("B%d+C%d", row, row), currencyFmt)
	row++

	f.SetCellValue(sheet, cell("A", row), "Tax Savings from Harvesting")
	setCurrency(f, sheet, "D", row, pbImpact.totalSavings, currencyFmt)
	row++

	f.SetCellValue(sheet, cell("A", row), fmt.Sprintf("Net %d Tax After Harvesting", year))
	f.SetCellStyle(sheet, cell("A", row), cell("A", row), headerStyle)
	setCurrency(f, sheet, "D", row, pbImpact.taxAfter, currencyFmt)
	row++
	row++

	f.SetCellValue(sheet, cell("A", row), "See 'Tax Harvesting' sheet for full netting detail and sales history")
	row++
	f.SetCellValue(sheet, cell("A", row), "Wash sale rule: do NOT repurchase same/similar security within 30 days")
	row++
	f.SetCellValue(sheet, cell("A", row), "Applies across ALL accounts including IRAs (Rev. Ruling 2008-5)")
	row++
	row++

	// Action Steps
	f.SetCellValue(sheet, cell("A", row), "Action Steps")
	f.SetCellStyle(sheet, cell("A", row), cell("A", row), headerStyle)
	row++

	estimatedTaxDeadline := nextEstimatedTaxDeadlineAfterConversion()
	steps := []string{
		fmt.Sprintf("1. Harvest %s in equity/fund losses to offset %d realized gains (see Tax Exposure section above)", formatCurrency(harvestableTotal), time.Now().Year()),
		"2. Fix asset location issues inside IRA — sell munis, buy bonds (tax-free)",
		"3. Rebalance inside IRA — trim overweight positions (tax-free)",
		fmt.Sprintf("4. Convert %s from rebalanced IRA to Roth (high-growth first, munis never)", formatCurrency(conversionAmt)),
		fmt.Sprintf("5. Pay estimated tax of %s by %s for the quarter you convert", formatCurrency(fedTaxOnConv+caTaxOnConv), estimatedTaxDeadline),
		"   Harvested losses from step 1 reduce taxable income on your annual return, not the estimated payment itself",
		"6. Consider trimming overweight taxable positions (evaluate tax impact)",
	}
	for _, s := range steps {
		f.SetCellValue(sheet, cell("A", row), s)
		row++
	}
}

func writeTaxHarvestingSheet(f *excelize.File, data *reportData, currencyFmt, pctFmt, headerStyle int) {
	sheet := "Tax Harvesting"
	f.NewSheet(sheet)

	year := time.Now().Year()
	fedLTCGRate := 0.15
	caCGRate := 0.093
	fedSTRate := 0.24

	row := 1

	f.SetCellValue(sheet, cell("A", row), fmt.Sprintf("%d Capital Gains & Tax Loss Harvesting", year))
	f.SetCellStyle(sheet, cell("A", row), cell("A", row), headerStyle)
	row++
	f.SetCellValue(sheet, cell("A", row), "Excludes bonds and non-traded REITs (illiquid, hold to maturity).")
	row++
	row++

	stGains := float64(toInt64Val(data.realizedGL.StGains)) / 1_000_000
	stLosses := float64(toInt64Val(data.realizedGL.StLosses)) / 1_000_000
	ltGains := float64(toInt64Val(data.realizedGL.LtGains)) / 1_000_000
	ltLosses := float64(toInt64Val(data.realizedGL.LtLosses)) / 1_000_000

	netST := stGains + stLosses
	netLT := ltGains + ltLosses
	stTax := math.Max(netST, 0) * (fedSTRate + caCGRate)
	ltTax := math.Max(netLT, 0) * (fedLTCGRate + caCGRate)
	totalTax := stTax + ltTax

	f.SetCellValue(sheet, cell("A", row), fmt.Sprintf("Realized Gains/Losses YTD (%d)", year))
	f.SetCellStyle(sheet, cell("A", row), cell("A", row), headerStyle)
	row++

	writeSubHeader(f, sheet, row, []string{"", "Gains", "Losses", "Net", "Tax Rate", "Tax Owed"}, headerStyle)
	row++

	f.SetCellValue(sheet, cell("A", row), "Short-Term")
	setCurrency(f, sheet, "B", row, stGains, currencyFmt)
	setCurrency(f, sheet, "C", row, stLosses, currencyFmt)
	setCurrency(f, sheet, "D", row, netST, currencyFmt)
	setPercent(f, sheet, "E", row, fedSTRate+caCGRate, pctFmt)
	setCurrency(f, sheet, "F", row, stTax, currencyFmt)
	row++

	f.SetCellValue(sheet, cell("A", row), "Long-Term")
	setCurrency(f, sheet, "B", row, ltGains, currencyFmt)
	setCurrency(f, sheet, "C", row, ltLosses, currencyFmt)
	setCurrency(f, sheet, "D", row, netLT, currencyFmt)
	setPercent(f, sheet, "E", row, fedLTCGRate+caCGRate, pctFmt)
	setCurrency(f, sheet, "F", row, ltTax, currencyFmt)
	row++

	f.SetCellValue(sheet, cell("A", row), "Total")
	f.SetCellStyle(sheet, cell("A", row), cell("A", row), headerStyle)
	setCurrencyFormula(f, sheet, "B", row, fmt.Sprintf("B%d+B%d", row-2, row-1), currencyFmt)
	setCurrencyFormula(f, sheet, "C", row, fmt.Sprintf("C%d+C%d", row-2, row-1), currencyFmt)
	setCurrencyFormula(f, sheet, "D", row, fmt.Sprintf("D%d+D%d", row-2, row-1), currencyFmt)
	setCurrency(f, sheet, "F", row, totalTax, currencyFmt)
	row++
	row++

	unrealizedST, unrealizedLT := computeUnrealizedByPeriod(data)

	f.SetCellValue(sheet, cell("A", row), "Unrealized Gains/Losses (Taxable Accounts)")
	f.SetCellStyle(sheet, cell("A", row), cell("A", row), headerStyle)
	row++

	writeSubHeader(f, sheet, row, []string{"", "Gains", "Losses", "Net"}, headerStyle)
	row++

	f.SetCellValue(sheet, cell("A", row), "Short-Term")
	setCurrency(f, sheet, "B", row, unrealizedST.gains, currencyFmt)
	setCurrency(f, sheet, "C", row, unrealizedST.losses, currencyFmt)
	setCurrencyFormula(f, sheet, "D", row, fmt.Sprintf("B%d+C%d", row, row), currencyFmt)
	row++

	f.SetCellValue(sheet, cell("A", row), "Long-Term")
	setCurrency(f, sheet, "B", row, unrealizedLT.gains, currencyFmt)
	setCurrency(f, sheet, "C", row, unrealizedLT.losses, currencyFmt)
	setCurrencyFormula(f, sheet, "D", row, fmt.Sprintf("B%d+C%d", row, row), currencyFmt)
	row++

	f.SetCellValue(sheet, cell("A", row), "Total")
	f.SetCellStyle(sheet, cell("A", row), cell("A", row), headerStyle)
	setCurrencyFormula(f, sheet, "B", row, fmt.Sprintf("B%d+B%d", row-2, row-1), currencyFmt)
	setCurrencyFormula(f, sheet, "C", row, fmt.Sprintf("C%d+C%d", row-2, row-1), currencyFmt)
	setCurrencyFormula(f, sheet, "D", row, fmt.Sprintf("D%d+D%d", row-2, row-1), currencyFmt)
	row++
	row++

	impact := computeHarvestingImpact(netST, netLT, unrealizedST, unrealizedLT, fedSTRate, fedLTCGRate, caCGRate)

	f.SetCellValue(sheet, cell("A", row), "Harvesting Impact")
	f.SetCellStyle(sheet, cell("A", row), cell("A", row), headerStyle)
	row++

	f.SetCellValue(sheet, cell("A", row), "ST losses offset ST gains first, then excess offsets LT gains. LT losses offset LT gains.")
	row++
	row++

	writeSubHeader(f, sheet, row, []string{"", "Short-Term", "Long-Term", "Total"}, headerStyle)
	row++

	f.SetCellValue(sheet, cell("A", row), "Realized Gains (net)")
	setCurrency(f, sheet, "B", row, netST, currencyFmt)
	setCurrency(f, sheet, "C", row, netLT, currencyFmt)
	setCurrency(f, sheet, "D", row, netST+netLT, currencyFmt)
	row++

	f.SetCellValue(sheet, cell("A", row), "Harvestable Losses")
	setCurrency(f, sheet, "B", row, unrealizedST.losses, currencyFmt)
	setCurrency(f, sheet, "C", row, unrealizedLT.losses, currencyFmt)
	setCurrency(f, sheet, "D", row, unrealizedST.losses+unrealizedLT.losses, currencyFmt)
	row++

	f.SetCellValue(sheet, cell("A", row), "Net After Harvesting")
	setCurrency(f, sheet, "B", row, impact.netSTAfter, currencyFmt)
	setCurrency(f, sheet, "C", row, impact.netLTAfter, currencyFmt)
	setCurrency(f, sheet, "D", row, impact.netSTAfter+impact.netLTAfter, currencyFmt)
	row++
	row++

	f.SetCellValue(sheet, cell("A", row), fmt.Sprintf("%d Tax Before Harvesting", year))
	setCurrency(f, sheet, "D", row, totalTax, currencyFmt)
	row++

	f.SetCellValue(sheet, cell("A", row), "Tax Savings from Harvesting")
	setCurrency(f, sheet, "D", row, impact.totalSavings, currencyFmt)
	row++

	f.SetCellValue(sheet, cell("A", row), fmt.Sprintf("%d Tax After Harvesting", year))
	f.SetCellStyle(sheet, cell("A", row), cell("A", row), headerStyle)
	setCurrency(f, sheet, "D", row, impact.taxAfter, currencyFmt)
	row++
	row++

	f.SetCellValue(sheet, cell("A", row), fmt.Sprintf("Realized Sales Detail (%d)", year))
	f.SetCellStyle(sheet, cell("A", row), cell("A", row), headerStyle)
	row++

	saleHeaders := []string{"Date", "Account", "Symbol", "Name", "Period", "Qty", "Cost Basis", "Proceeds", "Gain/Loss"}
	writeSubHeader(f, sheet, row, saleHeaders, headerStyle)
	row++

	saleFirstRow := row
	for _, d := range data.dispositions {
		if !isLiquid(d.SecurityType.String) {
			continue
		}
		qty := float64(d.QuantityMicros) / 1_000_000
		cost := float64(d.CostBasisMicros) / 1_000_000
		proceeds := float64(d.ProceedsMicros) / 1_000_000
		gain := float64(d.RealizedGainMicros) / 1_000_000
		period := "LT"
		if d.HoldingPeriod == "short_term" {
			period = "ST"
		}

		f.SetCellValue(sheet, cell("A", row), d.DisposedDate)
		f.SetCellValue(sheet, cell("B", row), d.AccountName)
		f.SetCellValue(sheet, cell("C", row), d.Symbol)
		f.SetCellValue(sheet, cell("D", row), d.SecurityName.String)
		f.SetCellValue(sheet, cell("E", row), period)
		f.SetCellValue(sheet, cell("F", row), qty)
		setCurrency(f, sheet, "G", row, cost, currencyFmt)
		setCurrency(f, sheet, "H", row, proceeds, currencyFmt)
		setCurrency(f, sheet, "I", row, gain, currencyFmt)
		row++
	}
	saleLastRow := row - 1

	f.SetCellValue(sheet, cell("A", row), "TOTAL")
	f.SetCellStyle(sheet, cell("A", row), cell("A", row), headerStyle)
	if saleFirstRow <= saleLastRow {
		setCurrencyFormula(f, sheet, "G", row, fmt.Sprintf("SUM(G%d:G%d)", saleFirstRow, saleLastRow), currencyFmt)
		setCurrencyFormula(f, sheet, "H", row, fmt.Sprintf("SUM(H%d:H%d)", saleFirstRow, saleLastRow), currencyFmt)
		setCurrencyFormula(f, sheet, "I", row, fmt.Sprintf("SUM(I%d:I%d)", saleFirstRow, saleLastRow), currencyFmt)
	}
	row++
	row++

	lotRows := buildLotRows(data)

	type positionSummary struct {
		symbol  string
		name    string
		stGain  float64
		stLoss  float64
		ltGain  float64
		ltLoss  float64
		mktVal  float64
		cost    float64
		estBasis bool
	}

	posMap := make(map[string]*positionSummary)
	for _, r := range lotRows {
		unrealized := r.mktVal - r.cost
		ps, ok := posMap[r.lot.Symbol]
		if !ok {
			ps = &positionSummary{
				symbol: r.lot.Symbol,
				name:   r.lot.SecurityName.String,
			}
			posMap[r.lot.Symbol] = ps
		}
		ps.mktVal += r.mktVal
		ps.cost += r.cost
		if r.estimated {
			ps.estBasis = true
		}
		if r.period == "ST" {
			if unrealized < 0 {
				ps.stLoss += unrealized
			} else {
				ps.stGain += unrealized
			}
		} else {
			if unrealized < 0 {
				ps.ltLoss += unrealized
			} else {
				ps.ltGain += unrealized
			}
		}
	}

	var positions []*positionSummary
	for _, ps := range posMap {
		positions = append(positions, ps)
	}
	sort.Slice(positions, func(i, j int) bool {
		ni := positions[i].stLoss + positions[i].ltLoss
		nj := positions[j].stLoss + positions[j].ltLoss
		return ni < nj
	})

	f.SetCellValue(sheet, cell("A", row), "By Position (Taxable Accounts)")
	f.SetCellStyle(sheet, cell("A", row), cell("A", row), headerStyle)
	row++

	writeSubHeader(f, sheet, row, []string{"Symbol", "Name", "Mkt Value", "Cost Basis", "ST Gain", "ST Loss", "LT Gain", "LT Loss", "Net Unrealized", "Est."}, headerStyle)
	row++

	posFirstRow := row
	for _, ps := range positions {
		net := ps.stGain + ps.stLoss + ps.ltGain + ps.ltLoss
		f.SetCellValue(sheet, cell("A", row), ps.symbol)
		f.SetCellValue(sheet, cell("B", row), ps.name)
		setCurrency(f, sheet, "C", row, ps.mktVal, currencyFmt)
		setCurrency(f, sheet, "D", row, ps.cost, currencyFmt)
		setCurrency(f, sheet, "E", row, ps.stGain, currencyFmt)
		setCurrency(f, sheet, "F", row, ps.stLoss, currencyFmt)
		setCurrency(f, sheet, "G", row, ps.ltGain, currencyFmt)
		setCurrency(f, sheet, "H", row, ps.ltLoss, currencyFmt)
		setCurrency(f, sheet, "I", row, net, currencyFmt)
		if ps.estBasis {
			f.SetCellValue(sheet, cell("J", row), "~")
		}
		row++
	}
	posLastRow := row - 1

	f.SetCellValue(sheet, cell("A", row), "TOTAL")
	f.SetCellStyle(sheet, cell("A", row), cell("A", row), headerStyle)
	if posFirstRow <= posLastRow {
		for _, col := range []string{"C", "D", "E", "F", "G", "H", "I"} {
			setCurrencyFormula(f, sheet, col, row, fmt.Sprintf("SUM(%s%d:%s%d)", col, posFirstRow, col, posLastRow), currencyFmt)
		}
	}
	row++
	row++

	f.SetCellValue(sheet, cell("A", row), "All Lots (Taxable Accounts)")
	f.SetCellStyle(sheet, cell("A", row), cell("A", row), headerStyle)
	row++

	lotHeaders := []string{"Account", "Symbol", "Name", "Acquired", "Period", "Days", "Qty", "Cost Basis", "Mkt Value", "Unrealized G/L", "% G/L", "Action", "Est."}
	writeSubHeader(f, sheet, row, lotHeaders, headerStyle)
	row++

	sort.Slice(lotRows, func(i, j int) bool {
		gi := lotRows[i].mktVal - lotRows[i].cost
		gj := lotRows[j].mktVal - lotRows[j].cost
		return gi < gj
	})

	lotFirstRow := row
	for _, r := range lotRows {
		qty := float64(r.lot.RemainingMicros) / 1_000_000
		unrealized := r.mktVal - r.cost

		f.SetCellValue(sheet, cell("A", row), r.lot.AccountName)
		f.SetCellValue(sheet, cell("B", row), r.lot.Symbol)
		f.SetCellValue(sheet, cell("C", row), r.lot.SecurityName.String)
		f.SetCellValue(sheet, cell("D", row), r.lot.AcquiredDate)
		f.SetCellValue(sheet, cell("E", row), r.period)
		f.SetCellValue(sheet, cell("F", row), r.daysHeld)
		f.SetCellValue(sheet, cell("G", row), qty)
		setCurrency(f, sheet, "H", row, r.cost, currencyFmt)
		setCurrency(f, sheet, "I", row, r.mktVal, currencyFmt)
		setCurrencyFormula(f, sheet, "J", row, fmt.Sprintf("I%d-H%d", row, row), currencyFmt)
		if r.cost > 0 {
			setPercentFormula(f, sheet, "K", row, fmt.Sprintf("J%d/H%d", row, row), pctFmt)
		}
		if unrealized < 0 {
			f.SetCellValue(sheet, cell("L", row), "HARVEST")
		} else {
			f.SetCellValue(sheet, cell("L", row), "HOLD")
		}
		if r.estimated {
			f.SetCellValue(sheet, cell("M", row), "~")
		}
		row++
	}
	lotLastRow := row - 1

	f.SetCellValue(sheet, cell("A", row), "TOTAL")
	f.SetCellStyle(sheet, cell("A", row), cell("A", row), headerStyle)
	if lotFirstRow <= lotLastRow {
		setCurrencyFormula(f, sheet, "H", row, fmt.Sprintf("SUM(H%d:H%d)", lotFirstRow, lotLastRow), currencyFmt)
		setCurrencyFormula(f, sheet, "I", row, fmt.Sprintf("SUM(I%d:I%d)", lotFirstRow, lotLastRow), currencyFmt)
		setCurrencyFormula(f, sheet, "J", row, fmt.Sprintf("I%d-H%d", row, row), currencyFmt)
	}
}

func writePlaybookDetailSheet(f *excelize.File, data *reportData, currencyFmt, pctFmt, headerStyle int) {
	sheet := "Playbook Detail"
	f.NewSheet(sheet)

	bracketCeiling := bracketCeilings[defaultTargetBracket]
	portfolioIncome := computePortfolioOrdinaryIncome(data)
	totalOrdinary := portfolioIncome + defaultOtherIncome
	fedTaxableIncome := math.Max(totalOrdinary-fedStdDed2026, 0)
	conversionAmt := math.Max(bracketCeiling-fedTaxableIncome, 0)

	acctMap := make(map[string]db.Account)
	for _, a := range data.accounts {
		acctMap[a.ID] = a
	}

	row := 1

	// Table 1: Rebalancing Plan
	if data.analysis != nil && len(data.analysis.RebalanceDeltas) > 0 {
		f.SetCellValue(sheet, cell("A", row), "Rebalancing Plan")
		f.SetCellStyle(sheet, cell("A", row), cell("A", row), headerStyle)
		row++

		rebalHeaders := []string{"Symbol", "Name", "Sector", "Current Wt", "Target Wt", "Delta ($)", "Action", "Best Account", "Note"}
		writeSubHeader(f, sheet, row, rebalHeaders, headerStyle)
		row++

		for _, d := range data.analysis.RebalanceDeltas {
			qd := data.quotes[d.Symbol]
			name := qd.Name
			if name == "" {
				name = d.Name
			}
			f.SetCellValue(sheet, cell("A", row), d.Symbol)
			f.SetCellValue(sheet, cell("B", row), name)
			f.SetCellValue(sheet, cell("C", row), d.Sector)
			setPercent(f, sheet, "D", row, d.CurrentWeight, pctFmt)
			setPercent(f, sheet, "E", row, d.TargetWeight, pctFmt)
			setCurrency(f, sheet, "F", row, d.DeltaDollars, currencyFmt)
			f.SetCellValue(sheet, cell("G", row), d.Action)
			f.SetCellValue(sheet, cell("H", row), d.BestAccount)
			if d.Note != "" {
				f.SetCellValue(sheet, cell("I", row), d.Note)
			}
			row++
		}
		row++
	}

	// Consolidation Suggestions
	if data.analysis != nil && len(data.analysis.ConsolidationRecs) > 0 {
		f.SetCellValue(sheet, cell("A", row), "Consolidation Suggestions")
		f.SetCellStyle(sheet, cell("A", row), cell("A", row), headerStyle)
		row++

		writeSubHeader(f, sheet, row, []string{"Symbol", "Name", "Current Weight", "Merge Into", "Reason"}, headerStyle)
		row++

		for _, cr := range data.analysis.ConsolidationRecs {
			qd := data.quotes[cr.Symbol]
			name := qd.Name
			if name == "" {
				name = cr.Name
			}
			f.SetCellValue(sheet, cell("A", row), cr.Symbol)
			f.SetCellValue(sheet, cell("B", row), name)
			setPercent(f, sheet, "C", row, cr.CurrentWeight, pctFmt)
			f.SetCellValue(sheet, cell("D", row), cr.MergeInto)
			f.SetCellValue(sheet, cell("E", row), cr.Reason)
			row++
		}
		row++
	}

	// Table 2: Roth Conversion — Post-Rebalance IRA
	f.SetCellValue(sheet, cell("A", row), "Roth Conversion — Post-Rebalance IRA Positions")
	f.SetCellStyle(sheet, cell("A", row), cell("A", row), headerStyle)
	row++

	convHeaders := []string{"Account", "Symbol", "Name", "Growth Type", "Mkt Value", "% of Portfolio", "Location", "Action", "Convert Method", "Convert Value", "Running Total", "Status"}
	writeSubHeader(f, sheet, row, convHeaders, headerStyle)
	row++

	symbolWeights := computeSymbolWeights(data)
	rebalanceAdj := buildRebalanceAdjustments(data)

	var iraPositions []iraPosition
	for _, h := range data.allHoldings {
		acct := acctMap[h.AccountID]
		if acct.AccountType != "traditional_ira" && acct.AccountType != "sep_ira" {
			continue
		}

		qty := nullFloat64ToFloat(h.QuantityMicros) / 1_000_000
		cost := nullFloat64ToFloat(h.CostBasisMicros) / 1_000_000
		mv := marketValue(h.Symbol, h.SecurityType.String, qty, cost, data.quotes)

		if adj, ok := rebalanceAdj[h.Symbol]; ok {
			mv += adj
			if mv < 0 {
				mv = 0
			}
		}

		if mv <= 0 {
			continue
		}

		qd := data.quotes[h.Symbol]
		name := bestName(h.SecurityName.String, qd.Name)
		sector := resolveSector(qd, h.SecurityType.String, h.Symbol, name)

		gt := classifyGrowthType(h.Symbol, name, h.SecurityType.String)
		overweight := symbolWeights[h.Symbol] > positionConcentrationLimit
		action := determineConversionActionV2(gt, sector, overweight)
		priority := conversionPriorityV2(action)
		method := iraConvertMethod(h.Symbol, name, h.SecurityType.String)
		location := locationAssessment(gt, sector)

		iraPositions = append(iraPositions, iraPosition{
			account:    h.AccountName,
			symbol:     h.Symbol,
			name:       name,
			growthType: gt,
			mktValue:   mv,
			pctPortf:   mv / data.totalValue,
			method:     method,
			action:     action,
			priority:   priority,
			location:   location,
		})
	}

	for _, a := range data.accounts {
		if a.AccountType != "traditional_ira" && a.AccountType != "sep_ira" {
			continue
		}
		cash := float64(data.cashByAcct[a.ID]) / 1_000_000
		if cash <= 0 {
			continue
		}
		iraPositions = append(iraPositions, iraPosition{
			account:    a.Name,
			symbol:     "CASH",
			name:       "Cash Balance",
			growthType: "Cash",
			mktValue:   cash,
			pctPortf:   cash / data.totalValue,
			method:     "Sell First",
			action:     "KEEP IN IRA",
			priority:   6,
			location:   "Well-located",
		})
	}

	sort.Slice(iraPositions, func(i, j int) bool {
		if iraPositions[i].priority != iraPositions[j].priority {
			return iraPositions[i].priority < iraPositions[j].priority
		}
		return iraPositions[i].mktValue > iraPositions[j].mktValue
	})

	firstConvRow := row
	for i, p := range iraPositions {
		f.SetCellValue(sheet, cell("A", row), p.account)
		f.SetCellValue(sheet, cell("B", row), p.symbol)
		f.SetCellValue(sheet, cell("C", row), p.name)
		f.SetCellValue(sheet, cell("D", row), p.growthType)
		setCurrency(f, sheet, "E", row, p.mktValue, currencyFmt)
		setPercent(f, sheet, "F", row, p.pctPortf, pctFmt)
		f.SetCellValue(sheet, cell("G", row), p.location)
		f.SetCellValue(sheet, cell("H", row), p.action)
		f.SetCellValue(sheet, cell("I", row), p.method)

		if p.action == "KEEP IN IRA" {
			setCurrency(f, sheet, "J", row, 0, currencyFmt)
		} else if i == 0 || allPriorKeep(iraPositions[:i]) {
			setCurrency(f, sheet, "J", row, math.Min(p.mktValue, conversionAmt), currencyFmt)
		} else {
			prevRow := row - 1
			setCurrencyFormula(f, sheet, "J", row,
				fmt.Sprintf("MIN(E%d,MAX(%g-K%d,0))", row, conversionAmt, prevRow), currencyFmt)
		}

		setCurrencyFormula(f, sheet, "K", row, fmt.Sprintf("SUM(J$%d:J%d)", firstConvRow, row), currencyFmt)

		f.SetCellFormula(sheet, cell("L", row),
			fmt.Sprintf("IF(J%d=0,\"-\",IF(J%d<E%d,\"PARTIAL\",\"CONVERT\"))", row, row, row))

		row++
	}
	row++

	// Table 3: Tax Loss Harvest — Recommended Sales
	f.SetCellValue(sheet, cell("A", row), "Tax Loss Harvest — Recommended Sales")
	f.SetCellStyle(sheet, cell("A", row), cell("A", row), headerStyle)
	row++

	recHeaders := []string{"Account", "Symbol", "Name", "Period", "Qty", "Cost Basis", "Mkt Value", "Harvestable Loss", "Tax Savings", "Est. Basis"}
	writeSubHeader(f, sheet, row, recHeaders, headerStyle)
	row++

	fedLTCGRate := 0.15
	caCGRate := 0.093
	fedSTRate := 0.24

	lotRows := buildLotRows(data)
	recs := aggregateHarvestRecs(lotRows)

	sort.Slice(recs, func(i, j int) bool {
		return recs[i].loss < recs[j].loss
	})

	recFirstRow := row
	for _, rec := range recs {
		rate := fedLTCGRate + caCGRate
		if rec.period == "ST" {
			rate = fedSTRate + caCGRate
		}
		f.SetCellValue(sheet, cell("A", row), rec.account)
		f.SetCellValue(sheet, cell("B", row), rec.symbol)
		f.SetCellValue(sheet, cell("C", row), rec.name)
		f.SetCellValue(sheet, cell("D", row), rec.period)
		f.SetCellValue(sheet, cell("E", row), rec.qty)
		setCurrency(f, sheet, "F", row, rec.cost, currencyFmt)
		setCurrency(f, sheet, "G", row, rec.mktVal, currencyFmt)
		setCurrency(f, sheet, "H", row, rec.loss, currencyFmt)
		setCurrency(f, sheet, "I", row, -rec.loss*rate, currencyFmt)
		if rec.estBasis {
			f.SetCellValue(sheet, cell("J", row), "~")
		}
		row++
	}
	recLastRow := row - 1

	f.SetCellValue(sheet, cell("A", row), "TOTAL")
	f.SetCellStyle(sheet, cell("A", row), cell("A", row), headerStyle)
	if recFirstRow <= recLastRow {
		setCurrencyFormula(f, sheet, "F", row, fmt.Sprintf("SUM(F%d:F%d)", recFirstRow, recLastRow), currencyFmt)
		setCurrencyFormula(f, sheet, "G", row, fmt.Sprintf("SUM(G%d:G%d)", recFirstRow, recLastRow), currencyFmt)
		setCurrencyFormula(f, sheet, "H", row, fmt.Sprintf("SUM(H%d:H%d)", recFirstRow, recLastRow), currencyFmt)
		setCurrencyFormula(f, sheet, "I", row, fmt.Sprintf("SUM(I%d:I%d)", recFirstRow, recLastRow), currencyFmt)
	}
	row++
	row++

	// Table 4: All Unrealized Lots — Taxable
	f.SetCellValue(sheet, cell("A", row), "All Unrealized Lots — Taxable Accounts")
	f.SetCellStyle(sheet, cell("A", row), cell("A", row), headerStyle)
	row++

	lotHeaders := []string{"Account", "Symbol", "Name", "Acquired", "Period", "Days", "Qty", "Cost Basis", "Mkt Value", "Unrealized G/L", "% G/L", "Harvest?", "Est. Basis"}
	writeSubHeader(f, sheet, row, lotHeaders, headerStyle)
	row++

	sort.Slice(lotRows, func(i, j int) bool {
		gi := lotRows[i].mktVal - lotRows[i].cost
		gj := lotRows[j].mktVal - lotRows[j].cost
		return gi < gj
	})

	firstDataRow := row
	for _, r := range lotRows {
		qty := float64(r.lot.RemainingMicros) / 1_000_000
		unrealized := r.mktVal - r.cost

		f.SetCellValue(sheet, cell("A", row), r.lot.AccountName)
		f.SetCellValue(sheet, cell("B", row), r.lot.Symbol)
		f.SetCellValue(sheet, cell("C", row), r.lot.SecurityName.String)
		f.SetCellValue(sheet, cell("D", row), r.lot.AcquiredDate)
		f.SetCellValue(sheet, cell("E", row), r.period)
		f.SetCellValue(sheet, cell("F", row), r.daysHeld)
		f.SetCellValue(sheet, cell("G", row), qty)
		setCurrency(f, sheet, "H", row, r.cost, currencyFmt)
		setCurrency(f, sheet, "I", row, r.mktVal, currencyFmt)
		setCurrencyFormula(f, sheet, "J", row, fmt.Sprintf("I%d-H%d", row, row), currencyFmt)
		if r.cost > 0 {
			setPercentFormula(f, sheet, "K", row, fmt.Sprintf("J%d/H%d", row, row), pctFmt)
		}
		if unrealized < 0 {
			f.SetCellValue(sheet, cell("L", row), "SELL")
		}
		if r.estimated {
			f.SetCellValue(sheet, cell("M", row), "~")
		}
		row++
	}
	lastDataRow := row - 1

	row++
	f.SetCellValue(sheet, cell("A", row), "TOTAL")
	f.SetCellStyle(sheet, cell("A", row), cell("A", row), headerStyle)
	if firstDataRow <= lastDataRow {
		setCurrencyFormula(f, sheet, "H", row, fmt.Sprintf("SUM(H%d:H%d)", firstDataRow, lastDataRow), currencyFmt)
		setCurrencyFormula(f, sheet, "I", row, fmt.Sprintf("SUM(I%d:I%d)", firstDataRow, lastDataRow), currencyFmt)
		setCurrencyFormula(f, sheet, "J", row, fmt.Sprintf("I%d-H%d", row, row), currencyFmt)
	}
}

type iraPosition struct {
	account    string
	symbol     string
	name       string
	growthType string
	mktValue   float64
	pctPortf   float64
	method     string
	action     string
	priority   int
	location   string
}

type lotRow struct {
	lot       db.ListOpenLotsByAccountTypeRow
	cost      float64
	mktVal    float64
	daysHeld  int
	period    string
	estimated bool
}

type unrealizedTotals struct {
	gains  float64
	losses float64
}

type harvestRec struct {
	symbol   string
	name     string
	account  string
	period   string
	qty      float64
	cost     float64
	mktVal   float64
	loss     float64
	estBasis bool
}

func computePortfolioOrdinaryIncome(data *reportData) float64 {
	var income float64
	for _, p := range data.positions {
		qd := data.quotes[p.Symbol]
		if qd.DividendRate > 0 {
			qty := nullFloat64ToFloat(p.QuantityMicros) / 1_000_000
			income += qty * qd.DividendRate
		}
	}
	return income
}

func computeAccountTypeValues(data *reportData) (taxable, ira, roth float64) {
	acctMap := make(map[string]db.Account)
	for _, a := range data.accounts {
		acctMap[a.ID] = a
	}

	for _, h := range data.allHoldings {
		qty := nullFloat64ToFloat(h.QuantityMicros) / 1_000_000
		cost := nullFloat64ToFloat(h.CostBasisMicros) / 1_000_000
		mv := marketValue(h.Symbol, h.SecurityType.String, qty, cost, data.quotes)

		acct := acctMap[h.AccountID]
		switch acct.AccountType {
		case "traditional_ira", "sep_ira":
			ira += mv
		case "roth_ira":
			roth += mv
		default:
			taxable += mv
		}
	}

	for _, a := range data.accounts {
		cash := float64(data.cashByAcct[a.ID]) / 1_000_000
		switch a.AccountType {
		case "traditional_ira", "sep_ira":
			ira += cash
		case "roth_ira":
			roth += cash
		default:
			taxable += cash
		}
	}
	return
}

func computeUnrealizedByPeriod(data *reportData) (st, lt unrealizedTotals) {
	now := time.Now()
	for _, lot := range data.taxableLots {
		if lot.RemainingMicros <= 0 || lot.QuantityMicros <= 0 {
			continue
		}
		if !isLiquid(lot.SecurityType.String) {
			continue
		}
		qty := float64(lot.RemainingMicros) / 1_000_000
		cost := float64(lot.CostBasisMicros) / 1_000_000 * float64(lot.RemainingMicros) / float64(lot.QuantityMicros)
		mktVal := cost
		if qd, ok := data.quotes[lot.Symbol]; ok {
			mktVal = qty * qd.Price
		}

		acquired, _ := time.Parse("2006-01-02", lot.AcquiredDate)
		days := int(now.Sub(acquired).Hours() / 24)
		unrealized := mktVal - cost

		if days <= 365 {
			if unrealized < 0 {
				st.losses += unrealized
			} else {
				st.gains += unrealized
			}
		} else {
			if unrealized < 0 {
				lt.losses += unrealized
			} else {
				lt.gains += unrealized
			}
		}
	}
	return
}

func buildLotRows(data *reportData) []lotRow {
	now := time.Now()
	var rows []lotRow
	for _, lot := range data.taxableLots {
		if lot.RemainingMicros <= 0 || lot.QuantityMicros <= 0 {
			continue
		}
		if !isLiquid(lot.SecurityType.String) {
			continue
		}
		qty := float64(lot.RemainingMicros) / 1_000_000
		cost := float64(lot.CostBasisMicros) / 1_000_000 * float64(lot.RemainingMicros) / float64(lot.QuantityMicros)
		mktVal := cost
		if qd, ok := data.quotes[lot.Symbol]; ok {
			mktVal = qty * qd.Price
		}

		acquired, _ := time.Parse("2006-01-02", lot.AcquiredDate)
		daysHeld := int(now.Sub(acquired).Hours() / 24)
		period := "LT"
		if daysHeld <= 365 {
			period = "ST"
		}

		rows = append(rows, lotRow{
			lot:       lot,
			cost:      cost,
			mktVal:    mktVal,
			daysHeld:  daysHeld,
			period:    period,
			estimated: lot.EstimatedBasis != 0,
		})
	}
	return rows
}

func aggregateHarvestRecs(rows []lotRow) []*harvestRec {
	recMap := make(map[string]*harvestRec)
	for _, r := range rows {
		unrealized := r.mktVal - r.cost
		if unrealized >= 0 {
			continue
		}
		key := r.lot.Symbol + "|" + r.lot.AccountName + "|" + r.period
		if rec, ok := recMap[key]; ok {
			rec.qty += float64(r.lot.RemainingMicros) / 1_000_000
			rec.cost += r.cost
			rec.mktVal += r.mktVal
			rec.loss += unrealized
			if r.estimated {
				rec.estBasis = true
			}
		} else {
			recMap[key] = &harvestRec{
				symbol:   r.lot.Symbol,
				name:     r.lot.SecurityName.String,
				account:  r.lot.AccountName,
				period:   r.period,
				qty:      float64(r.lot.RemainingMicros) / 1_000_000,
				cost:     r.cost,
				mktVal:   r.mktVal,
				loss:     unrealized,
				estBasis: r.estimated,
			}
		}
	}

	var recs []*harvestRec
	for _, rec := range recMap {
		recs = append(recs, rec)
	}
	return recs
}

func classifyGrowthType(symbol, name, secType string) string {
	if isMoneyMarket(symbol, name) || symbol == "CASH" {
		return "Cash"
	}
	if secType == "bond" {
		return "Stable"
	}

	broadIndex := map[string]bool{
		"FXAIX": true, "VOO": true, "SPY": true, "VTI": true,
		"IVV": true, "VFIAX": true, "FSKAX": true, "VTSAX": true,
	}
	if broadIndex[symbol] {
		return "Index"
	}

	upper := strings.ToUpper(name)
	if strings.Contains(upper, "INDEX") || strings.Contains(upper, "S&P 500") || strings.Contains(upper, "TOTAL MARKET") {
		return "Index"
	}

	if secType == "equity" || secType == "etf" {
		return "High Growth"
	}

	if isMutualFundSymbol(symbol) {
		return "Growth Fund"
	}

	if secType == "" {
		return "High Growth"
	}

	return "Growth Fund"
}

func isMutualFundSymbol(symbol string) bool {
	return len(symbol) == 5 && symbol[4] == 'X'
}

func determineConversionAction(growthType string, overweight bool) string {
	switch growthType {
	case "High Growth", "Growth Fund", "Index":
		if overweight {
			return "TRIM & CONVERT"
		}
		return "CONVERT IN-KIND"
	default:
		return "KEEP IN IRA"
	}
}

func determineConversionActionV2(growthType, sector string, overweight bool) string {
	if sector == "Municipal Bonds" {
		return "SELL & CONVERT CASH"
	}
	switch growthType {
	case "High Growth":
		return "CONVERT IN-KIND"
	case "Growth Fund":
		return "CONVERT IN-KIND"
	case "Index":
		return "CONVERT IN-KIND"
	case "Stable":
		return "KEEP IN IRA"
	case "Cash":
		return "KEEP IN IRA"
	default:
		return "CONVERT IN-KIND"
	}
}

func conversionPriorityV2(action string) int {
	switch action {
	case "SELL & CONVERT CASH":
		return 1
	case "CONVERT IN-KIND":
		return 2
	case "KEEP IN IRA":
		return 6
	default:
		return 5
	}
}

func locationAssessment(growthType, sector string) string {
	if sector == "Municipal Bonds" {
		return "MISALLOCATED — muni in IRA"
	}
	switch growthType {
	case "High Growth", "Growth Fund":
		return "Convert to Roth"
	case "Index":
		return "Convert to Roth"
	case "Stable":
		return "Well-located"
	case "Cash":
		return "Well-located"
	}
	return "Evaluate"
}

func actionPriority(action string) int {
	switch action {
	case "TRIM & CONVERT":
		return 1
	case "CONVERT IN-KIND":
		return 2
	case "KEEP IN IRA":
		return 3
	default:
		return 4
	}
}

func iraConvertMethod(symbol, name, secType string) string {
	if isMoneyMarket(symbol, name) || symbol == "CASH" {
		return "Sell First"
	}
	if secType == "bond" {
		return "Sell First"
	}
	return "In-Kind"
}

func computeSymbolWeights(data *reportData) map[string]float64 {
	symbolValue := make(map[string]float64)
	for _, h := range data.allHoldings {
		qty := nullFloat64ToFloat(h.QuantityMicros) / 1_000_000
		cost := nullFloat64ToFloat(h.CostBasisMicros) / 1_000_000
		mv := marketValue(h.Symbol, h.SecurityType.String, qty, cost, data.quotes)
		symbolValue[h.Symbol] += mv
	}
	weights := make(map[string]float64, len(symbolValue))
	for sym, val := range symbolValue {
		weights[sym] = val / data.totalValue
	}
	return weights
}

func buildRebalanceAdjustments(data *reportData) map[string]float64 {
	adj := make(map[string]float64)
	if data.analysis == nil {
		return adj
	}
	for _, d := range data.analysis.RebalanceDeltas {
		adj[d.Symbol] += d.DeltaDollars
	}
	return adj
}

func allPriorKeep(positions []iraPosition) bool {
	for _, p := range positions {
		if p.action != "KEEP IN IRA" {
			return false
		}
	}
	return true
}

func findIRMAABracket(magi float64) (current irmaaThreshold, nextCliff float64) {
	current = irmaaThresholds[0]
	for i, t := range irmaaThresholds {
		if magi >= t.magi {
			current = t
			if i+1 < len(irmaaThresholds) {
				nextCliff = irmaaThresholds[i+1].magi
			}
		}
	}
	return
}

func formatAccountType(at string) string {
	switch at {
	case "traditional_ira":
		return "Traditional IRA"
	case "sep_ira":
		return "SEP IRA"
	case "roth_ira":
		return "Roth IRA"
	case "taxable":
		return "Taxable"
	}
	return at
}

func isMuniBond(secType, name string) bool {
	if secType != "bond" {
		return false
	}
	upper := strings.ToUpper(name)
	return strings.Contains(upper, "MUNI") || strings.Contains(upper, "MUNICIPAL")
}

func isLiquid(secType string) bool {
	switch secType {
	case "bond", "reit":
		return false
	}
	return true
}

type harvestingImpact struct {
	netSTAfter   float64
	netLTAfter   float64
	totalSavings float64
	taxAfter     float64
}

func computeHarvestingImpact(netST, netLT float64, unrealizedST, unrealizedLT unrealizedTotals, fedSTRate, fedLTCGRate, caCGRate float64) harvestingImpact {
	taxBefore := math.Max(netST, 0)*(fedSTRate+caCGRate) + math.Max(netLT, 0)*(fedLTCGRate+caCGRate)

	afterST := netST + unrealizedST.losses
	afterLT := netLT + unrealizedLT.losses

	// excess ST losses cross over to offset LT gains
	if afterST < 0 {
		afterLT += afterST
		afterST = 0
	}
	// excess LT losses cross over to offset ST gains
	if afterLT < 0 {
		afterST += afterLT
		afterLT = 0
	}

	afterST = math.Max(afterST, 0)
	afterLT = math.Max(afterLT, 0)

	taxAfter := afterST*(fedSTRate+caCGRate) + afterLT*(fedLTCGRate+caCGRate)

	return harvestingImpact{
		netSTAfter:   afterST,
		netLTAfter:   afterLT,
		totalSavings: taxBefore - taxAfter,
		taxAfter:     taxAfter,
	}
}

func computeStrategyIncome(data *reportData, instruments []string) float64 {
	instSet := make(map[string]bool, len(instruments))
	for _, s := range instruments {
		instSet[s] = true
	}
	var income float64
	for _, p := range data.positions {
		if !instSet[p.Symbol] {
			continue
		}
		qd := data.quotes[p.Symbol]
		if qd.DividendRate > 0 {
			qty := nullFloat64ToFloat(p.QuantityMicros) / 1_000_000
			income += qty * qd.DividendRate
		}
	}
	return income
}

func nextEstimatedTaxDate() string {
	now := time.Now()
	deadlines := []time.Time{
		time.Date(now.Year(), 4, 15, 0, 0, 0, 0, time.Local),
		time.Date(now.Year(), 6, 15, 0, 0, 0, 0, time.Local),
		time.Date(now.Year(), 9, 15, 0, 0, 0, 0, time.Local),
		time.Date(now.Year()+1, 1, 15, 0, 0, 0, 0, time.Local),
	}
	for _, d := range deadlines {
		if now.Before(d) {
			return d.Format("Jan 2, 2006")
		}
	}
	return deadlines[3].Format("Jan 2, 2006")
}

func nextEstimatedTaxDeadlineAfterConversion() string {
	now := time.Now()
	deadlines := []time.Time{
		time.Date(now.Year(), 4, 15, 0, 0, 0, 0, time.Local),
		time.Date(now.Year(), 6, 15, 0, 0, 0, 0, time.Local),
		time.Date(now.Year(), 9, 15, 0, 0, 0, 0, time.Local),
		time.Date(now.Year()+1, 1, 15, 0, 0, 0, 0, time.Local),
	}
	// Assume conversion happens ~30 days from now; find the deadline after that
	conversionDate := now.AddDate(0, 0, 30)
	for _, d := range deadlines {
		if conversionDate.Before(d) {
			return d.Format("Jan 2, 2006")
		}
	}
	return deadlines[3].Format("Jan 2, 2006")
}
