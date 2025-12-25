// Copyright (c) 2025 JoeGlenn1213
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package server

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"strings"
)

// AuthConfig holds authentication configuration
type AuthConfig struct {
	Enabled      bool
	Username     string
	PasswordHash string // SHA256 hash with salt: "salt:hash"
}

// AuthMiddleware provides HTTP Basic Authentication
type AuthMiddleware struct {
	username     string
	passwordHash string // format: "salt:hash" or just "hash" for simple mode
	realm        string
}

// NewAuthMiddleware creates a new authentication middleware
func NewAuthMiddleware(username, passwordHash string) *AuthMiddleware {
	return &AuthMiddleware{
		username:     username,
		passwordHash: passwordHash,
		realm:        "LGH Repository Access",
	}
}

// Wrap wraps an http.Handler with authentication
func (a *AuthMiddleware) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Health endpoint is always public
		if r.URL.Path == "/health" {
			next.ServeHTTP(w, r)
			return
		}

		user, pass, ok := r.BasicAuth()
		if !ok {
			a.unauthorized(w)
			return
		}

		// Constant-time comparison for username
		usernameMatch := subtle.ConstantTimeCompare([]byte(user), []byte(a.username)) == 1

		// Check password against hash
		passwordMatch := a.checkPassword(pass)

		if !usernameMatch || !passwordMatch {
			a.unauthorized(w)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// checkPassword verifies password against stored hash
// Supports format: "salt:hash" or simple "hash"
func (a *AuthMiddleware) checkPassword(password string) bool {
	parts := strings.SplitN(a.passwordHash, ":", 2)

	if len(parts) == 2 {
		// Salted hash format: "salt:hash"
		salt := parts[0]
		expectedHash := parts[1]
		actualHash := hashWithSalt(password, salt)
		return subtle.ConstantTimeCompare([]byte(expectedHash), []byte(actualHash)) == 1
	}

	// Simple hash format (not recommended but supported)
	simpleHash := sha256Hex(password)
	return subtle.ConstantTimeCompare([]byte(a.passwordHash), []byte(simpleHash)) == 1
}

// unauthorized sends a 401 response
func (a *AuthMiddleware) unauthorized(w http.ResponseWriter) {
	w.Header().Set("WWW-Authenticate", fmt.Sprintf(`Basic realm="%s", charset="UTF-8"`, a.realm))
	http.Error(w, "Unauthorized", http.StatusUnauthorized)
}

// HashPassword creates a salted SHA256 hash of a password
// Returns format: "salt:hash"
func HashPassword(password string) (string, error) {
	salt, err := generateSalt(16)
	if err != nil {
		return "", err
	}
	hash := hashWithSalt(password, salt)
	return fmt.Sprintf("%s:%s", salt, hash), nil
}

// hashWithSalt creates HMAC-SHA256 hash with salt
func hashWithSalt(password, salt string) string {
	h := hmac.New(sha256.New, []byte(salt))
	h.Write([]byte(password))
	return hex.EncodeToString(h.Sum(nil))
}

// sha256Hex creates a simple SHA256 hash
func sha256Hex(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}

// generateSalt generates a random salt
func generateSalt(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// GenerateToken generates a secure random token
func GenerateToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// LoadAuthFromEnv loads auth configuration from environment variables
// LGH_AUTH_USER and LGH_AUTH_PASSWORD_HASH
func LoadAuthFromEnv() *AuthConfig {
	user := os.Getenv("LGH_AUTH_USER")
	hash := os.Getenv("LGH_AUTH_PASSWORD_HASH")

	if user == "" || hash == "" {
		return &AuthConfig{Enabled: false}
	}

	return &AuthConfig{
		Enabled:      true,
		Username:     user,
		PasswordHash: hash,
	}
}

// ValidatePasswordHash checks if a password matches the hash
func ValidatePasswordHash(password, hash string) bool {
	parts := strings.SplitN(hash, ":", 2)
	if len(parts) == 2 {
		salt := parts[0]
		expectedHash := parts[1]
		actualHash := hashWithSalt(password, salt)
		return subtle.ConstantTimeCompare([]byte(expectedHash), []byte(actualHash)) == 1
	}
	simpleHash := sha256Hex(password)
	return subtle.ConstantTimeCompare([]byte(hash), []byte(simpleHash)) == 1
}
