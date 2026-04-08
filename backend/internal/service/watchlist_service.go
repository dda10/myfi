package service

import (
	"database/sql"

	"myfi-backend/internal/domain/watchlist"
)

// Type alias bridging domain/watchlist service into the service package
// for backward compatibility during migration.

type WatchlistService = watchlist.WatchlistService

func NewWatchlistService(database *sql.DB) *WatchlistService {
	return watchlist.NewWatchlistService(database)
}
