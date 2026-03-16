---
name: import
description: Import brokerage transaction CSVs into the holdings service, process tax lots, and compare computed holdings against scraped positions.
user_invocable: true
argument: "<broker> [account-dir] — broker name and optional specific account directory (e.g., etrade maya-3758)"
---

# Import — Brokerage Data Pipeline

Import transaction CSVs from `data/imports/<broker>/` into the holdings service, process tax lots, and compare the resulting holdings against scraped position data.

## Invocation

```
/import etrade              # all accounts
/import etrade maya-3758    # specific account
```

## Allowed Tools

- `Bash` — only: `go run`, `ls`, `wc`
- `Read`, `Glob`

## Account Mapping

Each broker has a mapping from directory names to account names used by the holdings CLI.

### E*TRADE

| Directory | Account Name |
|-----------|-------------|
| maya-3758 | Maya |
| joint-2060 | Joint 2060 |
| joint-3652 | Joint 3652 |
| joint-2813 | Joint 2813 |

### Merrill Lynch

| Directory | Account Name |
|-----------|-------------|
| managed-2241 | Managed |
| other-0282 | Other |

## Workflow

### 1. Pre-flight

- Confirm `data/imports/<broker>/` exists
- List account directories that contain CSV files
- If a specific account was given, validate it exists
- For each account, list available CSVs (sorted by year) and most recent `positions_*.json`
- Ask user which accounts to import (default: all with CSVs)

### 2. Pipeline (per account)

Run from `services/holdings/` directory. All commands use `go run cmd/main.go`.

For each account, run these steps in order:

```bash
# 1. Clear existing data for clean reimport
go run cmd/main.go lots clear --account-name "<name>"

# 2. Import all CSVs (year files ascending)
go run cmd/main.go import --broker etrade --account-name "<name>" --file <csv1> --file <csv2> ...

# 3. Process tax lots (FIFO matching)
go run cmd/main.go lots process --account-name "<name>"

# 4. Auto-fix lot gaps from positions JSON (if available)
go run cmd/main.go lots check --fix --account-name "<name>" --positions-file <most_recent_positions.json>

# 5. Rebuild lots with opening balances included
go run cmd/main.go lots process --account-name "<name>"

# 6. Show computed holdings
go run cmd/main.go holdings list --account-name "<name>"

# 7. Verify no remaining gaps
go run cmd/main.go lots check --account-name "<name>"

# 8. Generate cash transactions from trade history
go run cmd/main.go cash generate --account-name "<name>"

# 9. Show cash balance
go run cmd/main.go cash balance --account-name "<name>"

# 10. Reconcile cash against positions JSON (sets opening balance if needed)
go run cmd/main.go cash reconcile --account-name "<name>" --positions-file <most_recent_positions.json>
```

**Auto-fix behavior:** Step 4 uses the most recent `positions_*.json` for the account. If no positions file exists, fall back to `lots check` without `--fix` and report gaps. The positions file provides `price_paid` (average cost per share) which is used to compute approximate cost basis for opening balance transactions.

**File ordering:** Year files ascending (`transactions_2021.csv`, `transactions_2022.csv`, ...). Pass all as `--file` flags to a single import command. Opening balances are synthesized automatically from the positions JSON in step 4 — do NOT import `transactions_opening.csv` files.

**CSV paths:** Use paths relative to `services/holdings/`, e.g., `../../data/imports/etrade/maya-3758/transactions_2026.csv`. Only use files from `data/imports/<broker>/`.

### 3. Position Comparison

After `holdings list` output, compare against the most recent `positions_*.json` for that account:

1. Read the positions JSON from `data/imports/<broker>/<dir>/`
2. Parse the `holdings list` output for symbol quantities
3. Compare each symbol's quantity (holdings computed vs positions scraped)
4. Print a comparison table:

```
Symbol    Computed    Scraped     Match
APP       140.00      140         ✓
WMPXX     118200.00   118359.56   ✗ (money market — daily accrual)
SOFI      10000.00    10000       ✓
```

**Expected mismatches:**
- Money market funds (WMPXX, VMFXX, etc.) — quantities change daily from accrual
- Positions from before CSV date range — `lots check` will flag these as needing opening balances

### 4. Summary

After all accounts, print:

```
Account       Txns   Holdings   Lot Gaps   Cash
Maya          245    11         0          $100,061.82
Joint 2060    180    8          2          $45,230.00
...
```

Flag accounts that need opening balances (lot gaps > 0).

## Key Behaviors

- **Standalone.** This skill runs independently from `/gather`. It works with whatever CSVs are already present.
- **Clean reimport.** Always clears and reimports from scratch — idempotent.
- **All file paths relative to `services/holdings/`.** The Go CLI runs from that directory.
- **Don't stop on lot gaps.** Lot gaps are expected for accounts without full history. Report them but continue.
- **Position comparison is informational.** Mismatches don't block the pipeline.
