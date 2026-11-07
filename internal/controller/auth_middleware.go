package controller

import (
	"context"
	"net/http"
	"strings"

	"redditclone/internal/controller/dto"
	"redditclone/internal/domain"
	"redditclone/internal/pkg/auth"
)

type contextKey string

const userContextKey contextKey = "user"

func UserFromContext(ctx context.Context) (*domain.User, bool) {
	u, ok := ctx.Value(userContextKey).(*domain.User)
	return u, ok
}

func AuthMiddleware(tokenManager *auth.TokenManager) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				writeJSON(w, http.StatusUnauthorized, dto.ErrorResponse{Message: "unauthorized"})
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				writeJSON(w, http.StatusUnauthorized, dto.ErrorResponse{Message: "unauthorized"})
				return
			}

			user, err := tokenManager.ParseToken(parts[1])
			if err != nil {
				writeJSON(w, http.StatusUnauthorized, dto.ErrorResponse{Message: "unauthorized"})
				return
			}

			ctx := context.WithValue(r.Context(), userContextKey, user)
			next(w, r.WithContext(ctx))
		}
	}
}
