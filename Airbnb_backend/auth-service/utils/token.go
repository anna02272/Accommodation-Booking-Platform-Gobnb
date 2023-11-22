package utils

import (
	"auth-service/config"
	"fmt"
	"github.com/golang-jwt/jwt"
	"time"
)

func CreateToken(username string) (string, error) {
	config := config.LoadConfig()
	var secretKey = []byte((config.SecretKey))

	token := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"username": username,
			"exp":      time.Now().Add(time.Hour * 2).Unix(),
		})

	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func VerifyToken(tokenString string) error {
	config := config.LoadConfig()
	var secretKey = []byte((config.SecretKey))
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})
	if err != nil {
		return err
	}
	if !token.Valid {
		return fmt.Errorf("invalid token")
	}

	return nil
}
func ParseTokenClaims(tokenString string) (jwt.MapClaims, error) {
	token, _ := jwt.Parse(tokenString, nil)
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}
	return claims, nil
}
