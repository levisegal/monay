# Transaction Types & Edge Cases

## DRIP (Dividend Reinvestment Plan)

Instead of receiving cash dividends, DRIP automatically purchases additional shares.

**How it appears in E*TRADE CSV:**
```csv
12/30/25,Dividend,EQ,GILD,31.518,-3926.73,124.58692,0,GILEAD SCIENCE DIVIDEND REINVESTMENT
```

- `TransactionType`: "Dividend" 
- `Quantity`: positive (shares being purchased)
- `Amount`: negative (dividend "spent" on shares)
- `Price`: per-share price

**Importer handling:**
- "Dividend" with quantity > 0 → `TransactionTypeBuy` (creates a lot)
- "Dividend" with quantity = 0 → `TransactionTypeDividend` (cash dividend, no lot)

**Gotcha:** DRIP accounts often have positions predating transaction history. Use `lots check` to identify gaps and add opening balances.

---

## Reorganization (Mergers, Spinoffs, Symbol Changes)

Corporate actions that change shares without a traditional buy/sell.

**How it appears in E*TRADE CSV:**
```csv
04/28/23,Reorganization,EQ,PRVB,-200,5000,25,0,PROVENTION BIO MERGER
04/28/23,Reorganization,EQ,SNY,50,5000,100,0,SANOFI RECEIVED FROM MERGER
```

**Importer handling:**
- Negative quantity → `TransactionTypeReorgOut` (disposes lot, like a sell)
- Positive quantity → `TransactionTypeReorgIn` (creates lot, like a buy)

---

## Security Transfers (ACAT, Account Transfers)

Shares moved from another brokerage or account.

**How it appears:**
```csv
02/25/25,Transfer,EQ,TSLA,100,25000,250,0,SECURITY TRANSFER IN
```

**Importer handling:**
- `TransactionTypeSecurityTransfer` → creates a lot with the transferred cost basis

**Note:** If cost basis isn't included in the transfer, you may need to add an opening balance manually.

---

## Money Market Funds (VMFXX, WMPXX, etc.)

Sweep accounts that hold uninvested cash as shares of a money market fund.

**Characteristics:**
- Price is always $1.00
- Quantity = dollar amount
- Dividends often auto-reinvest monthly

**Gotcha:** Some brokers don't export money market transactions in the standard transaction CSV. Check your actual cash balance against imported VMFXX/WMPXX positions.

---

## Bonds

Not yet fully supported. Bond transactions have additional complexity:
- Accrued interest at purchase/sale
- Coupon payments (interest income)
- Par value vs market value
- Maturity dates

See `docs/cash-tracking.md` for planned bond support.


