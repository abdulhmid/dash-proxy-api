package service

import (
	"context"
	"fmt"
	"time"

	"api-source-proxy/internal/dto"
	"api-source-proxy/internal/model"
	"api-source-proxy/internal/repository"
	"api-source-proxy/pkg/auth"
	apperrors "api-source-proxy/pkg/errors"
)

type AuthService struct {
	userRepo    repository.UserRepository
	apiKeyRepo  repository.ApiKeyRepository
	jwtSecret   string
	jwtExpiry   int
	adminUser   string
	adminPass   string
	adminEmail  string
}

func NewAuthService(userRepo repository.UserRepository, apiKeyRepo repository.ApiKeyRepository, jwtSecret string, jwtExpiry int, adminUser, adminPass, adminEmail string) *AuthService {
	return &AuthService{
		userRepo:   userRepo,
		apiKeyRepo: apiKeyRepo,
		jwtSecret:  jwtSecret,
		jwtExpiry:  jwtExpiry,
		adminUser:  adminUser,
		adminPass:  adminPass,
		adminEmail: adminEmail,
	}
}

func (s *AuthService) InitAdmin(ctx context.Context) error {
	_, err := s.userRepo.GetByUsername(ctx, s.adminUser)
	if err == apperrors.ErrNotFound {
		hash, err := auth.HashPassword(s.adminPass)
		if err != nil {
			return fmt.Errorf("hash admin password: %w", err)
		}
		admin := &model.User{
			Username:     s.adminUser,
			Email:        s.adminEmail,
			PasswordHash: hash,
			Role:         "admin",
			IsActive:     true,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}
		return s.userRepo.Create(ctx, admin)
	}
	return err
}

func (s *AuthService) Login(ctx context.Context, req dto.LoginRequest) (*dto.LoginResponse, error) {
	user, err := s.userRepo.GetByUsername(ctx, req.Username)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.ErrUnauthorized, "Invalid credentials")
	}

	if !user.IsActive {
		return nil, apperrors.Wrap(apperrors.ErrForbidden, "Account is disabled")
	}

	if !auth.CheckPassword(req.Password, user.PasswordHash) {
		return nil, apperrors.Wrap(apperrors.ErrUnauthorized, "Invalid credentials")
	}

	token, err := auth.GenerateJWT(s.jwtSecret, user.ID, user.Username, user.Role, s.jwtExpiry)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.ErrInternal, "Failed to generate token")
	}

	return &dto.LoginResponse{
		Token:    token,
		Username: user.Username,
		Role:     user.Role,
	}, nil
}

func (s *AuthService) ValidateApiKey(ctx context.Context, keyStr string) (*model.User, *model.ApiKey, error) {
	hash := auth.ValidateApiKeyHash(keyStr)

	apiKey, err := s.apiKeyRepo.GetByKeyHash(ctx, hash)
	if err != nil {
		return nil, nil, apperrors.Wrap(apperrors.ErrUnauthorized, "Invalid API key")
	}

	if !apiKey.IsActive {
		return nil, nil, apperrors.Wrap(apperrors.ErrForbidden, "API key is deactivated")
	}

	if apiKey.ExpiresAt != nil && apiKey.ExpiresAt.Before(time.Now()) {
		return nil, nil, apperrors.Wrap(apperrors.ErrForbidden, "API key has expired")
	}

	user, err := s.userRepo.GetByID(ctx, apiKey.UserID)
	if err != nil {
		return nil, nil, apperrors.Wrap(apperrors.ErrUnauthorized, "User not found")
	}

	if !user.IsActive {
		return nil, nil, apperrors.Wrap(apperrors.ErrForbidden, "User account is disabled")
	}

	_ = s.apiKeyRepo.UpdateLastUsed(ctx, apiKey.ID)

	return user, apiKey, nil
}


