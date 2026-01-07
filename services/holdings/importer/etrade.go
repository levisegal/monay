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

type ETradeParser struct{}

func (p *ETradeParser) Parse(ctx context.Context, r io.Reader) (*ImportResult, error) {
	reader := csv.NewReader(r)
	reader.FieldsPerRecord = -1

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

		if len(record) >= 2 && record[0] == "For Account:" {
			externalAccountNumber = strings.TrimSpace(record[1])
			continue
		}

		if record[0] == "TransactionDate" {
			headerFound = true
			continue
		}

		if !headerFound {
			continue
		}

		if len(record) < 9 {
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
	dateStr := strings.TrimSpace(record[0])
	txnType := strings.TrimSpace(record[1])
	symbol := normalizeSymbol(strings.TrimSpace(record[3]))
	quantityStr := strings.TrimSpace(record[4])
	amountStr := strings.TrimSpace(record[5])
	priceStr := strings.TrimSpace(record[6])
	commissionStr := strings.TrimSpace(record[7])
	description := strings.TrimSpace(record[8])

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
	}, nil
}

func mapETradeTransactionType(txnType string, quantity, amount decimal.Decimal) TransactionType {
	switch txnType {
	case "Bought":
		return TransactionTypeBuy
	case "Sold":
		return TransactionTypeSell
	case "Dividend", "Qualified Dividend":
		return TransactionTypeDividend
	case "Interest Income", "Interest":
		return TransactionTypeInterest
	case "Online Transfer":
		if amount.IsPositive() {
			return TransactionTypeTransferIn
		}
		return TransactionTypeTransferOut
	case "Transfer":
		// Security transfer - shares coming in from another account
		// Treat as a buy for lot purposes (creates a cost basis lot)
		if quantity.IsPositive() || amount.IsPositive() {
			return TransactionTypeSecurityTransfer
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

// normalizeSymbol handles CUSIP-to-ticker mapping and cleanup
func normalizeSymbol(symbol string) string {
	// Skip empty or whitespace-only
	if strings.TrimSpace(symbol) == "" {
		return ""
	}

	// Known CUSIP mappings (E*TRADE sometimes uses CUSIPs for reorgs)
	cusipMap := map[string]string{
		"74374N102": "PRVB", // Provention Bio
	}
	if ticker, ok := cusipMap[symbol]; ok {
		return ticker
	}

	// Skip internal account references like #2145605
	if strings.HasPrefix(symbol, "#") {
		return ""
	}

	return symbol
}
