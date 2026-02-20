# Portfolio Implementation Plan

> See [portfolio.md](./portfolio.md) for the feature spec (API design, architecture, configuration).

## Goal

Add real market data to the Monay portfolio dashboard:
1. Migrate Holdings Service from Postgres to SQLite
2. Migrate Holdings Service from Connect RPC to chi REST
3. Create new Portfolio Service (Python/FastAPI) as gateway with chart cache
4. Wire frontend with TanStack Query and generated API client

## Critical Files

| File | Action |
|------|--------|
| `/services/holdings/database/database.go` | Rewrite: Postgres → SQLite |
| `/services/holdings/database/schema.sql` | New: SQLite schema (tables created on startup) |
| `/services/holdings/database/queries.sql` | Update: sqlc queries for SQLite |
| `/services/holdings/server/server.go` | Rewrite: Connect RPC → chi REST |
| `/services/holdings/server/router.go` | New: chi router with REST endpoints |
| `/services/holdings/service/holdings.go` | Update: handlers for REST routes |
| `/services/holdings/plaid/` | Delete: Plaid integration removed |
| `/services/holdings/service/plaid.go` | Delete: Plaid handlers removed |
| `/services/holdings/Makefile` | Update: remove proto.generate, update targets |
| `/services/holdings/build/*.Dockerfile` | Update: remove Postgres/Plaid deps |
| `/services/holdings/env.example` | Update: SQLite config only |
| `/services/portfolio/` | New: entire Python/FastAPI service |
| `/services/portfolio/main.py` | New: FastAPI app entry point |
| `/services/portfolio/services/market.py` | New: yfinance wrapper with cache |
| `/services/portfolio/services/cache.py` | New: SQLite daily price cache |
| `/services/portfolio/services/holdings_client.py` | New: HTTP client for Holdings |
| `/services/portfolio/Makefile` | New: dev/run targets |
| `/services/portfolio/build/dev.Dockerfile` | New: Python/uv container |
| `/services/client/package.json` | Update: add tanstack-query, orval |
| `/services/client/orval.config.ts` | New: OpenAPI client generation config |
| `/services/client/lib/api/` | Generated: typed API client |
| `/services/client/app/page.tsx` | Update: use generated hooks instead of mock data |
| `/api/monay/v1beta1/` | Archive: protobuf files no longer needed |

## Implementation Steps

### Phase 0: Holdings Service SQLite Migration

> Fresh start - no data migration from Postgres. Re-import from CSVs after migration.

#### 0.0 Export Verified Data as Reference

Before migration, snapshot current Postgres data as CSV fixtures:

```bash
# Export to fixtures directory
psql $DATABASE_URL -c "COPY accounts TO STDOUT WITH CSV HEADER" > fixtures/accounts.csv
psql $DATABASE_URL -c "COPY holdings TO STDOUT WITH CSV HEADER" > fixtures/holdings.csv
psql $DATABASE_URL -c "COPY lots TO STDOUT WITH CSV HEADER" > fixtures/lots.csv
psql $DATABASE_URL -c "COPY transactions TO STDOUT WITH CSV HEADER" > fixtures/transactions.csv
psql $DATABASE_URL -c "COPY cash_balances TO STDOUT WITH CSV HEADER" > fixtures/cash_balances.csv
```

Store in `/services/holdings/fixtures/` - verified against broker screenshots, useful for regression testing.

#### 0.1 Update sqlc Config

Update `sqlc.yaml` to target SQLite:
```yaml
version: "2"
sql:
  - engine: "sqlite"
    queries: "database/queries.sql"
    schema: "database/schema.sql"
    gen:
      go:
        package: "database"
        out: "database"
```

#### 0.2 Create SQLite Schema

Create `database/schema.sql`:
```sql
CREATE TABLE IF NOT EXISTS accounts (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    institution TEXT,
    type TEXT,
    subtype TEXT,
    created_at TEXT DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS holdings (
    id TEXT PRIMARY KEY,
    account_id TEXT NOT NULL REFERENCES accounts(id),
    symbol TEXT NOT NULL,
    name TEXT,
    quantity REAL NOT NULL,
    price REAL,
    value REAL,
    UNIQUE(account_id, symbol)
);

CREATE TABLE IF NOT EXISTS lots (
    id TEXT PRIMARY KEY,
    account_id TEXT NOT NULL REFERENCES accounts(id),
    symbol TEXT NOT NULL,
    quantity REAL NOT NULL,
    cost_basis REAL NOT NULL,
    acquired_date TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS transactions (
    id TEXT PRIMARY KEY,
    account_id TEXT NOT NULL REFERENCES accounts(id),
    symbol TEXT,
    type TEXT NOT NULL,
    quantity REAL,
    price REAL,
    amount REAL,
    date TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS cash_balances (
    account_id TEXT PRIMARY KEY REFERENCES accounts(id),
    balance REAL NOT NULL,
    as_of_date TEXT NOT NULL
);
```

#### 0.3 Update Database Connection

```bash
go get modernc.org/sqlite
```

Rewrite `database/database.go`:
- Replace `pgx` with `modernc.org/sqlite` (pure Go, no CGO required)
- Open DB with `sql.Open("sqlite", "./holdings.db")`
- Create tables on startup by executing schema.sql
- sqlc generates code compatible with `database/sql` interface

#### 0.4 Update Queries

Update `database/queries.sql` for SQLite syntax:
- Replace `$1, $2` with `?, ?` or `@param` syntax
- Replace `RETURNING *` patterns (SQLite support varies)
- Run `sqlc generate`

#### 0.5 Remove Postgres Dependencies

- Remove `github.com/jackc/pgx/v5` from go.mod
- Remove goose migrations (tables created on startup)
- Delete `database/sql/migrations/`

### Phase 1: Holdings Service REST Migration

#### 1.1 Add Dependencies

```bash
cd /services/holdings
go get github.com/go-chi/chi/v5 github.com/go-chi/cors
```

#### 1.2 Create chi Router

Create `server/router.go`:
- Mount routes under `/api/v1/`
- Add CORS middleware for local dev
- Add request logging middleware

#### 1.3 Implement REST Endpoints

Update `service/` to expose handlers as HTTP handlers instead of Connect RPC:

- `GET /api/v1/accounts` - List all accounts
- `GET /api/v1/accounts/{id}` - Get single account
- `GET /api/v1/holdings` - List holdings (optional `?account_id=` filter)

#### 1.4 Update Server Entry Point

Rewrite `server/server.go`:
- Remove Connect RPC imports and handlers
- Use chi router instead of http.ServeMux
- Keep h2c for HTTP/2 support

#### 1.5 Remove Connect RPC and Plaid

- Remove `connectrpc.com/connect` from go.mod
- Remove `connectrpc.com/grpchealth` and `grpcreflect`
- Remove `github.com/plaid/plaid-go/v29` from go.mod
- Delete generated protobuf code in `gen/`
- Delete `/services/holdings/plaid/` directory
- Delete `/services/holdings/service/plaid.go`
- Archive `/api/monay/v1beta1/*.proto` files

#### 1.6 Update Build Configs

Update `/services/holdings/Makefile`:
- Remove `proto.generate` target (no more protobufs)
- Update `run`/`dev` targets if needed

Update `/services/holdings/build/`:
- `dev.Dockerfile` - remove Postgres client, Plaid env vars
- `ci.Dockerfile` - same
- `.air.toml` - update if paths changed

Update `/services/holdings/env.example`:
- Remove `MONAY_HOLDINGS_PLAID_*` vars
- Remove `MONAY_HOLDINGS_POSTGRES_*` vars
- Add `MONAY_HOLDINGS_DB_PATH=./holdings.db`

### Phase 2: Portfolio Service (Python/FastAPI)

#### 2.1 Create Directory Structure

```
/services/portfolio/
├── main.py
├── config.py
├── pyproject.toml
├── routers/
│   ├── __init__.py
│   ├── portfolio.py
│   ├── quotes.py
│   └── chart.py
├── services/
│   ├── __init__.py
│   ├── market.py
│   └── holdings_client.py
└── models/
    ├── __init__.py
    ├── portfolio.py
    ├── quote.py
    └── chart.py
```

#### 2.2 Initialize with uv

```bash
cd /services/portfolio
uv init
uv add fastapi uvicorn[standard] yfinance httpx pydantic-settings aiosqlite
```

Dependencies:
- `fastapi` - Web framework
- `uvicorn[standard]` - ASGI server
- `yfinance` - Yahoo Finance API
- `httpx` - Async HTTP client
- `pydantic-settings` - Configuration
- `aiosqlite` - Async SQLite for daily price cache

#### 2.3 Implement Config

Create `config.py` with pydantic-settings:
- `MONAY_PORTFOLIO_LISTEN_ADDR` (default: `:8889`)
- `MONAY_PORTFOLIO_HOLDINGS_URL` (default: `http://localhost:8888`)
- `MONAY_PORTFOLIO_CACHE_PATH` (default: `./cache.db`)

#### 2.4 Implement Holdings Client

Create `services/holdings_client.py`:
- `get_accounts()` - Fetch accounts from Holdings Service
- `get_holdings(account_id: str | None)` - Fetch holdings

#### 2.5 Implement Daily Price Cache

Create `services/cache.py`:
- Create table on startup:
  ```sql
  CREATE TABLE IF NOT EXISTS daily_prices (
      symbol TEXT NOT NULL,
      date TEXT NOT NULL,
      open REAL, high REAL, low REAL, close REAL, volume INTEGER,
      PRIMARY KEY (symbol, date)
  );
  ```
- `get_daily_prices(symbol, start_date, end_date)` - Query cached data
- `store_daily_prices(symbol, points)` - Insert/replace price points

#### 2.6 Implement Market Service

Create `services/market.py`:
- `get_quotes(symbols: list[str])` - Fetch current prices via yfinance (always live)
- `get_daily_chart(symbol: str, start_date, end_date)` - Cache-first fetch:
  1. Query cache for existing daily data
  2. Identify missing dates
  3. Fetch missing from yfinance
  4. Store new data in cache
  5. Return combined result
- `get_intraday_chart(symbol: str, range: str)` - Fetch live from yfinance (no cache)

#### 2.7 Implement Routers

Create routers per [portfolio.md](./portfolio.md) API spec:
- `routers/quotes.py` - `GET /api/v1/quotes`
- `routers/chart.py` - `GET /api/v1/chart/{symbol}`
- `routers/portfolio.py` - `GET /api/v1/portfolio/enriched`

#### 2.8 Create Main App

Create `main.py`:
- Mount routers
- Add CORS middleware
- Expose `/openapi.json` (automatic with FastAPI)

#### 2.9 Create Build Configs

Create `/services/portfolio/Makefile`:
```makefile
.PHONY: dev run

dev:
	uv run uvicorn main:app --reload --port 8889

run:
	uv run uvicorn main:app --port 8889
```

Create `/services/portfolio/build/dev.Dockerfile`:
```dockerfile
FROM python:3.12-slim
WORKDIR /app
RUN pip install uv
COPY pyproject.toml uv.lock ./
RUN uv sync
COPY . .
CMD ["uv", "run", "uvicorn", "main:app", "--host", "0.0.0.0", "--port", "8889"]
```

Create `/services/portfolio/.env.example`:
```
MONAY_PORTFOLIO_LISTEN_ADDR=:8889
MONAY_PORTFOLIO_HOLDINGS_URL=http://localhost:8888
MONAY_PORTFOLIO_CACHE_PATH=./cache.db
```

### Phase 3: Frontend Integration

#### 3.1 Add Dependencies

```bash
cd /services/client
pnpm add @tanstack/react-query
pnpm add -D orval
```

#### 3.2 Create Orval Config

Create `orval.config.ts`:
```typescript
export default {
  portfolio: {
    input: 'http://localhost:8889/openapi.json',
    output: {
      target: './lib/api/portfolio.ts',
      client: 'react-query',
      mode: 'tags-split',
    },
  },
};
```

#### 3.3 Generate API Client

```bash
pnpm orval
```

This generates typed hooks in `lib/api/portfolio.ts`.

#### 3.4 Add Query Provider

Update `app/layout.tsx`:
- Wrap app with `QueryClientProvider`

#### 3.5 Update Dashboard Page

Update `app/page.tsx`:
- Import generated hooks (`useGetPortfolioEnriched`, `useGetQuotes`, `useGetChart`)
- Replace mock data with query results
- Add loading/error states

#### 3.6 Wire Components

Update components to use real data:
- Holdings table uses `enrichedPortfolio.holdings`
- Stats cards use `enrichedPortfolio.summary`
- Chart panel uses `useGetChart` hook

### Phase 4: Verification

#### 4.1 Start Services

```bash
# Terminal 1: Holdings Service
cd /services/holdings && go run ./cmd/main.go serve

# Terminal 2: Portfolio Service
cd /services/portfolio && uv run uvicorn main:app --port 8889

# Terminal 3: Frontend
cd /services/client && pnpm dev
```

#### 4.2 Verify Endpoints

```bash
# Holdings Service
curl http://localhost:8888/api/v1/accounts
curl http://localhost:8888/api/v1/holdings

# Portfolio Service
curl http://localhost:8889/api/v1/quotes?symbols=AAPL,MSFT
curl http://localhost:8889/api/v1/chart/AAPL?range=1y
curl http://localhost:8889/api/v1/portfolio/enriched
```

#### 4.3 Verify Frontend

Open http://localhost:7777 and verify:
- [ ] Holdings table shows real prices from yfinance
- [ ] Day change indicators reflect actual market movement
- [ ] Clicking a holding opens detail panel with live chart
- [ ] Portfolio summary reflects aggregated values

### Phase 5: Visual Regression Testing (MCP/Playwright)

#### 5.1 Baseline Screenshots

Before migration, capture baseline screenshots from the postgres-backed version:
- Store in `services/client/tests/screenshots/baseline/`
- `dashboard-full.png` - Full dashboard view
- `holdings-table.png` - Holdings table section
- `portfolio-chart.png` - Portfolio chart section

#### 5.2 Create Playwright Test Script

Create `services/client/tests/visual-regression.spec.ts`:
```typescript
import { test, expect } from '@playwright/test';

test.describe('Visual Regression', () => {
  test('dashboard matches baseline', async ({ page }) => {
    await page.goto('http://localhost:7777');
    await page.waitForSelector('[data-testid="holdings-table"]');

    // Compare full dashboard
    await expect(page).toHaveScreenshot('dashboard-full.png', {
      maxDiffPixelRatio: 0.1, // Allow 10% diff for dynamic data
    });
  });

  test('holdings table structure matches', async ({ page }) => {
    await page.goto('http://localhost:7777');
    await page.waitForSelector('[data-testid="holdings-table"]');

    const table = page.locator('[data-testid="holdings-table"]');
    await expect(table).toHaveScreenshot('holdings-table.png', {
      maxDiffPixelRatio: 0.1,
    });
  });
});
```

#### 5.3 Run Visual Comparison via MCP

Use Playwright MCP to:
1. Navigate to http://localhost:7777
2. Wait for data to load
3. Take screenshot of current state
4. Compare against baseline in `tests/screenshots/baseline/`
5. Report differences

```bash
# Run playwright tests
cd services/client && pnpm playwright test visual-regression

# Or use MCP directly for ad-hoc comparison
# browser_navigate -> browser_snapshot -> browser_take_screenshot
```

#### 5.4 Acceptance Criteria

- [ ] Dashboard layout unchanged (structure, sections)
- [ ] Holdings table columns match (symbol, shares, price, value, gain/loss)
- [ ] Portfolio chart renders with data
- [ ] No missing UI elements from postgres version
- [ ] Dynamic values (prices, changes) can differ but structure must match

### Phase 6: Data Reimport & Sanity Check

#### 6.1 Locate Original Data Sources

Check for original CSV files from brokerages:
```bash
# Common locations
ls ~/Downloads/*etrade*.csv
ls ~/Downloads/*fidelity*.csv
ls ~/Documents/finances/*.csv
```

Or check if old Postgres is still accessible:
```bash
# If postgres container still exists
docker ps -a | grep postgres
```

#### 6.2 Reimport Holdings Data

Using the import scripts in `services/holdings/scripts/`:

```bash
cd services/holdings

# Import E*Trade account
./scripts/import-account.sh /path/to/etrade-holdings.csv

# Import transactions if available
./scripts/import-transactions.sh /path/to/etrade-transactions.csv
```

#### 6.3 Verify Data Loaded

```bash
# Check accounts
curl http://localhost:8888/api/v1/accounts

# Check holdings
curl http://localhost:8888/api/v1/holdings

# Should see accounts and holdings matching the original data
```

#### 6.4 Screenshot Sanity Check vs E*Trade

Use MCP Playwright to compare dashboard against E*Trade checkpoint:

1. Navigate to dashboard: `http://localhost:7777`
2. Take screenshot of holdings table
3. Compare against `etrade-checkpoint.png` in repo root

**Comparison checklist (doesn't need to be 100% match):**
- [ ] Same accounts appear (Stock Plan, Brokerage, etc.)
- [ ] Holdings symbols match (AAPL, MSFT, etc.)
- [ ] Share quantities approximately match
- [ ] Total values in same ballpark (exact prices will differ due to market)
- [ ] No missing accounts or holdings

```bash
# MCP commands for comparison:
# 1. browser_navigate to http://localhost:7777
# 2. browser_snapshot to see current state
# 3. browser_take_screenshot filename=dashboard-current.png
# 4. Compare dashboard-current.png vs etrade-checkpoint.png
```

#### 6.5 Acceptance Criteria

- [ ] All accounts from E*Trade checkpoint appear in dashboard
- [ ] Holdings count matches (same number of positions)
- [ ] No data loss from migration
- [ ] Portfolio totals within reasonable range (market fluctuation OK)

## Account Inventory

All accounts that should be loaded and visible in the dashboard:

### E*Trade Accounts (6)
| Account Name | Type | Transactions | Notes |
|--------------|------|--------------|-------|
| Employee Stock Plan RSU | etrade | 1 | RSU grants |
| Stock Plan ESPP | etrade | 1 | Employee stock purchase |
| Stock Plan DRIP | etrade | 1 | Dividend reinvestment |
| E*Trade Joint-2060 | etrade | 223 | Joint brokerage |
| E*Trade Joint-3652 | etrade | 51 | Joint brokerage |
| E*Trade Maya-3758 | etrade | 147 | Individual brokerage |

### LPL Accounts (2)
| Account Name | Type | Transactions | Notes |
|--------------|------|--------------|-------|
| LPL Bond-5516 | lpl | 342 | Bond portfolio |
| LPL Equities-4015 | lpl | 322 | Equity portfolio |

### Merrill Accounts (2)
| Account Name | Type | Transactions | Notes |
|--------------|------|--------------|-------|
| Merrill Managed-2241 | merrill | 1275 | Managed account |
| Merrill Other-0282 | merrill | 55 | Other brokerage |

**Total: 10 accounts, 2,418 transactions**

---

### FINAL ACCEPTANCE TEST

**This is the exit condition for Ralph. Do not declare ALL_DONE until this passes.**

Using Playwright MCP, verify the dashboard matches etrade-checkpoint.png:

```
1. browser_navigate to http://localhost:7777
2. browser_snapshot - verify page loads without errors
3. browser_console_messages - no errors
4. browser_take_screenshot filename=final-dashboard.png
5. Compare against etrade-checkpoint.png in repo root
```

#### E*Trade Verification (Primary)

Compare dashboard against `etrade-checkpoint.png`. The E*Trade accounts are the primary validation target:

**Account Presence (must match E*Trade screenshot):**
- [ ] Employee Stock Plan RSU visible with holdings
- [ ] Stock Plan ESPP visible with holdings
- [ ] Stock Plan DRIP visible with holdings
- [ ] E*Trade Joint-2060 visible with holdings
- [ ] E*Trade Joint-3652 visible with holdings
- [ ] E*Trade Maya-3758 visible with holdings

**Holdings Accuracy (spot check against E*Trade):**

Stock Plan accounts (all GILD):
- [ ] Employee Stock Plan RSU: 3,248 shares GILD (~$394k cost basis)
- [ ] Stock Plan DRIP: 150 shares GILD (~$18k cost basis)
- [ ] Stock Plan ESPP: 309 shares GILD (~$37.5k cost basis)

E*Trade Joint-2060 holdings should include:
- [ ] VFIAX, VGHCX (Vanguard funds)
- [ ] LLY, LMT, META, TSLA, ESLT (individual stocks)

E*Trade Joint-3652:
- [ ] GILD holdings (transferred from stock plans)

E*Trade Maya-3758 holdings should include:
- [ ] Biotech: ALLO, CRSP, DVAX, KURA, PRVB
- [ ] Tech: EA, ZS, UBER, RBLX, APP, AMZN
- [ ] VIG (Vanguard dividend fund)

General:
- [ ] No phantom holdings (holdings not in E*Trade)
- [ ] No missing holdings (holdings in E*Trade but not dashboard)

**Value Sanity:**
- [ ] Total portfolio value within 10% of E*Trade total (market fluctuation OK)
- [ ] Individual position values calculated correctly (shares × current price)
- [ ] Cost basis preserved from import (if displayed)

#### All Accounts Verification (Secondary)

**LPL Accounts:**
- [ ] LPL Bond-5516: Municipal bonds (CUSIPs like 870462SA7, 13048T5L3)
- [ ] LPL Equities-4015 should include: NVDA, AAPL, GOOG, QQQJ, SOXX, XLY, MGM

**Merrill Accounts:**
- [ ] Merrill Managed-2241 should include: AAPL, NVDA, META, LLY, MSFT, GOOGL, JPM, WMT, V
- [ ] Merrill Other-0282 visible with holdings

**Must match:**
- [ ] All 10 accounts visible in dashboard
- [ ] All 6 E*Trade accounts show correct holdings
- [ ] Both LPL accounts show bond/equity positions
- [ ] Both Merrill accounts show holdings
- [ ] Holdings show real market prices (not $0 or mock data)
- [ ] Charts render with actual data
- [ ] No console errors, no loading spinners stuck

#### API Verification

```bash
# Verify all accounts returned
curl -s http://localhost:8888/api/v1/accounts | jq length
# Expected: 10

# Verify holdings exist for E*Trade accounts
curl -s http://localhost:8888/api/v1/holdings | jq 'group_by(.account_id) | length'
# Expected: accounts with holdings

# Verify lots computed correctly (for tax lot tracking)
curl -s http://localhost:8888/api/v1/lots | jq length
# Expected: > 0
```