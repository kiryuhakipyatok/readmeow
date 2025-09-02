package utils

import (
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

func GenerateJWT(tokenTTL time.Duration, id, secret string) (string, *time.Time, error) {
	now := time.Now()
	t := now.Add(tokenTTL)
	ttl := jwt.NewNumericDate(t)
	iat := jwt.NewNumericDate(now)
	jti := uuid.New().String()
	claims := jwt.RegisteredClaims{
		Subject:   id,
		ExpiresAt: ttl,
		IssuedAt:  iat,
		ID:        jti,
		Issuer:    "readmeow",
		Audience:  []string{"readmeow-users"},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	jwtToken, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", nil, err
	}
	return jwtToken, &t, nil
}
