package utils

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/shiwangeesingh/goFirstProject/config"
)

type Claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// GenerateToken creates a JWT for a user
func GenerateToken(username string) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(config.JWTSecret)
}

// ValidateToken verifies the JWT
func ValidateToken(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return config.JWTSecret, nil
	})

	if err != nil || !token.Valid {
		return nil, err
	}
	return claims, nil
}
