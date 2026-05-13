# My Tashabbus

My Tashabbus is a production-minded foundation for monitoring Tashabbusli Budget operational progress inside an MFY/mahalla. It does not vote for citizens and does not collect SMS codes. Official voting happens outside this system.

Stage 0 creates only the project foundation. It intentionally does not include business logic, real auth, MFY records, street records, households, assignments, dashboards, or voting workflows.

## Apps

- `apps/api`: Go HTTP API, future source of truth for permissions and business rules.
- `apps/bot`: Go Telegram Bot service with `/start` and a Mini App button placeholder.
- `apps/admin`: React + TypeScript admin dashboard foundation.
- `apps/miniapp`: React + TypeScript Telegram Mini App foundation.

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

Check that required tools are installed:

```sh
make check-tools
```

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

The bot is behind the `bot` Docker Compose profile so a missing `BOT_TOKEN` does not block the default development stack.

## Run Locally

API:

```sh
cd apps/api
go run ./cmd/api
```

Bot:

```sh
cd apps/bot
BOT_TOKEN=your_token MINI_APP_URL=http://localhost:5174 go run ./cmd/bot
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

To review CI, open the GitHub repository, go to the Actions tab, and inspect the latest workflow run. If CI fails, fix those errors before starting Stage 1.

## Stage 1 Auth Examples

Development login:

```sh
curl -X POST http://localhost:8080/auth/dev-login \
  -H "Content-Type: application/json" \
  -d '{"full_name":"Dev Super Admin","role":"SUPER_ADMIN"}'
```

Current user:

```sh
curl http://localhost:8080/auth/me \
  -H "Authorization: Bearer <token>"
```

Telegram Mini App auth uses `POST /auth/telegram` with Telegram WebApp `initData`. In production, the API requires a real `BOT_TOKEN` and validates the Telegram signature. Unregistered Telegram users receive `USER_NOT_REGISTERED`.

## Project Stages

- Stage 0: project foundation.
- Stage 1: auth, users, roles, Telegram identity.
- Stage 2: MFY and streets.
- Stage 3: households and assignments.
- Stage 4: dashboards and statistics.
- Stage 5: Telegram Mini App real workflow.
- Stage 6: Telegram Bot notifications and reminders.
