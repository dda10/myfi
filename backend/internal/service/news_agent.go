package service

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"time"

	"myfi-backend/internal/model"

	"github.com/tmc/langchaingo/llms"
)

// ---------------------------------------------------------------------------
// News_Agent — fetches and summarizes relevant financial news
// ---------------------------------------------------------------------------
//
// Requirements satisfied:
//   - 11.1  Fetch news from CafeF RSS feed, filter by symbol mentions/keywords
//   - 11.2  Fallback to Google search when CafeF RSS insufficient
//   - 11.3  Return structured response with articles and sentiment summary
//   - 11.4  Limit to 10 most recent/relevant articles
//   - 11.5  Return empty result with unavailable flag if both sources fail

const (
	// CafeF RSS feed URL for stock market news
	cafeFRSSURL = "https://cafef.vn/rss/chung-khoan.rss"

	// Maximum articles to return per query
	maxArticles = 10

	// HTTP timeout for fetching RSS/search results
	newsHTTPTimeout = 15 * time.Second

	// Minimum articles from RSS before triggering Google fallback
	minRSSArticles = 3
)

// NewsAgent is the AI sub-agent responsible for fetching and summarizing
// financial news from CafeF RSS and Google search fallback.
type NewsAgent struct {
	httpClient *http.Client
}

// NewNewsAgent creates a NewsAgent with a configured HTTP client.
func NewNewsAgent() *NewsAgent {
	return &NewsAgent{
		httpClient: &http.Client{
			Timeout: newsHTTPTimeout,
		},
	}
}

// Name returns the agent identifier used by the orchestrator.
func (a *NewsAgent) Name() string { return "News_Agent" }

// Execute fetches news articles for the symbols in the query intent and returns
// an AgentMessage containing NewsAgentResponse in the payload.
func (a *NewsAgent) Execute(ctx context.Context, intent model.QueryIntent, llm llms.Model) (*model.AgentMessage, error) {
	symbols := intent.Symbols
	if len(symbols) == 0 {
		// If no symbols provided, fetch general market news
		symbols = []string{}
	}

	// Build keywords from symbols and common Vietnamese stock terms
	keywords := a.buildKeywords(symbols)

	// 1. Fetch from CafeF RSS feed (Requirement 11.1)
	articles, err := a.fetchCafeFRSS(ctx, keywords)
	if err != nil {
		log.Printf("[News_Agent] CafeF RSS fetch failed: %v", err)
	}

	// 2. If insufficient results, fallback to Google search (Requirement 11.2)
	if len(articles) < minRSSArticles {
		log.Printf("[News_Agent] RSS returned %d articles, triggering Google fallback", len(articles))
		googleArticles, googleErr := a.fetchGoogleSearch(ctx, symbols, keywords)
		if googleErr != nil {
			log.Printf("[News_Agent] Google search failed: %v", googleErr)
		} else {
			articles = a.mergeArticles(articles, googleArticles)
		}
	}

	// 3. Check if both sources failed (Requirement 11.5)
	if len(articles) == 0 && err != nil {
		return a.buildUnavailableResponse(), nil
	}

	// 4. Sort by relevance and recency, limit to 10 (Requirement 11.4)
	articles = a.rankAndLimit(articles, keywords)

	// 5. Generate summary using LLM (Requirement 11.3)
	summary := a.generateSummary(ctx, articles, symbols, llm)

	// Build response
	response := model.NewsAgentResponse{
		Articles: articles,
		Summary:  summary,
	}

	payload := make(map[string]interface{})
	payload["news"] = response

	msg := &model.AgentMessage{
		AgentName:   a.Name(),
		PayloadType: "news_data",
		Payload:     payload,
		Timestamp:   time.Now(),
	}

	log.Printf("[News_Agent] successfully fetched %d articles for symbols: %v", len(articles), symbols)
	return msg, nil
}

// ---------------------------------------------------------------------------
// CafeF RSS Feed Parsing
// ---------------------------------------------------------------------------

// cafeFRSSFeed represents the RSS feed structure from CafeF
type cafeFRSSFeed struct {
	XMLName xml.Name        `xml:"rss"`
	Channel cafeFRSSChannel `xml:"channel"`
}

type cafeFRSSChannel struct {
	Title       string         `xml:"title"`
	Description string         `xml:"description"`
	Items       []cafeFRSSItem `xml:"item"`
}

type cafeFRSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

// fetchCafeFRSS fetches and parses the CafeF RSS feed, filtering by keywords.
func (a *NewsAgent) fetchCafeFRSS(ctx context.Context, keywords []string) ([]model.NewsArticle, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, cafeFRSSURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create RSS request: %w", err)
	}

	req.Header.Set("User-Agent", "MyFi-NewsAgent/1.0")
	req.Header.Set("Accept", "application/rss+xml, application/xml, text/xml")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch RSS feed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("RSS feed returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read RSS body: %w", err)
	}

	var feed cafeFRSSFeed
	if err := xml.Unmarshal(body, &feed); err != nil {
		return nil, fmt.Errorf("failed to parse RSS XML: %w", err)
	}

	// Filter and convert articles
	var articles []model.NewsArticle
	for _, item := range feed.Channel.Items {
		// Filter by keywords if provided
		if len(keywords) > 0 && !a.matchesKeywords(item.Title+" "+item.Description, keywords) {
			continue
		}

		pubTime := a.parseRSSDate(item.PubDate)
		snippet := a.cleanHTML(item.Description)
		if len(snippet) > 200 {
			snippet = snippet[:200] + "..."
		}

		articles = append(articles, model.NewsArticle{
			Title:       a.cleanHTML(item.Title),
			Source:      "CafeF",
			URL:         item.Link,
			PublishedAt: pubTime,
			Snippet:     snippet,
		})
	}

	return articles, nil
}

// ---------------------------------------------------------------------------
// Google Search Fallback
// ---------------------------------------------------------------------------

// fetchGoogleSearch performs a web search for news about the given symbols.
// Note: This is a simplified implementation. In production, you would use
// Google Custom Search API or a similar service.
func (a *NewsAgent) fetchGoogleSearch(ctx context.Context, symbols []string, keywords []string) ([]model.NewsArticle, error) {
	if len(symbols) == 0 && len(keywords) == 0 {
		return nil, nil
	}

	// Build search query
	var queryParts []string
	for _, sym := range symbols {
		queryParts = append(queryParts, sym)
	}
	queryParts = append(queryParts, "chứng khoán", "tin tức")
	query := strings.Join(queryParts, " ")

	// Use DuckDuckGo HTML search as a fallback (no API key required)
	// In production, consider using Google Custom Search API
	searchURL := fmt.Sprintf("https://html.duckduckgo.com/html/?q=%s", url.QueryEscape(query))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, searchURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create search request: %w", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; MyFi-NewsAgent/1.0)")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute search: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("search returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read search response: %w", err)
	}

	// Parse search results (simplified HTML parsing)
	articles := a.parseSearchResults(string(body), keywords)
	return articles, nil
}

// parseSearchResults extracts article information from search HTML.
// This is a simplified parser for DuckDuckGo HTML results.
func (a *NewsAgent) parseSearchResults(html string, keywords []string) []model.NewsArticle {
	var articles []model.NewsArticle

	// Simple regex patterns to extract results
	// In production, use a proper HTML parser like goquery
	titlePattern := regexp.MustCompile(`<a[^>]*class="result__a"[^>]*href="([^"]*)"[^>]*>([^<]*)</a>`)
	snippetPattern := regexp.MustCompile(`<a[^>]*class="result__snippet"[^>]*>([^<]*)</a>`)

	titles := titlePattern.FindAllStringSubmatch(html, -1)
	snippets := snippetPattern.FindAllStringSubmatch(html, -1)

	for i, match := range titles {
		if len(match) < 3 {
			continue
		}

		articleURL := match[1]
		title := a.cleanHTML(match[2])

		snippet := ""
		if i < len(snippets) && len(snippets[i]) > 1 {
			snippet = a.cleanHTML(snippets[i][1])
		}

		// Filter by keywords
		if len(keywords) > 0 && !a.matchesKeywords(title+" "+snippet, keywords) {
			continue
		}

		// Extract source from URL
		source := a.extractDomain(articleURL)

		articles = append(articles, model.NewsArticle{
			Title:       title,
			Source:      source,
			URL:         articleURL,
			PublishedAt: time.Now(), // Search results don't always have dates
			Snippet:     snippet,
		})
	}

	return articles
}

// ---------------------------------------------------------------------------
// LLM Summary Generation
// ---------------------------------------------------------------------------

// generateSummary uses the LLM to create a concise summary of the news articles.
func (a *NewsAgent) generateSummary(ctx context.Context, articles []model.NewsArticle, symbols []string, llm llms.Model) string {
	if len(articles) == 0 {
		return "Không có tin tức liên quan được tìm thấy. (No relevant news found.)"
	}

	if llm == nil {
		// Fallback to simple summary without LLM
		return a.generateSimpleSummary(articles, symbols)
	}

	// Build prompt for LLM
	var articleTexts []string
	for i, art := range articles {
		if i >= 5 { // Limit to first 5 for summary
			break
		}
		articleTexts = append(articleTexts, fmt.Sprintf("- %s: %s", art.Title, art.Snippet))
	}

	symbolsStr := "thị trường chung"
	if len(symbols) > 0 {
		symbolsStr = strings.Join(symbols, ", ")
	}

	prompt := fmt.Sprintf(`Bạn là chuyên gia phân tích tin tức tài chính. Hãy tóm tắt ngắn gọn (2-3 câu) các tin tức sau về %s và đánh giá tâm lý thị trường (tích cực/tiêu cực/trung lập):

%s

Trả lời bằng tiếng Việt, ngắn gọn và súc tích.`, symbolsStr, strings.Join(articleTexts, "\n"))

	// Call LLM
	response, err := llms.GenerateFromSinglePrompt(ctx, llm, prompt)
	if err != nil {
		log.Printf("[News_Agent] LLM summary generation failed: %v", err)
		return a.generateSimpleSummary(articles, symbols)
	}

	return response
}

// generateSimpleSummary creates a basic summary without LLM.
func (a *NewsAgent) generateSimpleSummary(articles []model.NewsArticle, symbols []string) string {
	if len(articles) == 0 {
		return "Không có tin tức liên quan."
	}

	symbolsStr := "thị trường"
	if len(symbols) > 0 {
		symbolsStr = strings.Join(symbols, ", ")
	}

	return fmt.Sprintf("Tìm thấy %d tin tức liên quan đến %s. Tin mới nhất: %s",
		len(articles), symbolsStr, articles[0].Title)
}

// ---------------------------------------------------------------------------
// Helper Functions
// ---------------------------------------------------------------------------

// buildKeywords creates a list of keywords from symbols and common terms.
func (a *NewsAgent) buildKeywords(symbols []string) []string {
	keywords := make([]string, 0, len(symbols)*2)

	for _, sym := range symbols {
		keywords = append(keywords, strings.ToUpper(sym))
		keywords = append(keywords, strings.ToLower(sym))
	}

	return keywords
}

// matchesKeywords checks if text contains any of the keywords.
func (a *NewsAgent) matchesKeywords(text string, keywords []string) bool {
	if len(keywords) == 0 {
		return true // No filter
	}

	textLower := strings.ToLower(text)
	for _, kw := range keywords {
		if strings.Contains(textLower, strings.ToLower(kw)) {
			return true
		}
	}
	return false
}

// parseRSSDate parses various RSS date formats.
func (a *NewsAgent) parseRSSDate(dateStr string) time.Time {
	formats := []string{
		time.RFC1123Z,
		time.RFC1123,
		"Mon, 02 Jan 2006 15:04:05 -0700",
		"Mon, 02 Jan 2006 15:04:05 MST",
		"2006-01-02T15:04:05Z",
		"2006-01-02 15:04:05",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t
		}
	}

	return time.Now() // Default to now if parsing fails
}

// cleanHTML removes HTML tags and decodes entities.
func (a *NewsAgent) cleanHTML(s string) string {
	// Remove HTML tags
	tagPattern := regexp.MustCompile(`<[^>]*>`)
	s = tagPattern.ReplaceAllString(s, "")

	// Decode common HTML entities
	replacer := strings.NewReplacer(
		"&amp;", "&",
		"&lt;", "<",
		"&gt;", ">",
		"&quot;", "\"",
		"&#39;", "'",
		"&nbsp;", " ",
	)
	s = replacer.Replace(s)

	// Trim whitespace
	s = strings.TrimSpace(s)

	return s
}

// extractDomain extracts the domain name from a URL.
func (a *NewsAgent) extractDomain(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "Web"
	}

	host := parsed.Host
	// Remove www. prefix
	host = strings.TrimPrefix(host, "www.")

	// Extract main domain
	parts := strings.Split(host, ".")
	if len(parts) >= 2 {
		return parts[len(parts)-2] + "." + parts[len(parts)-1]
	}

	return host
}

// mergeArticles combines articles from multiple sources, removing duplicates.
func (a *NewsAgent) mergeArticles(primary, secondary []model.NewsArticle) []model.NewsArticle {
	seen := make(map[string]bool)
	var merged []model.NewsArticle

	// Add primary articles first
	for _, art := range primary {
		key := strings.ToLower(art.Title)
		if !seen[key] {
			seen[key] = true
			merged = append(merged, art)
		}
	}

	// Add secondary articles
	for _, art := range secondary {
		key := strings.ToLower(art.Title)
		if !seen[key] {
			seen[key] = true
			merged = append(merged, art)
		}
	}

	return merged
}

// rankAndLimit sorts articles by relevance and recency, then limits to maxArticles.
func (a *NewsAgent) rankAndLimit(articles []model.NewsArticle, keywords []string) []model.NewsArticle {
	if len(articles) == 0 {
		return articles
	}

	// Score each article
	type scoredArticle struct {
		article model.NewsArticle
		score   float64
	}

	scored := make([]scoredArticle, len(articles))
	now := time.Now()

	for i, art := range articles {
		score := 0.0

		// Recency score (higher for newer articles)
		hoursSince := now.Sub(art.PublishedAt).Hours()
		if hoursSince < 24 {
			score += 10.0
		} else if hoursSince < 72 {
			score += 5.0
		} else if hoursSince < 168 {
			score += 2.0
		}

		// Keyword match score
		text := strings.ToLower(art.Title + " " + art.Snippet)
		for _, kw := range keywords {
			if strings.Contains(text, strings.ToLower(kw)) {
				score += 3.0
			}
		}

		// Source preference (CafeF is primary)
		if art.Source == "CafeF" {
			score += 2.0
		}

		scored[i] = scoredArticle{article: art, score: score}
	}

	// Sort by score descending
	sort.Slice(scored, func(i, j int) bool {
		return scored[i].score > scored[j].score
	})

	// Limit to maxArticles
	limit := maxArticles
	if len(scored) < limit {
		limit = len(scored)
	}

	result := make([]model.NewsArticle, limit)
	for i := 0; i < limit; i++ {
		result[i] = scored[i].article
	}

	return result
}

// buildUnavailableResponse creates a response indicating news data is unavailable.
func (a *NewsAgent) buildUnavailableResponse() *model.AgentMessage {
	response := model.NewsAgentResponse{
		Articles: []model.NewsArticle{},
		Summary:  "Không thể truy cập nguồn tin tức. Vui lòng thử lại sau. (News sources unavailable. Please try again later.)",
	}

	payload := make(map[string]interface{})
	payload["news"] = response
	payload["unavailable"] = true

	return &model.AgentMessage{
		AgentName:   "News_Agent",
		PayloadType: "news_data",
		Payload:     payload,
		Timestamp:   time.Now(),
	}
}
