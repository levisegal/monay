# LPL Financial (Account View) Navigation Guide

## Login

- URL: `https://accountview.lpl.com/web/login`
- User logs in manually
- MFA: typical browser-based MFA
- After login, user lands on: Overview page (`/web/overview`)

## Known Accounts

| Account | Number | Type | Notes |
|---------|--------|------|-------|
| Bond Portfolio | 56005516 (••••5516) | Muni Bonds | ~20 CA muni bonds + cash |
| BROKER-NR | ••••3530 | — | Essentially empty ($0.02) |
| Equities | 25274015 (••••4015) | Equities | 12 equity/ETF positions + cash |

Advisor: Gerber Kawasaki

## Account Overview

Overview page (`/web/overview`) shows all accounts in a table with:
- Account Nickname / Number
- Account Value
- Day Change ($)
- Day Change (%)

Also shows: Value Over Time chart, Asset Allocation, Top Positions, Market Information.

## Per-Account Navigation

### Step 1: Click into account

Click account name on the overview page to go to Positions page for that account.

Alternatively, from any page, use the account selector dropdown (click account name near "Positions" or "Activity" heading) which shows checkboxes for all accounts.

### Step 2: Navigate to Positions

`Accounts` nav menu → `Positions` (URL: `/web/positions`)

Positions are grouped "By Account" and collapsed by default. Click the chevron (expand arrow) next to the account name to reveal individual positions.

### Portfolios Page

- Tab options: POSITIONS, OVERALL RETURNS, ESTIMATED INCOME, REALIZED GAIN/LOSS
- Columns: SYMBOL, DESCRIPTION, QUANTITY, PRICE, DAY CHANGE ($), DAY CHANGE (%), VALUE, AS OF, TOTAL COST BASIS, UNIT COST, U G/L ($), U G/L (%), HELD IN
- Bond positions show CUSIPs in the SYMBOL column (e.g., `072024F56`)
- Equity positions show tickers (e.g., `AAPL`, `NVDA`)
- The SYMBOL column is in a separate frozen left table from the data columns
- Cash position labeled "Insured Cash Account" with dashes for most fields
- "Data grid with N rows" in the group label tells total position count

### Step 3: Screenshot portfolio

```bash
playwright-cli screenshot --filename=data/imports/lpl/<account>/screenshot_portfolio_YYYY-MM-DD.png
```

### Step 4: Navigate to Activity

`Accounts` nav menu → `Activity` (URL: `/web/activity`)

### Step 5: Set date range

Click the Date Range button (shows current selection like "1 Month").
Select from dropdown: **1 Month**, **YTD**, **1 Year**, **Prev Years**, **Custom**

**Use "1 Year" for standard gather** — it covers ~12 months back.

Note: Date range resets to "1 Month" when navigating away from Activity page.

### Step 6: Download CSV

1. Click "Export" button
2. Select "Export to CSV" from dropdown
3. File downloads as `Activity.csv` to `.playwright-cli/`
4. Move immediately (next download overwrites)

```bash
mv .playwright-cli/Activity.csv data/imports/lpl/<account>/transactions_2025_2026.csv
```

### Step 7: Switch to next account

Use the account selector dropdown (click account name near heading). Uncheck current account, check next account. Re-select "1 Year" date range (it resets).

## CSV Format

```
Date,Activity,Symbol,Description,Quantity,Unit Price,Value,Held In,Account Nickname,Account Number
2/27/2026,interest,9999227,"INSURED CASH ACCOUNT 022726...",‐,‐,$0.03,cash,Bond Portfolio,56005516
2/12/2026,cash dividend,AAPL,APPLE INC 021226...,‐,‐,$52.55,cash,Equities,25274015
2/12/2026,dividend reinvest,AAPL,APPLE INC,0.191,$275.13,-$52.55,cash,Equities,25274015
```

Activity types seen: interest, reinvest interest, ach funds, cash dividend, dividend reinvest, buy, sell, fee

## Date Range Strategy

- **1 Year** covers standard gather needs (current year + most of prior year)
- **Prev Years** has a sub-menu with year options but doesn't work via accessibility tree automation
- **Custom** also has sub-menu inputs that don't render as accessible form elements
- For backfill, may need to manually interact with Prev Years or Custom date pickers

## Known Quirks

- CSV always downloads as `Activity.csv` — must move immediately between accounts
- Date range resets to "1 Month" when navigating to Positions and back to Activity
- "Prev Years" and "Custom" date range options have sub-menus that are inaccessible to playwright automation
- Account selector uses checkboxes (multi-select) — remember to uncheck previous account
- Bond positions show CUSIPs, not tickers — need CUSIP-to-ticker mapping for import
- Insured Cash Account uses symbol "9999227" in transactions
- Value column in CSV has tab character prefix (e.g., `\t$0.03`)

## Import Commands

```bash
cd services/holdings
# TODO: Create LPL import parser
```
