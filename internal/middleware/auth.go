package middleware

import (
	"crypto/subtle"
	"net/http"
)

const (
	// AuthHeaderKey is the header key for API key authentication
	AuthHeaderKey = "X-API-Key"
)

// APIKeyAuth middleware checks for a valid API key in the request header
func APIKeyAuth(validAPIKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get API key from header
			apiKey := r.Header.Get(AuthHeaderKey)
			if apiKey == "" {
				http.Error(w, "Unauthorized - Missing API key", http.StatusUnauthorized)
				return
			}

			// Use constant time comparison to prevent timing attacks
			if subtle.ConstantTimeCompare([]byte(apiKey), []byte(validAPIKey)) != 1 {
				http.Error(w, "Unauthorized - Invalid API key", http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
