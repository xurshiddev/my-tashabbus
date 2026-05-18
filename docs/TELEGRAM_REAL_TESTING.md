# Real Telegram Mini App Testing

This checklist is for local end-to-end testing with a real Telegram Mini App WebView and ngrok tunnels.

My Tashabbus is a one-bot, one-MFY operational tracking app. It must not automate official voting, collect SMS codes, impersonate citizens, or bypass official voting systems.

## Required Model

- One Telegram bot belongs to one MFY.
- Another MFY uses another bot/deployment/env set.
- `TELEGRAM_BOT_TOKEN` is the preferred bot token env.
- `BOT_TOKEN` is a legacy fallback only.
- `APP_MFY_NAME` and `APP_MFY_SLUG` identify this deployment's MFY.
- `APP_MFY_SLUG` is currently config-level only and is not stored in the database.
- `MFY_CHAIRMAN_TELEGRAM_ID` is auto-bootstrapped as `MFY_CHAIRMAN`.
- `ADMIN_TELEGRAM_ID` and `MFY_OWNER_TELEGRAM_ID` are legacy aliases.
- The Mini App authenticates requests with `X-Telegram-Init-Data`; it does not use JWT, `access_token`, localStorage sessions, `/auth/telegram`, or `Authorization: Bearer`.

## Required URLs

- API tunnel: `http://localhost:8080` -> `https://<api-ngrok>.ngrok-free.app`
- Mini App tunnel: `http://localhost:5174` -> `https://<miniapp-ngrok>.ngrok-free.app`

The Mini App uses the API ngrok URL as `VITE_API_BASE_URL`. The bot uses the Mini App ngrok URL as `MINI_APP_URL`.

These must be different URLs unless you have configured a reverse proxy that routes both frontend and API paths:

- `MINI_APP_URL=https://<miniapp-ngrok>.ngrok-free.app`
- `VITE_API_BASE_URL=https://<api-ngrok>.ngrok-free.app`

If `VITE_API_BASE_URL` points to the Mini App/Vite tunnel, API requests will hit the frontend and fail. `GET $VITE_API_BASE_URL/health` must return the API health JSON, not Vite HTML.

## Environment Files

Create local files from examples:

```sh
cp .env.local.test.example .env.local.test
cp .env.local.bot.example .env.local.bot
cp apps/miniapp/.env.local.example apps/miniapp/.env.local
```

Never commit real tokens or real tunnel URLs.

Example `.env.local.bot`:

```sh
TELEGRAM_BOT_TOKEN=123456789:replace-with-real-token
BOT_TOKEN=
MINI_APP_URL=https://your-miniapp-ngrok.ngrok-free.app
API_BASE_URL=https://your-api-ngrok.ngrok-free.app
BOT_ENV=development
MFY_CHAIRMAN_TELEGRAM_ID=123456789
ADMIN_TELEGRAM_ID=
MFY_OWNER_TELEGRAM_ID=
```

Example `.env.local.test`:

```sh
API_PUBLIC_URL=https://your-api-ngrok.ngrok-free.app
MINI_APP_PUBLIC_URL=https://your-miniapp-ngrok.ngrok-free.app
APP_MFY_NAME=My Tashabbus MFY
APP_MFY_SLUG=my-tashabbus-mfy
MFY_CHAIRMAN_TELEGRAM_ID=123456789
ADMIN_TELEGRAM_ID=
MFY_OWNER_TELEGRAM_ID=
```

Example `apps/miniapp/.env.local`:

```sh
VITE_API_BASE_URL=https://your-api-ngrok.ngrok-free.app
VITE_DEV_TELEGRAM_AUTH=false
VITE_ALLOWED_HOSTS=your-miniapp-ngrok.ngrok-free.app
```

## Checklist

1. Start backend:

```sh
unset DATABASE_URL
set -a
source .env.local.bot
source .env.local.test
set +a
export CORS_ALLOWED_ORIGINS="http://localhost:5173,http://localhost:5174,$MINI_APP_PUBLIC_URL"
docker compose -f docker-compose.dev.yml down -v
docker compose -f docker-compose.dev.yml up -d --build postgres api
```

2. Run migrations:

```sh
DATABASE_URL="postgres://my_tashabbus:my_tashabbus_password@localhost:5432/my_tashabbus?sslmode=disable" make migrate-up
```

3. Start API ngrok:

```sh
ngrok http 8080
```

4. Start Mini App:

```sh
cd apps/miniapp
npm run dev -- --host 0.0.0.0
```

5. Start Mini App ngrok:

```sh
ngrok http 5174
```

6. Update `.env.local.test`, `.env.local.bot`, and `apps/miniapp/.env.local` with the current ngrok URLs.

7. Restart the Mini App dev server after changing `apps/miniapp/.env.local`.

8. Restart the API after changing `TELEGRAM_BOT_TOKEN`, `APP_MFY_NAME`, `APP_MFY_SLUG`, `MFY_CHAIRMAN_TELEGRAM_ID`, `ADMIN_TELEGRAM_ID`, `MFY_OWNER_TELEGRAM_ID`, or `CORS_ALLOWED_ORIGINS`.

9. Verify the public API tunnel:

```sh
set -a
source .env.local.test
set +a
curl "$API_PUBLIC_URL/health"
```

10. Verify the API container has the preferred token env:

```sh
docker compose -f docker-compose.dev.yml exec api sh -c 'test -n "$TELEGRAM_BOT_TOKEN" && echo "TELEGRAM_BOT_TOKEN exists" || echo "TELEGRAM_BOT_TOKEN missing"'
```

11. Start the bot:

```sh
cd apps/bot
set -a
source ../../.env.local.bot
set +a
go run ./cmd/bot
```

12. In Telegram:

```text
/myid
/start
```

Send `/start` again after every `MINI_APP_URL` or ngrok change so Telegram uses a fresh WebApp button.

13. Open the Mini App from the newly sent `Mini Appni ochish` button.

Expected development diagnostics:

```text
Telegram WebApp object: present
Telegram initData: present
```

14. Owner bootstrap test:

- If your Telegram ID equals `MFY_CHAIRMAN_TELEGRAM_ID`, opening the Mini App should create or update you as `MFY_CHAIRMAN`.

15. Non-owner test:

- A Telegram user who is not assigned should see `USER_NOT_ASSIGNED` with their Telegram ID and an instruction to contact the MFY chairman.

## Troubleshooting

- Use these checks before debugging roles:

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

Expected results:

- `/health` returns `{"status":"ok","service":"my-tashabbus-api"}`.
- `/miniapp/me` without `X-Telegram-Init-Data` returns JSON with `TELEGRAM_INIT_DATA_MISSING`.
- `OPTIONS /miniapp/me` returns CORS headers including `Access-Control-Allow-Origin` and `Access-Control-Allow-Headers` with `X-Telegram-Init-Data` and `ngrok-skip-browser-warning`.
- `GET /miniapp/me` with fake init data returns `TELEGRAM_INIT_DATA_INVALID` and still includes `Access-Control-Allow-Origin`.
- If `/health` returns an HTML page, `VITE_API_BASE_URL` points to the frontend tunnel instead of the API tunnel.

- `ERR_NGROK_3200` means the ngrok tunnel is offline. Restart that tunnel and update env files.
- API ngrok has zero requests means the Mini App likely points to an old `VITE_API_BASE_URL`, the request was not triggered, or Vite was not restarted after env changes.
- Vite blocked host means `VITE_ALLOWED_HOSTS` is missing the Mini App ngrok host.
- CORS errors mean backend `CORS_ALLOWED_ORIGINS` is missing the Mini App ngrok URL.
- `TELEGRAM_BOT_TOKEN missing` in the API container means Mini App auth cannot validate Telegram initData.
- If `DATABASE_URL` inside the API container uses `localhost`, Docker Compose env is wrong. The API container must use the `postgres` hostname.
- Do not use an old `/start` button after changing bot code or `MINI_APP_URL`; send `/start` again.
- Free ngrok warning pages may require pressing `Visit Site` before Telegram can load the app reliably.
- Vite env changes are read at dev-server startup. Restart `npm run dev` after changing `apps/miniapp/.env.local`.
