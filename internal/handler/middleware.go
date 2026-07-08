package handler

import (
	"context"
	"net/http"
	"strings"

	"api-source-proxy/internal/model"
	"api-source-proxy/internal/service"
	"api-source-proxy/pkg/auth"
	"api-source-proxy/pkg/response"
)

type contextKey string

const (
	ContextUser   contextKey = "user"
	ContextApiKey contextKey = "api_key"
	ContextClaims contextKey = "claims"
)

func AdminAuth(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
				response.Error(w, r, http.StatusUnauthorized, "Missing or invalid authorization header")
				return
			}

			tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
			claims, err := auth.ValidateJWT(jwtSecret, tokenStr)
			if err != nil {
				response.Error(w, r, http.StatusUnauthorized, "Invalid or expired token")
				return
			}

			if claims.Role != "admin" {
				response.Error(w, r, http.StatusForbidden, "Admin access required")
				return
			}

			ctx := context.WithValue(r.Context(), ContextClaims, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func UserAuth(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
				response.Error(w, r, http.StatusUnauthorized, "Missing or invalid authorization header")
				return
			}

			tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
			claims, err := auth.ValidateJWT(jwtSecret, tokenStr)
			if err != nil {
				response.Error(w, r, http.StatusUnauthorized, "Invalid or expired token")
				return
			}

			ctx := context.WithValue(r.Context(), ContextClaims, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func ApiKeyAuth(authService *service.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			apiKey := r.Header.Get("X-API-Key")
			if apiKey == "" {
				response.Error(w, r, http.StatusUnauthorized, "Missing X-API-Key header")
				return
			}

			user, key, err := authService.ValidateApiKey(r.Context(), apiKey)
			if err != nil {
				response.Error(w, r, http.StatusUnauthorized, "Invalid API key")
				return
			}

			ctx := context.WithValue(r.Context(), ContextUser, user)
			ctx = context.WithValue(ctx, ContextApiKey, key)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetUser(ctx context.Context) *model.User {
	if u, ok := ctx.Value(ContextUser).(*model.User); ok {
		return u
	}
	return nil
}

func GetApiKey(ctx context.Context) *model.ApiKey {
	if k, ok := ctx.Value(ContextApiKey).(*model.ApiKey); ok {
		return k
	}
	return nil
}

func GetClaims(ctx context.Context) *auth.Claims {
	if c, ok := ctx.Value(ContextClaims).(*auth.Claims); ok {
		return c
	}
	return nil
}
