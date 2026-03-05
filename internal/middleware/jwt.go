package middleware

import (
	"context"
	"net/http"
	"strings"

	"recipebox-backend-go/internal/service"
	"recipebox-backend-go/internal/utils"
)

type contextKey string

const userIDContextKey contextKey = "userID"

func AuthJWT(authService *service.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := strings.TrimSpace(r.Header.Get("Authorization"))
			if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
				utils.Error(w, http.StatusUnauthorized, "missing bearer token")
				return
			}

			token := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
			if token == "" {
				utils.Error(w, http.StatusUnauthorized, "missing bearer token")
				return
			}

			userID, err := authService.ParseAccessToken(r.Context(), token)
			if err != nil {
				utils.Error(w, http.StatusUnauthorized, "invalid access token")
				return
			}

			ctx := context.WithValue(r.Context(), userIDContextKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func UserIDFromContext(ctx context.Context) (int64, bool) {
	v := ctx.Value(userIDContextKey)
	userID, ok := v.(int64)
	return userID, ok
}
