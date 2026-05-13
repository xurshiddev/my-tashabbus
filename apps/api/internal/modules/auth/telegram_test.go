package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestTelegramValidatorValidInitData(t *testing.T) {
	now := time.Unix(1700000000, 0)
	initData := signedInitData(t, "token", now, `{"id":123,"username":"dev","first_name":"Dev"}`)
	validator := NewTelegramValidator("token", 24*time.Hour)
	validator.now = func() time.Time { return now.Add(time.Minute) }

	user, err := validator.Validate(initData)
	if err != nil {
		t.Fatalf("validate init data: %v", err)
	}
	if user.ID != 123 {
		t.Fatalf("expected user id 123, got %d", user.ID)
	}
}

func TestTelegramValidatorRejectsModifiedData(t *testing.T) {
	now := time.Unix(1700000000, 0)
	initData := signedInitData(t, "token", now, `{"id":123}`)
	initData = strings.Replace(initData, `%7B%22id%22%3A123%7D`, `%7B%22id%22%3A456%7D`, 1)
	validator := NewTelegramValidator("token", 24*time.Hour)
	validator.now = func() time.Time { return now.Add(time.Minute) }

	if _, err := validator.Validate(initData); err != ErrInvalidInitData {
		t.Fatalf("expected ErrInvalidInitData, got %v", err)
	}
}

func TestTelegramValidatorRejectsMissingHash(t *testing.T) {
	validator := NewTelegramValidator("token", 24*time.Hour)
	if _, err := validator.Validate("auth_date=1700000000&user=%7B%22id%22%3A123%7D"); err != ErrInvalidInitData {
		t.Fatalf("expected ErrInvalidInitData, got %v", err)
	}
}

func TestTelegramValidatorRejectsOldAuthDate(t *testing.T) {
	now := time.Unix(1700000000, 0)
	initData := signedInitData(t, "token", now.Add(-48*time.Hour), `{"id":123}`)
	validator := NewTelegramValidator("token", 24*time.Hour)
	validator.now = func() time.Time { return now }

	if _, err := validator.Validate(initData); err != ErrOldInitData {
		t.Fatalf("expected ErrOldInitData, got %v", err)
	}
}

func TestTelegramValidatorRejectsMalformedUser(t *testing.T) {
	now := time.Unix(1700000000, 0)
	initData := signedInitData(t, "token", now, `{bad`)
	validator := NewTelegramValidator("token", 24*time.Hour)
	validator.now = func() time.Time { return now.Add(time.Minute) }

	if _, err := validator.Validate(initData); err != ErrInvalidInitData {
		t.Fatalf("expected ErrInvalidInitData, got %v", err)
	}
}

func signedInitData(t *testing.T, botToken string, authDate time.Time, user string) string {
	t.Helper()
	values := url.Values{}
	values.Set("auth_date", strconv.FormatInt(authDate.Unix(), 10))
	values.Set("user", user)
	pairs := []string{"auth_date=" + values.Get("auth_date"), "user=" + values.Get("user")}
	sort.Strings(pairs)
	dataCheckString := strings.Join(pairs, "\n")

	secretHMAC := hmac.New(sha256.New, []byte("WebAppData"))
	secretHMAC.Write([]byte(botToken))
	secret := secretHMAC.Sum(nil)

	dataHMAC := hmac.New(sha256.New, secret)
	dataHMAC.Write([]byte(dataCheckString))
	values.Set("hash", hex.EncodeToString(dataHMAC.Sum(nil)))
	return values.Encode()
}
