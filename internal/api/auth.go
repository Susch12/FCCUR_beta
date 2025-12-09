package api

import (
	"crypto/subtle"
	"net/http"
)

// AuthConfig holds authentication configuration
type AuthConfig struct {
	Username string
	Password string
	Enabled  bool
}

// withAuth wraps a handler with HTTP Basic Authentication
func (s *Server) withAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Skip auth if not configured
		if !s.authConfig.Enabled {
			next(w, r)
			return
		}

		// Get credentials from request
		username, password, ok := r.BasicAuth()
		if !ok {
			w.Header().Set("WWW-Authenticate", `Basic realm="FCCUR Upload"`)
			http.Error(w, "Authentication required", http.StatusUnauthorized)
			return
		}

		// Constant-time comparison to prevent timing attacks
		usernameMatch := subtle.ConstantTimeCompare([]byte(username), []byte(s.authConfig.Username)) == 1
		passwordMatch := subtle.ConstantTimeCompare([]byte(password), []byte(s.authConfig.Password)) == 1

		if !usernameMatch || !passwordMatch {
			w.Header().Set("WWW-Authenticate", `Basic realm="FCCUR Upload"`)
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}

		// Authentication successful
		next(w, r)
	}
}
