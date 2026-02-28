package taxlots

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type PositionsFile struct {
	Account   string            `json:"account"`
	Broker    string            `json:"broker"`
	Date      string            `json:"date"`
	Summary   PositionsSummary  `json:"summary"`
	Positions []ScrapedPosition `json:"positions"`
}

type PositionsSummary struct {
	NetValue            float64 `json:"net_value"`
	UnrealizedGain      float64 `json:"unrealized_gain"`
	UnrealizedGainPct   float64 `json:"unrealized_gain_pct"`
	DaysGain            float64 `json:"days_gain"`
	CashPurchasingPower float64 `json:"cash_purchasing_power"`
}

type ScrapedPosition struct {
	Symbol    string  `json:"symbol"`
	Quantity  float64 `json:"quantity"`
	PricePaid float64 `json:"price_paid"`
}

func LoadPositions(path string) (*PositionsFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read positions file: %w", err)
	}

	var pf PositionsFile
	if err := json.Unmarshal(data, &pf); err != nil {
		return nil, fmt.Errorf("failed to parse positions JSON: %w", err)
	}

	return &pf, nil
}

func (pf *PositionsFile) FindBySymbol(symbol string) *ScrapedPosition {
	for i := range pf.Positions {
		if strings.EqualFold(pf.Positions[i].Symbol, symbol) {
			return &pf.Positions[i]
		}
	}
	return nil
}
