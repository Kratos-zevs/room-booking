package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var secret = []byte("secret-key")

type Claims struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

func GenerateToken(role string) (string, error) {
	var userID string

	if role == "admin" {
		userID = "11111111-1111-1111-1111-111111111111"
	} else {
		userID = "22222222-2222-2222-2222-222222222222"
	}

	claims := Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}