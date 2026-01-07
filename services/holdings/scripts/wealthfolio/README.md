# Wealthfolio Import Scripts

Scripts to transform brokerage transaction exports into Wealthfolio's import format.

## Wealthfolio Transaction Types

| Type | Description |
|------|-------------|
| `BUY` | Purchase securities |
| `SELL` | Sell securities |
| `DIVIDEND` | Cash dividend received |
| `INTEREST` | Interest income |
| `DEPOSIT` | Cash deposit into account |
| `WITHDRAWAL` | Cash withdrawal from account |
| `ADD_HOLDING` | Shares added (reorg, gift, transfer in-kind) |
| `REMOVE_HOLDING` | Shares removed (reorg, gift, transfer out) |
| `TRANSFER_IN` | Cash transfer in |
| `TRANSFER_OUT` | Cash transfer out |
| `FEE` | Fees charged |
| `TAX` | Tax withheld |
| `SPLIT` | Stock split |

## Broker Mappings

### E*TRADE

| E*TRADE Type | → Wealthfolio | Notes |
|--------------|---------------|-------|
| `Bought` | `BUY` | |
| `Sold` | `SELL` | |
| `Dividend` | `DIVIDEND` | |
| `Qualified Dividend` | `DIVIDEND` | |
| `Interest` | `INTEREST` | |
| `Interest Income` | `INTEREST` | |
| `Transfer` (+amount) | `TRANSFER_IN` | ACH deposit |
| `Transfer` (-amount) | `TRANSFER_OUT` | ACH withdrawal |
| `Reorganization` (+qty) | `ADD_HOLDING` | Shares received from merger/spinoff |
| `Reorganization` (-qty) | `REMOVE_HOLDING` | Shares removed from merger |
| `Reorganization` (0 qty, +amt) | `DIVIDEND` | Cash in lieu of fractions |
| `Reorganization` (0 qty, -amt) | `FEE` | Reorg fee |
| `Misc Trade` | `FEE` | Usually reorg fees |
| `Adjustment` | `ADD_HOLDING` | Free shares received |

### LPL Financial

| LPL Type | → Wealthfolio | Notes |
|----------|---------------|-------|
| `buy` | `BUY` | |
| `sell` | `SELL` | |
| `interest` | `INTEREST` | Bond coupon payments |
| `reinvest interest` | `BUY` | Interest reinvested into cash account |
| `ica transfer` (+amt) | `TRANSFER_IN` | Internal cash transfer in |
| `ica transfer` (-amt) | `TRANSFER_OUT` | Internal cash transfer out |
| `journal` | `TRANSFER_OUT` | Distribution to linked account |
| `fee` | `FEE` | Advisory fees |

## Usage

```bash
# Combine E*TRADE transactions into Wealthfolio format
./combine.sh etrade maya-3758
./combine.sh etrade joint-2060

# Combine LPL transactions (when supported)
./combine.sh lpl bond-5516
```

## Adding New Brokers

1. Add mapping table to this README
2. Add broker case to `transforms/<broker>.awk`
3. Update `combine.sh` to detect broker and apply correct transform

## Known Issues

### Symbols to Skip
- `#XXXXXX` - E*TRADE internal account references
- `XXXXXXXX` (CUSIP format) - Bond identifiers, need ticker lookup
- `9999227` - LPL cash account reference
- `CASH` - Cash placeholder

### Transactions That Need Manual Review
- Reorganizations with CUSIP symbols (no ticker)
- Journal entries for distributions
- Corporate actions (mergers, spinoffs)

