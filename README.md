# Monay

Personal wealth intelligence platform for families. Aggregate accounts, ask questions in plain English, and manage portfolios across multiple households and institutions.

## Overview

Monay unifies investment data from multiple brokerage accounts and family members into a single queryable system. Instead of logging into 6 different portals, ask:

- *"What's my total net worth across all accounts?"*
- *"How much did my parents' portfolio gain today?"*
- *"What's my asset allocation vs target?"*
- *"What trades would rebalance in-laws back to 60/40?"*

The main interface is a **Next.js dashboard** showing holdings, performance, and allocation. A **conversational AI popup** lets you ask questions in natural language—think ChatGPT sidebar but backed by your real portfolio data.

## Philosophy

**This is personal software.** Built for one family, never meant to scale. This unlocks simplicity:

- **Home server + Tailscale** — docker-compose on a home machine, access via Tailscale
- **Postgres in a container** — No managed database needed
- **No auth** — Private network only; no OAuth/OIDC complexity
- **Hardcoded config** — Family names, account mappings can live in code
- **Automate first** — But manual fallbacks are fine for 3 households
- **Opinionated** — Optimize for your brokerages, your asset classes, your needs

If it works on your home server, it's production.

**Operational constraints:**
- Read-only (no automated trading)
- Manual OTP flows (family texts you codes when needed)
- CSV fallback for institutions with poor API support

## Features

| Feature | Description |
|---------|-------------|
| **Account Aggregation** | Plaid + CSV imports from Schwab, ETrade, Fidelity, LPL |
| **Conversational Queries** | Ask questions about holdings, performance, allocation |
| **Daily Snapshots** | Track portfolio changes over time |
| **Rebalancing** | Generate trade recommendations to hit target allocations |
| **Multi-Household** | Separate views for your money, parents, in-laws |
| **Household Filter** | Single-household view for screen sharing with family |
| **Performance Tracking** | Daily/weekly/monthly gains across accounts |
| **Live Prices** | Current quotes via market data API (replaces checking Schwab) |

## Target Institutions

| Institution | Integration | Priority |
|-------------|-------------|----------|
| Schwab | Plaid + CSV | High |
| E*Trade | Plaid + CSV | High |
| Fidelity | Plaid + CSV | Medium |
| LPL | CSV only | Medium |

## Architecture

```
┌─────────────────────────────────────────────────────────────────────┐
│                     monay-client (Next.js)                           │
│                                                                      │
│  ┌────────────────────────────────────┐  ┌───────────────────────┐  │
│  │  Dashboard                         │  │  AI Chat Popup        │  │
│  │  - Holdings table                  │  │  - Natural language   │  │
│  │  - Allocation pie chart            │  │  - Ask anything       │  │
│  │  - Daily P&L                       │  │  - Backed by MCP      │  │
│  │  - CSV upload                      │  │                       │  │
│  └────────────────────────────────────┘  └───────────────────────┘  │
└─────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────┐
│                     Go Services (Connect RPC)                        │
│                                                                      │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  ┌─────────┐  │
│  │  holdings    │  │  quotes      │  │  rebalancing │  │  mcp    │  │
│  └──────────────┘  └──────────────┘  └──────────────┘  └─────────┘  │
└─────────────────────────────────────────────────────────────────────┘
              │                │                │
              ▼                ▼                ▼
┌──────────────┐ ┌──────────────┐ ┌──────────────────────────────────┐
│ Plaid        │ │ CSV Importer │ │ Market Data (Yahoo Finance)      │
└──────────────┘ └──────────────┘ └──────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────┐
│                     PostgreSQL (RDS)                                 │
│                                                                      │
│  households │ accounts │ holdings │ snapshots │ target_allocations  │
└─────────────────────────────────────────────────────────────────────┘
```

### Data Model

```
households
├── id, name (e.g., "Levi", "Parents", "In-Laws")
│
├── accounts
│   ├── id, household_id, institution, account_number, account_type
│   ├── plaid_item_id (nullable - for Plaid-linked accounts)
│   │
│   └── holdings (point-in-time positions)
│       ├── id, account_id, symbol, quantity, price, market_value
│       └── as_of_date
│
├── target_allocations
│   ├── id, household_id, asset_class, target_pct
│   └── (e.g., "US Equity" → 60%, "Bonds" → 30%, "International" → 10%)
│
└── snapshots (daily portfolio state)
    ├── id, household_id, snapshot_date
    └── total_value, allocation_json
```

### Ingestion Flow

1. **Plaid-linked accounts**: Scheduler runs daily, pulls holdings via Plaid Investment APIs
2. **CSV uploads**: User uploads brokerage exports, parser normalizes to common schema
3. **Deduplication**: Holdings keyed by (account_id, symbol, as_of_date)

### Market Data

For live prices, use a free API. Options ranked by simplicity:

| Provider | Free Tier | Notes |
|----------|-----------|-------|
| **Yahoo Finance** | Unlimited (unofficial) | No API key, just works |
| **Alpha Vantage** | 25 req/day | Free API key, reliable |
| **Finnhub** | 60 req/min | Free API key |
| **Polygon.io** | 5 req/min | Free tier available |

For personal use, **Yahoo Finance** via an unofficial Go library is the path of least resistance—no signup, no key management. Cache quotes for 15 minutes to avoid hammering.

```
get_quotes(symbols []string) → map[symbol]Quote
  - price, change, change_pct, market_cap
  - cached in memory, refreshed on demand
```

### Rebalancing Engine

```
Input:
  - Current holdings across all accounts for a household
  - Target allocation (asset class → percentage)
  - Security → asset class mapping

Output:
  - Current vs target allocation diff
  - Suggested trades to rebalance (buy X, sell Y)
  - Tax-lot aware recommendations (optional)
```

The engine is **advisory only**—generates recommendations that you execute manually in each brokerage.

### Conversational Agent (MCP)

A popup/sidebar in the dashboard lets you ask questions in natural language. An MCP server exposes portfolio data as tools, powering the chat via Claude or a custom LLM integration.

```
Tools:
  - query_holdings(household?, account?, symbol?)
  - get_quotes(symbols)                 # live prices
  - get_portfolio_value(household?)     # holdings × current prices
  - get_daily_pnl(household?)           # today's gain/loss
  - get_allocation(household)
  - get_performance(household, period)  # day, week, month, ytd
  - calculate_rebalance(household)
  - compare_snapshots(household, date1, date2)
  - get_net_worth(household?)
```

**Example conversations:**

> "How much money do I have total?"
→ Queries all households, sums market values

> "What's my parents' biggest position?"
→ Filters holdings by household, sorts by market value

> "How much did we make today across everyone?"
→ Compares today's snapshot to yesterday's for all households

> "What would it take to get in-laws to 70/30 stocks/bonds?"
→ Runs rebalancing engine with custom target

> "Show me all the tech stocks across all accounts"
→ Filters by sector/asset class

## Project Structure

```
monay/
├── api/
│   └── monay/v1/             # Protobuf definitions
├── build/
│   ├── compose.yml
│   └── Makefile
├── kubernetes/               # K8s manifests (kustomize overlays)
├── terraform/
│   └── aws/
└── services/
    ├── holdings/             # Portfolio data, CSV import, Plaid sync
    │   ├── cmd/
    │   ├── csv/
    │   ├── plaid/
    │   └── database/
    ├── quotes/               # Market data service
    │   └── cmd/
    ├── rebalancing/          # Rebalancing engine
    │   └── cmd/
    ├── mcp/                  # MCP server for AI agent
    │   └── cmd/
    └── client/               # Next.js frontend
        ├── app/
        ├── components/
        └── package.json
```

## Infrastructure

### Local + Tailscale (default)

Run on a home server or always-on laptop, access via Tailscale:

```
docker-compose up     # Postgres, Go services, Next.js
```

- Access via Tailscale magic DNS (e.g., `monay.tailnet-name.ts.net`)
- Free, no cloud costs
- Good enough for personal/family use indefinitely

### AWS + Kubernetes (optional)

If you want "real" hosting (~$100/mo):

- **Compute**: EKS (Karpenter for nodes)
- **Database**: RDS Postgres
- **Ingress**: Kong + external-dns + cert-manager
- **Secrets**: secrets-store-csi-driver

## Non-Goals

- Automated trade execution
- Real-time streaming prices
- Multi-currency support
- Multi-tenancy / user management
- Brokerage API integrations beyond Plaid
- Complex auth (OAuth, OIDC, RBAC)

