# Architecture

## Monorepo Overview

The repository contains separate applications for the API, Telegram Bot, Admin Dashboard, and Telegram Mini App. Each app can be developed and deployed independently while sharing project rules and documentation.

## API

`apps/api` is the Go HTTP API. It is the future source of truth for permissions, validation, business rules, persistence, and reporting data.

## Telegram Bot

`apps/bot` is a Go Telegram Bot service. It handles Telegram updates and user-facing bot commands. The bot must not directly own business logic, bypass permissions, or write directly to PostgreSQL. It should call the API or shared service boundaries.

## Admin Dashboard

`apps/admin` is a React + TypeScript dashboard for administrators and MFY leadership workflows.

## Telegram Mini App

`apps/miniapp` is a React + TypeScript Telegram Mini App for future street leader and responsible person workflows.

## PostgreSQL

PostgreSQL stores application data. The API owns database access. Bot and Mini App flows should call the API instead of directly writing to the database.

## Stage 1 Auth Flow

- Admin development login calls `POST /auth/dev-login` in non-production environments and receives a JWT.
- Telegram Mini App authentication calls `POST /auth/telegram` with Telegram WebApp `initData`.
- The API validates Telegram initData with HMAC-SHA256 and the bot token. It does not call Telegram network APIs for this check.
- Authenticated requests use `Authorization: Bearer <token>`.
- API middleware loads the current user, rejects inactive users, and applies role guards.
- The Bot opens the Mini App but does not own authentication, permissions, or database writes.

## Future Modules

- Auth
- Users
- MFYs
- Streets
- Households
- Assignments
- Dashboard and reports

Stage 0 only creates foundations for these modules. Business logic is postponed until the relevant stage.
