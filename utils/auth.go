package utils

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/argon2"
)

func CreateHash(password string) (string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	hash := argon2.IDKey([]byte(password), salt, 1, 64*1024, 4, 32)
	encoded := fmt.Sprintf("$argon2id$%s$%s",
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash))
	return encoded, nil
}

func IsValid(encoded, password string) bool {
	parts := strings.Split(encoded, "$")
	if len(parts) != 4 {
		return false
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[2])
	if err != nil {
		return false
	}
	expectedHash, err := base64.RawStdEncoding.DecodeString(parts[3])
	if err != nil {
		return false
	}
	hash := argon2.IDKey([]byte(password), salt, 1, 64*1024, 4, 32)
	return string(hash) == string(expectedHash)
}

func CreateJWT(userID uuid.UUID) (string, error) {
	godotenv.Load()
	secret := os.Getenv("JWT_SECRET")
	claims := jwt.MapClaims{
		"sub": userID.String(),
		"exp": time.Now().Add(time.Hour * 24).Unix(), // 24h
		"iat": time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}
