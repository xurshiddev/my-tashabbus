# Architecture

## Monorepo Overview

The repository contains separate applications for the API, Telegram Bot, legacy/development Admin Dashboard, and Telegram Mini App. The active MVP is a single-MFY, bot-first deployment.

One Telegram bot belongs to one MFY. Another MFY should use another bot/deployment/env set.

## API

`apps/api` is the Go HTTP API. It is the source of truth for permissions, validation, business rules, persistence, and reporting data.

At startup, the API ensures the deployment MFY exists from `APP_MFY_NAME`. `APP_MFY_SLUG` is currently config-level metadata only and is not persisted in the `mfys` table.

## Telegram Bot

`apps/bot` is a Go Telegram Bot service. It handles Telegram updates and user-facing bot commands. The bot uses `TELEGRAM_BOT_TOKEN` with `BOT_TOKEN` as a legacy fallback. The bot must not directly own business logic, bypass permissions, or write directly to PostgreSQL.

## Admin Dashboard

`apps/admin` is a React + TypeScript dashboard kept as legacy/development tooling. It is not required for the Mini App MVP.

## Telegram Mini App

`apps/miniapp` is a React + TypeScript Telegram Mini App for MFY chairman, street leader, and responsible person workflows.

## PostgreSQL

PostgreSQL stores application data. The API owns database access. Bot and Mini App flows should call the API instead of directly writing to the database.

## Mini App Auth Flow

- The Mini App reads raw `window.Telegram.WebApp.initData`.
- Every protected Mini App request sends `X-Telegram-Init-Data`.
- The API validates Telegram initData with HMAC-SHA256 and `TELEGRAM_BOT_TOKEN`. It does not call Telegram network APIs for this check.
- The API extracts the Telegram user ID and resolves a local user in the deployment MFY.
- If the Telegram user ID equals the resolved chairman env ID, the API creates or updates that user as `MFY_CHAIRMAN`.
- Chairman env resolution is `MFY_CHAIRMAN_TELEGRAM_ID`, then `ADMIN_TELEGRAM_ID`, then `MFY_OWNER_TELEGRAM_ID`.
- This is not a platform `SUPER_ADMIN`; it is the chairman/owner for the current single-MFY deployment.
- Unknown non-owner users receive `USER_NOT_ASSIGNED`.
- The Mini App does not use JWT, `access_token`, localStorage sessions, `/auth/telegram`, `Authorization`, or `Bearer`.
- The Bot opens the Mini App but does not own authentication, permissions, or database writes.

## Legacy Admin/JWT Flow

- Admin development login calls `POST /auth/dev-login` in non-production environments and receives a JWT.
- Legacy authenticated Admin/API requests use `Authorization: Bearer <token>`.
- These routes remain for development support, but they are not required for the Mini App MVP.

## Stage 2 MFY And Street Modules

- The API owns MFY, street, and street leader assignment business rules.
- `mfys` stores MFY metadata and optional target vote planning context. It does not implement voting logic.
- `streets` stores streets inside an MFY and planned household counts. Household records are deferred to Stage 3.
- `street_leader_assignments` links one active street leader to a street at a time.
- SUPER_ADMIN exists as a legacy/development role for older Admin flows.
- MFY_CHAIRMAN is scoped by `users.mfy_id`.
- STREET_LEADER can view only streets with an active assignment row.
- Bot and Mini App call the API and do not write directly to PostgreSQL.

## Stage 3 Household And Responsible Modules

- The API owns household validation, responsible person assignment rules, and household audit logging.
- `households` stores aggregate household progress only: house number, expected/contacted/reported voted counts, status, notes, and assigned responsible user.
- `responsible_assignments` stores active or historical responsible person ranges for streets.
- `household_change_logs` records important household edits for auditability.
- SUPER_ADMIN exists as a legacy/development role for older Admin flows.
- MFY_CHAIRMAN is scoped to households and assignments in their own MFY.
- STREET_LEADER is scoped to actively assigned streets.
- RESPONSIBLE_PERSON can view and update only households assigned to them.
- The system does not store citizen credentials, SMS codes, passport data, or official voting data.

## Future Modules

- Auth
- Users
- MFYs
- Streets
- Households
- Assignments
- Dashboard and reports

Stage 0 only creates foundations for these modules. Business logic is postponed until the relevant stage.
