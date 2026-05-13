# Project Rules

## Engineering Rules

- Keep code clean, explicit, and easy to test.
- Do not write quick-and-dirty solutions.
- Prefer small, focused packages with clear responsibilities.
- Business logic must not live inside HTTP handlers or Telegram handlers.
- The API owns permission checks and business rules.
- Every feature must include validation and permission checks before it is considered complete.
- Every important change must be testable with automated tests or a clear manual verification path.
- Avoid global mutable state except for bootstrap wiring where it is appropriate.
- Keep configuration explicit and environment-driven.
- Keep error handling clear and actionable.

## Development Rules

- Use test-first or test-aware development.
- Add or update tests for behavior changes.
- Run relevant tests and builds after changes.
- AI coders must report what changed and what was tested.
- AI coders must state any commands that could not be run and why.
- Do not implement voting automation, SMS code collection, or voting on behalf of citizens.

## Stage 0 Boundaries

- Do not implement business modules yet.
- Do not add MFY, street, household, or assignment tables yet.
- Do not add fake complex auth.
- Do not connect the Telegram Bot directly to PostgreSQL.
