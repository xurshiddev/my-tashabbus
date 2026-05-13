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

func NewTelegramValidator(botToken string, ttl time.Duration) *TelegramValidator {
	return &TelegramValidator{botToken: botToken, ttl: ttl, now: time.Now}
}

func (v *TelegramValidator) Validate(initData string) (TelegramUser, error) {
	if v.botToken == "" {
		return TelegramUser{}, ErrTelegramTokenNeeded
	}
	values, err := url.ParseQuery(initData)
	if err != nil {
		return TelegramUser{}, ErrInvalidInitData
	}
	hash := values.Get("hash")
	if hash == "" {
		return TelegramUser{}, ErrInvalidInitData
	}
	authDateRaw := values.Get("auth_date")
	if authDateRaw == "" {
		return TelegramUser{}, ErrInvalidInitData
	}
	authDateUnix, err := strconv.ParseInt(authDateRaw, 10, 64)
	if err != nil {
		return TelegramUser{}, ErrInvalidInitData
	}
	authDate := time.Unix(authDateUnix, 0)
	if v.now().Sub(authDate) > v.ttl {
		return TelegramUser{}, ErrOldInitData
	}

	if !v.validHash(values, hash) {
		return TelegramUser{}, ErrInvalidInitData
	}
	userRaw := values.Get("user")
	if userRaw == "" {
		return TelegramUser{}, ErrInvalidInitData
	}
	var user TelegramUser
	if err := json.Unmarshal([]byte(userRaw), &user); err != nil {
		return TelegramUser{}, ErrInvalidInitData
	}
	if user.ID == 0 {
		return TelegramUser{}, ErrInvalidInitData
	}
	return user, nil
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
