package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

func GenerateToken(userId uint, email string, tokenType string) (string, error) {
	var expirationTime time.Time
	if tokenType == "login" {
		expirationTime = time.Now().Add(24 * time.Hour)
	} else if tokenType == "reset" {
		expirationTime = time.Now().Add(15 * time.Minute)
	} else {
		return "", fmt.Errorf("invalid token type")
	}

	claims := &jwt.MapClaims{
		"UserId":    userId,
		"Email":     email,
		"exp":       expirationTime.Unix(),
		"TokenType": tokenType,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	jwtKey := []byte("my_secret_key")
	signedToken, err := token.SignedString([]byte(jwtKey))
	if err != nil {
		return "", err
	}
	return signedToken, nil
}

func ValidateToken(tokenStr string) (jwt.MapClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &jwt.MapClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte("my_secret_key"), nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	return *claims, nil
}
