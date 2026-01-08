package importer

import (
	"bufio"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

type MerrillParser struct{}

func (p *MerrillParser) Parse(ctx context.Context, r io.Reader) (*ImportResult, error) {
	// Merrill CSVs have weird formatting: `"value" ,"value"` (space before comma)
	// Preprocess to normalize: `"value","value"`
	normalized := normalizeMerrillCSV(r)

	reader := csv.NewReader(normalized)
	reader.FieldsPerRecord = -1
	reader.LazyQuotes = true
	reader.TrimLeadingSpace = true

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

		// Look for header row
		firstCol := strings.TrimSpace(record[0])
		if strings.Contains(strings.ToLower(firstCol), "trade date") {
			headerFound = true
			continue
		}

		if !headerFound {
			continue
		}

		// Skip empty rows or separator rows
		if len(record) < 9 || firstCol == "" || firstCol == "," {
			continue
		}

		// Extract account number from Account column (index 2)
		if externalAccountNumber == "" && len(record) > 2 {
			acct := strings.TrimSpace(record[2])
			if acct != "" {
				externalAccountNumber = acct
			}
		}

		txn, err := parseMerrillRow(record)
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

// normalizeMerrillCSV fixes Merrill's weird CSV formatting where there's a space
// before commas: `"value" ,"value"` → `"value","value"`
func normalizeMerrillCSV(r io.Reader) io.Reader {
	var result strings.Builder
	scanner := bufio.NewScanner(r)
	// Increase buffer for long lines (Merrill descriptions can be very long)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	for scanner.Scan() {
		line := scanner.Text()
		// Replace `" ,"` with `","`
		normalized := strings.ReplaceAll(line, "\" ,\"", "\",\"")
		// Remove trailing space after last quote (Merrill quirk)
		normalized = strings.TrimRight(normalized, " ")
		result.WriteString(normalized)
		result.WriteByte('\n')
	}

	return strings.NewReader(result.String())
}

func parseMerrillRow(record []string) (*Transaction, error) {
	// Columns: Trade Date, Settlement Date, Account, Description, Type, Symbol/CUSIP, Quantity, Price, Amount
	dateStr := strings.TrimSpace(record[0])
	description := strings.TrimSpace(record[3])
	symbol := strings.TrimSpace(record[5])
	quantityStr := cleanMerrillAmount(record[6])
	priceStr := cleanMerrillAmount(record[7])
	amountStr := cleanMerrillAmount(record[8])

	if dateStr == "" {
		return nil, nil
	}

	// Parse date MM/DD/YYYY
	date, err := time.Parse("01/02/2006", dateStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse date %s: %w", dateStr, err)
	}

	// Parse amounts first (needed for transaction type detection)
	quantity, _ := decimal.NewFromString(quantityStr)
	price, _ := decimal.NewFromString(priceStr)
	amount, _ := decimal.NewFromString(amountStr)

	// Determine transaction type from description (and qty/amount for stock splits)
	transactionType := mapMerrillTransactionType(description, quantity, amount)
	if transactionType == "" {
		return nil, nil
	}

	// Normalize CUSIP to symbol for known securities
	symbol = normalizeMerrillSymbol(symbol)

	// Skip internal cash account transactions
	if symbol == "990286916" || strings.Contains(symbol, "TMCXX") {
		if transactionType != TransactionTypeInterest {
			return nil, nil
		}
	}

	// For sells, quantity is negative in CSV - we want absolute value
	// For buys/opening balance, use the quantity as-is (positive)
	absQuantity := quantity.Abs()

	// For amount, take absolute value
	absAmount := amount.Abs()

	return &Transaction{
		Symbol:          symbol,
		SecurityName:    extractMerrillSecurityName(description),
		TransactionType: transactionType,
		TransactionDate: date,
		QuantityMicros:  toMicros(absQuantity),
		PriceMicros:     toMicros(price),
		AmountMicros:    toMicros(absAmount),
		FeesMicros:      0,
		Description:     description,
	}, nil
}

func mapMerrillTransactionType(description string, quantity, amount decimal.Decimal) TransactionType {
	desc := strings.ToLower(description)

	switch {
	case strings.HasPrefix(desc, "sale "):
		return TransactionTypeSell
	case strings.HasPrefix(desc, "purchase "):
		return TransactionTypeBuy
	case strings.HasPrefix(desc, "opening balance"):
		return TransactionTypeOpeningBalance

	// Stock split handling: "Dividend X HOLDING Y PAY DATE" with qty > 0 and amount = 0
	case strings.HasPrefix(desc, "dividend ") && !quantity.IsZero() && amount.IsZero():
		// Stock split shares - treat as buy with $0 cost
		return TransactionTypeBuy

	case strings.HasPrefix(desc, "dividend "), strings.HasPrefix(desc, "foreign dividend "):
		return TransactionTypeDividend

	case strings.HasPrefix(desc, "interest "):
		return TransactionTypeInterest
	case strings.HasPrefix(desc, "bank interest "):
		return TransactionTypeInterest

	case strings.HasPrefix(desc, "short term capital gain"):
		return TransactionTypeCapGain
	case strings.HasPrefix(desc, "long term capital gain"):
		return TransactionTypeCapGain

	case strings.HasPrefix(desc, "advisory program fee"):
		return TransactionTypeFee

	case strings.HasPrefix(desc, "reinvestment share"):
		// DRIP shares - treat as buy with $0 cost
		return TransactionTypeBuy

	// Bond redemption at maturity
	case strings.HasPrefix(desc, "redemption "):
		return TransactionTypeSell

	// Exchange transactions (muni bond exchanges) - buy or sell based on qty sign
	case strings.HasPrefix(desc, "exchange "):
		if quantity.IsNegative() {
			return TransactionTypeSell
		}
		return TransactionTypeBuy

	// Skip these:
	case strings.HasPrefix(desc, "stock dividend due bill"):
		// Temporary placeholder entries that cancel out - skip
		return ""
	case strings.HasPrefix(desc, "reinvestment program"):
		// Cash side of DRIP - skip (the "Reinvestment Share(s)" has the actual shares)
		return ""
	case strings.HasPrefix(desc, "deposit ml bank"), strings.HasPrefix(desc, "withdrawal ml"):
		// Internal cash sweep - skip
		return ""
	case strings.HasPrefix(desc, "check accumulation"):
		// Cash sweep accumulation - skip
		return ""
	case strings.HasPrefix(desc, "foreign tax"):
		// Foreign withholding tax - skip for now
		return ""

	default:
		return ""
	}
}

// normalizeMerrillSymbol maps CUSIPs to symbols for securities that Merrill
// reports inconsistently (e.g., using CUSIP at maturity but symbol elsewhere)
func normalizeMerrillSymbol(symbol string) string {
	cusipToSymbol := map[string]string{
		"46435U432": "IBMN", // iShares iBonds Dec 2025 Term Muni
	}
	if mapped, ok := cusipToSymbol[symbol]; ok {
		return mapped
	}
	return symbol
}

func cleanMerrillAmount(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, "$", "")
	s = strings.ReplaceAll(s, ",", "")
	s = strings.ReplaceAll(s, "\"", "")
	if s == "" || s == "-" {
		return "0"
	}
	return s
}

func extractMerrillSecurityName(description string) string {
	desc := strings.TrimSpace(description)

	// Remove common prefixes
	prefixes := []string{
		"Sale  ", "Sale ", "Purchase  ", "Purchase ",
		"Opening Balance - ",
		"Dividend ", "Foreign Dividend ", "Bank Interest ", "Interest ",
		"Short Term Capital Gain ", "Long Term Capital Gain ",
		"Advisory Program Fee ",
		"Reinvestment Share(s) ", "Reinvestment Program ",
		"Redemption ", "Exchange ",
	}

	for _, prefix := range prefixes {
		if strings.HasPrefix(desc, prefix) {
			desc = strings.TrimPrefix(desc, prefix)
			break
		}
	}

	// Cut at common noise phrases
	cutoffs := []string{
		" VSP ", " EXECUTED ", " PER ADVISORY", " THIS SALE",
		" HOLDING ", " PAY DATE", " AGENT REINV",
		" PROSPECTUS", " PRODUCT DESCRIPTION",
	}

	for _, cutoff := range cutoffs {
		if idx := strings.Index(desc, cutoff); idx > 0 {
			desc = desc[:idx]
			break
		}
	}

	return strings.TrimSpace(desc)
}
