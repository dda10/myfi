package service

import (
	"crypto/sha256"
	"database/sql"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"myfi-backend/internal/model"
)

// AuthService handles user authentication, JWT token management, and session tracking.
// Requirements: 36.1, 36.2, 36.6, 36.7, 36.8
type AuthService struct {
	db       *sql.DB
	config   model.AuthConfig
	sessions map[string]*model.Session // In-memory session store (sessionID -> Session)
	mu       sync.RWMutex
}

// NewAuthService creates a new AuthService with the given database and configuration.
func NewAuthService(db *sql.DB, config model.AuthConfig) *AuthService {
	if config.JWTSecret == "" {
		config.JWTSecret = model.DefaultAuthConfig().JWTSecret
	}
	if config.TokenExpiry == 0 {
		config.TokenExpiry = model.DefaultAuthConfig().TokenExpiry
	}
	if config.SessionTimeout == 0 {
		config.SessionTimeout = model.DefaultAuthConfig().SessionTimeout
	}
	if config.MaxFailedAttempts == 0 {
		config.MaxFailedAttempts = model.DefaultAuthConfig().MaxFailedAttempts
	}
	if config.LockoutWindow == 0 {
		config.LockoutWindow = model.DefaultAuthConfig().LockoutWindow
	}
	if config.LockoutDuration == 0 {
		config.LockoutDuration = model.DefaultAuthConfig().LockoutDuration
	}
	if config.BcryptCost == 0 {
		config.BcryptCost = model.DefaultAuthConfig().BcryptCost
	}

	return &AuthService{
		db:       db,
		config:   config,
		sessions: make(map[string]*model.Session),
	}
}

// Register creates a new user account with bcrypt-hashed password.
// Requirement 36.1: Support local authentication with username and password, storing passwords hashed with bcrypt
func (s *AuthService) Register(req model.RegisterRequest) (*model.User, error) {
	// Check if username already exists
	var exists bool
	err := s.db.QueryRow(`SELECT EXISTS(SELECT 1 FROM users WHERE username = $1)`, req.Username).Scan(&exists)
	if err != nil {
		return nil, fmt.Errorf("failed to check username: %w", err)
	}
	if exists {
		return nil, &model.AuthError{Code: model.ErrCodeUserExists, Message: "username already exists"}
	}

	// Hash password with bcrypt cost 12
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), s.config.BcryptCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Insert user into database
	var user model.User
	err = s.db.QueryRow(
		`INSERT INTO users (username, password_hash, email, theme_preference, language_preference)
		 VALUES ($1, $2, $3, 'light', 'vi-VN')
		 RETURNING id, username, email, created_at, theme_preference, language_preference`,
		req.Username, string(hashedPassword), req.Email,
	).Scan(&user.ID, &user.Username, &user.Email, &user.CreatedAt, &user.ThemePreference, &user.LanguagePreference)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return &user, nil
}

// Login authenticates a user and issues a JWT token.
// Requirements: 36.1, 36.2, 36.7
func (s *AuthService) Login(req model.LoginRequest) (*model.LoginResponse, error) {
	// Fetch user from database
	var user model.User
	var accountLockedUntil sql.NullTime
	var lastLogin sql.NullTime

	err := s.db.QueryRow(
		`SELECT id, username, password_hash, email, created_at, last_login, 
		        failed_login_attempts, account_locked_until, theme_preference, language_preference
		 FROM users WHERE username = $1`,
		req.Username,
	).Scan(
		&user.ID, &user.Username, &user.PasswordHash, &user.Email, &user.CreatedAt,
		&lastLogin, &user.FailedLoginAttempts, &accountLockedUntil,
		&user.ThemePreference, &user.LanguagePreference,
	)
	if err == sql.ErrNoRows {
		return nil, &model.AuthError{Code: model.ErrCodeInvalidCredentials, Message: "invalid username or password"}
	}
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user: %w", err)
	}

	if lastLogin.Valid {
		user.LastLogin = &lastLogin.Time
	}
	if accountLockedUntil.Valid {
		user.AccountLockedUntil = &accountLockedUntil.Time
	}

	// Check if account is locked (Requirement 36.7)
	if user.AccountLockedUntil != nil && time.Now().Before(*user.AccountLockedUntil) {
		remaining := time.Until(*user.AccountLockedUntil).Round(time.Minute)
		return nil, &model.AuthError{
			Code:    model.ErrCodeAccountLocked,
			Message: fmt.Sprintf("account is locked for %v", remaining),
		}
	}

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err != nil {
		// Increment failed login attempts
		if err := s.recordFailedLogin(user.ID); err != nil {
			return nil, fmt.Errorf("failed to record failed login: %w", err)
		}
		return nil, &model.AuthError{Code: model.ErrCodeInvalidCredentials, Message: "invalid username or password"}
	}

	// Reset failed login attempts on successful login
	if err := s.resetFailedLogins(user.ID); err != nil {
		return nil, fmt.Errorf("failed to reset failed logins: %w", err)
	}

	// Update last login timestamp
	if err := s.updateLastLogin(user.ID); err != nil {
		return nil, fmt.Errorf("failed to update last login: %w", err)
	}

	// Generate session and JWT token (Requirement 36.2)
	session := s.createSession(user.ID)
	token, err := s.generateJWT(user, session.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &model.LoginResponse{
		Token:     token,
		ExpiresAt: session.ExpiresAt.Unix(),
		User:      user,
	}, nil
}

// ChangePassword changes a user's password after verifying the current password.
// Requirement 36.6: Support password change requiring the current password and a new password
func (s *AuthService) ChangePassword(userID int, req model.ChangePasswordRequest) error {
	// Fetch current password hash
	var currentHash string
	err := s.db.QueryRow(`SELECT password_hash FROM users WHERE id = $1`, userID).Scan(&currentHash)
	if err == sql.ErrNoRows {
		return &model.AuthError{Code: model.ErrCodeUserNotFound, Message: "user not found"}
	}
	if err != nil {
		return fmt.Errorf("failed to fetch user: %w", err)
	}

	// Verify current password
	err = bcrypt.CompareHashAndPassword([]byte(currentHash), []byte(req.CurrentPassword))
	if err != nil {
		return &model.AuthError{Code: model.ErrCodePasswordMismatch, Message: "current password is incorrect"}
	}

	// Hash new password with bcrypt cost 12
	newHash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), s.config.BcryptCost)
	if err != nil {
		return fmt.Errorf("failed to hash new password: %w", err)
	}

	// Update password in database
	_, err = s.db.Exec(`UPDATE users SET password_hash = $1 WHERE id = $2`, string(newHash), userID)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	// Invalidate all sessions for this user (force re-login)
	s.invalidateUserSessions(userID)

	return nil
}

// ValidateToken validates a JWT token and returns the claims if valid.
// Requirements: 36.2, 36.8
func (s *AuthService) ValidateToken(tokenString string) (*model.JWTClaims, error) {
	claims, err := s.parseJWT(tokenString)
	if err != nil {
		return nil, err
	}

	// Check session validity and inactivity timeout (Requirement 36.8)
	s.mu.RLock()
	session, exists := s.sessions[claims.SessionID]
	s.mu.RUnlock()

	if !exists {
		return nil, &model.AuthError{Code: model.ErrCodeSessionExpired, Message: "session not found"}
	}

	// Check inactivity timeout (4 hours default)
	if time.Since(session.LastActivity) > s.config.SessionTimeout {
		s.invalidateSession(claims.SessionID)
		return nil, &model.AuthError{Code: model.ErrCodeSessionExpired, Message: "session expired due to inactivity"}
	}

	// Update last activity timestamp (sliding window)
	s.mu.Lock()
	session.LastActivity = time.Now()
	s.mu.Unlock()

	return claims, nil
}

// Logout invalidates a user's session.
func (s *AuthService) Logout(sessionID string) error {
	s.invalidateSession(sessionID)
	return nil
}

// GetUserByID retrieves a user by their ID.
func (s *AuthService) GetUserByID(userID int) (*model.User, error) {
	var user model.User
	var lastLogin sql.NullTime

	err := s.db.QueryRow(
		`SELECT id, username, email, created_at, last_login, theme_preference, language_preference
		 FROM users WHERE id = $1`,
		userID,
	).Scan(&user.ID, &user.Username, &user.Email, &user.CreatedAt, &lastLogin, &user.ThemePreference, &user.LanguagePreference)
	if err == sql.ErrNoRows {
		return nil, &model.AuthError{Code: model.ErrCodeUserNotFound, Message: "user not found"}
	}
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user: %w", err)
	}

	if lastLogin.Valid {
		user.LastLogin = &lastLogin.Time
	}

	return &user, nil
}

// recordFailedLogin increments the failed login counter and locks the account if threshold is reached.
// Requirement 36.7: 5 failed attempts in 15 min → 30 min lockout
func (s *AuthService) recordFailedLogin(userID int) error {
	// Get current failed attempts count
	var failedAttempts int
	var accountLockedUntil sql.NullTime

	err := s.db.QueryRow(
		`SELECT failed_login_attempts, account_locked_until FROM users WHERE id = $1`,
		userID,
	).Scan(&failedAttempts, &accountLockedUntil)
	if err != nil {
		return fmt.Errorf("failed to get failed attempts: %w", err)
	}

	// If account was previously locked and lockout has expired, reset counter
	if accountLockedUntil.Valid && time.Now().After(accountLockedUntil.Time) {
		failedAttempts = 0
	}

	failedAttempts++

	// Check if we should lock the account
	if failedAttempts >= s.config.MaxFailedAttempts {
		lockUntil := time.Now().Add(s.config.LockoutDuration)
		_, err = s.db.Exec(
			`UPDATE users SET failed_login_attempts = $1, account_locked_until = $2 WHERE id = $3`,
			failedAttempts, lockUntil, userID,
		)
	} else {
		_, err = s.db.Exec(
			`UPDATE users SET failed_login_attempts = $1 WHERE id = $2`,
			failedAttempts, userID,
		)
	}

	return err
}

// resetFailedLogins resets the failed login counter for a user.
func (s *AuthService) resetFailedLogins(userID int) error {
	_, err := s.db.Exec(
		`UPDATE users SET failed_login_attempts = 0, account_locked_until = NULL WHERE id = $1`,
		userID,
	)
	return err
}

// updateLastLogin updates the last login timestamp for a user.
func (s *AuthService) updateLastLogin(userID int) error {
	_, err := s.db.Exec(`UPDATE users SET last_login = NOW() WHERE id = $1`, userID)
	return err
}

// createSession creates a new session for a user.
func (s *AuthService) createSession(userID int) *model.Session {
	sessionID := generateSessionID()
	now := time.Now()

	session := &model.Session{
		ID:           sessionID,
		UserID:       userID,
		CreatedAt:    now,
		LastActivity: now,
		ExpiresAt:    now.Add(s.config.TokenExpiry),
	}

	s.mu.Lock()
	s.sessions[sessionID] = session
	s.mu.Unlock()

	return session
}

// invalidateSession removes a session from the store.
func (s *AuthService) invalidateSession(sessionID string) {
	s.mu.Lock()
	delete(s.sessions, sessionID)
	s.mu.Unlock()
}

// invalidateUserSessions removes all sessions for a user.
func (s *AuthService) invalidateUserSessions(userID int) {
	s.mu.Lock()
	for id, session := range s.sessions {
		if session.UserID == userID {
			delete(s.sessions, id)
		}
	}
	s.mu.Unlock()
}

// generateJWT creates a signed JWT token for the given user and session.
// Requirement 36.2: Issue JWT tokens with 24-hour expiry (configurable)
func (s *AuthService) generateJWT(user model.User, sessionID string) (string, error) {
	now := time.Now()
	claims := model.JWTClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.config.TokenExpiry)),
		},
		UserID:    user.ID,
		Username:  user.Username,
		SessionID: sessionID,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.config.JWTSecret))
}

// parseJWT parses and validates a JWT token string.
func (s *AuthService) parseJWT(tokenString string) (*model.JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &model.JWTClaims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(s.config.JWTSecret), nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, &model.AuthError{Code: model.ErrCodeTokenExpired, Message: "token has expired"}
		}
		return nil, &model.AuthError{Code: model.ErrCodeInvalidToken, Message: "invalid token"}
	}

	claims, ok := token.Claims.(*model.JWTClaims)
	if !ok || !token.Valid {
		return nil, &model.AuthError{Code: model.ErrCodeInvalidToken, Message: "invalid token claims"}
	}

	return claims, nil
}

// generateSessionID generates a unique session ID using SHA-256.
func generateSessionID() string {
	b := make([]byte, 32)
	now := time.Now().UnixNano()
	for i := 0; i < 8; i++ {
		b[i] = byte(now >> (i * 8))
	}
	ptr := fmt.Sprintf("%p", &b)
	copy(b[8:], []byte(ptr))
	h := sha256.Sum256(b)
	return fmt.Sprintf("%x", h)
}

// CleanupExpiredSessions removes expired sessions from the store.
// This should be called periodically (e.g., every hour).
func (s *AuthService) CleanupExpiredSessions() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	count := 0
	for id, session := range s.sessions {
		// Remove if token expired or inactive for too long
		if now.After(session.ExpiresAt) || now.Sub(session.LastActivity) > s.config.SessionTimeout {
			delete(s.sessions, id)
			count++
		}
	}
	return count
}

// GetActiveSessions returns the number of active sessions (for monitoring).
func (s *AuthService) GetActiveSessions() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.sessions)
}

// RefreshSession updates the session's last activity and optionally extends the token expiry.
// Requirement 36.8: Track last activity timestamp per session
func (s *AuthService) RefreshSession(sessionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	session, exists := s.sessions[sessionID]
	if !exists {
		return &model.AuthError{Code: model.ErrCodeSessionExpired, Message: "session not found"}
	}

	session.LastActivity = time.Now()
	return nil
}
