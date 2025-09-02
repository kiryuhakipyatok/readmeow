package utils

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func GenetateJWT(tokenTTL time.Duration, id, secret string) (string, *time.Time, error) {
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

func GetIdFromLocals(locals any) (string, error) {
	token, ok := locals.(*jwt.Token)
	if !ok {
		return "", errors.New("invalid local data")
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", errors.New("invalid claims data")
	}
	return claims["sub"].(string), nil
}
