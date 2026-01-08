# Cash & Income Tracking

## Overview

Track cash flows per account: interest, dividends, fees, transfers, and buy/sell proceeds.

## Current State

**What we have:**
- Transactions imported (buy, sell, interest, dividend, transfer, fee)
- Tax lots for positions (cost basis, remaining shares)
- Holdings rollup from lots

**What's missing:**
- Cash balance tracking
- Income aggregation (interest, dividends)
- Fee tracking
- Bond-specific handling (coupon payments, accrued interest)

## Proposed Schema

### `cash_transactions` table
```sql
create table monay.cash_transactions (
    id text primary key,
    account_id text not null references monay.accounts(id),
    transaction_id text references monay.transactions(id),  -- optional link
    transaction_date date not null,
    cash_type text not null,  -- interest, dividend, fee, proceeds, deposit, withdrawal
    amount_micros bigint not null,  -- positive = inflow, negative = outflow
    security_id text references monay.securities(id),  -- for interest/dividend source
    description text,
    created_at timestamptz default now()
);
```

### Cash Types
| Type | Description | Sign |
|------|-------------|------|
| `interest` | Bond coupons, sweep interest | + |
| `dividend` | Stock dividends | + |
| `proceeds` | Sell proceeds | + |
| `purchase` | Buy outflow | - |
| `fee` | Advisory fees, reorg fees | - |
| `deposit` | ACH/wire in | + |
| `withdrawal` | ACH/wire out | - |
| `distribution` | Transfer to linked account | - |

## Bond-Specific Considerations

### What makes bonds different:
1. **Accrued interest** - buyer pays seller for interest since last coupon
2. **Premium/discount** - bonds trade above/below par
3. **Coupon payments** - semi-annual interest based on face value
4. **Maturity/call** - principal returned at par

### LPL Transaction Mapping
| LPL Activity | Cash Type | Notes |
|--------------|-----------|-------|
| `buy` | `purchase` | Amount includes accrued interest |
| `sell` | `proceeds` | Amount includes accrued interest |
| `interest` | `interest` | Coupon payment |
| `reinvest interest` | `interest` + `purchase` | Interest reinvested |
| `ica transfer` | `deposit`/`withdrawal` | Internal cash movement |
| `journal` | `distribution` | Transfer to linked account |
| `fee` | `fee` | Advisory fee |

## Implementation Tasks

### Phase 1: Cash Balance
- [ ] Add `cash_transactions` table + migration
- [ ] Update transaction processor to create cash entries
- [ ] Add `cash balance` command to show current cash

### Phase 2: Income Reporting  
- [ ] Add `income summary` command (interest, dividends by period)
- [ ] Group by security for bond interest tracking
- [ ] YTD/monthly rollups

### Phase 3: Bond Enhancements
- [ ] Track accrued interest on purchases/sales
- [ ] Bond maturity/call handling
- [ ] Amortization of premium/discount (optional)

## CLI Commands

```bash
# Show cash balance
go run cmd/main.go cash balance --account-name "Bond Portfolio"

# Show income summary
go run cmd/main.go income summary --account-name "Bond Portfolio" --year 2024

# Show income by security
go run cmd/main.go income by-security --account-name "Bond Portfolio"
```

## Example Output

```
=== Bond Portfolio: Cash Activity (2024) ===

INCOME:
  Interest (bonds)     $4,850.00
  Interest (sweep)        $12.50
  ─────────────────────────────
  Total Income         $4,862.50

EXPENSES:
  Advisory Fees         -$847.77
  ─────────────────────────────
  Net Income           $4,014.73

DISTRIBUTIONS:
  To linked account   -$5,575.00

CURRENT CASH:          $1,234.56
```


