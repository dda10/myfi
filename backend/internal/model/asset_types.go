package model

import (
	"fmt"
	"strings"
)

// AssetType represents the type of asset
type AssetType string

const (
	VNStock AssetType = "vn_stock"
	Crypto  AssetType = "crypto"
	Gold    AssetType = "gold"
	Savings AssetType = "savings"
	Bond    AssetType = "bond"
	Cash    AssetType = "cash"
)

// ValidAssetTypes contains all supported asset types for validation
var ValidAssetTypes = map[AssetType]bool{
	VNStock: true,
	Crypto:  true,
	Gold:    true,
	Savings: true,
	Bond:    true,
	Cash:    true,
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
