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

const microsMultiplier = 1_000_000

var etradeCashEquivalents = map[string]bool{
	"WMPXX": true, // Allspring Money Market Premier
	"VMFXX": true, // Vanguard Federal Money Market
}

type ETradeParser struct{}

func (p *ETradeParser) Parse(ctx context.Context, r io.Reader) (*ImportResult, error) {
	reader := csv.NewReader(r)
	reader.FieldsPerRecord = -1
	reader.LazyQuotes = true

	var transactions []Transaction
	var externalAccountNumber string
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

		if strings.HasPrefix(record[0], "Account Activity for ") {
			externalAccountNumber = extractAccountNumber(record[0])
			continue
		}

		if record[0] == "Activity/Trade Date" {
			headerFound = true
			continue
		}

		if !headerFound {
			continue
		}

		if len(record) < 11 {
			continue
		}

		txn, err := parseETradeRow(record)
		if err != nil {
			continue
		}
		if txn != nil {
			transactions = append(transactions, *txn)
		}
	}

	return &ImportResult{
		ExternalAccountNumber: externalAccountNumber,
		Transactions:          transactions,
		Positions:             nil,
	}, nil
}

func parseETradeRow(record []string) (*Transaction, error) {
	dateStr := strings.TrimSpace(record[1])
	txnType := strings.TrimSpace(record[3])
	description := strings.TrimSpace(record[4])
	symbol := normalizeSymbol(strings.TrimSpace(record[5]))
	quantityStr := strings.TrimSpace(record[7])
	priceStr := strings.TrimSpace(record[8])
	amountStr := strings.TrimSpace(record[9])
	commissionStr := strings.TrimSpace(record[10])

	if dateStr == "" {
		return nil, nil
	}

	date, err := time.Parse("01/02/06", dateStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse date %s: %w", dateStr, err)
	}

	quantity, _ := decimal.NewFromString(quantityStr)
	amount, _ := decimal.NewFromString(amountStr)
	price, _ := decimal.NewFromString(priceStr)
	commission, _ := decimal.NewFromString(commissionStr)

	transactionType := mapETradeTransactionType(txnType, quantity, amount)
	if transactionType == "" {
		return nil, nil
	}

	return &Transaction{
		Symbol:          symbol,
		SecurityName:    extractSecurityName(description),
		TransactionType: transactionType,
		TransactionDate: date,
		QuantityMicros:  toMicros(quantity.Abs()),
		PriceMicros:     toMicros(price),
		AmountMicros:    toMicros(amount.Abs()),
		FeesMicros:      toMicros(commission),
		Description:     description,
		CashEquivalent:  etradeCashEquivalents[symbol],
	}, nil
}

func mapETradeTransactionType(txnType string, quantity, amount decimal.Decimal) TransactionType {
	switch txnType {
	case "Bought":
		return TransactionTypeBuy
	case "Opening Balance":
		return TransactionTypeOpeningBalance
	case "Sold":
		return TransactionTypeSell
	case "Dividend":
		// DRIP: quantity > 0 means reinvesting dividend into shares (buy)
		if quantity.IsPositive() {
			return TransactionTypeBuy
		}
		return TransactionTypeDividend
	case "Qualified Dividend":
		return TransactionTypeDividend
	case "Interest Income", "Interest":
		return TransactionTypeInterest
	case "Online Transfer":
		if amount.IsPositive() {
			return TransactionTypeTransferIn
		}
		return TransactionTypeTransferOut
	case "Transfer":
		if quantity.IsPositive() {
			return TransactionTypeSecurityTransfer
		}
		if amount.IsPositive() {
			return TransactionTypeTransferIn
		}
		return TransactionTypeTransferOut
	case "Reorganization":
		if quantity.IsPositive() {
			return TransactionTypeReorgIn
		}
		return TransactionTypeReorgOut
	case "LT Cap Gain Distribution", "ST Cap Gain Distribution":
		return TransactionTypeCapGain
	case "Misc Trade", "Adjustment":
		return TransactionTypeOther
	default:
		return ""
	}
}

func extractAccountNumber(line string) string {
	// "Account Activity for maya -3758 from ..." → "#####3758"
	idx := strings.LastIndex(line, "-")
	if idx < 0 {
		return ""
	}
	rest := line[idx+1:]
	if spaceIdx := strings.Index(rest, " "); spaceIdx > 0 {
		rest = rest[:spaceIdx]
	}
	return "#####" + strings.TrimSpace(rest)
}

func extractSecurityName(description string) string {
	parts := strings.SplitN(description, " ", 4)
	if len(parts) >= 3 {
		return strings.Join(parts[:3], " ")
	}
	return description
}

func toMicros(d decimal.Decimal) int64 {
	return d.Mul(decimal.NewFromInt(microsMultiplier)).IntPart()
}

func normalizeSymbol(symbol string) string {
	if strings.TrimSpace(symbol) == "" || symbol == "--" {
		return ""
	}

	cusipMap := map[string]string{
		"74374N102": "PRVB",
	}
	if ticker, ok := cusipMap[symbol]; ok {
		return ticker
	}

	if strings.HasPrefix(symbol, "#") {
		return ""
	}

	return symbol
}
