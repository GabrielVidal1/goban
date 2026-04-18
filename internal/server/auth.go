package server

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"sync"
)

var (
	authToken string
	authMu    sync.RWMutex
)

// GenerateAuthToken creates a random 32-byte hex token.
func GenerateAuthToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// GetAuthToken returns the current auth token (thread-safe).
func GetAuthToken() string {
	authMu.RLock()
	defer authMu.RUnlock()
	return authToken
}

// SetAuthToken sets the auth token (thread-safe).
func SetAuthToken(token string) {
	authMu.Lock()
	defer authMu.Unlock()
	authToken = token
}

// AuthMiddleware returns an http.HandlerFunc wrapper that requires a valid token.
// The token is checked from: ?token= in query params, or X-Auth-Token header.
func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.URL.Query().Get("token")
		if token == "" {
			token = r.Header.Get("X-Auth-Token")
		}

		if token == "" || token != GetAuthToken() {
			writeError(w, http.StatusUnauthorized, "unauthorized: invalid or missing authentication token")
			return
		}

		next.ServeHTTP(w, r)
	}
}
