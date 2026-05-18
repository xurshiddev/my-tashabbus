# My Tashabbus

My Tashabbus is a production-minded system for monitoring Tashabbusli Budget operational progress inside one MFY/mahalla. It does not vote for citizens and does not collect SMS codes. Official voting happens outside this system.

The active MVP is bot-first and Mini App-first:

- One Telegram bot belongs to exactly one MFY.
- Another MFY should run another bot/deployment/env set.
- The Admin app and JWT login remain legacy/development tooling only.
- The Telegram Mini App does not use JWT, `access_token`, localStorage sessions, `/auth/telegram`, `Authorization`, or `Bearer`.
- The Mini App authenticates protected API requests with `X-Telegram-Init-Data`.

## Apps

- `apps/api`: Go HTTP API and source of truth for permissions and business rules.
- `apps/bot`: Go Telegram Bot service with `/start`, `/myid`, and a Telegram WebApp button.
- `apps/admin`: React + TypeScript legacy/development admin dashboard.
- `apps/miniapp`: React + TypeScript Telegram Mini App for the MVP workflow.

## Prerequisites

- Go 1.22+
- Node.js 20+ and npm
- Docker and Docker Compose
- `golang-migrate`
- `sqlc`

## Environment Setup

Copy `.env.example` to `.env` and adjust values for local development.

```sh
cp .env.example .env
```

For detailed local setup instructions, see `docs/LOCAL_SETUP.md`.

Required single-MFY values:

```sh
TELEGRAM_BOT_TOKEN=123456789:replace-with-real-token
APP_MFY_NAME=My Tashabbus MFY
APP_MFY_SLUG=my-tashabbus-mfy
MFY_CHAIRMAN_TELEGRAM_ID=123456789
MINI_APP_URL=https://your-miniapp-url.example
VITE_API_BASE_URL=http://localhost:8080
CORS_ALLOWED_ORIGINS=http://localhost:5173,http://localhost:5174
```

`BOT_TOKEN` is supported only as a legacy fallback. Prefer `TELEGRAM_BOT_TOKEN`.

`MFY_CHAIRMAN_TELEGRAM_ID` identifies the chairman/owner of this single-MFY deployment. It is not a platform super admin. `ADMIN_TELEGRAM_ID` and `MFY_OWNER_TELEGRAM_ID` are supported as legacy aliases in that order after `MFY_CHAIRMAN_TELEGRAM_ID`.

`APP_MFY_SLUG` is currently config-level metadata. It is returned in Mini App context but is not persisted in the `mfys` table yet.

Check that required tools are installed:

```sh
make check-tools
```

For real Telegram Mini App testing through ngrok, use `docs/TELEGRAM_REAL_TESTING.md`. Keep local secrets in `.env.local.bot`, `.env.local.test`, and `apps/miniapp/.env.local`; use the matching `.example` files as templates.

If this folder is not already a Git repository, initialize and push it before relying on GitHub Actions:

```sh
git init
git add .
git commit -m "chore: add stage 0 project foundation and validation infrastructure"
git branch -M main
git remote add origin <repo-url>
git push -u origin main
```

## Run With Docker

```sh
make docker-up
curl http://localhost:8080/health
```

The bot is behind the `bot` Docker Compose profile so a missing `TELEGRAM_BOT_TOKEN` does not block the default development stack.

When using Docker, do not globally export a localhost `DATABASE_URL` before starting Compose. The API container must use the Docker hostname `postgres`; use inline `DATABASE_URL=... make migrate-up` only for terminal migrations.

## Run Locally

API:

```sh
cd apps/api
go run ./cmd/api
```

Bot:

```sh
cd apps/bot
TELEGRAM_BOT_TOKEN=your_token MINI_APP_URL=http://localhost:5174 go run ./cmd/bot
```

Admin:

```sh
cd apps/admin
npm install
npm run dev
```

Mini App:

```sh
cd apps/miniapp
npm install
npm run dev
```

## Tests And Builds

```sh
make test
make build
make lint
```

Frontend builds require dependencies to be installed in each frontend app.

For a full Stage 0 local validation run:

```sh
make validate-local
```

This checks local tools, validates code, starts Postgres and the API with Docker Compose, verifies `GET /health`, and shuts the Docker stack down.

## CI

GitHub Actions CI validates:

- API Go modules, tests, and build.
- Bot Go modules, tests, and build.
- Admin install, build, and lint.
- Mini App install, build, and lint.
- Docker Compose configuration and API/Bot image builds.

Stage 1 must not start until either `make validate-local` passes locally or GitHub Actions CI passes successfully.

To review CI, open the GitHub repository, go to the Actions tab, and inspect the latest workflow run.

## Mini App MVP Auth

The Mini App sends raw Telegram WebApp init data on every protected request:

```http
X-Telegram-Init-Data: <raw Telegram WebApp initData>
```

Current Mini App user:

```sh
curl http://localhost:8080/miniapp/me \
  -H "X-Telegram-Init-Data: <raw Telegram WebApp initData>"
```

If the Telegram ID equals `MFY_CHAIRMAN_TELEGRAM_ID`, the backend bootstraps the user as `MFY_CHAIRMAN`. Unknown non-chairman users receive `USER_NOT_ASSIGNED` and should send their Telegram ID to the MFY chairman.

Legacy JWT/dev endpoints such as `/auth/dev-login`, `/auth/me`, `/users`, and Admin screens still exist for development, but they are not required for the Mini App MVP.

The current Mini App UI is a role-only smoke test. It calls only `GET /miniapp/me` and shows the MFY name, user, and role. Street, household, assignment, progress, dashboard, and statistics tools are intentionally hidden until this first Telegram auth path is stable.

## Current Mini App Endpoint Names

These names exist in the API and are documented as-is for later MVP workflow wiring:

- `GET /miniapp/me`
- `GET /miniapp/my/streets`
- `GET /miniapp/my/households`
- `GET /miniapp/streets/{street_id}/households`
- `POST /miniapp/streets/{street_id}/households`
- `PATCH /miniapp/households/{id}`
- `GET /miniapp/households/{id}/logs`
- `GET /miniapp/streets/{street_id}/responsibles`
- `POST /miniapp/streets/{street_id}/responsibles`
- `POST /miniapp/responsible-assignments/{id}/deactivate`

Missing MVP work, intentionally not implemented in Cleanup-1:

- Chairman street creation from Mini App.
- Street leader assignment by Telegram ID from Mini App.
- Progress/dashboard/statistics endpoints.

Future role assignment design:

- A user sends `/myid` to the bot.
- The user gives that Telegram ID to the MFY chairman.
- The chairman later assigns `STREET_LEADER` or `RESPONSIBLE_PERSON` by Telegram ID in the Mini App.
- A later improvement can show a "new unassigned users" list. The bot/Mini App should not assume access to Telegram contact lists.

## Legacy Admin/JWT Examples

Development login for legacy Admin/dev tooling:

```sh
curl -X POST http://localhost:8080/auth/dev-login \
  -H "Content-Type: application/json" \
  -d '{"full_name":"Dev Super Admin","role":"SUPER_ADMIN"}'
```

Current legacy JWT user:

```sh
curl http://localhost:8080/auth/me \
  -H "Authorization: Bearer <token>"
```

Household data is aggregate operational tracking only. It must not include citizen credentials, SMS codes, passport data, or any official voting automation.

## Project Stages

- Stage 0: project foundation.
- Stage 1: auth, users, roles, Telegram identity.
- Stage 2: MFY and streets.
- Stage 3: households and assignments.
- Stage 4: dashboards and statistics.
- Stage 5: Telegram Mini App real workflow.
- Stage 6: Telegram Bot notifications and reminders.
