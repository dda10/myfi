package service

import (
	"myfi-backend/internal/domain/knowledge"
)

// Type aliases bridging domain/knowledge services into the service package
// for backward compatibility during migration. These will be removed once
// all services are migrated to domain packages.

type KnowledgeBase = knowledge.KnowledgeBase
type ObservationFilters = knowledge.ObservationFilters
type AccuracyMetrics = knowledge.AccuracyMetrics

var NewKnowledgeBase = knowledge.NewKnowledgeBase
