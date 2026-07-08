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

type UserService struct {
	userRepo   repository.UserRepository
	apiKeyRepo repository.ApiKeyRepository
}

func NewUserService(userRepo repository.UserRepository, apiKeyRepo repository.ApiKeyRepository) *UserService {
	return &UserService{userRepo: userRepo, apiKeyRepo: apiKeyRepo}
}

func (s *UserService) Create(ctx context.Context, req dto.CreateUserRequest) (*model.User, error) {
	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.ErrInternal, "Failed to hash password")
	}

	role := req.Role
	if role == "" {
		role = "user"
	}

	user := &model.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: hash,
		Role:         role,
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		if err == apperrors.ErrConflict {
			return nil, apperrors.Wrap(apperrors.ErrConflict, "Username or email already exists")
		}
		return nil, fmt.Errorf("create user: %w", err)
	}

	return user, nil
}

func (s *UserService) CreateApiKey(ctx context.Context, req dto.CreateApiKeyRequest) (*dto.CreateApiKeyResponse, error) {
	user, err := s.userRepo.GetByID(ctx, req.UserID)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.ErrNotFound, "User not found")
	}
	if !user.IsActive {
		return nil, apperrors.Wrap(apperrors.ErrForbidden, "User is not active")
	}

	rawKey, keyHash, prefix, err := auth.GenerateApiKey()
	if err != nil {
		return nil, apperrors.Wrap(apperrors.ErrInternal, "Failed to generate API key")
	}

	apiKey := &model.ApiKey{
		UserID:    user.ID,
		KeyHash:   keyHash,
		KeyPrefix: prefix,
		Name:      req.Name,
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if req.ExpiryDay != nil && *req.ExpiryDay > 0 {
		exp := time.Now().AddDate(0, 0, *req.ExpiryDay)
		apiKey.ExpiresAt = &exp
	}

	if err := s.apiKeyRepo.Create(ctx, apiKey); err != nil {
		return nil, fmt.Errorf("create api key: %w", err)
	}

	return &dto.CreateApiKeyResponse{
		ID:        apiKey.ID,
		Key:       rawKey,
		KeyPrefix: apiKey.KeyPrefix,
		Name:      apiKey.Name,
	}, nil
}

func (s *UserService) UpdateUser(ctx context.Context, id string, req dto.UpdateUserRequest) (*model.User, error) {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Username != "" {
		user.Username = req.Username
	}
	if req.Email != "" {
		user.Email = req.Email
	}
	if req.Role != "" {
		user.Role = req.Role
	}
	if req.Password != "" {
		hash, err := auth.HashPassword(req.Password)
		if err != nil {
			return nil, apperrors.Wrap(apperrors.ErrInternal, "Failed to hash password")
		}
		user.PasswordHash = hash
	}

	if err := s.userRepo.Update(ctx, user); err != nil {
		if err == apperrors.ErrConflict {
			return nil, apperrors.Wrap(apperrors.ErrConflict, "Username or email already exists")
		}
		return nil, fmt.Errorf("update user: %w", err)
	}

	return user, nil
}

func (s *UserService) GetByID(ctx context.Context, id string) (*model.User, error) {
	return s.userRepo.GetByID(ctx, id)
}

func (s *UserService) ListUsers(ctx context.Context) ([]model.User, error) {
	return s.userRepo.List(ctx)
}

func (s *UserService) ListApiKeys(ctx context.Context) ([]model.ApiKey, error) {
	return s.apiKeyRepo.List(ctx)
}

func (s *UserService) RevokeApiKey(ctx context.Context, id string) error {
	return s.apiKeyRepo.Deactivate(ctx, id)
}
