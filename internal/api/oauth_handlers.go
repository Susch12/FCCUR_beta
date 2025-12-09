package api

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/jesus/FCCUR/internal/auth"
	"github.com/jesus/FCCUR/internal/models"
	"github.com/jesus/FCCUR/internal/storage"
)

// OAuth2Login initiates OAuth2 login flow
func (s *Server) OAuth2Login(w http.ResponseWriter, r *http.Request) {
	if s.oauth2Config == nil || !s.oauth2Config.Enabled {
		respondJSON(w, http.StatusNotImplemented, map[string]string{
			"error": "OAuth2 is not configured",
		})
		return
	}

	// Generate state parameter
	state, err := auth.GenerateState()
	if err != nil {
		log.Printf("Error generating state: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Store state in session cookie (valid for 10 minutes)
	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		Path:     "/",
		MaxAge:   600, // 10 minutes
		HttpOnly: true,
		Secure:   r.TLS != nil,
		SameSite: http.SameSiteLaxMode,
	})

	// Redirect to OAuth2 provider
	authURL := s.oauth2Config.GetAuthURL(state)
	respondJSON(w, http.StatusOK, map[string]string{
		"auth_url": authURL,
	})
}

// OAuth2Callback handles OAuth2 callback
func (s *Server) OAuth2Callback(w http.ResponseWriter, r *http.Request) {
	if s.oauth2Config == nil || !s.oauth2Config.Enabled {
		http.Error(w, "OAuth2 is not configured", http.StatusNotImplemented)
		return
	}

	// Verify state parameter
	stateCookie, err := r.Cookie("oauth_state")
	if err != nil {
		http.Error(w, "Missing state cookie", http.StatusBadRequest)
		return
	}

	state := r.URL.Query().Get("state")
	if state == "" || state != stateCookie.Value {
		http.Error(w, "Invalid state parameter", http.StatusBadRequest)
		return
	}

	// Clear state cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})

	// Get authorization code
	code := r.URL.Query().Get("code")
	if code == "" {
		errorDesc := r.URL.Query().Get("error_description")
		if errorDesc == "" {
			errorDesc = "No authorization code received"
		}
		http.Error(w, errorDesc, http.StatusBadRequest)
		return
	}

	// Exchange code for token
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	token, err := s.oauth2Config.ExchangeCode(ctx, code)
	if err != nil {
		log.Printf("Error exchanging code: %v", err)
		http.Error(w, "Failed to exchange authorization code", http.StatusInternalServerError)
		return
	}

	// Get user info
	userInfo, err := s.oauth2Config.GetUserInfo(ctx, token.AccessToken)
	if err != nil {
		log.Printf("Error getting user info: %v", err)
		http.Error(w, "Failed to get user information", http.StatusInternalServerError)
		return
	}

	email := userInfo.GetEmail()
	fullName := userInfo.GetFullName()

	// Check if user exists
	user, err := s.db.GetUserByEmail(email)
	if err != nil {
		if err == storage.ErrUserNotFound {
			// Create new user with student role by default
			user, err = s.db.CreateUser(email, "", fullName, models.RoleStudent)
			if err != nil {
				log.Printf("Error creating OAuth2 user: %v", err)
				http.Error(w, "Failed to create user", http.StatusInternalServerError)
				return
			}
			log.Printf("Created new OAuth2 user: %s", email)
		} else {
			log.Printf("Error getting user: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	}

	// Check if user is active
	if !user.IsActive {
		respondJSON(w, http.StatusForbidden, map[string]string{
			"error": "Account is deactivated",
		})
		return
	}

	// Generate JWT tokens
	jwtToken, expiresAt, err := s.jwtManager.GenerateToken(user.ID, user.Email, string(user.Role), user.IsAdmin)
	if err != nil {
		log.Printf("Error generating token: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	refreshToken, err := s.jwtManager.GenerateRefreshToken()
	if err != nil {
		log.Printf("Error generating refresh token: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Create session
	ipAddress := getIPAddress(r)
	userAgent := r.UserAgent()
	_, err = s.db.CreateSession(user.ID, jwtToken, refreshToken, ipAddress, userAgent, expiresAt)
	if err != nil {
		log.Printf("Error creating session: %v", err)
	}

	// Update last login
	if err := s.db.UpdateUserLastLogin(user.ID); err != nil {
		log.Printf("Error updating last login: %v", err)
	}

	// Return auth response
	respondJSON(w, http.StatusOK, models.AuthResponse{
		Token:        jwtToken,
		RefreshToken: refreshToken,
		User:         user,
		ExpiresIn:    int64(s.jwtManager.GetAccessExpiration().Seconds()),
	})
}

// OAuth2Config returns OAuth2 configuration status
func (s *Server) OAuth2Config(w http.ResponseWriter, r *http.Request) {
	if s.oauth2Config == nil || !s.oauth2Config.Enabled {
		respondJSON(w, http.StatusOK, map[string]bool{
			"enabled": false,
		})
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"enabled":  true,
		"provider": "microsoft",
	})
}
