# [Broker Name] Navigation Guide

## Login

- URL: `<login_url>`
- User logs in manually
- MFA: [describe typical MFA flow]
- After login, user lands on: [describe landing page]

## Known Accounts

| Account | Identifier | Type | Notes |
|---------|-----------|------|-------|
| name-XXXX | XXXX | Type | Notes |

## Account Overview

[Describe the account list/overview page. What does it look like? How are accounts listed?]

## Per-Account Navigation

### Step 1: Click into account

[How to select a specific account from the overview]

### Step 2: Screenshot portfolio

```bash
playwright-cli screenshot --path data/imports/<broker>/<account>/screenshot_portfolio.png
```

### Step 3: Navigate to Transactions

[How to get to transaction history from the account detail page]

### Step 4: Set date range

[How to set the date range. What controls exist? What format do dates use?]

### Step 5: Download CSV

[How to initiate the CSV download. What button/link? Any format selection needed?]

### Step 6: Move and verify

```bash
mkdir -p data/imports/<broker>/<account>
mv ~/Downloads/<file>.csv data/imports/<broker>/<account>/transactions_<year>.csv
```

Expected header format:
```
[paste first 3-5 lines of a sample CSV]
```

### Step 7: Return to account overview

[How to navigate back to the account list]

## Date Range Strategy

[How to handle multi-year downloads. One file per year? Single file?]

## Known Quirks

- [Any browser-specific issues]
- [Session timeout behavior]
- [Download file naming patterns]
- [Account types with different navigation]

## Import Commands

```bash
cd services/holdings
./scripts/import-account.sh <broker> "<Account Name>" ../../data/imports/<broker>/<account>
```
