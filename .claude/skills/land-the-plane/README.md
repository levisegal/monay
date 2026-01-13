# Land the Plane Skill

Mandatory session completion protocol ensuring all work is properly committed and pushed before ending a work session.

## Purpose

Prevents work from being stranded locally by enforcing a structured workflow that ends with successful `git push`.

## Key Principle

**Work is NOT considered complete until `git push` succeeds.**

## Workflow

1. **File remaining issues** - Document follow-up work
2. **Execute quality gates** - Run tests, linters, builds
3. **Update issue tracking** - Close finished, update in-progress
4. **Commit all changes** - Clear commit messages
5. **Push to remote** - `git pull --rebase && bd sync && git push`
6. **Clean up** - Clear stashes, prune branches
7. **Verify and hand off** - Confirm push success, provide context

## Invocation

- User says "land the plane", "wrap up", "end session", "finish up"
- Session is ending
- User explicitly requests to commit/push work

## Critical Rules

- Never stop before pushing completed work
- Never delegate pushing with "ready when you are"
- Resolve any push failures and retry until successful
- All changes must be committed AND pushed
- Quality gates must pass or failures must be documented

## Example Usage

```bash
# User says: "land the plane"
# Claude:
# 1. Creates issues for follow-up work
# 2. Runs make test, pnpm build, etc.
# 3. Closes completed beads
# 4. Commits all changes
# 5. Executes: git pull --rebase && bd sync && git push
# 6. Verifies git status shows "up to date"
# 7. Provides handoff summary
```

## Integration with Beads

This skill is designed to work seamlessly with the beads issue tracking system:
- Uses `bd new task` to create follow-up issues
- Uses `bd close` to close completed work
- Uses `bd update` to note progress
- Uses `bd sync` before pushing to sync JSONL

## Related

- See [AGENTS.md](https://github.com/steveyegge/beads/blob/main/AGENTS.md) in the beads repository for the original specification
