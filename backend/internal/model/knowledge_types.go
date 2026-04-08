package model

import (
	"myfi-backend/internal/domain/knowledge"
)

// --- Type aliases bridging domain/knowledge types into model for backward compatibility ---
// These will be removed once all services are migrated to domain packages.

type PatternType = knowledge.PatternType

const (
	PatternAccumulation = knowledge.PatternAccumulation
	PatternDistribution = knowledge.PatternDistribution
	PatternBreakout     = knowledge.PatternBreakout
)

type PatternObservation = knowledge.PatternObservation
