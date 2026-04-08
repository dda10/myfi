package model

import "myfi-backend/internal/domain/auth"

// --- Type aliases bridging domain/auth types into model for backward compatibility ---

type User = auth.User
type Session = auth.Session
type LoginRequest = auth.LoginRequest
type LoginResponse = auth.LoginResponse
type RegisterRequest = auth.RegisterRequest
type ChangePasswordRequest = auth.ChangePasswordRequest
type JWTClaims = auth.JWTClaims
type AuthConfig = auth.AuthConfig
type AuthError = auth.AuthError

var DefaultAuthConfig = auth.DefaultAuthConfig

const (
	ErrCodeInvalidCredentials = auth.ErrCodeInvalidCredentials
	ErrCodeAccountLocked      = auth.ErrCodeAccountLocked
	ErrCodeUserNotFound       = auth.ErrCodeUserNotFound
	ErrCodeUserExists         = auth.ErrCodeUserExists
	ErrCodeInvalidToken       = auth.ErrCodeInvalidToken
	ErrCodeTokenExpired       = auth.ErrCodeTokenExpired
	ErrCodeSessionExpired     = auth.ErrCodeSessionExpired
	ErrCodePasswordMismatch   = auth.ErrCodePasswordMismatch
)
