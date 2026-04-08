package market

// SearchEntry represents a single entry in the in-memory search index.
type SearchEntry struct {
	Symbol      string `json:"symbol"`
	CompanyName string `json:"companyName"`
	Exchange    string `json:"exchange"`
	Sector      string `json:"sector,omitempty"`
}

// SearchResult represents a single result from a search query.
type SearchResult struct {
	Symbol      string  `json:"symbol"`
	CompanyName string  `json:"companyName"`
	Exchange    string  `json:"exchange"`
	Sector      string  `json:"sector,omitempty"`
	Score       float64 `json:"score"`
}
