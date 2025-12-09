package storage

import (
	"database/sql"
	"time"

	"github.com/jesus/FCCUR/internal/models"
)


// CreateUser creates a new user
func (s *SQLiteDB) CreateUser(email, passwordHash, fullName string, role models.UserRole) (*models.User, error) {
	result, err := s.db.Exec(`
		INSERT INTO users (email, password_hash, full_name, role, email_verified)
		VALUES (?, ?, ?, ?, 0)
	`, email, passwordHash, fullName, role)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return s.GetUserByID(id)
}

// GetUserByID retrieves a user by ID
func (s *SQLiteDB) GetUserByID(id int64) (*models.User, error) {
	user := &models.User{}
	err := s.db.QueryRow(`
		SELECT id, email, password_hash, full_name, role, assigned_courses,
		       is_active, is_admin, email_verified, verification_token,
		       reset_token, reset_token_expiry, last_login, created_at, updated_at
		FROM users WHERE id = ?
	`, id).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.FullName,
		&user.Role, &user.AssignedCourses,
		&user.IsActive, &user.IsAdmin, &user.EmailVerified, &user.VerificationToken,
		&user.ResetToken, &user.ResetTokenExpiry, &user.LastLogin,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return user, nil
}

// GetUserByEmail retrieves a user by email
func (s *SQLiteDB) GetUserByEmail(email string) (*models.User, error) {
	user := &models.User{}
	err := s.db.QueryRow(`
		SELECT id, email, password_hash, full_name, role, assigned_courses,
		       is_active, is_admin, email_verified, verification_token,
		       reset_token, reset_token_expiry, last_login, created_at, updated_at
		FROM users WHERE email = ?
	`, email).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.FullName,
		&user.Role, &user.AssignedCourses,
		&user.IsActive, &user.IsAdmin, &user.EmailVerified, &user.VerificationToken,
		&user.ResetToken, &user.ResetTokenExpiry, &user.LastLogin,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return user, nil
}

// UpdateUserLastLogin updates the last login timestamp
func (s *SQLiteDB) UpdateUserLastLogin(userID int64) error {
	_, err := s.db.Exec(`
		UPDATE users SET last_login = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, userID)
	return err
}

// SetUserResetToken sets a password reset token
func (s *SQLiteDB) SetUserResetToken(userID int64, token string, expiry time.Time) error {
	_, err := s.db.Exec(`
		UPDATE users
		SET reset_token = ?, reset_token_expiry = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, token, expiry, userID)
	return err
}

// GetUserByResetToken retrieves a user by reset token
func (s *SQLiteDB) GetUserByResetToken(token string) (*models.User, error) {
	user := &models.User{}
	err := s.db.QueryRow(`
		SELECT id, email, password_hash, full_name, role, assigned_courses,
		       is_active, is_admin, email_verified, verification_token,
		       reset_token, reset_token_expiry, last_login, created_at, updated_at
		FROM users
		WHERE reset_token = ? AND reset_token_expiry > CURRENT_TIMESTAMP
	`, token).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.FullName,
		&user.Role, &user.AssignedCourses,
		&user.IsActive, &user.IsAdmin, &user.EmailVerified, &user.VerificationToken,
		&user.ResetToken, &user.ResetTokenExpiry, &user.LastLogin,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrInvalidToken
	}
	if err != nil {
		return nil, err
	}
	return user, nil
}

// UpdateUserPassword updates a user's password and clears reset token
func (s *SQLiteDB) UpdateUserPassword(userID int64, passwordHash string) error {
	_, err := s.db.Exec(`
		UPDATE users
		SET password_hash = ?, reset_token = NULL, reset_token_expiry = NULL,
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, passwordHash, userID)
	return err
}

// SetUserEmailVerified marks user's email as verified
func (s *SQLiteDB) SetUserEmailVerified(userID int64) error {
	_, err := s.db.Exec(`
		UPDATE users
		SET email_verified = 1, verification_token = NULL, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, userID)
	return err
}

// CreateSession creates a new session
func (s *SQLiteDB) CreateSession(userID int64, token, refreshToken, ipAddress, userAgent string, expiresAt time.Time) (*models.Session, error) {
	result, err := s.db.Exec(`
		INSERT INTO sessions (user_id, token, refresh_token, ip_address, user_agent, expires_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`, userID, token, refreshToken, ipAddress, userAgent, expiresAt)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return s.GetSessionByID(id)
}

// GetSessionByID retrieves a session by ID
func (s *SQLiteDB) GetSessionByID(id int64) (*models.Session, error) {
	session := &models.Session{}
	err := s.db.QueryRow(`
		SELECT id, user_id, token, refresh_token, ip_address, user_agent, expires_at, created_at
		FROM sessions WHERE id = ?
	`, id).Scan(
		&session.ID, &session.UserID, &session.Token, &session.RefreshToken,
		&session.IPAddress, &session.UserAgent, &session.ExpiresAt, &session.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrSessionNotFound
	}
	if err != nil {
		return nil, err
	}
	return session, nil
}

// GetSessionByToken retrieves a session by token
func (s *SQLiteDB) GetSessionByToken(token string) (*models.Session, error) {
	session := &models.Session{}
	err := s.db.QueryRow(`
		SELECT id, user_id, token, refresh_token, ip_address, user_agent, expires_at, created_at
		FROM sessions WHERE token = ?
	`, token).Scan(
		&session.ID, &session.UserID, &session.Token, &session.RefreshToken,
		&session.IPAddress, &session.UserAgent, &session.ExpiresAt, &session.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrSessionNotFound
	}
	if err != nil {
		return nil, err
	}

	// Check if session is expired
	if time.Now().After(session.ExpiresAt) {
		return nil, ErrSessionExpired
	}

	return session, nil
}

// GetSessionByRefreshToken retrieves a session by refresh token
func (s *SQLiteDB) GetSessionByRefreshToken(refreshToken string) (*models.Session, error) {
	session := &models.Session{}
	err := s.db.QueryRow(`
		SELECT id, user_id, token, refresh_token, ip_address, user_agent, expires_at, created_at
		FROM sessions WHERE refresh_token = ?
	`, refreshToken).Scan(
		&session.ID, &session.UserID, &session.Token, &session.RefreshToken,
		&session.IPAddress, &session.UserAgent, &session.ExpiresAt, &session.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrSessionNotFound
	}
	if err != nil {
		return nil, err
	}
	return session, nil
}

// DeleteSession deletes a session by token
func (s *SQLiteDB) DeleteSession(token string) error {
	_, err := s.db.Exec(`DELETE FROM sessions WHERE token = ?`, token)
	return err
}

// DeleteUserSessions deletes all sessions for a user
func (s *SQLiteDB) DeleteUserSessions(userID int64) error {
	_, err := s.db.Exec(`DELETE FROM sessions WHERE user_id = ?`, userID)
	return err
}

// CleanExpiredSessions removes expired sessions
func (s *SQLiteDB) CleanExpiredSessions() error {
	_, err := s.db.Exec(`DELETE FROM sessions WHERE expires_at < CURRENT_TIMESTAMP`)
	return err
}

// ListUserSessions lists all active sessions for a user
func (s *SQLiteDB) ListUserSessions(userID int64) ([]*models.Session, error) {
	rows, err := s.db.Query(`
		SELECT id, user_id, token, refresh_token, ip_address, user_agent, expires_at, created_at
		FROM sessions
		WHERE user_id = ? AND expires_at > CURRENT_TIMESTAMP
		ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []*models.Session
	for rows.Next() {
		session := &models.Session{}
		err := rows.Scan(
			&session.ID, &session.UserID, &session.Token, &session.RefreshToken,
			&session.IPAddress, &session.UserAgent, &session.ExpiresAt, &session.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, session)
	}

	return sessions, rows.Err()
}
