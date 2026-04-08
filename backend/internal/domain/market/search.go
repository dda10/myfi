package market

// ---------------------------------------------------------------------------
// SearchService — in-memory fuzzy search for VN stock symbols
// ---------------------------------------------------------------------------
//
// Requirements satisfied:
//   - 37.3 In-memory index of all VN symbols + company names
//   - 37.4 Fuzzy search with sub-200ms response
//   - 37.6 Search by symbol or company name
//   - 37.7 Return symbol, company name, exchange, sector, relevance score

import (
	"context"
	"log"
	"strings"
	"sync"
	"time"

	"myfi-backend/internal/infra"
)

// SearchService provides fast in-memory fuzzy search across all VN stock symbols.
type SearchService struct {
	marketData *MarketDataService
	cache      *infra.Cache

	mu    sync.RWMutex
	index []SearchEntry
}

// NewSearchService creates a new SearchService.
func NewSearchService(md *MarketDataService, cache *infra.Cache) *SearchService {
	return &SearchService{
		marketData: md,
		cache:      cache,
	}
}

// BuildIndex loads all VN stock symbols and company names into the in-memory index.
// Should be called at startup and periodically refreshed (e.g., daily).
func (s *SearchService) BuildIndex(ctx context.Context) error {
	start := time.Now()

	symbols, err := s.marketData.GetAllSymbols(ctx)
	if err != nil {
		return err
	}

	entries := make([]SearchEntry, 0, len(symbols))
	for _, sym := range symbols {
		entries = append(entries, SearchEntry{
			Symbol:      sym.Symbol,
			CompanyName: sym.CompanyName,
			Exchange:    sym.Exchange,
			Sector:      sym.Sector,
		})
	}

	s.mu.Lock()
	s.index = entries
	s.mu.Unlock()

	log.Printf("[SearchService] Built index with %d entries in %v", len(entries), time.Since(start))
	return nil
}

// Search performs fuzzy matching against the in-memory index.
// Must return within 200ms. Returns up to `limit` results sorted by relevance.
func (s *SearchService) Search(ctx context.Context, query string, limit int) ([]SearchResult, error) {
	if query == "" {
		return nil, nil
	}
	if limit <= 0 {
		limit = 10
	}

	s.mu.RLock()
	index := s.index
	s.mu.RUnlock()

	if len(index) == 0 {
		return nil, nil
	}

	queryLower := strings.ToLower(strings.TrimSpace(query))
	queryUpper := strings.ToUpper(strings.TrimSpace(query))

	var results []SearchResult

	for _, entry := range index {
		score := computeMatchScore(entry, queryLower, queryUpper)
		if score > 0 {
			results = append(results, SearchResult{
				Symbol:      entry.Symbol,
				CompanyName: entry.CompanyName,
				Exchange:    entry.Exchange,
				Sector:      entry.Sector,
				Score:       score,
			})
		}
	}

	// Sort by score descending (simple insertion sort for small result sets)
	for i := 1; i < len(results); i++ {
		for j := i; j > 0 && results[j].Score > results[j-1].Score; j-- {
			results[j], results[j-1] = results[j-1], results[j]
		}
	}

	if len(results) > limit {
		results = results[:limit]
	}

	return results, nil
}

// IndexSize returns the number of entries in the search index.
func (s *SearchService) IndexSize() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.index)
}

// computeMatchScore calculates a relevance score for a search entry against a query.
// Higher score = better match. Returns 0 for no match.
func computeMatchScore(entry SearchEntry, queryLower, queryUpper string) float64 {
	symbolUpper := strings.ToUpper(entry.Symbol)
	nameLower := strings.ToLower(entry.CompanyName)

	var score float64

	// Exact symbol match (highest priority)
	if symbolUpper == queryUpper {
		score += 100
	} else if strings.HasPrefix(symbolUpper, queryUpper) {
		// Symbol prefix match
		score += 80
	} else if strings.Contains(symbolUpper, queryUpper) {
		// Symbol contains query
		score += 60
	}

	// Company name matching
	if strings.Contains(nameLower, queryLower) {
		score += 40
		// Bonus for match at word boundary
		if strings.HasPrefix(nameLower, queryLower) {
			score += 20
		}
		// Bonus for shorter names (more specific match)
		if len(entry.CompanyName) < 30 {
			score += 5
		}
	}

	// Fuzzy matching: check if all query chars appear in order
	if score == 0 {
		fuzzyScore := fuzzyMatch(queryLower, nameLower)
		if fuzzyScore > 0 {
			score += fuzzyScore * 20
		}
		fuzzySymbol := fuzzyMatch(queryUpper, symbolUpper)
		if fuzzySymbol > 0 {
			score += fuzzySymbol * 30
		}
	}

	return score
}

// fuzzyMatch returns a score (0-1) for how well the query chars appear in order in the target.
func fuzzyMatch(query, target string) float64 {
	if len(query) == 0 || len(target) == 0 {
		return 0
	}

	qi := 0
	for ti := 0; ti < len(target) && qi < len(query); ti++ {
		if query[qi] == target[ti] {
			qi++
		}
	}

	if qi == len(query) {
		return float64(qi) / float64(len(target))
	}
	return 0
}
