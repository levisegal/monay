# Portfolio Feature Spec

## Overview

Real market data integration for the Monay portfolio dashboard. Three components:

1. **Holdings Service (Go/chi)** - Internal service for accounts and holdings data
2. **Portfolio Service (Python/FastAPI)** - Gateway that aggregates holdings with live market data
3. **Frontend (Next.js)** - Dashboard with TanStack Query and generated API client

## Architecture

```
Client (Next.js :7777)
    → TanStack Query + orval-generated client
        → Portfolio Service (Python/FastAPI :8889)
            → REST → Holdings Service (Go/chi :8888, internal)
            → yfinance → Yahoo Finance
```

## API Design

### Portfolio Service (gateway, :8889)

**GET /api/v1/portfolio/enriched**

Returns holdings enriched with live market data.

```json
{
  "summary": {
    "total_value": 125000.00,
    "total_cost_basis": 100000.00,
    "total_unrealized_gain": 25000.00,
    "day_change": 1500.00,
    "day_change_percent": 1.21,
    "holding_count": 15
  },
  "holdings": [
    {
      "account_id": "acc_123",
      "symbol": "AAPL",
      "name": "Apple Inc.",
      "quantity": 100,
      "cost_basis": 15000.00,
      "current_price": 185.50,
      "market_value": 18550.00,
      "unrealized_gain": 3550.00,
      "unrealized_gain_percent": 23.67,
      "day_change": 230.00,
      "allocation_percent": 14.84,
      "asset_type": "equity"
    }
  ]
}
```

**GET /api/v1/quotes?symbols=AAPL,GOOGL**

Returns current quotes for symbols.

```json
{
  "quotes": [
    {
      "symbol": "AAPL",
      "name": "Apple Inc.",
      "price": 185.50,
      "change": 2.30,
      "change_percent": 1.25,
      "previous_close": 183.20,
      "volume": 45000000,
      "asset_type": "equity"
    }
  ]
}
```

**GET /api/v1/chart/{symbol}?range=1y&interval=1d**

Returns historical chart data.

```json
{
  "symbol": "AAPL",
  "points": [
    {
      "timestamp": "2024-01-17",
      "open": 180.00,
      "high": 182.00,
      "low": 179.50,
      "close": 181.25,
      "volume": 50000000
    }
  ]
}
```

### Holdings Service (internal, :8888)

**GET /api/v1/accounts**
```json
{
  "accounts": [
    {
      "id": "acc_123",
      "name": "Individual Brokerage",
      "institution": "E*Trade",
      "type": "brokerage"
    }
  ]
}
```

**GET /api/v1/holdings?account_id=acc_123**
```json
{
  "holdings": [
    {
      "account_id": "acc_123",
      "symbol": "AAPL",
      "name": "Apple Inc.",
      "quantity": 100,
      "cost_basis": 15000.00
    }
  ]
}
```

Data populated via CSV import (CLI), not Plaid.

## Database

Both services use SQLite for simplicity (no Postgres to run, no migrations).

### Holdings Service

SQLite file at `holdings.db`. Tables created on startup:
- `accounts` - Linked brokerage accounts
- `holdings` - Current positions
- `lots` - Tax lots with cost basis
- `transactions` - Trade history
- `cash_balances` - Cash tracking

sqlc works with SQLite - keep existing query patterns.

### Portfolio Service

SQLite file at `cache.db`. Tables created on startup:
- `daily_prices` - Daily OHLCV data (immutable, cached forever)

```sql
CREATE TABLE IF NOT EXISTS daily_prices (
    symbol TEXT NOT NULL,
    date TEXT NOT NULL,
    open REAL,
    high REAL,
    low REAL,
    close REAL,
    volume INTEGER,
    PRIMARY KEY (symbol, date)
);
```

**Caching strategy:**
- **Daily data**: Cached forever (immutable historical prices)
- **Intraday data**: Fetched live from yfinance (no cache, recent data is fast)

Cache-first fetch for daily: check cache → fetch missing dates from yfinance → store.

## Configuration

### Holdings Service
```
MONAY_HOLDINGS_LISTEN_ADDR=:8888
MONAY_HOLDINGS_DB_PATH=./holdings.db
```

### Portfolio Service
```
MONAY_PORTFOLIO_LISTEN_ADDR=:8889
MONAY_PORTFOLIO_HOLDINGS_URL=http://localhost:8888
MONAY_PORTFOLIO_CACHE_PATH=./cache.db
```
