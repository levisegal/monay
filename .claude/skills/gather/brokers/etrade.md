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
