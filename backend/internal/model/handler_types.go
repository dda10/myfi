package model

// ChatRequest represents a chat request to the AI agent.
type ChatRequest struct {
	Message      string `json:"message"`
	Symbol       string `json:"symbol"`
	Provider     string `json:"provider"`
	Model        string `json:"model"`
	ApiKey       string `json:"apiKey"`
	AwsAccessKey string `json:"awsAccessKey"`
	AwsSecretKey string `json:"awsSecretKey"`
	AwsRegion    string `json:"awsRegion"`
}

// ModelsRequest represents a request to list available models for a provider.
type ModelsRequest struct {
	Provider     string `json:"provider"`
	ApiKey       string `json:"apiKey"`
	AwsAccessKey string `json:"awsAccessKey"`
	AwsSecretKey string `json:"awsSecretKey"`
	AwsRegion    string `json:"awsRegion"`
}

// RSS represents an RSS feed.
type RSS struct {
	Channel Channel `xml:"channel"`
}

// Channel represents an RSS channel.
type Channel struct {
	Title       string `xml:"title"`
	Description string `xml:"description"`
	Items       []Item `xml:"item"`
}

// Item represents an RSS item.
type Item struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

// ScreenerStock represents a stock entry in the screener output (handler-level).
type ScreenerStock struct {
	Symbol     string  `json:"symbol"`
	Exchange   string  `json:"exchange"`
	MarketCap  float64 `json:"marketCap"`
	PE         float64 `json:"pe"`
	PB         float64 `json:"pb"`
	EVEBITDA   float64 `json:"evEbitda"`
	ReportTerm string  `json:"reportTerm"`
}
