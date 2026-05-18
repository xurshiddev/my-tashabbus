# Telegram Bot Flow

## MVP Flow

1. User sends `/start`.
2. Bot replies with a role-aware welcome message for this MFY deployment.
3. Bot sends a Telegram `web_app` button labeled `Mini Appni ochish` when `MINI_APP_URL` is configured.
4. User opens the Mini App from the newest `web_app` button.
5. The Mini App authenticates API requests with `X-Telegram-Init-Data`.

The Mini App button must use the Telegram WebApp payload:

```json
{
  "text": "Mini Appni ochish",
  "web_app": {
    "url": "https://miniapp.example"
  }
}
```

It must not be a normal root-level `url` button, because normal URL buttons open outside the Telegram Mini App context and do not provide `initData`.

Send `/start` again after changing `MINI_APP_URL` or an ngrok tunnel. Old Telegram messages keep the old button URL.

## Single-MFY Config

- One bot belongs to one MFY.
- Another MFY uses another bot/deployment/env set.
- The preferred token env is `TELEGRAM_BOT_TOKEN`.
- `BOT_TOKEN` is only a legacy fallback.
- `MFY_CHAIRMAN_TELEGRAM_ID` identifies the Telegram user who should be bootstrapped as `MFY_CHAIRMAN`.
- `ADMIN_TELEGRAM_ID` and `MFY_OWNER_TELEGRAM_ID` are supported as legacy aliases.
- `MINI_APP_URL` must point to the current Mini App public URL.

## Auth Boundary

The bot can present the Mini App button and the `/myid` helper, but permissions and business rules are owned by the API.

The Mini App does not use JWT, `access_token`, localStorage sessions, `/auth/telegram`, or `Authorization: Bearer`. It sends raw Telegram WebApp init data on every protected request:

```http
X-Telegram-Init-Data: <raw Telegram WebApp initData>
```

The API verifies Telegram initData using `TELEGRAM_BOT_TOKEN`, resolves the Telegram user, bootstraps the resolved chairman Telegram ID as `MFY_CHAIRMAN`, and enforces role permissions.

The bot must not bypass API permissions or write directly to the database.

For local testing, `/myid` returns the sender's Telegram ID. Non-owner users who are not assigned should send that ID to the MFY chairman.
