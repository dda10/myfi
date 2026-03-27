package model

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// User represents a user account in the system.
type User struct {
	ID                  int        `json:"id"`
	Username            string     `json:"username"`
	PasswordHash        string     `json:"-"` // Never expose in JSON
	Email               *string    `json:"email,omitempty"`
	CreatedAt           time.Time  `json:"createdAt"`
	LastLogin           *time.Time `json:"lastLogin,omitempty"`
	FailedLoginAttempts int        `json:"-"`
	AccountLockedUntil  *time.Time `json:"-"`
	ThemePreference     string     `json:"themePreference"`
	LanguagePreference  string     `json:"languagePreference"`
}

// Session represents an active user session.
type Session struct {
	ID           string    `json:"id"`
	UserID       int       `json:"userId"`
	Token        string    `json:"-"` // Never expose in JSON
	CreatedAt    time.Time `json:"createdAt"`
	LastActivity time.Time `json:"lastActivity"`
	ExpiresAt    time.Time `json:"expiresAt"`
}

// LoginRequest represents a login request payload.
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse represents a successful login response.
type LoginResponse struct {
	Token     string `json:"token"`
	ExpiresAt int64  `json:"expiresAt"`
	User      User   `json:"user"`
}

// RegisterRequest represents a user registration request.
type RegisterRequest struct {
	Username string  `json:"username" binding:"required,min=3,max=50"`
	Password string  `json:"password" binding:"required,min=8"`
	Email    *string `json:"email,omitempty"`
}

// ChangePasswordRequest represents a password change request.
type ChangePasswordRequest struct {
	CurrentPassword string `json:"currentPassword" binding:"required"`
	NewPassword     string `json:"newPassword" binding:"required,min=8"`
}

// JWTClaims represents the claims stored in a JWT token.
// Embeds jwt.RegisteredClaims for standard fields (exp, iat, etc.)
type JWTClaims struct {
	jwt.RegisteredClaims
	UserID    int    `json:"user_id"`
	Username  string `json:"username"`
	SessionID string `json:"session_id"`
}

// AuthConfig holds authentication configuration.
type AuthConfig struct {
	JWTSecret         string        // Secret key for signing JWT tokens
	TokenExpiry       time.Duration // JWT token expiry duration (default 24h)
	SessionTimeout    time.Duration // Session inactivity timeout (default 4h)
	MaxFailedAttempts int           // Max failed login attempts before lockout (default 5)
	LockoutWindow     time.Duration // Window for counting failed attempts (default 15min)
	LockoutDuration   time.Duration // Account lockout duration (default 30min)
	BcryptCost        int           // Bcrypt cost factor (default 12)
}

// DefaultAuthConfig returns the default authentication configuration.
func DefaultAuthConfig() AuthConfig {
	return AuthConfig{
		JWTSecret:         "change-me-in-production", // Should be overridden via env
		TokenExpiry:       24 * time.Hour,
		SessionTimeout:    4 * time.Hour,
		MaxFailedAttempts: 5,
		LockoutWindow:     15 * time.Minute,
		LockoutDuration:   30 * time.Minute,
		BcryptCost:        12,
	}
}

// AuthError represents an authentication error with a specific code.
type AuthError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e *AuthError) Error() string {
	return e.Message
}

// Common authentication error codes
const (
	ErrCodeInvalidCredentials = "INVALID_CREDENTIALS"
	ErrCodeAccountLocked      = "ACCOUNT_LOCKED"
	ErrCodeUserNotFound       = "USER_NOT_FOUND"
	ErrCodeUserExists         = "USER_EXISTS"
	ErrCodeInvalidToken       = "INVALID_TOKEN"
	ErrCodeTokenExpired       = "TOKEN_EXPIRED"
	ErrCodeSessionExpired     = "SESSION_EXPIRED"
	ErrCodePasswordMismatch   = "PASSWORD_MISMATCH"
)
