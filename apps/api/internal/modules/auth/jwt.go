package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/my-tashabbus/api/internal/modules/users"
)

type TokenManager struct {
	secret []byte
	ttl    time.Duration
	now    func() time.Time
}

func NewTokenManager(secret string, ttl time.Duration) *TokenManager {
	return &TokenManager{
		secret: []byte(secret),
		ttl:    ttl,
		now:    time.Now,
	}
}

func (m *TokenManager) GenerateAccessToken(user users.User) (string, int64, error) {
	now := m.now().UTC()
	expiresAt := now.Add(m.ttl)
	claims := jwt.MapClaims{
		"sub":  user.ID.String(),
		"role": string(user.Role),
		"iat":  now.Unix(),
		"exp":  expiresAt.Unix(),
	}
	if user.TelegramID != nil {
		claims["telegram_id"] = *user.TelegramID
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(m.secret)
	if err != nil {
		return "", 0, fmt.Errorf("sign token: %w", err)
	}
	return signed, int64(m.ttl.Seconds()), nil
}

func (m *TokenManager) ParseAccessToken(tokenValue string) (TokenClaims, error) {
	parsed, err := jwt.Parse(tokenValue, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, ErrInvalidToken
		}
		return m.secret, nil
	}, jwt.WithExpirationRequired())
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return TokenClaims{}, ErrExpiredToken
		}
		return TokenClaims{}, ErrInvalidToken
	}
	claims, ok := parsed.Claims.(jwt.MapClaims)
	if !ok || !parsed.Valid {
		return TokenClaims{}, ErrInvalidToken
	}

	sub, err := claims.GetSubject()
	if err != nil || sub == "" {
		return TokenClaims{}, ErrInvalidToken
	}
	userID, err := uuid.Parse(sub)
	if err != nil {
		return TokenClaims{}, ErrInvalidToken
	}
	roleRaw, ok := claims["role"].(string)
	if !ok || !users.IsValidRole(users.Role(roleRaw)) {
		return TokenClaims{}, ErrInvalidToken
	}
	expiresAt, err := claims.GetExpirationTime()
	if err != nil {
		return TokenClaims{}, ErrInvalidToken
	}
	issuedAt, err := claims.GetIssuedAt()
	if err != nil {
		return TokenClaims{}, ErrInvalidToken
	}

	var telegramID *int64
	if raw, ok := claims["telegram_id"].(float64); ok {
		value := int64(raw)
		telegramID = &value
	}

	return TokenClaims{
		UserID:     userID,
		Role:       users.Role(roleRaw),
		TelegramID: telegramID,
		IssuedAt:   issuedAt.Time,
		ExpiresAt:  expiresAt.Time,
	}, nil
}
