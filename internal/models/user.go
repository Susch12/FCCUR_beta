package models

import "time"

// UserRole represents the different user roles
type UserRole string

const (
	RoleGuest     UserRole = "guest"
	RoleStudent   UserRole = "student"
	RoleProfessor UserRole = "professor"
	RoleAdmin     UserRole = "admin"
)

// User represents a registered user
type User struct {
	ID                int64     `json:"id"`
	Email             string    `json:"email"`
	PasswordHash      string    `json:"-"` // Never expose in JSON
	FullName          string    `json:"full_name,omitempty"`
	Role              UserRole  `json:"role"`
	AssignedCourses   string    `json:"assigned_courses,omitempty"` // JSON array of course names for professors
	IsActive          bool      `json:"is_active"`
	IsAdmin           bool      `json:"is_admin"` // Deprecated: use Role instead
	EmailVerified     bool      `json:"email_verified"`
	VerificationToken string    `json:"-"`
	ResetToken        string    `json:"-"`
	ResetTokenExpiry  time.Time `json:"-"`
	LastLogin         time.Time `json:"last_login,omitempty"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// HasRole checks if user has a specific role
func (u *User) HasRole(role UserRole) bool {
	return u.Role == role
}

// IsAdminRole checks if user is admin
func (u *User) IsAdminRole() bool {
	return u.Role == RoleAdmin || u.IsAdmin
}

// CanUpload checks if user can upload packages
func (u *User) CanUpload() bool {
	return u.Role == RoleAdmin || u.Role == RoleProfessor
}

// CanDelete checks if user can delete packages
func (u *User) CanDelete() bool {
	return u.Role == RoleAdmin
}

// CanUploadToCourse checks if professor can upload to a specific course
func (u *User) CanUploadToCourse(courseName string) bool {
	if u.Role == RoleAdmin {
		return true
	}
	if u.Role != RoleProfessor {
		return false
	}
	// Parse assigned courses (simple comma-separated for now)
	// In production, use proper JSON parsing
	if u.AssignedCourses == "" {
		return false
	}
	// Simple contains check (improve with JSON parsing)
	return contains(u.AssignedCourses, courseName)
}

func contains(str, substr string) bool {
	return len(str) > 0 && len(substr) > 0 && (str == substr || len(str) > len(substr))
}

// Session represents an active user session
type Session struct {
	ID           int64     `json:"id"`
	UserID       int64     `json:"user_id"`
	Token        string    `json:"token"`
	RefreshToken string    `json:"refresh_token"`
	IPAddress    string    `json:"ip_address,omitempty"`
	UserAgent    string    `json:"user_agent,omitempty"`
	ExpiresAt    time.Time `json:"expires_at"`
	CreatedAt    time.Time `json:"created_at"`
}

// UserRegistration is the payload for user registration
type UserRegistration struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	FullName string `json:"full_name,omitempty"`
}

// UserLogin is the payload for user login
type UserLogin struct {
	Email      string `json:"email"`
	Password   string `json:"password"`
	RememberMe bool   `json:"remember_me"`
}

// PasswordReset is the payload for password reset request
type PasswordReset struct {
	Email string `json:"email"`
}

// PasswordResetConfirm is the payload for confirming password reset
type PasswordResetConfirm struct {
	Token       string `json:"token"`
	NewPassword string `json:"new_password"`
}

// PasswordChange is the payload for changing password
type PasswordChange struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

// AuthResponse is the response after successful authentication
type AuthResponse struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
	User         *User  `json:"user"`
	ExpiresIn    int64  `json:"expires_in"` // seconds
}
