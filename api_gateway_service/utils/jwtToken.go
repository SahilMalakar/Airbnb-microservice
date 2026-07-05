package utils

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/sahilmalakar/airbnb-microservice/api-gateway/config"
)

var jwtSecretKey = []byte(config.GetEnvString("ACCESS_KEY_TOKEN", ""))
var refreshSecretKey = []byte(config.GetEnvString("REFRESH_TOKEN_SECRET", ""))

func CreateAccessToken(id int64, email string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":    id,
		"email": email,
		"exp":   time.Now().Add(time.Minute * 30).Unix(),
	})

	tokenString, err := token.SignedString(jwtSecretKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}
	return tokenString, nil
}

func CreateRefreshToken(id int64) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":  id,
		"exp": time.Now().Add(time.Hour * 24 * 3).Unix(),
	})

	tokenString, err := token.SignedString(refreshSecretKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign refresh token: %w", err)
	}
	return tokenString, nil
}

// VerifyAccessToken validates a token signed with jwtSecretKey (30-min access tokens).
func VerifyAccessToken(tokenString string) (jwt.MapClaims, error) {
	return verifyToken(tokenString, jwtSecretKey)
}

// VerifyRefreshToken validates a token signed with refreshSecretKey (3-day refresh tokens).
func VerifyRefreshToken(tokenString string) (jwt.MapClaims, error) {
	return verifyToken(tokenString, refreshSecretKey)
}

func verifyToken(tokenString string, secret []byte) (jwt.MapClaims, error) {
	claims := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return secret, nil
	})
	if err != nil || !token.Valid {
		return nil, fmt.Errorf("invalid or expired token")
	}
	return claims, nil
}