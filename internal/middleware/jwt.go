package middleware

import (
	"Zoo_List/internal/config"
	"Zoo_List/internal/models"
	"context"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
)

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Missing authorization header", http.StatusUnauthorized)
			return
		}

		if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
			http.Error(w, "Invalid authorization format", http.StatusUnauthorized)
			return
		}

		tokenString := authHeader[7:]
		cfg := config.Load()
		token, err := jwt.ParseWithClaims(tokenString, &models.Claims{}, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
			}
			secret := cfg.JWTSecret
			return []byte(secret), nil
		})
		if err == nil {
			if claims, ok := token.Claims.(*models.Claims); ok {
				ctx := context.WithValue(r.Context(), "user_id", claims.UserID)
				next.ServeHTTP(w, r.WithContext(ctx))
			} else {
				http.Error(w, "Invalid token claims", http.StatusUnauthorized)
				return
			}
		} else {
			fmt.Printf("JWT validation error: %v\n", err)
			http.Error(w, "Error 401", http.StatusUnauthorized)
			return
		}
	})
}
