package taxlots

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/levisegal/monay/services/holdings/database"
	"github.com/levisegal/monay/services/holdings/gen/db"
)

type Processor struct {
	queries *db.Queries
}

func NewProcessor(queries *db.Queries) *Processor {
	return &Processor{queries: queries}
}

func (p *Processor) ProcessTransactions(ctx context.Context, accountID string) error {
	if err := p.queries.DeleteLotsByAccount(ctx, accountID); err != nil {
		return fmt.Errorf("failed to clear lot dispositions: %w", err)
	}
	if err := p.queries.DeleteLotsForAccount(ctx, accountID); err != nil {
		return fmt.Errorf("failed to clear lots: %w", err)
	}

	slog.Info("cleared existing lots for account", "account_id", accountID)

	txns, err := p.queries.ListTransactionsByAccount(ctx, accountID)
	if err != nil {
		return fmt.Errorf("failed to list transactions: %w", err)
	}

	sorted := sortByDateAsc(txns)

	for _, txn := range sorted {
		if !txn.SecurityID.Valid {
			continue
		}

		switch txn.TransactionType {
		case "buy", "security_transfer":
			// Both buys and security transfers create lots
			if err := p.processBuy(ctx, txn); err != nil {
				return fmt.Errorf("failed to process buy %s: %w", txn.ID, err)
			}
		case "sell":
			if err := p.processSell(ctx, txn); err != nil {
				return fmt.Errorf("failed to process sell %s: %w", txn.ID, err)
			}
		}
	}

	return nil
}

func (p *Processor) processBuy(ctx context.Context, txn db.ListTransactionsByAccountRow) error {
	if !txn.QuantityMicros.Valid || txn.QuantityMicros.Int64 == 0 {
		return nil
	}

	_, err := p.queries.CreateLot(ctx, db.CreateLotParams{
		ID:              database.NewID(database.PrefixLot),
		AccountID:       txn.AccountID,
		SecurityID:      txn.SecurityID.String,
		TransactionID:   txn.ID,
		AcquiredDate:    txn.TransactionDate,
		QuantityMicros:  txn.QuantityMicros.Int64,
		RemainingMicros: txn.QuantityMicros.Int64,
		CostBasisMicros: txn.AmountMicros,
	})
	if err != nil {
		return err
	}

	slog.Debug("created lot",
		"transaction_id", txn.ID,
		"symbol", txn.Symbol,
		"quantity", txn.QuantityMicros.Int64,
	)

	return nil
}

func (p *Processor) processSell(ctx context.Context, txn db.ListTransactionsByAccountRow) error {
	if !txn.QuantityMicros.Valid || txn.QuantityMicros.Int64 == 0 {
		return nil
	}

	lots, err := p.queries.ListLotsByAccountAndSecurity(ctx, db.ListLotsByAccountAndSecurityParams{
		AccountID:  txn.AccountID,
		SecurityID: txn.SecurityID.String,
	})
	if err != nil {
		return fmt.Errorf("failed to list lots: %w", err)
	}

	remainingToSell := txn.QuantityMicros.Int64
	proceeds := txn.AmountMicros
	sellDate := txn.TransactionDate.Time

	for _, lot := range lots {
		if remainingToSell <= 0 {
			break
		}

		if lot.RemainingMicros <= 0 {
			continue
		}

		sellFromLot := min(remainingToSell, lot.RemainingMicros)

		costPerMicro := float64(lot.CostBasisMicros) / float64(lot.QuantityMicros)
		costBasis := int64(costPerMicro * float64(sellFromLot))

		proceedsPerMicro := float64(proceeds) / float64(txn.QuantityMicros.Int64)
		lotProceeds := int64(proceedsPerMicro * float64(sellFromLot))

		gain := lotProceeds - costBasis

		holdingPeriod := "short_term"
		acquiredDate := lot.AcquiredDate.Time
		if sellDate.Sub(acquiredDate) > 365*24*time.Hour {
			holdingPeriod = "long_term"
		}

		_, err := p.queries.CreateLotDisposition(ctx, db.CreateLotDispositionParams{
			ID:                 database.NewID(database.PrefixLotDisposition),
			LotID:              lot.ID,
			SellTransactionID:  txn.ID,
			DisposedDate:       txn.TransactionDate,
			QuantityMicros:     sellFromLot,
			CostBasisMicros:    costBasis,
			ProceedsMicros:     lotProceeds,
			RealizedGainMicros: gain,
			HoldingPeriod:      holdingPeriod,
		})
		if err != nil {
			return fmt.Errorf("failed to create disposition: %w", err)
		}

		newRemaining := lot.RemainingMicros - sellFromLot
		err = p.queries.UpdateLotRemaining(ctx, db.UpdateLotRemainingParams{
			ID:              lot.ID,
			RemainingMicros: newRemaining,
		})
		if err != nil {
			return fmt.Errorf("failed to update lot remaining: %w", err)
		}

		slog.Debug("matched sell to lot",
			"lot_id", lot.ID,
			"quantity", sellFromLot,
			"cost_basis", costBasis,
			"proceeds", lotProceeds,
			"gain", gain,
			"holding_period", holdingPeriod,
		)

		remainingToSell -= sellFromLot
	}

	if remainingToSell > 0 {
		slog.Warn("sell quantity exceeds available lots",
			"transaction_id", txn.ID,
			"symbol", txn.Symbol,
			"unmatched_quantity", remainingToSell,
		)
	}

	return nil
}

func sortByDateAsc(txns []db.ListTransactionsByAccountRow) []db.ListTransactionsByAccountRow {
	sorted := make([]db.ListTransactionsByAccountRow, len(txns))
	copy(sorted, txns)

	for i := 0; i < len(sorted)-1; i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[j].TransactionDate.Time.Before(sorted[i].TransactionDate.Time) {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	return sorted
}
