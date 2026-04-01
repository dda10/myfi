package service

import "time"

// ObservationFilters defines query filters for pattern observations.
type ObservationFilters struct {
	Symbol        string
	PatternType   string
	StartDate     time.Time
	EndDate       time.Time
	MinConfidence int
	Limit         int
	Offset        int
}
