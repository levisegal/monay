# E*TRADE Navigation Guide

## Login

- URL: `https://us.etrade.com/home` (redirects to `/etx/hw/v2/accountshome` after login)
- **Old URL `/etx/hw/accountlist` is dead** — returns 404
- User logs in manually at the E*TRADE login page
- MFA: typically SMS or app-based — user handles this
- After login, user should land on the Complete View page

## Known Accounts

| Account | Last 4 | Type | In Activity Dropdown | Notes |
|---------|--------|------|---------------------|-------|
| maya-3758 | 3758 | Individual | Yes | Standard brokerage |
| joint-2060 | 2060 | Joint | Yes | Joint brokerage |
| joint-3652 | 3652 | Joint | Yes | Linked to Stock Plan (GILD) |
| joint-2813 | 2813 | Joint | Yes | Linked to Stock Plan (KITE) |
| stockplan-3652 | 3652 | Stock Plan (GILD) | No | Not in Activity dropdown |
| stockplan-2813 | 2813 | Stock Plan (KITE) | No | Not in Activity dropdown, $0.00 balance |

## Complete View (Account Overview)

URL: `https://us.etrade.com/etx/hw/v2/accountshome`

Shows all accounts with name, last 4 digits, value, and day's gain. Stock plan accounts and their linked brokerage accounts are grouped together.

## Portfolios Page (Positions)

URL: `https://us.etrade.com/etx/pxy/portfolios/positions`

Shows positions for a single account at a time. Account is selected via a native HTML `<select>` dropdown — use `playwright-cli select` to change it.

### Navigation

1. `playwright-cli goto https://us.etrade.com/etx/pxy/portfolios/positions`
2. Select account from the `<select>` dropdown
3. `playwright-cli snapshot` — capture the accessibility tree

### Grid Row Structure

Each position is a `row` containing a `rowheader` (symbol) and `gridcell` values in this order:

```
rowheader  → Symbol
gridcell 1 → Last Price
gridcell 2 → Change $
gridcell 3 → Change %
gridcell 4 → Quantity
gridcell 5 → Price Paid
gridcell 6 → Day's Gain
gridcell 7 → Total Gain
gridcell 8 → Total Gain %
gridcell 9 → Value
```

Example row from the snapshot YAML:
```yaml
row "APP Alert Note 418.68 6.68 1.62% 140 398.0588 935.20 2,886.97 5.18% 58,615.20":
  - rowheader "APP"
  - gridcell "418.68"
  - gridcell "6.68"
  - gridcell "1.62%"
  - gridcell "140"
  - gridcell "398.0588"
  - gridcell "935.20"
  - gridcell "2,886.97"
  - gridcell "5.18%"
  - gridcell "58,615.20"
```

Map to JSON fields:
- `symbol` ← rowheader
- `last_price` ← gridcell 1
- `quantity` ← gridcell 4
- `price_paid` ← gridcell 5
- `total_gain` ← gridcell 7 (strip commas)
- `total_gain_pct` ← gridcell 8 (strip `%`)
- `value` ← gridcell 9 (strip commas)

### Account Summary

The summary section uses definition list `term`/`definition` pairs:

| Term | Example | JSON field |
|------|---------|------------|
| Net Account Value | $1,151,838.24 | `net_value` |
| Unrealized Gain | $179,665.83 | `unrealized_gain` |
| Unrealized Gain % | 20.60% | `unrealized_gain_pct` |
| Day's Gain | -$1,537.65 | `days_gain` |
| Cash Purchasing Power | $100,061.82 | `cash_purchasing_power` |

Strip `$`, `,`, and `%` when parsing. Negative values use `-$` prefix (e.g., `-$1,537.65` → `-1537.65`).

### Stock Plan Accounts

Stock Plan accounts (ESPP/RSU) use a different page layout ("Stock Plan Holdings") with different columns. **Skip position scraping for stock plan accounts** — only scrape regular brokerage and joint accounts.

## Activity Page (Transactions)

URL: `https://us.etrade.com/etx/pxy/accounts/transactions`

This is the **new** Activity experience (as of 2026). Key elements:
- **Account dropdown** — select which account to view
- **Activity type dropdown** — filter by transaction type (default: All Transactions)
- **Duration dropdown** — date range with options: Last 7/30/60/90 Days, Current Year, Prior Year, quarterly ranges
- **Download menuitem** — in the top action bar (Refresh | Help | Print | Download)

First visit shows an intro modal ("Introducing the new Activity experience") — dismiss it.

## Per-Account Download Flow

### Step 1: Navigate to Activity page

```
playwright-cli goto https://us.etrade.com/etx/pxy/accounts/transactions
```

Dismiss the intro modal if it appears (click Close button).

### Step 2: Select account

Open the "Accounts" combobox and select the target account by name.

### Step 3: Select duration

Open the "Duration" combobox. Select:
- **"Current Year"** for current year transactions
- **"Prior Year"** for previous year transactions

Download one year at a time.

### Step 4: Click Download

Click the "Download" menuitem in the top action bar. A dialog appears.

### Step 5: Select CSV format

The dialog defaults to "Spreadsheet format including Microsoft Excel". **Change it to "CSV file"** using the "Select data format" combobox.

### Step 6: Click Download in dialog

Click the "Download" button. File downloads to `.playwright-cli/DownloadTxnHistory.csv`.

### Step 7: Move and verify

```bash
mv .playwright-cli/DownloadTxnHistory.csv data/imports/etrade/<account>/transactions_<year>.csv
head -5 data/imports/etrade/<account>/transactions_<year>.csv
wc -l data/imports/etrade/<account>/transactions_<year>.csv
```

Expected header (new format):
```
All Transactions Activity Types

Account Activity for maya -3758 from Current Year

Total:,-18247.05

Activity/Trade Date,Transaction Date,Settlement Date,Activity Type,Description,Symbol,Cusip,Quantity #,Price $,Amount $,Commission,Category,Note
```

Confirm the account name in "Account Activity for..." matches the target account.

### Step 8: Repeat for next year/account

Change Duration or Account and repeat steps 4-7. **Move each download immediately** — the filename is always `DownloadTxnHistory.csv` and will be overwritten.

## Date Range Strategy

Download one CSV per year:
- Current year (2026): select "Current Year"
- Previous year (2025): select "Prior Year"

For incremental updates, just download the current year.

## Backfill

**Known duration options:** Last 7 Days, Last 30 Days, Last 60 Days, Last 90 Days, Current Year, Prior Year, Q4-Q1 2025 quarterly, Custom.

Preset quarterly options only go back to Q1 of the prior year. Use **Custom** date ranges to reach older data.

### Custom Date Range Limits

- **Lookback: ~2 years** from today (dates older than ~24 months are rejected as "invalid")
- **Max range per request: ~6 months** — full-year custom ranges fail
- **Strategy:** Download each year in two halves, then combine:
  - H1: `02/22/<year>` to `06/30/<year>` (start date = 2 years before today)
  - H2: `07/01/<year>` to `12/31/<year>`
- Jan 1 through ~Feb 21 of the oldest reachable year is lost
- **Pre-2024 data is not available via the web UI** (as of Feb 2026)

### Backfill Procedure

For each account:

1. Select Custom duration
2. Fill "Date from" and "Date to" for H1 of the target year
3. Press Enter to load results
4. Download → switch format to CSV → click Download → move file
5. Change dates to H2 and repeat
6. Combine H1 and H2 into `transactions_<year>.csv` (H2 rows first, H1 rows second, maintaining descending date order; strip footer disclaimers)
7. If both halves have zero data rows, stop — broker has no more data

### Existing Files

Skip years where `transactions_<year>.csv` already exists. Always re-download the current year.

## Known Quirks

- **Stock Plan accounts** are not accessible from the Activity page dropdown. Skip them during gather.
- **Intro modal** appears on first visit to Activity page per session — dismiss with Close button.
- **Download filename** is always `DownloadTxnHistory.csv` regardless of account/date range.
- **CSV format changed** — the new Activity page exports a different format than the classic E*TRADE transactions page. The importer may need updating.
- **Session timeout:** E*TRADE sessions expire after ~15 min of inactivity.
- **Browser may die after download** — check that playwright-cli is still connected before the next operation.

## Import Commands

After gathering, import with:

```bash
cd services/holdings
./scripts/import-account.sh etrade "Maya" ../../data/imports/etrade/maya-3758
./scripts/import-account.sh etrade "Joint 2060" ../../data/imports/etrade/joint-2060
./scripts/import-account.sh etrade "Joint 3652" ../../data/imports/etrade/joint-3652
./scripts/import-account.sh etrade "Joint 2813" ../../data/imports/etrade/joint-2813
```

**Note:** The importer currently expects the old CSV format. The new format will need a parser update before import will work.
