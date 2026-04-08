package watchlist

import (
	"database/sql"
	"fmt"
	"log"
)

// WatchlistService manages watchlists and their symbols with database persistence.
type WatchlistService struct {
	db *sql.DB
}

// NewWatchlistService creates a new WatchlistService.
func NewWatchlistService(database *sql.DB) *WatchlistService {
	return &WatchlistService{db: database}
}

// CreateWatchlist creates a new named watchlist for a user.
func (s *WatchlistService) CreateWatchlist(userID string, name string) (*Watchlist, error) {
	if name == "" {
		return nil, fmt.Errorf("watchlist name cannot be empty")
	}

	var w Watchlist
	err := s.db.QueryRow(
		`INSERT INTO watchlists (user_id, name) VALUES ($1, $2)
		 RETURNING id, user_id, name, created_at`,
		userID, name,
	).Scan(&w.ID, &w.UserID, &w.Name, &w.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create watchlist: %w", err)
	}
	w.Symbols = []WatchlistSymbol{}
	return &w, nil
}

// RenameWatchlist renames an existing watchlist owned by the given user.
func (s *WatchlistService) RenameWatchlist(userID string, watchlistID int, newName string) error {
	if newName == "" {
		return fmt.Errorf("watchlist name cannot be empty")
	}

	result, err := s.db.Exec(
		`UPDATE watchlists SET name = $1 WHERE id = $2 AND user_id = $3`,
		newName, watchlistID, userID,
	)
	if err != nil {
		return fmt.Errorf("failed to rename watchlist: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("watchlist not found")
	}
	return nil
}

// DeleteWatchlist deletes a watchlist owned by the given user and all its symbols (CASCADE).
func (s *WatchlistService) DeleteWatchlist(userID string, watchlistID int) error {
	result, err := s.db.Exec(
		`DELETE FROM watchlists WHERE id = $1 AND user_id = $2`,
		watchlistID, userID,
	)
	if err != nil {
		return fmt.Errorf("failed to delete watchlist: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("watchlist not found")
	}
	return nil
}

// GetWatchlists returns all watchlists for a user, each with its symbols loaded.
func (s *WatchlistService) GetWatchlists(userID string) ([]Watchlist, error) {
	rows, err := s.db.Query(
		`SELECT id, user_id, name, created_at FROM watchlists WHERE user_id = $1 ORDER BY created_at ASC`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query watchlists: %w", err)
	}
	defer rows.Close()

	var watchlists []Watchlist
	for rows.Next() {
		var w Watchlist
		if err := rows.Scan(&w.ID, &w.UserID, &w.Name, &w.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan watchlist: %w", err)
		}
		w.Symbols = []WatchlistSymbol{}
		watchlists = append(watchlists, w)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating watchlists: %w", err)
	}

	// Load symbols for each watchlist
	for i := range watchlists {
		symbols, err := s.GetWatchlistSymbols(watchlists[i].ID)
		if err != nil {
			return nil, err
		}
		watchlists[i].Symbols = symbols
	}

	return watchlists, nil
}

// GetWatchlistSymbols returns all symbols for a given watchlist, ordered by position.
func (s *WatchlistService) GetWatchlistSymbols(watchlistID int) ([]WatchlistSymbol, error) {
	rows, err := s.db.Query(
		`SELECT id, watchlist_id, symbol, position, price_alert_above, price_alert_below, created_at
		 FROM watchlist_symbols WHERE watchlist_id = $1 ORDER BY position ASC`,
		watchlistID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query watchlist symbols: %w", err)
	}
	defer rows.Close()

	symbols := []WatchlistSymbol{}
	for rows.Next() {
		var ws WatchlistSymbol
		if err := rows.Scan(&ws.ID, &ws.WatchlistID, &ws.Symbol, &ws.Position, &ws.PriceAlertAbove, &ws.PriceAlertBelow, &ws.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan watchlist symbol: %w", err)
		}
		symbols = append(symbols, ws)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating watchlist symbols: %w", err)
	}
	return symbols, nil
}

// AddSymbol adds a symbol to a watchlist at the next position.
func (s *WatchlistService) AddSymbol(watchlistID int, symbol string) error {
	if symbol == "" {
		return fmt.Errorf("symbol cannot be empty")
	}

	// Get the next position
	var maxPos sql.NullInt64
	err := s.db.QueryRow(
		`SELECT MAX(position) FROM watchlist_symbols WHERE watchlist_id = $1`,
		watchlistID,
	).Scan(&maxPos)
	if err != nil {
		return fmt.Errorf("failed to get max position: %w", err)
	}

	nextPos := 0
	if maxPos.Valid {
		nextPos = int(maxPos.Int64) + 1
	}

	_, err = s.db.Exec(
		`INSERT INTO watchlist_symbols (watchlist_id, symbol, position) VALUES ($1, $2, $3)`,
		watchlistID, symbol, nextPos,
	)
	if err != nil {
		return fmt.Errorf("failed to add symbol to watchlist: %w", err)
	}
	return nil
}

// RemoveSymbol removes a symbol from a watchlist and reorders remaining symbols.
func (s *WatchlistService) RemoveSymbol(watchlistID int, symbol string) error {
	result, err := s.db.Exec(
		`DELETE FROM watchlist_symbols WHERE watchlist_id = $1 AND symbol = $2`,
		watchlistID, symbol,
	)
	if err != nil {
		return fmt.Errorf("failed to remove symbol: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("symbol not found in watchlist")
	}

	// Reorder remaining symbols to close gaps
	_, err = s.db.Exec(
		`WITH ranked AS (
			SELECT id, ROW_NUMBER() OVER (ORDER BY position ASC) - 1 AS new_pos
			FROM watchlist_symbols WHERE watchlist_id = $1
		)
		UPDATE watchlist_symbols SET position = ranked.new_pos
		FROM ranked WHERE watchlist_symbols.id = ranked.id`,
		watchlistID,
	)
	if err != nil {
		return fmt.Errorf("failed to reorder symbols after removal: %w", err)
	}
	return nil
}

// ReorderSymbols sets the order of symbols in a watchlist based on the provided slice.
// The symbols slice must contain exactly the symbols currently in the watchlist.
func (s *WatchlistService) ReorderSymbols(watchlistID int, symbols []string) error {
	if len(symbols) == 0 {
		return fmt.Errorf("symbols list cannot be empty")
	}

	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck

	// Verify all symbols exist in the watchlist
	var count int
	err = tx.QueryRow(
		`SELECT COUNT(*) FROM watchlist_symbols WHERE watchlist_id = $1`,
		watchlistID,
	).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to count symbols: %w", err)
	}
	if count != len(symbols) {
		return fmt.Errorf("provided %d symbols but watchlist has %d", len(symbols), count)
	}

	// Update positions
	for i, sym := range symbols {
		result, err := tx.Exec(
			`UPDATE watchlist_symbols SET position = $1 WHERE watchlist_id = $2 AND symbol = $3`,
			i, watchlistID, sym,
		)
		if err != nil {
			return fmt.Errorf("failed to update position for %s: %w", sym, err)
		}
		rows, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("failed to check rows affected for %s: %w", sym, err)
		}
		if rows == 0 {
			return fmt.Errorf("symbol %s not found in watchlist", sym)
		}
	}

	return tx.Commit()
}

// SetPriceAlert sets the above and/or below price alert thresholds for a symbol in a watchlist.
// Pass nil to clear a threshold.
func (s *WatchlistService) SetPriceAlert(watchlistID int, symbol string, above, below *float64) error {
	result, err := s.db.Exec(
		`UPDATE watchlist_symbols SET price_alert_above = $1, price_alert_below = $2
		 WHERE watchlist_id = $3 AND symbol = $4`,
		above, below, watchlistID, symbol,
	)
	if err != nil {
		return fmt.Errorf("failed to set price alert: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("symbol not found in watchlist")
	}
	return nil
}

// EnsureDefaultWatchlist ensures a default "My Watchlist" exists for the user,
// creating it with popular VN stocks if it doesn't exist yet.
// This should be called when a new user accesses the platform for the first time.
func (s *WatchlistService) EnsureDefaultWatchlist(userID string) (*Watchlist, error) {
	return s.GetDefaultWatchlist(userID)
}

// GetDefaultWatchlist returns the user's default watchlist ("My Watchlist"),
// creating it with popular VN stocks if it doesn't exist yet.
func (s *WatchlistService) GetDefaultWatchlist(userID string) (*Watchlist, error) {
	var w Watchlist
	err := s.db.QueryRow(
		`SELECT id, user_id, name, created_at FROM watchlists WHERE user_id = $1 AND name = $2`,
		userID, "My Watchlist",
	).Scan(&w.ID, &w.UserID, &w.Name, &w.CreatedAt)
	if err == sql.ErrNoRows {
		return s.createDefaultWatchlist(userID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query default watchlist: %w", err)
	}

	symbols, err := s.GetWatchlistSymbols(w.ID)
	if err != nil {
		return nil, err
	}
	w.Symbols = symbols
	return &w, nil
}

// GetAllWatchedSymbols returns all unique symbols across all watchlists for a user.
// This is used by the Monitor_Agent to determine which symbols to scan.
func (s *WatchlistService) GetAllWatchedSymbols(userID string) ([]string, error) {
	rows, err := s.db.Query(
		`SELECT DISTINCT ws.symbol
		 FROM watchlist_symbols ws
		 JOIN watchlists w ON ws.watchlist_id = w.id
		 WHERE w.user_id = $1
		 ORDER BY ws.symbol`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query watched symbols: %w", err)
	}
	defer rows.Close()

	var symbols []string
	for rows.Next() {
		var sym string
		if err := rows.Scan(&sym); err != nil {
			return nil, fmt.Errorf("failed to scan symbol: %w", err)
		}
		symbols = append(symbols, sym)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating watched symbols: %w", err)
	}
	return symbols, nil
}

// SyncToMonitorAgent syncs the user's watchlist symbols to the Monitor_Agent scan list.
// This is a stub — Monitor_Agent integration will be implemented in a later task.
func (s *WatchlistService) SyncToMonitorAgent(userID string) error {
	symbols, err := s.GetAllWatchedSymbols(userID)
	if err != nil {
		return fmt.Errorf("failed to get watched symbols for sync: %w", err)
	}
	log.Printf("SyncToMonitorAgent: user %s has %d symbols to scan: %v", userID, len(symbols), symbols)
	// TODO: Push symbols to Monitor_Agent scan list when Monitor_Agent is implemented
	return nil
}

// createDefaultWatchlist creates a default "My Watchlist" with popular VN stocks for a new user.
func (s *WatchlistService) createDefaultWatchlist(userID string) (*Watchlist, error) {
	w, err := s.CreateWatchlist(userID, "My Watchlist")
	if err != nil {
		return nil, fmt.Errorf("failed to create default watchlist: %w", err)
	}

	defaultSymbols := []string{"VNM", "FPT", "SSI", "HPG", "MWG"}
	for _, sym := range defaultSymbols {
		if err := s.AddSymbol(w.ID, sym); err != nil {
			log.Printf("warning: failed to add default symbol %s: %v", sym, err)
		}
	}

	// Reload symbols
	symbols, err := s.GetWatchlistSymbols(w.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to query default watchlist symbols: %w", err)
	}
	w.Symbols = symbols
	return w, nil
}
