package api

import (
	"net/http"
	"strings"

	"go1f/pkg/auth"
)

func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/signin" {
			next(w, r)
			return
		}

		tokenString := extractToken(r)
		if tokenString == "" {
			writeJSON(w, ErrorResponse{Error: "Authorization token required"}, http.StatusUnauthorized)
			return
		}

		valid, err := auth.ValidateToken(tokenString)
		if err != nil || !valid {
			writeJSON(w, ErrorResponse{Error: "Invalid token"}, http.StatusUnauthorized)
			return
		}

		next(w, r)
	}
}

func extractToken(r *http.Request) string {
	bearerToken := r.Header.Get("Authorization")
	if len(bearerToken) > 7 && strings.HasPrefix(bearerToken, "Bearer ") {
		return bearerToken[7:]
	}

	if cookie, err := r.Cookie("token"); err == nil {
		return cookie.Value
	}

	return ""
}
