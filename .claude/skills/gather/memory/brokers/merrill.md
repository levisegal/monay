# Merrill Lynch — Broker Memory

## Accounts

- Managed (CMA 5VT-22241): ~$1.8M, managed/advisory, ~900 txns/year
- Other (CMA 5VT-10282): ~$190K, just BREIT + sweep, ~45 txns/year

## Navigation

- Main nav buttons: Accounts, Holdings, Transfers & Bill Pay, Research, Rewards & Benefits
- Activity page: Accounts dropdown → Activity (not a direct top-level nav item)
- Holdings page: direct "Holdings" button in top nav
- Account selector is a dropdown with checkboxes, not a simple select

## Export Behavior

- CSV export: Export All Activity → CSV → confirmation popup → download
- File downloads to `.playwright-cli/` with name `ExportData{DDMMYYYY}{HHMMSS}.csv`
- Both accounts come in one file — must split by "Managed CMA" / "Other CMA" in Account field
- Max online history: ~24 months (Feb 2024 → present as of Feb 2026)
- Export triggers a "may take longer" confirmation popup — must click Yes

## CSV Format Quirks

- Spaces before commas: `"value" ,"value"` (not standard CSV)
- "Type" column usually empty — type info is in "Description" prefix
- Amount includes `$` sign: `"$2.49"` or `"-$600.00"`
- Header has 10 columns including a trailing empty column `" "`
- Existing parser in `services/holdings/importer/merrill.go` handles these quirks

## Positions Scraping

- Holdings page "Account" view shows per-account tables
- No "price_paid" field — compute as `(value - unrealized_gain) / quantity`
- ML Bank Deposit (990286916) and TMCXX (money market) show `--` for gain
- Bonds show quantity as face value, price per $100
- Cash Balance is a separate row (not a position)
- Accrued Interest rows appear after bond positions
