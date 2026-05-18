#!/usr/bin/env sh
set -eu

require_env() {
	if [ -z "${1:-}" ]; then
		echo "Missing required env var: $2" >&2
		exit 1
	fi
}

require_tool() {
	if ! command -v "$1" >/dev/null 2>&1; then
		echo "Missing required tool: $1" >&2
		exit 1
	fi
}

require_env "${API_PUBLIC_URL:-}" "API_PUBLIC_URL"
require_env "${ADMIN_TELEGRAM_ID:-}" "ADMIN_TELEGRAM_ID"
require_env "${ADMIN_TELEGRAM_USERNAME:-}" "ADMIN_TELEGRAM_USERNAME"
require_tool curl
require_tool python3

API_PUBLIC_URL="${API_PUBLIC_URL%/}"

LOGIN_RESPONSE="$(curl -sS -w '\n%{http_code}' \
	-X POST "$API_PUBLIC_URL/auth/dev-login" \
	-H "Content-Type: application/json" \
	-d '{"full_name":"Khurshid Admin","role":"SUPER_ADMIN"}')"
LOGIN_STATUS="$(printf '%s' "$LOGIN_RESPONSE" | tail -n 1)"
LOGIN_BODY="$(printf '%s' "$LOGIN_RESPONSE" | sed '$d')"

if [ "$LOGIN_STATUS" -lt 200 ] || [ "$LOGIN_STATUS" -ge 300 ]; then
	echo "Dev login failed with HTTP $LOGIN_STATUS:" >&2
	printf '%s\n' "$LOGIN_BODY" >&2
	exit 1
fi

TOKEN="$(printf '%s' "$LOGIN_BODY" | python3 -c 'import json,sys; d=json.load(sys.stdin).get("data", {}); print(d.get("access_token", ""))')"
USER_ID="$(printf '%s' "$LOGIN_BODY" | python3 -c 'import json,sys; d=json.load(sys.stdin).get("data", {}); print(d.get("user", {}).get("id", ""))')"

if [ -z "$TOKEN" ] || [ -z "$USER_ID" ]; then
	echo "Dev login response did not include token or user id:" >&2
	printf '%s\n' "$LOGIN_BODY" >&2
	exit 1
fi

PATCH_BODY="$(python3 -c 'import json,os; print(json.dumps({"telegram_id": int(os.environ["ADMIN_TELEGRAM_ID"]), "telegram_username": os.environ["ADMIN_TELEGRAM_USERNAME"]}))')"
PATCH_RESPONSE="$(curl -sS -w '\n%{http_code}' \
	-X PATCH "$API_PUBLIC_URL/users/$USER_ID/telegram" \
	-H "Content-Type: application/json" \
	-H "Authorization: Bearer $TOKEN" \
	-d "$PATCH_BODY")"
PATCH_STATUS="$(printf '%s' "$PATCH_RESPONSE" | tail -n 1)"
PATCH_BODY_RESPONSE="$(printf '%s' "$PATCH_RESPONSE" | sed '$d')"

if [ "$PATCH_STATUS" -lt 200 ] || [ "$PATCH_STATUS" -ge 300 ]; then
	echo "Telegram admin binding failed with HTTP $PATCH_STATUS:" >&2
	printf '%s\n' "$PATCH_BODY_RESPONSE" >&2
	exit 1
fi

echo "Telegram admin binding completed successfully."
