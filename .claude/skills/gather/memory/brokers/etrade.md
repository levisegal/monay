# E*TRADE Gather Findings

## CSV Format (New Activity Page)

The new E*TRADE Activity page exports a **different CSV format** than the classic export.

**New format** (Activity page download):
```
All Transactions Activity Types

Account Activity for maya -3758 from Current Year

Total:,-18247.05

Activity/Trade Date,Transaction Date,Settlement Date,Activity Type,Description,Symbol,Cusip,Quantity #,Price $,Amount $,Commission,Category,Note
02/17/26,02/17/26,02/18/26,Bought,UNITED PARCEL SER INC CL-B UNSOLICITED TRADE,UPS,--,78.0,115.85,-9036.3,0.0,--,--
```

**Old format** (classic transactions page, what the importer expects):
```
For Account:,#####3758

TransactionDate,TransactionType,SecurityType,Symbol,Quantity,Amount,Price,Commission,Description
```

Key differences:
- Header: `Account Activity for maya -3758` vs `For Account:,#####3758`
- Columns: New has `Activity/Trade Date,Transaction Date,Settlement Date,Activity Type,Description,Symbol,Cusip,Quantity #,Price $,Amount $,Commission,Category,Note`
- Old has: `TransactionDate,TransactionType,SecurityType,Symbol,Quantity,Amount,Price,Commission,Description`
- New format has 3 date columns (Activity, Transaction, Settlement)
- New format includes Cusip, Commission, Category, Note columns
- Activity types differ: "Bought" vs "Buy", "Online Transfer" vs "Transfer", "Interest Income" vs "Interest"
- Amounts: negative for purchases in new format (e.g., `-9036.3`)
- Stock plan-linked accounts (e.g., joint-3652/GILD) include an extra disclaimer paragraph about Morgan Stanley activity dates

**Action needed:** Update `services/holdings/importer/etrade.go` to handle the new format, or find the classic export page.

## Accounts

The Complete View page shows 6 accounts:

| Account | Last 4 | Type | In Activity Dropdown |
|---------|--------|------|---------------------|
| maya | -3758 | Individual | Yes |
| Joint JTWROS | -2060 | Joint | Yes |
| Joint JTWROS | -3652 | Joint (linked to GILD stock plan) | Yes |
| Joint JTWROS | -2813 | Joint (linked to KITE stock plan) | Yes |
| Stock Plan (GILD) | -3652 | Stock Plan | No |
| Stock Plan (KITE) | -2813 | Stock Plan ($0.00 balance) | No |

## Navigation

- **Login URL:** `https://us.etrade.com/etx/pxy/login` (direct, better than `/home` which lands on marketing)
- **Complete View:** `https://us.etrade.com/etx/hw/v2/accountshome`
- **Activity page:** `https://us.etrade.com/etx/pxy/accounts/transactions`
- **Old account list URL is dead:** `https://us.etrade.com/etx/hw/accountlist` → 404
- Intro modal ("Introducing the new Activity experience") may appear on first visit — dismiss with Close button
- Activity page has its own account selector dropdown — doesn't inherit from portfolio navigation

## Download Flow

1. Navigate to Activity page
2. Select account from dropdown
3. Select Duration (Current Year, Prior Year, etc.)
4. Click Download menuitem in top action bar
5. Dialog appears — **format always defaults to Excel, must switch to "CSV file" every time** (not sticky)
6. Click Download button in dialog
7. File downloads to `.playwright-cli/DownloadTxnHistory.csv` (always same filename)
8. **Move immediately** before next download

**Duration resets to "Last 30 Days" when switching accounts** — must re-select after each account change.

## Custom Date Range (Backfill)

The Duration dropdown "Custom" option allows arbitrary date ranges with these constraints:
- **Lookback limit: ~2 years** from today (e.g., Feb 22, 2024 works on Feb 22, 2026; Jan 1, 2024 fails)
- **Max range per request: ~6 months** — full-year ranges fail even within the lookback window
- **Strategy:** Download in two halves per year: H1 (Feb 22-Jun 30) and H2 (Jul 1-Dec 31), combine afterward
- Jan 1 to ~Feb 21 of the oldest reachable year is lost (within the 2-year cutoff)
- Pre-2024 data is not available via the web UI at all (as of Feb 2026)

## Files Gathered (as of 2026-02-21)

| Account | 2024 | 2025 | 2026 |
|---------|------|------|------|
| maya-3758 | 8 rows (EA divs) | 40 lines | 33 lines |
| joint-2060 | 42 rows (PTOAX, VGHCX, VFIAX, etc.) | 84 lines | 25 lines |
| joint-3652 | 8 rows (GILD divs) | 26 lines | 15 lines |
| joint-2813 | 6 rows (FCX divs) | 26 lines | 20 lines |
| stockplan-3652 (GILD) | skip (not in dropdown) | skip | skip |
| stockplan-2813 (KITE) | skip ($0 balance) | skip | skip |
