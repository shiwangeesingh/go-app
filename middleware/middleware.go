package middleware

import (
	"net/http"
	"strings"

	"github.com/shiwangeesingh/go-app/utils"
)

// AuthMiddleware validates JWT from Authorization header
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get the Authorization header
		authHeader := r.Header.Get("Authorization")

		// Check if Authorization header is present
		if authHeader == "" {
			http.Error(w, "🚫 Unauthorized: No token provided", http.StatusUnauthorized)
			return
		}

		// Extract token (expecting format: "Bearer <token>")
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "🚫 Unauthorized: Invalid token format", http.StatusUnauthorized)
			return
		}

		tokenString := parts[1]

		// Validate token
		claims, err := utils.ValidateToken(tokenString)
		if err != nil {
			http.Error(w, "🚫 Unauthorized: Invalid or expired token", http.StatusUnauthorized)
			return
		}

		// Attach user data to request context
		r.Header.Set("X-User", claims.Username)

		// Continue to next handler
		next.ServeHTTP(w, r)
	})
}
