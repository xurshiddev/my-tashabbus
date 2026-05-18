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
- Household change logs for important updates.
- Admin household and responsible assignment placeholders.
- Mini App "My Households" flow.

Stage 3 intentionally does not include dashboard statistics, reports, official voting, SMS collection, or automated voting.

## Cleanup-1

- Align documentation and env examples with the single-MFY product decision.
- Prefer `TELEGRAM_BOT_TOKEN`; keep `BOT_TOKEN` only as a legacy fallback.
- Document that one bot belongs to one MFY and another MFY uses another deployment/env set.
- Document that the Mini App authenticates protected requests with `X-Telegram-Init-Data`.
- Keep Admin/JWT as legacy/development tooling only.
- Do not normalize Mini App endpoint names yet; document the current names.
- Document that `APP_MFY_SLUG` is currently config-level metadata and is not persisted in the database.

Cleanup-1 intentionally does not add street creation, street leader assignment, progress dashboards, statistics, reports, official voting, or SMS collection.

## Next MVP Workflow Work

- Add Mini App chairman street creation.
- Add Mini App street leader assignment by Telegram ID.
- Add focused route tests for Mini App owner bootstrap and role-scoped workflows.
- Run real Telegram Mini App testing end to end before Stage 4.

## Stage 4

- Dashboard and statistics foundation.
- MFY dashboard.
- Street dashboard.
- Responsible person progress.
- Status counts and safe aggregate reporting.

## Stage 5

- Telegram Mini App real workflow.

## Stage 6

- Telegram Bot notifications and reminders.
