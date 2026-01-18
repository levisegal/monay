# Holdings Service Spec

## Overview

Investment holdings aggregation service. Data imported via CSV exports from brokerages.

## Architecture

```
Holdings Service (Go :8888)
├── CSV Importers
│   ├── E*Trade
│   ├── Fidelity
│   ├── Schwab
│   ├── Vanguard
│   ├── LPL
│   └── Merrill Lynch
├── Tax Lot Tracking
│   └── Track cost basis per lot
└── Cash Balance Tracking
    └── Generate cash transactions from trades
```

## Components

### CSV Importers

Parse transaction history exports from brokerages:

| Broker | Export Format | Notes |
|--------|---------------|-------|
| E*Trade | Gains & Losses CSV | RSU/ESPP lots supported |
| Fidelity | Activity CSV | |
| Schwab | Transaction History CSV | |
| Vanguard | Transaction History CSV | |
| LPL | Positions/Transactions CSV | Bond positions |
| Merrill Lynch | Activity CSV | |

### Tax Lot Tracking

Track individual purchase lots for tax reporting:
- Cost basis per lot
- Acquisition date (short-term vs long-term)
- Wash sale adjustments

### Cash Balance Tracking

Generate cash transactions from trade activity:
- Set opening cash balance from statement
- Auto-generate cash flows from buys/sells/dividends
- View current cash balance and ledger

## CLI Commands

```bash
# Server
go run cmd/main.go serve              # Start API server

# Accounts
go run cmd/main.go accounts list      # List accounts
go run cmd/main.go accounts delete    # Delete account

# Holdings
go run cmd/main.go holdings list      # List holdings

# Import
go run cmd/main.go import             # Import from CSV

# Tax Lots
go run cmd/main.go lots list          # List tax lots
go run cmd/main.go lots process       # Process lots

# Cash
go run cmd/main.go cash balance       # View cash balance
go run cmd/main.go cash ledger        # View cash ledger
go run cmd/main.go cash set           # Set opening balance
go run cmd/main.go cash generate      # Generate cash transactions
```

## API (Connect RPC → REST migration pending)

Current API uses Connect RPC. Migration to chi REST planned in [portfolio-implementation-plan.md](./portfolio-implementation-plan.md).

**Planned endpoints** (REST):
- `GET /api/v1/accounts` - List accounts
- `GET /api/v1/accounts/{id}` - Get account
- `GET /api/v1/holdings` - List holdings (optional `?account_id=` filter)

## Configuration

```
MONAY_HOLDINGS_LISTEN_ADDR=:8888
MONAY_HOLDINGS_DB_PATH=./holdings.db
```

## Database

PostgreSQL with tables for:
- Accounts
- Holdings
- Tax lots
- Transactions
- Cash balances

Migrations in `database/sql/migrations/`.
