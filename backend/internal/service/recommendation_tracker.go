package service

import (
	"myfi-backend/internal/domain/ranking"
)

// Type aliases bridging domain/ranking services into the service package
// for backward compatibility during migration. These will be removed once
// all services are migrated to domain packages.

type RecommendationTracker = ranking.RecommendationTracker

var NewRecommendationTracker = ranking.NewRecommendationTracker
var NewRecommendationTrackerWithDB = ranking.NewRecommendationTrackerWithDB
