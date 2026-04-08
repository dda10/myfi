package auth

import (
	"crypto/sha256"
	"database/sql"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// AuthService handles user authentication, JWT token management, and session tracking.
type AuthService struct {
	db       *sql.DB
	config   AuthConfig
	sessions map[string]*Session // In-memory session store (sessionID -> Session)
	mu       sync.RWMutex
}

// NewAuthService creates a new AuthService with the given database and configuration.
func NewAuthService(db *sql.DB, config AuthConfig) *AuthService {
	if config.JWTSecret == "" {
		config.JWTSecret = DefaultAuthConfig().JWTSecret
	}
	if config.TokenExpiry == 0 {
		config.TokenExpiry = DefaultAuthConfig().TokenExpiry
	}
	if config.SessionTimeout == 0 {
		config.SessionTimeout = DefaultAuthConfig().SessionTimeout
	}
	if config.MaxFailedAttempts == 0 {
		config.MaxFailedAttempts = DefaultAuthConfig().MaxFailedAttempts
	}
	if config.LockoutWindow == 0 {
		config.LockoutWindow = DefaultAuthConfig().LockoutWindow
	}
	if config.LockoutDuration == 0 {
		config.LockoutDuration = DefaultAuthConfig().LockoutDuration
	}
	if config.BcryptCost == 0 {
		config.BcryptCost = DefaultAuthConfig().BcryptCost
	}

	return &AuthService{
		db:       db,
		config:   config,
		sessions: make(map[string]*Session),
	}
}

// Register creates a new user account with bcrypt-hashed password.
func (s *AuthService) Register(req RegisterRequest) (*User, error) {
	var exists bool
	err := s.db.QueryRow(`SELECT EXISTS(SELECT 1 FROM users WHERE username = $1)`, req.Username).Scan(&exists)
	if err != nil {
		return nil, fmt.Errorf("failed to check username: %w", err)
	}
	if exists {
		return nil, &AuthError{Code: ErrCodeUserExists, Message: "username already exists"}
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), s.config.BcryptCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	var user User
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
func (s *AuthService) Login(req LoginRequest) (*LoginResponse, error) {
	var user User
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
		return nil, &AuthError{Code: ErrCodeInvalidCredentials, Message: "invalid username or password"}
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

	if user.AccountLockedUntil != nil && time.Now().Before(*user.AccountLockedUntil) {
		remaining := time.Until(*user.AccountLockedUntil).Round(time.Minute)
		return nil, &AuthError{
			Code:    ErrCodeAccountLocked,
			Message: fmt.Sprintf("account is locked for %v", remaining),
		}
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err != nil {
		if err := s.recordFailedLogin(user.ID); err != nil {
			return nil, fmt.Errorf("failed to record failed login: %w", err)
		}
		return nil, &AuthError{Code: ErrCodeInvalidCredentials, Message: "invalid username or password"}
	}

	if err := s.resetFailedLogins(user.ID); err != nil {
		return nil, fmt.Errorf("failed to reset failed logins: %w", err)
	}

	if err := s.updateLastLogin(user.ID); err != nil {
		return nil, fmt.Errorf("failed to update last login: %w", err)
	}

	session := s.createSession(user.ID)
	token, err := s.generateJWT(user, session.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &LoginResponse{
		Token:     token,
		ExpiresAt: session.ExpiresAt.Unix(),
		User:      user,
	}, nil
}

// ChangePassword changes a user's password after verifying the current password.
func (s *AuthService) ChangePassword(userID string, req ChangePasswordRequest) error {
	var currentHash string
	err := s.db.QueryRow(`SELECT password_hash FROM users WHERE id = $1`, userID).Scan(&currentHash)
	if err == sql.ErrNoRows {
		return &AuthError{Code: ErrCodeUserNotFound, Message: "user not found"}
	}
	if err != nil {
		return fmt.Errorf("failed to fetch user: %w", err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(currentHash), []byte(req.CurrentPassword))
	if err != nil {
		return &AuthError{Code: ErrCodePasswordMismatch, Message: "current password is incorrect"}
	}

	newHash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), s.config.BcryptCost)
	if err != nil {
		return fmt.Errorf("failed to hash new password: %w", err)
	}

	_, err = s.db.Exec(`UPDATE users SET password_hash = $1 WHERE id = $2`, string(newHash), userID)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	s.invalidateUserSessions(userID)
	return nil
}

// ValidateToken validates a JWT token and returns the claims if valid.
func (s *AuthService) ValidateToken(tokenString string) (*JWTClaims, error) {
	claims, err := s.parseJWT(tokenString)
	if err != nil {
		return nil, err
	}

	s.mu.RLock()
	session, exists := s.sessions[claims.SessionID]
	s.mu.RUnlock()

	if !exists {
		return nil, &AuthError{Code: ErrCodeSessionExpired, Message: "session not found"}
	}

	if time.Since(session.LastActivity) > s.config.SessionTimeout {
		s.invalidateSession(claims.SessionID)
		return nil, &AuthError{Code: ErrCodeSessionExpired, Message: "session expired due to inactivity"}
	}

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
func (s *AuthService) GetUserByID(userID string) (*User, error) {
	var user User
	var lastLogin sql.NullTime

	err := s.db.QueryRow(
		`SELECT id, username, email, created_at, last_login, theme_preference, language_preference
		 FROM users WHERE id = $1`,
		userID,
	).Scan(&user.ID, &user.Username, &user.Email, &user.CreatedAt, &lastLogin, &user.ThemePreference, &user.LanguagePreference)
	if err == sql.ErrNoRows {
		return nil, &AuthError{Code: ErrCodeUserNotFound, Message: "user not found"}
	}
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user: %w", err)
	}

	if lastLogin.Valid {
		user.LastLogin = &lastLogin.Time
	}

	return &user, nil
}

func (s *AuthService) recordFailedLogin(userID string) error {
	var failedAttempts int
	var accountLockedUntil sql.NullTime

	err := s.db.QueryRow(
		`SELECT failed_login_attempts, account_locked_until FROM users WHERE id = $1`,
		userID,
	).Scan(&failedAttempts, &accountLockedUntil)
	if err != nil {
		return fmt.Errorf("failed to get failed attempts: %w", err)
	}

	if accountLockedUntil.Valid && time.Now().After(accountLockedUntil.Time) {
		failedAttempts = 0
	}

	failedAttempts++

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

func (s *AuthService) resetFailedLogins(userID string) error {
	_, err := s.db.Exec(
		`UPDATE users SET failed_login_attempts = 0, account_locked_until = NULL WHERE id = $1`,
		userID,
	)
	return err
}

func (s *AuthService) updateLastLogin(userID string) error {
	_, err := s.db.Exec(`UPDATE users SET last_login = NOW() WHERE id = $1`, userID)
	return err
}

func (s *AuthService) createSession(userID string) *Session {
	sessionID := generateSessionID()
	now := time.Now()

	session := &Session{
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

func (s *AuthService) invalidateSession(sessionID string) {
	s.mu.Lock()
	delete(s.sessions, sessionID)
	s.mu.Unlock()
}

func (s *AuthService) invalidateUserSessions(userID string) {
	s.mu.Lock()
	for id, session := range s.sessions {
		if session.UserID == userID {
			delete(s.sessions, id)
		}
	}
	s.mu.Unlock()
}

func (s *AuthService) generateJWT(user User, sessionID string) (string, error) {
	now := time.Now()
	claims := JWTClaims{
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

func (s *AuthService) parseJWT(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(s.config.JWTSecret), nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, &AuthError{Code: ErrCodeTokenExpired, Message: "token has expired"}
		}
		return nil, &AuthError{Code: ErrCodeInvalidToken, Message: "invalid token"}
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, &AuthError{Code: ErrCodeInvalidToken, Message: "invalid token claims"}
	}

	return claims, nil
}

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

// UpdateSettings updates user preferences (theme, language).
func (s *AuthService) UpdateSettings(userID string, theme, language *string) error {
	if theme != nil {
		if _, err := s.db.Exec(`UPDATE users SET theme_preference = $1 WHERE id = $2`, *theme, userID); err != nil {
			return fmt.Errorf("failed to update theme: %w", err)
		}
	}
	if language != nil {
		if _, err := s.db.Exec(`UPDATE users SET language_preference = $1 WHERE id = $2`, *language, userID); err != nil {
			return fmt.Errorf("failed to update language: %w", err)
		}
	}
	return nil
}

// CleanupExpiredSessions removes expired sessions from the store.
func (s *AuthService) CleanupExpiredSessions() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	count := 0
	for id, session := range s.sessions {
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
func (s *AuthService) RefreshSession(sessionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	session, exists := s.sessions[sessionID]
	if !exists {
		return &AuthError{Code: ErrCodeSessionExpired, Message: "session not found"}
	}

	session.LastActivity = time.Now()
	return nil
}
