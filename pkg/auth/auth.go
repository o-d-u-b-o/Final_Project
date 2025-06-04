package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var secretKey []byte

func init() {
	password := os.Getenv("TODO_PASSWORD")
	if password != "" {
		hash := sha256.Sum256([]byte(password))
		secretKey = []byte(hex.EncodeToString(hash[:]))
	}
}

type Claims struct {
	jwt.RegisteredClaims
	PasswordHash string `json:"pwd_hash"`
}

func GenerateToken() (string, error) {
	if len(secretKey) == 0 {

		return "", nil
	}

	password := os.Getenv("TODO_PASSWORD")
	hash := sha256.Sum256([]byte(password))

	claims := &Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(8 * time.Hour)),
		},
		PasswordHash: hex.EncodeToString(hash[:]),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secretKey)
}

func ValidateToken(tokenString string) (bool, error) {
	if len(secretKey) == 0 {
		return true, nil
	}

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})

	if err != nil {
		return false, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		currentPassword := os.Getenv("TODO_PASSWORD")
		currentHash := sha256.Sum256([]byte(currentPassword))
		return claims.PasswordHash == hex.EncodeToString(currentHash[:]), nil
	}

	return false, nil
}
