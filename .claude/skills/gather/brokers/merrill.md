# Merrill Lynch Navigation Guide

## Login

- URL: `https://olui2.fs.ml.com/accounts-overview/`
- User logs in manually (may redirect through Bank of America SSO)
- After login, user lands on Accounts Overview dashboard

## Known Accounts

| Directory | Account Name | Identifier | Type |
|-----------|-------------|-----------|------|
| managed-2241 | Managed | CMA 5VT-22241 | Managed/advisory |
| other-0282 | Other | CMA 5VT-10282 | Self-directed |

## Account Overview

Dashboard shows account cards with name, identifier, and total value. Top nav has: Accounts, Holdings, Transfers & Bill Pay, Research, Rewards & Benefits.

## Holdings / Portfolios Page

URL: `/TFPHoldings/HoldingsByAccount.aspx`

Navigate via top nav → Holdings button (e2124).

### Page Structure

- Account selector dropdown at top ("Go To All Accounts")
- "View by" radio buttons: Product Class, Security, Account
- Use **Account** view to get per-account position breakdowns

### Columns

Symbol, Description, Quantity, Price, Day's Change, Value, Day's Value Change, Unrealized Gain/Loss, Last Updated

### Special Rows

- **Cash Balance**: appears at bottom of each account section (not a position)
- **Accrued Interest**: appears after bond positions
- **ML Bank Deposit Program (990286916)**: sweep account, appears as a position at $1.00
- **TMCXX**: money market fund (Blackrock Liquidity Funds TempCash)
- **Cumulative Investment Return**: summary rows between position groups

### Positions JSON

Merrill doesn't show `price_paid` directly. Compute from: `price_paid = (value - unrealized_gain) / quantity`

For bonds: quantity is face value, price is per $100 face value. Same formula applies.

For money market/deposit positions with `--` gain: use `price_paid = 1.00`.

## Activity / Transactions Page

URL: `/TFPActivity/Activity.aspx`

Navigate via top nav → Accounts dropdown → Activity link.

### Date Range

Combobox "Show activity for" with options:
- Last 30 days, Current month, Prior month
- Last 3/6/9/12/24 months
- Custom range

**Maximum online range: ~24 months.** Activity prior to ~Feb 2024 requires paper statements.

### Account Filter

"Accounts in View" dropdown → checkboxes for each account → Apply/Cancel.

Default is all accounts.

### Export

1. Click "Export All Activity" link
2. Submenu appears: CSV (.CSV), Tab-Delimited (.DNL), Plain Text (.TXT)
3. Click "Export as Comma-Separated Spreadsheet (.CSV)"
4. Confirmation popup: "Exporting All Activities may take longer than expected. Would you like to continue?" → Click "Yes"
5. File downloads to `.playwright-cli/` directory

Filename format: `ExportData{DDMMYYYY}{HHMMSS}.csv`

### CSV Format

```
"Trade Date" ,"Settlement Date" ,"Account" ,"Description" ,"Type" ,"Symbol/ CUSIP" ,"Quantity" ,"Price" ,"Amount" ," "
"02/27/2026" ,"02/27/2026" ,"Managed CMA 5VT-22241" ,"Bank Interest ML BANK DEPOSIT PROGRAM FROM 01/30 THRU 02/26" ,"" ,"990286916" ,"" ,"" ,"$2.49" ,""
```

Header row has spaces before commas (Merrill's quirky formatting). The existing parser in `services/holdings/importer/merrill.go` handles this.

Both accounts are mixed in one export file. Split by account name + year for import.

### Splitting

Match account by `"Managed CMA"` or `"Other CMA"` in the line.
Match year by trade date at start of line: `^"MM/DD/YYYY"`.
Don't use year in the whole line — descriptions can contain dates like "PAY DATE 03/07/2024".

## Date Range Strategy

Export "Last 24 months" in one shot. Split resulting CSV by account and year.
Current year always re-downloaded.

## Known Quirks

- Export includes metadata rows (exported on, selected accounts, total line) — these don't match the data header and are safely ignored by grep-based splitting
- Amount field includes `$` prefix and negative amounts use `-$` (e.g., `-$600.00`)
- "Type" column is often empty — transaction type is embedded in the "Description" field
- Quantities can be negative for withdrawals/sells
- The page shows "Day's Change" and "Day's Value Change" as $0.00 after market hours
- Bond quantities are face values (e.g., 30,000), prices are per $100 face value
