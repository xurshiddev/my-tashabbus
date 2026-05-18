# Mini App Flow

The Telegram Mini App is the active MVP interface for My Tashabbus. The Admin app and JWT login remain only as legacy/development tooling.

## Architecture

- One Telegram bot serves one MFY.
- A different MFY should run a separate bot/deployment with its own environment values.
- The deployment MFY is configured by `APP_MFY_NAME` and `APP_MFY_SLUG`.
- `APP_MFY_SLUG` is currently config-level metadata only; the `mfys` table does not persist a slug column yet.
- The MFY chairman/owner is configured by `MFY_CHAIRMAN_TELEGRAM_ID`.
- `ADMIN_TELEGRAM_ID` and `MFY_OWNER_TELEGRAM_ID` are legacy aliases, resolved after `MFY_CHAIRMAN_TELEGRAM_ID`.
- This is not a platform `SUPER_ADMIN`; it is the chairman for the current single-MFY deployment.

## Authentication

The Mini App does not use JWT, `access_token`, localStorage sessions, `/auth/telegram`, `Authorization`, or `Bearer` headers.

For every protected request, the Mini App sends the raw Telegram WebApp init data exactly as Telegram provided it:

```http
X-Telegram-Init-Data: <raw Telegram WebApp initData>
```

The backend verifies this header on every Mini App request using `TELEGRAM_BOT_TOKEN`. `BOT_TOKEN` exists only as a legacy fallback.

The Mini App must not use `initDataUnsafe` for authentication. It may use safe diagnostic metadata, but it must never display raw `initData`, bot tokens, JWTs, or secrets.

## Startup Flow

1. Telegram opens the Mini App from the bot `web_app` button.
2. The Mini App checks `window.Telegram.WebApp`.
3. The Mini App reads `window.Telegram.WebApp.initData`.
4. The Mini App calls `GET /miniapp/me` with `X-Telegram-Init-Data`.
5. The API validates Telegram initData, resolves the local user, and returns the current user plus deployment MFY.
6. If the Telegram user ID matches `MFY_CHAIRMAN_TELEGRAM_ID`, the backend bootstraps that user as `MFY_CHAIRMAN`.
7. If the user is not assigned, the API returns `USER_NOT_ASSIGNED`.

## Current Smoke-Test UI

The current Mini App UI is intentionally minimal:

- It calls only `GET /miniapp/me`.
- It shows the MFY name, current user, and role.
- It includes a `Qayta tekshirish` button.
- It does not show street, household, assignment, progress, dashboard, or statistics tools yet.

## User Messages

If `initData` is missing:

```text
Mini App Telegram ichidan ochilishi kerak.
```

If the user is not assigned:

```text
Siz hali ushbu MFY tizimiga biriktirilmagansiz. Telegram ID: <id>. MFY raisiga murojaat qiling.
```

If initData is invalid:

```text
Telegram autentifikatsiyasi tasdiqlanmadi. Bot token yoki sozlamalarni tekshiring.
```

If initData is expired:

```text
Sessiya tugagan. Mini App'ni Telegramdan qayta oching.
```

## Existing Mini App Endpoints

These endpoint names exist in the API and are documented for later MVP workflow wiring, but the current smoke-test UI only calls `GET /miniapp/me`:

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

Future cleanup may normalize names such as `/miniapp/my-households`, but no endpoint renaming is part of Cleanup-1.

## Role Flows

`MFY_CHAIRMAN`:

- Sees the deployment MFY context.
- Can work inside the deployment MFY through Mini App endpoints that support chairman scope.
- Missing MVP work: street creation and street leader assignment by Telegram ID.

## Future Role Assignment

Do not implement this before the role-only smoke test is stable:

- Users send `/myid` to the bot.
- Users give their Telegram ID to the MFY chairman.
- The chairman later assigns `STREET_LEADER` or `RESPONSIBLE_PERSON` by Telegram ID.
- A later improvement can show "Yangi kirgan foydalanuvchilar" as a selectable list.
- The bot/Mini App should not assume access to Telegram contact lists.

`STREET_LEADER`:

- Sees assigned streets.
- Can view and manage households inside assigned streets where endpoints allow it.
- Can assign responsible persons within assigned streets where endpoints allow it.

`RESPONSIBLE_PERSON`:

- Sees assigned households.
- Can update household status/counts/notes where endpoints allow it.

The Mini App must not collect SMS codes, store citizen credentials, automate official voting, or interact with official voting systems.
