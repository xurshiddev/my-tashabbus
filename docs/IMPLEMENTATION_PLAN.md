# Implementation Plan

## Stage 0

- Project foundation.
- Go API health endpoint.
- Go Telegram Bot `/start` placeholder.
- React Admin foundation.
- React Telegram Mini App foundation.
- PostgreSQL Docker Compose foundation.
- Documentation, Makefile, tests, and project rules.

## Stage 0.2

- Validation infrastructure.
- GitHub Actions CI.
- Local setup guide.
- Toolchain check script.
- `validate-local` command.

## Stage 0.3

- Git repository readiness review.
- CI workflow review.
- GitHub push and Actions documentation.
- CI fix preparation before Stage 1.

Stage 1 can start only after:

- API tests pass.
- Bot tests pass.
- Admin build passes.
- Mini App build passes.
- Docker Compose config passes.
- API `/health` works either locally or in a CI-supported environment.

## Stage 1

- Auth, users, roles, and Telegram identity binding foundation.
- PostgreSQL users table.
- JWT access token foundation.
- Admin development login.
- Telegram Mini App initData authentication endpoint.
- Current-user and role guard middleware.

Stage 1 intentionally does not include MFY, street, household, assignment, dashboard statistic, report, or voting workflow logic.

## Stage 2

- MFY table and API module.
- Streets table and API module.
- Street leader assignment table and API module.
- Scoped Stage 2 permissions for SUPER_ADMIN, MFY_CHAIRMAN, and STREET_LEADER.
- Admin MFY/street management placeholders.
- Mini App "My Streets" view.

Stage 2 intentionally does not include households, responsible person assignment ranges, vote tracking, dashboard statistics, reports, official voting, or SMS collection.

## Stage 3

- Households and responsible person assignments.
- Household ranges.
- Responsible person scoped permissions.
- Household status foundation.

## Stage 4

- Dashboards and statistics.

## Stage 5

- Telegram Mini App real workflow.

## Stage 6

- Telegram Bot notifications and reminders.
