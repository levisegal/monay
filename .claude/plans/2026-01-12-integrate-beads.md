# Plan: Integrate Beads Task Tracker with AI-Driven Seeding

## Goal

Install and configure beads (a git-backed issue tracker for AI agents) in the monay monorepo, then create an AI-driven workflow to seed initial tasks from existing documentation.

## Overview

Beads is a distributed, git-backed graph issue tracker designed for AI agents. It stores issues as JSONL in `.beads/` and provides dependency tracking, hierarchical task IDs, and JSON output optimized for agents.

**Key decisions:**
1. Initialize beads at monorepo root (not per-service)
2. Use short issue prefix: `my-` (for monay)
3. Create an AGENTS.md file to instruct AI agents to use beads
4. Seed tasks via AI agent reading existing docs and filing epics/issues

## Critical Files

- `/Users/levi/dev/levisegal/monay/AGENTS.md` - New file to configure AI agents to use beads
- `/Users/levi/dev/levisegal/monay/.beads/` - Directory created by `bd init` (stores issues.jsonl)
- `/Users/levi/dev/levisegal/monay/scripts/seed-beads.md` - Prompt file for AI-driven task seeding
- `/Users/levi/dev/levisegal/monay/README.md` - Source of project features and TODOs
- `/Users/levi/dev/levisegal/monay/services/holdings/docs/cash-tracking.md` - Source of implementation tasks

## Implementation Steps

### 1. Install beads CLI

Install the `bd` command globally:

```bash
# Via Go (preferred since monay is a Go project)
go install github.com/steveyegge/beads/cmd/bd@latest

# Or via Homebrew
brew install steveyegge/beads/bd

# Or via npm
npm install -g @beads/bd
```

### 2. Initialize beads in monorepo

```bash
cd /Users/levi/dev/levisegal/monay
bd init --prefix my-
```

This creates:
- `.beads/` directory with `issues.jsonl`
- Local SQLite cache for fast queries

### 3. Create AGENTS.md

Create `/Users/levi/dev/levisegal/monay/AGENTS.md` to instruct AI agents:

```markdown
# Agent Instructions

Use `bd` (beads) for task tracking in this repository.

## Workflow

1. **Starting work**: Run `bd ready` to see available tasks
2. **Pick a task**: Run `bd show <id>` to see details
3. **Update status**: `bd update <id> --status in_progress`
4. **File new issues**: `bd create "Title" -p 1 --description="Details"`
5. **Complete work**: `bd close <id> --reason "completed"`
6. **Before stopping**: Run `bd sync` then `git push`

## Issue Prefix

All issues use the `my-` prefix (e.g., `my-a1b2`).

## Filing Issues

File beads issues for:
- Work taking longer than 2 minutes
- Follow-up tasks discovered during implementation
- Code review findings
- Bug reports

## Dependencies

Use `bd dep add <child> <parent>` to link related tasks.
```

### 4. Create seeding prompt file

Create `/Users/levi/dev/levisegal/monay/scripts/seed-beads.md` with a prompt for AI agents to seed tasks:

```markdown
# Seed Monay Beads Tasks

Read the following documentation and file beads epics and issues for all planned work.

## Source Documents

1. README.md - Project overview, features, and architecture TODOs
2. services/holdings/docs/cash-tracking.md - Implementation tasks for cash/income tracking

## Instructions

1. First, read all source documents thoroughly
2. Create epics for major feature areas:
   - Stock Plan Importer (from README TODO)
   - Income Reporting (Phase 2 from cash-tracking.md)
   - Bond Enhancements (Phase 3 from cash-tracking.md)
   - Frontend Dashboard (from README architecture)
   - MCP Server (from README architecture)

3. Under each epic, create child issues for specific implementation tasks
4. Set appropriate priorities:
   - P0: Blocking/critical
   - P1: High priority, current sprint
   - P2: Normal priority
   - P3: Nice to have

5. Add dependencies where tasks must be done in order
6. Use clear, actionable titles (e.g., "Add income summary CLI command")
7. Include descriptions with acceptance criteria

## Example Commands

```bash
# Create epic
bd create "Stock Plan Importer" -t epic -p 1 --description="Parse E*Trade Stock Plan exports for RSU/ESPP lots"

# Create child task under epic
bd create "Parse E*Trade Gains & Losses CSV" -p 1 --parent my-xxxx --description="Extract vest dates, quantities, and FMV from Stock Plan exports"
```

## After Seeding

Run `bd sync` to commit all issues, then `bd list` to verify.
```

### 5. Run initial task seeding

After steps 1-4, run the seeding process:

```bash
# Start a Claude Code session and provide the seeding prompt
cd /Users/levi/dev/levisegal/monay
# Claude reads scripts/seed-beads.md and files issues
```

The AI agent will:
1. Read README.md and cash-tracking.md
2. File epics for major features
3. Create child issues with dependencies
4. Run `bd sync` to commit everything

### 6. Verify and adjust

```bash
# List all issues
bd list

# View dependency tree
bd dep tree

# Check for ready tasks
bd ready

# Clean up if needed
bd doctor --fix
```

## Tasks to Seed (Reference)

Based on existing documentation:

**Epic: Stock Plan Importer**
- Parse E*Trade Stock Plan "Gains & Losses" exports
- Parse E*Trade "Tax Information" exports
- Auto-create lots for vest/purchase events
- Link stock plan lots to associated brokerage account

**Epic: Income Reporting (Phase 2)**
- Add `income summary` CLI command
- Group income by security for bond interest tracking
- Add YTD/monthly income rollups

**Epic: Bond Enhancements (Phase 3)**
- Track accrued interest on bond purchases/sales
- Handle bond maturity and call events
- Amortization of premium/discount (optional)

**Epic: Frontend Dashboard**
- Holdings table view
- Allocation pie chart
- Daily P&L display
- CSV upload component

**Epic: MCP Server**
- query_holdings tool
- get_portfolio_value tool
- get_daily_pnl tool
- calculate_rebalance tool

## Open Questions

(none)
