# Telegram Bot Flow

## Stage 0 Flow

1. User sends `/start`.
2. Bot replies with: `Assalomu alaykum! My Tashabbus tizimiga xush kelibsiz.`
3. Bot sends a `Mini Appni ochish` button when `MINI_APP_URL` is configured.

## Future Flow

- Link Telegram IDs to application users through an API-backed identity flow.
- Open the Mini App with role-aware context.
- Validate user permissions through the API before exposing workflow data.
- Send notifications and reminders for follow-up tasks.

The bot must not bypass API permissions or write directly to the database.

## Stage 1 Auth Boundary

The bot can present the Mini App button, but Telegram identity binding and access checks are owned by the API. A user must be registered and bound to a Telegram ID before Mini App authentication can return an access token.

In Stage 2, the Mini App can show streets assigned by the API. The bot still does not own permissions, street assignments, or database writes.
