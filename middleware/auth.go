package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/eclipseron/digital-wallet-app/dto"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type ContextString string

var USERID ContextString = "USERID"

func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var response dto.ResponseModel
		response.ID = uuid.New()
		response.Timestamp = time.Now().UTC()
		w.Header().Add("Content-Type", "application/json")

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			w.WriteHeader(http.StatusUnauthorized)
			response.Data = dto.ErrorModel{Message: "missing authorization header"}
			json.NewEncoder(w).Encode(&response)
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			w.WriteHeader(http.StatusUnauthorized)
			response.Data = dto.ErrorModel{Message: "invalid authorization header"}
			json.NewEncoder(w).Encode(&response)
			return
		}

		token := parts[1]
		claims := jwt.MapClaims{}
		parsedToken, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (any, error) {
			return []byte(os.Getenv("JWT_SECRET")), nil
		})
		if err != nil || !parsedToken.Valid {
			w.WriteHeader(http.StatusUnauthorized)
			response.Data = dto.ErrorModel{Message: "invalid or expired token"}
			json.NewEncoder(w).Encode(&response)
			return
		}

		sub, ok := claims["sub"].(string)
		if !ok {
			w.WriteHeader(http.StatusUnauthorized)
			response.Data = dto.ErrorModel{Message: "invalid token payload"}
			json.NewEncoder(w).Encode(&response)
			return
		}

		ctx := context.WithValue(r.Context(), USERID, sub)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
