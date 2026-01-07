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

type LPLParser struct{}

func (p *LPLParser) Parse(ctx context.Context, r io.Reader) (*ImportResult, error) {
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

		// Header: Date,Activity,Symbol,Description,Quantity,Unit Price,Value,Held In,Account Nickname,Account Number
		if record[0] == "Date" {
			headerFound = true
			continue
		}

		if !headerFound {
			continue
		}

		if len(record) < 10 {
			continue
		}

		// Extract account number from last column
		if externalAccountNumber == "" {
			externalAccountNumber = strings.TrimSpace(record[9])
		}

		txn, err := parseLPLRow(record)
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

func parseLPLRow(record []string) (*Transaction, error) {
	dateStr := strings.TrimSpace(record[0])
	activity := strings.ToLower(strings.TrimSpace(record[1]))
	symbol := strings.TrimSpace(record[2])
	description := strings.TrimSpace(record[3])
	quantityStr := strings.TrimSpace(record[4])
	priceStr := cleanLPLAmount(record[5])
	valueStr := cleanLPLAmount(record[6])

	if dateStr == "" {
		return nil, nil
	}

	// Parse date MM/DD/YYYY
	date, err := time.Parse("1/2/2006", dateStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse date %s: %w", dateStr, err)
	}

	// Clean quantity
	if quantityStr == "-" {
		quantityStr = "0"
	}
	quantity, _ := decimal.NewFromString(quantityStr)
	price, _ := decimal.NewFromString(priceStr)
	value, _ := decimal.NewFromString(valueStr)

	transactionType := mapLPLTransactionType(activity, value)
	if transactionType == "" {
		return nil, nil
	}

	// Skip internal cash account transactions for buy/sell
	if (transactionType == TransactionTypeBuy || transactionType == TransactionTypeSell) && symbol == "9999227" {
		return nil, nil
	}

	// Clean up symbol
	symbol = normalizeLPLSymbol(symbol)

	return &Transaction{
		Symbol:          symbol,
		SecurityName:    extractLPLSecurityName(description),
		TransactionType: transactionType,
		TransactionDate: date,
		QuantityMicros:  toMicros(quantity.Abs()),
		PriceMicros:     toMicros(price),
		AmountMicros:    toMicros(value.Abs()),
		FeesMicros:      0,
		Description:     description,
	}, nil
}

func mapLPLTransactionType(activity string, value decimal.Decimal) TransactionType {
	switch activity {
	case "buy":
		return TransactionTypeBuy
	case "sell":
		return TransactionTypeSell
	case "interest":
		return TransactionTypeInterest
	case "reinvest interest":
		// Interest reinvested into cash account - treat as interest
		return TransactionTypeInterest
	case "ica transfer":
		if value.IsPositive() {
			return TransactionTypeTransferIn
		}
		return TransactionTypeTransferOut
	case "journal":
		// Distribution to linked account
		return TransactionTypeTransferOut
	case "fee":
		return TransactionTypeFee
	default:
		return ""
	}
}

func cleanLPLAmount(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, "$", "")
	s = strings.ReplaceAll(s, "\t", "")
	s = strings.ReplaceAll(s, ",", "")
	if s == "-" {
		return "0"
	}
	return s
}

func normalizeLPLSymbol(symbol string) string {
	// Skip internal cash account
	if symbol == "9999227" || symbol == "CASH" {
		return ""
	}
	return symbol
}

func extractLPLSecurityName(description string) string {
	// Bond descriptions are very long, extract the issuer name
	// e.g. "LOS ANGELES CA UNI SCH DIST RFDG SER A B/E CPN  5.000% DUE..."
	parts := strings.Split(description, " CPN ")
	if len(parts) > 0 {
		return strings.TrimSpace(parts[0])
	}
	return description
}

