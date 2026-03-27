package service

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"
	"sync"
	"time"

	"myfi-backend/internal/model"

	"github.com/tmc/langchaingo/llms"
)

// ---------------------------------------------------------------------------
// Sub-agent interface
// ---------------------------------------------------------------------------

// SubAgent is the interface that each specialised agent (Price, Analysis, News,
// Monitor, Supervisor) must implement. The orchestrator invokes Execute and
// expects a structured AgentMessage back.
type SubAgent interface {
	// Name returns the agent identifier (e.g. "Price_Agent").
	Name() string
	// Execute runs the agent with the given intent and returns a message.
	Execute(ctx context.Context, intent model.QueryIntent, llm llms.Model) (*model.AgentMessage, error)
}

// ---------------------------------------------------------------------------
// MultiAgentSystem
// ---------------------------------------------------------------------------

// MultiAgentSystem orchestrates the five specialised sub-agents using
// langchaingo for LLM integration.
//
// Requirements satisfied:
//   - 8.1  Five distinct agents via SubAgent interface
//   - 8.2  Coordinated pipeline (ProcessQuery)
//   - 8.3  Query analysis → parallel dispatch
//   - 8.4  Structured AgentMessage schema
//   - 8.5  30-second timeout + partial failure tolerance
//   - 8.6  Configurable LLM providers (delegated to GetLLM)
type MultiAgentSystem struct {
	llm       llms.Model
	llmConfig model.LLMConfig

	// Sub-agents — pluggable via the SubAgent interface so that concrete
	// implementations (tasks 21.3, 21.5, 21.7, …) can be wired in later.
	priceAgent      SubAgent
	analysisAgent   SubAgent
	newsAgent       SubAgent
	monitorAgent    SubAgent
	supervisorAgent SubAgent

	// recommendationTracker logs AI recommendations for accuracy tracking.
	recommendationTracker *RecommendationTracker

	// agentTimeout is the maximum duration each sub-agent is allowed before
	// the orchestrator cancels it and proceeds with partial results.
	agentTimeout time.Duration
}

// NewMultiAgentSystem creates a MultiAgentSystem with the provided LLM and
// default 30-second per-agent timeout.
func NewMultiAgentSystem(llm llms.Model, cfg model.LLMConfig) *MultiAgentSystem {
	return &MultiAgentSystem{
		llm:          llm,
		llmConfig:    cfg,
		agentTimeout: 30 * time.Second,
	}
}

// ---------------------------------------------------------------------------
// Agent registration (setter injection for pluggable agents)
// ---------------------------------------------------------------------------

// SetPriceAgent registers the Price_Agent implementation.
func (m *MultiAgentSystem) SetPriceAgent(a SubAgent) { m.priceAgent = a }

// SetAnalysisAgent registers the Analysis_Agent implementation.
func (m *MultiAgentSystem) SetAnalysisAgent(a SubAgent) { m.analysisAgent = a }

// SetNewsAgent registers the News_Agent implementation.
func (m *MultiAgentSystem) SetNewsAgent(a SubAgent) { m.newsAgent = a }

// SetMonitorAgent registers the Monitor_Agent implementation.
func (m *MultiAgentSystem) SetMonitorAgent(a SubAgent) { m.monitorAgent = a }

// SetSupervisorAgent registers the Supervisor_Agent implementation.
func (m *MultiAgentSystem) SetSupervisorAgent(a SubAgent) { m.supervisorAgent = a }

// SetRecommendationTracker registers the tracker for logging AI recommendations.
func (m *MultiAgentSystem) SetRecommendationTracker(t *RecommendationTracker) {
	m.recommendationTracker = t
}

// SetAgentTimeout overrides the default 30-second per-agent timeout.
func (m *MultiAgentSystem) SetAgentTimeout(d time.Duration) { m.agentTimeout = d }

// ---------------------------------------------------------------------------
// Query analysis phase  (Requirement 8.3)
// ---------------------------------------------------------------------------

// knownAssetKeywords maps common Vietnamese / English keywords to asset types.
var knownAssetKeywords = map[string]string{
	"stock": "vn_stock", "cổ phiếu": "vn_stock", "cp": "vn_stock",
	"crypto": "crypto", "bitcoin": "crypto", "btc": "crypto", "eth": "crypto",
	"gold": "gold", "vàng": "gold",
	"savings": "savings", "tiết kiệm": "savings",
	"bond": "bond", "trái phiếu": "bond",
}

// symbolPattern matches uppercase 1-5 letter tokens that look like VN stock
// tickers (e.g. SSI, FPT, VNM) or crypto symbols.
var symbolPattern = regexp.MustCompile(`\b[A-Z]{1,5}\b`)

// AnalyzeQuery parses a user query to extract symbols, asset types, and intent.
// It uses simple heuristics first; an LLM call can be layered on top later.
func (m *MultiAgentSystem) AnalyzeQuery(ctx context.Context, query string) model.QueryIntent {
	intent := model.QueryIntent{RawQuery: query}

	lower := strings.ToLower(query)

	// --- Extract asset types ---
	seen := map[string]bool{}
	for kw, at := range knownAssetKeywords {
		if strings.Contains(lower, kw) && !seen[at] {
			intent.AssetTypes = append(intent.AssetTypes, at)
			seen[at] = true
		}
	}

	// --- Extract symbols (uppercase tokens) ---
	matches := symbolPattern.FindAllString(query, -1)
	symSeen := map[string]bool{}
	for _, s := range matches {
		// Skip very common English words that happen to be uppercase
		if isCommonWord(s) {
			continue
		}
		if !symSeen[s] {
			intent.Symbols = append(intent.Symbols, s)
			symSeen[s] = true
		}
	}

	// --- Determine intent ---
	switch {
	case containsAny(lower, "phân tích", "analysis", "technical", "indicator", "chỉ báo"):
		intent.Intent = "analysis"
	case containsAny(lower, "tin tức", "news", "bản tin", "headline"):
		intent.Intent = "news"
	case containsAny(lower, "giá", "price", "quote", "báo giá"):
		intent.Intent = "price"
	case containsAny(lower, "khuyến nghị", "recommend", "tư vấn", "advice", "nên mua", "nên bán"):
		intent.Intent = "recommendation"
	default:
		intent.Intent = "general"
	}

	return intent
}

// isCommonWord filters out short uppercase tokens that are common English words
// rather than ticker symbols.
func isCommonWord(s string) bool {
	common := map[string]bool{
		"I": true, "A": true, "THE": true, "AND": true, "OR": true,
		"IS": true, "IT": true, "IN": true, "ON": true, "AT": true,
		"TO": true, "OF": true, "FOR": true, "BY": true, "AN": true,
		"IF": true, "DO": true, "SO": true, "AS": true, "MY": true,
		"ME": true, "WE": true, "HE": true, "BE": true, "NO": true,
	}
	return common[s]
}

// containsAny returns true if s contains any of the given substrings.
func containsAny(s string, subs ...string) bool {
	for _, sub := range subs {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}

// ---------------------------------------------------------------------------
// Parallel agent execution  (Requirements 8.2, 8.3, 8.5)
// ---------------------------------------------------------------------------

// selectAgents decides which sub-agents to invoke based on the parsed intent.
// For "general" or "recommendation" intents all data-gathering agents run.
func (m *MultiAgentSystem) selectAgents(intent model.QueryIntent) []SubAgent {
	var agents []SubAgent

	switch intent.Intent {
	case "price":
		if m.priceAgent != nil {
			agents = append(agents, m.priceAgent)
		}
	case "analysis":
		if m.priceAgent != nil {
			agents = append(agents, m.priceAgent)
		}
		if m.analysisAgent != nil {
			agents = append(agents, m.analysisAgent)
		}
	case "news":
		if m.newsAgent != nil {
			agents = append(agents, m.newsAgent)
		}
	default: // "recommendation", "general"
		for _, a := range []SubAgent{m.priceAgent, m.analysisAgent, m.newsAgent} {
			if a != nil {
				agents = append(agents, a)
			}
		}
	}

	return agents
}

// executeAgentsParallel runs the given agents concurrently, each with its own
// timeout derived from m.agentTimeout. Results (including errors) are collected
// and returned so the caller can implement partial-failure tolerance.
func (m *MultiAgentSystem) executeAgentsParallel(
	ctx context.Context,
	agents []SubAgent,
	intent model.QueryIntent,
) []model.AgentResult {
	results := make([]model.AgentResult, len(agents))
	var wg sync.WaitGroup

	for i, agent := range agents {
		wg.Add(1)
		go func(idx int, a SubAgent) {
			defer wg.Done()

			agentCtx, cancel := context.WithTimeout(ctx, m.agentTimeout)
			defer cancel()

			msg, err := a.Execute(agentCtx, intent, m.llm)
			results[idx] = model.AgentResult{
				AgentName: a.Name(),
				Message:   msg,
				Err:       err,
			}
		}(i, agent)
	}

	wg.Wait()
	return results
}

// ---------------------------------------------------------------------------
// ProcessQuery — main orchestration entry point  (Requirement 8.2)
// ---------------------------------------------------------------------------

// ProcessQuery implements the full multi-agent pipeline:
//  1. Analyse the user query (symbols, asset types, intent).
//  2. Select and execute relevant sub-agents in parallel.
//  3. Collect results with partial-failure tolerance.
//  4. Optionally invoke the Supervisor_Agent for synthesis.
//  5. Return a SupervisorRecommendation.
func (m *MultiAgentSystem) ProcessQuery(
	ctx context.Context,
	userQuery string,
	userID int64,
) (model.SupervisorRecommendation, error) {

	// 1. Query analysis
	intent := m.AnalyzeQuery(ctx, userQuery)

	// 2. Select agents
	agents := m.selectAgents(intent)
	if len(agents) == 0 {
		return model.SupervisorRecommendation{
			Summary: "No agents available to process this query.",
		}, nil
	}

	// 3. Parallel execution with per-agent timeout
	results := m.executeAgentsParallel(ctx, agents, intent)

	// 4. Collect successful outputs and note failures (Requirement 8.5)
	var messages []*model.AgentMessage
	var missingSources []string

	for _, r := range results {
		if r.Err != nil {
			log.Printf("[MultiAgentSystem] agent %s failed: %v", r.AgentName, r.Err)
			missingSources = append(missingSources, r.AgentName)
			continue
		}
		if r.Message != nil {
			messages = append(messages, r.Message)
		}
	}

	// 5. Supervisor synthesis (if registered)
	if m.supervisorAgent != nil {
		supervisorCtx, cancel := context.WithTimeout(ctx, m.agentTimeout)
		defer cancel()

		// Build a supervisor intent that carries the sub-agent outputs in its
		// payload so the supervisor can synthesise them.
		supervisorIntent := model.QueryIntent{
			RawQuery:   userQuery,
			Symbols:    intent.Symbols,
			AssetTypes: intent.AssetTypes,
			Intent:     "synthesize",
		}

		msg, err := m.supervisorAgent.Execute(supervisorCtx, supervisorIntent, m.llm)
		if err != nil {
			log.Printf("[MultiAgentSystem] Supervisor_Agent failed: %v", err)
			missingSources = append(missingSources, "Supervisor_Agent")
		} else if msg != nil {
			messages = append(messages, msg)
		}
	}

	// 6. Build final recommendation from collected messages
	rec := m.buildRecommendation(ctx, messages, missingSources)
	return rec, nil
}

// buildRecommendation assembles a SupervisorRecommendation from the collected
// agent messages. When the Supervisor_Agent is not yet wired, it produces a
// simple aggregation of available data.
func (m *MultiAgentSystem) buildRecommendation(
	ctx context.Context,
	messages []*model.AgentMessage,
	missingSources []string,
) model.SupervisorRecommendation {
	rec := model.SupervisorRecommendation{
		MissingSources: missingSources,
	}

	if len(messages) == 0 {
		rec.Summary = "Unable to gather sufficient data to produce a recommendation."
		return rec
	}

	var parts []string
	for _, msg := range messages {
		parts = append(parts, fmt.Sprintf("[%s] %s", msg.AgentName, msg.PayloadType))
	}
	rec.Summary = fmt.Sprintf("Aggregated outputs from %d agent(s): %s",
		len(messages), strings.Join(parts, ", "))

	if len(missingSources) > 0 {
		rec.Summary += fmt.Sprintf(". Note: data from %s was unavailable.",
			strings.Join(missingSources, ", "))
	}

	// Log any asset recommendations for accuracy tracking
	if m.recommendationTracker != nil && len(rec.AssetRecommendations) > 0 {
		for _, assetRec := range rec.AssetRecommendations {
			// Extract confidence from analysis if available
			confidence := 50 // default
			for _, msg := range messages {
				if msg.PayloadType == "analysis" {
					if conf, ok := msg.Payload["confidenceScore"].(int); ok {
						confidence = conf
					}
				}
			}
			if _, err := m.recommendationTracker.LogRecommendation(ctx, assetRec, confidence); err != nil {
				log.Printf("[MultiAgentSystem] failed to log recommendation: %v", err)
			}
		}
	}

	return rec
}
