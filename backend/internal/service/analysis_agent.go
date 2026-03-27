package service

// ---------------------------------------------------------------------------
// Analysis_Agent — performs technical, fundamental, and sector-relative analysis
// ---------------------------------------------------------------------------
//
// Requirements satisfied:
//   - 10.1  Compute all 21 technical indicators
//   - 10.2  Use default parameters for key indicators
//   - 10.3  Identify support/resistance levels, volume anomalies
//   - 10.4  Evaluate fundamental metrics vs sector averages
//   - 10.5  Produce structured analysis summary with confidence score
//   - 10.6  Omit indicators with insufficient data
//   - 10.7  Retrieve ICB sector classification from Sector_Service
//   - 10.8  Compare stock vs sector performance (1w, 1m, 3m, 1y)
//   - 10.9  Compare fundamentals vs sector medians
//   - 10.10 Evaluate sector trend and factor into confidence
//   - 10.11 Detect sector rotation signals
//   - 10.12 Include sector context in recommendation
//   - 10.13 Generate composite signal (strongly bullish to strongly bearish)

// AnalysisAgent is the AI sub-agent responsible for technical analysis,
// fundamental analysis, and sector-relative analysis.
type AnalysisAgent struct {
	priceService  *PriceService
	sectorService *SectorService
}

// NewAnalysisAgent creates an AnalysisAgent with the required service dependencies.
func NewAnalysisAgent(priceService *PriceService, sectorService *SectorService) *AnalysisAgent {
	return &AnalysisAgent{
		priceService:  priceService,
		sectorService: sectorService,
	}
}

// Name returns the agent identifier used by the orchestrator.
func (a *AnalysisAgent) Name() string { return "Analysis_Agent" }
