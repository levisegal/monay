---
name: land-the-plane
description: Mandatory session completion protocol that ensures all work is properly committed and pushed before ending a work session. Prevents work from being stranded locally.
---

# Land the Plane - Session Completion Protocol

## Core Principle

**Work is NOT considered complete until `git push` succeeds.**

This skill implements a mandatory workflow ensuring all changes are safely pushed to the remote repository before ending a work session. Never stop before pushing completed work.

## When This Skill Activates

This skill should be invoked when:
- User says "land the plane", "wrap up", "end session", "finish up"
- Session is about to end
- User explicitly requests to commit/push work
- Wrapping up a significant piece of work

## Mandatory Workflow Steps

### Step 1: File Remaining Issues

Create beads issues for any follow-up work discovered during this session:

```bash
# For each piece of follow-up work needed:
bd new task "[Clear description]" --priority [0-4] --deps [parent-issue-id]
```

**Examples of follow-up work:**
- Bugs discovered but not fixed
- Features partially implemented
- Tech debt identified
- Tests that need writing
- Documentation updates needed
- Refactoring opportunities

**Important:** Don't leave work undocumented. If you thought "we should do X later," create an issue NOW.

### Step 2: Execute Quality Gates

If code was modified, run appropriate checks:

**For Go services:**
```bash
make test           # Run tests
make lint           # Run linters
make build          # Verify builds
```

**For Next.js/TypeScript:**
```bash
pnpm tsc:check      # TypeScript compilation
pnpm lint           # ESLint
pnpm build          # Production build (if appropriate)
```

**For Python services:**
```bash
poetry run ruff check    # Linting
poetry run pytest        # Tests
```

**Critical:** If any quality gate fails, you MUST:
1. Fix the issues OR
2. Create a bead issue documenting the failure and why it's acceptable to proceed OR
3. Revert the breaking changes

**Never push failing tests or broken builds without explicit documentation.**

### Step 3: Update Issue Tracking

**Close finished work:**
```bash
# For each completed issue:
bd close [issue-id] --reason "[What was completed and verified]"
```

**Update in-progress work:**
```bash
# Update issues that are partially complete:
bd update [issue-id] --notes "[Progress made, what's left to do]"
```

**Review current issues:**
```bash
bd list
```

Ensure the issue list accurately reflects reality. No zombie issues, no forgotten work.

### Step 4: Commit All Changes

**Review what's changed:**
```bash
git status
git diff
```

**Stage and commit:**
```bash
git add .
git commit -m "[Clear commit message]

[Optional detailed explanation]

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>"
```

**Commit message guidelines:**
- Start with imperative verb (Add, Update, Fix, Remove, Refactor)
- Be specific about WHAT changed
- Include WHY if not obvious
- Reference bead IDs if applicable

**Examples:**
- ✅ "Add Paper Portfolio design system and Next.js dashboard"
- ✅ "Fix holdings table to handle empty states (my-abc)"
- ✅ "Refactor CSV importer to support E*Trade format"
- ❌ "Updates" (too vague)
- ❌ "WIP" (never push WIP)

### Step 5: Push to Remote

**Execute the push sequence:**

```bash
# Pull latest changes with rebase
git pull --rebase

# Sync beads (updates JSONL from database)
bd sync

# Push everything to remote
git push

# Verify success
git status
```

**Expected output from `git status`:**
```
On branch [branch-name]
Your branch is up to date with 'origin/[branch-name]'.

nothing to commit, working tree clean
```

**If push fails:**

**Conflict during rebase:**
```bash
# Resolve conflicts in files
# Then:
git add .
git rebase --continue
git push
```

**Push rejected (non-fast-forward):**
```bash
# NEVER force push to main/master
# For feature branches:
git pull --rebase
git push
```

**Authentication failure:**
```bash
# Check credentials
# Retry push
git push
```

**Retry until push succeeds.** Do NOT proceed if push fails.

### Step 6: Clean Up

**Clear stashes (if any):**
```bash
git stash list
# If there are stashes you don't need:
git stash drop stash@{0}
```

**Prune remote branches (optional):**
```bash
git fetch --prune
git branch -vv | grep ': gone]' | awk '{print $1}' | xargs git branch -d
```

### Step 7: Verify and Hand Off

**Final verification:**
```bash
git log -1 --oneline    # Show last commit
git status              # Confirm clean state
bd list                 # Show current issue state
```

**Provide context for next session:**

Write a brief summary of:
1. **What was completed** - Key accomplishments this session
2. **What's in progress** - Partially completed work with context
3. **What's next** - Logical next steps
4. **Important notes** - Any gotchas, decisions made, or context needed

**Example handoff:**
```
✓ Session completed and pushed successfully

Completed:
- Built Next.js dashboard with Paper Portfolio design system (my-0dl closed)
- Created holdings table component with stat cards
- Set up Tailwind config with custom Paper Portfolio theme

In Progress:
- Frontend Dashboard epic (my-css) - basic layout done, need CSV upload and charts

Next Steps:
- Implement CSV upload component (my-css.3)
- Add allocation pie chart (my-css.2)
- Set up Yahoo Finance integration for live prices (my-qih)

Notes:
- Dev server running on localhost:3000
- Using Lora font from Google Fonts for serif typography
- Paper texture implemented via SVG data URI in Tailwind config
```

## Critical Rules

1. **Never stop before pushing** - If session must end before push succeeds, document exactly why and what's needed
2. **Never delegate pushing** - Don't say "ready to push when you are" - YOU must execute the push
3. **Resolve push failures** - Retry until `git push` succeeds
4. **All changes committed** - No uncommitted work, no stashes (unless explicitly documented)
5. **Quality gates passed** - Tests pass, builds succeed, or failures documented
6. **Issues updated** - Closed work is closed, in-progress is noted, follow-ups created

## Error Handling

**"I can't push because X"** → Document X as a bead issue, explain clearly why push is blocked

**"Tests are failing"** → Fix them OR create issue documenting why it's OK to proceed OR revert

**"I'm not sure what to commit"** → Review git diff, break into logical commits

**"Push failed with conflict"** → Resolve conflict, rebase, retry push

**"Session ending before push"** → NOT ACCEPTABLE - extend session to complete push

## Non-Goals

This skill does NOT:
- Create pull requests (that's a separate workflow)
- Deploy to production
- Run CI/CD pipelines (those run automatically on push)
- Handle multi-repository synchronization

## Usage Examples

**Standard session end:**
```
User: "land the plane"
Assistant: [Executes all 7 steps, verifies push success, provides handoff summary]
```

**After completing a feature:**
```
User: "wrap this up"
Assistant: [Files follow-up issues, runs tests, closes bead, commits, pushes, confirms]
```

**Quick commit-and-push:**
```
User: "commit and push this"
Assistant: [Skips some steps if appropriate, focuses on commit message and push verification]
```

---

**Remember:** The plane isn't landed until `git push` succeeds and you've verified `git status` shows "up to date with origin".
