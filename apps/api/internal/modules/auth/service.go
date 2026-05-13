package auth

import (
	"context"
	"errors"
	"time"

	"github.com/my-tashabbus/api/internal/config"
	"github.com/my-tashabbus/api/internal/modules/users"
)

type Service struct {
	cfg               config.Config
	users             *users.Service
	tokens            *TokenManager
	telegramValidator *TelegramValidator
}

func NewService(cfg config.Config, usersService *users.Service) (*Service, error) {
	if cfg.AppEnv == "production" && (cfg.JWTSecret == "" || cfg.JWTSecret == "dev-secret-change-me") {
		return nil, ErrWeakJWTSecret
	}
	tokenTTL := time.Duration(cfg.JWTAccessTokenTTLMinutes) * time.Minute
	return &Service{
		cfg:               cfg,
		users:             usersService,
		tokens:            NewTokenManager(cfg.JWTSecret, tokenTTL),
		telegramValidator: NewTelegramValidator(cfg.BotToken, 24*time.Hour),
	}, nil
}

func (s *Service) TokenManager() *TokenManager {
	return s.tokens
}

func (s *Service) Me(ctx context.Context, user users.User) users.User {
	return user
}

func (s *Service) DevLogin(ctx context.Context, input DevLoginRequest) (TokenResponse, error) {
	if s.cfg.AppEnv == "production" {
		return TokenResponse{}, ErrDevLoginDisabled
	}
	createInput := users.CreateUserInput{
		FullName:         input.FullName,
		TelegramID:       input.TelegramID,
		TelegramUsername: input.TelegramUsername,
		Role:             input.Role,
	}
	if input.TelegramID != nil {
		existing, err := s.users.GetByTelegramID(ctx, *input.TelegramID)
		if err == nil {
			return s.tokenResponse(existing)
		}
		if !errors.Is(err, users.ErrUserNotFound) {
			return TokenResponse{}, err
		}
	} else {
		existingUsers, err := s.users.List(ctx, 100, 0)
		if err != nil {
			return TokenResponse{}, err
		}
		for _, existing := range existingUsers {
			if existing.FullName == input.FullName && existing.Role == input.Role {
				return s.tokenResponse(existing)
			}
		}
	}
	user, err := s.users.Create(ctx, createInput)
	if err != nil {
		return TokenResponse{}, err
	}
	return s.tokenResponse(user)
}

func (s *Service) TelegramLogin(ctx context.Context, input TelegramAuthRequest) (TokenResponse, error) {
	var telegramUser TelegramUser
	if s.cfg.AppEnv != "production" && s.cfg.TelegramAuthDevMode && input.DevTelegramID != nil {
		telegramUser = TelegramUser{
			ID:        *input.DevTelegramID,
			Username:  input.DevUsername,
			FirstName: input.DevFullName,
		}
	} else {
		user, err := s.telegramValidator.Validate(input.InitData)
		if err != nil {
			return TokenResponse{}, err
		}
		telegramUser = user
	}

	user, err := s.users.GetByTelegramID(ctx, telegramUser.ID)
	if err != nil {
		if errors.Is(err, users.ErrUserNotFound) {
			return TokenResponse{}, ErrUserNotRegistered
		}
		return TokenResponse{}, err
	}
	if !user.IsActive {
		return TokenResponse{}, users.ErrUserInactive
	}
	return s.tokenResponse(user)
}

func (s *Service) tokenResponse(user users.User) (TokenResponse, error) {
	token, expiresIn, err := s.tokens.GenerateAccessToken(user)
	if err != nil {
		return TokenResponse{}, err
	}
	return TokenResponse{
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresIn:   expiresIn,
		User:        user,
	}, nil
}
