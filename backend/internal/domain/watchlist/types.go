package watchlist

import "time"

// Watchlist represents a named watchlist belonging to a user.
type Watchlist struct {
	ID        int               `json:"id"`
	UserID    string            `json:"userId"`
	Name      string            `json:"name"`
	Symbols   []WatchlistSymbol `json:"symbols"`
	CreatedAt time.Time         `json:"createdAt"`
}

// WatchlistSymbol represents a symbol entry within a watchlist.
type WatchlistSymbol struct {
	ID              int       `json:"id"`
	WatchlistID     int       `json:"watchlistId"`
	Symbol          string    `json:"symbol"`
	Position        int       `json:"position"`
	PriceAlertAbove *float64  `json:"priceAlertAbove,omitempty"`
	PriceAlertBelow *float64  `json:"priceAlertBelow,omitempty"`
	CreatedAt       time.Time `json:"createdAt"`
}
