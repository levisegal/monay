# CLAUDE.md

This file provides guidance to Claude Code when working with code in this repository.

## Development Commands

### Build Commands
```bash
# Build holdings for current platform (macOS ARM64)
cd services/holdings && make build.darwin-arm64

# Build all platforms
make build

# Build specific platforms
make build.linux-arm64
make build.linux-amd64

# Generate sqlc code (after SQL query/schema changes)
cd services/holdings && make db.generate

# Docker operations (from build/)
make build          # build all services
make up             # start all services
make down           # stop all services
make build.service SERVICE=holdings  # build single service

# Run services directly
cd services/holdings && make run       # Go REST server on :8888
cd services/portfolio && make dev      # Python FastAPI on :8889
```

### Key Services

| Service | Language | Port | Role |
|---------|----------|------|------|
| **holdings** | Go (chi) | 8888 | Accounts, positions, lots, transactions |
| **portfolio** | Python (FastAPI) | 8889 | Market data, quotes, OHLCV cache |
| **client** | Next.js | 7777 | Dashboard UI (TanStack Query + orval) |

### Directory Structure

```
monay/
├── build/              # Docker compose files, top-level Makefile
├── services/
│   ├── holdings/       # Go REST API + SQLite
│   │   ├── cmd/        # Cobra CLI entrypoint
│   │   ├── config/     # Env-based config (caarlos0/env)
│   │   ├── database/   # SQLite schema + sqlc queries
│   │   ├── gen/db/     # sqlc-generated Go code
│   │   ├── server/     # chi router + handlers
│   │   └── build/      # Dockerfiles
│   ├── portfolio/      # Python FastAPI + SQLite cache
│   │   ├── routers/    # FastAPI route handlers
│   │   ├── services/   # Market data + cache logic
│   │   ├── models/     # Pydantic models
│   │   └── build/      # Dockerfiles
│   └── client/         # Next.js dashboard
│       ├── components/ # React components
│       └── gen/        # orval-generated API client
└── specs/              # Implementation plans
```

## Database

All services use **SQLite** (no PostgreSQL).

### Holdings (Go)
- Driver: `modernc.org/sqlite` (pure Go, no CGO required)
- Schema: embedded via `//go:embed` in `database/sql/schema.sql`, applied on startup
- Queries: sqlc with SQLite engine, named `@param` syntax
- Tables: accounts, securities, positions, transactions, lots, lot_dispositions, cash_transactions
- Money: stored as `integer` micros (1,000,000 micros = $1.00)

### Portfolio (Python)
- Driver: `aiosqlite` (async)
- Cache tables: daily_prices, quote_cache
- Caches yfinance market data to avoid repeated API calls

### Common Workflows
1. **Schema Changes**: Edit `services/holdings/database/sql/schema.sql`, then `cd services/holdings && make db.generate`
2. **Query Changes**: Edit files in `database/sql/queries/`, then `make db.generate`
3. **Local Testing**: `cd services/holdings && make run`

## REST API

Holdings service uses **chi** router with JSON responses (no protobuf/gRPC).

Endpoints:
- `GET /api/v1/health` - Health check
- `GET /api/v1/version` - Version info
- `GET /api/v1/accounts` - List accounts
- `GET /api/v1/accounts/{id}` - Get account
- `GET /api/v1/holdings` - List holdings (optional `?account_id=` filter)

## Development Prerequisites

1. **Go** (CGO not required — uses pure-Go SQLite driver)
2. **direnv** for environment variable management (no dotenv libraries)
3. **Python 3 + uv** for portfolio service
4. **Docker** with buildx for multi-platform builds
5. **sqlc** for Go query code generation

## Docker Build Pattern

Each service has a `build/` directory with:
- `dev.Dockerfile` - Development image with hot-reload (Air for Go, uvicorn --reload for Python)
- `ci.Dockerfile` - Production image for CI pipelines
- Docker Compose profiles: `monay-apis` (holdings + portfolio), `dashboard` (client)

## Go Coding Conventions

### CLI Structure
Use Cobra for all CLI commands:
- `cmd/main.go` at service root calls `cmd.Execute(ctx)`
- `cmd/cmd/root.go` defines root command and registers subcommands
- `cmd/cmd/<feature>.go` for each feature's subcommands

### Configuration Pattern
Each service has a `config/config.go` with:
- `Load()` singleton loader with `sync.Once`
- `NewConfig()` creates config by merging defaults with env vars
- `defaultConfig()` returns sensible defaults
- Uses `github.com/caarlos0/env/v11` for parsing with service prefix (e.g., `MONAY_`)
- Uses `dario.cat/mergo` to merge env config over defaults

### Happy Path / Early Return Pattern
Handle errors and edge cases with early returns to keep main logic flat:

```go
op, err := getOperation(ctx, id)
if err != nil {
    return nil, err
}

if op.Status == "FAILURE" {
    return nil, fmt.Errorf("operation failed")
}

return processData(op)
```

### No Generic Helpers or Utils
Don't create generic `helpers.go` or `utils.go` files. Keep code in domain-specific files. Duplication is acceptable — prefer clarity over DRY when it means keeping logic close to where it's used.

### No Inline Closures
Extract anonymous functions to package-level or method-level functions:

```go
// Bad
func (s *Service) UpsertItems(ctx context.Context, items []Item) {
    stringPtr := func(s string) *string {
        if s == "" { return nil }
        return &s
    }
    for _, item := range items {
        s.db.Insert(stringPtr(item.Name))
    }
}

// Good
func (s *Service) UpsertItems(ctx context.Context, items []Item) {
    for _, item := range items {
        s.db.Insert(stringPtr(item.Name))
    }
}

func stringPtr(s string) *string {
    if s == "" { return nil }
    return &s
}
```

### Functional Options Pattern
Use variadic functional options for configurable operations:

```go
type Option func(*Config)

func WithTimeout(d time.Duration) Option {
    return func(c *Config) { c.Timeout = d }
}

func DoThing(ctx context.Context, opts ...Option) {
    cfg := &Config{}
    for _, opt := range opts {
        opt(cfg)
    }
}
```

### No Function Passing in Structs
Use simple switches or explicit code instead of function fields for iteration.

## Go Testing Conventions

### External Package Tests
Use the `_test` package suffix to test only the public API.

### Golden File Tests
For data transformations that must match external implementations, use golden files:
1. Store real test data in `testdata/` directory
2. Name files by expected output
3. Document how expected values were computed

### What to Test
- **DO**: Pure data transformations, lookup/matching logic, business rules with clear inputs/outputs
- **DON'T**: HTTP call wiring, database insert/query mechanics, external service integration

## Code Style Preferences
- SQL queries (sqlc) should use named parameters (e.g., `@account_id`) instead of positional (`$1`, `$2`)
- Avoid generic type name suffixes like `Data`, `List`, `Info`, `Result` — prefer domain-specific names
- Helper functions placed **below** the functions that call them
- Comments should explain **why**, constraints, and non-obvious tradeoffs — not restate the obvious
- Scripts go into the `scripts/` directory in the associated service
- Always use structured logging (e.g., `slog.Info`, `slog.Debug`)
- When referencing an "account number" be clear if it's an apex account number

## SQL Style Preferences
- Use SQLFluff formatting: lowercase keywords, consistent indentation
- Trailing commas on multi-line SELECT lists
- One clause per line (SELECT, FROM, WHERE, GROUP BY, ORDER BY)
- Align columns in SELECT lists

## ID Generation
- Use prefixed KSUIDs for all primary keys (Stripe-style)
- Format: `{prefix}_{ksuid}` where prefix indicates resource type
- Examples: `acct_2OVFytmKTeZPMhvaU5g0EbHh7hn`, `sec_2OVAmoKVfUfq7BjyrlMTNaMxcJM`
- Store as `text` in SQLite
- Prefixes should be short (3-4 chars) and domain-specific
