# Signs (Learnings)

When Ralph makes mistakes, add guidance here to prevent recurrence.

## Go Patterns
- Always use `slog.Info/Debug` for logging, never `fmt.Println`
- Use `shopspring/decimal` for money amounts, not float64
- Helpers go BELOW callers, not above

## Frontend Patterns
- Use existing component patterns from `components/ui/`
- Check TypeScript types before committing
- Use TanStack Query for server state
- Reference PAPER_PORTFOLIO_DESIGN.md for design system

## Database Patterns
- Run `make db.generate` after changing queries

## Proto Patterns
- Service definitions in `*-service.proto`
- Message definitions in domain-specific files
- Run `make proto.generate` after changes

## Common Mistakes
<!-- Add learnings from failures here -->
