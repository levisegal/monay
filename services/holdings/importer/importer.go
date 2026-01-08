package importer

import (
	"context"
	"fmt"
	"io"
	"time"
)

type TransactionType string

const (
	TransactionTypeBuy              TransactionType = "buy"
	TransactionTypeSell             TransactionType = "sell"
	TransactionTypeDividend         TransactionType = "dividend"
	TransactionTypeInterest         TransactionType = "interest"
	TransactionTypeSplit            TransactionType = "split"
	TransactionTypeTransferIn       TransactionType = "transfer_in"
	TransactionTypeTransferOut      TransactionType = "transfer_out"
	TransactionTypeSecurityTransfer TransactionType = "security_transfer" // shares transferred in
	TransactionTypeReorgIn          TransactionType = "reorg_in"          // shares received from reorg
	TransactionTypeReorgOut         TransactionType = "reorg_out"         // shares removed from reorg (cash merger, etc)
	TransactionTypeCapGain          TransactionType = "cap_gain"          // capital gain distributions
	TransactionTypeOpeningBalance   TransactionType = "opening_balance"   // manual opening lot
	TransactionTypeFee              TransactionType = "fee"               // advisory fees, etc
	TransactionTypeOther            TransactionType = "other"
)

type Transaction struct {
	Symbol          string
	SecurityName    string
	TransactionType TransactionType
	TransactionDate time.Time
	QuantityMicros  int64 // quantity * 1,000,000
	PriceMicros     int64 // price * 1,000,000
	AmountMicros    int64 // amount * 1,000,000
	FeesMicros      int64 // fees * 1,000,000
	Description     string
}

type Position struct {
	Symbol            string
	SecurityName      string
	QuantityMicros    int64 // quantity * 1,000,000
	CostBasisMicros   int64 // cost basis * 1,000,000
	MarketValueMicros int64 // market value * 1,000,000
	AsOfDate          time.Time
}

type ImportResult struct {
	ExternalAccountNumber string
	Transactions          []Transaction
	Positions             []Position
}

type Broker string

const (
	BrokerETrade   Broker = "etrade"
	BrokerSchwab   Broker = "schwab"
	BrokerFidelity Broker = "fidelity"
	BrokerVanguard Broker = "vanguard"
	BrokerLPL      Broker = "lpl"
	BrokerMerrill  Broker = "merrill"
)

type Parser interface {
	Parse(ctx context.Context, r io.Reader) (*ImportResult, error)
}

func GetParser(broker Broker) (Parser, error) {
	switch broker {
	case BrokerETrade:
		return &ETradeParser{}, nil
	case BrokerSchwab:
		return &SchwabParser{}, nil
	case BrokerFidelity:
		return &FidelityParser{}, nil
	case BrokerVanguard:
		return &VanguardParser{}, nil
	case BrokerLPL:
		return &LPLParser{}, nil
	case BrokerMerrill:
		return &MerrillParser{}, nil
	default:
		return nil, fmt.Errorf("unsupported broker: %s", broker)
	}
}
