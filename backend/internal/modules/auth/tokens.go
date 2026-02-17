package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type tokenManager struct {
	secret []byte
}

type jwtClaims struct {
	UserID   string `json:"user_id"`
	TenantID string `json:"tenant_id"`
	Role     string `json:"role"`
	Email    string `json:"email"`
	Type     string `json:"type"`
	jwt.RegisteredClaims
}

func newTokenManager(secret string) *tokenManager {
	return &tokenManager{secret: []byte(secret)}
}

func (m *tokenManager) createAccessToken(user User, ttl time.Duration) (string, time.Time, error) {
	expiresAt := time.Now().UTC().Add(ttl)
	claims := jwtClaims{
		UserID:   user.ID,
		TenantID: user.TenantID,
		Role:     string(user.Role),
		Email:    user.Email,
		Type:     "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
			Subject:   user.ID,
		},
	}
	return m.sign(claims)
}

func (m *tokenManager) createRefreshToken(user User, ttl time.Duration) (string, time.Time, error) {
	expiresAt := time.Now().UTC().Add(ttl)
	claims := jwtClaims{
		UserID:   user.ID,
		TenantID: user.TenantID,
		Role:     string(user.Role),
		Email:    user.Email,
		Type:     "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
			Subject:   user.ID,
		},
	}
	return m.sign(claims)
}

func (m *tokenManager) sign(claims jwtClaims) (string, time.Time, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(m.secret)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("sign token: %w", err)
	}
	return signed, claims.ExpiresAt.Time, nil
}

func (m *tokenManager) parse(token string) (*jwtClaims, error) {
	parsed, err := jwt.ParseWithClaims(token, &jwtClaims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return m.secret, nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := parsed.Claims.(*jwtClaims)
	if !ok || !parsed.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}
	return claims, nil
}
