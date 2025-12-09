package api

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/jesus/FCCUR/internal/auth"
	"github.com/jesus/FCCUR/internal/models"
	"github.com/jesus/FCCUR/internal/storage"
)

// Register handles user registration
func (s *Server) Register(w http.ResponseWriter, r *http.Request) {
	var req models.UserRegistration
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Validate email
	if err := auth.ValidateEmail(req.Email); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid email address"})
		return
	}

	// Validate password
	if err := auth.ValidatePassword(req.Password); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Password must be at least 8 characters with uppercase, lowercase, and digit",
		})
		return
	}

	// Check if email already exists
	existingUser, err := s.db.GetUserByEmail(req.Email)
	if err == nil && existingUser != nil {
		respondJSON(w, http.StatusConflict, map[string]string{"error": "Email already registered"})
		return
	}

	// Hash password
	passwordHash, err := auth.HashPassword(req.Password)
	if err != nil {
		log.Printf("Error hashing password: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Create user with default student role
	user, err := s.db.CreateUser(req.Email, passwordHash, req.FullName, models.RoleStudent)
	if err != nil {
		log.Printf("Error creating user: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Generate tokens
	token, expiresAt, err := s.jwtManager.GenerateToken(user.ID, user.Email, string(user.Role), user.IsAdmin)
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
	_, err = s.db.CreateSession(user.ID, token, refreshToken, ipAddress, userAgent, expiresAt)
	if err != nil {
		log.Printf("Error creating session: %v", err)
	}

	// Return auth response
	respondJSON(w, http.StatusCreated, models.AuthResponse{
		Token:        token,
		RefreshToken: refreshToken,
		User:         user,
		ExpiresIn:    int64(s.jwtManager.GetAccessExpiration().Seconds()),
	})
}

// Login handles user login
func (s *Server) Login(w http.ResponseWriter, r *http.Request) {
	var req models.UserLogin
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Get user by email
	user, err := s.db.GetUserByEmail(req.Email)
	if err != nil {
		if err == storage.ErrUserNotFound {
			respondJSON(w, http.StatusUnauthorized, map[string]string{"error": "Invalid credentials"})
			return
		}
		log.Printf("Error getting user: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Check if user is active
	if !user.IsActive {
		respondJSON(w, http.StatusForbidden, map[string]string{"error": "Account is deactivated"})
		return
	}

	// Verify password
	if !auth.CheckPassword(req.Password, user.PasswordHash) {
		respondJSON(w, http.StatusUnauthorized, map[string]string{"error": "Invalid credentials"})
		return
	}

	// Generate tokens
	token, expiresAt, err := s.jwtManager.GenerateToken(user.ID, user.Email, string(user.Role), user.IsAdmin)
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
	_, err = s.db.CreateSession(user.ID, token, refreshToken, ipAddress, userAgent, expiresAt)
	if err != nil {
		log.Printf("Error creating session: %v", err)
	}

	// Update last login
	if err := s.db.UpdateUserLastLogin(user.ID); err != nil {
		log.Printf("Error updating last login: %v", err)
	}

	// Return auth response
	respondJSON(w, http.StatusOK, models.AuthResponse{
		Token:        token,
		RefreshToken: refreshToken,
		User:         user,
		ExpiresIn:    int64(s.jwtManager.GetAccessExpiration().Seconds()),
	})
}

// Logout handles user logout
func (s *Server) Logout(w http.ResponseWriter, r *http.Request) {
	token := extractToken(r)
	if token == "" {
		respondJSON(w, http.StatusOK, map[string]string{"message": "Logged out"})
		return
	}

	// Delete session
	if err := s.db.DeleteSession(token); err != nil {
		log.Printf("Error deleting session: %v", err)
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Logged out successfully"})
}

// LogoutAll handles logging out from all devices
func (s *Server) LogoutAll(w http.ResponseWriter, r *http.Request) {
	// Get current user from context
	claims, err := s.getCurrentUser(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Delete all user sessions
	if err := s.db.DeleteUserSessions(claims.UserID); err != nil {
		log.Printf("Error deleting user sessions: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Logged out from all devices"})
}

// RefreshToken handles token refresh
func (s *Server) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Get session by refresh token
	session, err := s.db.GetSessionByRefreshToken(req.RefreshToken)
	if err != nil {
		if err == storage.ErrSessionNotFound {
			respondJSON(w, http.StatusUnauthorized, map[string]string{"error": "Invalid refresh token"})
			return
		}
		log.Printf("Error getting session: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Get user
	user, err := s.db.GetUserByID(session.UserID)
	if err != nil {
		log.Printf("Error getting user: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Generate new tokens
	token, expiresAt, err := s.jwtManager.GenerateToken(user.ID, user.Email, string(user.Role), user.IsAdmin)
	if err != nil {
		log.Printf("Error generating token: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	newRefreshToken, err := s.jwtManager.GenerateRefreshToken()
	if err != nil {
		log.Printf("Error generating refresh token: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Delete old session and create new one
	s.db.DeleteSession(session.Token)
	ipAddress := getIPAddress(r)
	userAgent := r.UserAgent()
	_, err = s.db.CreateSession(user.ID, token, newRefreshToken, ipAddress, userAgent, expiresAt)
	if err != nil {
		log.Printf("Error creating session: %v", err)
	}

	// Return new tokens
	respondJSON(w, http.StatusOK, models.AuthResponse{
		Token:        token,
		RefreshToken: newRefreshToken,
		User:         user,
		ExpiresIn:    int64(s.jwtManager.GetAccessExpiration().Seconds()),
	})
}

// RequestPasswordReset handles password reset requests
func (s *Server) RequestPasswordReset(w http.ResponseWriter, r *http.Request) {
	var req models.PasswordReset
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Get user
	user, err := s.db.GetUserByEmail(req.Email)
	if err != nil {
		// Don't reveal if email exists
		respondJSON(w, http.StatusOK, map[string]string{
			"message": "If the email exists, a reset link will be sent",
		})
		return
	}

	// Generate reset token
	resetToken, err := auth.GenerateVerificationToken()
	if err != nil {
		log.Printf("Error generating reset token: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Set reset token (expires in 1 hour)
	expiry := time.Now().Add(1 * time.Hour)
	if err := s.db.SetUserResetToken(user.ID, resetToken, expiry); err != nil {
		log.Printf("Error setting reset token: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// In production, send email with reset link
	// For now, log the token (REMOVE IN PRODUCTION)
	log.Printf("Password reset token for %s: %s", user.Email, resetToken)

	respondJSON(w, http.StatusOK, map[string]string{
		"message": "If the email exists, a reset link will be sent",
	})
}

// ResetPassword handles password reset confirmation
func (s *Server) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var req models.PasswordResetConfirm
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Validate new password
	if err := auth.ValidatePassword(req.NewPassword); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Password must be at least 8 characters with uppercase, lowercase, and digit",
		})
		return
	}

	// Get user by reset token
	user, err := s.db.GetUserByResetToken(req.Token)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid or expired reset token"})
		return
	}

	// Hash new password
	passwordHash, err := auth.HashPassword(req.NewPassword)
	if err != nil {
		log.Printf("Error hashing password: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Update password
	if err := s.db.UpdateUserPassword(user.ID, passwordHash); err != nil {
		log.Printf("Error updating password: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Invalidate all sessions
	s.db.DeleteUserSessions(user.ID)

	respondJSON(w, http.StatusOK, map[string]string{"message": "Password reset successfully"})
}

// ChangePassword handles password change (when logged in)
func (s *Server) ChangePassword(w http.ResponseWriter, r *http.Request) {
	// Get current user
	claims, err := s.getCurrentUser(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req models.PasswordChange
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Get user
	user, err := s.db.GetUserByID(claims.UserID)
	if err != nil {
		log.Printf("Error getting user: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Verify current password
	if !auth.CheckPassword(req.CurrentPassword, user.PasswordHash) {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "Current password is incorrect"})
		return
	}

	// Validate new password
	if err := auth.ValidatePassword(req.NewPassword); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Password must be at least 8 characters with uppercase, lowercase, and digit",
		})
		return
	}

	// Hash new password
	passwordHash, err := auth.HashPassword(req.NewPassword)
	if err != nil {
		log.Printf("Error hashing password: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Update password
	if err := s.db.UpdateUserPassword(user.ID, passwordHash); err != nil {
		log.Printf("Error updating password: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Password changed successfully"})
}

// GetCurrentUser returns the current authenticated user
func (s *Server) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	claims, err := s.getCurrentUser(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	user, err := s.db.GetUserByID(claims.UserID)
	if err != nil {
		log.Printf("Error getting user: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusOK, user)
}

// Helper functions

func extractToken(r *http.Request) string {
	bearerToken := r.Header.Get("Authorization")
	if len(strings.Split(bearerToken, " ")) == 2 {
		return strings.Split(bearerToken, " ")[1]
	}
	return ""
}

func (s *Server) getCurrentUser(r *http.Request) (*auth.JWTClaims, error) {
	token := extractToken(r)
	if token == "" {
		return nil, auth.ErrInvalidToken
	}

	return s.jwtManager.ValidateToken(token)
}

func getIPAddress(r *http.Request) string {
	// Check X-Forwarded-For header first
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		return strings.Split(forwarded, ",")[0]
	}
	return r.RemoteAddr
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
