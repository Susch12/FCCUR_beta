package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidToken     = errors.New("invalid token")
	ErrTokenExpired     = errors.New("token expired")
	ErrWeakPassword     = errors.New("password does not meet requirements")
	ErrInvalidEmail     = errors.New("invalid email address")
)

// JWTClaims represents the JWT payload (simplified, no external deps)
type JWTClaims struct {
	UserID    int64  `json:"user_id"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	IsAdmin   bool   `json:"is_admin"` // Deprecated: use Role instead
	ExpiresAt int64  `json:"exp"`
	IssuedAt  int64  `json:"iat"`
}

// JWTManager handles JWT operations
type JWTManager struct {
	secret            []byte
	accessExpiration  time.Duration
	refreshExpiration time.Duration
}

// NewJWTManager creates a new JWT manager
func NewJWTManager(secret string, accessExpiration, refreshExpiration time.Duration) *JWTManager {
	return &JWTManager{
		secret:            []byte(secret),
		accessExpiration:  accessExpiration,
		refreshExpiration: refreshExpiration,
	}
}

// GenerateToken creates a new JWT token
func (j *JWTManager) GenerateToken(userID int64, email string, role string, isAdmin bool) (string, time.Time, error) {
	now := time.Now()
	expiresAt := now.Add(j.accessExpiration)

	claims := JWTClaims{
		UserID:    userID,
		Email:     email,
		Role:      role,
		IsAdmin:   isAdmin,
		ExpiresAt: expiresAt.Unix(),
		IssuedAt:  now.Unix(),
	}

	// Simple JWT implementation (header.payload.signature)
	// In production, consider using github.com/golang-jwt/jwt
	claimsJSON, err := json.Marshal(claims)
	if err != nil {
		return "", time.Time{}, err
	}

	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))
	payload := base64.RawURLEncoding.EncodeToString(claimsJSON)
	signature := j.sign(header + "." + payload)

	token := header + "." + payload + "." + signature

	return token, expiresAt, nil
}

// GenerateRefreshToken creates a random refresh token
func (j *JWTManager) GenerateRefreshToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// ValidateToken validates and parses a JWT token
func (j *JWTManager) ValidateToken(token string) (*JWTClaims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, ErrInvalidToken
	}

	header, payload, signature := parts[0], parts[1], parts[2]

	// Verify signature
	expectedSignature := j.sign(header + "." + payload)
	if signature != expectedSignature {
		return nil, ErrInvalidToken
	}

	// Decode payload
	claimsJSON, err := base64.RawURLEncoding.DecodeString(payload)
	if err != nil {
		return nil, ErrInvalidToken
	}

	var claims JWTClaims
	if err := json.Unmarshal(claimsJSON, &claims); err != nil {
		return nil, ErrInvalidToken
	}

	// Check expiration
	if time.Now().Unix() > claims.ExpiresAt {
		return nil, ErrTokenExpired
	}

	return &claims, nil
}

// sign creates a signature for the token using HMAC-SHA256
func (j *JWTManager) sign(data string) string {
	mac := hmac.New(sha256.New, j.secret)
	mac.Write([]byte(data))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

// HashPassword hashes a password using bcrypt
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPassword compares a password with its hash
func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// ValidatePassword checks if password meets requirements
func ValidatePassword(password string) error {
	if len(password) < 8 {
		return ErrWeakPassword
	}

	hasUpper := false
	hasLower := false
	hasDigit := false

	for _, char := range password {
		switch {
		case char >= 'A' && char <= 'Z':
			hasUpper = true
		case char >= 'a' && char <= 'z':
			hasLower = true
		case char >= '0' && char <= '9':
			hasDigit = true
		}
	}

	if !hasUpper || !hasLower || !hasDigit {
		return ErrWeakPassword
	}

	return nil
}

// ValidateEmail checks if email is valid (basic check)
func ValidateEmail(email string) error {
	if !strings.Contains(email, "@") || !strings.Contains(email, ".") {
		return ErrInvalidEmail
	}
	if len(email) < 3 {
		return ErrInvalidEmail
	}
	return nil
}

// GenerateVerificationToken generates a random verification token
func GenerateVerificationToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// GetAccessExpiration returns the access token expiration duration
func (j *JWTManager) GetAccessExpiration() time.Duration {
	return j.accessExpiration
}

// GetRefreshExpiration returns the refresh token expiration duration
func (j *JWTManager) GetRefreshExpiration() time.Duration {
	return j.refreshExpiration
}
