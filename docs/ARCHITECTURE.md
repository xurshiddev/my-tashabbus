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

## Future Modules

- Auth
- Users
- MFYs
- Streets
- Households
- Assignments
- Dashboard and reports

Stage 0 only creates foundations for these modules. Business logic is postponed until the relevant stage.
