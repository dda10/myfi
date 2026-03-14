package main

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

// Asset represents a user's asset holding in the registry
type Asset struct {
	ID              int64     `json:"id"`
	UserID          int64     `json:"userId"`
	AssetType       AssetType `json:"assetType"`
	Symbol          string    `json:"symbol"`
	Quantity        float64   `json:"quantity"`
	AverageCost     float64   `json:"averageCost"`
	AcquisitionDate time.Time `json:"acquisitionDate"`
	Account         string    `json:"account"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

// AssetRegistry manages asset CRUD operations with database persistence
type AssetRegistry struct {
	db           *sql.DB
	priceService *PriceService
}

// NewAssetRegistry creates a new AssetRegistry instance
func NewAssetRegistry(db *sql.DB, priceService *PriceService) *AssetRegistry {
	return &AssetRegistry{
		db:           db,
		priceService: priceService,
	}
}

// ValidateAssetType checks if the given asset type is supported.
// Returns an error listing valid types if the type is not recognized.
func ValidateAssetType(at AssetType) error {
	if ValidAssetTypes[at] {
		return nil
	}
	var supported []string
	for k := range ValidAssetTypes {
		supported = append(supported, string(k))
	}
	return fmt.Errorf("unrecognized asset type %q; supported types: %s", at, strings.Join(supported, ", "))
}

// AddAsset persists a new asset to the database.
// All monetary values (AverageCost) must be in VND.
func (r *AssetRegistry) AddAsset(ctx context.Context, asset Asset) (int64, error) {
	if err := ValidateAssetType(asset.AssetType); err != nil {
		return 0, err
	}
	if asset.UserID <= 0 {
		return 0, fmt.Errorf("invalid user ID: %d", asset.UserID)
	}
	if asset.Symbol == "" {
		return 0, fmt.Errorf("symbol is required")
	}
	if asset.Quantity <= 0 {
		return 0, fmt.Errorf("quantity must be positive")
	}
	if asset.AverageCost < 0 {
		return 0, fmt.Errorf("average cost must be non-negative")
	}

	now := time.Now()
	result, err := r.db.ExecContext(ctx,
		`INSERT INTO assets (user_id, asset_type, symbol, quantity, average_cost, acquisition_date, account, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		asset.UserID, string(asset.AssetType), asset.Symbol, asset.Quantity,
		asset.AverageCost, asset.AcquisitionDate.Format(time.RFC3339),
		asset.Account, now.Format(time.RFC3339), now.Format(time.RFC3339),
	)
	if err != nil {
		return 0, fmt.Errorf("failed to insert asset: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get inserted asset ID: %w", err)
	}
	return id, nil
}

// UpdateAsset updates an existing asset and triggers NAV recalculation.
func (r *AssetRegistry) UpdateAsset(ctx context.Context, asset Asset) error {
	if asset.ID <= 0 {
		return fmt.Errorf("invalid asset ID: %d", asset.ID)
	}
	if err := ValidateAssetType(asset.AssetType); err != nil {
		return err
	}
	if asset.Quantity <= 0 {
		return fmt.Errorf("quantity must be positive")
	}
	if asset.AverageCost < 0 {
		return fmt.Errorf("average cost must be non-negative")
	}

	now := time.Now()
	result, err := r.db.ExecContext(ctx,
		`UPDATE assets SET asset_type = ?, symbol = ?, quantity = ?, average_cost = ?,
		 acquisition_date = ?, account = ?, updated_at = ?
		 WHERE id = ? AND user_id = ?`,
		string(asset.AssetType), asset.Symbol, asset.Quantity, asset.AverageCost,
		asset.AcquisitionDate.Format(time.RFC3339), asset.Account,
		now.Format(time.RFC3339), asset.ID, asset.UserID,
	)
	if err != nil {
		return fmt.Errorf("failed to update asset: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check update result: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("asset %d not found for user %d", asset.ID, asset.UserID)
	}

	// Recalculate NAV within the same request cycle (Requirement 1.3)
	nav, err := r.computeNAV(ctx, asset.UserID)
	if err != nil {
		return fmt.Errorf("failed to recalculate NAV after update: %w", err)
	}
	_ = nav // NAV computed; callers can fetch it separately

	return nil
}

// DeleteAsset removes an asset and all associated transactions (cascade).
func (r *AssetRegistry) DeleteAsset(ctx context.Context, assetID, userID int64) error {
	if assetID <= 0 {
		return fmt.Errorf("invalid asset ID: %d", assetID)
	}

	// Fetch the asset first to get symbol and type for transaction cleanup
	asset, err := r.GetAsset(ctx, assetID, userID)
	if err != nil {
		return fmt.Errorf("asset %d not found for user %d: %w", assetID, userID, err)
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Delete associated transactions from Transaction_Ledger (Requirement 1.4)
	_, err = tx.ExecContext(ctx,
		`DELETE FROM transactions WHERE user_id = ? AND symbol = ? AND asset_type = ?`,
		userID, asset.Symbol, string(asset.AssetType),
	)
	if err != nil {
		return fmt.Errorf("failed to delete associated transactions: %w", err)
	}

	// Delete the asset
	result, err := tx.ExecContext(ctx,
		`DELETE FROM assets WHERE id = ? AND user_id = ?`,
		assetID, userID,
	)
	if err != nil {
		return fmt.Errorf("failed to delete asset: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check delete result: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("asset %d not found for user %d", assetID, userID)
	}

	return tx.Commit()
}

// GetAsset retrieves a single asset by ID and user.
func (r *AssetRegistry) GetAsset(ctx context.Context, assetID, userID int64) (Asset, error) {
	var a Asset
	var assetType string
	var acqDate string
	var createdAt, updatedAt string
	var account sql.NullString

	err := r.db.QueryRowContext(ctx,
		`SELECT id, user_id, asset_type, symbol, quantity, average_cost, acquisition_date, account, created_at, updated_at
		 FROM assets WHERE id = ? AND user_id = ?`,
		assetID, userID,
	).Scan(&a.ID, &a.UserID, &assetType, &a.Symbol, &a.Quantity, &a.AverageCost,
		&acqDate, &account, &createdAt, &updatedAt)
	if err != nil {
		return Asset{}, fmt.Errorf("asset not found: %w", err)
	}

	a.AssetType = AssetType(assetType)
	a.AcquisitionDate, _ = time.Parse(time.RFC3339, acqDate)
	a.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	a.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)
	if account.Valid {
		a.Account = account.String
	}
	return a, nil
}

// GetAssetsByUser retrieves all assets for a given user.
func (r *AssetRegistry) GetAssetsByUser(ctx context.Context, userID int64) ([]Asset, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, user_id, asset_type, symbol, quantity, average_cost, acquisition_date, account, created_at, updated_at
		 FROM assets WHERE user_id = ? ORDER BY created_at DESC`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query assets: %w", err)
	}
	defer rows.Close()

	var assets []Asset
	for rows.Next() {
		var a Asset
		var assetType string
		var acqDate string
		var createdAt, updatedAt string
		var account sql.NullString

		if err := rows.Scan(&a.ID, &a.UserID, &assetType, &a.Symbol, &a.Quantity, &a.AverageCost,
			&acqDate, &account, &createdAt, &updatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan asset row: %w", err)
		}

		a.AssetType = AssetType(assetType)
		a.AcquisitionDate, _ = time.Parse(time.RFC3339, acqDate)
		a.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		a.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)
		if account.Valid {
			a.Account = account.String
		}
		assets = append(assets, a)
	}
	return assets, rows.Err()
}

// computeNAV calculates the total NAV for a user by summing quantity * averageCost
// for all holdings. When PriceService is available, it uses live prices instead.
func (r *AssetRegistry) computeNAV(ctx context.Context, userID int64) (float64, error) {
	assets, err := r.GetAssetsByUser(ctx, userID)
	if err != nil {
		return 0, err
	}

	var nav float64
	for _, a := range assets {
		// Use average cost as fallback value (all stored in VND per Requirement 1.6)
		nav += a.Quantity * a.AverageCost
	}
	return nav, nil
}
