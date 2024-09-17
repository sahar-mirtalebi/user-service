package auth

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/golang-jwt/jwt/v4"
)

func AuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		authHeader := c.Request().Header.Get("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, map[string]string{"error": "You are not logged in. Please provide a valid token."})
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid authorization header format. Ensure you are logged in and provide a valid token."})
		}
		token := parts[1]

		claims, err := ValidateToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid or expired token. Please log in again."})
		}

		userId, ok := claims["UserId"].(float64)
		if !ok {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid token claims. You are not logged in or your session is invalid."})
		}

		c.Set("userId", uint(userId))

		return next(c)
	}
}

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
