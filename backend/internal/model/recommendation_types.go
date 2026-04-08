package model

import (
	"myfi-backend/internal/domain/ranking"
)

// --- Type aliases bridging domain/ranking types into model for backward compatibility ---
// These will be removed once all services are migrated to domain packages.

type RecommendationAction = ranking.RecommendationAction

const (
	ActionBuy  = ranking.ActionBuy
	ActionSell = ranking.ActionSell
	ActionHold = ranking.ActionHold
)

type RecommendationRecord = ranking.RecommendationRecord
type AssetRecommendation = ranking.AssetRecommendation
type RecommendationAccuracy = ranking.RecommendationAccuracy
type RecommendationFilter = ranking.RecommendationFilter
type RecommendationSummary = ranking.RecommendationSummary
