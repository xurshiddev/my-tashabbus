# Local Setup

This guide explains how to prepare a local machine to validate and run My Tashabbus.

My Tashabbus is not an official voting system. It must not automate voting, collect SMS codes, or vote on behalf of citizens.

The active MVP is a single-MFY deployment:

- One Telegram bot belongs to one MFY.
- Another MFY uses another bot/deployment/env set.
- The Admin app and JWT login are legacy/development tooling only.
- The Telegram Mini App authenticates protected requests with `X-Telegram-Init-Data`, not JWT.

## Required Tools

- Go 1.22+
- Node.js 20+ and npm
- Docker Desktop or Docker Engine with Docker Compose

## Optional Tools

- `golang-migrate` for database migrations
- `sqlc` for future query generation

## macOS Setup Notes

Install:

- Go from the official Go installer or Homebrew.
- Node.js/npm from the official Node.js installer, `nvm`, or Homebrew.
- Docker Desktop for macOS.

Verify:

```sh
go version
node --version
npm --version
docker --version
docker compose version
```

## Ubuntu/Linux Setup Notes

Install:

- Go from the official Go downloads or your package manager.
- Node.js/npm from NodeSource, `nvm`, or your package manager.
- Docker Engine and Docker Compose plugin from Docker's official installation guide.

Verify:

```sh
go version
node --version
npm --version
docker --version
docker compose version
```

## Tool Check

From the repository root:

```sh
make check-tools
```

The script prints detected versions for required tools and reports missing optional tools without failing.

## Git Repository Setup

If this directory has not been initialized as a Git repository yet, run:

```sh
git init
git add .
git commit -m "chore: add stage 0 project foundation and validation infrastructure"
git branch -M main
git remote add origin <repo-url>
git push -u origin main
```

Replace `<repo-url>` with the GitHub repository URL.

If the repository already exists locally, check the current state before committing:

```sh
git status --short
```

## GitHub Actions

After pushing to GitHub:

1. Open the GitHub repository.
2. Go to the Actions tab.
3. Open the latest workflow run.
4. Inspect any failing job logs.
5. Fix CI errors before starting Stage 1.

Stage 1 can start only when either `make validate-local` passes locally or GitHub Actions CI passes successfully.

## Full Local Validation

From the repository root:

```sh
make validate-local
```

This runs tool checks, Go tests/builds, frontend installs/builds/lints, Docker Compose config validation, starts Postgres and API, checks `/health`, and stops the Docker stack.

## Real Telegram Mini App Testing

For real Telegram WebApp testing with ngrok, see `docs/TELEGRAM_REAL_TESTING.md`.

The Mini App public ngrok URL is used by the bot button. The API public ngrok URL is used by the Mini App frontend when it calls the backend.

Keep these URLs separate:

```sh
MINI_APP_URL=https://<miniapp-ngrok>.ngrok-free.app
VITE_API_BASE_URL=https://<api-ngrok>.ngrok-free.app
```

`MINI_APP_URL` is the frontend/Vite tunnel used by Telegram's WebApp button. `VITE_API_BASE_URL` is the Go API tunnel used by browser `fetch` calls. A single ngrok URL cannot serve both unless you configure a reverse proxy.

Minimum single-MFY env setup:

```sh
export TELEGRAM_BOT_TOKEN="123456789:replace-with-real-token"
export APP_MFY_NAME="My Tashabbus MFY"
export APP_MFY_SLUG="my-tashabbus-mfy"
export MFY_CHAIRMAN_TELEGRAM_ID="123456789"
export MINI_APP_URL="https://<miniapp-ngrok>.ngrok-free.app"
```

`BOT_TOKEN` is only a legacy fallback. Prefer `TELEGRAM_BOT_TOKEN`.

`MFY_CHAIRMAN_TELEGRAM_ID` is the chairman/owner of this single-MFY deployment, not a platform super admin. `ADMIN_TELEGRAM_ID` and `MFY_OWNER_TELEGRAM_ID` are accepted as legacy aliases.

`APP_MFY_SLUG` is currently config-level metadata only; the database MFY row is created from `APP_MFY_NAME`.

Configure the backend CORS list with the current Mini App origin before restarting the API container:

```sh
export CORS_ALLOWED_ORIGINS="http://localhost:5173,http://localhost:5174,https://<miniapp-ngrok>.ngrok-free.app"
unset DATABASE_URL
docker compose -f docker-compose.dev.yml down
docker compose -f docker-compose.dev.yml up -d --build postgres api
```

Before opening Telegram, verify the API URL and CORS preflight:

```sh
API_URL="https://api-ngrok-url.ngrok-free.app"
MINIAPP_ORIGIN="https://miniapp-ngrok-url.ngrok-free.app"

curl -i "$API_URL/health"
curl -i "$API_URL/miniapp/me"
curl -i -X OPTIONS "$API_URL/miniapp/me" \
  -H "Origin: $MINIAPP_ORIGIN" \
  -H "Access-Control-Request-Method: GET" \
  -H "Access-Control-Request-Headers: X-Telegram-Init-Data, Content-Type, ngrok-skip-browser-warning"

curl -i "$API_URL/miniapp/me" \
  -H "Origin: $MINIAPP_ORIGIN" \
  -H "X-Telegram-Init-Data: test" \
  -H "ngrok-skip-browser-warning: true"
```

`/health` must return API JSON. `/miniapp/me` without the header should return `TELEGRAM_INIT_DATA_MISSING`. `/miniapp/me` with fake init data should return `TELEGRAM_INIT_DATA_INVALID`. The OPTIONS response must allow `X-Telegram-Init-Data` and `ngrok-skip-browser-warning`.

Use `DATABASE_URL` inline only for local terminal migrations:

```sh
DATABASE_URL="postgres://my_tashabbus:my_tashabbus_password@localhost:5432/my_tashabbus?sslmode=disable" make migrate-up
```

Do not export a localhost `DATABASE_URL` globally before running Docker Compose. The API container must use the `postgres` hostname inside Docker.

Mini App Vite env:

```sh
cat > apps/miniapp/.env.local <<'EOF'
VITE_API_BASE_URL=https://<api-ngrok>.ngrok-free.app
VITE_DEV_TELEGRAM_AUTH=false
VITE_ALLOWED_HOSTS=<miniapp-ngrok>.ngrok-free.app
EOF
```

Restart `npm run dev` after changing `apps/miniapp/.env.local`.

Start the bot with the current Mini App URL:

```sh
cd apps/bot
TELEGRAM_BOT_TOKEN="$TELEGRAM_BOT_TOKEN" MINI_APP_URL="$MINI_APP_URL" go run ./cmd/bot
```

After every `MINI_APP_URL` or ngrok change, send `/start` again in Telegram so the new WebApp button is used.

Expected real Mini App behavior:

- The chairman Telegram ID from `MFY_CHAIRMAN_TELEGRAM_ID` becomes `MFY_CHAIRMAN`.
- Non-chairman, unassigned users see `USER_NOT_ASSIGNED` and should send their Telegram ID to the MFY chairman.
- The Mini App never stores an access token and never sends `Authorization: Bearer`.

## Manual Validation Commands

API:

```sh
cd apps/api && go mod tidy && go test ./... && go build ./...
```

Bot:

```sh
cd apps/bot && go mod tidy && go test ./... && go build ./...
```

Admin:

```sh
cd apps/admin && npm install && npm run build && npm run lint
```

Mini App:

```sh
cd apps/miniapp && npm install && npm run build && npm run lint
```

Docker:

```sh
docker compose -f docker-compose.dev.yml config
docker compose -f docker-compose.dev.yml up -d postgres api
curl http://localhost:8080/health
docker compose -f docker-compose.dev.yml down
```

Expected health response:

```json
{
  "status": "ok",
  "service": "my-tashabbus-api"
}
```

## Troubleshooting

`go: command not found`

Install Go 1.22+ and confirm `go version` works in a new terminal.

`npm: command not found`

Install Node.js 20+ with npm and confirm `node --version` and `npm --version` work.

`docker: command not found`

Install Docker Desktop or Docker Engine and confirm `docker --version` works.

Docker daemon not running

Start Docker Desktop or the Docker service, then rerun `docker compose -f docker-compose.dev.yml config`.

Port 8080 already in use

Stop the process using port 8080 or change the API port in `.env` and Docker Compose settings.

Postgres port conflict

If local PostgreSQL already uses port 5432, change the host port mapping in `docker-compose.dev.yml`, for example `5433:5432`.

`TELEGRAM_BOT_TOKEN` empty in development

This is acceptable when the bot is not running. The bot is optional in Docker Compose and idles safely in development when no token is configured.

For real Telegram Mini App auth, the API container needs the same `TELEGRAM_BOT_TOKEN` that belongs to the bot. Telegram `initData` validation happens in the API. `BOT_TOKEN` is only a legacy fallback.
