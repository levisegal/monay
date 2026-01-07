// Package taxlots provides FIFO lot matching and gap analysis for tax lot tracking.
//
// # Era Detection
//
// A key concept is "ownership eras" - periods when you held a security separated by
// complete exits. The analyzer detects when unmatched sells are from a PREVIOUS era
// vs the CURRENT era to avoid false positives.
//
// Example: ZS (Zscaler)
//
//	Era 1 (old position):
//	  2015: Bought 1000 shares    <- NOT in transaction history (too old)
//	  2018: Sold 500 shares       <- In history, but no matching buy = UNMATCHED
//	  2019: Sold 500 shares       <- In history, but no matching buy = UNMATCHED
//	  (position fully closed)
//
//	Era 2 (new position):
//	  2023-11-13: Bought 300 shares  <- In history, creates lot
//	  (currently held)
//
// Without era detection: "ZS has 1000 unmatched sells AND 300 current shares = NEEDS REVIEW"
// With era detection: "Last unmatched sell (2019) is BEFORE current lot (2023) = SAFE TO IGNORE"
//
// The current 300 shares are NOT affected by the old unmatched sells because they're
// from a completely separate ownership period.
package taxlots

import (
	"context"
	"time"

	"github.com/levisegal/monay/services/holdings/gen/db"
)

type SymbolGap struct {
	Symbol                string
	SecurityID            string
	UnmatchedMicros       int64
	RemainingMicros       int64
	NeedsOpeningBalance   bool
	LastUnmatchedSellDate time.Time
	EarliestLotDate       time.Time
}

type AnalysisResult struct {
	Gaps []SymbolGap
}

type Analyzer struct {
	queries *db.Queries
}

func NewAnalyzer(queries *db.Queries) *Analyzer {
	return &Analyzer{queries: queries}
}

func (a *Analyzer) Analyze(ctx context.Context, accountID string) (*AnalysisResult, error) {
	txns, err := a.queries.ListTransactionsByAccount(ctx, accountID)
	if err != nil {
		return nil, err
	}

	sorted := sortByDateAsc(txns)

	lots := make(map[string][]simulatedLot)
	unmatched := make(map[string]int64)
	lastUnmatchedSell := make(map[string]time.Time)

	for _, txn := range sorted {
		if !txn.SecurityID.Valid {
			continue
		}

		secID := txn.SecurityID.String
		txnDate := txn.TransactionDate.Time

		switch txn.TransactionType {
		case "buy", "security_transfer", "opening_balance":
			if txn.QuantityMicros.Valid && txn.QuantityMicros.Int64 > 0 {
				lots[secID] = append(lots[secID], simulatedLot{
					quantityMicros:  txn.QuantityMicros.Int64,
					remainingMicros: txn.QuantityMicros.Int64,
					acquiredDate:    txnDate,
				})
			}
		case "sell":
			if !txn.QuantityMicros.Valid || txn.QuantityMicros.Int64 == 0 {
				continue
			}
			remaining := txn.QuantityMicros.Int64
			for i := range lots[secID] {
				if remaining <= 0 {
					break
				}
				if lots[secID][i].remainingMicros <= 0 {
					continue
				}
				take := min(remaining, lots[secID][i].remainingMicros)
				lots[secID][i].remainingMicros -= take
				remaining -= take
			}
			if remaining > 0 {
				unmatched[secID] += remaining
				lastUnmatchedSell[secID] = txnDate
			}
		}
	}

	remainingBySymbol, err := a.queries.SumRemainingBySymbol(ctx, accountID)
	if err != nil {
		return nil, err
	}

	symbolMap := make(map[string]db.SumRemainingBySymbolRow)
	for _, row := range remainingBySymbol {
		symbolMap[row.SecurityID] = row
	}

	var gaps []SymbolGap
	for secID, unmatchedQty := range unmatched {
		row, ok := symbolMap[secID]
		symbol := secID
		var remainingMicros int64
		if ok {
			symbol = row.Symbol
			remainingMicros = row.RemainingMicros
		}

		var earliestLotDate time.Time
		for _, lot := range lots[secID] {
			if lot.remainingMicros > 0 {
				if earliestLotDate.IsZero() || lot.acquiredDate.Before(earliestLotDate) {
					earliestLotDate = lot.acquiredDate
				}
			}
		}

		lastSellDate := lastUnmatchedSell[secID]

		// Era-aware opening balance detection:
		//
		// We only need an opening balance if unmatched sells could affect CURRENT holdings.
		// This happens when sells occurred DURING the current ownership era (after you
		// acquired the shares you still hold).
		//
		// Example 1: NEEDS opening balance (same era)
		//   2015: Bought 1000 shares    <- missing from history
		//   2023: Sold 500 shares       <- unmatched, AFTER earliest lot would be
		//   Current: 500 shares held    <- these ARE affected by missing 2015 buy
		//   => lastSellDate (2023) >= earliestLotDate (would be 2015 if we had it)
		//   => NeedsOpeningBalance = true
		//
		// Example 2: SAFE to ignore (different eras)
		//   2015: Bought 1000 shares    <- missing from history (Era 1)
		//   2019: Sold 1000 shares      <- unmatched sell (Era 1 closed)
		//   2023: Bought 300 shares     <- in history, new era (Era 2)
		//   Current: 300 shares held    <- NOT affected by old missing buy
		//   => lastSellDate (2019) < earliestLotDate (2023)
		//   => NeedsOpeningBalance = false
		//
		needsOpeningBalance := remainingMicros > 0 && !earliestLotDate.IsZero() &&
			!lastSellDate.Before(earliestLotDate)

		gaps = append(gaps, SymbolGap{
			Symbol:                symbol,
			SecurityID:            secID,
			UnmatchedMicros:       unmatchedQty,
			RemainingMicros:       remainingMicros,
			NeedsOpeningBalance:   needsOpeningBalance,
			LastUnmatchedSellDate: lastSellDate,
			EarliestLotDate:       earliestLotDate,
		})
	}

	return &AnalysisResult{Gaps: gaps}, nil
}

type simulatedLot struct {
	quantityMicros  int64
	remainingMicros int64
	acquiredDate    time.Time
}
