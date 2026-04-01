# Requirements Document

## Introduction

EziStock is a Vietnamese stock quantitative analysis and AI-powered investment platform focused exclusively on stocks listed on HOSE, HNX, and UPCOM exchanges. The platform employs a microservices architecture with a Go backend (handling market data, portfolio, and market services via vnstock-go) and a Python microservice (hosting the multi-agent AI system and Alpha Mining engine using LangChain/LangGraph). The multi-agent AI system is the core differentiator, comprising specialized agents (Technical Analyst, News Analyst, Investment Advisor, Strategy Builder) that pre-process data into insights before feeding to LLMs, following the principle that the quantitative system underneath is the real differentiator while the agent is the storytelling layer. The Alpha Mining engine implements a four-layer architecture (Data Layer, Model Layer, Backtest Layer, Deployment Layer) for adaptive signal discovery that moves beyond fixed factors to data-driven, regime-aware learning with multi-strategy ensemble and consensus-based stock ranking. The frontend is Next.js with TypeScript and Tailwind CSS, featuring TradingView-style charting with lightweight-charts. The platform provides deep individual stock analysis with AI-generated investment theses and valuations, a market dashboard with sector heatmaps and hot topics, AI-powered stock ranking with configurable factor groups and backtesting, proactive investment ideas with buy/sell signals, a smart screener with fundamental and technical filters, portfolio tracking with performance analytics, and macro economic indicators — all inspired by the Miquant platform feature set. The agent chat is proactive (suggesting hot topics and detecting outliers), citation-backed (linking responses to the data platform), and supports Vietnamese language natively.

**Migration Note:** This spec replaces the previous "Unified Finance Platform" / "MyFi" spec. The following features from the old spec are REMOVED and their code should be cleaned up during implementation: gold price services (Doji API integration), cryptocurrency services (CoinGecko integration), savings tracker (term deposits, bank accounts), FX service (USD/VND conversion), commodity service, fund service, goal planner, bond tracking, multi-asset portfolio (now stock-only), comparison engine, and all references to asset types other than VN stocks. The vnstock-go library now supports VCI, VND, KBS, and DNSE connectors (Go 1.24+); references to the old "77 columns" / "28 columns" terminology are outdated. Gold prices use SJC + BTMC (not Doji). FMP and Binance connectors are stubs (not implemented).

## Glossary

- **EziStock_Platform**: The Vietnamese stock analysis web application composed of a Go backend, Python AI microservice, and Next.js frontend
- **Go_Backend**: The Go microservice (Gin framework) responsible for market data ingestion, portfolio management, and market services, using vnstock-go for Vietnamese stock data from VCI and KBS sources
- **Python_AI_Service**: The Python microservice hosting the multi-agent AI system and Alpha Mining engine, built with LangChain/LangGraph, communicating with the Go_Backend via gRPC or REST
- **Data_Source_Router**: The Go_Backend module that determines the optimal data source (VCI or KBS) for each data category based on data richness, and handles failover between sources
- **Market_Data_Service**: The Go_Backend service that consolidates all data categories from the vnstock-go API (listing, company, financial, trading, market, valuation, macro) into a unified data layer
- **Price_Service**: The Go_Backend service that fetches and caches near real-time stock prices from VCI and KBS sources via the Data_Source_Router
- **Portfolio_Service**: The Go_Backend module responsible for tracking stock holdings, transactions, cost basis, P&L calculations, and NAV computation for Vietnamese stock portfolios
- **Chart_Engine**: The frontend charting module built on lightweight-charts providing candlestick charts with technical indicators, drawing tools, and theme-adaptive rendering
- **Multi_Agent_System**: The LangChain/LangGraph-based AI system in the Python_AI_Service composed of specialized agents: Technical_Analyst_Agent, News_Analyst_Agent, Investment_Advisor_Agent, and Strategy_Builder_Agent
- **Technical_Analyst_Agent**: An AI agent that computes technical indicators (RSI, MACD, Bollinger Bands, SMA, EMA, ADX, Stochastic, ATR, OBV, MFI, etc.), identifies support/resistance levels, overbought/oversold conditions, and generates price predictions
- **News_Analyst_Agent**: An AI agent that searches and analyzes Vietnamese financial news, identifies catalysts and risk factors, and assesses market sentiment with source citations
- **Investment_Advisor_Agent**: An AI agent that synthesizes technical analysis and news analysis into structured investment recommendations with confidence scores and reasoning
- **Strategy_Builder_Agent**: An AI agent that develops trading strategies based on multi-agent analysis, optimizes entry/exit timing, and produces actionable trade plans
- **Alpha_Mining_Engine**: The four-layer engine in the Python_AI_Service for adaptive signal discovery: Data Layer (multi-dimensional signal space), Model Layer (ML/DL models), Backtest Layer (rolling backtests with regime-aware validation), Deployment Layer (multi-strategy ensemble with consensus ranking)
- **Signal_Space**: The multi-dimensional data space in the Alpha_Mining_Engine Data Layer encompassing price, volume, fundamentals, money flow, technical indicators, and macro signals
- **Regime_Detector**: The component within the Alpha_Mining_Engine that identifies current market regime (bull, bear, sideways, risk-on, risk-off) and adapts strategy weights accordingly
- **Strategy_Ensemble**: The Deployment Layer component that combines multiple strategies via voting/consensus to produce higher-confidence stock rankings
- **Alpha_Decay_Monitor**: The component that detects when discovered signals lose predictive power and triggers automatic rebalancing or signal replacement
- **Sector_Service**: The Go_Backend module that fetches and caches ICB sector index data (VNIT, VNIND, VNCONS, VNCOND, VNHEAL, VNENE, VNUTI, VNREAL, VNFIN, VNMAT), computes sector performance metrics, and maps stocks to their ICB sector classification
- **ICB_Sector**: The Industry Classification Benchmark sector assigned to each stock on the VN market, provided by the vnstock-go API
- **Sector_Trend**: A computed metric indicating whether a sector index is in an uptrend, downtrend, or sideways state based on price action over a specified time period
- **Screener_Service**: The Go_Backend module providing advanced stock filtering with fundamental, technical, and sector-based filter criteria
- **AI_Ranking_Service**: The service that implements configurable factor-based stock ranking with backtesting, supporting factor groups (Quality, Value, Growth, Momentum, Volatility), universe selection (VN30, VN100, HOSE, HNX, UPCOM), and rebalancing frequency configuration
- **Investment_Idea**: A proactive buy/sell recommendation generated by the Multi_Agent_System containing: symbol, signal direction, entry/exit prices, confidence score, quantitative reasoning, and historical accuracy tracking
- **Smart_Money_Flow**: A composite metric tracking net foreign investor and institutional investor buying/selling activity for a given symbol, used as a proxy for informed capital movement
- **VCI_Source**: Vietcap Securities HTTP API and GraphQL endpoint providing OHLCV history, real-time quotes, listings, index data, company profiles, officers, financial statements, and financial ratios
- **KBS_Source**: KB Securities HTTP API providing OHLCV history, real-time quotes, listings, index data, company profiles, officers, financial statements, financial ratios, plus KBS-specific methods for company events, company news, insider trading, symbols by group/industry classification
- **VND_Source**: VNDirect dchart API providing OHLCV history and listings as a third data source option
- **NAV**: Net Asset Value — the total value of all stock holdings in the user's portfolio, denominated in VND
- **Transaction_Ledger**: The record of all buy, sell, dividend, and corporate action events for stock holdings
- **Knowledge_Base**: A persistent store of market observations, detected patterns, alpha signals, and their outcomes used to improve recommendations over time
- **Watchlist_Service**: The Go_Backend module that persists named watchlists with per-symbol price alert thresholds
- **Performance_Engine**: The Go_Backend module computing portfolio performance metrics including TWR, MWRR/XIRR, NAV equity curve, and benchmark comparisons against VN-Index and VN30
- **Risk_Service**: The Go_Backend module computing portfolio-level and per-holding risk metrics including Sharpe ratio, max drawdown, portfolio beta, volatility, and Value at Risk
- **Corporate_Action_Service**: The Go_Backend module that fetches dividend calendars, stock split events, and bonus share events, and auto-adjusts cost basis in the Transaction_Ledger
- **Macro_Service**: The Go_Backend module that fetches and serves Vietnamese macroeconomic indicators (interbank rates, government bond yields, FX rates, CPI, GDP growth)
- **Auth_Service**: The Go_Backend module handling user authentication with JWT tokens, bcrypt password hashing, and session management
- **Heatmap_Engine**: The frontend component rendering market sector heatmap visualizations with color-coded performance by sector and individual stock
- **Analyst_IQ_Service**: The service that aggregates analyst reports from Vietnamese brokerages, scores analyst accuracy, and provides consensus recommendations
- **Research_Service**: The Python_AI_Service module that generates periodic research reports (Factor Snapshots, Sector Deep-Dives, Market Outlooks) using data from the Alpha_Mining_Engine and Multi_Agent_System, with PDF export capability
- **Mission**: A user-defined scheduled monitoring task with a trigger condition (price threshold, schedule, event), target symbols, action (alert, report, agent analysis), and notification preference; evaluated by the Go_Backend scheduler
- **Global_Search**: The frontend search component accessible via ⌘K that provides instant fuzzy search across all VN stock symbols and company names with sub-200ms response time
- **Liquidity_Filter**: The Go_Backend module that computes tradability scores (0-100) for each stock based on volume, trading value, bid-ask spread, zero-volume days, and free-float ratio, classifying stocks into Tier 1 (highly liquid), Tier 2 (moderate), and Tier 3 (illiquid) to prevent recommendations of untradeable stocks
- **Feedback_Loop_Engine**: The Python_AI_Service module that closes the loop between recommendations and outcomes, computing per-agent accuracy scores, detecting systematic biases, injecting historical performance context into agent prompts, and feeding outcome data back into Alpha_Mining_Engine model retraining


## Requirements

### Requirement 1: Microservices Architecture

**User Story:** As a developer, I want the platform to use a microservices architecture with Go for data/market services and Python for AI/ML workloads, so that each service can be independently scaled and use the best language for its domain.

#### Acceptance Criteria

1. THE EziStock_Platform SHALL be composed of three services: Go_Backend (market data, portfolio, market services), Python_AI_Service (multi-agent AI, Alpha Mining engine), and a Next.js frontend
2. THE Go_Backend SHALL communicate with the Python_AI_Service via gRPC for low-latency inter-service calls, with REST as a fallback
3. THE Go_Backend SHALL use the vnstock-go library (Go 1.24+) for all Vietnamese stock market data access, importing connectors via `_ "github.com/dda10/vnstock-go/all"` to register VCI, KBS, VND, and DNSE connectors
4. THE Python_AI_Service SHALL use LangChain and LangGraph for multi-agent orchestration and workflow management
5. THE EziStock_Platform SHALL use Docker Compose for local development with separate containers for Go_Backend, Python_AI_Service, PostgreSQL, and Redis
6. IF the Python_AI_Service is unavailable, THEN THE Go_Backend SHALL serve market data and portfolio features independently and return a service degradation notice for AI-dependent features
7. THE Go_Backend SHALL implement health check endpoints for both itself and the Python_AI_Service at GET /api/health
8. EACH service SHALL be packaged as a Docker container image suitable for deployment to Kubernetes (AWS EKS, GKE, or self-managed) or container orchestration platforms (AWS ECS, Docker Swarm)

### Requirement 2: Multi-Source Data Routing (VCI/KBS/VND)

**User Story:** As an investor, I want the platform to intelligently select the best data source (VCI, KBS, or VND) for each data category, so that I always get the most complete and reliable Vietnamese stock information.

#### Acceptance Criteria

1. THE Data_Source_Router SHALL maintain a source preference mapping that specifies the primary and fallback sources for each data category: OHLCV history (VCI primary, KBS fallback, VND tertiary), real-time quotes (VCI primary, KBS fallback), listings (VCI primary, KBS fallback, VND tertiary), index data (VCI primary, KBS fallback), company profiles (VCI primary, KBS fallback), officers (VCI primary, KBS fallback), financial statements (VCI primary, KBS fallback), financial ratios (VCI primary, KBS fallback), company events (KBS only), company news (KBS only), and insider trading (KBS only)
2. THE Data_Source_Router SHALL select the primary source for each data category based on which source provides more complete data fields for that category
3. WHEN the primary source for a data category fails to respond within 10 seconds, THE Data_Source_Router SHALL automatically route the request to the next fallback source in the preference chain
4. WHEN the primary source returns empty or incomplete data for a specific symbol, THE Data_Source_Router SHALL fetch from the fallback source and return whichever response contains more populated fields
5. IF all sources fail for a data category, THEN THE Data_Source_Router SHALL return the last cached result with a stale indicator flag
6. THE Data_Source_Router SHALL implement circuit breaker logic per source: after 3 consecutive failures within 60 seconds, THE Data_Source_Router SHALL skip that source until it recovers
7. THE Data_Source_Router SHALL log each source selection decision (chosen source, reason, response time) for monitoring
8. THE Data_Source_Router SHALL use KBS-specific methods (type assertion on the KBS connector) for data categories only available from KBS: company events, company news, insider trading, symbols by group, and symbols by industry

### Requirement 3: Real-Time Stock Price Service

**User Story:** As an investor, I want near real-time prices for Vietnamese stocks with automatic failover between data sources, so that I can see accurate valuations at any moment.

#### Acceptance Criteria

1. WHEN a price request is received for VN stocks, THE Price_Service SHALL fetch quotes via the Data_Source_Router using the vnstock-go client's RealTimeQuotes method (supported by VCI and KBS connectors)
2. IF the primary source returns zero or null prices, THEN THE Price_Service SHALL request the Data_Source_Router to fetch from the fallback source, then fall back to QuoteHistory using the last 10 days of OHLCV data
3. THE Price_Service SHALL cache VN stock prices with a 15-minute TTL
4. THE Price_Service SHALL batch multiple symbol requests into a single RealTimeQuotes call (vnstock-go fetches concurrently)
5. IF a data source API call fails after 3 retries with exponential backoff, THEN THE Price_Service SHALL return the last cached price with a stale indicator flag

### Requirement 4: Multi-Agent AI System Architecture (Core Feature)

**User Story:** As an investor, I want an AI system composed of specialized agents that pre-process data into insights before feeding to LLMs, so that I receive comprehensive, citation-backed investment advice grounded in quantitative analysis.

#### Acceptance Criteria

1. THE Multi_Agent_System SHALL be implemented in the Python_AI_Service using LangChain and LangGraph with four specialized agents: Technical_Analyst_Agent, News_Analyst_Agent, Investment_Advisor_Agent, and Strategy_Builder_Agent
2. THE Multi_Agent_System SHALL pre-process raw market data into structured insights before feeding to LLM agents, following the principle that the quantitative system is the differentiator and the agent is the storytelling layer
3. THE Multi_Agent_System SHALL use different LLM models for different tasks: lightweight models for data extraction, capable models for analysis, and conversational models for user-facing responses
4. THE Investment_Advisor_Agent SHALL synthesize outputs from Technical_Analyst_Agent and News_Analyst_Agent into unified investment recommendations with confidence scores and structured reasoning
5. THE Strategy_Builder_Agent SHALL develop trading strategies based on multi-agent analysis, optimizing entry/exit timing and producing actionable trade plans
6. IF any agent fails or times out within 30 seconds, THEN THE Multi_Agent_System SHALL proceed with outputs from remaining agents and note the missing data source in the final response
7. THE Multi_Agent_System SHALL support configurable LLM providers (OpenAI, Anthropic, Google, Qwen) via LangChain provider abstraction
8. THE Multi_Agent_System SHALL include a citation mechanism that cross-references agent responses against the data platform, linking each claim to its source data point

### Requirement 5: Technical Analyst Agent

**User Story:** As an investor, I want a dedicated AI agent that computes comprehensive technical indicators, identifies patterns, and generates price predictions, so that I have thorough technical analysis for every stock I evaluate.

#### Acceptance Criteria

1. WHEN the Technical_Analyst_Agent receives OHLCV data for a stock, THE Technical_Analyst_Agent SHALL compute the following indicators: RSI(14), MACD(12,26,9), Bollinger Bands(20,2), SMA(20), SMA(50), SMA(200), EMA(12), EMA(26), ADX(14), Stochastic(14,3,3), ATR(14), OBV, MFI(14), Aroon(25), Parabolic SAR(0.02,0.2), Supertrend(10,3), Williams %R(14), VWAP, and ROC(12)
2. THE Technical_Analyst_Agent SHALL identify support and resistance levels from historical price data using pivot points and volume profile analysis
3. THE Technical_Analyst_Agent SHALL classify overbought conditions (RSI > 70, MFI > 80) and oversold conditions (RSI < 30, MFI < 20) and detect divergences between indicators and price
4. THE Technical_Analyst_Agent SHALL generate a composite technical signal (strongly bullish, bullish, neutral, bearish, strongly bearish) by aggregating signals from all computed indicators with bullish vs bearish count
5. THE Technical_Analyst_Agent SHALL detect candlestick patterns including: hammer, engulfing, doji, morning star, evening star, and three white soldiers/black crows
6. THE Technical_Analyst_Agent SHALL identify stocks crossing key moving averages (MA50, MA100, MA200) and classify the crossover direction (golden cross, death cross)
7. THE Technical_Analyst_Agent SHALL compute Smart_Money_Flow by aggregating net foreign and institutional investor buy/sell volumes and classify as strong inflow, moderate inflow, neutral, moderate outflow, or strong outflow

### Requirement 6: News Analyst Agent

**User Story:** As an investor, I want a dedicated AI agent that searches and analyzes Vietnamese financial news with sentiment scoring and source citations, so that I understand the catalysts and risks affecting my stocks.

#### Acceptance Criteria

1. WHEN the News_Analyst_Agent receives stock symbols, THE News_Analyst_Agent SHALL fetch company news using the KBS connector's CompanyNews method (via the Go_Backend), and supplement with web search of Vietnamese financial news sources (CafeF, VnExpress, Vietstock) for broader market context
2. THE News_Analyst_Agent SHALL analyze each article to identify: catalysts (positive drivers), risk factors (negative drivers), and neutral information
3. THE News_Analyst_Agent SHALL produce a sentiment score (positive, negative, neutral) for each queried symbol with confidence level and supporting evidence
4. THE News_Analyst_Agent SHALL return structured responses with source citations linking each insight to the original article URL
5. THE News_Analyst_Agent SHALL limit returned articles to the 10 most recent and relevant items per query
6. IF news sources are unavailable, THEN THE News_Analyst_Agent SHALL return an empty result set with a flag indicating news data is unavailable

### Requirement 7: Investment Advisor Agent

**User Story:** As an investor, I want an AI agent that synthesizes technical and news analysis into clear investment recommendations with confidence scores, so that I receive actionable advice backed by quantitative reasoning.

#### Acceptance Criteria

1. WHEN the Investment_Advisor_Agent receives outputs from Technical_Analyst_Agent and News_Analyst_Agent, THE Investment_Advisor_Agent SHALL synthesize the information into a unified analysis context
2. THE Investment_Advisor_Agent SHALL produce recommendations containing: specific buy/sell/hold action, target price with upside/downside percentage, confidence score (0-100), risk assessment (low/medium/high), and structured reasoning referencing both technical and news factors
3. THE Investment_Advisor_Agent SHALL incorporate the user's current portfolio holdings and NAV from the Portfolio_Service to provide portfolio-aware recommendations
4. THE Investment_Advisor_Agent SHALL incorporate sector context from the Sector_Service including sector trend direction and the stock's relative performance within its ICB_Sector
5. THE Investment_Advisor_Agent SHALL query the Knowledge_Base for historical observations of similar patterns and incorporate historical accuracy into the analysis
6. IF the user's portfolio data is unavailable, THEN THE Investment_Advisor_Agent SHALL provide general market analysis without portfolio-specific recommendations

### Requirement 8: Strategy Builder Agent

**User Story:** As an investor, I want an AI agent that develops trading strategies with optimized entry/exit timing based on multi-agent analysis, so that I can execute well-timed trades with clear risk parameters.

#### Acceptance Criteria

1. WHEN the Strategy_Builder_Agent receives analysis from the Investment_Advisor_Agent, THE Strategy_Builder_Agent SHALL develop a trading strategy with specific entry price, stop-loss price (ATR-based), and take-profit price (risk-reward ratio based)
2. THE Strategy_Builder_Agent SHALL optimize entry/exit timing by combining technical signals, volume patterns, and market regime from the Regime_Detector
3. THE Strategy_Builder_Agent SHALL produce structured Trading_Signal outputs containing: symbol, entry price, stop-loss, take-profit, risk/reward ratio, confidence score, signal direction (long/short), suggested position size as percentage of NAV, and reasoning
4. THE Strategy_Builder_Agent SHALL produce structured Investment_Signal outputs for long-term recommendations containing: symbol, entry price zone (low-high), target price, suggested holding period, fundamental reasoning, and key metrics
5. THE Strategy_Builder_Agent SHALL factor in the current market regime (bull, bear, sideways) when calibrating position sizes and stop-loss distances

### Requirement 9: Alpha Mining Engine — Data Layer

**User Story:** As an investor, I want the platform to construct a multi-dimensional signal space from diverse data sources, so that the Alpha Mining engine can discover non-obvious predictive signals beyond traditional fixed factors.

#### Acceptance Criteria

1. THE Alpha_Mining_Engine Data Layer SHALL construct a Signal_Space encompassing the following dimensions: price signals (returns, momentum, mean reversion), volume signals (volume ratio, volume trend, unusual volume), fundamental signals (P/E, P/B, ROE, ROA, revenue growth, profit growth, debt-to-equity), money flow signals (foreign flow, institutional flow, net buy/sell), technical indicator signals (RSI, MACD, Bollinger Band width, ADX), and macro signals (interbank rates, VN-Index trend, sector rotation)
2. THE Alpha_Mining_Engine Data Layer SHALL normalize all signals to comparable scales using z-score normalization within each signal category
3. THE Alpha_Mining_Engine Data Layer SHALL handle missing data by forward-filling for up to 5 trading days and marking longer gaps as unavailable
4. THE Alpha_Mining_Engine Data Layer SHALL refresh signal data daily after market close (after 15:00 ICT) from the Go_Backend Market_Data_Service
5. THE Alpha_Mining_Engine Data Layer SHALL maintain a rolling history of signal values for a minimum of 3 years to support backtesting

### Requirement 10: Alpha Mining Engine — Model Layer

**User Story:** As an investor, I want the platform to use ML/DL models for adaptive signal discovery that detects non-linear interactions and adapts to market regimes, so that the system finds alpha signals that fixed-factor models miss.

#### Acceptance Criteria

1. THE Alpha_Mining_Engine Model Layer SHALL implement ML models (gradient boosting, random forest) for adaptive signal discovery from the Signal_Space
2. THE Alpha_Mining_Engine Model Layer SHALL detect non-linear interactions between signals that traditional linear factor models cannot capture
3. THE Alpha_Mining_Engine Model Layer SHALL implement regime-aware learning where model weights and feature importance adapt based on the current market regime identified by the Regime_Detector
4. THE Regime_Detector SHALL classify the Vietnamese market into regimes (bull, bear, sideways, risk-on, risk-off) using VN-Index trend, market breadth, foreign flow direction, and volatility levels
5. THE Alpha_Mining_Engine Model Layer SHALL retrain models on a configurable schedule (default: weekly) using rolling windows to prevent overfitting to stale patterns
6. THE Alpha_Mining_Engine Model Layer SHALL output ranked signal importance scores indicating which signals are currently most predictive for each market regime

### Requirement 11: Alpha Mining Engine — Backtest Layer

**User Story:** As an investor, I want the platform to validate discovered signals through rigorous backtesting with regime-aware validation and transaction cost simulation, so that I can trust the signals are robust and not artifacts of overfitting.

#### Acceptance Criteria

1. THE Alpha_Mining_Engine Backtest Layer SHALL implement rolling backtests using walk-forward analysis with configurable training and testing windows
2. THE Alpha_Mining_Engine Backtest Layer SHALL perform regime-aware validation by evaluating signal performance separately for each market regime (bull, bear, sideways)
3. THE Alpha_Mining_Engine Backtest Layer SHALL analyze signal stability by measuring consistency of signal rankings across multiple time periods
4. THE Alpha_Mining_Engine Backtest Layer SHALL simulate transaction costs including: broker commission (0.15-0.25% per trade), market impact for large orders, and slippage estimates for Vietnamese stocks
5. THE Alpha_Mining_Engine Backtest Layer SHALL compute backtest metrics including: cumulative return, annualized return, Sharpe ratio, max drawdown, win rate, profit factor, and information ratio vs VN-Index benchmark
6. THE Alpha_Mining_Engine Backtest Layer SHALL detect alpha decay by monitoring signal predictive power over time and flagging signals whose performance degrades below a configurable threshold
7. THE Alpha_Mining_Engine Backtest Layer SHALL present backtest results with cumulative returns charts, monthly and yearly performance tables, and drawdown visualization

### Requirement 12: Alpha Mining Engine — Deployment Layer

**User Story:** As an investor, I want the platform to combine multiple strategies via consensus voting to produce high-confidence stock rankings, so that I benefit from diversified signal sources rather than relying on a single strategy.

#### Acceptance Criteria

1. THE Alpha_Mining_Engine Deployment Layer SHALL implement a Strategy_Ensemble that combines multiple strategies via voting/consensus to produce stock rankings
2. THE Strategy_Ensemble SHALL cross-validate between strategies, requiring agreement from a configurable minimum number of strategies (default: 2 out of 3) before including a stock in the top ranking
3. THE Strategy_Ensemble SHALL produce a consensus-based stock ranking with composite scores reflecting the agreement level across strategies
4. THE Alpha_Decay_Monitor SHALL continuously monitor deployed signals for performance degradation and trigger automatic rebalancing when alpha decay is detected
5. THE Alpha_Decay_Monitor SHALL account for the Vietnamese market's faster alpha decay compared to developed markets by using shorter monitoring windows (default: 20 trading days)
6. THE Deployment Layer SHALL support continuous learning by feeding backtest outcomes back into the Model Layer for model improvement

### Requirement 13: AI-Powered Stock Ranking with Backtesting

**User Story:** As an investor, I want to configure factor-based stock rankings with customizable factor groups, universe selection, and backtesting, so that I can build and validate my own quantitative strategies.

#### Acceptance Criteria

1. THE AI_Ranking_Service SHALL support configurable factor groups including: Quality (ROE, ROA, profit margin), Value (P/E, P/B, EV/EBITDA), Growth (revenue growth, profit growth), Momentum (price momentum 1M/3M/6M/12M), and Volatility (standard deviation, beta)
2. THE AI_Ranking_Service SHALL support universe selection from: VN30, VN100, HOSE, HNX, UPCOM, or custom symbol lists
3. THE AI_Ranking_Service SHALL support configurable rebalancing frequency: monthly, quarterly, or semi-annually
4. THE AI_Ranking_Service SHALL allow users to customize factor weights within each factor group
5. WHEN a ranking configuration is submitted, THE AI_Ranking_Service SHALL compute factor scores for all stocks in the selected universe and produce a ranked list
6. THE AI_Ranking_Service SHALL provide backtesting for any ranking configuration, computing cumulative returns vs benchmark (VN-Index), monthly and yearly performance tables, and portfolio composition over time
7. THE AI_Ranking_Service SHALL display top holdings at each rebalancing period with entry/exit details

### Requirement 14: Proactive Investment Ideas

**User Story:** As an investor, I want the platform to proactively generate buy/sell investment ideas with entry/exit prices, confidence scores, and quantitative reasoning, so that I receive actionable opportunities without having to ask.

#### Acceptance Criteria

1. THE Multi_Agent_System SHALL proactively scan the VN stock universe (HOSE, HNX, UPCOM) on a configurable schedule (default: daily after market close) to identify investment opportunities
2. WHEN an opportunity is identified, THE Multi_Agent_System SHALL generate an Investment_Idea containing: symbol, signal direction (buy/sell), entry price, stop-loss price, take-profit price, confidence score (0-100), and structured reasoning backed by quantitative analysis
3. THE EziStock_Platform SHALL track historical accuracy of Investment_Ideas by recording the price outcome at 1-day, 7-day, 14-day, and 30-day intervals after idea generation
4. THE EziStock_Platform frontend SHALL display Investment_Ideas in a dedicated view sorted by confidence score with filtering by signal direction and recency
5. THE Multi_Agent_System SHALL detect outliers and unusual market conditions (volume spikes, price gaps, unusual foreign flow) and proactively suggest investigation topics to the user
6. THE EziStock_Platform SHALL deduplicate Investment_Ideas by not generating a new idea for the same symbol and direction within a 48-hour window unless the confidence score increases by 10 or more points

### Requirement 15: Deep Stock Analysis Page

**User Story:** As an investor, I want a comprehensive stock analysis page with AI-generated investment thesis, AI valuation with target price, technical signal summary, fundamental metrics, news sentiment, financial statements, and smart money flow tracking, so that I have all the information needed to make informed decisions on a single page.

#### Acceptance Criteria

1. THE EziStock_Platform frontend SHALL display a stock analysis page for each symbol containing: AI-generated investment thesis (luận điểm đầu tư), AI valuation with target price and upside/downside percentage, technical signal summary (bullish/bearish/neutral), fundamental metrics dashboard, news sentiment with source citations, financial statements (income, balance sheet, cash flow), revenue/profit trend charts, and Smart_Money_Flow tracking
2. WHEN a stock analysis page is loaded, THE Python_AI_Service SHALL generate an investment thesis by invoking the Multi_Agent_System pipeline for the requested symbol
3. THE EziStock_Platform frontend SHALL display the AI valuation with a clear target price, current price, and percentage upside/downside with visual indicator
4. THE EziStock_Platform frontend SHALL display technical signals as a summary card showing the composite signal (bullish/bearish/neutral) with individual indicator breakdowns
5. THE EziStock_Platform frontend SHALL display fundamental metrics in a dashboard format including: P/E, P/B, EV/EBITDA, ROE, ROA, revenue growth, profit growth, dividend yield, and debt-to-equity with comparison against sector averages
6. THE EziStock_Platform frontend SHALL display Smart_Money_Flow with foreign and institutional net buy/sell volumes and flow classification
7. THE EziStock_Platform frontend SHALL display financial statements with quarterly and yearly views for income statement, balance sheet, and cash flow statement

### Requirement 16: Market Dashboard

**User Story:** As an investor, I want a market dashboard showing hot topics, interbank rates, bond yields, MA crossover distribution, sector performance, AI valuation rankings, and global market context, so that I have a comprehensive market overview at a glance.

#### Acceptance Criteria

1. THE EziStock_Platform frontend SHALL display a market dashboard with the following sections: hot market topics/themes, interbank rates with yield curve, government bond yields with yield curve, top stocks crossing MA (MA50, MA100, MA200) with distribution chart, sector performance table (11 ICB sectors with percentage change), AI valuation rankings (top undervalued/overvalued stocks), global markets (US, Europe, Asia indices), and FX rates (DXY, EUR/USD, USD/VND)
2. THE Multi_Agent_System SHALL identify and surface hot market topics/themes (chủ đề nổi bật) by analyzing recent news, price movements, and sector rotations
3. THE Macro_Service SHALL provide interbank rate data and government bond yield data for yield curve rendering
4. THE Technical_Analyst_Agent SHALL compute MA crossover statistics across the entire VN stock universe and provide distribution data (number of stocks crossing MA50/MA100/MA200 up vs down)
5. THE Sector_Service SHALL provide sector performance data for all 11 ICB sectors with percentage change over configurable time periods
6. THE AI_Ranking_Service SHALL provide AI valuation rankings showing the most undervalued and overvalued stocks based on fundamental analysis
7. THE EziStock_Platform frontend SHALL auto-refresh market dashboard data every 5 minutes during trading hours (9:00-15:00 ICT) and every 30 minutes outside trading hours

### Requirement 17: Advanced Charting Engine

**User Story:** As a trader, I want TradingView-style charts with comprehensive technical indicators and drawing tools, so that I can perform thorough technical analysis directly in the app.

#### Acceptance Criteria

1. THE Chart_Engine SHALL render candlestick (OHLCV) charts using the lightweight-charts library with data fetched via the Go_Backend from the optimal source for OHLCV data
2. THE Chart_Engine SHALL support the following time intervals: 1 minute, 5 minutes, 15 minutes, 1 hour, 1 day, 1 week, 1 month
3. THE Chart_Engine SHALL support technical indicators including: SMA, EMA, VWAP, ADX, Aroon, Parabolic SAR, Supertrend, RSI, MACD, Stochastic, Bollinger Bands, Keltner Channel, ATR, OBV, MFI, and Linear Regression
4. WHEN a user selects an indicator, THE Chart_Engine SHALL compute the indicator values from OHLCV data and overlay the result on the chart (overlay indicators on price pane, oscillators in separate pane below)
5. THE Chart_Engine SHALL allow the user to configure parameters for each indicator and add multiple instances with different parameters
6. THE Chart_Engine SHALL provide drawing tools including: trend lines, horizontal lines, Fibonacci retracement, and rectangle selection
7. WHEN a user draws on the chart, THE Chart_Engine SHALL persist the drawing state in local storage
8. THE Chart_Engine SHALL display a volume histogram below the price chart synchronized with the same time axis

### Requirement 18: Smart Stock Screener

**User Story:** As an investor, I want an advanced stock screener with fundamental, technical, and sector filters with saveable presets, so that I can discover investment opportunities matching my strategy.

#### Acceptance Criteria

1. THE Screener_Service SHALL fetch stock data via the Data_Source_Router from the source providing the most complete financial data
2. THE Screener_Service SHALL support filtering by fundamental criteria: P/E ratio range, P/B ratio range, market capitalization minimum, EV/EBITDA range, ROE range, ROA range, revenue growth rate range, profit growth rate range, dividend yield range, and debt-to-equity ratio range
3. THE Screener_Service SHALL support filtering by ICB_Sector classification, allowing selection of one or more sectors
4. THE Screener_Service SHALL support filtering by exchange (HOSE, HNX, UPCOM) with multi-select
5. THE Screener_Service SHALL support filtering by technical criteria including: RSI range, MACD signal, MA crossover status, and volume anomaly detection
6. THE Screener_Service SHALL support sorting by multiple criteria with ascending or descending direction and pagination with configurable page size (default 20)
7. THE EziStock_Platform backend SHALL expose REST API endpoints for saving, listing, updating, and deleting filter presets per user (maximum 10 presets)
8. THE EziStock_Platform frontend SHALL provide built-in default filter presets (Value Investing, High Growth, High Dividend, Low Debt, Momentum) that the user can apply with one click
9. IF the Screener_Service cannot retrieve financial data for a symbol, THEN THE Screener_Service SHALL exclude that symbol from filtered results and log the data gap

### Requirement 19: Portfolio Tracking and Performance

**User Story:** As an investor, I want to track my Vietnamese stock portfolio with real-time P&L, performance analytics, and benchmark comparison, so that I can measure and optimize my investment performance.

#### Acceptance Criteria

1. WHEN a user records a buy transaction, THE Portfolio_Service SHALL credit the stock holding with the specified quantity and cost basis
2. WHEN a user records a sell transaction, THE Portfolio_Service SHALL debit the stock holding and compute realized P&L using the weighted average cost method
3. THE Portfolio_Service SHALL compute unrealized P&L for each holding by comparing the current market price from Price_Service against the weighted average cost basis
4. THE Portfolio_Service SHALL compute the total NAV by summing the current market value of all stock holdings
5. WHEN the NAV is computed, THE Portfolio_Service SHALL break down the allocation by sector as both absolute VND values and percentage of total NAV
6. THE Transaction_Ledger SHALL record each transaction with: symbol, quantity, unit price, total value, transaction date, and transaction type (buy, sell, dividend)
7. IF a sell transaction quantity exceeds the current holding quantity, THEN THE Portfolio_Service SHALL reject the transaction and return an error
8. THE Performance_Engine SHALL compute time-weighted return (TWR), money-weighted return (MWRR/XIRR), daily NAV snapshots, and equity curve data
9. THE Performance_Engine SHALL compute benchmark comparisons against VN-Index and VN30 over matching time periods
10. THE Risk_Service SHALL compute portfolio-level and per-holding risk metrics: Sharpe ratio, max drawdown, portfolio beta, annualized volatility, and Value at Risk (VaR)

### Requirement 20: Sector Heatmap and Sector Analysis

**User Story:** As an investor, I want a market sector heatmap visualization and sector performance analysis, so that I can identify which sectors are gaining or losing momentum and make sector-informed decisions.

#### Acceptance Criteria

1. THE Heatmap_Engine SHALL render a market heatmap visualization showing all stocks grouped by ICB sector with color-coded performance (green gradient for positive, red gradient for negative) sized by market capitalization
2. THE Sector_Service SHALL maintain a mapping of each VN stock symbol to its ICB_Sector classification using data from the vnstock-go API
3. THE Sector_Service SHALL fetch OHLCV history for all 10 ICB sector indices and compute performance metrics over: today, 1 week, 1 month, 3 months, 6 months, and 1 year
4. THE Sector_Service SHALL determine the Sector_Trend (uptrend, downtrend, sideways) for each ICB sector by comparing the sector index price against its SMA(20) and SMA(50)
5. THE Sector_Service SHALL compute median fundamental metrics (P/E, P/B, ROE, ROA, dividend yield) for stocks within each ICB_Sector to serve as sector averages
6. THE EziStock_Platform frontend SHALL display a sector performance table showing all 11 ICB sectors with percentage change and trend indicators
7. WHEN the user clicks on a sector in the heatmap, THE EziStock_Platform frontend SHALL display a detail panel showing: sector index chart, top 5 stocks by market cap in that sector, and sector median fundamentals
8. THE Sector_Service SHALL cache sector data with a 30-minute TTL during trading hours and 6-hour TTL outside trading hours

### Requirement 21: Proactive Agent Chat

**User Story:** As an investor, I want an AI chat that proactively suggests hot market topics and questions, provides citation-backed responses, and supports Vietnamese language natively, so that I can interact naturally with the platform's intelligence.

#### Acceptance Criteria

1. WHEN a user opens the agent chat, THE EziStock_Platform frontend SHALL display proactive suggestions including: hot market topics, trending stocks, and suggested questions based on the user's portfolio and recent market events
2. WHEN a user sends a message, THE Go_Backend SHALL route the request to the Python_AI_Service which orchestrates the full Multi_Agent_System pipeline
3. THE Multi_Agent_System SHALL return citation-backed responses where each claim links to its source data point on the EziStock_Platform (stock page, chart, financial data)
4. THE EziStock_Platform frontend chat SHALL render structured recommendations with distinct sections for: market data, technical analysis, news summary, and actionable advice
5. THE EziStock_Platform frontend chat SHALL support conversation history by sending the last 10 messages as context on each request
6. IF the Multi_Agent_System pipeline takes longer than 45 seconds, THEN THE Go_Backend SHALL return a partial response with whatever agent outputs have completed
7. THE Multi_Agent_System SHALL support Vietnamese language natively for both input and output, with the ability to switch to English
8. THE Multi_Agent_System SHALL detect outliers and unusual conditions in the user's portfolio or watched stocks and proactively surface them in the chat

### Requirement 22: Comprehensive Market Data Service

**User Story:** As an investor, I want the platform to provide access to all available data categories from the vnstock-go API through a unified data layer, so that the AI agents, dashboard, and features have comprehensive Vietnamese stock market data.

#### Acceptance Criteria

1. THE Market_Data_Service SHALL provide listing information using vnstock-go's Listing method, including: all stock symbols, market indices via IndexCurrent/IndexHistory (VN-Index, HNX-Index, UPCOM-Index, VN30, VN100), and exchange filtering (HOSE, HNX, UPCOM) — supported by VCI, VND, and KBS connectors
2. THE Market_Data_Service SHALL provide company information using vnstock-go's CompanyProfile method (returning shareholders, subsidiaries, ownership, charter capital history, labor structure) and Officers method — supported by VCI, KBS, and DNSE connectors
3. THE Market_Data_Service SHALL provide financial reports using vnstock-go's FinancialStatement method with type parameter (income, balance, cashflow) and period parameter (annual, quarterly) — supported by VCI and KBS connectors
4. THE Market_Data_Service SHALL provide trading statistics using vnstock-go's RealTimeQuotes and QuoteHistory methods with interval support (1m, 5m, 15m, 30m, 1H, 1D, 1W, 1M) — supported by VCI, VND, KBS, and DNSE connectors
5. THE Market_Data_Service SHALL provide market statistics using vnstock-go's IndexCurrent and IndexHistory methods for market indices and ICB sector indices — supported by VCI and KBS connectors
6. THE Market_Data_Service SHALL provide KBS-specific data using type assertion on the KBS connector: CompanyEvents (dividends, AGM, rights issues), CompanyNews, InsiderTrading, SymbolsByGroup (VN30, VN100, HNX30, etc.), and SymbolsByIndustry
7. THE Market_Data_Service SHALL expose a unified REST API with endpoints organized by data category: /api/market/listing, /api/market/company/{symbol}, /api/market/finance/{symbol}, /api/market/trading/{symbol}, /api/market/statistics, /api/market/valuation, and /api/market/macro
8. THE Market_Data_Service SHALL cache data with appropriate TTLs: listing (24h), company (6h), financial reports (24h), trading (15min), market statistics (30min), valuation (1h)
9. IF a data category is unavailable from the primary source, THEN THE Market_Data_Service SHALL attempt the alternative source and return last cached data with stale indicator if both fail

### Requirement 23: Macro Economics Dashboard

**User Story:** As an investor, I want a Vietnamese macroeconomic indicators dashboard showing interbank rates, bond yields, FX rates, and key economic data, so that I can understand the macro context affecting the stock market.

#### Acceptance Criteria

1. THE Macro_Service SHALL provide Vietnamese macroeconomic data including: interbank lending rates (overnight, 1W, 2W, 1M, 3M, 6M, 12M), government bond yields (1Y, 2Y, 3Y, 5Y, 10Y, 15Y, 30Y), FX rates (USD/VND, EUR/VND, JPY/VND, CNY/VND, DXY), CPI data, and GDP growth figures
2. THE EziStock_Platform frontend SHALL display interbank rates with a yield curve visualization
3. THE EziStock_Platform frontend SHALL display government bond yields with a yield curve visualization
4. THE EziStock_Platform frontend SHALL display FX rate trends with historical charts
5. THE Macro_Service SHALL cache macroeconomic data with a 1-hour TTL during trading hours
6. THE Multi_Agent_System SHALL incorporate macro data from the Macro_Service when generating investment recommendations to contextualize stock analysis within the broader economic environment

### Requirement 24: Analyst IQ — Analyst Report Aggregation

**User Story:** As an investor, I want aggregated analyst reports from Vietnamese brokerages with analyst accuracy scoring, so that I can evaluate the quality of analyst recommendations and incorporate consensus views.

#### Acceptance Criteria

1. THE Analyst_IQ_Service SHALL aggregate analyst reports and recommendations from major Vietnamese brokerages (SSI, VPS, HSC, VCBS, MBS, VNDS)
2. THE Analyst_IQ_Service SHALL track analyst accuracy by comparing past recommendations against actual price outcomes at 1-month, 3-month, and 6-month intervals
3. THE Analyst_IQ_Service SHALL compute a consensus recommendation (strong buy, buy, hold, sell, strong sell) and consensus target price for each covered stock
4. THE EziStock_Platform frontend SHALL display analyst reports with: analyst name, brokerage, recommendation, target price, report date, and accuracy score
5. THE Investment_Advisor_Agent SHALL incorporate analyst consensus data from the Analyst_IQ_Service when generating investment recommendations

### Requirement 25: Watchlist Management with Price Alerts

**User Story:** As an investor, I want to manage multiple named watchlists with per-symbol price alerts, so that I can organize stocks I'm tracking and get notified when prices hit my targets.

#### Acceptance Criteria

1. THE Watchlist_Service SHALL support multiple named watchlists per user with configurable symbol ordering
2. THE Watchlist_Service SHALL support per-symbol price alert thresholds (upper and lower bounds) for each watchlist entry
3. WHEN a stock price crosses a configured alert threshold, THE EziStock_Platform SHALL generate an in-app notification for the user
4. THE Watchlist_Service SHALL persist watchlist data in the database for cross-session access
5. THE EziStock_Platform frontend SHALL display watchlist symbols with real-time price updates, daily change, and alert status indicators

### Requirement 26: Knowledge Base for Pattern Learning

**User Story:** As an investor, I want the AI system to maintain a knowledge base of past market observations and their outcomes, so that the system improves its pattern detection and recommendations over time.

#### Acceptance Criteria

1. THE Knowledge_Base SHALL persist each observation from the Multi_Agent_System with: symbol, pattern type, detection date, confidence score, supporting data snapshot, and current price at detection time
2. THE Knowledge_Base SHALL track the outcome of each observation by recording the price change at 1-day, 7-day, 14-day, and 30-day intervals after detection
3. WHEN the Multi_Agent_System detects a new pattern, THE Multi_Agent_System SHALL query the Knowledge_Base for historical observations of the same pattern type and compute the historical success rate
4. THE Knowledge_Base SHALL provide a query interface filtered by symbol, pattern type, date range, and minimum confidence score
5. THE Knowledge_Base SHALL retain observations for a minimum of 2 years
6. THE Knowledge_Base SHALL compute aggregate accuracy metrics per pattern type: total observations, success count, failure count, average price change, and average confidence score

### Requirement 27: Corporate Action Tracking

**User Story:** As an investor, I want the platform to automatically track dividends, stock splits, and bonus shares and adjust my cost basis accordingly, so that my portfolio valuations remain accurate after corporate events.

#### Acceptance Criteria

1. THE Corporate_Action_Service SHALL fetch dividend calendars, stock split events, and bonus share events using the KBS connector's CompanyEvents method (via type assertion on the KBS connector) through the Data_Source_Router
2. WHEN a stock split or bonus share event occurs for a holding, THE Corporate_Action_Service SHALL auto-adjust the quantity and cost basis in the Transaction_Ledger
3. WHEN a dividend payment is recorded, THE Corporate_Action_Service SHALL record the dividend in the Transaction_Ledger and update the holding's dividend income
4. THE EziStock_Platform frontend SHALL display a calendar of upcoming corporate events (ex-dividend dates, payment dates, split dates) for stocks in the user's portfolio and watchlists
5. THE Corporate_Action_Service SHALL cache corporate action data with a 6-hour TTL

### Requirement 28: Data Persistence and API

**User Story:** As an investor, I want my portfolio data, transactions, knowledge base, and settings to persist across sessions, so that I do not lose my financial records or AI learning data.

#### Acceptance Criteria

1. THE Go_Backend SHALL persist all Portfolio_Service, Transaction_Ledger, Knowledge_Base, Watchlist_Service, and Screener filter preset data to PostgreSQL
2. WHEN the Go_Backend starts, THE Go_Backend SHALL run database migrations to create or update the required schema
3. THE Go_Backend SHALL expose CRUD REST API endpoints for: stock holdings, transactions, watchlists, filter presets, and user settings
4. THE EziStock_Platform frontend SHALL store user preferences (selected LLM provider, chart settings, theme) in local storage
5. IF a database write operation fails, THEN THE Go_Backend SHALL return an error with a descriptive message and ensure atomic transactions

### Requirement 29: User Authentication and Security

**User Story:** As a user, I want the platform to be protected by authentication with secure session management, so that my portfolio data is accessible only to me.

#### Acceptance Criteria

1. THE Auth_Service SHALL support local authentication with username and password, storing passwords hashed with bcrypt
2. THE Auth_Service SHALL issue JWT tokens upon successful authentication with a configurable expiry period (default 24 hours)
3. THE Go_Backend SHALL require a valid JWT token for all REST API endpoints except GET /api/health and POST /api/auth/login
4. WHEN no valid session exists, THE EziStock_Platform frontend SHALL display a login screen and prevent access to other views
5. IF 5 failed login attempts occur within 15 minutes for the same account, THEN THE Auth_Service SHALL lock the account for 30 minutes
6. THE Auth_Service SHALL auto-expire sessions after a configurable inactivity period (default 4 hours)

### Requirement 30: Dark Theme and Multi-Language Support

**User Story:** As a user, I want to toggle between light and dark themes and switch between Vietnamese and English, so that I can use the platform comfortably in my preferred language and visual mode.

#### Acceptance Criteria

1. THE EziStock_Platform frontend SHALL provide a theme toggle between light and dark modes with persistent preference in local storage
2. WHEN the theme toggle is activated, THE EziStock_Platform frontend SHALL update all UI components immediately without page reload, including chart colors
3. THE EziStock_Platform frontend SHALL provide a language selector between Vietnamese (vi-VN) and English (en-US) with persistent preference
4. WHEN the language is changed, THE EziStock_Platform frontend SHALL update all UI text immediately without page reload
5. THE EziStock_Platform frontend SHALL default to Vietnamese (vi-VN) for new users
6. THE EziStock_Platform frontend SHALL format numbers, dates, and currency according to the selected locale

### Requirement 31: Real-Time Price Updates on Frontend

**User Story:** As an investor, I want the dashboard to show near real-time price updates for my stocks without manual refresh, so that I always see current valuations.

#### Acceptance Criteria

1. THE EziStock_Platform frontend SHALL poll the Price_Service at configurable intervals: every 15 seconds during trading hours (9:00-15:00 ICT), every 300 seconds outside trading hours
2. WHEN new price data is received, THE EziStock_Platform frontend SHALL update the NAV display, allocation chart, and individual holding values without full page reload
3. THE EziStock_Platform frontend SHALL display a visual indicator showing price freshness: green for data less than 1 minute old, yellow for 1-5 minutes, red for older than 5 minutes
4. IF a price update request fails, THEN THE EziStock_Platform frontend SHALL retain the last known price and display a stale data warning icon

### Requirement 32: AI Recommendation Audit Trail

**User Story:** As an investor, I want every AI recommendation to be logged with full inputs, outputs, and outcome tracking, so that I can evaluate the AI's accuracy and the system can learn from past performance.

#### Acceptance Criteria

1. THE Knowledge_Base SHALL record every Investment_Advisor_Agent recommendation with: full input data (price, technical analysis, news analysis), output (recommended action, target price, confidence), and timestamp
2. THE Knowledge_Base SHALL track recommendation outcomes by recording price changes at 1-day, 7-day, 14-day, and 30-day intervals after the recommendation
3. THE EziStock_Platform frontend SHALL display a recommendation history view showing past recommendations with their outcomes and accuracy metrics
4. THE Multi_Agent_System SHALL incorporate recommendation accuracy data from the Knowledge_Base to calibrate future confidence scores

### Requirement 33: Export and Reporting

**User Story:** As an investor, I want to export my transaction history, P&L reports, and portfolio snapshots as CSV and PDF, so that I can maintain records and file taxes.

#### Acceptance Criteria

1. THE Go_Backend SHALL generate CSV exports of: transaction history, portfolio snapshot with current valuations, and P&L summary by holding
2. THE Go_Backend SHALL generate PDF exports of: portfolio report with charts, P&L summary, and tax-friendly capital gains summary (0.1% withholding on sell value for VN stocks)
3. THE EziStock_Platform frontend SHALL provide export buttons on portfolio, transaction, and P&L views
4. THE Go_Backend SHALL compute VN personal income tax liability for securities trading based on current regulations (0.1% on sell value)

### Requirement 34: Rate Limiting and API Quota Management

**User Story:** As a developer, I want centralized API rate limiting and quota management for upstream data sources, so that the platform does not get blocked by VCI or KBS for excessive requests.

#### Acceptance Criteria

1. THE Data_Source_Router SHALL enforce per-source request rate limits with configurable requests-per-second and requests-per-minute thresholds
2. WHEN the rate limit is reached for a source, THE Data_Source_Router SHALL queue excess requests and process them when capacity becomes available
3. IF the request queue depth exceeds a configurable maximum (default 100), THEN THE Data_Source_Router SHALL reject new requests with a rate limit exceeded error
4. THE Data_Source_Router SHALL log rate limit events for monitoring and tuning

### Requirement 35: Research Report Library

**User Story:** As an investor, I want access to a library of AI-generated research reports (factor snapshots, sector deep-dives, strategy performance reviews) with PDF download, so that I can study structured quantitative research and share it with others.

#### Acceptance Criteria

1. THE Python_AI_Service SHALL generate periodic research reports including: weekly Factor Snapshot (comparing strategy returns across Foreign, Value, Momentum, Quality, Low Volatility vs VNINDEX), monthly Sector Deep-Dive (sector rotation analysis with top picks), and quarterly Market Outlook (macro + technical + fundamental synthesis)
2. THE EziStock_Platform frontend SHALL display a Research section listing all published reports with: title, report type tag (Chiến lược / Ngành / Thị trường), publication date, author (EziStock Research Team), and thumbnail preview
3. THE EziStock_Platform frontend SHALL render research reports inline with charts, tables, and formatted text
4. THE EziStock_Platform frontend SHALL provide a "Tải PDF" (Download PDF) button for each report that generates and downloads a styled PDF version of the report
5. THE Factor Snapshot report SHALL include: average annual returns by strategy over 5Y/1Y/1M periods, yearly strategy return comparison table (showing which strategy outperformed each year), and strategy ranking by period
6. THE Research reports SHALL be generated on a configurable schedule (default: weekly for Factor Snapshot, monthly for Sector Deep-Dive) by the Python_AI_Service using data from the Alpha_Mining_Engine and Multi_Agent_System

### Requirement 36: Mission System — Scheduled Monitoring Tasks

**User Story:** As an investor, I want to create scheduled monitoring missions (price alerts, news monitoring, periodic stock scans) that run automatically and notify me, so that I stay informed without manually checking the platform.

#### Acceptance Criteria

1. THE EziStock_Platform SHALL support user-defined Missions, where each Mission is a scheduled monitoring task with: name, trigger type (price threshold, schedule-based, event-based), target symbols, action (alert, report, agent analysis), and notification preference (in-app, email)
2. THE EziStock_Platform SHALL support the following Mission types: Price Alert (notify when a stock crosses a price threshold), News Monitor (notify when news mentioning specified symbols is detected), Periodic Scan (run a screener or agent analysis on a schedule), and Portfolio Check (daily portfolio health summary)
3. WHEN a Mission trigger condition is met, THE EziStock_Platform SHALL execute the configured action and deliver a notification to the user
4. THE EziStock_Platform frontend SHALL display a "Nhiệm vụ" (Missions) tab in the top navigation showing all active missions with their status (active, paused, triggered), last trigger time, and next scheduled run
5. THE EziStock_Platform frontend SHALL provide a mission creation wizard allowing users to configure: mission name, trigger type and parameters, target symbols, action type, and notification preferences
6. THE Go_Backend SHALL implement a scheduler that evaluates Mission trigger conditions at configurable intervals (default: every 5 minutes during trading hours, every 30 minutes outside)
7. THE EziStock_Platform SHALL limit each user to a maximum of 20 active Missions
8. THE EziStock_Platform SHALL persist Mission configurations and trigger history in the database

### Requirement 37: Global Stock Search (⌘K)

**User Story:** As an investor, I want a global stock search accessible from anywhere in the app via keyboard shortcut (⌘K), so that I can quickly find and navigate to any Vietnamese stock.

#### Acceptance Criteria

1. THE EziStock_Platform frontend SHALL display a persistent search bar in the top navigation with placeholder text "Tìm kiếm cổ phiếu..." and a ⌘K keyboard shortcut indicator
2. WHEN the user presses ⌘K (or Ctrl+K on Windows), THE EziStock_Platform frontend SHALL open a focused search modal/overlay
3. THE search SHALL support fuzzy matching against: stock symbol (e.g., "HPG"), company name (e.g., "Hòa Phát"), and ICB sector name
4. THE search SHALL return results within 200ms by querying a pre-loaded in-memory index of all VN stock symbols and company names (fetched from Market_Data_Service listing endpoint and cached)
5. WHEN the user selects a search result, THE EziStock_Platform frontend SHALL navigate to the stock analysis page for that symbol
6. THE search results SHALL display: symbol, company name, exchange (HOSE/HNX/UPCOM), current price, and daily change percentage
7. THE search SHALL show recent searches (last 5) when opened with an empty query
8. THE search modal SHALL be dismissible via Escape key or clicking outside

### Requirement 38: Mobile Responsiveness

**User Story:** As a mobile user, I want the platform to work on my smartphone with touch-optimized interactions, so that I can check my portfolio and market data on the go.

#### Acceptance Criteria

1. THE EziStock_Platform frontend SHALL implement responsive design with breakpoints for mobile (< 768px), tablet (768px-1024px), and desktop (> 1024px)
2. THE Chart_Engine SHALL support touch gestures including: pinch-to-zoom, two-finger pan, and tap to show crosshair
3. THE EziStock_Platform frontend SHALL implement mobile-optimized navigation using a bottom navigation bar on mobile devices
4. THE EziStock_Platform frontend SHALL optimize touch target sizes to minimum 44x44 pixels for all interactive elements on mobile

### Requirement 39: Stock Quality Scoring and Liquidity Filter

**User Story:** As an investor, I want the platform to score every stock on tradability and quality, filtering out illiquid "trash" stocks from all AI recommendations, rankings, screener results, and investment ideas, so that I only see actionable stocks I can actually buy and sell without excessive slippage.

#### Acceptance Criteria

1. THE Go_Backend SHALL implement a Liquidity_Filter that computes a tradability score (0-100) for each stock based on: average daily trading volume over 20 days, average daily trading value (VND) over 20 days, bid-ask spread (where available), number of zero-volume trading days in the last 20 sessions, and free-float ratio
2. THE Liquidity_Filter SHALL classify stocks into tiers: Tier 1 (highly liquid, score ≥ 70), Tier 2 (moderately liquid, score 40-69), Tier 3 (illiquid, score < 40)
3. THE Liquidity_Filter SHALL apply configurable minimum thresholds with defaults: minimum average daily volume of 50,000 shares, minimum average daily value of 500 million VND, and maximum zero-volume days of 3 out of 20
4. THE Multi_Agent_System SHALL apply the Liquidity_Filter before generating any Investment_Idea or recommendation, excluding Tier 3 stocks from buy recommendations entirely
5. THE AI_Ranking_Service SHALL exclude stocks below the configured liquidity threshold from all ranking universes before computing factor scores
6. THE Screener_Service SHALL include a liquidity tier filter (Tier 1 only, Tier 1+2, All) with default set to Tier 1+2, and display the tradability score alongside each screener result
7. THE Investment_Advisor_Agent SHALL include a liquidity warning in recommendations for Tier 2 stocks, noting potential slippage risk and suggesting position size limits
8. THE Strategy_Builder_Agent SHALL adjust position sizing for Tier 2 stocks by reducing the maximum suggested position size by 50% compared to Tier 1 stocks, and SHALL never suggest positions in Tier 3 stocks
9. THE Alpha_Mining_Engine SHALL exclude Tier 3 stocks from the Signal_Space to prevent illiquid stocks from polluting signal discovery and backtest results
10. THE EziStock_Platform frontend SHALL display the tradability score and liquidity tier badge (green for Tier 1, yellow for Tier 2, red for Tier 3) on the stock analysis page, screener results, and investment ideas
11. THE Liquidity_Filter SHALL refresh tradability scores daily after market close (after 15:00 ICT) and cache results with a 24-hour TTL
12. THE EziStock_Platform frontend SHALL allow users to configure their personal liquidity threshold preference (strict: Tier 1 only, normal: Tier 1+2, relaxed: all tiers) which applies across all platform features

### Requirement 40: Data Storage Strategy (Cost-Optimized)

**User Story:** As a developer, I want a cost-optimized data storage strategy that minimizes database costs by using cheap object storage for bulk historical data and reserving expensive relational storage only for data that needs it, so that the platform can store years of market data without high infrastructure costs.

#### Acceptance Criteria

1. THE EziStock_Platform SHALL use PostgreSQL as the primary relational database ONLY for data requiring transactional integrity and complex queries: user accounts, portfolios, transactions, watchlists, filter presets, missions, active knowledge base observations (last 90 days), and corporate action records
2. THE EziStock_Platform SHALL use S3-compatible object storage (AWS S3 or MinIO for local dev) with Parquet columnar format as the primary store for bulk historical data including: OHLCV price history (3+ years), signal space values, backtest results, liquidity score history, and archived knowledge base observations (older than 90 days)
3. THE Python_AI_Service SHALL read historical data directly from S3 Parquet files using PyArrow or DuckDB for in-process analytical queries, avoiding the need to load bulk time-series data into PostgreSQL
4. THE Go_Backend SHALL implement a tiered data lifecycle: hot data (last 30 days of OHLCV, active prices) in Redis cache, warm data (last 90 days) in PostgreSQL, cold data (90+ days) in S3 Parquet files
5. THE Go_Backend SHALL run a nightly archival job (after market close) that moves data older than 90 days from PostgreSQL to S3 Parquet, keeping PostgreSQL lean and query-fast
6. THE EziStock_Platform SHALL use Redis as the caching layer for: real-time stock prices (15min TTL), market data (30min TTL), sector performance (30min TTL), listing data (24h TTL), session tokens, and rate limiter counters
7. THE EziStock_Platform SHALL use S3 for: generated research report PDFs, portfolio export files (CSV/PDF), trained ML model artifacts (versioned for rollback), and static frontend assets served via CloudFront CDN
8. THE Go_Backend SHALL implement a storage abstraction layer so that the choice between local filesystem (dev) and S3 (prod) is configurable via environment variables without code changes
9. THE EziStock_Platform SHALL use S3 Intelligent-Tiering or lifecycle policies to automatically move infrequently accessed data (older than 1 year) to S3 Infrequent Access, and data older than 3 years to S3 Glacier for minimal storage cost
10. THE EziStock_Platform SHALL implement database connection pooling in the Go_Backend (default: min 5, max 25 connections) and use the smallest viable PostgreSQL instance (e.g., db.t3.micro for dev, db.t3.small for prod) since bulk data lives in S3
11. THE EziStock_Platform SHALL implement a data retention policy: OHLCV in S3 retained indefinitely (Glacier after 3 years), signal space in S3 retained for 5 years, knowledge base hot data in PostgreSQL for 90 days then archived to S3 for 2 years, chat history retained for 1 year in PostgreSQL, audit logs retained for 3 years in S3
12. THE EziStock_Platform SHALL use database migrations (managed by the Go_Backend at startup) to create and evolve all PostgreSQL schemas, with migration files versioned in the repository
13. FOR local development, THE EziStock_Platform SHALL use Docker volumes for PostgreSQL and Redis persistence, and a local MinIO container as S3-compatible object storage with the same Parquet read/write paths as production

### Requirement 41: Continuous Learning Feedback Loop

**User Story:** As an investor, I want the AI agents and Alpha Mining models to continuously learn from their own past recommendation outcomes and backtest results, so that the system's accuracy improves over time and I can trust that poor-performing strategies are automatically corrected.

#### Acceptance Criteria

1. THE Python_AI_Service SHALL implement a Feedback_Loop_Engine that closes the loop between recommendations → outcomes → model improvement across both the Multi_Agent_System and Alpha_Mining_Engine
2. THE Feedback_Loop_Engine SHALL compute a rolling accuracy score for each agent (Technical_Analyst_Agent, News_Analyst_Agent, Investment_Advisor_Agent, Strategy_Builder_Agent) based on the outcome of their past recommendations at 7-day and 30-day intervals, stored in the Knowledge_Base
3. THE Investment_Advisor_Agent SHALL receive its own historical accuracy metrics as context in every prompt, including: overall hit rate (% of recommendations where price moved in the predicted direction), average confidence vs actual outcome correlation, and worst-performing recommendation patterns
4. THE Strategy_Builder_Agent SHALL adjust its default risk parameters (stop-loss distance, position sizing, confidence thresholds) based on the rolling backtest performance of its past Trading_Signals — widening stops if recent signals hit stop-loss too frequently, tightening if win rate is high
5. THE Alpha_Mining_Engine Model Layer SHALL incorporate recommendation outcome data as an additional training signal during weekly retraining, weighting recently successful signal combinations higher and penalizing signal combinations that led to failed recommendations
6. THE Feedback_Loop_Engine SHALL compute a per-pattern-type success rate from the Knowledge_Base (e.g., "accumulation patterns detected with confidence >70 had a 65% success rate at 14 days") and inject this as context into the Investment_Advisor_Agent's prompt when that pattern type is detected again
7. THE Feedback_Loop_Engine SHALL detect systematic biases in agent recommendations (e.g., consistently overestimating upside for a specific sector, or consistently wrong on short-term timing) and generate bias correction factors that are applied to future confidence scores
8. THE EziStock_Platform frontend SHALL display a "Model Performance" dashboard showing: per-agent accuracy trends over time, Alpha Mining model accuracy by regime, recommendation hit rate by confidence bucket (0-30, 30-60, 60-80, 80-100), and a comparison of predicted vs actual returns
9. THE Feedback_Loop_Engine SHALL run automatically after each outcome tracking update (1d, 7d, 14d, 30d) and store updated accuracy metrics in the Knowledge_Base
10. IF an agent's rolling 30-day accuracy drops below 40%, THEN THE Feedback_Loop_Engine SHALL flag the agent for review and reduce its weight in the Investment_Advisor_Agent's synthesis until accuracy recovers above 50%

### Requirement 42: Notification Delivery System

**User Story:** As an investor, I want to receive notifications via in-app bell icon, browser push, and email when missions trigger, price alerts fire, or investment ideas are generated, so that I never miss important market events.

#### Acceptance Criteria

1. THE EziStock_Platform SHALL support three notification channels: in-app (bell icon with unread badge), browser push notifications (Web Push API), and email (via SMTP or SendGrid/AWS SES)
2. THE EziStock_Platform frontend SHALL display a notification bell icon in the top navigation bar showing the count of unread notifications, with a dropdown panel listing recent notifications sorted by time
3. WHEN a Mission trigger fires, a price alert is hit, or a new Investment_Idea is generated, THE Go_Backend SHALL create a notification record and deliver it via the user's configured channels
4. THE EziStock_Platform SHALL allow users to configure notification preferences per channel: enable/disable each channel, quiet hours (e.g., no push/email between 22:00-07:00 ICT), and per-notification-type toggles (missions, price alerts, investment ideas, research reports)
5. THE Go_Backend SHALL persist all notifications in PostgreSQL with: type, title, body, link (deep link to relevant page), read/unread status, and created timestamp
6. THE EziStock_Platform frontend SHALL mark notifications as read when clicked, and provide a "mark all as read" action
7. THE Go_Backend SHALL implement email notifications using a configurable provider (SMTP for dev, SendGrid or AWS SES for prod) with HTML templates for: price alerts, investment ideas, mission triggers, and weekly portfolio summary
8. THE Go_Backend SHALL rate-limit notifications to prevent spam: maximum 1 notification per symbol per channel per hour for price alerts, maximum 10 email notifications per user per day

### Requirement 43: API Documentation and Contracts

**User Story:** As a developer, I want auto-generated API documentation with an interactive Swagger UI and protobuf contracts for inter-service communication, so that the frontend, Go backend, and Python AI service stay in sync and new developers can onboard quickly.

#### Acceptance Criteria

1. THE Go_Backend SHALL generate an OpenAPI 3.0 specification from annotated handler code (using swaggo/swag or similar) covering all REST API endpoints with: request/response schemas, query parameters, path parameters, authentication requirements, and example payloads
2. THE Go_Backend SHALL serve an interactive Swagger UI at GET /api/docs that renders the OpenAPI spec, allowing developers to browse and test API endpoints directly in the browser
3. THE Go_Backend SHALL serve the raw OpenAPI JSON spec at GET /api/docs/swagger.json for consumption by code generators and API clients
4. THE inter-service communication between Go_Backend and Python_AI_Service SHALL be defined using Protocol Buffer (.proto) files that specify all gRPC service methods, request messages, and response messages
5. THE .proto files SHALL be versioned in the repository under a shared /proto directory accessible to both Go and Python services
6. THE EziStock_Platform CI pipeline SHALL validate that the OpenAPI spec and protobuf definitions are up-to-date with the implementation by running spec generation and diff checks on every pull request
7. THE OpenAPI spec SHALL group endpoints by tag matching the platform features: Market Data, Portfolio, Screener, Watchlist, AI Chat, Ranking, Missions, Notifications, Auth, and Export

### Requirement 44: Distributed Tracing and Observability

**User Story:** As a developer, I want distributed tracing across Go backend and Python AI service calls, so that I can debug latency issues and trace a user request through the entire multi-agent pipeline.

#### Acceptance Criteria

1. THE Go_Backend and Python_AI_Service SHALL implement OpenTelemetry-compatible distributed tracing with trace ID propagation across gRPC and REST calls
2. EVERY user-facing API request SHALL generate a unique trace ID that is propagated to all downstream service calls, including Go → Python gRPC calls and Python → LLM API calls
3. THE Go_Backend SHALL include the trace ID in all structured log entries (JSON format) for correlation
4. THE EziStock_Platform SHALL support exporting traces to Jaeger, AWS X-Ray, or any OpenTelemetry-compatible collector via environment configuration
5. THE Python_AI_Service SHALL instrument each agent execution (Technical_Analyst, News_Analyst, Investment_Advisor, Strategy_Builder) as separate spans within the parent trace, recording: agent name, execution duration, LLM model used, token count, and success/failure status

### Requirement 45: Graceful Degradation and Offline Handling

**User Story:** As an investor, I want the platform to remain usable when the backend is slow, partially down, or my internet connection is unstable, so that I can still view cached data and recent analysis.

#### Acceptance Criteria

1. THE EziStock_Platform frontend SHALL implement a service worker that caches the most recent responses for: portfolio summary, watchlist data, last-viewed stock analysis, and market dashboard
2. WHEN the Go_Backend is unreachable, THE EziStock_Platform frontend SHALL serve cached data with a visible "Offline — showing cached data" banner and disable write operations (transactions, mission creation)
3. WHEN the Python_AI_Service is unavailable but the Go_Backend is healthy, THE EziStock_Platform SHALL display all market data and portfolio features normally, with AI-dependent features (chat, investment ideas, stock analysis thesis) showing a "AI service temporarily unavailable" placeholder
4. THE EziStock_Platform frontend SHALL implement optimistic UI updates for write operations (e.g., adding to watchlist) with automatic retry and rollback on failure
5. THE EziStock_Platform frontend SHALL display a connection status indicator (green/yellow/red) in the top navigation bar

### Requirement 46: End-to-End Testing Strategy

**User Story:** As a developer, I want an automated E2E testing strategy that validates critical user flows across frontend, Go backend, and Python AI service, so that regressions are caught before deployment.

#### Acceptance Criteria

1. THE EziStock_Platform SHALL implement E2E tests using Playwright (or Cypress) covering critical user flows: login → view dashboard → search stock → view analysis → add to watchlist → create mission → view portfolio
2. THE E2E test suite SHALL run against a Docker Compose environment with all services (Go_Backend, Python_AI_Service with mocked LLM responses, PostgreSQL, Redis, MinIO)
3. THE Python_AI_Service SHALL support a test mode where LLM calls are replaced with deterministic mock responses, enabling reproducible E2E tests without incurring LLM costs
4. THE Go_Backend SHALL implement integration tests for all REST API endpoints using httptest with an in-memory or test PostgreSQL database
5. THE CI pipeline SHALL run E2E tests on every pull request targeting the main branch, blocking merge on failure
6. THE E2E test suite SHALL complete within 10 minutes to maintain fast CI feedback loops

### Requirement 47: Input Validation and Security Hardening

**User Story:** As a developer, I want comprehensive input validation, XSS prevention, and CSRF protection across all API endpoints and frontend forms, so that the platform is resistant to common web security vulnerabilities.

#### Acceptance Criteria

1. THE Go_Backend SHALL validate all API request inputs using struct tag validation (e.g., go-playground/validator) and reject malformed requests with 400 Bad Request and descriptive error messages
2. THE Go_Backend SHALL use parameterized queries (never string concatenation) for all database operations to prevent SQL injection
3. THE Go_Backend SHALL sanitize all user-provided text fields (watchlist names, mission names, chat messages) to strip HTML/script tags before storage and rendering
4. THE EziStock_Platform frontend SHALL implement Content Security Policy (CSP) headers to prevent XSS attacks
5. THE Go_Backend SHALL implement CSRF protection for all state-changing endpoints using SameSite cookie attributes and Origin header validation
6. THE Go_Backend SHALL set security headers on all responses: X-Content-Type-Options: nosniff, X-Frame-Options: DENY, Strict-Transport-Security (HSTS) in production
7. THE Go_Backend SHALL log all authentication failures and suspicious input patterns (SQL injection attempts, XSS payloads) for security monitoring

### Requirement 48: LLM Cost Control and Token Budget

**User Story:** As a platform operator, I want to control and monitor LLM API costs with per-user token budgets, model selection optimization, and cost tracking, so that the platform remains financially sustainable as usage grows.

#### Acceptance Criteria

1. THE Python_AI_Service SHALL track token usage (input + output tokens) for every LLM call, recording: user ID, agent name, LLM model, token count, estimated cost (USD), and timestamp
2. THE Python_AI_Service SHALL enforce a configurable per-user daily token budget (default: 100,000 tokens/day for free tier, 500,000 for premium) and return a budget exceeded error when the limit is reached
3. THE Python_AI_Service SHALL implement model routing that selects the cheapest adequate model for each task: lightweight models (e.g., GPT-4o-mini, Claude Haiku) for data extraction and summarization, capable models (e.g., GPT-4o, Claude Sonnet) for analysis and recommendations
4. THE Python_AI_Service SHALL cache LLM responses for identical inputs (same symbol + same data snapshot) with a configurable TTL (default: 1 hour) to avoid redundant LLM calls
5. THE EziStock_Platform SHALL expose a cost dashboard (admin-only) showing: total daily/weekly/monthly LLM spend, per-user token consumption, per-agent cost breakdown, and cost trend charts
6. THE Python_AI_Service SHALL implement prompt optimization by compressing pre-processed insights to minimize input token count while preserving information density
7. IF a user exceeds 80% of their daily token budget, THEN THE EziStock_Platform SHALL display a warning in the chat interface indicating remaining budget

### Requirement 49: Legal Disclaimer and Investment Advice Compliance

**User Story:** As a platform operator, I want clear legal disclaimers that AI-generated recommendations are not financial advice, so that the platform complies with Vietnamese securities regulations and limits legal liability.

#### Acceptance Criteria

1. THE EziStock_Platform frontend SHALL display a persistent disclaimer footer on all pages containing AI-generated content (stock analysis, investment ideas, chat responses, AI ranking) stating: "Thông tin trên đây chỉ mang tính chất tham khảo, không phải lời khuyên đầu tư. EziStock không chịu trách nhiệm về quyết định đầu tư của bạn." (English: "The information above is for reference only and does not constitute investment advice. EziStock is not responsible for your investment decisions.")
2. THE EziStock_Platform frontend SHALL require new users to acknowledge the disclaimer during first login before accessing AI features
3. EVERY AI-generated recommendation (Investment_Idea, chat response, stock analysis thesis) SHALL include an inline disclaimer tag linking to the full terms of service
4. THE EziStock_Platform SHALL maintain a Terms of Service page and Privacy Policy page accessible from the footer of every page
5. THE EziStock_Platform SHALL NOT use language implying guaranteed returns, certainty of predictions, or professional financial advisory status in any AI-generated content or marketing copy

### Requirement 50: Cloud Deployment and Scalability

**User Story:** As a developer, I want the platform to be deployable to AWS or any Kubernetes cluster with container orchestration, auto-scaling, and infrastructure-as-code, so that the platform can scale to handle growing users and remain highly available.

#### Acceptance Criteria

1. EACH service (Go_Backend, Python_AI_Service, Next.js frontend) SHALL have a production-ready multi-stage Dockerfile that produces minimal container images (Go: scratch/distroless, Python: slim, Next.js: standalone output)
2. THE EziStock_Platform SHALL provide Kubernetes manifests (or Helm charts) for deploying all services, including: Deployments, Services, Ingress, ConfigMaps, Secrets, HorizontalPodAutoscaler, and PersistentVolumeClaims
3. THE Go_Backend Deployment SHALL be configured with a HorizontalPodAutoscaler that scales between 2 and 10 replicas based on CPU utilization (target 70%) and request rate
4. THE Python_AI_Service Deployment SHALL be configured with a HorizontalPodAutoscaler that scales between 1 and 5 replicas based on CPU utilization (target 60%), with resource requests/limits accounting for ML model memory requirements
5. THE EziStock_Platform SHALL use managed PostgreSQL (e.g., AWS RDS, or a PostgreSQL StatefulSet with persistent volumes) and managed Redis (e.g., AWS ElastiCache, or a Redis Deployment with persistent volumes) for stateful services
6. THE EziStock_Platform SHALL provide a Docker Compose file for local development that mirrors the production architecture (Go_Backend, Python_AI_Service, PostgreSQL, Redis, Next.js frontend)
7. THE EziStock_Platform SHALL provide Terraform or CDK infrastructure-as-code for provisioning AWS resources including: EKS cluster (or ECS), RDS PostgreSQL, ElastiCache Redis, ALB/NLB load balancer, ECR container registry, S3 for research report PDFs, CloudFront for frontend static assets, and Route 53 DNS
8. THE Go_Backend and Python_AI_Service SHALL expose health check endpoints (liveness and readiness probes) compatible with Kubernetes health checking at /healthz (liveness) and /readyz (readiness)
9. THE EziStock_Platform SHALL implement centralized logging (stdout/stderr in JSON format) compatible with CloudWatch, ELK, or any log aggregator
10. THE EziStock_Platform SHALL implement structured metrics (Prometheus-compatible) for: request latency, error rates, cache hit ratios, agent response times, and data source availability
11. THE EziStock_Platform SHALL support environment-based configuration via environment variables or Kubernetes ConfigMaps/Secrets for: database connection strings, Redis URLs, LLM API keys, data source credentials, and service URLs
12. THE EziStock_Platform SHALL provide a CI/CD pipeline configuration (GitHub Actions) that: runs tests, builds container images, pushes to container registry, and deploys to the target Kubernetes cluster or AWS environment
13. THE Kubernetes Ingress SHALL terminate TLS and route traffic to the appropriate service based on path: / to frontend, /api/* to Go_Backend, with the Go_Backend proxying AI requests to Python_AI_Service internally

## Correctness Properties

1. **Data Source Failover**: For any data category, if the primary source fails, the system must return data from the fallback source or cached data — never an empty response when cached data exists.
2. **Portfolio NAV Consistency**: The NAV must always equal the sum of (quantity × current price) for all holdings. Buy/sell transactions must maintain double-entry consistency.
3. **Agent Partial Failure Tolerance**: If any sub-agent fails, the Multi_Agent_System must still produce a response using available agent outputs — never a total failure when at least one agent succeeds.
4. **Alpha Decay Detection**: When a signal's predictive power degrades below the configured threshold over the monitoring window, the Alpha_Decay_Monitor must flag it within one monitoring cycle.
5. **Strategy Ensemble Consensus**: A stock must only appear in the top ranking if it meets the minimum strategy agreement threshold (default 2/3).
6. **Investment Idea Deduplication**: No duplicate Investment_Idea for the same symbol and direction within 48 hours unless confidence increases by ≥10 points.
7. **Rate Limit Queue Bound**: The request queue depth must never exceed the configured maximum; excess requests must be rejected immediately.
8. **Mission Trigger Accuracy**: When a Mission's trigger condition is met, the notification must be delivered within one evaluation cycle of the condition being true.
9. **Search Latency**: Global stock search results must be returned within 200ms for any query against the full VN stock universe.
10. **Citation Integrity**: Every claim in an agent response that references a data point must link to a valid, verifiable source on the platform.
11. **Liquidity Gate**: No Investment_Idea, AI Ranking result, or Strategy_Builder_Agent recommendation shall ever include a Tier 3 (illiquid) stock. Tier 2 stocks must always carry a liquidity warning and reduced position sizing.
12. **Feedback Loop Convergence**: An agent whose 30-day rolling accuracy drops below 40% must have its synthesis weight reduced within one feedback cycle. Its weight must not be restored until accuracy recovers above 50%.