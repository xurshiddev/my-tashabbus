# Local Setup

This guide explains how to prepare a local machine to validate the Stage 0 foundation for My Tashabbus.

My Tashabbus is not an official voting system. It must not automate voting, collect SMS codes, or vote on behalf of citizens.

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

`BOT_TOKEN` empty in development

This is expected for Stage 0. The bot is optional in Docker Compose and idles safely in development when no token is configured.
