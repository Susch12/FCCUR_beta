package storage

import "errors"

// Package errors
var (
	ErrPackageNotFound = errors.New("package not found")
)

// User errors
var (
	ErrUserNotFound    = errors.New("user not found")
	ErrEmailExists     = errors.New("email already exists")
	ErrInvalidToken    = errors.New("invalid or expired token")
	ErrSessionNotFound = errors.New("session not found")
	ErrSessionExpired  = errors.New("session expired")
)
