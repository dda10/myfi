# Implementation Plan: Unified Finance Platform

## Overview

This implementation plan breaks down the unified finance platform into sequential, incremental tasks. The platform consolidates VN stocks, gold, crypto, savings, and bonds into a single dashboard with real-time price tracking, advanced charting with 21 technical indicators, and a multi-agent AI advisory system.

The backend is built with Go (Gin framework) using vnstock-go for Vietnamese stock market data. The frontend is Next.js 16 with TypeScript, Tailwind CSS, and lightweight-charts. The database is SQLite (dev) / PostgreSQL (prod).

Each task builds on previous work, with checkpoints to validate progress. Tasks marked with `*` are optional and can be skipped for faster MVP delivery.

## Tasks

- [x] 1. Backend infrastructure and database setup
  - Set up Go project structure with Gin framework
  - Configure database connection (SQLite dev, PostgreSQL prod)
  - Create database migration system
  - Implement all database schemas (users, assets, transactions, savings_accounts, nav_snapshots, pattern_observations, alerts, watchlists, watchlist_symbols, filter_presets, recommendation_audit_log, financial_goals, stock_sector_mapping, cache_entries)
  - Set up environment variable configuration
  - _Requirements: 20.1, 20.2_

- [ ] 2. Data layer foundation - Data_Source_Router and caching
  - [x] 2.1 Implement Data_Source_Router with source preference mapping
    - Create DataCategory enum and SourcePreference struct
    - Implement source selection logic for all 12 data categories
    - Implement timeout-based failover (10 second timeout)
    - Implement empty data failover with field completeness checking
    - _Requirements: 2.1, 2.2, 2.3, 2.4_
  
  - [ ]* 2.2 Write property tests for Data_Source_Router
    - **Property 6: Data Source Timeout Failover**
    - **Property 7: Empty Data Failover**
    - **Property 8: Cache Fallback on Total Failure**
    - **Property 9: Source Selection Logging**
    - **Validates: Requirements 2.3, 2.4, 2.8, 2.9**
  
  - [x] 2.3 Implement Rate_Limiter with per-source limits
    - Create RateLimiter struct with per-source RateLimit tracking
    - Implement request queuing when limits reached
    - Implement queue depth limit (max 100)
    - Expose rate limit metrics endpoint
    - _Requirements: 33.1, 33.2, 33.4, 33.6, 33.7_
  
  - [x] 2.4 Implement Circuit_Breaker pattern
    - Create CircuitBreaker struct with state machine (Closed, Open, HalfOpen)
    - Track consecutive failures per source (threshold: 3 failures in 60s)
    - Implement recovery logic (60s timeout, test request)
    - _Requirements: 2.10_
  
  - [ ]* 2.5 Write property test for Circuit_Breaker
    - **Property 10: Circuit Breaker State Transition**
    - **Validates: Requirements 2.10**
  
  - [x] 2.6 Implement Cache with TTL support
    - Create in-memory Cache struct with RWMutex
    - Implement Get/Set with expiration checking
    - Implement database-backed cache for persistence
    - _Requirements: 3.6_


- [ ] 3. Price services - VN stocks, crypto, gold
  - [x] 3.1 Implement Price_Service for VN stocks
    - Create PriceService struct with Data_Source_Router integration
    - Implement GetQuotes for batch stock price fetching via VCI/KBS
    - Implement GetHistoricalData for OHLCV data
    - Implement cache integration with 15-minute TTL
    - Implement retry logic with exponential backoff (3 attempts)
    - _Requirements: 3.1, 3.5, 3.6, 3.7, 3.8_
  
  - [ ]* 3.2 Write property tests for Price_Service
    - **Property 11: Cache TTL Enforcement**
    - **Property 12: Price Request Batching**
    - **Property 13: Retry Exhaustion Cache Fallback**
    - **Validates: Requirements 3.6, 3.7, 3.8**
  
  - [x] 3.3 Implement Gold_Service
    - Create GoldService struct
    - Implement Doji API integration with price conversion (×1000)
    - Implement SJC fallback endpoint
    - Implement cache with 1-hour TTL
    - _Requirements: 5.1, 5.2, 5.3, 5.4, 5.5_
  
  - [ ]* 3.4 Write property tests for Gold_Service
    - **Property 19: Gold Price Conversion**
    - **Property 20: Gold Source Failover**
    - **Property 21: Gold Cache Fallback**
    - **Validates: Requirements 5.2, 5.3, 5.4**
  
  - [x] 3.5 Implement crypto price fetching
    - Integrate CoinGecko API for crypto prices
    - Implement VND conversion
    - Implement cache with 5-minute TTL
    - _Requirements: 3.2_
  
  - [x] 3.6 Implement FX_Service for USD/VND rates
    - Fetch USD/VND from CoinGecko (USDT/VND pair)
    - Implement hardcoded fallback (25,400)
    - Implement cache with 1-hour TTL
    - _Requirements: 29.1, 29.2_

- [x] 4. Checkpoint - Verify price services
  - Test VCI/KBS stock price fetching with real symbols (VNM, FPT, SSI)
  - Test gold price fetching from Doji
  - Test crypto price fetching from CoinGecko
  - Verify cache TTL enforcement
  - Verify failover logic works correctly
  - Ensure all tests pass, ask the user if questions arise.

- [x] 5. Portfolio engine - Asset registry and transactions
  - [x] 5.1 Implement Asset_Registry
    - Create Asset struct with all required fields
    - Implement AddAsset with database persistence
    - Implement UpdateAsset with NAV recalculation
    - Implement DeleteAsset with cascade to Transaction_Ledger
    - Implement asset type validation (vn_stock, crypto, gold, savings, bond, cash)
    - Ensure all monetary values stored in VND
    - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.5, 1.6_
  
  - [ ]* 5.2 Write property tests for Asset_Registry
    - **Property 1: Asset Persistence Round-Trip**
    - **Property 2: Asset Update NAV Consistency**
    - **Property 3: Asset Deletion Cascade**
    - **Property 4: Invalid Asset Type Rejection**
    - **Property 5: VND Currency Invariant**
    - **Validates: Requirements 1.2, 1.3, 1.4, 1.5, 1.6**
  
  - [x] 5.3 Implement Transaction_Ledger
    - Create Transaction struct with all transaction types
    - Implement RecordTransaction with database persistence
    - Implement transaction type validation (buy, sell, deposit, withdrawal, interest, dividend)
    - _Requirements: 4.6_
  
  - [x] 5.4 Implement Portfolio_Engine core logic
    - Implement buy transaction processing (double-entry accounting)
    - Implement sell transaction processing with weighted average cost P&L
    - Implement unrealized P&L calculation
    - Implement NAV computation (sum of all holdings)
    - Implement allocation breakdown by asset type
    - Implement insufficient holdings validation
    - _Requirements: 4.1, 4.2, 4.3, 4.4, 4.5, 4.7_
  
  - [x] 5.5 Write property tests for Portfolio_Engine
    - **Property 14: Buy Transaction Double-Entry**
    - **Property 15: Sell Transaction P&L Computation**
    - **Property 16: Unrealized P&L Calculation**
    - **Property 17: NAV Aggregation**
    - **Property 18: Insufficient Holdings Rejection**
    - **Validates: Requirements 4.1, 4.2, 4.3, 4.4, 4.7**

- [ ] 6. Savings tracker and interest calculation
  - [x] 6.1 Implement Savings_Tracker
    - Create SavingsAccount struct
    - Implement compound interest calculation: A = P × (1 + r/n)^(n×t)
    - Implement maturity date detection
    - Include accrued interest in NAV calculation
    - Support bank current accounts with zero/specified interest
    - _Requirements: 6.1, 6.2, 6.3, 6.4, 6.5_
  
  - [ ]* 6.2 Write property tests for Savings_Tracker
    - **Property 22: Savings Interest Calculation**
    - **Property 23: Term Deposit Maturity Detection**
    - **Validates: Requirements 6.2, 6.3**

- [x] 7. Checkpoint - Verify portfolio engine
  - Test asset CRUD operations
  - Test buy/sell transactions with P&L calculation
  - Test NAV computation with multiple asset types
  - Test savings interest calculation
  - Verify database persistence
  - Ensure all tests pass, ask the user if questions arise.

- [x] 8. Sector service - ICB classification and performance
  - [x] 8.1 Implement Sector_Service
    - Create SectorService struct with stock-to-sector mapping
    - Implement GetStockSector for ICB classification lookup
    - Implement GetSectorPerformance for individual sector metrics
    - Implement GetAllSectorPerformances for all 10 ICB sectors
    - Fetch sector index OHLCV data via Data_Source_Router
    - Compute performance metrics (today, 1w, 1m, 3m, 6m, 1y)
    - Implement sector trend classification (uptrend/downtrend/sideways)
    - Compute sector median fundamentals (P/E, P/B, ROE, ROA, div yield, debt-to-equity)
    - Implement cache with 30-min TTL (trading hours), 6-hour TTL (off-hours)
    - Refresh stock-to-sector mapping daily at 9:00 ICT
    - _Requirements: 21.1, 21.2, 21.3, 21.4, 21.5, 21.6, 21.7, 21.8, 21.9_
  
  - [ ]* 8.2 Write property test for Sector_Service
    - **Property 39: Sector Trend Classification**
    - **Validates: Requirements 21.4**

- [x] 9. Market data service - Unified data layer
  - [x] 9.1 Implement Market_Data_Service for listing data
    - Fetch all stock symbols via Data_Source_Router
    - Fetch market indices (VN30, VN100, VNMID, VNSML, VNALL)
    - Fetch government bonds
    - Fetch exchange information (HOSE, HNX, UPCOM)
    - Implement cache with 24-hour TTL
    - _Requirements: 23.1_
  
  - [x] 9.2 Implement Market_Data_Service for company data
    - Fetch company overview/profile
    - Fetch major shareholders
    - Fetch management team/officers
    - Fetch company news
    - Implement cache with 6-hour TTL
    - _Requirements: 23.2_
  
  - [x] 9.3 Implement Market_Data_Service for financial reports
    - Fetch income statements (yearly/quarterly)
    - Fetch balance sheets (yearly/quarterly)
    - Fetch cash flow statements (yearly/quarterly)
    - Fetch financial ratios
    - Implement cache with 24-hour TTL
    - _Requirements: 23.3_
  
  - [x] 9.4 Implement Market_Data_Service for trading statistics
    - Integrate real-time quotes (via Price_Service)
    - Fetch OHLCV history (all intervals)
    - Fetch intraday tick data
    - Fetch order book/price depth
    - _Requirements: 23.4_
  
  - [x] 9.5 Implement Market_Data_Service for market statistics
    - Fetch market index data
    - Fetch ICB sector index data (via Sector_Service)
    - Compute market breadth (advancing vs declining)
    - Fetch foreign trading data
    - Implement cache with 30-minute TTL
    - _Requirements: 23.5_
  
  - [x] 9.6 Implement Market_Data_Service for valuation metrics
    - Compute market-level P/E, P/B, EV/EBITDA
    - Compute sector-level valuation metrics
    - Compute stock-level valuation metrics
    - Implement cache with 1-hour TTL
    - _Requirements: 23.6_
  
  - [x] 9.7 Implement Fund_Service for open fund data
    - Fetch fund list
    - Fetch fund NAV
    - Fetch fund performance metrics
    - _Requirements: 23.7_
  
  - [x] 9.8 Implement Commodity_Service
    - Integrate VN gold prices (via Gold_Service)
    - Fetch global gold OHLCV (if available)
    - Fetch crude oil, natural gas, steel, iron ore prices
    - Fetch agricultural commodity prices
    - Fetch VN pork prices
    - _Requirements: 23.8_
  
  - [x] 9.9 Implement Macro_Service
    - Fetch macroeconomic indicators for VN market
    - _Requirements: 23.9_
  
  - [x] 9.10 Create unified REST API endpoints
    - Implement /api/market/listing
    - Implement /api/market/company/:symbol
    - Implement /api/market/finance/:symbol
    - Implement /api/market/trading/:symbol
    - Implement /api/market/statistics
    - Implement /api/market/valuation
    - Implement /api/market/funds
    - Implement /api/market/commodities
    - Implement /api/market/macro
    - Support batch requests for trading statistics
    - _Requirements: 23.10, 23.11, 23.12, 23.13, 23.14, 23.15_

- [x] 10. Checkpoint - Verify market data services
  - Test sector classification and performance metrics
  - Test company data fetching
  - Test financial report fetching
  - Test market statistics
  - Verify cache behavior
  - Ensure all tests pass, ask the user if questions arise.

- [x] 11. Screener service with advanced filtering
  - [x] 11.1 Implement Screener_Service
    - Create ScreenerFilters struct with all filter criteria
    - Fetch stock data via Data_Source_Router (replace mocked data)
    - Implement filtering by fundamentals (P/E, P/B, market cap, EV/EBITDA, ROE, ROA, revenue growth, profit growth, div yield, debt-to-equity)
    - Implement filtering by ICB_Sector (multi-select)
    - Implement filtering by exchange (HOSE, HNX, UPCOM)
    - Implement filtering by Sector_Trend
    - Implement sorting (multiple criteria, asc/desc)
    - Implement pagination (default 20 per page)
    - Handle missing financial data gracefully (exclude symbol, log gap)
    - _Requirements: 18.1, 18.2, 18.3, 18.4, 18.5, 18.6, 18.7, 18.8, 18.15_
  
  - [x] 11.2 Implement Filter_Preset management
    - Implement SavePreset with database persistence
    - Implement GetPresets for user
    - Implement UpdatePreset
    - Implement DeletePreset
    - Limit to 10 presets per user
    - _Requirements: 18.9, 18.10_
  
  - [ ]* 11.3 Write property test for Screener_Service
    - **Property 37: Screener Graceful Degradation**
    - **Validates: Requirements 18.15**

- [x] 12. Comparison engine for stock analysis
  - [x] 12.1 Implement Comparison_Engine
    - Support up to 10 stocks simultaneously
    - Implement Valuation mode (P/E, P/B time-series)
    - Implement Performance mode (normalized % returns)
    - Implement Correlation mode (Pearson correlation matrix)
    - Support time periods: 3M, 6M, 1Y, 3Y, 5Y
    - Support sector-based stock grouping
    - Handle missing OHLCV data (exclude stock, return warning)
    - _Requirements: 24.1, 24.2, 24.3, 24.4, 24.5, 24.6, 24.18_
  
  - [ ]* 12.2 Write property test for Comparison_Engine
    - **Property 40: Correlation Matrix Symmetry**
    - **Validates: Requirements 24.4**

- [x] 13. Watchlist service with price alerts
  - [x] 13.1 Implement Watchlist_Service
    - Implement CreateWatchlist, RenameWatchlist, DeleteWatchlist
    - Implement AddSymbol, RemoveSymbol, ReorderSymbols
    - Implement per-symbol price alert thresholds (above/below)
    - Persist all watchlists in database
    - Sync watchlist symbols to Monitor_Agent scan list
    - Create default watchlist for new users
    - _Requirements: 25.1, 25.2, 25.3, 25.5, 25.6, 25.7_
  
  - [ ]* 13.2 Write property test for Watchlist_Service
    - **Property 41: Price Alert Triggering**
    - **Validates: Requirements 25.4**

- [x] 14. Performance and risk analytics
  - [x] 14.1 Implement Performance_Engine
    - Implement ComputeTWR (time-weighted return with chain-linking)
    - Implement ComputeMWRR (money-weighted return using IRR)
    - Implement StoreNAVSnapshot (daily at 15:00 ICT)
    - Implement GetEquityCurve
    - Fetch benchmark data (VN-Index, VN30) via Data_Source_Router
    - Compute benchmark comparison with alpha
    - Compute performance breakdown by asset type
    - _Requirements: 26.1, 26.2, 26.3, 26.4, 26.5, 26.6_
  
  - [ ]* 14.2 Write property tests for Performance_Engine
    - **Property 42: TWR Calculation Correctness**
    - **Property 43: MWRR Calculation Correctness**
    - **Validates: Requirements 26.1, 26.2**
  
  - [x] 14.3 Implement Risk_Service
    - Implement ComputeSharpeRatio (use VN risk-free rate 4.5% default)
    - Implement ComputeMaxDrawdown from NAV history
    - Implement ComputeBeta against VN-Index (1-year regression)
    - Implement annualized volatility (σ × √252)
    - Implement ComputeVaR at 95% confidence (historical simulation)
    - Compute per-holding risk contribution
    - _Requirements: 27.1, 27.2, 27.3, 27.4, 27.5, 27.8_
  
  - [ ]* 14.4 Write property tests for Risk_Service
    - **Property 44: Sharpe Ratio Calculation**
    - **Property 45: Maximum Drawdown Calculation**
    - **Property 46: Beta Calculation**
    - **Property 47: Volatility Calculation**
    - **Property 48: VaR Calculation**
    - **Validates: Requirements 27.1, 27.2, 27.3, 27.4, 27.5**

- [x] 15. Checkpoint - Verify analytics services
  - Test screener with various filter combinations
  - Test stock comparison across all modes
  - Test watchlist CRUD operations
  - Test TWR and MWRR calculations
  - Test risk metrics computation
  - Ensure all tests pass, ask the user if questions arise.

- [x] 16. Export and reporting service
  - [x] 16.1 Implement Export_Service
    - Implement CSV export for transaction history
    - Implement CSV export for portfolio snapshot
    - Implement PDF export for portfolio report
    - Implement tax report (capital gains by asset type and year)
    - Support date range filtering
    - Include both VND and USD columns in exports
    - _Requirements: 28.1, 28.2, 28.3, 28.4, 28.5, 29.7_

- [x] 17. Corporate action tracking
  - [x] 17.1 Implement Corporate_Action_Service
    - Fetch dividend calendar from VCI/KBS via Data_Source_Router
    - Fetch stock split and bonus share events
    - Auto-record dividend payments as transactions
    - Auto-adjust cost basis and quantity for splits/bonus shares
    - Track dividend history per holding
    - Compute yield-on-cost
    - _Requirements: 30.1, 30.2, 30.3, 30.7_

- [x] 18. Goal planning service
  - [x] 18.1 Implement Goal_Planner
    - Implement CreateGoal, UpdateGoal, DeleteGoal
    - Compute progress percentage (current NAV / target amount)
    - Compute required monthly contribution
    - Support goal categories (retirement, emergency_fund, property, education, custom)
    - Persist goals in database
    - _Requirements: 31.1, 31.2, 31.3, 31.6, 31.7_

- [x] 19. Backtesting engine
  - [x] 19.1 Implement Backtest_Engine
    - Accept strategy rules (entry/exit conditions, stop-loss, take-profit)
    - Run strategy against historical OHLCV data
    - Compute results (total return, win rate, max drawdown, Sharpe ratio, trades, avg holding period)
    - Support all 21 technical indicators in strategy rules
    - Provide preset strategies (RSI Oversold Bounce, MACD Crossover, Bollinger Band Squeeze)
    - _Requirements: 32.1, 32.2, 32.3, 32.5, 32.6_

- [x] 20. Checkpoint - Verify business logic services
  - Test export generation (CSV and PDF)
  - Test corporate action processing
  - Test goal planning calculations
  - Test backtesting with preset strategies
  - Ensure all tests pass, ask the user if questions arise.


- [ ] 21. AI agent system - Multi-agent architecture
  - [x] 21.1 Implement Multi_Agent_System orchestration
    - Create MultiAgentSystem struct with LLM integration (langchaingo)
    - Implement agent message schema (AgentMessage struct)
    - Implement query analysis phase (parse symbols, asset types, intent)
    - Implement parallel agent execution with 30-second timeout per agent
    - Implement partial failure tolerance (proceed with available outputs)
    - Support configurable LLM providers (OpenAI, Anthropic, Google, Qwen, Bedrock)
    - _Requirements: 8.1, 8.2, 8.3, 8.4, 8.5, 8.6_
  
  - [ ]* 21.2 Write property test for Multi_Agent_System
    - **Property 26: Agent Partial Failure Tolerance**
    - **Validates: Requirements 8.5**
  
  - [x] 21.3 Implement Price_Agent
    - Fetch current prices via Data_Source_Router
    - Fetch historical OHLCV data
    - Return structured PriceAgentResponse
    - Handle source failover transparently
    - Format all prices in VND
    - _Requirements: 9.1, 9.2, 9.3, 9.4, 9.5_
  
  - [ ]* 21.4 Write property test for Price_Agent
    - **Property 27: Price Agent VND Formatting**
    - **Validates: Requirements 9.5**
  
  - [x] 21.5 Implement Analysis_Agent
    - Compute all 21 technical indicators with default parameters
    - Identify support/resistance levels (local minima/maxima)
    - Detect volume anomalies (>2x 20-day average)
    - Retrieve ICB sector classification from Sector_Service
    - Compare stock vs sector performance (1w, 1m, 3m, 1y)
    - Compare fundamentals vs sector medians
    - Evaluate sector trend and rotation signals
    - Generate composite signal (strongly bullish to strongly bearish)
    - Return AnalysisAgentResponse with confidence score 0-100
    - _Requirements: 10.1, 10.2, 10.3, 10.4, 10.5, 10.6, 10.7, 10.8, 10.9, 10.10, 10.11, 10.12, 10.13_
  
  - [ ]* 21.6 Write property test for Analysis_Agent
    - **Property 28: Indicator Omission on Insufficient Data**
    - **Validates: Requirements 10.6**
  
  - [x] 21.7 Implement News_Agent
    - Fetch CafeF RSS feed
    - Filter articles by symbol mentions and keywords
    - Fallback to Google search if RSS insufficient
    - Limit to 10 most recent/relevant articles
    - Use LLM to summarize articles
    - Return NewsAgentResponse
    - _Requirements: 11.1, 11.2, 11.3, 11.4, 11.5_

- [ ] 22. Autonomous monitoring and knowledge base
  - [x] 22.1 Implement Pattern_Detector
    - Detect accumulation patterns (price consolidation 5% for 10+ days, volume >1.5x avg, institutional buying)
    - Detect distribution patterns (price near highs, volume on down days, institutional selling)
    - Detect breakout signals (price above resistance, volume >2x avg)
    - Generate confidence scores 0-100
    - _Requirements: 12.3, 12.4, 12.5_
  
  - [ ]* 22.2 Write property tests for Pattern_Detector
    - **Property 29: Accumulation Pattern Detection**
    - **Property 30: Pattern Observation Completeness**
    - **Validates: Requirements 12.3, 12.6**
  
  - [x] 22.3 Implement Monitor_Agent
    - Run on schedule (30min trading hours, 2hr off-hours)
    - Fetch OHLCV data for all watchlist symbols
    - Use Pattern_Detector to identify patterns
    - Generate PatternObservation with supporting data
    - Store observations in Knowledge_Base
    - Trigger Alert_Service for confidence ≥60
    - Implement 5-minute timeout with graceful failure handling
    - _Requirements: 12.1, 12.2, 12.6, 12.7, 12.8, 12.9_
  
  - [ ]* 22.4 Write property test for Monitor_Agent
    - **Property 31: Monitor Scan Timeout Handling**
    - **Validates: Requirements 12.9**
  
  - [x] 22.5 Implement Knowledge_Base
    - Persist observations with all required fields
    - Track outcomes at 1d, 7d, 14d, 30d intervals
    - Implement QueryObservations with filters
    - Compute aggregate accuracy metrics per pattern type
    - Retain observations for 2 years
    - _Requirements: 13.1, 13.2, 13.3, 13.4, 13.6, 13.7_
  
  - [ ]* 22.6 Write property test for Knowledge_Base
    - **Property 32: Knowledge Base Outcome Tracking**
    - **Property 33: Pattern Accuracy Metric Computation**
    - **Validates: Requirements 13.2, 13.7**
  
  - [x] 22.7 Implement Alert_Service
    - Create alerts for confidence ≥60
    - Deliver in-app notifications
    - Include symbol, pattern type, confidence, explanation, timestamp, chart link
    - Persist alerts in database
    - Support user alert preferences (min confidence, pattern types, symbols)
    - Implement deduplication (48-hour window, unless confidence +10)
    - Mark alerts as expired after 24 hours if not viewed
    - _Requirements: 14.1, 14.2, 14.3, 14.4, 14.5, 14.6, 14.7_
  
  - [ ]* 22.8 Write property test for Alert_Service
    - **Property 34: Alert Deduplication**
    - **Validates: Requirements 14.6**

- [ ] 23. Supervisor agent and recommendation audit
  - [x] 23.1 Implement Supervisor_Agent
    - Orchestrate sub-agents based on query intent
    - Incorporate portfolio context from Portfolio_Engine
    - Query Knowledge_Base for relevant patterns
    - Apply NAV-based position sizing logic
    - Check diversification (flag if single asset >40% NAV)
    - Incorporate sector context from Analysis_Agent
    - Format structured SupervisorRecommendation
    - Handle partial failures gracefully
    - _Requirements: 15.1, 15.2, 15.3, 15.4, 15.5, 15.6, 15.7, 15.8, 15.9, 15.10_
  
  - [ ]* 23.2 Write property test for Supervisor_Agent
    - **Property 35: Diversification Warning**
    - **Validates: Requirements 15.8**
  
  - [x] 23.3 Implement Recommendation_Audit_Log
    - Persist every recommendation with full inputs/outputs
    - Track outcomes at 1d, 7d, 14d, 30d intervals
    - Compute recommendation accuracy metrics
    - Feed outcomes back to Knowledge_Base
    - Retain audit records for 2 years
    - _Requirements: 35.1, 35.2, 35.4, 35.5, 35.6_

- [x] 24. Checkpoint - Verify AI agent system
  - Test Price_Agent with various asset types
  - Test Analysis_Agent with all 21 indicators
  - Test News_Agent with real symbols
  - Test Pattern_Detector with historical data
  - Test Monitor_Agent autonomous scanning
  - Test Supervisor_Agent recommendation synthesis
  - Verify Knowledge_Base outcome tracking
  - Ensure all tests pass, ask the user if questions arise.

- [x] 25. Authentication and retry'
- [ ] security
  - [x] 25.1 Implement Auth_Service
    - Implement user registration with bcrypt password hashing (cost 12)
    - Implement login with JWT token issuance (24-hour expiry)
    - Implement password change requiring current password
    - Implement account lockout (5 failed attempts in 15 min → 30 min lockout)
    - Implement session management with 4-hour inactivity timeout
    - _Requirements: 36.1, 36.2, 36.6, 36.7, 36.8_
  
  - [x] 25.2 Implement JWT middleware
    - Extract and verify JWT from Authorization header
    - Check token expiration
    - Attach user_id to request context
    - Protect all endpoints except /api/health and /api/auth/login
    - _Requirements: 36.3, 36.4_
  
  - [x] 25.3 Implement security measures
    - Enforce HTTPS in production
    - Configure CORS (frontend origin only, credentials: true)
    - Implement per-user rate limiting (100 req/min)
    - Implement per-IP rate limiting (200 req/min for unauth endpoints)
    - Validate all inputs and use parameterized queries
    - _Requirements: 36.9_

- [x] 26. REST API endpoints - Backend routes
  - [x] 26.1 Implement authentication endpoints
    - POST /api/auth/login
    - POST /api/auth/logout
    - POST /api/auth/change-password
    - GET /api/auth/me
  
  - [x] 26.2 Implement price service endpoints
    - GET /api/prices/quotes
    - GET /api/prices/history
    - GET /api/prices/gold
    - GET /api/prices/crypto
    - GET /api/prices/fx
  
  - [x] 26.3 Implement portfolio endpoints
    - GET /api/portfolio/summary
    - POST /api/portfolio/assets
    - PUT /api/portfolio/assets/:id
    - DELETE /api/portfolio/assets/:id
    - GET /api/portfolio/transactions
    - POST /api/portfolio/transactions
    - GET /api/portfolio/performance
    - GET /api/portfolio/risk
  
  - [x] 26.4 Implement sector service endpoints
    - GET /api/sectors/performance
    - GET /api/sectors/:sector/performance
    - GET /api/sectors/symbol/:symbol
    - GET /api/sectors/:sector/averages
    - GET /api/sectors/:sector/stocks
  
  - [x] 26.5 Implement screener endpoints
    - POST /api/screener
    - GET /api/screener/presets
    - POST /api/screener/presets
    - DELETE /api/screener/presets/:id
  
  - [x] 26.6 Implement comparison endpoints
    - GET /api/comparison/valuation
    - GET /api/comparison/performance
    - GET /api/comparison/correlation
  
  - [x] 26.7 Implement watchlist endpoints
    - GET /api/watchlists
    - POST /api/watchlists
    - PUT /api/watchlists/:id
    - DELETE /api/watchlists/:id
    - POST /api/watchlists/:id/symbols
    - DELETE /api/watchlists/:id/symbols/:symbol
    - PUT /api/watchlists/:id/reorder
  
  - [x] 26.8 Implement AI chat endpoints
    - POST /api/chat (integrate full multi-agent pipeline)
    - POST /api/models
    - _Requirements: 19.1, 19.2, 19.3_
  
  - [x] 26.9 Implement alert endpoints
    - GET /api/alerts
    - PUT /api/alerts/:id/viewed
    - PUT /api/alerts/preferences
  
  - [x] 26.10 Implement knowledge base endpoints
    - GET /api/knowledge/observations
    - GET /api/knowledge/accuracy/:patternType
  
  - [x] 26.11 Implement market data endpoints
    - GET /api/market/listing
    - GET /api/market/company/:symbol
    - GET /api/market/finance/:symbol
    - GET /api/market/statistics
    - GET /api/market/commodities
    - GET /api/market/macro
  
  - [x] 26.12 Implement goal endpoints
    - GET /api/goals
    - POST /api/goals
    - PUT /api/goals/:id
    - DELETE /api/goals/:id
    - GET /api/goals/:id/progress
  
  - [x] 26.13 Implement backtest endpoint
    - POST /api/backtest
  
  - [x] 26.14 Implement export endpoints
    - GET /api/export/transactions
    - GET /api/export/snapshot
    - GET /api/export/report
    - GET /api/export/tax
  
  - [x] 26.15 Implement health and metrics endpoints
    - GET /api/health
    - GET /api/metrics/rate-limits

- [x] 27. Checkpoint - Verify backend API
  - Test all authentication flows
  - Test all CRUD operations
  - Test AI chat with multi-agent pipeline
  - Test rate limiting
  - Test error handling
  - Verify JWT authentication on protected endpoints
  - Ensure all tests pass, ask the user if questions arise.

- [x] 28. Frontend infrastructure setup
  - [x] 28.1 Set up Next.js 16 project structure
    - Configure App Router
    - Set up TypeScript configuration
    - Configure Tailwind CSS
    - Install dependencies (lightweight-charts, lucide-react, framer-motion)
    - Set up environment variables for API base URL
  
  - [x] 28.2 Implement Theme_Service
    - Create ThemeContext with light/dark modes
    - Implement theme toggle button
    - Persist theme preference in localStorage
    - Apply theme to all UI components
    - Ensure WCAG AA contrast ratios (4.5:1 normal text, 3:1 large text)
    - _Requirements: 37.1, 37.2, 37.3, 37.4, 37.5, 37.6, 37.8_
  
  - [x] 28.3 Implement I18n_Service
    - Create I18nContext with vi-VN and en-US locales
    - Implement language selector dropdown
    - Persist language preference in localStorage
    - Translate all static UI text
    - Implement locale-aware number formatting (VN: period thousands, comma decimal; EN: comma thousands, period decimal)
    - Implement locale-aware date formatting (VN: dd/MM/yyyy; EN: MM/dd/yyyy)
    - Implement locale-aware time formatting (VN: 24-hour; EN: 12-hour AM/PM)
    - Implement locale-aware currency formatting
    - Support dynamic text interpolation
    - _Requirements: 38.1, 38.2, 38.3, 38.4, 38.5, 38.6, 38.7, 38.8, 38.9, 38.10, 38.13_
  
  - [x] 28.4 Implement authentication context
    - Create AuthContext with login/logout functions
    - Store JWT in httpOnly cookie
    - Implement protected route wrapper
    - Redirect to login when no valid session
    - _Requirements: 36.4, 36.5_

- [x] 29. Chart engine with 21 technical indicators
  - [x] 29.1 Implement Chart_Engine core
    - Integrate lightweight-charts library
    - Create ChartEngine class with candlestick series
    - Implement volume histogram
    - Support time intervals (1m, 5m, 15m, 1h, 1d, 1w, 1M)
    - Fetch OHLCV data from backend
    - _Requirements: 7.1, 7.2, 7.12_
  
  - [x] 29.2 Implement trend indicators
    - SMA (Simple Moving Average)
    - EMA (Exponential Moving Average)
    - VWAP (Volume Weighted Average Price)
    - VWMA (Volume Weighted Moving Average)
    - ADX (Average Directional Movement Index)
    - Aroon Indicator
    - Parabolic SAR (Stop and Reverse)
    - Supertrend
    - _Requirements: 7.3_
  
  - [x] 29.3 Implement momentum indicators
    - RSI (Relative Strength Index)
    - MACD (Moving Average Convergence Divergence)
    - Williams %R
    - CMO (Chande Momentum Oscillator)
    - Stochastic Oscillator
    - ROC (Rate of Change)
    - Momentum
    - _Requirements: 7.4_
  
  - [x] 29.4 Implement volatility indicators
    - Bollinger Bands
    - Keltner Channel
    - ATR (Average True Range)
    - Standard Deviation
    - _Requirements: 7.5_
  
  - [x] 29.5 Implement volume and statistics indicators
    - OBV (On-Balance Volume)
    - Linear Regression
    - _Requirements: 7.6, 7.7_
  
  - [x] 29.6 Implement indicator configuration
    - Allow user to configure indicator parameters
    - Support multiple instances of same indicator with different params
    - Overlay indicators on price pane, oscillators in separate pane
    - _Requirements: 7.8, 7.9, 7.14_
  
  - [x] 29.7 Implement drawing tools
    - Trend lines
    - Horizontal lines
    - Fibonacci retracement
    - Rectangle selection
    - Persist drawings in localStorage
    - _Requirements: 7.10, 7.11_
  
  - [x] 29.8 Implement chart theme adaptation
    - Apply theme-specific colors to charts
    - Re-render charts on theme change without losing state
    - _Requirements: 37.7, 37.9_
  
  - [ ]* 29.9 Write property tests for Chart_Engine
    - **Property 24: Chart Drawing Persistence**
    - **Property 25: Indicator Recalculation on Interval Change**
    - **Validates: Requirements 7.11, 7.13**

- [x] 30. Checkpoint - Verify chart engine
  - Test chart rendering with real OHLCV data
  - Test all 21 indicators
  - Test drawing tools and persistence
  - Test theme adaptation
  - Test interval switching
  - Ensure all tests pass, ask the user if questions arise.


- [x] 31. Dashboard and portfolio views
  - [x] 31.1 Implement Dashboard component
    - Display total NAV with real portfolio data (replace hardcoded ASSET_DATA)
    - Display asset allocation pie chart by type
    - Display NAV change (24-hour absolute and percentage)
    - Display 5 most recent transactions
    - Display quick metric cards (gold rate, Bitcoin price with 24h change)
    - Display notification panel with unread alerts
    - Navigate to chart view when alert clicked
    - _Requirements: 17.1, 17.2, 17.3, 17.4, 17.5, 17.6, 17.7_
  
  - [x] 31.2 Implement real-time price updates
    - Poll Price_Service every 15s for stocks (trading hours), 60s for crypto, 300s for gold
    - Update NAV, allocation, holdings without page reload
    - Reduce stock polling to 300s outside trading hours (before 9:00 or after 15:00 ICT)
    - Display freshness indicator (green <1min, yellow 1-5min, red >5min)
    - Retain last known price on failure with stale warning
    - _Requirements: 16.1, 16.2, 16.3, 16.4, 16.5_
  
  - [ ]* 31.3 Write property test for price freshness
    - **Property 36: Price Freshness Indicator**
    - **Validates: Requirements 16.4**
  
  - [x] 31.4 Implement Portfolio view
    - Display all holdings with current prices, market values, unrealized P&L
    - Display transaction history with filters
    - Display performance metrics (TWR, MWRR, equity curve chart)
    - Display risk metrics panel
    - Support time period selection (1W, 1M, 3M, 6M, 1Y, YTD, ALL)
    - Provide export buttons (CSV, PDF)
    - _Requirements: 26.7, 27.6, 28.7_
  
  - [x] 31.5 Implement multi-currency display toggle
    - Add VND/USD toggle control
    - Convert all values using FX_Service rate
    - Display current USD/VND rate in header
    - _Requirements: 29.3, 29.4, 29.6_

- [x] 32. Screener and comparison views
  - [x] 32.1 Implement Screener view
    - Display filter panel with all criteria (fundamentals, sectors, exchanges, sector trends)
    - Display results table with sortable columns
    - Support pagination
    - Allow saving/loading filter presets
    - Provide built-in default presets (Value Investing, High Growth, High Dividend, Low Debt)
    - _Requirements: 18.11, 18.12, 18.13, 18.14_
  
  - [x] 32.2 Implement Stock Comparison view
    - Display stock selector (add/remove symbols)
    - Display sector/industry group dropdown for auto-population
    - Implement three tabs: Định giá (Valuation), Hiệu suất (Performance), Tương quan (Correlation)
    - Display interactive charts for Valuation (P/E, P/B time-series)
    - Display interactive charts for Performance (normalized % returns)
    - Display correlation matrix heatmap
    - Support time period selection (3M, 6M, 1Y, 3Y, 5Y)
    - Allow hide/show individual stock lines
    - Provide "Clear All" button
    - Display prompt when fewer than 2 stocks selected
    - _Requirements: 24.7, 24.8, 24.9, 24.10, 24.11, 24.12, 24.13, 24.14, 24.17_

- [x] 33. Sector trend dashboard
  - [x] 33.1 Implement Sector Trend Dashboard
    - Display heatmap with all 10 ICB sectors (color-coded performance)
    - Support time period switching (today, 1w, 1m, 3m, 6m, 1y)
    - Display bar chart comparing sector performance
    - Display sector trend indicators (uptrend/downtrend/sideways arrows)
    - Display top 3 and bottom 3 performing sectors
    - Show detail panel on sector click (index chart, top 5 stocks, median fundamentals)
    - Auto-refresh every 5 minutes (trading hours), 30 minutes (off-hours)
    - _Requirements: 22.1, 22.2, 22.3, 22.4, 22.5, 22.6, 22.7, 22.10_

- [x] 34. AI chat widget
  - [x] 34.1 Implement Chat Widget
    - Integrate with POST /api/chat endpoint (multi-agent pipeline)
    - Detect asset symbols and types in user messages
    - Render structured recommendations with sections (data, analysis, news, advice)
    - Support conversation history (last 10 messages)
    - Handle 45-second timeout with partial response
    - _Requirements: 19.1, 19.2, 19.3, 19.4, 19.5, 19.6_
  
  - [x] 34.2 Implement recommendation history view
    - Display past recommendations with outcomes
    - Show price changes at 1d, 7d, 14d, 30d intervals
    - Support filters by date range, symbol, action type
    - _Requirements: 35.3, 35.7_

- [x] 35. Additional frontend features
  - [x] 35.1 Implement Watchlist management UI
    - Display all watchlists with symbols
    - Support create/rename/delete watchlist
    - Support add/remove/reorder symbols
    - Support setting price alert thresholds per symbol
    - _Requirements: 25.1, 25.2, 25.3_
  
  - [x] 35.2 Implement Corporate Actions calendar
    - Display upcoming ex-dividend dates, payment dates, AGM dates
    - Show 3-day advance notifications for ex-dividend dates
    - Display dividend history per holding
    - _Requirements: 30.4, 30.5_
  
  - [x] 35.3 Implement Goal Planning UI
    - Display goal progress with progress bars
    - Show projected completion dates
    - Display required monthly contributions
    - Support create/edit/delete goals
    - _Requirements: 31.4, 31.5_
  
  - [x] 35.4 Implement Backtesting UI
    - Provide strategy rule builder (select indicators, set conditions/thresholds)
    - Support symbol and timeframe selection
    - Display backtest results (equity curve, trade markers, metrics)
    - _Requirements: 32.8_
  
  - [x] 35.5 Implement Offline mode support
    - Detect online/offline status
    - Display "Offline Mode" banner when offline
    - Serve cached data with timestamps
    - Auto-refresh on connectivity restore
    - Display per-source health indicators
    - _Requirements: 34.1, 34.2, 34.3, 34.4, 34.6, 34.7_

- [x] 36. Checkpoint - Verify frontend features
  - Test dashboard with real data
  - Test portfolio view with transactions
  - Test screener with various filters
  - Test stock comparison across all modes
  - Test sector trend dashboard
  - Test AI chat widget
  - Test watchlist management
  - Test goal planning
  - Test backtesting UI
  - Test offline mode
  - Ensure all tests pass, ask the user if questions arise.

- [ ] 37. Integration and end-to-end testing
  - [ ]* 37.1 Write integration tests for critical flows
    - User registration and login flow
    - Portfolio creation and transaction recording
    - Price fetching and NAV calculation
    - AI chat query with multi-agent pipeline
    - Pattern detection and alert delivery
    - Screener with real data
    - Stock comparison
    - Export generation
  
  - [ ]* 37.2 Write E2E tests for user journeys
    - Complete portfolio setup journey
    - Complete AI advisory journey
    - Complete screener to watchlist journey
    - Complete goal planning journey
    - Complete backtesting journey
  
  - [ ]* 37.3 Write remaining property tests
    - **Property 38: Database Transaction Atomicity**
    - **Property 49: Rate Limit Enforcement**
    - **Property 50: JWT Token Expiration**
    - **Property 51: Account Lockout After Failed Attempts**
    - **Property 52: Theme Persistence**
    - **Property 53: Language Persistence**
    - **Property 54: Locale Number Formatting**
    - **Property 55: Locale Date Formatting**
    - **Property 56: Locale Time Formatting**
    - **Property 57: Locale Currency Formatting**
    - **Property 58: Chart Theme Adaptation**
    - **Property 59: Multi-Currency Conversion Consistency**
    - **Property 60: Corporate Action Cost Basis Adjustment**
    - **Property 61: Goal Progress Calculation**
    - **Validates: Requirements 20.5, 33.1, 36.2, 36.7, 37.3, 37.4, 38.3, 38.4, 38.7, 38.8, 38.9, 38.10, 37.9, 29.4, 30.3, 31.2**

- [x] 38. Performance optimization and polish
  - [x] 38.1 Optimize backend performance
    - Implement database connection pooling
    - Add database indexes for frequently queried fields
    - Optimize N+1 query patterns
    - Implement request batching where possible
    - Profile and optimize slow endpoints
  
  - [x] 38.2 Optimize frontend performance
    - Implement code splitting for route-based chunks
    - Lazy load heavy components (charts, comparison)
    - Optimize re-renders with React.memo
    - Implement virtual scrolling for large lists
    - Optimize bundle size
  
  - [x] 38.3 Implement error boundaries
    - Add React error boundaries for graceful failure
    - Implement global error handler in backend
    - Log errors to monitoring service
  
  - [x] 38.4 Add loading states and skeletons
    - Implement loading skeletons for all async data
    - Add progress indicators for long operations
    - Implement optimistic UI updates where appropriate
  
  - [x] 38.5 Accessibility improvements
    - Ensure keyboard navigation works throughout
    - Add ARIA labels to interactive elements
    - Verify screen reader compatibility
    - Test with accessibility tools

- [x] 39. Documentation and deployment preparation
  - [ ] 39.1 Write API documentation
    - Document all REST endpoints with request/response examples
    - Document authentication flow
    - Document rate limits and quotas
    - Document error codes and messages
  
  - [ ] 39.2 Write deployment guide
    - Document environment variables
    - Document database setup and migrations
    - Document production deployment steps
    - Document monitoring and logging setup
  
  - [ ] 39.3 Write user guide
    - Document key features and workflows
    - Create tutorial for first-time users
    - Document AI chat usage and best practices
    - Document screener and comparison tools
  
  - [ ] 39.4 Set up production environment
    - Configure PostgreSQL database
    - Set up HTTPS with SSL certificates
    - Configure environment variables
    - Set up logging and monitoring
    - Configure backup strategy

- [x] 40. Final checkpoint and launch preparation
  - Run full test suite (unit, property, integration, E2E)
  - Verify all 38 requirements are implemented
  - Verify all 61 correctness properties are tested
  - Perform security audit
  - Perform performance testing under load
  - Test with real market data during trading hours
  - Verify offline mode works correctly
  - Test multi-language support thoroughly
  - Test dark theme across all views
  - Prepare launch checklist
  - Ensure all tests pass, ask the user if questions arise.

## Notes

- Tasks marked with `*` are optional and can be skipped for faster MVP delivery
- Each task references specific requirements for traceability
- Property tests validate universal correctness properties from the design document
- Checkpoints ensure incremental validation at logical milestones
- The implementation follows a bottom-up approach: data layer → business logic → AI agents → API → frontend
- Backend uses Go with Gin framework, vnstock-go library, and langchaingo for AI
- Frontend uses Next.js 16 with TypeScript, Tailwind CSS, and lightweight-charts
- Database is SQLite for development, PostgreSQL for production
- All monetary values are stored in VND as the base currency
- VCI is the primary data source with KBS as fallback
- The platform supports 21 technical indicators across 5 categories
- The multi-agent AI system uses 5 specialized agents with autonomous monitoring
- The Knowledge_Base learns from pattern outcomes to improve recommendations over time


- [ ] 41. Mobile responsiveness and PWA
  - [ ] 41.1 Implement responsive design
    - Add responsive breakpoints (mobile <768px, tablet 768-1024px, desktop >1024px)
    - Use mobile-first CSS approach with progressive enhancement
    - Implement mobile-optimized navigation (hamburger menu or bottom nav bar)
    - Optimize touch target sizes to minimum 44x44 pixels
    - Use viewport-relative units and flexible layouts
    - Implement swipe gestures for tab navigation and modal dismissal
    - _Requirements: 39.1, 39.2, 39.4, 39.8, 39.9, 39.11_
  
  - [ ] 41.2 Implement PWA capabilities
    - Register service worker for offline support
    - Create web app manifest file for "Add to Home Screen"
    - Cache critical assets (HTML, CSS, JS, fonts) for offline access
    - Implement lazy loading for images and heavy components
    - _Requirements: 39.5, 39.6, 39.7, 39.12_
  
  - [ ] 41.3 Implement touch-optimized chart controls
    - Support pinch-to-zoom, two-finger pan, tap for crosshair
    - Support long-press for context menu
    - Adapt chart controls with larger touch-friendly buttons
    - Simplify indicator configuration panels for mobile
    - _Requirements: 39.3, 39.10_

- [ ] 42. Real-time WebSocket price streaming
  - [ ] 42.1 Implement WebSocket server
    - Create WebSocket endpoint at /ws/prices
    - Implement JWT authentication in connection handshake
    - Implement subscription management (subscribe/unsubscribe symbols)
    - Implement heartbeat/ping-pong for connection health
    - Limit subscriptions to max 100 symbols per connection
    - Broadcast market status changes (open, close, halt)
    - _Requirements: 40.1, 40.2, 40.5, 40.7, 40.10, 40.11_
  
  - [ ] 42.2 Implement WebSocket client
    - Establish WebSocket connection on dashboard load
    - Subscribe to portfolio and watchlist symbols
    - Update UI immediately on price updates
    - Implement auto-reconnect with exponential backoff (1s, 2s, 4s, 8s, max 30s)
    - Fall back to HTTP polling after 5 failed reconnection attempts
    - _Requirements: 40.3, 40.6, 40.8, 40.9_
  
  - [ ] 42.3 Integrate WebSocket with Price_Service
    - Publish price updates to WebSocket clients when fetching from upstream
    - _Requirements: 40.4_

  - [ ]* 42.4 Write property tests for WebSocket
    - **Property 62: WebSocket Reconnection Backoff**
    - **Property 63: WebSocket Subscription Limit**
    - **Validates: Requirements 40.8, 40.10**

- [ ] 43. Checkpoint - Verify mobile and WebSocket
  - Test responsive layout on mobile, tablet, and desktop viewports
  - Test PWA install and offline caching
  - Test touch gestures on chart engine
  - Test WebSocket connection, subscription, and price streaming
  - Test WebSocket reconnection and HTTP polling fallback
  - Ensure all tests pass, ask the user if questions arise.

- [ ] 44. Push notifications
  - [ ] 44.1 Implement push notification backend
    - Implement Web Push protocol for sending notifications
    - Store push subscriptions in database
    - Send push for high-priority alerts (price alerts, patterns >80 confidence, ex-dividend <24h)
    - Implement quiet hours (default 22:00-07:00)
    - Implement notification grouping (max 5 per hour, summary for overflow)
    - Clean up expired subscriptions (7 consecutive failed deliveries)
    - _Requirements: 41.3, 41.4, 41.8, 41.9, 41.10_
  
  - [ ] 44.2 Implement push notification frontend
    - Request push notification permission on first login
    - Register with Web Push API and store subscription in backend
    - Display notifications with title, body, icon, action buttons (View, Dismiss)
    - Navigate to relevant view on notification click
    - Allow per-alert-type notification preferences (push, in-app, disabled)
    - _Requirements: 41.1, 41.2, 41.5, 41.6, 41.7_

- [ ] 45. Email and SMS alerts
  - [ ] 45.1 Implement email alert delivery
    - Integrate SMTP service or email API (SendGrid, AWS SES)
    - Send email for critical and high-priority alerts
    - Format emails with HTML including summary, data, and app link
    - Implement rate limiting (max 20 emails per day per user)
    - Provide unsubscribe link in all emails
    - _Requirements: 42.3, 42.7, 42.8, 42.10_
  
  - [ ] 45.2 Implement SMS alert delivery
    - Integrate SMS gateway API (Twilio, AWS SNS)
    - Send SMS only for critical-priority alerts
    - Implement rate limiting (max 5 SMS per day per user)
    - _Requirements: 42.4, 42.6, 42.8_
  
  - [ ] 45.3 Implement alert priority and channel configuration
    - Define priority levels (critical, high, medium, low)
    - Allow per-alert-type channel configuration (in-app, email, SMS)
    - Support separate quiet hours per channel
    - _Requirements: 42.1, 42.2, 42.5, 42.7, 42.9_

- [ ] 46. Social features and community
  - [ ] 46.1 Implement community performance sharing
    - Allow opt-in to anonymous performance sharing
    - Publish anonymized metrics (return %, Sharpe, max drawdown, allocation %)
    - Display community leaderboard (top performers by return, Sharpe, risk-adjusted)
    - Support time period selection (1M, 3M, 6M, 1Y)
    - _Requirements: 43.1, 43.2, 43.3_
  
  - [ ] 46.2 Implement public watchlists
    - Allow publishing watchlists as public with shareable link
    - Support following public watchlists from other users
    - Track follower count and display popularity metrics
    - Display followed watchlists in "Community Watchlists" section
    - _Requirements: 43.4, 43.5, 43.6_

  - [ ] 46.3 Implement community sentiment tracking
    - Aggregate sentiment from public watchlists (symbol add count in past 7 days)
    - Display sentiment indicators on stock detail pages (discussion volume, sentiment %, trending)
    - Display "Community Trending" section on dashboard (top 10 most-discussed stocks)
    - _Requirements: 43.7, 43.8_
  
  - [ ] 46.4 Implement peer comparison
    - Compare user performance against anonymized peer benchmarks by portfolio size
    - Display peer comparison charts showing performance percentile
    - _Requirements: 43.9, 43.10_
  
  - [ ] 46.5 Implement privacy controls and moderation
    - Allow hiding profile from leaderboards
    - Implement moderation for public watchlists and comments
    - _Requirements: 43.11, 43.12_

- [ ] 47. Transaction import and broker integration
  - [ ] 47.1 Implement CSV transaction import
    - Support flexible column mapping interface
    - Provide preset templates for major VN brokers (SSI, VPS, HSC, VCBS, Vietcombank, MB)
    - Auto-detect broker format when possible
    - Display preview of parsed transactions before import
    - Validate transactions (required fields, data types, logical consistency)
    - Support transaction types (buy, sell, dividend, split, bonus, rights, deposit, withdrawal)
    - _Requirements: 44.1, 44.2, 44.3, 44.4, 44.5_
  
  - [ ] 47.2 Implement transaction reconciliation
    - Display reconciliation view with imported vs existing transactions
    - Detect conflicts and provide resolution options (skip, overwrite, merge)
    - Implement duplicate detection (symbol, date, quantity, price, type matching)
    - _Requirements: 44.6, 44.12_

  - [ ] 47.3 Implement broker API integration
    - Integrate SSI iBoard API for automated transaction sync
    - Integrate VPS API if accessible
    - Sync transactions daily at configurable time (default 16:00 ICT)
    - Store last sync timestamp and display sync status
    - _Requirements: 44.7, 44.8, 44.9_
  
  - [ ] 47.4 Implement bank statement parsing
    - Parse cash flow transactions from VN bank statements (Vietcombank, Techcombank, VPBank)
    - Detect deposits and withdrawals
    - _Requirements: 44.10_
  
  - [ ] 47.5 Implement import history log
    - Display all import operations with timestamp, source, transaction count, errors
    - _Requirements: 44.11_

- [ ] 48. Portfolio rebalancing tools
  - [ ] 48.1 Implement target allocation management
    - Allow defining target allocation % for asset types and holdings/sectors
    - Display target vs actual allocation comparison chart
    - _Requirements: 45.1, 45.2_
  
  - [ ] 48.2 Implement rebalancing suggestions
    - Compute rebalancing when deviation exceeds threshold (default 5%)
    - Generate buy/sell recommendations with quantities and costs
    - Provide rebalancing simulator to preview portfolio state
    - Factor in transaction costs (broker fees, taxes)
    - _Requirements: 45.3, 45.4, 45.5, 45.7_
  
  - [ ] 48.3 Implement rebalancing strategies
    - Support threshold-based rebalancing
    - Support calendar-based rebalancing (monthly/quarterly)
    - Support opportunistic rebalancing
    - _Requirements: 45.6_

  - [ ] 48.4 Implement tax-loss harvesting
    - Identify holdings with unrealized losses for tax offset
    - Suggest selling losing positions and replacing with similar assets
    - Display estimated tax savings
    - _Requirements: 45.8, 45.9, 45.10_
  
  - [ ] 48.5 Implement rebalancing history and constraints
    - Track rebalancing history with outcomes and performance impact
    - Support constraints (min trade size, max trades, excluded holdings)
    - _Requirements: 45.11, 45.12_

- [ ] 49. Voice input and natural language queries
  - [ ] 49.1 Implement voice input
    - Integrate Web Speech API with Vietnamese support (vi-VN)
    - Add microphone button in chat widget
    - Display visual indicator during recording (animated waveform)
    - Support push-to-talk and toggle mode
    - _Requirements: 46.1, 46.2, 46.3, 46.4_
  
  - [ ] 49.2 Implement natural language query parsing
    - Parse portfolio queries (profit, best performer, dividend income)
    - Extract intent and entities (time periods, asset types, metrics, actions)
    - Query Portfolio_Engine and Transaction_Ledger for answers
    - Format responses in natural language with conversational tone
    - _Requirements: 46.5, 46.6, 46.7, 46.8_
  
  - [ ] 49.3 Implement voice output and commands
    - Support text-to-speech for AI responses
    - Implement voice command shortcuts (show portfolio, check price, set alert)
    - Handle ambiguous queries with clarifying questions
    - _Requirements: 46.9, 46.10, 46.11_
  
  - [ ] 49.4 Implement voice interaction logging
    - Log voice interactions for quality improvement
    - _Requirements: 46.12_

- [ ] 50. Advanced tax optimization and reporting
  - [ ] 50.1 Implement VN tax computation
    - Compute VN personal income tax (0.1% on sell value for stocks)
    - Compute capital gains tax for other assets
    - Track cost basis adjustments for corporate actions
    - _Requirements: 47.1, 47.8_
  
  - [ ] 50.2 Implement tax report generation
    - Generate annual tax report (total sell value, tax paid, realized gains/losses, dividend income)
    - Generate detailed transaction ledger for tax filing
    - Generate capital gains summary by asset type
    - Generate dividend income summary
    - Export to PDF with Vietnamese formatting
    - _Requirements: 47.2, 47.7, 47.11_
  
  - [ ] 50.3 Implement tax optimization suggestions
    - Identify tax-loss harvesting opportunities
    - Compute optimal timing for selling winning positions
    - Suggest tax-efficient withdrawal strategies
    - Compute wash sale violations (sell at loss, repurchase within 30 days)
    - _Requirements: 47.3, 47.4, 47.6, 47.9_
  
  - [ ] 50.4 Implement tax dashboard and scenario modeling
    - Display YTD tax paid, projected year-end liability, harvesting opportunities
    - Provide tax scenario modeling for planned trades
    - _Requirements: 47.5, 47.10_
  
  - [ ] 50.5 Implement tax lot tracking
    - Support FIFO, LIFO, or specific identification methods
    - _Requirements: 47.12_

- [ ] 51. Educational content and onboarding
  - [ ] 51.1 Implement interactive onboarding
    - Create onboarding flow (account setup, add first asset, dashboard tour, AI chat intro)
    - Implement step-by-step tutorial system with tooltips and guided tours
    - Track user progress and display completion badges
    - _Requirements: 48.1, 48.2, 48.10_
  
  - [ ] 51.2 Implement financial glossary and help
    - Create searchable glossary (Vietnamese and English)
    - Display contextual help tooltips on complex UI elements
    - _Requirements: 48.3, 48.4_
  
  - [ ] 51.3 Implement investment strategy templates
    - Provide templates (value, growth, dividend, index, sector rotation)
    - Include predefined screener filters and allocation targets
    - _Requirements: 48.5_
  
  - [ ] 51.4 Implement educational content
    - Provide video tutorials or animated guides
    - Create "Learn" section with articles (VN market basics, technical/fundamental analysis, portfolio/risk management)
    - Provide example portfolios (demo data) for exploration
    - Provide "Tips & Tricks" section
    - Allow replaying tutorials from help menu
    - _Requirements: 48.6, 48.7, 48.8, 48.11, 48.12_
  
  - [ ] 51.5 Implement progressive disclosure
    - Show basic features first, gradually introduce advanced features
    - Unlock features as milestones are reached
    - _Requirements: 48.9_

- [ ] 52. Data quality and validation
  - [ ] 52.1 Implement data anomaly detection
    - Flag sudden price jumps >20% without volume increase
    - Flag prices outside ceiling/floor range
    - Flag zero or negative prices
    - Log anomalies and attempt alternative source fetch
    - _Requirements: 49.1, 49.2_
  
  - [ ] 52.2 Implement OHLCV data validation
    - Validate logical consistency (high >= low, high >= open/close, etc.)
    - Detect and flag gaps in time series data
    - _Requirements: 49.4, 49.5_
  
  - [ ] 52.3 Implement data source reliability tracking
    - Track uptime %, average response time, error rate, data completeness per source
    - Display data quality indicators on dashboard (reliability score, last update, warnings)
    - Display data quality badge on price displays (verified/unverified/anomaly)
    - _Requirements: 49.3, 49.6, 49.11_
  
  - [ ] 52.4 Implement user-reported data issues
    - Add "Report Data Issue" button on stock detail pages
    - Track issues in database (symbol, issue type, reporter, timestamp, resolution)
    - _Requirements: 49.7, 49.8_
  
  - [ ] 52.5 Implement data quality reporting
    - Verify historical data accuracy against cached values
    - Generate daily data quality reports for admins
    - Implement cross-source validation for critical data (flag >5% discrepancies)
    - _Requirements: 49.9, 49.10, 49.12_

- [ ] 53. Performance optimization and caching
  - [ ] 53.1 Implement frontend performance optimizations
    - Implement code splitting for route-based chunks
    - Implement service worker with cache-first strategy
    - Implement optimistic UI updates
    - Implement virtual scrolling for long lists
    - Implement request deduplication
    - Implement skeleton screens and progressive loading
    - _Requirements: 50.1, 50.2, 50.3, 50.7, 50.12, 50.14_
  
  - [ ] 53.2 Implement backend caching and optimization
    - Implement Redis or in-memory caching with appropriate TTLs
    - Optimize database queries with proper indexing
    - Implement database connection pooling
    - Implement database read replicas for read-heavy operations
    - Implement API response compression (gzip or brotli)
    - _Requirements: 50.4, 50.5, 50.6, 50.9, 50.11_
  
  - [ ] 53.3 Implement chart and image optimizations
    - Use canvas-based rendering for charts
    - Implement image lazy loading, responsive images with srcset, WebP format
    - _Requirements: 50.8, 50.10_
  
  - [ ] 53.4 Implement CDN for static assets
    - Use CDN with edge caching for static assets
    - _Requirements: 50.13_

- [ ] 54. Compliance and audit trail
  - [ ] 54.1 Implement audit logging
    - Log all user actions (login/logout, portfolio changes, transactions, settings, exports)
    - Record user ID, action type, timestamp, IP, user agent, affected resources, old/new values, result
    - Implement immutable audit logs
    - Retain logs for configurable period (default 7 years)
    - _Requirements: 51.1, 51.2, 51.3, 51.4_
  
  - [ ] 54.2 Implement audit log query interface
    - Provide query interface for admins (filter by user, date, action type, resource)
    - _Requirements: 51.5_
  
  - [ ] 54.3 Implement data retention and privacy
    - Support data deletion after account closure (preserve audit logs)
    - Support GDPR-style data export (JSON format)
    - Implement data anonymization for deleted accounts
    - _Requirements: 51.6, 51.7, 51.8_
  
  - [ ] 54.4 Implement API and security logging
    - Log all API calls (endpoint, method, params, status, response time, user ID)
    - Log security events (failed logins, password changes, session expirations, suspicious activity)
    - _Requirements: 51.9, 51.10_
  
  - [ ] 54.5 Implement compliance reporting and backup
    - Generate compliance reports (total users, active users, storage, API usage, security events)
    - Implement automated backup of audit logs with encryption at rest
    - _Requirements: 51.11, 51.12_

- [ ] 55. Advanced charting features
  - [ ] 55.1 Implement chart templates
    - Allow saving chart templates (indicators, drawings, interval, style)
    - Allow naming, saving, and loading templates
    - Persist templates in backend database
    - _Requirements: 52.1, 52.2_
  
  - [ ] 55.2 Implement multi-timeframe indicators
    - Display indicator values from higher timeframes on lower timeframe charts
    - _Requirements: 52.3_
  
  - [ ] 55.3 Implement chart replay mode
    - Allow stepping through historical price action bar-by-bar
    - Hide future price data in replay mode
    - Allow placing simulated trades for practice
    - _Requirements: 52.4, 52.5_
  
  - [ ] 55.4 Implement automated pattern recognition
    - Detect chart patterns (head and shoulders, double top/bottom, triangles, flag, pennant)
    - Draw patterns on chart with labels
    - Provide pattern analysis (bullish/bearish, target price, stop loss)
    - Allow setting pattern recognition alerts
    - _Requirements: 52.6, 52.7, 52.8_
  
  - [ ] 55.5 Implement advanced chart features
    - Implement volume profile analysis (horizontal histogram)
    - Support custom time ranges (arbitrary start/end dates)
    - Implement comparison mode (multiple symbols, normalized % scale)
    - Implement chart annotations (text notes, arrows, shapes)
    - _Requirements: 52.9, 52.10, 52.11, 52.12_

  - [ ] 55.6 Implement chart state sync and shortcuts
    - Sync chart state across devices (indicators, drawings, annotations, zoom)
    - Implement keyboard shortcuts (I: add indicator, T: trend line, C: crosshair, +/-: zoom, R: reset)
    - Display keyboard shortcuts help modal (? key)
    - _Requirements: 52.13, 52.14_

- [ ] 56. Portfolio stress testing and scenario analysis
  - [ ] 56.1 Implement historical stress testing
    - Replay portfolio during past crises (2008 GFC, 2020 COVID, 2011 VN banking crisis)
    - Compute drawdown, recovery time, comparison vs VN-Index
    - _Requirements: 53.1, 53.2_
  
  - [ ] 56.2 Implement custom scenario analysis
    - Allow defining hypothetical conditions (VN-Index change, sector shocks, currency, interest rates)
    - Compute portfolio impact (NAV change, affected holdings, risk contribution)
    - _Requirements: 53.3, 53.4_
  
  - [ ] 56.3 Implement Monte Carlo simulation
    - Run 10,000 simulations with randomized returns based on historical volatility
    - Display probability distribution, confidence intervals (10th, 50th, 90th percentile)
    - Show probability of reaching financial goals
    - _Requirements: 53.6, 53.7_
  
  - [ ] 56.4 Implement risk analysis and recommendations
    - Compute correlation breakdown scenarios
    - Identify concentration risks (single holding >20%, sector >40%, asset type >60%)
    - Provide "What-If" calculator for adding/removing positions
    - Compute tail risk metrics (CVaR, expected shortfall, max loss at 99%)
    - Display stress test recommendations (diversification, hedging, safe-haven allocation)
    - _Requirements: 53.5, 53.8, 53.9, 53.10, 53.11, 53.12_

  - [ ] 56.5 Implement stress test visualization
    - Display NAV drawdown chart, recovery timeline, worst-case projections
    - _Requirements: 53.5_

- [ ] 57. Public API and webhooks
  - [ ] 57.1 Implement public REST API
    - Expose endpoints for portfolio, transactions, prices, market data, AI recommendations
    - Implement API key authentication with rate limiting per key
    - Implement API versioning (v1, v2)
    - Implement rate limiting (100 req/min authenticated, 10 req/min unauthenticated)
    - _Requirements: 54.1, 54.2, 54.9, 54.11_
  
  - [ ] 57.2 Implement API documentation
    - Create OpenAPI/Swagger specification
    - Provide interactive API explorer
    - _Requirements: 54.3_
  
  - [ ] 57.3 Implement webhook system
    - Allow registering webhook URLs for events (price alerts, NAV milestones, pattern detections)
    - Send HTTP POST with JSON payload on events
    - Implement retry logic with exponential backoff (3 retries: 1s, 5s, 15s)
    - Implement webhook signature verification (HMAC-SHA256)
    - _Requirements: 54.4, 54.5, 54.6, 54.7_
  
  - [ ] 57.4 Implement webhook management UI
    - Allow registering, testing, viewing logs, disabling webhooks
    - _Requirements: 54.8_
  
  - [ ] 57.5 Implement API analytics and SDKs
    - Provide API usage analytics (request count, error rate, endpoints, violations)
    - Provide SDKs or client libraries (JavaScript, Python, Go)
    - _Requirements: 54.10, 54.12_

- [ ] 58. Integration with external services
  - [ ] 58.1 Implement portfolio export integrations
    - Export to Yahoo Finance CSV, Google Sheets format, Personal Capital format
    - Implement one-click export to Google Sheets with auto sheet creation
    - Implement OAuth integration with Google account
    - _Requirements: 55.1, 55.2, 55.3_
  
  - [ ] 58.2 Implement tax and accounting exports
    - Generate tax reports compatible with VN tax software
    - Export to accounting software formats (QuickBooks CSV, Xero, generic CSV)
    - _Requirements: 55.4, 55.5_
  
  - [ ] 58.3 Implement scheduled exports
    - Support automatic exports to Google Drive or Dropbox (daily, weekly, monthly)
    - _Requirements: 55.7_
  
  - [ ] 58.4 Implement watchlist and calendar imports
    - Import watchlists from TradingView, Yahoo Finance, generic symbol lists
    - Integrate with Google Calendar/Outlook for corporate events, ex-dividend, earnings dates
    - _Requirements: 55.8, 55.10_
  
  - [ ] 58.5 Implement automation integrations
    - Implement Zapier integration for custom workflows
    - Implement IFTTT integration for simple automation rules
    - _Requirements: 55.11, 55.12_
  
  - [ ] 58.6 Implement integrations UI
    - Display "Connect" section in settings with integration status
    - _Requirements: 55.6_
  
  - [ ] 58.7 Implement advisor reporting
    - Generate reports for financial advisors (PDF with charts, Excel with breakdowns)
    - _Requirements: 55.9_

- [ ] 59. User experience enhancements
  - [ ] 59.1 Implement command palette
    - Create command palette (Cmd+K or Ctrl+K) with fuzzy search
    - Support action execution (navigate, add transaction, create alert, search symbols, settings)
    - _Requirements: 56.1, 56.2_
  
  - [ ] 59.2 Implement keyboard shortcuts
    - Add shortcuts (A: add transaction, C: chart, S: screener, H: chat, R: refresh)
    - Display shortcuts help modal (? key)
    - _Requirements: 56.3, 56.4_
  
  - [ ] 59.3 Implement customizable dashboard layouts
    - Implement drag-and-drop grid system for widgets
    - Provide widget library (NAV, allocation, watchlist, transactions, alerts, indices, sector heatmap, AI insights)
    - Persist layouts in backend per user
    - Provide dashboard presets (Trader, Long-term Investor, Dividend Investor)
    - _Requirements: 56.5, 56.6, 56.7, 56.13_
  
  - [ ] 59.4 Implement undo/redo functionality
    - Implement undo/redo for transaction operations (Cmd+Z / Cmd+Shift+Z)
    - Maintain undo history of last 20 actions per session
    - _Requirements: 56.8, 56.9_
  
  - [ ] 59.5 Implement bulk operations and quick actions
    - Support bulk transaction editing (select multiple, batch delete/edit/export)
    - Provide quick actions menu (right-click or long-press on holdings, transactions, watchlist)
    - _Requirements: 56.10, 56.11_

  - [ ] 59.6 Implement smart search and auto-switching
    - Implement smart search in header (symbol lookup, company name, action suggestions)
    - Implement dark mode auto-switching (system preference or time-based)
    - _Requirements: 56.12, 56.14_

- [ ] 60. Enhanced security features
  - [ ] 60.1 Implement two-factor authentication
    - Support TOTP (Google Authenticator, Authy compatible)
    - Require 2FA setup during account creation or in settings
    - Require password + TOTP for login when enabled
    - Provide 10 backup codes during setup
    - _Requirements: 57.1, 57.2, 57.3, 57.4_
  
  - [ ] 60.2 Implement biometric authentication
    - Support fingerprint and Face ID using WebAuthn
    - _Requirements: 57.5_
  
  - [ ] 60.3 Implement session management
    - Display all active sessions (device, browser, location, last activity)
    - Allow revoking individual sessions or all except current
    - _Requirements: 57.6, 57.7_
  
  - [ ] 60.4 Implement suspicious activity detection
    - Flag login from new device, unusual location, multiple failed attempts, unusual API usage
    - Send email alert and require additional verification for suspicious sessions
    - Implement device fingerprinting for trusted devices
    - _Requirements: 57.8, 57.9, 57.10_

  - [ ] 60.5 Implement additional security measures
    - Support security questions for account recovery
    - Implement API request signing for sensitive operations (prevent CSRF)
    - Enforce password complexity (min 12 chars, uppercase, lowercase, number, special)
    - Implement password breach detection (Have I Been Pwned API)
    - Implement rate limiting on auth endpoints (max 5 login attempts per 15 min per IP)
    - _Requirements: 57.11, 57.12, 57.13, 57.14, 57.15_

- [ ] 61. Market calendar and events
  - [ ] 61.1 Implement market calendar backend
    - Maintain VN market holidays and trading schedule
    - Fetch earnings calendar data (symbol, date, estimated/actual EPS)
    - Fetch IPO calendar data (company, date, price range, lot size, subscription period)
    - Fetch economic calendar (GDP, CPI, interest rates, trade balance, FDI)
    - Fetch corporate action calendar (splits, bonus shares, rights issues, M&A dates)
    - _Requirements: 58.1, 58.3, 58.6, 58.7, 58.12_
  
  - [ ] 61.2 Implement market calendar frontend
    - Display calendar view with all event types
    - Highlight earnings dates for holdings
    - Support filtering by event type
    - Display "Today's Events" widget on dashboard
    - _Requirements: 58.2, 58.4, 58.8, 58.11_
  
  - [ ] 61.3 Implement calendar alerts and export
    - Send reminders for earnings (1 week, 1 day, day of)
    - Allow custom reminders with configurable lead time
    - Support calendar export (Google Calendar, Outlook, iCal)
    - _Requirements: 58.5, 58.9, 58.10_

- [ ] 62. Vietnamese financial news sentiment analysis
  - [ ] 62.1 Implement VN news source integration
    - Fetch from CafeF, VietStock, Đầu tư Chứng khoán, Nhịp sống kinh tế, Báo Đầu tư
    - Extract key entities (company names, symbols, sectors, key figures)
    - _Requirements: 59.1, 59.3_
  
  - [ ] 62.2 Implement Vietnamese sentiment analysis
    - Implement Vietnamese NLP sentiment classification (positive, negative, neutral)
    - Aggregate sentiment per symbol over rolling windows (24h, 7d, 30d)
    - _Requirements: 59.2, 59.4_
  
  - [ ] 62.3 Implement sentiment display
    - Display sentiment indicators on stock detail pages (current, trend, news count)
    - Display sentiment timeline chart
    - Provide news summaries in Vietnamese and English
    - _Requirements: 59.5, 59.10, 59.12_
  
  - [ ] 62.4 Implement breaking news detection
    - Detect breaking news and significant sentiment shifts
    - Trigger alerts for holdings when sentiment changes >30 points in 24h
    - _Requirements: 59.11_
  
  - [ ] 62.5 Integrate sentiment into AI agents
    - Incorporate news and community sentiment into Analysis_Agent confidence scores
    - Reference significant news events in Supervisor_Agent recommendations
    - _Requirements: 59.8, 59.9_

  - [ ] 62.6 Implement community forum monitoring
    - Monitor CafeF forum, VietStock forum, Facebook groups
    - Compute community sentiment from posts (frequency, keywords, engagement)
    - _Requirements: 59.6, 59.7_

- [ ] 63. Customizable alert scheduling and grouping
  - [ ] 63.1 Implement alert scheduling
    - Allow per-alert-type schedules (trading hours, business hours, 24/7)
    - Support quiet hours configuration
    - Support day-of-week filtering
    - _Requirements: 60.1, 60.2, 60.3_
  
  - [ ] 63.2 Implement alert grouping and priority
    - Group multiple alerts of same type within 1-hour window
    - Send summary notification for grouped alerts
    - Define priority levels (critical, high, medium, low)
    - Configure delivery channels per priority level
    - _Requirements: 60.4, 60.5, 60.6, 60.7_
  
  - [ ] 63.3 Implement alert digest and snoozing
    - Support daily summary email (digest mode)
    - Support alert snoozing (1 hour, 1 day, 1 week)
    - _Requirements: 60.8, 60.9_
  
  - [ ] 63.4 Implement alert management UI and smart throttling
    - Display active, snoozed, history, configuration per type
    - Implement smart throttling for repeated similar alerts
    - Support complex alert rules with conditions
    - _Requirements: 60.10, 60.11, 60.12_

- [ ] 64. Social sentiment from VN investor community
  - [ ] 64.1 Implement community forum monitoring
    - Monitor CafeF forum, VietStock forum, subreddits, Facebook groups
    - Extract discussion metrics per symbol (post count, comment count, sentiment distribution)
    - _Requirements: 61.1, 61.2_
  
  - [ ] 64.2 Implement trending stock detection
    - Identify trending stocks (3x discussion volume vs 7-day average)
    - Display "Community Trending" section on dashboard (top 10 stocks)
    - _Requirements: 61.4, 61.5_
  
  - [ ] 64.3 Implement topic extraction and analysis
    - Extract discussion topics using topic modeling (earnings, technical analysis, news, rumors)
    - Display word cloud or topic summary per symbol
    - Identify influential community members and weight sentiment
    - _Requirements: 61.6, 61.7, 61.8_
  
  - [ ] 64.4 Implement community sentiment display
    - Display indicators on stock detail pages (volume, sentiment %, trending status)
    - Display sentiment timeline
    - _Requirements: 61.3, 61.10_
  
  - [ ] 64.5 Integrate community sentiment into AI
    - Incorporate into Supervisor_Agent recommendations
    - Allow alerts for sentiment changes (trending, sentiment shift)
    - Implement spam and bot detection
    - _Requirements: 61.9, 61.11, 61.12_

- [ ] 65. Portfolio performance attribution
  - [ ] 65.1 Implement performance attribution computation
    - Break down total return into contributions from holdings, asset types, sectors, timing
    - Compute holding-level attribution (contribution in VND and %, weight, return)
    - Compute sector attribution (contribution, weight vs benchmark, selection vs allocation effect)
    - Compute asset type attribution
    - _Requirements: 62.1, 62.2, 62.4, 62.5_
  
  - [ ] 65.2 Implement timing and currency attribution
    - Compute timing attribution (actual vs buy-and-hold returns)
    - Compute currency attribution (USD/VND impact on crypto and USD holdings)
    - _Requirements: 62.6, 62.8_
  
  - [ ] 65.3 Implement attribution visualization
    - Display attribution chart (top contributors and detractors)
    - Display waterfall charts (starting NAV, contributions, ending NAV)
    - _Requirements: 62.3, 62.7_
  
  - [ ] 65.4 Implement decision analysis
    - Identify best and worst investment decisions
    - Display "What Went Right / What Went Wrong" summary
    - Compute skill vs luck analysis
    - _Requirements: 62.9, 62.10, 62.11_
  
  - [ ] 65.5 Integrate attribution into AI recommendations
    - Incorporate insights into Supervisor_Agent (double down, cut losses, learn from past)
    - _Requirements: 62.12_


- [ ] 66. Checkpoint - Verify new features (Requirements 39-62)
  - Test mobile responsiveness and PWA installation
  - Test WebSocket price streaming and reconnection
  - Test push notifications, email, and SMS alerts
  - Test social features and community sentiment
  - Test transaction import (CSV and broker API)
  - Test portfolio rebalancing and tax-loss harvesting
  - Test voice input and natural language queries
  - Test tax optimization and reporting
  - Test educational content and onboarding
  - Test data quality validation and anomaly detection
  - Test performance optimizations (caching, code splitting, CDN)
  - Test compliance audit trail and data retention
  - Test advanced charting (templates, replay, pattern recognition)
  - Test stress testing and Monte Carlo simulation
  - Test public API and webhooks
  - Test external service integrations
  - Test UX enhancements (command palette, keyboard shortcuts, custom layouts)
  - Test enhanced security (2FA, biometric, session management)
  - Test market calendar and events
  - Test Vietnamese news sentiment analysis
  - Test customizable alert scheduling and grouping
  - Test social sentiment from VN community
  - Test portfolio performance attribution
  - Ensure all tests pass, ask the user if questions arise.


- [ ] 67. Integration testing for new features
  - [ ]* 67.1 Write integration tests for mobile and real-time features
    - Test PWA offline caching and service worker
    - Test WebSocket connection lifecycle and price streaming
    - Test push notification delivery and click handling
    - Test email and SMS alert delivery
  
  - [ ]* 67.2 Write integration tests for social and import features
    - Test community performance sharing and leaderboards
    - Test public watchlist following
    - Test CSV transaction import with various broker formats
    - Test broker API sync
  
  - [ ]* 67.3 Write integration tests for advanced features
    - Test portfolio rebalancing suggestions
    - Test voice input and natural language query processing
    - Test tax report generation
    - Test stress testing and Monte Carlo simulation
  
  - [ ]* 67.4 Write integration tests for API and integrations
    - Test public API authentication and rate limiting
    - Test webhook delivery and retry logic
    - Test Google Sheets export
    - Test calendar integration
  
  - [ ]* 67.5 Write property tests for new features
    - **Property 64: PWA Offline Cache Consistency**
    - **Property 65: WebSocket Message Ordering**
    - **Property 66: Push Notification Grouping**
    - **Property 67: Alert Priority Routing**
    - **Property 68: CSV Import Duplicate Detection**
    - **Property 69: Rebalancing Cost Effectiveness**
    - **Property 70: Tax-Loss Harvesting Opportunity Detection**
    - **Property 71: Voice Query Intent Extraction**
    - **Property 72: Sentiment Aggregation Accuracy**
    - **Property 73: Monte Carlo Simulation Convergence**
    - **Property 74: Webhook Signature Verification**
    - **Property 75: Chart Pattern Detection Accuracy**
    - **Property 76: Performance Attribution Sum Consistency**
    - **Validates: Requirements 39.7, 40.6, 41.9, 42.7, 44.12, 45.7, 45.8, 46.6, 59.4, 53.6, 54.7, 52.6, 62.1**


- [ ] 68. Final checkpoint and launch preparation
  - Run full test suite (unit, property, integration, E2E)
  - Verify all 62 requirements are implemented
  - Verify all 76 correctness properties are tested
  - Perform security audit (including 2FA, biometric, session management)
  - Perform performance testing under load
  - Test with real market data during trading hours
  - Verify offline mode and PWA functionality
  - Test multi-language support thoroughly
  - Test dark theme across all views
  - Test mobile responsiveness on various devices
  - Test WebSocket stability under high load
  - Test all alert delivery channels (in-app, push, email, SMS)
  - Verify data quality monitoring and anomaly detection
  - Test compliance audit trail completeness
  - Prepare launch checklist
  - Ensure all tests pass, ask the user if questions arise.

## Notes

- Tasks marked with `*` are optional and can be skipped for faster MVP delivery
- Each task references specific requirements for traceability
- Property tests validate universal correctness properties from the design document
- Checkpoints ensure incremental validation at logical milestones
- The implementation follows a bottom-up approach: data layer → business logic → AI agents → API → frontend
- Backend uses Go with Gin framework, vnstock-go library, and langchaingo for AI
- Frontend uses Next.js 16 with TypeScript, Tailwind CSS, and lightweight-charts
- Database is SQLite for development, PostgreSQL for production
- All monetary values are stored in VND as the base currency
- VCI is the primary data source with KBS as fallback
- The platform supports 21 technical indicators across 5 categories
- The multi-agent AI system uses 5 specialized agents with autonomous monitoring
- The Knowledge_Base learns from pattern outcomes to improve recommendations over time
- New features (Requirements 39-62) add mobile support, real-time streaming, social features, advanced analytics, and enhanced security
- Total of 76 correctness properties to be tested across all features

