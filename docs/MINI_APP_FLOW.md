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
