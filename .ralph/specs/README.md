# Specs Directory

Lookup table for feature specifications. Read descriptions below to find the right spec, then drill into the linked files for details.

## Pattern

- `<feature>.md` - Feature spec (architecture, API design, configuration)
- `<feature>-implementation-plan.md` - Implementation plan (steps, tasks, critical files)

## Specifications

| Feature | Description | Spec |
|---------|-------------|------|
| Holdings | Investment data service. Plaid account linking, CSV importers (E*Trade, Fidelity, Schwab, Vanguard, LPL, Merrill), tax lot tracking, cash balance tracking. CLI and Connect RPC API. | [holdings.md](./holdings.md) |
| Portfolio | Market data integration. SQLite for both services (Holdings uses sqlc, Portfolio uses aiosqlite for daily price cache). New Python/FastAPI gateway (uv) enriches holdings with live yfinance prices. Intraday fetched live, daily cached forever. Holdings migrates Connect RPC → chi REST. Frontend wired with TanStack Query + orval. | [portfolio.md](./portfolio.md) |
