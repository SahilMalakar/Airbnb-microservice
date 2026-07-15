package utils

import (
	"fmt"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/sahilmalakar/airbnb-microservice/api-gateway/config"
)

var (
	jwtSecretKey     []byte
	refreshSecretKey []byte
	secretsOnce      sync.Once
)

func MustLoadSecrets() {
	secretsOnce.Do(func() {
		jwtSecretKey = []byte(config.RequireEnvString("ACCESS_KEY_TOKEN"))
		refreshSecretKey = []byte(config.RequireEnvString("REFRESH_TOKEN_SECRET"))
	})
}

// Shared TTL constants — referenced by both token creation here and cookie
// MaxAge in authCookies.go, so the two can't drift out of sync the way the
// old inline literals (30*time.Minute / 3*24*time.Hour, duplicated in two
// files) risked doing.
const (
	AccessTokenTTL  = 30 * time.Minute
	RefreshTokenTTL = 3 * 24 * time.Hour
)

func CreateAccessToken(id int64, email string, name string, familyID string, roles []string, permissions []string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":          id,
		"email":       email,
		"name":        name,
		"familyId":    familyID,
		"roles":       roles,
		"permissions": permissions,
		"exp":         time.Now().Add(AccessTokenTTL).Unix(),
	})

	tokenString, err := token.SignedString(jwtSecretKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}
	return tokenString, nil
}

// NewRefreshFamilyID generates a new refresh-token family identifier, used
// once per login/signup. Every subsequent rotation of that session's
// refresh token keeps the same familyId, which is what lets reuse be
// detected against the whole chain rather than a single token.
func NewRefreshFamilyID() string {
	return uuid.NewString()
}

// NewTokenID generates a fresh jti for a single refresh token issuance —
// called both for a family's first token and on every rotation.
func NewTokenID() string {
	return uuid.NewString()
}

// SignRefreshToken signs a refresh token for an existing (familyID, jti)
// pair. It doesn't generate or persist anything — callers generate jti via
// NewTokenID and record it in a cache.RefreshTokenStore first (see
// service.issueRefreshToken / rotateRefreshToken).
func SignRefreshToken(id int64, familyID string, jti string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":       id,
		"familyId": familyID,
		"jti":      jti,
		"exp":      time.Now().Add(RefreshTokenTTL).Unix(),
	})

	tokenString, err := token.SignedString(refreshSecretKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign refresh token: %w", err)
	}
	return tokenString, nil
}

func VerifyAccessToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return jwtSecretKey, nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to parse access token: %w", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid access token")
	}

	return claims, nil
}

func VerifyRefreshToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return refreshSecretKey, nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to parse refresh token: %w", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid refresh token")
	}

	return claims, nil
}
