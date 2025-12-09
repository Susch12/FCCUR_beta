package storage

import (
	"database/sql"
	"time"

	"github.com/jesus/FCCUR/internal/models"
)

// CreateUser creates a new user
func (p *PostgresDB) CreateUser(email, passwordHash, fullName string, role models.UserRole) (*models.User, error) {
	ctx, cancel := p.getContext()
	defer cancel()

	var id int64
	err := p.pool.QueryRow(ctx, `
		INSERT INTO users (email, password_hash, full_name, role, email_verified)
		VALUES ($1, $2, $3, $4, false)
		RETURNING id
	`, email, passwordHash, fullName, role).Scan(&id)

	if err != nil {
		return nil, err
	}

	return p.GetUserByID(id)
}

// GetUserByID retrieves a user by ID
func (p *PostgresDB) GetUserByID(id int64) (*models.User, error) {
	ctx, cancel := p.getContext()
	defer cancel()

	user := &models.User{}
	err := p.pool.QueryRow(ctx, `
		SELECT id, email, password_hash, full_name, role, assigned_courses,
		       is_active, is_admin, email_verified, verification_token,
		       reset_token, reset_token_expiry, last_login, created_at, updated_at
		FROM users WHERE id = $1
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
	return user, err
}

// GetUserByEmail retrieves a user by email
func (p *PostgresDB) GetUserByEmail(email string) (*models.User, error) {
	ctx, cancel := p.getContext()
	defer cancel()

	user := &models.User{}
	err := p.pool.QueryRow(ctx, `
		SELECT id, email, password_hash, full_name, role, assigned_courses,
		       is_active, is_admin, email_verified, verification_token,
		       reset_token, reset_token_expiry, last_login, created_at, updated_at
		FROM users WHERE email = $1
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
	return user, err
}

// UpdateUserLastLogin updates the last login timestamp
func (p *PostgresDB) UpdateUserLastLogin(userID int64) error {
	ctx, cancel := p.getContext()
	defer cancel()

	_, err := p.pool.Exec(ctx, `
		UPDATE users SET last_login = CURRENT_TIMESTAMP
		WHERE id = $1
	`, userID)
	return err
}

// SetUserResetToken sets a password reset token
func (p *PostgresDB) SetUserResetToken(userID int64, token string, expiry time.Time) error {
	ctx, cancel := p.getContext()
	defer cancel()

	_, err := p.pool.Exec(ctx, `
		UPDATE users
		SET reset_token = $1, reset_token_expiry = $2
		WHERE id = $3
	`, token, expiry, userID)
	return err
}

// GetUserByResetToken retrieves a user by reset token
func (p *PostgresDB) GetUserByResetToken(token string) (*models.User, error) {
	ctx, cancel := p.getContext()
	defer cancel()

	user := &models.User{}
	err := p.pool.QueryRow(ctx, `
		SELECT id, email, password_hash, full_name, role, assigned_courses,
		       is_active, is_admin, email_verified, verification_token,
		       reset_token, reset_token_expiry, last_login, created_at, updated_at
		FROM users
		WHERE reset_token = $1 AND reset_token_expiry > CURRENT_TIMESTAMP
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
	return user, err
}

// UpdateUserPassword updates a user's password and clears reset token
func (p *PostgresDB) UpdateUserPassword(userID int64, passwordHash string) error {
	ctx, cancel := p.getContext()
	defer cancel()

	_, err := p.pool.Exec(ctx, `
		UPDATE users
		SET password_hash = $1, reset_token = NULL, reset_token_expiry = NULL
		WHERE id = $2
	`, passwordHash, userID)
	return err
}

// SetUserEmailVerified marks user's email as verified
func (p *PostgresDB) SetUserEmailVerified(userID int64) error {
	ctx, cancel := p.getContext()
	defer cancel()

	_, err := p.pool.Exec(ctx, `
		UPDATE users
		SET email_verified = true, verification_token = NULL
		WHERE id = $1
	`, userID)
	return err
}

// CreateSession creates a new session
func (p *PostgresDB) CreateSession(userID int64, token, refreshToken, ipAddress, userAgent string, expiresAt time.Time) (*models.Session, error) {
	ctx, cancel := p.getContext()
	defer cancel()

	var id int64
	err := p.pool.QueryRow(ctx, `
		INSERT INTO sessions (user_id, token, refresh_token, ip_address, user_agent, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`, userID, token, refreshToken, ipAddress, userAgent, expiresAt).Scan(&id)

	if err != nil {
		return nil, err
	}

	return p.GetSessionByID(id)
}

// GetSessionByID retrieves a session by ID
func (p *PostgresDB) GetSessionByID(id int64) (*models.Session, error) {
	ctx, cancel := p.getContext()
	defer cancel()

	session := &models.Session{}
	err := p.pool.QueryRow(ctx, `
		SELECT id, user_id, token, refresh_token, ip_address, user_agent, expires_at, created_at
		FROM sessions WHERE id = $1
	`, id).Scan(
		&session.ID, &session.UserID, &session.Token, &session.RefreshToken,
		&session.IPAddress, &session.UserAgent, &session.ExpiresAt, &session.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, ErrSessionNotFound
	}
	return session, err
}

// GetSessionByToken retrieves a session by token
func (p *PostgresDB) GetSessionByToken(token string) (*models.Session, error) {
	ctx, cancel := p.getContext()
	defer cancel()

	session := &models.Session{}
	err := p.pool.QueryRow(ctx, `
		SELECT id, user_id, token, refresh_token, ip_address, user_agent, expires_at, created_at
		FROM sessions WHERE token = $1
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
func (p *PostgresDB) GetSessionByRefreshToken(refreshToken string) (*models.Session, error) {
	ctx, cancel := p.getContext()
	defer cancel()

	session := &models.Session{}
	err := p.pool.QueryRow(ctx, `
		SELECT id, user_id, token, refresh_token, ip_address, user_agent, expires_at, created_at
		FROM sessions WHERE refresh_token = $1
	`, refreshToken).Scan(
		&session.ID, &session.UserID, &session.Token, &session.RefreshToken,
		&session.IPAddress, &session.UserAgent, &session.ExpiresAt, &session.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, ErrSessionNotFound
	}
	return session, err
}

// DeleteSession deletes a session by token
func (p *PostgresDB) DeleteSession(token string) error {
	ctx, cancel := p.getContext()
	defer cancel()

	_, err := p.pool.Exec(ctx, `DELETE FROM sessions WHERE token = $1`, token)
	return err
}

// DeleteUserSessions deletes all sessions for a user
func (p *PostgresDB) DeleteUserSessions(userID int64) error {
	ctx, cancel := p.getContext()
	defer cancel()

	_, err := p.pool.Exec(ctx, `DELETE FROM sessions WHERE user_id = $1`, userID)
	return err
}

// CleanExpiredSessions removes expired sessions
func (p *PostgresDB) CleanExpiredSessions() error {
	ctx, cancel := p.getContext()
	defer cancel()

	_, err := p.pool.Exec(ctx, `DELETE FROM sessions WHERE expires_at < CURRENT_TIMESTAMP`)
	return err
}

// ListUserSessions lists all active sessions for a user
func (p *PostgresDB) ListUserSessions(userID int64) ([]*models.Session, error) {
	ctx, cancel := p.getContext()
	defer cancel()

	rows, err := p.pool.Query(ctx, `
		SELECT id, user_id, token, refresh_token, ip_address, user_agent, expires_at, created_at
		FROM sessions
		WHERE user_id = $1 AND expires_at > CURRENT_TIMESTAMP
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
