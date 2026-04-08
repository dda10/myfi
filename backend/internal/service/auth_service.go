package service

import (
	"database/sql"

	"myfi-backend/internal/domain/auth"
)

// Type alias bridging domain/auth service into the service package
// for backward compatibility during migration.

type AuthService = auth.AuthService

func NewAuthService(db *sql.DB, config auth.AuthConfig) *AuthService {
	return auth.NewAuthService(db, config)
}
