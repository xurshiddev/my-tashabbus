# Mini App Flow

The Telegram Mini App will be used by street leaders and responsible persons.

## Stage 0 Flow

- Detect whether the app is opened inside Telegram.
- Safely call `window.Telegram.WebApp.ready()` when available.
- Show placeholder sections for future role-based workflows.

## Future Flow

- Validate Telegram `initData` through the API.
- Show role-based screens.
- Let street leaders review streets and assignments.
- Let responsible persons update assigned household progress.
- Send all writes through the API so validation and permission checks remain centralized.

## Stage 1 Auth Flow

1. Mini App detects `window.Telegram.WebApp`.
2. If opened inside Telegram, it reads `WebApp.initData`.
3. User taps `Authenticate with Telegram`.
4. Mini App sends `init_data` to `POST /auth/telegram`.
5. API validates the Telegram signature and looks up a registered active user by `telegram_id`.
6. If found, API returns a JWT and current user.
7. If not found, API returns `USER_NOT_REGISTERED`.

Browser development mode may use a dev Telegram payload only when `VITE_DEV_TELEGRAM_AUTH=true` and the API allows Telegram auth dev mode.

## Stage 2 My Streets Flow

1. Authenticated Mini App user taps `Load My Streets`.
2. Mini App calls `GET /my/streets`.
3. STREET_LEADER users receive actively assigned streets.
4. MFY_CHAIRMAN users receive streets in their own MFY.
5. Users without Stage 2 street scope see a friendly no-streets message.

Household lists and household updates are not implemented until Stage 3.

## Stage 3 My Households Flow

1. Authenticated Mini App user taps `Mening xonadonlarim`.
2. Mini App calls `GET /my/households`.
3. RESPONSIBLE_PERSON users receive only households assigned to them.
4. STREET_LEADER users receive households in assigned streets.
5. MFY_CHAIRMAN users receive households in their own MFY.
6. Users can update allowed household fields through `PATCH /households/{id}`.
7. The API validates permissions and writes household change logs.

The Mini App does not collect SMS codes and does not interact with official voting systems.
