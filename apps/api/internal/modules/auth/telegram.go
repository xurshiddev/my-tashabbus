package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

type TelegramValidator struct {
	botToken string
	ttl      time.Duration
	now      func() time.Time
}

type TelegramValidationDiagnostics struct {
	InitDataPresent bool
	InitDataLength  int
	TelegramUserID  *int64
	AuthDate        *time.Time
	ServerTime      time.Time
	AgeSeconds      *int64
	MaxAgeSeconds   int64
	Result          string
}

func NewTelegramValidator(botToken string, ttl time.Duration) *TelegramValidator {
	return &TelegramValidator{botToken: botToken, ttl: ttl, now: time.Now}
}

func (v *TelegramValidator) Validate(initData string) (TelegramUser, error) {
	user, _, err := v.ValidateWithDiagnostics(initData)
	return user, err
}

func (v *TelegramValidator) ValidateWithDiagnostics(initData string) (TelegramUser, TelegramValidationDiagnostics, error) {
	diagnostics := TelegramValidationDiagnostics{
		InitDataPresent: initData != "",
		InitDataLength:  len(initData),
		ServerTime:      v.now(),
		MaxAgeSeconds:   int64(v.ttl.Seconds()),
		Result:          "started",
	}
	if v.botToken == "" {
		diagnostics.Result = "bot_token_missing"
		return TelegramUser{}, diagnostics, ErrTelegramTokenNeeded
	}
	values, err := url.ParseQuery(initData)
	if err != nil {
		diagnostics.Result = "parse_failed"
		return TelegramUser{}, diagnostics, ErrInvalidInitData
	}
	hash := values.Get("hash")
	if hash == "" {
		diagnostics.Result = "hash_missing"
		return TelegramUser{}, diagnostics, ErrInvalidInitData
	}
	authDateRaw := values.Get("auth_date")
	if authDateRaw == "" {
		diagnostics.Result = "auth_date_missing"
		return TelegramUser{}, diagnostics, ErrInvalidInitData
	}
	authDateUnix, err := strconv.ParseInt(authDateRaw, 10, 64)
	if err != nil {
		diagnostics.Result = "auth_date_invalid"
		return TelegramUser{}, diagnostics, ErrInvalidInitData
	}
	authDate := time.Unix(authDateUnix, 0)
	diagnostics.AuthDate = &authDate
	ageSeconds := int64(diagnostics.ServerTime.Sub(authDate).Seconds())
	diagnostics.AgeSeconds = &ageSeconds
	if diagnostics.ServerTime.Sub(authDate) > v.ttl {
		diagnostics.Result = "expired"
		return TelegramUser{}, diagnostics, ErrOldInitData
	}

	if !v.validHash(values, hash) {
		diagnostics.Result = "hash_invalid"
		return TelegramUser{}, diagnostics, ErrInvalidInitData
	}
	userRaw := values.Get("user")
	if userRaw == "" {
		diagnostics.Result = "user_missing"
		return TelegramUser{}, diagnostics, ErrInvalidInitData
	}
	var user TelegramUser
	if err := json.Unmarshal([]byte(userRaw), &user); err != nil {
		diagnostics.Result = "user_json_invalid"
		return TelegramUser{}, diagnostics, ErrInvalidInitData
	}
	if user.ID == 0 {
		diagnostics.Result = "user_id_missing"
		return TelegramUser{}, diagnostics, ErrInvalidInitData
	}
	diagnostics.TelegramUserID = &user.ID
	diagnostics.Result = "valid"
	return user, diagnostics, nil
}

func (v *TelegramValidator) validHash(values url.Values, hash string) bool {
	pairs := make([]string, 0, len(values))
	for key, value := range values {
		if key == "hash" {
			continue
		}
		pairs = append(pairs, key+"="+value[0])
	}
	sort.Strings(pairs)
	dataCheckString := strings.Join(pairs, "\n")

	secretHMAC := hmac.New(sha256.New, []byte("WebAppData"))
	secretHMAC.Write([]byte(v.botToken))
	secret := secretHMAC.Sum(nil)

	dataHMAC := hmac.New(sha256.New, secret)
	dataHMAC.Write([]byte(dataCheckString))
	expected := hex.EncodeToString(dataHMAC.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(hash))
}
