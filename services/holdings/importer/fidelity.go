package importer

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

var fidelityCashEquivalents = map[string]bool{
	"FDRXX": true, // Fidelity Government Cash Reserves
	"SPAXX": true, // Fidelity Government Money Market
	"FCASH": true, // Fidelity cash
}

type FidelityParser struct{}

func (p *FidelityParser) Parse(ctx context.Context, r io.Reader) (*ImportResult, error) {
	reader := csv.NewReader(r)
	reader.FieldsPerRecord = -1
	reader.LazyQuotes = true

	var transactions []Transaction
	headerFound := false

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read CSV: %w", err)
		}

		if len(record) == 0 {
			continue
		}

		first := strings.TrimSpace(record[0])

		if first == "Run Date" {
			headerFound = true
			continue
		}

		if !headerFound {
			continue
		}

		if len(record) < 11 {
			continue
		}

		txn, err := parseFidelityRow(record)
		if err != nil {
			continue
		}
		if txn != nil {
			transactions = append(transactions, *txn)
		}
	}

	return &ImportResult{
		Transactions: transactions,
	}, nil
}

func parseFidelityRow(record []string) (*Transaction, error) {
	dateStr := strings.TrimSpace(record[0])
	action := strings.TrimSpace(record[1])
	symbol := strings.TrimSpace(record[2])
	description := strings.TrimSpace(record[3])
	priceStr := cleanFidelityAmount(record[5])
	quantityStr := cleanFidelityAmount(record[6])
	commissionStr := cleanFidelityAmount(record[7])
	feesStr := cleanFidelityAmount(record[8])
	amountStr := cleanFidelityAmount(record[10])

	if dateStr == "" {
		return nil, nil
	}

	date, err := time.Parse("01/02/2006", dateStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse date %s: %w", dateStr, err)
	}

	quantity, _ := decimal.NewFromString(quantityStr)
	amount, _ := decimal.NewFromString(amountStr)
	price, _ := decimal.NewFromString(priceStr)
	commission, _ := decimal.NewFromString(commissionStr)
	fees, _ := decimal.NewFromString(feesStr)
	totalFees := commission.Add(fees)

	symbol = normalizeFidelitySymbol(symbol)

	txnType := mapFidelityTransactionType(action, quantity, amount)
	if txnType == "" {
		return nil, nil
	}

	return &Transaction{
		Symbol:          symbol,
		SecurityName:    description,
		TransactionType: txnType,
		TransactionDate: date,
		QuantityMicros:  toMicros(quantity.Abs()),
		PriceMicros:     toMicros(price),
		AmountMicros:    toMicros(amount.Abs()),
		FeesMicros:      toMicros(totalFees),
		Description:     action,
		CashEquivalent:  fidelityCashEquivalents[symbol],
	}, nil
}

func mapFidelityTransactionType(action string, quantity, amount decimal.Decimal) TransactionType {
	upper := strings.ToUpper(action)

	switch {
	case strings.HasPrefix(upper, "YOU BOUGHT"):
		return TransactionTypeBuy

	case strings.HasPrefix(upper, "YOU SOLD"):
		return TransactionTypeSell

	case strings.HasPrefix(upper, "DIVIDEND RECEIVED"):
		return TransactionTypeDividend

	case strings.HasPrefix(upper, "INTEREST EARNED"):
		return TransactionTypeInterest

	case strings.HasPrefix(upper, "FEE CHARGED"):
		return TransactionTypeFee

	case strings.HasPrefix(upper, "REINVESTMENT"):
		// Money market DRIP — skip (cash symbols already filtered above,
		// but if a non-cash symbol somehow gets here, treat as buy)
		return TransactionTypeBuy

	case strings.HasPrefix(upper, "MERGER MER FROM"):
		return TransactionTypeReorgIn

	case strings.HasPrefix(upper, "MERGER MER PAYOUT"):
		// Old shares removed OR cash payout from merger
		if quantity.IsNegative() {
			return TransactionTypeReorgOut
		}
		// Cash payout (no shares) — treat as reorg cash proceeds
		return TransactionTypeReorgOut

	case strings.HasPrefix(upper, "REVERSE SPLIT"):
		return TransactionTypeReorgOut

	case strings.HasPrefix(upper, "IN LIEU OF FRX SHARE"):
		// Cash in lieu of fractional shares from reorg
		return TransactionTypeReorgIn

	case strings.HasPrefix(upper, "TRANSFERRED TO"):
		return TransactionTypeTransferOut

	case strings.HasPrefix(upper, "CASH CONTRIBUTION"):
		return TransactionTypeTransferIn

	default:
		return ""
	}
}

func normalizeFidelitySymbol(symbol string) string {
	symbol = strings.TrimSpace(symbol)
	if symbol == "" || symbol == "--" {
		return ""
	}

	cusipMap := map[string]string{
		"86800U104": "SMCI",
		"38000Q102": "GLYCO",
		"87164U201": "SYN",
		"87164U409": "TOVX",
		"04624N107": "AST",
		"315994103": "FCASH",
	}
	if ticker, ok := cusipMap[symbol]; ok {
		return ticker
	}

	return symbol
}

func cleanFidelityAmount(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, "$", "")
	s = strings.ReplaceAll(s, ",", "")
	if s == "" || s == "--" {
		return "0"
	}
	return s
}
