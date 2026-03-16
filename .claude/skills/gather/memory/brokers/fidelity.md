# Fidelity — Broker Memory

## Accounts

### Login 1 (gathered 2026-03-02)

| Directory | Account Number | Account Type | Balance |
|-----------|---------------|--------------|---------|
| rollover-0850 | 257524850 | Rollover IRA | $2,830,795 |
| sep-0740 | 249148740 | SEP-IRA | $56,745 |
| roth-2094 | 114712094 | ROTH IRA | $6,520 |
| joint-7807 | X66087807 | Joint WROS - TOD | $3,672 |
| astellas-3509 | 93509 | ASTELLAS RSP | $40 |

Zero-balance accounts (skipped): GILEAD SCIENCES 401K 40049, INSTIL BIO RET PLAN 56193

### Login 2 (gathered 2026-03-04)

| Directory | Account Number | Account Type | Balance |
|-----------|---------------|--------------|---------|
| rollover-1820 | 114641820 | Rollover IRA | $735,773 |
| roth-2108 | 114712108 | ROTH IRA | $278,255 |
| joint-7807 | X66087807 | Joint WROS - TOD | $3,767 |

Zero-balance accounts (skipped): JEWISH FED CNCL-L A 68440

**Note:** Joint WROS X66087807 appears on BOTH logins (same account, shared between users).

## Login

- URL: `https://digital.fidelity.com/prgw/digital/login/full-page`
- Manual login + MFA required
- Post-login lands on Portfolio Summary: `/ftgw/digital/portfolio/summary`

## Navigation

- Account selector in left nav, always visible
- Tabs: Summary, Positions, Activity & Orders, Balances, More
- Clicking an account in the selector changes the account for whichever tab you're on
- Account refs in the selector remain stable across tab switches

## Positions Page

- URL pattern: `/ftgw/digital/portfolio/positions#<acct_number>`
- Table columns: Symbol, Last price, Today's gain/loss, Total gain/loss, Current value, % of account, Quantity, Cost basis, 52-week range
- **Symbol rows and data rows are in SEPARATE rowgroups** — parse by matching order
- First data row corresponds to the account header (empty), not the first position
- "Not Priced Today" appears on retirement plan positions (updated less frequently)
- Some positions use CUSIPs instead of tickers (e.g., 437355100, 14022L801)

## Activity Page

- URL pattern: `/ftgw/digital/portfolio/activity#<acct_number>`
- Time period filter: button shows current range (e.g., "Past 30 days")
- Custom date range: radio button "Custom" → From Date / To Date inputs
- Date inputs are HTML `type="date"` requiring YYYY-MM-DD format
- **365-day maximum range** — error "Your request exceeds the 365-day range limit" if exceeded
- Min date: 5 years back from today (e.g., 2021-03-02 for 2026-03-02)
- **Leap year gotcha**: 2023-03-02 → 2024-03-02 is 366 days. Use 2023-03-03 instead.
- Apply button sometimes stays disabled after filling dates — click the To Date field to trigger validation

## CSV Download

- Download button → submenu → "Download as CSV"
- Filename pattern: `History_for_Account_{acct_number}.csv` (downloads as `History-for-Account-{acct_number}.csv`)
- Downloads land in `.playwright-cli/` — move immediately

### Standard CSV Format (brokerage accounts)

```
Run Date,Action,Symbol,Description,Type,Price ($),Quantity,Commission ($),Fees ($),Accrued Interest ($),Amount ($),Cash Balance ($),Settlement Date
```

Used by: rollover-0850, sep-0740, roth-2094, joint-7807

### Retirement Plan CSV Format (Astellas RSP)

```
Date,Investment,Transaction Type,Shares/Unit,Amount ($)
```

Used by: astellas-3509. Completely different columns — needs separate parser.

## Quirks

- Cash symbols vary: SPAXX (money market), FDRXX (money market), FCASH (via CUSIP 315994103)
- Joint account X66087807 uses CUSIP 315994103 for cash, not a ticker
- Rollover IRA has 15 positions but only 3 transactions in 5 years (transferred in-kind)
- Astellas RSP is a tiny retirement plan ($40) with different CSV format
- BOM (`\xEF\xBB\xBF`) at start of CSV files (UTF-8 BOM)
- Blank line after BOM before header row in some files
