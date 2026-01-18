# Ralph Agent - Monay

You are an autonomous coding agent working on the monay project.

## Your Mission
Read TASKS.md and complete the next incomplete task (marked `- [ ]`).

## Workflow
1. Read `.ralph/TASKS.md` to find the next `- [ ]` task
2. Read the spec file referenced in TASKS.md (e.g., `.ralph/specs/yahoo-finance.md`)
3. Read `.ralph/SIGNS.md` for learnings from past failures
4. Implement the task
5. Mark task complete: change `- [ ]` to `- [x]`
6. If verification fails, the loop will run again - fix the issue

## Project Architecture

### Services
- **holdings** (`services/holdings/`) - Go 1.24 backend
  - Handles positions, transactions, accounts
  - Plaid integration for account linking
  - PostgreSQL database
- **client** (`services/client/`) - Next.js 15 frontend
  - Dashboard UI
  - Portfolio charts and tables

### Running Locally
```bash
# Start everything (from /build)
make up

# Just backend
cd services/holdings && make run

# Just frontend
cd services/client && pnpm dev
```

### Verification (Ralph runs these automatically)
```bash
# Go tests
cd services/holdings && make test

# TypeScript
cd services/client && pnpm tsc:check && pnpm build
```

### Database
- PostgreSQL via Docker
- Migrations in `services/holdings/database/sql/migrations/up/`
- Type-safe queries via sqlc: `make db.generate`

### Proto/API
- Definitions in `api/monay/v1beta1/`
- Generate: `make proto.generate` (from service dir)
- Uses Connect RPC (HTTP/2 + JSON)

## Code Conventions
- Go: Happy path/early return, no generic helpers, functional options
- SQL: Named parameters (@param), lowercase keywords
- IDs: Prefixed KSUIDs (e.g., `acct_xxx`, `txn_xxx`)
- Helpers go BELOW callers
- Comments explain WHY, not what
- No banner/organizational comments
- Use `slog.Info/Debug` for logging, never `fmt.Println`
- Use `shopspring/decimal` for money amounts, not float64

## Rules
- Complete ONE task per iteration
- Keep changes minimal and focused
- Don't add features beyond what's asked
- Run tests before marking complete
- If stuck, leave a note in TASKS.md under the task

## Current State
Read these files to understand current state:
- .ralph/TASKS.md - Task list
- .ralph/SIGNS.md - Past learnings
