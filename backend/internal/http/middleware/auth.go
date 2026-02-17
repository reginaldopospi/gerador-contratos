package middleware

import (
	"context"
	"net/http"
	"strings"

	"gerador-contratos/backend/internal/http/httpx"
	"gerador-contratos/backend/internal/modules/auth"
	"gerador-contratos/backend/internal/modules/common"
)

type authContextKey string

const claimsContextKey authContextKey = "auth_claims"

type AuthValidator interface {
	ValidateAccessToken(token string) (*auth.AuthClaims, error)
}

func RequireAuth(validator AuthValidator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authorization := strings.TrimSpace(r.Header.Get("Authorization"))
			parts := strings.Fields(authorization)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				httpx.WriteError(w, common.NewUnauthorized("missing_token", "token de acesso ausente"))
				return
			}

			token := strings.TrimSpace(parts[1])
			if token == "" {
				httpx.WriteError(w, common.NewUnauthorized("missing_token", "token de acesso ausente"))
				return
			}

			claims, err := validator.ValidateAccessToken(token)
			if err != nil {
				httpx.WriteError(w, err)
				return
			}

			ctx := context.WithValue(r.Context(), claimsContextKey, *claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetClaims(ctx context.Context) (auth.AuthClaims, bool) {
	claims, ok := ctx.Value(claimsContextKey).(auth.AuthClaims)
	return claims, ok
}
