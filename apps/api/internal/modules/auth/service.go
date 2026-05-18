package auth

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/my-tashabbus/api/internal/config"
	"github.com/my-tashabbus/api/internal/modules/users"
)

type Service struct {
	cfg               config.Config
	users             *users.Service
	tokens            *TokenManager
	telegramValidator *TelegramValidator
	deploymentMFY     *DeploymentMFY
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

func (s *Service) AppEnv() string {
	return s.cfg.AppEnv
}

func (s *Service) SetDeploymentMFY(mfy DeploymentMFY) {
	s.deploymentMFY = &mfy
}

func (s *Service) DeploymentMFY() (DeploymentMFY, bool) {
	if s.deploymentMFY == nil {
		return DeploymentMFY{}, false
	}
	return *s.deploymentMFY, true
}

func (s *Service) TelegramDiagnostics(initData string) TelegramValidationDiagnostics {
	_, diagnostics, _ := s.telegramValidator.ValidateWithDiagnostics(initData)
	return diagnostics
}

func (s *Service) UserFromTelegramInitData(ctx context.Context, initData string) (users.User, TelegramValidationDiagnostics, error) {
	telegramUser, diagnostics, err := s.telegramValidator.ValidateWithDiagnostics(initData)
	if err != nil {
		return users.User{}, diagnostics, err
	}
	user, err := s.users.GetByTelegramID(ctx, telegramUser.ID)
	if err != nil {
		if errors.Is(err, users.ErrUserNotFound) {
			owner, ownerErr := s.bootstrapMFYOwner(ctx, telegramUser)
			if ownerErr == nil {
				return owner, diagnostics, nil
			}
			return users.User{}, diagnostics, ownerErr
		}
		return users.User{}, diagnostics, err
	}
	if s.cfg.MFYOwnerTelegramID != 0 && telegramUser.ID == s.cfg.MFYOwnerTelegramID {
		owner, err := s.ensureOwnerUser(ctx, user, telegramUser)
		if err != nil {
			return users.User{}, diagnostics, err
		}
		user = owner
	}
	if !user.IsActive {
		return users.User{}, diagnostics, users.ErrUserInactive
	}
	return user, diagnostics, nil
}

func (s *Service) bootstrapMFYOwner(ctx context.Context, telegramUser TelegramUser) (users.User, error) {
	if s.cfg.MFYOwnerTelegramID == 0 || telegramUser.ID != s.cfg.MFYOwnerTelegramID {
		return users.User{}, ErrUserNotAssigned
	}
	if s.deploymentMFY == nil {
		return users.User{}, ErrMFYContextMissing
	}
	fullName := telegramDisplayName(telegramUser)
	username := optionalString(telegramUser.Username)
	user, err := s.users.Create(ctx, users.CreateUserInput{
		FullName:         fullName,
		TelegramID:       &telegramUser.ID,
		TelegramUsername: username,
		Role:             users.RoleMFYChairman,
		MFYID:            &s.deploymentMFY.ID,
	})
	if errors.Is(err, users.ErrTelegramIDConflict) {
		existing, getErr := s.users.GetByTelegramID(ctx, telegramUser.ID)
		if getErr != nil {
			return users.User{}, getErr
		}
		return s.ensureOwnerUser(ctx, existing, telegramUser)
	}
	return user, err
}

func (s *Service) ensureOwnerUser(ctx context.Context, user users.User, telegramUser TelegramUser) (users.User, error) {
	if s.deploymentMFY == nil {
		return users.User{}, ErrMFYContextMissing
	}
	needsUpdate := user.Role != users.RoleMFYChairman || user.MFYID == nil || *user.MFYID != s.deploymentMFY.ID || !user.IsActive
	if !needsUpdate {
		return user, nil
	}
	return s.users.Update(ctx, user.ID, users.UpdateUserInput{
		FullName: firstNonEmpty(user.FullName, telegramDisplayName(telegramUser)),
		Phone:    user.Phone,
		Role:     users.RoleMFYChairman,
		MFYID:    &s.deploymentMFY.ID,
		IsActive: true,
	})
}

func telegramDisplayName(user TelegramUser) string {
	name := strings.TrimSpace(strings.Join([]string{user.FirstName, user.LastName}, " "))
	if name != "" {
		return name
	}
	if strings.TrimSpace(user.Username) != "" {
		return "@" + strings.TrimSpace(user.Username)
	}
	return "MFY Chairman"
}

func optionalString(value string) *string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return "MFY Chairman"
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
