# Seed Monay Beads Tasks

Read the following documentation and file beads epics and issues for all planned work.

## Source Documents

1. `README.md` - Project overview, features list, and architecture TODOs
2. `services/holdings/docs/cash-tracking.md` - Implementation tasks for cash/income tracking (Phases 2-3)

## Instructions

### 1. Read Source Documents

First, read all source documents thoroughly to understand:
- Current project state (what's implemented)
- Planned features (from README features table)
- Specific implementation tasks (from cash-tracking.md)

### 2. Create Epics for Major Feature Areas

File epics for these high-level features from the documentation:

**From README.md:**
- Stock Plan Importer (RSUs, ESPP) - mentioned in "Stock Plan Accounts" section
- Frontend Dashboard (Next.js) - from architecture diagram
- MCP Server for AI queries - from architecture diagram
- Quotes Service - from architecture diagram
- Rebalancing Engine - from architecture diagram

**From cash-tracking.md:**
- Income Reporting (Phase 2)
- Bond Enhancements (Phase 3)

### 3. Create Child Issues Under Each Epic

Break down each epic into specific, actionable tasks. Use the implementation tasks from the docs.

**Example for Stock Plan Importer:**
- Parse E*Trade "Gains & Losses" CSV format
- Parse E*Trade "Tax Information" CSV format
- Auto-create lots for RSU vest events
- Auto-create lots for ESPP purchase events
- Link stock plan lots to brokerage account

**Example for Income Reporting (Phase 2 from cash-tracking.md):**
- Add `income summary` CLI command
- Add grouping by security for bond interest
- Add YTD income rollup
- Add monthly income rollup

**Example for Bond Enhancements (Phase 3 from cash-tracking.md):**
- Track accrued interest on bond purchases
- Track accrued interest on bond sales
- Handle bond maturity events
- Handle bond call events
- (Optional) Implement premium/discount amortization

### 4. Set Priorities

Use this priority scheme:
- **P0**: Blocking/critical (nothing currently blocking)
- **P1**: High priority, current development focus
- **P2**: Normal priority, near-term work
- **P3**: Nice to have, future work

Default to P2 unless the task is clearly urgent or clearly low-priority.

### 5. Add Dependencies

Link tasks that must be done in order using `bd dep add <child> <parent>`.

**Example:**
- "Link stock plan lots to brokerage" depends on "Parse Gains & Losses CSV"
- "Add monthly rollup" might depend on "Add income summary command"

### 6. Use Clear, Actionable Titles

Good titles:
- "Add income summary CLI command"
- "Parse E*Trade Gains & Losses CSV"
- "Track accrued interest on bond purchases"

Bad titles:
- "Fix stuff"
- "Income reporting"
- "Bonds"

### 7. Include Descriptions with Acceptance Criteria

Each issue should have:
- What needs to be done
- Why it's needed (context)
- How you'll know it's done (acceptance criteria)

**Example:**
```
Title: Add income summary CLI command

Description:
Implement `monay cash income summary` command to show interest and dividend income for an account.

Context: Users need to see total income by type (interest, dividends) for tax reporting and performance tracking.

Acceptance Criteria:
- Command accepts --account-name and optional --year flags
- Groups cash transactions by type (interest, dividend)
- Shows totals for each type and grand total
- Outputs in human-readable format (see cash-tracking.md for example)

Reference: services/holdings/docs/cash-tracking.md Phase 2
```

## Example Commands

```bash
# Create an epic
bd create "Stock Plan Importer" -t epic -p 2 --description="Parse E*Trade Stock Plan exports (Gains & Losses, Tax Information) to auto-create lots for RSU/ESPP transactions. Links stock plan lots to associated brokerage accounts."

# Create a child task under the epic (use the epic ID returned above)
bd create "Parse E*Trade Gains & Losses CSV" -p 2 --parent my-xxxx --description="Parse E*Trade Stock Plan 'Gains & Losses' export CSV to extract vest dates, quantities, FMV, and cost basis for RSU/ESPP lots."

# Add a dependency between two tasks
bd dep add my-child my-parent
```

## After Seeding

1. Run `bd sync` to commit all issues to git
2. Run `bd list` to view all issues
3. Run `bd ready` to see what's unblocked and ready to work on
4. Run `bd dep tree` to visualize the dependency graph

## Tips

- Focus on extracting concrete tasks from the docs, not inventing new work
- Keep epic descriptions high-level, task descriptions specific
- Don't file issues for work that's already done (check README current state)
- Use the exact terminology from the docs (e.g., "Gains & Losses" not "gains and losses")
- Reference source files in descriptions (e.g., "See README.md Stock Plan Accounts section")
