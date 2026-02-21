---
name: gather
description: Automates downloading transaction CSVs from brokerage websites using playwright-cli browser automation. The user logs in manually; the skill handles navigation, screenshots, and file management.
user_invocable: true
argument: "<broker> — which brokerage to gather from (e.g., etrade)"
---

# Gather — Brokerage Data Collection

Automate the tedious part of downloading transaction CSVs from brokerage sites. You drive the browser with playwright-cli while the user handles login/MFA.

## Invocation

```
/gather etrade
```

The argument is the broker name. Load the matching guide from `brokers/<broker>.md`. If no guide exists, tell the user and offer to help create one from `brokers/_template.md`.

## Allowed Tools

- `Bash` — only: `playwright-cli *`, `mv`, `ls`, `wc`, `mkdir`, `head`
- `Read`, `Write`, `Edit`, `Glob`

## Memory: Pre-flight

Before starting any gather session, load knowledge in this order:

1. **`memory/MEMORY.md`** — cross-broker learnings (browser modes, download behavior, reliability)
2. **`memory/brokers/<broker>.md`** — broker-specific learnings (CSV formats, nav quirks, account lists)
3. **`memory/sessions/`** — today's session log (if it exists) + yesterday's (recent raw context)

All paths are relative to this skill directory (`.claude/skills/gather/`).

## Memory: Session Logging

During the gather session, append observations to `memory/sessions/YYYY-MM-DD.md` as they happen. Multiple gather runs on the same day append to the same file. Use timestamped entries:

```markdown
## 14:30 — etrade: Account selector behavior
The Activity page dropdown doesn't inherit the account from portfolio nav.
Had to explicitly select maya-3758 from the Accounts combobox.
```

Log these kinds of things:
- Navigation surprises (new modals, changed URLs, timeouts)
- Download quirks (wrong format, missing files, filename changes)
- Account changes (new accounts, removed accounts, balance snapshots)
- Browser issues (crashes, reconnection attempts)
- Successful patterns worth remembering

Session logs are append-only — never edit previous entries.

## Workflow

### 1. Pre-flight

- Load memory (see "Memory: Pre-flight" above)
- Load the broker guide from `brokers/<broker>.md`
- Print known accounts for the broker (from the guide + memory)
- Confirm target directory exists: `data/imports/<broker>/`
- Check what files already exist (avoid re-downloading)
- Ask user which accounts to gather (default: all remaining)
- Remind user to have credentials ready

### 2. Browser Setup

```bash
playwright-cli open --headed <broker_login_url>
```

Tell the user: "Please log in and complete any MFA. Let me know when you're on the main account overview."

**Wait for user confirmation before proceeding.** Do not attempt to automate login.

### 3. Verify Login

After user confirms:

```bash
playwright-cli snapshot
```

Describe what's visible. Confirm you can see the expected account overview page. If not, ask the user to navigate there manually.

### 4. Per-Account Loop

For each account (from the broker guide):

**a. Navigate to account**
- Follow broker guide navigation steps
- After each navigation action: `playwright-cli snapshot`
- Describe what's visible, confirm it matches expectations
- If unexpected state: offer retry, manual nav, or skip
- Log anything surprising to the session log

**b. Screenshot portfolio**
```bash
playwright-cli screenshot --filename=data/imports/<broker>/<account>/screenshot_portfolio.png
```

**c. Navigate to transactions**
- Follow broker guide steps to reach transaction history
- Set date range per broker guide (typically: Jan 1 of previous year through today)
- Snapshot and confirm the date range looks correct

**d. Download CSV**
- Follow broker guide download steps
- Downloads land in `.playwright-cli/` directory — move immediately after each download
- Move to target: `data/imports/<broker>/<account>/transactions_<year>.csv`

**e. Verify download**
```bash
head -5 data/imports/<broker>/<account>/transactions_<year>.csv
wc -l data/imports/<broker>/<account>/transactions_<year>.csv
```

Confirm the file header matches expected format from broker guide.

**f. Check browser is still alive before next account**
- Browser can die after downloads — re-open if needed
- The user may need to re-login if session expired

### 5. Post-Gather Summary & Distillation

Print a summary table of gathered files with row counts.

Then print the import commands.

**Distill session into long-term memory:**

1. Read today's session log (`memory/sessions/YYYY-MM-DD.md`)
2. Extract durable learnings — things that would help future sessions
3. Update `memory/MEMORY.md` with cross-broker learnings (browser behavior, download patterns)
4. Update `memory/brokers/<broker>.md` with broker-specific learnings (nav changes, format changes, new accounts)
5. Remove entries from long-term memory that today's session contradicts
6. Session log stays as-is (append-only, never edited after the fact)

## Key Behaviors

- **Never automate login or MFA.** The user handles all authentication.
- **Snapshot after every navigation.** Describe what you see. Ask user to confirm before proceeding.
- **On unexpected state:** Don't guess. Offer: (1) retry the action, (2) user navigates manually, (3) skip this account.
- **File naming:** `transactions_<year>.csv` — if the broker exports a single file spanning years, name it `transactions_<start>_<end>.csv`.
- **Idempotent:** If a file already exists at the target path, ask before overwriting.
- **Move downloads immediately.** Files download to `.playwright-cli/` and may be overwritten by the next download.
- **Log as you go.** Append observations to the session log throughout the gather, not just at the end.
- **Distill after every session.** Promote durable learnings from session logs to long-term memory.

## File Organization

```
.claude/skills/gather/
  SKILL.md
  brokers/                              # Reference docs (navigation guides)
    etrade.md
    _template.md
  memory/                               # Learned knowledge
    MEMORY.md                           # Cross-broker (curated)
    brokers/
      etrade.md                         # Broker-specific (curated)
    sessions/
      2026-02-20.md                     # Raw session log (append-only)

data/imports/
  <broker>/
    <account>/
      transactions_2025.csv
      transactions_2026.csv
      screenshot_portfolio.png
      screenshot_transactions.png
```

Clean split: `brokers/` = how to navigate, `memory/` = what we've learned.

## Adding New Brokers

Copy `.claude/skills/gather/brokers/_template.md` and fill in the broker-specific details. The template has all the sections you need.
