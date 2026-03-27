# Requirements Document

## Introduction

MyFi is a unified personal finance platform that consolidates all asset types (VN stocks, gold, crypto, savings, bank accounts, bonds) into a single dashboard. The platform provides near real-time price tracking, advanced TradingView-style charting with 21 technical indicators, a multi-agent AI advisory system powered by langchaingo, and NAV-based portfolio recommendations. The backend is built with Go (Gin framework) using vnstock-go library which supports both VCI (Vietcap Securities) and KBS data sources for Vietnamese stock market data, CoinGecko for crypto, and Doji/SJC for gold. The frontend is Next.js 16 with TypeScript, Tailwind CSS, and lightweight-charts. The platform includes an autonomous AI monitoring system that continuously scans for market patterns (e.g., accumulation phases) and proactively alerts the user, building a knowledge base of observations to improve recommendations over time. The platform provides sector/industry-level analysis using ICB sector indices (VNIT, VNIND, VNCONS, VNCOND, VNHEAL, VNENE, VNUTI, VNREAL, VNFIN, VNMAT) to contextualize individual stock performance within their sector, track sector trends over multiple time horizons, and power an enhanced stock screener with comprehensive fundamental filters. A comprehensive market data service consolidates all data categories available from the vnstock-go API (listing info, company info, financial reports, trading statistics, market statistics, market valuation, open fund data, commodity market, macroeconomics) into a unified data layer. A stock comparison tool enables side-by-side valuation, performance, and correlation analysis across multiple stocks with sector-based grouping. The platform supports multiple named watchlists with per-symbol price alerts and backend persistence for Monitor_Agent integration. Portfolio performance analytics provide time-weighted return (TWR), money-weighted return (MWRR/XIRR), NAV equity curve charting, and benchmark comparison against VN-Index and VN30. Risk metrics including Sharpe ratio, max drawdown, portfolio beta, volatility, and Value at Risk (VaR) are computed at both portfolio and per-holding levels. Export and reporting capabilities allow transaction history, P&L reports, and tax-friendly capital gains summaries to be exported as CSV and PDF. A multi-currency display layer with FX_Service tracks USD/VND rates and provides a frontend toggle between VND and USD display. Dividend and corporate action tracking automatically adjusts cost basis for stock splits and bonus shares, records dividend payments, and provides a calendar of upcoming corporate events. Goal-based financial planning lets users set financial targets and track progress against NAV with Supervisor_Agent integration. A backtesting engine enables users to define indicator-based trading strategies and simulate them against historical OHLCV data. Centralized API rate limiting and quota management in the Data_Source_Router prevents upstream API blocking with per-source request limits and request queuing. Offline and degraded mode support ensures graceful operation when internet connectivity is lost, serving cached data with clear offline indicators. An AI recommendation audit trail logs every Supervisor_Agent recommendation with full inputs, outputs, and outcome tracking to feed back into the Knowledge_Base. User authentication and security with JWT-based session management, bcrypt password hashing, and account lockout protection secure all API endpoints. The platform supports dark theme and multi-language (Vietnamese/English) capabilities with persistent user preferences, locale-aware formatting, and theme-adaptive chart rendering. A dual-mode Trading and Investment Recommendation Engine proactively scans the entire VN stock universe to identify optimal trading (short-term) and investment (long-term) opportunities, producing structured signals with entry prices, stop-loss/take-profit levels, confidence scores, and reasoning backed by technical indicators (including MFI — Money Flow Index as the 22nd indicator), volume anomaly detection, price action analysis, money flow analysis (institutional and foreign flow tracking), sector rotation momentum, and fundamental analysis.

## Glossary

- **MyFi_Platform**: The unified personal finance web application (Go backend + Next.js frontend)
- **Portfolio_Engine**: The backend module responsible for tracking asset holdings, transactions, cost basis, and P&L calculations
- **Price_Service**: The backend service that fetches and caches near real-time prices from VCI, KBS, CoinGecko, Doji, and other data sources
- **Data_Source_Router**: The backend module that determines the optimal data source (VCI or KBS) for each data category based on data richness, and handles failover between sources
- **Market_Data_Service**: The backend service that consolidates all data categories from the vnstock-go API (listing, company, financial, trading, market, valuation, fund, commodity, macro) into a unified data layer accessible to all platform components
- **Chart_Engine**: The frontend charting module built on lightweight-charts providing candlestick charts with 21 technical indicators across trend, momentum, volatility, volume, and statistics categories, plus drawing tools
- **Multi_Agent_System**: The langchaingo-based AI system composed of specialized agents (Price_Agent, Analysis_Agent, News_Agent, Monitor_Agent, Supervisor_Agent)
- **Price_Agent**: An AI agent responsible for fetching current and historical price data for any asset type
- **Analysis_Agent**: An AI agent that performs technical analysis using all 21 supported indicators, fundamental analysis, and sector-relative analysis using price action, volume, sector context, and predefined rules
- **News_Agent**: An AI agent that fetches and summarizes relevant financial news from CafeF RSS, Google search, and other sources
- **Monitor_Agent**: An autonomous AI agent that continuously scans market data for predefined patterns (accumulation, distribution, breakout) and generates proactive alerts
- **Supervisor_Agent**: The orchestrating AI agent that gathers outputs from all sub-agents and produces NAV-based recommendations
- **Knowledge_Base**: A persistent store of market observations, detected patterns, and their outcomes used by the Monitor_Agent to learn and improve pattern detection accuracy over time
- **Pattern_Detector**: The component within the Monitor_Agent that identifies market patterns such as accumulation (gom hàng), distribution, breakouts, and volume anomalies
- **Alert_Service**: The backend service that delivers proactive notifications to the user when the Monitor_Agent detects actionable market patterns
- **NAV**: Net Asset Value — the total value of all assets in the user's portfolio minus liabilities, denominated in VND
- **Asset_Registry**: The data model and storage layer for all supported asset types and their metadata
- **VCI_Source**: Vietcap Securities HTTP API — provides 77 columns per symbol for price board, supports OHLCV chart, intraday, order book, company overview, shareholders, officers, news, and financial statements
- **KBS_Source**: KBS Securities HTTP API — provides 28 columns per symbol for price board, serves as alternative stable source for stock data
- **Transaction_Ledger**: The record of all buy, sell, deposit, withdrawal, and interest accrual events across asset types
- **Gold_Service**: The backend module fetching gold prices from Doji API (primary) and SJC (fallback)
- **Savings_Tracker**: The module that tracks savings accounts, term deposits, and computes accrued interest
- **Sector_Service**: The backend module that fetches and caches ICB sector index data (VNIT, VNIND, VNCONS, VNCOND, VNHEAL, VNENE, VNUTI, VNREAL, VNFIN, VNMAT), computes sector performance metrics, and maps individual stocks to their ICB sector classification
- **ICB_Sector**: The Industry Classification Benchmark sector assigned to each stock on the VN market, provided by the vnstock-go API via VCI/KBS data sources
- **Sector_Trend**: A computed metric indicating whether a sector index is in an uptrend, downtrend, or sideways state based on price action over a specified time period
- **Screener_Service**: The backend module that provides advanced stock filtering with real data from the Data_Source_Router, supporting fundamental, sector, and trend-based filter criteria
- **Filter_Preset**: A saved set of screener filter criteria that a user can name, persist, and reapply
- **Comparison_Engine**: The backend module that computes valuation, performance, and correlation comparisons across multiple selected stocks for the stock comparison tool
- **Correlation_Matrix**: A matrix showing the statistical correlation coefficients between the price movements of selected stocks over a specified time period
- **Commodity_Service**: The backend module that fetches commodity market data (gold VN/global, crude oil, natural gas, steel, iron ore, agricultural commodities, VN pork prices) from available data sources
- **Macro_Service**: The backend module that fetches and serves macroeconomic indicators relevant to the VN market
- **Fund_Service**: The backend module that fetches and serves open fund (mutual fund) data and performance metrics
- **Watchlist_Service**: The backend module that persists named watchlists, manages symbol ordering, and stores per-symbol price alert thresholds, syncing watchlist data to the Monitor_Agent for scanning
- **Performance_Engine**: The backend module that computes portfolio performance metrics including time-weighted return (TWR), money-weighted return (MWRR/XIRR), daily NAV snapshots, equity curve data, and benchmark comparisons against VN-Index and VN30
- **Risk_Service**: The backend module that computes portfolio-level and per-holding risk metrics including Sharpe ratio, max drawdown, portfolio beta, annualized volatility, Value at Risk (VaR), and per-holding risk contribution
- **Export_Service**: The backend module that generates CSV and PDF exports of transaction history, portfolio snapshots, P&L reports, and tax-friendly capital gains summaries
- **FX_Service**: The backend module that fetches and caches the USD/VND exchange rate from CoinGecko (USDT/VND pair) with a hardcoded fallback of 25,400, enabling multi-currency display across the platform
- **Corporate_Action_Service**: The backend module that fetches dividend calendars, stock split events, and bonus share events from VCI/KBS via the Data_Source_Router, and auto-adjusts cost basis and holdings in the Transaction_Ledger
- **Goal_Planner**: The backend module that manages user-defined financial goals (target amount, target date, associated asset types), computes progress against current NAV, and calculates required monthly contributions
- **Backtest_Engine**: The backend module that accepts indicator-based trading strategy rules, runs them against historical OHLCV data, and computes simulation results including total return, win rate, max drawdown, Sharpe ratio, and trade statistics
- **Rate_Limiter**: The module within the Data_Source_Router that enforces per-source request rate limits, queues excess requests, and rejects requests when queue depth exceeds a configurable maximum
- **Recommendation_Audit_Log**: The persistent store that records every Supervisor_Agent recommendation with full inputs (price data, analysis data, news data), outputs (recommended actions), and outcome tracking (price changes at 1-day, 7-day, 14-day, 30-day intervals)
- **Auth_Service**: The backend module that handles user authentication with bcrypt-hashed passwords, JWT token issuance and validation, session management, account lockout after failed attempts, and password change functionality
- **Theme_Service**: The frontend service that manages theme state (light/dark mode), persists theme preference in local storage, and coordinates theme updates across all UI components including charts
- **I18n_Service**: The frontend internationalization service that manages language state (Vietnamese/English), provides translation strings, persists language preference in local storage, and handles locale-aware formatting for numbers, dates, and currency
- **Recommendation_Engine**: The backend module that orchestrates the dual-mode (Trading and Investment) recommendation pipeline, coordinating the Stock_Scanner, Analysis_Agent, and Supervisor_Agent to produce structured Trading_Signal and Investment_Signal outputs for VN stocks
- **Trading_Signal**: A structured short-term trading recommendation containing: symbol, entry price, stop-loss price (ATR-based), take-profit price (risk-reward ratio based), risk/reward ratio, confidence score (0–100), signal direction (long/short), and reasoning
- **Investment_Signal**: A structured long-term investment recommendation containing: symbol, entry price zone (low–high), target price, suggested holding period (weeks/months), fundamental reasoning, confidence score (0–100), and key fundamental metrics supporting the recommendation
- **Money_Flow_Index**: A volume-weighted momentum indicator (MFI) that combines price and volume data to measure buying and selling pressure on a scale of 0–100, where values above 80 indicate overbought conditions and below 20 indicate oversold conditions; the 22nd technical indicator supported by the platform
- **Smart_Money_Flow**: A composite metric tracking net foreign investor and institutional investor buying/selling activity for a given symbol, used as a proxy for informed capital movement in the VN stock market
- **Stock_Scanner**: The backend module that proactively scans the entire VN stock universe (HOSE, HNX, UPCOM), ranks stocks by composite signal strength across technical, volume, money flow, and fundamental dimensions, and feeds the top candidates to the Recommendation_Engine

## Requirements

### Requirement 1: Unified Asset Registry

**User Story:** As an investor, I want to register and manage all my asset types (VN stocks, gold, crypto, savings, bank accounts, bonds) in one place, so that I have a single source of truth for my entire portfolio.

#### Acceptance Criteria

1. THE Asset_Registry SHALL support the following asset types: VN stocks, gold (physical and SJC), cryptocurrency, savings accounts, term deposits, bank current accounts, and bonds
2. WHEN a user adds a new asset, THE Asset_Registry SHALL persist the asset with its type, quantity, acquisition cost, acquisition date, and associated account
3. WHEN a user edits an existing asset, THE Asset_Registry SHALL update the asset record and recalculate the NAV within the same request cycle
4. WHEN a user deletes an asset, THE Asset_Registry SHALL remove the asset and all associated transactions from the Transaction_Ledger
5. IF an asset type is not recognized, THEN THE Asset_Registry SHALL return an error specifying the list of supported asset types
6. THE Asset_Registry SHALL store all monetary values in VND as the base currency

### Requirement 2: VCI/KBS Dual-Source Routing and Best-Data Selection

**User Story:** As an investor, I want the platform to intelligently select the best data source (VCI or KBS) for each data category, so that I always get the most complete and reliable information available.

#### Acceptance Criteria

1. THE Data_Source_Router SHALL maintain a source preference mapping that specifies the primary source (VCI_Source or KBS_Source) for each data category: price quotes, OHLCV history, intraday data, order book, company overview, shareholders, officers/directors, news, income statements, balance sheets, cash flow statements, and financial ratios
2. THE Data_Source_Router SHALL select the primary source for each data category based on which source provides more complete data fields for that category
3. WHEN the primary source for a data category fails to respond within 10 seconds, THE Data_Source_Router SHALL automatically route the request to the alternative source for that category
4. WHEN the primary source returns empty or incomplete data for a specific symbol (zero fields populated or missing key fields), THE Data_Source_Router SHALL fetch the same data from the alternative source and return whichever response contains more populated fields
5. THE Data_Source_Router SHALL use VCI_Source as the default primary source for price board data (77 columns) and fall back to KBS_Source (28 columns) when VCI_Source is unavailable
6. THE Data_Source_Router SHALL use the source that provides director and officer information for a given symbol, falling back to the alternative source when the primary returns no records
7. THE Data_Source_Router SHALL use the source that provides financial health data (financial statements and ratios) for a given symbol, falling back to the alternative source when the primary returns no records
8. IF both VCI_Source and KBS_Source fail for a data category, THEN THE Data_Source_Router SHALL return the last cached result for that category with a stale indicator flag
9. THE Data_Source_Router SHALL log each source selection decision (chosen source, reason, response time) for monitoring and tuning the source preference mapping
10. THE Data_Source_Router SHALL implement circuit breaker logic per source: after 3 consecutive failures within 60 seconds for a source, THE Data_Source_Router SHALL route all requests to the alternative source until the failed source recovers

### Requirement 3: Real-Time Price Service with Dual-Source Fallback

**User Story:** As an investor, I want near real-time prices for all my asset types with automatic failover between data sources, so that I can see accurate valuations at any moment.

#### Acceptance Criteria

1. WHEN a price request is received for VN stocks, THE Price_Service SHALL fetch quotes via the Data_Source_Router which selects the optimal source (VCI_Source or KBS_Source) for price data
2. WHEN a price request is received for cryptocurrency, THE Price_Service SHALL fetch quotes from CoinGecko API with VND conversion
3. WHEN a price request is received for gold, THE Price_Service SHALL fetch buy/sell prices from Doji API and multiply by 1000 to convert to VND
4. IF the Doji API is unavailable, THEN THE Price_Service SHALL fall back to the SJC API for gold prices
5. IF the primary stock data source returns zero or null prices, THEN THE Price_Service SHALL request the Data_Source_Router to fetch from the alternative source before falling back to the OHLCV chart endpoint using the last 10 days of history
6. THE Price_Service SHALL cache VN stock prices with a 15-minute TTL, gold prices with a 1-hour TTL, and crypto prices with a 5-minute TTL
7. THE Price_Service SHALL batch multiple symbol requests into a single API call where the data source supports it
8. IF a data source API call fails after 3 retries with exponential backoff, THEN THE Price_Service SHALL return the last cached price with a stale indicator flag

### Requirement 4: Portfolio Tracking and P&L

**User Story:** As an investor, I want to track all my transactions and see real-time profit/loss for each asset and the overall portfolio, so that I can measure my investment performance.

#### Acceptance Criteria

1. WHEN a user records a buy transaction, THE Portfolio_Engine SHALL debit the cash account and credit the asset holding with the specified quantity and cost basis
2. WHEN a user records a sell transaction, THE Portfolio_Engine SHALL credit the cash account, debit the asset holding, and compute realized P&L using the weighted average cost method
3. THE Portfolio_Engine SHALL compute unrealized P&L for each holding by comparing the current market price from Price_Service against the weighted average cost basis
4. THE Portfolio_Engine SHALL compute the total NAV by summing the current market value of all holdings across all asset types
5. WHEN the NAV is computed, THE Portfolio_Engine SHALL break down the allocation by asset type as both absolute VND values and percentage of total NAV
6. THE Transaction_Ledger SHALL record each transaction with: asset type, symbol, quantity, unit price, total value, transaction date, and transaction type (buy, sell, deposit, withdrawal, interest, dividend)
7. IF a sell transaction quantity exceeds the current holding quantity, THEN THE Portfolio_Engine SHALL reject the transaction and return an error indicating insufficient holdings

### Requirement 5: Gold Price Integration

**User Story:** As a gold investor, I want to see current Doji and SJC gold buy/sell prices, so that I can track the value of my gold holdings accurately.

#### Acceptance Criteria

1. WHEN the Gold_Service fetches prices from Doji API, THE Gold_Service SHALL parse the response and return buy and sell prices for each gold type (SJC, PNJ, 9999, etc.)
2. THE Gold_Service SHALL convert Doji API prices from thousands VND to full VND by multiplying by 1000
3. IF the Doji API returns an error or times out within 10 seconds, THEN THE Gold_Service SHALL attempt to fetch from the SJC fallback endpoint
4. IF both Doji and SJC APIs fail, THEN THE Gold_Service SHALL return the last cached gold prices with a stale indicator
5. WHEN gold prices are fetched, THE Gold_Service SHALL cache the result with a 1-hour TTL

### Requirement 6: Savings and Bank Interest Tracking

**User Story:** As a saver, I want to track my savings accounts and term deposits with automatic interest calculation, so that I can see the true value of my cash holdings.

#### Acceptance Criteria

1. WHEN a user adds a savings account, THE Savings_Tracker SHALL record the principal amount, annual interest rate, compounding frequency (monthly, quarterly, yearly), start date, and maturity date
2. THE Savings_Tracker SHALL compute accrued interest for each savings account based on the compounding formula: A = P × (1 + r/n)^(n×t), where P is principal, r is annual rate, n is compounding frequency, and t is elapsed time in years
3. WHEN a term deposit reaches its maturity date, THE Savings_Tracker SHALL flag the deposit as matured in the portfolio view
4. THE Savings_Tracker SHALL include accrued interest in the NAV calculation for each savings account
5. WHEN a user adds a bank current account, THE Savings_Tracker SHALL track the balance as a cash asset with zero or specified interest rate

### Requirement 7: Advanced Charting Engine with 21 Technical Indicators

**User Story:** As a trader, I want TradingView-style charts with a comprehensive set of 21 technical indicators across trend, momentum, volatility, volume, and statistics categories, plus drawing tools, so that I can perform thorough technical analysis directly in the app.

#### Acceptance Criteria

1. THE Chart_Engine SHALL render candlestick (OHLCV) charts using the lightweight-charts library with data fetched via the Data_Source_Router from the optimal source for OHLCV data
2. THE Chart_Engine SHALL support the following time intervals: 1 minute, 5 minutes, 15 minutes, 1 hour, 1 day, 1 week, 1 month
3. THE Chart_Engine SHALL support the following 8 trend indicators: SMA (Simple Moving Average), EMA (Exponential Moving Average), VWAP (Volume Weighted Average Price), VWMA (Volume Weighted Moving Average), ADX (Average Directional Movement Index), Aroon Indicator, Parabolic SAR (Stop and Reverse), and Supertrend
4. THE Chart_Engine SHALL support the following 7 momentum indicators: RSI (Relative Strength Index), MACD (Moving Average Convergence Divergence), Williams %R, CMO (Chande Momentum Oscillator), Stochastic Oscillator, ROC (Rate of Change), and Momentum
5. THE Chart_Engine SHALL support the following 4 volatility indicators: Bollinger Bands, Keltner Channel, ATR (Average True Range), and Standard Deviation
6. THE Chart_Engine SHALL support the following volume indicator: OBV (On-Balance Volume)
7. THE Chart_Engine SHALL support the following statistics indicator: Linear Regression
8. WHEN a user selects an indicator, THE Chart_Engine SHALL compute the indicator values from the OHLCV data and overlay the result on the chart (overlay indicators on the price pane, oscillator indicators in a separate pane below)
9. THE Chart_Engine SHALL allow the user to configure parameters for each indicator (e.g., period length for SMA/EMA, overbought/oversold levels for RSI, K/D periods for Stochastic)
10. THE Chart_Engine SHALL provide drawing tools including: trend lines, horizontal lines, Fibonacci retracement, and rectangle selection
11. WHEN a user draws on the chart, THE Chart_Engine SHALL persist the drawing state in local storage so drawings survive page reloads
12. THE Chart_Engine SHALL display a volume histogram below the price chart synchronized with the same time axis
13. WHEN the user switches time intervals, THE Chart_Engine SHALL fetch new OHLCV data from the backend and re-render the chart with all active indicators recalculated
14. THE Chart_Engine SHALL allow the user to add multiple instances of the same indicator with different parameters (e.g., SMA(20) and SMA(50) simultaneously)

### Requirement 8: Multi-Agent AI System Architecture

**User Story:** As an investor, I want an AI system composed of specialized agents that each handle a distinct responsibility (price fetching, analysis, news, autonomous monitoring, supervision), so that I receive comprehensive and well-reasoned financial advice both on-demand and proactively.

#### Acceptance Criteria

1. THE Multi_Agent_System SHALL be implemented using the langchaingo library with five distinct agents: Price_Agent, Analysis_Agent, News_Agent, Monitor_Agent, and Supervisor_Agent
2. THE Multi_Agent_System SHALL execute agents in a coordinated pipeline where the Supervisor_Agent orchestrates the other agents and aggregates their outputs for on-demand queries
3. WHEN the Supervisor_Agent receives a user query, THE Supervisor_Agent SHALL determine which sub-agents to invoke based on the query content and dispatch requests to each relevant agent in parallel where possible
4. THE Multi_Agent_System SHALL pass structured data between agents using a defined message schema containing agent name, payload type, and payload content
5. IF any sub-agent fails or times out within 30 seconds, THEN THE Supervisor_Agent SHALL proceed with the outputs from the remaining agents and note the missing data source in the final response
6. THE Multi_Agent_System SHALL support configurable LLM providers (OpenAI, Anthropic, Google, Qwen, Bedrock) as already implemented in the existing getLLM function
7. THE Monitor_Agent SHALL operate autonomously on a configurable schedule independent of user chat queries

### Requirement 9: Price Agent with Dual-Source Support

**User Story:** As part of the AI system, I want a dedicated agent that fetches current and historical prices for any asset using the best available data source, so that other agents have accurate market data to work with.

#### Acceptance Criteria

1. WHEN the Price_Agent receives a symbol and asset type, THE Price_Agent SHALL fetch the current price via the Data_Source_Router which selects the optimal source (VCI_Source or KBS_Source for stocks, CoinGecko for crypto, Doji for gold)
2. WHEN the Price_Agent receives a request for historical data, THE Price_Agent SHALL fetch OHLCV history for the specified time range and interval via the Data_Source_Router
3. THE Price_Agent SHALL return a structured response containing: current price, price change, percentage change, volume, data source used, and historical data points if requested
4. IF the primary data source fails, THEN THE Price_Agent SHALL use the fallback source as determined by the Data_Source_Router
5. THE Price_Agent SHALL format all prices in VND for the Supervisor_Agent

### Requirement 10: Analysis Agent with Sector Context and Full Indicator Suite

**User Story:** As part of the AI system, I want a dedicated agent that performs technical analysis using all 22 supported indicators (including MFI), fundamental analysis, money flow analysis, and sector-relative analysis, so that the supervisor can incorporate comprehensive analytical insights with sector context into recommendations.

#### Acceptance Criteria

1. WHEN the Analysis_Agent receives OHLCV data, THE Analysis_Agent SHALL compute all applicable technical indicators from the 22 supported indicators: trend indicators (SMA, EMA, VWAP, VWMA, ADX, Aroon, Parabolic SAR, Supertrend), momentum indicators (RSI, MACD, Williams %R, CMO, Stochastic Oscillator, ROC, Momentum), volatility indicators (Bollinger Bands, Keltner Channel, ATR, Standard Deviation), volume indicators (OBV, MFI), and statistics indicators (Linear Regression)
2. THE Analysis_Agent SHALL use the following default parameters for key indicators: SMA(20), SMA(50), EMA(12), EMA(26), RSI(14), MACD(12,26,9), Bollinger Bands(20,2), ADX(14), Aroon(25), Parabolic SAR(0.02,0.2), Supertrend(10,3), Stochastic(14,3,3), ATR(14), Keltner Channel(20,2), Williams %R(14), CMO(14), ROC(12), OBV, and MFI(14)
3. THE Analysis_Agent SHALL identify price action patterns including: support/resistance levels, trend direction (uptrend, downtrend, sideways), volume anomalies (volume exceeding 2x the 20-day average), and candlestick patterns (hammer, engulfing, doji, morning/evening star)
14. WHEN the Analysis_Agent receives OHLCV and volume data, THE Analysis_Agent SHALL compute the Money_Flow_Index (MFI) using a 14-period default, classifying readings above 80 as overbought, below 20 as oversold, and divergences between MFI and price as potential reversal signals
15. THE Analysis_Agent SHALL compute Smart_Money_Flow for each analyzed stock by aggregating net foreign investor and net institutional investor buy/sell volumes from the price board data, and classify the flow as strong inflow, moderate inflow, neutral, moderate outflow, or strong outflow
16. THE Analysis_Agent SHALL include money flow analysis in the structured analysis summary containing: MFI value and classification, Smart_Money_Flow direction and magnitude, and whether MFI divergence from price is detected
4. WHEN the Analysis_Agent receives financial ratio data, THE Analysis_Agent SHALL evaluate fundamental metrics including P/E, P/B, EV/EBITDA, ROE, ROA, revenue growth, profit growth, dividend yield, and debt-to-equity and compare them against sector averages from the Sector_Service
5. THE Analysis_Agent SHALL produce a structured analysis summary containing: trend assessment, indicator signals (bullish/bearish/neutral for each computed indicator), key price levels, sector-relative performance, and a confidence score from 0 to 100
6. IF insufficient data is available to compute an indicator (fewer data points than the indicator period), THEN THE Analysis_Agent SHALL omit that indicator from the summary and note the reason
7. WHEN the Analysis_Agent analyzes a stock, THE Analysis_Agent SHALL retrieve the stock's ICB_Sector classification from the Sector_Service and include the sector name in the analysis context
8. WHEN the Analysis_Agent analyzes a stock, THE Analysis_Agent SHALL compare the stock's price performance (1-week, 1-month, 3-month, 1-year returns) against the corresponding ICB sector index performance over the same periods to determine if the stock is outperforming or underperforming its sector
9. WHEN the Analysis_Agent analyzes a stock, THE Analysis_Agent SHALL compare the stock's fundamental metrics (P/E, P/B, ROE, ROA, dividend yield) against the median values of stocks within the same ICB_Sector
10. THE Analysis_Agent SHALL evaluate the Sector_Trend for the stock's sector and factor the sector momentum (uptrend, downtrend, sideways) into the overall analysis confidence score
11. THE Analysis_Agent SHALL detect sector rotation signals by comparing relative performance changes across ICB sectors over 1-week and 1-month periods and flag sectors gaining or losing momentum
12. WHEN the Analysis_Agent produces a recommendation, THE Analysis_Agent SHALL include a sector context section containing: sector name, sector trend direction, stock vs sector relative performance, and whether the sector is currently in a rotation phase (capital flowing in or out)
13. THE Analysis_Agent SHALL generate a composite signal summary that aggregates signals from all computed indicators into an overall technical outlook (strongly bullish, bullish, neutral, bearish, strongly bearish) with the count of bullish vs bearish indicator signals

### Requirement 11: News Agent

**User Story:** As part of the AI system, I want a dedicated agent that fetches and summarizes relevant financial news, so that the supervisor can factor current events into recommendations.

#### Acceptance Criteria

1. WHEN the News_Agent receives a query with asset symbols, THE News_Agent SHALL fetch news articles from CafeF RSS feed and filter for articles mentioning the specified symbols or related keywords
2. WHEN the CafeF RSS feed does not contain relevant results, THE News_Agent SHALL perform a web search using Google for recent news related to the specified assets
3. THE News_Agent SHALL return a structured response containing: a list of relevant articles (title, source, date, URL) and a concise summary of the overall news sentiment (positive, negative, neutral) for each queried asset
4. THE News_Agent SHALL limit the returned articles to the 10 most recent and relevant items per query
5. IF both CafeF RSS and Google search fail, THEN THE News_Agent SHALL return an empty result set with a flag indicating news data is unavailable

### Requirement 12: Autonomous Market Monitor Agent

**User Story:** As an investor, I want the AI system to continuously and autonomously monitor the market for patterns like accumulation phases (gom hàng), so that I receive proactive alerts about potential opportunities without having to ask.

#### Acceptance Criteria

1. THE Monitor_Agent SHALL run autonomously on a configurable schedule (default: every 30 minutes during trading hours 9:00–15:00 ICT, every 2 hours outside trading hours)
2. THE Monitor_Agent SHALL scan OHLCV data and volume data for a configurable watchlist of symbols via the Data_Source_Router
3. THE Monitor_Agent SHALL detect accumulation patterns (gom hàng) by identifying the combination of: price consolidation within a 5% range over 10 or more trading days, daily volume exceeding 1.5x the 20-day average volume, and net foreign or institutional buying from the price board data
4. THE Monitor_Agent SHALL detect distribution patterns by identifying the combination of: price near recent highs, increasing volume on down days, and net institutional selling
5. THE Monitor_Agent SHALL detect breakout signals by identifying price movement above resistance levels with volume exceeding 2x the 20-day average
6. WHEN the Monitor_Agent detects a pattern matching any defined signal, THE Monitor_Agent SHALL generate a structured observation containing: symbol, pattern type, confidence score (0–100), supporting data points, and detection timestamp
7. WHEN the Monitor_Agent generates an observation, THE Alert_Service SHALL deliver a notification to the user containing the symbol, pattern type, confidence score, and a brief explanation of the detected signal
8. THE Monitor_Agent SHALL store each observation in the Knowledge_Base for future reference and learning
9. IF the Monitor_Agent fails to complete a scan cycle within 5 minutes, THEN THE Monitor_Agent SHALL log the failure, skip the incomplete symbols, and retry on the next scheduled cycle

### Requirement 13: Knowledge Base for Pattern Learning

**User Story:** As an investor, I want the AI system to maintain a knowledge base of past market observations and their outcomes, so that the system improves its pattern detection and recommendations over time.

#### Acceptance Criteria

1. THE Knowledge_Base SHALL persist each observation from the Monitor_Agent with: symbol, pattern type, detection date, confidence score, supporting data snapshot, and current price at detection time
2. THE Knowledge_Base SHALL track the outcome of each observation by recording the price change at 1-day, 7-day, 14-day, and 30-day intervals after detection
3. WHEN the Monitor_Agent detects a new pattern, THE Monitor_Agent SHALL query the Knowledge_Base for historical observations of the same pattern type and compute the historical success rate (percentage of past observations where the price moved in the predicted direction)
4. THE Knowledge_Base SHALL provide a query interface that returns past observations filtered by symbol, pattern type, date range, and minimum confidence score
5. WHEN the Supervisor_Agent produces recommendations, THE Supervisor_Agent SHALL incorporate relevant Knowledge_Base observations (e.g., "this stock showed accumulation signals 2 weeks ago and has since risen 15%, a similar pattern is now appearing in stock X")
6. THE Knowledge_Base SHALL retain observations for a minimum of 2 years to build sufficient historical data for pattern accuracy assessment
7. THE Knowledge_Base SHALL compute and store aggregate accuracy metrics per pattern type: total observations, success count, failure count, average price change after detection, and average confidence score at detection

### Requirement 14: Proactive Alert Service

**User Story:** As an investor, I want to receive proactive alerts when the autonomous monitor detects actionable patterns, so that I can act on opportunities without constantly watching the market.

#### Acceptance Criteria

1. WHEN the Monitor_Agent generates an observation with a confidence score of 60 or above, THE Alert_Service SHALL create an alert notification for the user
2. THE Alert_Service SHALL deliver alerts through the MyFi_Platform frontend as in-app notifications displayed in a notification panel
3. THE Alert_Service SHALL include in each alert: symbol, pattern type, confidence score, brief explanation, detection timestamp, and a link to view the stock's chart
4. THE Alert_Service SHALL persist all alerts in the database so the user can review past alerts
5. THE Alert_Service SHALL allow the user to configure alert preferences: minimum confidence threshold, specific pattern types to monitor, and symbols to include or exclude
6. THE Alert_Service SHALL deduplicate alerts by not sending a new alert for the same symbol and pattern type within a 48-hour window unless the confidence score increases by 10 or more points
7. IF the user has not viewed an alert within 24 hours, THEN THE Alert_Service SHALL mark the alert as expired but retain the record for historical reference

### Requirement 15: Supervisor Agent and NAV-Based Recommendations with Knowledge Integration

**User Story:** As an investor, I want the supervisor AI agent to synthesize all data from sub-agents, my current portfolio, and the knowledge base of past observations to give me actionable buy/sell/hold recommendations, so that I can make informed decisions about managing my NAV.

#### Acceptance Criteria

1. WHEN the Supervisor_Agent receives aggregated outputs from Price_Agent, Analysis_Agent, and News_Agent, THE Supervisor_Agent SHALL synthesize the information into a unified analysis context
2. THE Supervisor_Agent SHALL incorporate the user's current NAV, asset allocation breakdown, and individual holding positions from the Portfolio_Engine into the analysis
3. THE Supervisor_Agent SHALL query the Knowledge_Base for recent observations related to the queried symbols and incorporate historical pattern accuracy into the analysis
4. THE Supervisor_Agent SHALL produce recommendations that include: specific buy/sell/hold actions for queried assets, suggested position sizes as a percentage of NAV, risk assessment (low/medium/high), and reasoning for each recommendation
5. THE Supervisor_Agent SHALL reference relevant Knowledge_Base entries in recommendations (e.g., citing past pattern detections and their outcomes for similar stocks)
6. THE Supervisor_Agent SHALL identify new investment opportunities based on the screener data, technical signals, news sentiment, sector trend data from the Sector_Service, and Monitor_Agent observations that align with the user's current portfolio composition
11. WHEN the Recommendation_Engine produces a Trading_Signal or Investment_Signal, THE Supervisor_Agent SHALL incorporate the structured signal (entry price, stop-loss, take-profit, confidence score, reasoning) into the final recommendation output, formatting it as an actionable trade or investment plan
12. WHEN the user requests trading recommendations, THE Supervisor_Agent SHALL delegate to the Recommendation_Engine in trading mode and return Trading_Signal outputs with exact entry, stop-loss, and take-profit prices
13. WHEN the user requests investment recommendations, THE Supervisor_Agent SHALL delegate to the Recommendation_Engine in investment mode and return Investment_Signal outputs with entry price zone, target price, and suggested holding period
7. WHEN producing recommendations, THE Supervisor_Agent SHALL incorporate sector context from the Analysis_Agent including sector trend direction, sector rotation signals, and the stock's relative performance within its ICB_Sector
8. WHEN producing recommendations, THE Supervisor_Agent SHALL respect portfolio diversification by flagging when a single asset type exceeds 40% of total NAV
9. THE Supervisor_Agent SHALL format the final response in a structured format containing: summary, individual asset recommendations, portfolio-level suggestions, identified opportunities, sector context, and relevant knowledge base insights
10. IF the user's NAV data is unavailable, THEN THE Supervisor_Agent SHALL provide general market analysis without portfolio-specific recommendations and inform the user to set up portfolio tracking

### Requirement 16: Real-Time Price Updates on Frontend

**User Story:** As an investor, I want the dashboard to show near real-time price updates for all my assets without manual refresh, so that I always see current valuations.

#### Acceptance Criteria

1. THE MyFi_Platform frontend SHALL poll the Price_Service at configurable intervals: every 15 seconds for stocks during trading hours (9:00–15:00 ICT), every 60 seconds for crypto, and every 300 seconds for gold
2. WHEN new price data is received, THE MyFi_Platform frontend SHALL update the NAV display, asset allocation chart, and individual holding values without a full page reload
3. WHILE the VN stock market is outside trading hours (before 9:00 or after 15:00 ICT), THE MyFi_Platform frontend SHALL reduce stock polling frequency to every 300 seconds
4. THE MyFi_Platform frontend SHALL display a visual indicator (colored dot or timestamp) showing the freshness of each price: green for data less than 1 minute old, yellow for 1–5 minutes, red for older than 5 minutes
5. IF a price update request fails, THEN THE MyFi_Platform frontend SHALL retain the last known price and display a stale data warning icon

### Requirement 17: Unified Dashboard with NAV Overview and Alert Panel

**User Story:** As an investor, I want a single dashboard view that shows my total NAV, asset allocation, recent transactions, quick metrics, and proactive AI alerts, so that I get a complete financial picture at a glance.

#### Acceptance Criteria

1. THE MyFi_Platform frontend SHALL display the total NAV computed from real portfolio data (replacing the current hardcoded ASSET_DATA) with VND formatting
2. THE MyFi_Platform frontend SHALL display a pie chart showing asset allocation by type (stocks, gold, crypto, savings, bonds, cash) with values and percentages derived from the Portfolio_Engine
3. THE MyFi_Platform frontend SHALL display the NAV change (absolute VND and percentage) over the past 24 hours, computed by comparing current NAV against the NAV snapshot from 24 hours prior
4. THE MyFi_Platform frontend SHALL display the 5 most recent transactions from the Transaction_Ledger with transaction type, asset, amount, and date
5. THE MyFi_Platform frontend SHALL display quick metric cards for gold rate (from Gold_Service) and Bitcoin price (from CoinGecko) with 24-hour change percentages
6. THE MyFi_Platform frontend SHALL display a notification panel showing unread alerts from the Alert_Service with pattern type, symbol, confidence score, and detection time
7. WHEN the user clicks an alert in the notification panel, THE MyFi_Platform frontend SHALL navigate to the chart view for the alerted symbol with the relevant time range pre-selected

### Requirement 18: Enhanced Stock Screener with Sector Filtering and Advanced Criteria

**User Story:** As an investor, I want to filter stocks based on comprehensive fundamental criteria, sector/industry classification, and sector trend data using real market data from the best available source, so that I can discover investment opportunities matching my strategy with full sector context.

#### Acceptance Criteria

1. THE Screener_Service SHALL fetch stock screener data via the Data_Source_Router which selects the source providing the most complete financial data for each symbol (replacing the current mocked ScreenerStock data in market.go)
2. THE Screener_Service SHALL support filtering by the following fundamental criteria: P/E ratio range, P/B ratio range, market capitalization minimum, EV/EBITDA range, ROE range, ROA range, revenue growth rate range, profit growth rate range, dividend yield range, and debt-to-equity ratio range
3. THE Screener_Service SHALL support filtering by ICB_Sector classification, allowing the user to select one or more sectors (VNIT, VNIND, VNCONS, VNCOND, VNHEAL, VNENE, VNUTI, VNREAL, VNFIN, VNMAT) to include in results
4. THE Screener_Service SHALL support filtering by exchange (HOSE, HNX, UPCOM) with multi-select
5. THE Screener_Service SHALL support filtering by Sector_Trend, allowing the user to show only stocks belonging to sectors currently in an uptrend, downtrend, or sideways state as computed by the Sector_Service
6. WHEN a screener request is received, THE Screener_Service SHALL return matching stocks sorted by market capitalization descending by default
7. THE Screener_Service SHALL support sorting by multiple criteria (market cap, P/E, P/B, ROE, ROA, dividend yield, revenue growth, profit growth) with ascending or descending direction for each
8. THE Screener_Service SHALL support pagination with configurable page size (default 20) and page number
9. THE MyFi_Platform backend SHALL expose a REST API endpoint for saving, listing, updating, and deleting Filter_Preset records, each containing a name and the full set of filter criteria
10. THE MyFi_Platform backend SHALL limit the number of saved Filter_Preset records to a configurable maximum (default 10) per user
11. THE MyFi_Platform frontend SHALL display screener results in a sortable table with columns for: symbol, exchange, ICB_Sector name, market cap, P/E, P/B, EV/EBITDA, ROE, ROA, revenue growth, profit growth, dividend yield, debt-to-equity, and sector trend indicator
12. THE MyFi_Platform frontend SHALL provide a filter panel allowing the user to set ranges for each supported fundamental criterion, select sectors, select exchanges, and toggle the sector trend filter
13. THE MyFi_Platform frontend SHALL allow the user to save the current filter configuration as a named Filter_Preset and load any previously saved preset
14. THE MyFi_Platform frontend SHALL display a set of built-in default filter presets (e.g., Value Investing, High Growth, High Dividend, Low Debt) that the user can apply with one click
15. IF the Screener_Service cannot retrieve financial data for a symbol, THEN THE Screener_Service SHALL exclude that symbol from the filtered results and log the data gap

### Requirement 19: AI Chat Integration with Multi-Agent Pipeline

**User Story:** As an investor, I want the chat widget to use the full multi-agent pipeline instead of the current single-agent flow, so that I receive comprehensive AI-powered advice.

#### Acceptance Criteria

1. WHEN a user sends a message through the chat widget, THE MyFi_Platform backend SHALL route the request through the Supervisor_Agent which orchestrates the full multi-agent pipeline
2. THE MyFi_Platform backend SHALL detect asset symbols and asset types mentioned in the user message and pass them to the appropriate sub-agents
3. WHEN the multi-agent pipeline completes, THE MyFi_Platform backend SHALL return the Supervisor_Agent's structured response to the chat widget
4. THE MyFi_Platform frontend chat widget SHALL render structured recommendations with distinct sections for market data, analysis, news summary, knowledge base insights, and actionable advice
5. THE MyFi_Platform frontend chat widget SHALL support conversation history by sending the last 10 messages as context to the backend on each request
6. IF the multi-agent pipeline takes longer than 45 seconds, THEN THE MyFi_Platform backend SHALL return a partial response with whatever agent outputs have completed and indicate which agents timed out

### Requirement 20: Data Persistence

**User Story:** As an investor, I want my portfolio data, transactions, settings, knowledge base, and alert history to persist across sessions, so that I do not lose my financial records or AI learning data.

#### Acceptance Criteria

1. THE MyFi_Platform backend SHALL persist all Asset_Registry, Transaction_Ledger, Savings_Tracker, Knowledge_Base, Alert_Service, Filter_Preset, and Sector_Service mapping data to a database (SQLite for local development, PostgreSQL for production)
2. WHEN the backend starts, THE MyFi_Platform backend SHALL run database migrations to create or update the required schema
3. THE MyFi_Platform backend SHALL expose CRUD REST API endpoints for assets, transactions, savings accounts, user settings, alerts, knowledge base observations, and filter presets
4. THE MyFi_Platform frontend SHALL store user preferences (selected LLM provider, chart settings, watchlist, alert preferences) in local storage as currently implemented
5. IF a database write operation fails, THEN THE MyFi_Platform backend SHALL return an error with a descriptive message and not leave the data in a partially written state (transactions SHALL be atomic)

### Requirement 21: Sector Data Service

**User Story:** As an investor, I want the platform to maintain up-to-date sector classification and performance data for all VN stocks, so that sector-level analysis is available across the AI agents, screener, and dashboard.

#### Acceptance Criteria

1. THE Sector_Service SHALL maintain a mapping of each VN stock symbol to its ICB_Sector classification using data from the vnstock-go API (VCI_Source or KBS_Source via the Data_Source_Router)
2. THE Sector_Service SHALL fetch OHLCV history for all 10 ICB sector indices (VNIT, VNIND, VNCONS, VNCOND, VNHEAL, VNENE, VNUTI, VNREAL, VNFIN, VNMAT) via the Data_Source_Router
3. THE Sector_Service SHALL compute performance metrics for each ICB sector index over the following time periods: today (intraday change), 1 week, 1 month, 3 months, 6 months, and 1 year
4. THE Sector_Service SHALL determine the Sector_Trend (uptrend, downtrend, sideways) for each ICB sector by comparing the sector index price against its SMA(20) and SMA(50): uptrend when price is above both, downtrend when price is below both, sideways otherwise
5. THE Sector_Service SHALL compute the median fundamental metrics (P/E, P/B, ROE, ROA, dividend yield, debt-to-equity) for stocks within each ICB_Sector to serve as sector averages for the Analysis_Agent
6. THE Sector_Service SHALL cache sector index data with a 30-minute TTL during trading hours (9:00–15:00 ICT) and a 6-hour TTL outside trading hours
7. THE Sector_Service SHALL refresh the stock-to-sector mapping once per day at market open (9:00 ICT)
8. IF the Data_Source_Router fails to fetch sector index data for a specific sector, THEN THE Sector_Service SHALL return the last cached data for that sector with a stale indicator flag
9. THE Sector_Service SHALL expose REST API endpoints for: retrieving all sector performance summaries, retrieving the sector classification for a given symbol, and retrieving sector average fundamentals for a given ICB_Sector

### Requirement 22: Sector Trend Dashboard

**User Story:** As an investor, I want a dashboard view that visualizes sector/industry performance trends over multiple time periods, so that I can identify which sectors are gaining or losing momentum and make sector-informed investment decisions.

#### Acceptance Criteria

1. THE MyFi_Platform frontend SHALL display a dedicated Sector Trend dashboard view accessible from the main navigation
2. THE MyFi_Platform frontend SHALL display a heatmap visualization showing all 10 ICB sectors with color-coded performance (green gradient for positive, red gradient for negative) for the selected time period
3. THE MyFi_Platform frontend SHALL allow the user to switch the heatmap time period between: today, 1 week, 1 month, 3 months, 6 months, and 1 year
4. THE MyFi_Platform frontend SHALL display a bar chart comparing sector performance across all 10 ICB sectors for the selected time period, sorted by performance descending
5. THE MyFi_Platform frontend SHALL display a Sector_Trend indicator (uptrend, downtrend, sideways) next to each sector name using directional arrows or icons
6. THE MyFi_Platform frontend SHALL display a summary panel showing the top 3 performing sectors and bottom 3 performing sectors for the selected time period
7. WHEN the user clicks on a sector in the heatmap or bar chart, THE MyFi_Platform frontend SHALL display a detail panel showing: sector index price chart (mini candlestick), list of top 5 stocks by market cap in that sector with their individual performance, and the sector's median fundamental metrics
8. THE MyFi_Platform frontend SHALL fetch sector data from the Sector_Service REST API endpoints
9. WHEN the Analysis_Agent or Supervisor_Agent produces recommendations, THE Analysis_Agent or Supervisor_Agent SHALL reference the current sector trend data from the Sector_Service to contextualize stock-level advice with sector-level momentum
10. THE MyFi_Platform frontend SHALL auto-refresh sector trend data every 5 minutes during trading hours (9:00–15:00 ICT) and every 30 minutes outside trading hours

### Requirement 23: Comprehensive Market Data Service

**User Story:** As an investor, I want the platform to provide access to all available data categories from the vnstock-go API (listing information, company information, financial reports, trading statistics, market statistics, market valuation, open fund data, commodity market, macroeconomics) through a unified data layer, so that the AI agents, dashboard, and other features have comprehensive market data to work with.

#### Acceptance Criteria

1. THE Market_Data_Service SHALL provide listing information including: all stock symbols, market indices (VN30, VN100, VNMID, VNSML, VNALL), government bonds, and exchange information (HOSE, HNX, UPCOM) fetched via the Data_Source_Router from the vnstock-go API
2. THE Market_Data_Service SHALL provide company information including: company overview/profile, major shareholders, management team/officers, and company news for any given stock symbol, fetched via the Data_Source_Router
3. THE Market_Data_Service SHALL provide financial reports including: income statements, balance sheets, cash flow statements, and financial ratios for any given stock symbol, with support for both yearly and quarterly reporting periods, fetched via the Data_Source_Router
4. THE Market_Data_Service SHALL provide trading statistics including: real-time quotes (via price board), OHLCV history (daily, weekly, monthly, and intraday intervals), intraday tick data, order book/price depth, and bid/ask depth for any given stock symbol, fetched via the Data_Source_Router
5. THE Market_Data_Service SHALL provide market statistics including: market index data for VN30, VN100, VNMID, VNSML, and VNALL; ICB sector index data for VNIT, VNIND, VNCONS, VNCOND, VNHEAL, VNENE, VNUTI, VNREAL, VNFIN, and VNMAT; market breadth (advancing vs declining stocks); and foreign trading data, fetched via the Data_Source_Router
6. THE Market_Data_Service SHALL provide market valuation metrics including: P/E, P/B, EV/EBITDA, market capitalization, and dividend yield at the market level, sector level, and individual stock level, computed from financial data fetched via the Data_Source_Router
7. THE Fund_Service SHALL provide open fund (mutual fund) data including: fund list, fund NAV, and fund performance metrics fetched from available data sources
8. THE Commodity_Service SHALL provide commodity market data including: VN gold prices (Doji/SJC), global gold OHLCV, crude oil prices, natural gas prices, steel (HRC) prices, iron ore prices, agricultural commodity prices (corn, soybean, sugar), and VN pork prices fetched from available data sources
9. THE Macro_Service SHALL provide macroeconomic indicators relevant to the VN market fetched from available data sources
10. THE Market_Data_Service SHALL expose a unified REST API with endpoints organized by data category: /api/market/listing, /api/market/company/{symbol}, /api/market/finance/{symbol}, /api/market/trading/{symbol}, /api/market/statistics, /api/market/valuation, /api/market/funds, /api/market/commodities, and /api/market/macro
11. THE Market_Data_Service SHALL cache listing information with a 24-hour TTL, company information with a 6-hour TTL, financial reports with a 24-hour TTL, trading statistics with the same TTL as the Price_Service (15 minutes for stocks), market statistics with a 30-minute TTL, and valuation metrics with a 1-hour TTL
12. WHEN any AI agent (Price_Agent, Analysis_Agent, News_Agent, Monitor_Agent, Supervisor_Agent) requires market data, THE Market_Data_Service SHALL serve as the single data access layer providing all data categories through a consistent internal Go interface
13. THE MyFi_Platform frontend SHALL be able to access all data categories through the unified REST API for rendering dashboards, charts, company profiles, and financial reports
14. IF a data category is unavailable from the primary source, THEN THE Market_Data_Service SHALL attempt the alternative source via the Data_Source_Router and return the last cached data with a stale indicator if both sources fail
15. THE Market_Data_Service SHALL support batch requests for trading statistics, allowing multiple symbols to be queried in a single API call for price board and OHLCV data

### Requirement 24: Stock Comparison Tool

**User Story:** As an investor, I want to compare multiple stocks side-by-side across valuation metrics, price performance, and correlation, with the ability to group stocks by sector, so that I can make informed relative investment decisions (inspired by smoney.com.vn/so-sanh-co-phieu).

#### Acceptance Criteria

1. THE Comparison_Engine SHALL support comparing up to 10 stocks simultaneously across three comparison modes: Valuation (Định giá), Performance (Hiệu suất), and Correlation (Tương quan)
2. WHEN the Valuation comparison mode is selected, THE Comparison_Engine SHALL fetch P/E and P/B ratios for all selected stocks from the Market_Data_Service and return time-series data for charting these ratios over the selected time period
3. WHEN the Performance comparison mode is selected, THE Comparison_Engine SHALL compute percentage price returns for all selected stocks from OHLCV data fetched via the Market_Data_Service and return time-series data normalized to percentage change from the start of the selected period
4. WHEN the Correlation comparison mode is selected, THE Comparison_Engine SHALL compute a Correlation_Matrix of daily price returns between all pairs of selected stocks over the selected time period using Pearson correlation coefficients
5. THE Comparison_Engine SHALL support the following time period selections for all comparison modes: 3 months, 6 months, 1 year, 3 years, and 5 years
6. THE Comparison_Engine SHALL support sector-based stock grouping by accepting an ICB_Sector code and returning all stocks belonging to that sector from the Sector_Service, allowing the user to auto-populate the comparison list with stocks from a specific sector (e.g., "8355 Ngân hàng" for Banking)
7. THE MyFi_Platform frontend SHALL display a dedicated Stock Comparison view accessible from the main navigation with three tabs: Định giá (Valuation), Hiệu suất (Performance), and Tương quan (Correlation)
8. THE MyFi_Platform frontend SHALL display an interactive chart in the Valuation tab showing P/E and P/B ratio time-series lines for all selected stocks, with a time period selector (3M, 6M, 1Y, 3Y, 5Y)
9. THE MyFi_Platform frontend SHALL display an interactive chart in the Performance tab showing normalized percentage return lines for all selected stocks on a single chart, with a time period selector (3M, 6M, 1Y, 3Y, 5Y)
10. THE MyFi_Platform frontend SHALL display a color-coded Correlation_Matrix heatmap in the Correlation tab showing correlation coefficients between all pairs of selected stocks
11. THE MyFi_Platform frontend SHALL provide a stock selector that allows the user to add individual stocks by symbol (displaying company name alongside the symbol) and remove individual stocks from the comparison
12. THE MyFi_Platform frontend SHALL provide a sector/industry group dropdown that, when a sector is selected, auto-populates the comparison list with stocks from that sector (e.g., selecting "8355 Ngân hàng" adds all banking sector stocks)
13. THE MyFi_Platform frontend SHALL allow the user to hide/show individual stock lines on the Valuation and Performance charts without removing the stock from the comparison list
14. THE MyFi_Platform frontend SHALL provide a "Clear All" button that removes all selected stocks from the comparison
15. THE Comparison_Engine SHALL expose REST API endpoints: GET /api/comparison/valuation, GET /api/comparison/performance, and GET /api/comparison/correlation, each accepting a list of symbols and a time period parameter
16. WHEN the Analysis_Agent or Supervisor_Agent analyzes a stock, THE Analysis_Agent or Supervisor_Agent SHALL be able to request comparison data from the Comparison_Engine to provide peer comparison context in recommendations
17. IF fewer than 2 stocks are selected for comparison, THEN THE MyFi_Platform frontend SHALL display a prompt instructing the user to select at least 2 stocks
18. IF OHLCV data is unavailable for a selected stock over the requested time period, THEN THE Comparison_Engine SHALL exclude that stock from the comparison result and return a warning indicating which stocks were excluded and why

### Requirement 25: Watchlist Management

**User Story:** As an investor, I want to create multiple named watchlists with per-symbol price alerts and backend persistence, so that I can organize the symbols I track, get notified when prices hit my targets, and ensure the Monitor_Agent scans the symbols I care about.

#### Acceptance Criteria

1. THE Watchlist_Service SHALL allow the user to create, rename, and delete multiple named watchlists
2. THE Watchlist_Service SHALL allow the user to add, remove, and reorder symbols within a watchlist
3. WHEN a user sets a per-symbol price alert (above or below a specified VND threshold), THE Watchlist_Service SHALL persist the alert threshold in the database
4. WHEN the Price_Service reports a price that crosses a configured alert threshold for a watched symbol, THE Alert_Service SHALL deliver a notification to the user containing the symbol, triggered threshold, and current price
5. THE Watchlist_Service SHALL sync all watchlist symbols to the backend so the Monitor_Agent uses the union of all watchlist symbols as its scan list
6. THE Watchlist_Service SHALL persist all watchlists and their symbol orderings in the database
7. WHEN a new user accesses the platform for the first time, THE Watchlist_Service SHALL create a default watchlist named "My Watchlist" containing the symbols VNM, FPT, SSI, HPG, and MWG

### Requirement 26: Portfolio Performance Analytics

**User Story:** As an investor, I want to see time-weighted and money-weighted returns, an equity curve of my portfolio NAV over time, and benchmark comparisons against VN-Index and VN30, so that I can evaluate how well my portfolio is performing relative to the market.

#### Acceptance Criteria

1. THE Performance_Engine SHALL compute the time-weighted return (TWR) for the portfolio by chain-linking sub-period returns between each external cash flow event
2. THE Performance_Engine SHALL compute the money-weighted return (MWRR/XIRR) for the portfolio using the internal rate of return method applied to all cash flow dates and amounts
3. THE Performance_Engine SHALL store daily NAV snapshots in the database at market close (15:00 ICT) for constructing the historical equity curve
4. THE MyFi_Platform frontend SHALL display an equity curve chart showing portfolio NAV over time using the stored daily NAV snapshots
5. THE Performance_Engine SHALL fetch benchmark index data (VN-Index, VN30) via the Data_Source_Router and compute benchmark returns over the same periods as the portfolio for comparison
6. THE Performance_Engine SHALL compute performance breakdown by asset type (stocks, gold, crypto, savings) showing each asset type's contribution to total portfolio return
7. THE MyFi_Platform frontend SHALL support time period selection for performance views: 1W, 1M, 3M, 6M, 1Y, YTD, and ALL

### Requirement 27: Risk Metrics

**User Story:** As an investor, I want to see portfolio-level and per-holding risk metrics including Sharpe ratio, max drawdown, beta, volatility, and Value at Risk, so that I can understand and manage the risk profile of my investments.

#### Acceptance Criteria

1. THE Risk_Service SHALL compute the Sharpe ratio for the portfolio using the VN risk-free rate (current VN government bond yield or configurable default of 4.5% per annum)
2. THE Risk_Service SHALL compute the maximum drawdown from the NAV history as the largest peak-to-trough percentage decline
3. THE Risk_Service SHALL compute the portfolio beta against VN-Index by regressing daily portfolio returns against daily VN-Index returns over the trailing 1-year period
4. THE Risk_Service SHALL compute annualized volatility as the standard deviation of daily portfolio returns multiplied by the square root of 252 (trading days per year)
5. THE Risk_Service SHALL compute Value at Risk (VaR) at the 95% confidence level using the historical simulation method over the trailing 1-year period
6. THE MyFi_Platform frontend SHALL display risk metrics on the dashboard in a dedicated risk summary panel
7. WHEN the Analysis_Agent analyzes a stock, THE Analysis_Agent SHALL include per-stock risk metrics (volatility, beta, max drawdown) in the analysis output
8. THE Risk_Service SHALL compute per-holding risk contribution showing each holding's percentage contribution to total portfolio volatility

### Requirement 28: Export and Reporting

**User Story:** As an investor, I want to export transaction history, P&L reports, and portfolio snapshots to CSV and PDF, including a tax-friendly format for VN capital gains reporting, so that I can maintain records and file taxes accurately.

#### Acceptance Criteria

1. THE Export_Service SHALL export transaction history to CSV containing all Transaction_Ledger fields (asset type, symbol, quantity, unit price, total value, transaction date, transaction type)
2. THE Export_Service SHALL export a portfolio snapshot to CSV containing current holdings with symbol, quantity, average cost, current price, market value, unrealized P&L, and asset type
3. THE Export_Service SHALL export a portfolio report to PDF containing NAV summary, asset allocation chart, and P&L breakdown by holding
4. THE Export_Service SHALL generate a tax report showing realized capital gains and losses grouped by asset type and tax year, formatted for VN capital gains reporting
5. THE Export_Service SHALL support date range selection for all export types, filtering transactions and snapshots to the specified period
6. THE MyFi_Platform backend SHALL expose REST API endpoints for generating each export type: GET /api/export/transactions, GET /api/export/snapshot, GET /api/export/report, and GET /api/export/tax
7. THE MyFi_Platform frontend SHALL provide download buttons in the portfolio view and transaction history view for each supported export format

### Requirement 29: Multi-Currency Display

**User Story:** As an investor, I want to toggle the display currency between VND and USD with proper FX rate tracking, so that I can view my portfolio values in either currency and understand my USD-equivalent exposure.

#### Acceptance Criteria

1. THE FX_Service SHALL fetch the USD/VND exchange rate from CoinGecko using the USDT/VND trading pair, with a hardcoded fallback rate of 25,400 VND per USD when the API is unavailable
2. THE FX_Service SHALL cache the USD/VND rate with a 1-hour TTL
3. THE MyFi_Platform frontend SHALL provide a toggle control to switch the display currency between VND and USD
4. WHEN USD display is selected, THE MyFi_Platform frontend SHALL convert all VND-denominated values using the current FX rate from the FX_Service
5. THE Price_Service SHALL store crypto prices in both USD (native) and VND (converted) denominations
6. THE MyFi_Platform frontend SHALL display the current USD/VND exchange rate in the dashboard header
7. THE Export_Service SHALL include both VND and USD value columns in all exported CSV and PDF reports

### Requirement 30: Dividend and Corporate Action Tracking

**User Story:** As an investor, I want the platform to track ex-dividend dates, dividend payments, stock splits, and bonus shares, auto-adjust my cost basis accordingly, and show a calendar of upcoming corporate actions, so that I maintain accurate portfolio records and never miss important events.

#### Acceptance Criteria

1. THE Corporate_Action_Service SHALL fetch dividend calendar, stock split, and bonus share events from VCI/KBS via the Data_Source_Router
2. WHEN an ex-dividend date passes for a holding in the portfolio, THE Corporate_Action_Service SHALL automatically record the dividend payment as a transaction in the Transaction_Ledger
3. WHEN a stock split or bonus share event occurs for a holding, THE Corporate_Action_Service SHALL auto-adjust the cost basis and quantity in the Portfolio_Engine to reflect the new share count and adjusted per-share cost
4. THE MyFi_Platform frontend SHALL display an upcoming corporate actions calendar showing ex-dividend dates, payment dates, and AGM dates for all holdings
5. WHEN an ex-dividend date for a holding is 3 days away, THE Alert_Service SHALL notify the user with the symbol, ex-date, and expected dividend amount
6. WHEN the Analysis_Agent analyzes a stock, THE Analysis_Agent SHALL factor dividend yield and upcoming dividend events into the analysis and recommendations
7. THE Corporate_Action_Service SHALL track dividend history per holding to compute yield-on-cost (total dividends received divided by original cost basis)

### Requirement 31: Goal-Based Financial Planning

**User Story:** As an investor, I want to set financial goals with target amounts and dates, track my progress against current NAV, and have the Supervisor_Agent factor my goals into recommendations, so that I can plan and work toward specific financial milestones.

#### Acceptance Criteria

1. THE Goal_Planner SHALL allow the user to create, edit, and delete financial goals with the following fields: name, target amount (VND), target date, and associated asset types
2. THE Goal_Planner SHALL compute progress percentage for each goal as: current NAV of associated assets divided by target amount
3. THE Goal_Planner SHALL compute the required monthly contribution to reach the goal on time based on the remaining shortfall and months until the target date
4. THE MyFi_Platform frontend SHALL display goal progress on the dashboard with a progress bar and projected completion date
5. WHEN the Supervisor_Agent produces recommendations, THE Supervisor_Agent SHALL reference active goals and incorporate goal progress into advice (e.g., suggesting increased contributions when behind target)
6. THE Goal_Planner SHALL support goal categories: retirement, emergency fund, property, education, and custom
7. THE Goal_Planner SHALL persist all goals in the database

### Requirement 32: Backtesting and Strategy Simulation

**User Story:** As a trader, I want to define simple trading rules using technical indicators and backtest them against historical OHLCV data, so that I can evaluate strategy performance before risking real capital.

#### Acceptance Criteria

1. THE Backtest_Engine SHALL accept strategy rules defined as: entry condition (indicator-based, e.g., "RSI < 30 AND MACD crosses up"), exit condition, stop-loss percentage, and take-profit percentage
2. THE Backtest_Engine SHALL run the strategy against historical OHLCV data for a specified symbol and time range fetched via the Data_Source_Router
3. THE Backtest_Engine SHALL compute backtest results containing: total return, win rate, max drawdown, Sharpe ratio, number of trades, and average holding period
4. THE MyFi_Platform frontend SHALL display backtest results with an equity curve chart and trade entry/exit markers overlaid on the OHLCV chart
5. THE Backtest_Engine SHALL support backtesting with any of the 21 supported technical indicators (SMA, EMA, RSI, MACD, Bollinger Bands, Stochastic, ADX, Aroon, Parabolic SAR, Supertrend, VWAP, VWMA, Williams %R, CMO, ROC, Momentum, Keltner Channel, ATR, Standard Deviation, OBV, Linear Regression)
6. THE Backtest_Engine SHALL provide preset strategies including: "RSI Oversold Bounce" (entry: RSI < 30, exit: RSI > 70), "MACD Crossover" (entry: MACD line crosses above signal, exit: MACD line crosses below signal), and "Bollinger Band Squeeze" (entry: price touches lower band with narrowing bandwidth, exit: price touches upper band)
7. THE MyFi_Platform backend SHALL expose a REST API endpoint POST /api/backtest for running backtests, accepting strategy rules, symbol, and time range as input
8. THE MyFi_Platform frontend SHALL provide a UI for defining strategy rules (selecting indicators, setting conditions and thresholds), selecting symbol and timeframe, and viewing backtest results

### Requirement 33: API Rate Limiting and Quota Management

**User Story:** As a platform operator, I want centralized rate limiting in the Data_Source_Router to prevent upstream API blocking, with per-source limits and request queuing, so that the platform maintains reliable access to all external data sources.

#### Acceptance Criteria

1. THE Rate_Limiter SHALL enforce per-source request limits: CoinGecko maximum 10 requests per minute, VCI maximum 60 requests per minute, KBS maximum 60 requests per minute, and Doji maximum 30 requests per minute
2. WHEN a rate limit is reached for a source, THE Rate_Limiter SHALL queue pending requests and process them when the rate limit window resets
3. THE Rate_Limiter SHALL log each rate limit event with the source name, current queue depth, and estimated wait time
4. THE Data_Source_Router SHALL integrate the Rate_Limiter and check rate limits before dispatching any external API call
5. THE Rate_Limiter SHALL support configurable rate limits via environment variables or a configuration file, allowing per-source limits to be adjusted without code changes
6. THE Rate_Limiter SHALL expose metrics: current request count per source within the active window, queue depth per source, and total throttle events count
7. IF the queue depth for a source exceeds a configurable maximum (default 100), THEN THE Rate_Limiter SHALL reject new requests for that source with a "rate limited" error response

### Requirement 34: Offline and Degraded Mode

**User Story:** As an investor, I want the platform to handle internet unavailability gracefully by showing cached data with clear offline indicators, so that I can still review my portfolio and recent data when connectivity is lost.

#### Acceptance Criteria

1. THE MyFi_Platform frontend SHALL detect network connectivity status (online/offline) using browser navigator.onLine API and online/offline events
2. WHILE the frontend is offline, THE MyFi_Platform frontend SHALL display a persistent "Offline Mode" banner at the top of the page
3. WHILE the frontend is offline, THE MyFi_Platform frontend SHALL serve all data from local cache and last known API responses
4. WHEN network connectivity is restored, THE MyFi_Platform frontend SHALL automatically refresh all stale data from the backend
5. THE MyFi_Platform backend SHALL expose a health check endpoint (GET /api/health) that returns the connectivity status of each external data source (VCI, KBS, CoinGecko, Doji)
6. THE MyFi_Platform frontend SHALL display per-source health indicators showing which data sources are currently available and which are degraded or unavailable
7. THE MyFi_Platform frontend SHALL display the timestamp of the last successful data fetch alongside all cached data values

### Requirement 35: AI Recommendation Audit Trail

**User Story:** As an investor, I want every AI recommendation logged with full inputs and outputs, with outcome tracking fed back into the Knowledge_Base, so that I can review past advice, assess its accuracy, and the system improves over time.

#### Acceptance Criteria

1. THE Recommendation_Audit_Log SHALL persist every Supervisor_Agent recommendation with: timestamp, user query, sub-agent inputs (price data, analysis data, news data), final recommendation output, symbols involved, and recommended actions (buy, sell, hold)
2. THE Recommendation_Audit_Log SHALL track the outcome of each recommendation by recording the price changes of involved symbols at 1-day, 7-day, 14-day, and 30-day intervals after the recommendation timestamp
3. THE MyFi_Platform frontend SHALL display a recommendation history view showing past recommendations with their recorded outcomes and price changes
4. THE Recommendation_Audit_Log SHALL compute recommendation accuracy metrics: percentage of buy recommendations that resulted in positive returns at the 14-day mark, and percentage of sell recommendations that avoided losses at the 14-day mark
5. THE Knowledge_Base SHALL ingest recommendation outcomes from the Recommendation_Audit_Log to improve future Supervisor_Agent reasoning by providing historical accuracy context for similar recommendations
6. THE Recommendation_Audit_Log SHALL retain all audit records for a minimum of 2 years
7. THE MyFi_Platform backend SHALL expose a REST API endpoint GET /api/recommendations/history supporting filters by date range, symbol, and action type (buy, sell, hold)

### Requirement 36: User Authentication and Security

**User Story:** As a user, I want the platform to be protected by authentication with secure session management, so that my financial data and portfolio are accessible only to me.

#### Acceptance Criteria

1. THE Auth_Service SHALL support local authentication with username and password, storing passwords hashed with bcrypt
2. THE Auth_Service SHALL issue JWT tokens upon successful authentication with a configurable expiry period (default 24 hours)
3. THE MyFi_Platform backend SHALL require a valid JWT token for all REST API endpoints except GET /api/health and POST /api/auth/login
4. WHEN no valid session exists, THE MyFi_Platform frontend SHALL display a login screen and prevent access to any other views
5. THE MyFi_Platform frontend SHALL store the JWT in an httpOnly cookie rather than localStorage to prevent XSS-based token theft
6. THE Auth_Service SHALL support password change requiring the current password and a new password
7. IF 5 failed login attempts occur within 15 minutes for the same account, THEN THE Auth_Service SHALL lock the account for 30 minutes and return an error indicating the lockout duration
8. THE Auth_Service SHALL auto-expire sessions after a configurable inactivity period (default 4 hours) by tracking last activity timestamp per session
9. WHILE the MyFi_Platform is deployed in production, THE MyFi_Platform SHALL enforce HTTPS for all client-server communication


### Requirement 37: Dark Theme Support

**User Story:** As a user, I want to toggle between light and dark themes with persistent preference, so that I can use the platform comfortably in different lighting conditions and reduce eye strain during extended sessions.

#### Acceptance Criteria

1. THE Theme_Service SHALL manage theme state with two supported modes: light and dark
2. THE MyFi_Platform frontend SHALL provide a theme toggle button in the header or settings panel that switches between light and dark modes
3. WHEN the theme toggle is activated, THE Theme_Service SHALL update all UI components immediately without requiring a page reload
4. THE Theme_Service SHALL persist the user's theme preference in local storage
5. WHEN the user first accesses the platform without a saved preference, THE Theme_Service SHALL default to light mode
6. THE Theme_Service SHALL apply a dark theme color palette to all UI components including: navigation, cards, tables, forms, modals, buttons, text, and background colors
7. THE Chart_Engine SHALL adapt chart colors based on the current theme: light background with dark candles and indicators for light theme, dark background with light candles and indicators for dark theme
8. THE Theme_Service SHALL ensure sufficient contrast ratios in dark mode to maintain accessibility compliance (WCAG AA minimum 4.5:1 for normal text, 3:1 for large text)
9. WHEN the theme changes, THE Chart_Engine SHALL re-render all active charts with the theme-appropriate color scheme without losing chart state (zoom level, indicators, drawings)
10. THE Theme_Service SHALL apply theme-specific colors to data visualizations including pie charts, bar charts, heatmaps, and correlation matrices to ensure readability in both modes

### Requirement 38: Multi-Language Support

**User Story:** As a user, I want to switch between Vietnamese and English languages with persistent preference and locale-aware formatting, so that I can use the platform in my preferred language with properly formatted numbers, dates, and currency.

#### Acceptance Criteria

1. THE I18n_Service SHALL manage language state with two supported locales: Vietnamese (vi-VN) and English (en-US)
2. THE MyFi_Platform frontend SHALL provide a language selector dropdown in the header or settings panel with options: "Tiếng Việt" and "English"
3. WHEN the language selector is changed, THE I18n_Service SHALL update all UI text immediately without requiring a page reload
4. THE I18n_Service SHALL persist the user's language preference in local storage
5. WHEN the user first accesses the platform without a saved preference, THE I18n_Service SHALL default to Vietnamese (vi-VN)
6. THE I18n_Service SHALL translate all static UI text including: navigation labels, button text, form labels, error messages, validation messages, chart labels, table headers, and help text
7. THE I18n_Service SHALL format numbers according to the selected locale: Vietnamese uses period (.) for thousands separator and comma (,) for decimal separator; English uses comma (,) for thousands separator and period (.) for decimal separator
8. THE I18n_Service SHALL format dates according to the selected locale: Vietnamese uses dd/MM/yyyy format; English uses MM/dd/yyyy format
9. THE I18n_Service SHALL format currency display based on the selected language: Vietnamese displays "₫" or "VND" suffix; English displays "VND" prefix or suffix based on context
10. THE I18n_Service SHALL format time according to the selected locale: Vietnamese uses 24-hour format (HH:mm); English uses 12-hour format with AM/PM (hh:mm AM/PM)
11. THE I18n_Service SHALL provide translation strings for all AI agent responses, alert messages, and notification text generated by the Multi_Agent_System
12. THE I18n_Service SHALL translate sector names, asset type labels, and financial term labels (P/E, P/B, ROE, etc.) according to the selected language
13. THE I18n_Service SHALL support dynamic text interpolation for translated strings containing variable values (e.g., "Portfolio value: {amount}" where {amount} is formatted according to locale)
14. WHEN the Chart_Engine renders charts, THE Chart_Engine SHALL use locale-appropriate number formatting for axis labels, tooltips, and legend values based on the current I18n_Service language setting
### Requirement 39: Mobile Responsiveness and Progressive Web App

**User Story:** As a mobile user, I want the platform to work seamlessly on my smartphone with touch-optimized interactions and offline capabilities, so that I can manage my portfolio on the go.

#### Acceptance Criteria

1. THE MyFi_Platform frontend SHALL implement responsive design with breakpoints for mobile (< 768px), tablet (768px - 1024px), and desktop (> 1024px) viewports
2. THE MyFi_Platform frontend SHALL use mobile-first CSS approach with progressive enhancement for larger screens
3. THE Chart_Engine SHALL support touch gestures including: pinch-to-zoom, two-finger pan, tap to show crosshair, and long-press for context menu
4. THE MyFi_Platform frontend SHALL implement a mobile-optimized navigation pattern using a hamburger menu or bottom navigation bar on mobile devices
5. THE MyFi_Platform frontend SHALL register a service worker to enable Progressive Web App (PWA) capabilities
6. THE MyFi_Platform frontend SHALL provide a web app manifest file enabling "Add to Home Screen" functionality on iOS and Android
7. WHEN installed as a PWA, THE MyFi_Platform SHALL cache critical assets (HTML, CSS, JS, fonts) for offline access
8. THE MyFi_Platform frontend SHALL optimize touch target sizes to minimum 44x44 pixels for all interactive elements on mobile
9. THE MyFi_Platform frontend SHALL implement swipe gestures for navigation between tabs and dismissing modals on mobile
10. THE Chart_Engine SHALL adapt chart controls for mobile with larger touch-friendly buttons and simplified indicator configuration panels
11. THE MyFi_Platform frontend SHALL use viewport-relative units and flexible layouts to ensure content fits without horizontal scrolling on all screen sizes
12. THE MyFi_Platform frontend SHALL lazy-load images and heavy components to improve mobile performance on slower networks

### Requirement 40: Real-Time WebSocket Price Streaming

**User Story:** As an investor, I want real-time price updates pushed to my browser via WebSocket instead of polling, so that I see instant price changes with lower latency and reduced server load.

#### Acceptance Criteria

1. THE MyFi_Platform backend SHALL implement a WebSocket server endpoint at /ws/prices for real-time price streaming
2. WHEN a client connects to the WebSocket endpoint, THE MyFi_Platform backend SHALL authenticate the connection using the JWT token passed in the connection handshake
3. THE MyFi_Platform frontend SHALL establish a WebSocket connection on dashboard load and subscribe to price updates for all symbols in the user's portfolio and watchlists
4. THE Price_Service SHALL publish price updates to all connected WebSocket clients whenever new price data is fetched from upstream sources
5. THE MyFi_Platform backend SHALL support subscription management allowing clients to subscribe and unsubscribe from specific symbols dynamically
6. WHEN a price update is received via WebSocket, THE MyFi_Platform frontend SHALL update the UI immediately without waiting for the next polling interval
7. THE MyFi_Platform backend SHALL implement WebSocket heartbeat/ping-pong to detect and close stale connections
8. IF the WebSocket connection is lost, THEN THE MyFi_Platform frontend SHALL automatically attempt to reconnect with exponential backoff (1s, 2s, 4s, 8s, max 30s)
9. THE MyFi_Platform frontend SHALL fall back to HTTP polling if WebSocket connection fails after 5 reconnection attempts
10. THE MyFi_Platform backend SHALL limit each WebSocket connection to subscribing to a maximum of 100 symbols to prevent resource exhaustion
11. THE MyFi_Platform backend SHALL broadcast market status changes (market open, market close, trading halt) to all connected clients via WebSocket

### Requirement 41: Push Notifications

**User Story:** As an investor, I want to receive push notifications on my browser and mobile device for critical alerts, so that I never miss important market events even when the app is not open.

#### Acceptance Criteria

1. THE MyFi_Platform frontend SHALL request push notification permission from the user on first login or when enabling alerts
2. THE MyFi_Platform frontend SHALL register for push notifications using the Web Push API and store the push subscription in the backend
3. THE Alert_Service SHALL send push notifications for high-priority alerts including: price alerts triggered, Monitor_Agent pattern detections with confidence > 80, and upcoming ex-dividend dates within 24 hours
4. THE MyFi_Platform backend SHALL implement Web Push protocol to send notifications to subscribed clients even when the browser is closed
5. THE MyFi_Platform frontend SHALL display push notifications with: title, body text, icon, and action buttons (View, Dismiss)
6. WHEN a user clicks a push notification, THE MyFi_Platform frontend SHALL open the app and navigate to the relevant view (chart for price alerts, alert detail for pattern detections)
7. THE MyFi_Platform frontend SHALL allow users to configure notification preferences per alert type: push, in-app, or disabled
8. THE Alert_Service SHALL respect quiet hours configured by the user (default: no push notifications between 22:00 and 07:00 local time)
9. THE Alert_Service SHALL implement notification grouping to prevent spam: maximum 5 push notifications per hour, with additional alerts queued and sent as a summary
10. THE MyFi_Platform backend SHALL clean up expired push subscriptions (failed delivery for 7 consecutive days) from the database

### Requirement 42: Email and SMS Alerts

**User Story:** As an investor, I want to receive critical alerts via email and SMS in addition to in-app notifications, so that I can stay informed through my preferred communication channels.

#### Acceptance Criteria

1. THE Alert_Service SHALL support three delivery channels: in-app, email, and SMS, with per-alert-type configuration
2. THE MyFi_Platform frontend SHALL allow users to configure email and SMS preferences in the settings panel, including: enabled channels per alert type, email address, and phone number
3. THE Alert_Service SHALL send email alerts using an SMTP service or email API (SendGrid, AWS SES, or similar) for high-priority events
4. THE Alert_Service SHALL send SMS alerts using an SMS gateway API (Twilio, AWS SNS, or similar) for critical events only
5. THE Alert_Service SHALL define alert priority levels: critical (price alerts, large portfolio losses > 5%), high (Monitor_Agent patterns > 80 confidence), medium (ex-dividend reminders), and low (general notifications)
6. THE Alert_Service SHALL send SMS only for critical-priority alerts to minimize SMS costs
7. THE Alert_Service SHALL send email for critical and high-priority alerts with HTML formatting including: alert summary, relevant data, and a link to view details in the app
8. THE Alert_Service SHALL implement rate limiting for email (max 20 per day) and SMS (max 5 per day) per user to prevent notification fatigue and control costs
9. THE Alert_Service SHALL allow users to configure quiet hours separately for each channel (e.g., no SMS between 22:00-07:00, but email allowed anytime)
10. THE Alert_Service SHALL provide an unsubscribe link in all email alerts allowing users to disable email notifications for specific alert types

### Requirement 43: Social Features and Community

**User Story:** As an investor, I want to share my portfolio performance anonymously with the community, follow other investors' public watchlists, and see aggregated sentiment, so that I can learn from the community and benchmark my performance.

#### Acceptance Criteria

1. THE MyFi_Platform backend SHALL allow users to opt-in to sharing their portfolio performance anonymously with the community
2. WHEN a user enables public sharing, THE MyFi_Platform backend SHALL publish anonymized performance metrics (total return %, Sharpe ratio, max drawdown, asset allocation percentages) without revealing absolute NAV values or specific holdings
3. THE MyFi_Platform frontend SHALL display a community leaderboard showing top-performing anonymized portfolios ranked by total return, Sharpe ratio, and risk-adjusted return over selectable time periods (1M, 3M, 6M, 1Y)
4. THE MyFi_Platform backend SHALL allow users to publish watchlists as public with a shareable link
5. THE MyFi_Platform frontend SHALL allow users to follow public watchlists from other users, with followed watchlists appearing in a "Community Watchlists" section
6. THE MyFi_Platform backend SHALL track the number of followers for each public watchlist and display popularity metrics
7. THE MyFi_Platform backend SHALL aggregate sentiment data from public watchlists by counting how many users have added each symbol to their watchlists in the past 7 days
8. THE MyFi_Platform frontend SHALL display community sentiment indicators on stock detail pages showing: number of users watching, trending status (increasing/decreasing watchlist adds), and sentiment score (bullish/bearish based on watchlist activity)
9. THE MyFi_Platform backend SHALL allow users to compare their portfolio performance against anonymized peer benchmarks grouped by portfolio size ranges (< 100M VND, 100M-500M, 500M-1B, > 1B)
10. THE MyFi_Platform frontend SHALL display peer comparison charts showing the user's performance percentile within their portfolio size cohort
11. THE MyFi_Platform backend SHALL implement privacy controls allowing users to hide their profile from leaderboards while still accessing community features
12. THE MyFi_Platform backend SHALL moderate public watchlists and comments to prevent spam and inappropriate content


### Requirement 44: Transaction Import and Broker Integration

**User Story:** As an investor, I want to import my transaction history from CSV files exported by my broker or sync directly with broker APIs, so that I can avoid manual data entry and ensure my portfolio is always up-to-date.

#### Acceptance Criteria

1. THE MyFi_Platform backend SHALL support CSV import for transaction history with a flexible column mapping interface
2. THE MyFi_Platform backend SHALL provide preset CSV templates for major VN brokers: SSI, VPS, HSC, VCBS, Vietcombank Securities, and MB Securities
3. WHEN a user uploads a CSV file, THE MyFi_Platform backend SHALL parse the file, detect the broker format automatically if possible, and display a preview of parsed transactions before import
4. THE MyFi_Platform backend SHALL validate imported transactions for: required fields (symbol, quantity, price, date, type), data type correctness, and logical consistency (sell quantity not exceeding holdings)
5. THE MyFi_Platform backend SHALL support importing the following transaction types from CSV: buy, sell, dividend received, stock split, bonus shares, rights issue, and cash deposit/withdrawal
6. THE MyFi_Platform frontend SHALL provide a reconciliation view showing imported transactions alongside existing transactions, with conflict detection and resolution options (skip, overwrite, merge)
7. THE MyFi_Platform backend SHALL support broker API integration for automated transaction sync where APIs are available (SSI iBoard API, VPS API if accessible)
8. WHEN broker API integration is enabled, THE MyFi_Platform backend SHALL sync transactions daily at a configurable time (default: 16:00 ICT after market close)
9. THE MyFi_Platform backend SHALL store the last sync timestamp per broker connection and display sync status in the frontend
10. THE MyFi_Platform backend SHALL support bank statement parsing for cash flow transactions, detecting deposits and withdrawals from common VN bank statement formats (Vietcombank, Techcombank, VPBank)
11. THE MyFi_Platform frontend SHALL provide a transaction import history log showing all import operations with: timestamp, source (CSV file name or broker API), number of transactions imported, and any errors or warnings
12. THE MyFi_Platform backend SHALL implement duplicate detection to prevent importing the same transaction multiple times based on: symbol, date, quantity, price, and transaction type matching

### Requirement 45: Portfolio Rebalancing Tools

**User Story:** As an investor, I want to see my target allocation vs actual allocation with automatic rebalancing suggestions, so that I can maintain my desired portfolio balance and optimize for my investment strategy.

#### Acceptance Criteria

1. THE MyFi_Platform backend SHALL allow users to define target allocation percentages for asset types (stocks, gold, crypto, savings, bonds, cash) and for individual holdings or sectors
2. THE MyFi_Platform frontend SHALL display a target vs actual allocation comparison chart showing deviations from target percentages
3. THE Portfolio_Engine SHALL compute rebalancing suggestions when actual allocation deviates from target by more than a configurable threshold (default: 5 percentage points)
4. WHEN rebalancing is needed, THE Portfolio_Engine SHALL generate specific buy/sell recommendations with quantities and estimated costs to bring the portfolio back to target allocation
5. THE MyFi_Platform frontend SHALL provide a rebalancing simulator allowing users to preview the portfolio state after executing suggested trades before committing
6. THE Portfolio_Engine SHALL support multiple rebalancing strategies: threshold-based (rebalance when deviation exceeds threshold), calendar-based (rebalance monthly/quarterly), and opportunistic (rebalance when market conditions favor it)
7. THE Portfolio_Engine SHALL factor in transaction costs (broker fees, taxes) when computing rebalancing recommendations to ensure rebalancing is cost-effective
8. THE MyFi_Platform backend SHALL support tax-loss harvesting by identifying holdings with unrealized losses that can be sold to offset capital gains for tax purposes
9. WHEN tax-loss harvesting opportunities are detected, THE Portfolio_Engine SHALL suggest selling losing positions and optionally replacing them with similar assets to maintain allocation
10. THE MyFi_Platform frontend SHALL display the estimated tax savings from tax-loss harvesting recommendations
11. THE Portfolio_Engine SHALL track rebalancing history showing past rebalancing actions, their outcomes, and performance impact
12. THE MyFi_Platform backend SHALL allow users to set rebalancing constraints such as: minimum trade size, maximum number of trades per rebalancing, and excluded holdings (do not sell)

### Requirement 46: Voice Input and Natural Language Queries

**User Story:** As a user, I want to interact with the AI chat using voice input in Vietnamese and ask natural language questions about my portfolio, so that I can get information hands-free and in a conversational manner.

#### Acceptance Criteria

1. THE MyFi_Platform frontend SHALL integrate Web Speech API for voice input with Vietnamese language support (vi-VN)
2. THE MyFi_Platform frontend SHALL provide a microphone button in the chat widget that activates voice recording when pressed
3. WHEN voice recording is active, THE MyFi_Platform frontend SHALL display a visual indicator (animated waveform or pulsing icon) and transcribe speech to text in real-time
4. THE MyFi_Platform frontend SHALL support both push-to-talk (hold button to record) and toggle mode (click to start, click to stop) for voice input
5. THE Supervisor_Agent SHALL parse natural language portfolio queries such as: "How much did I make on tech stocks this quarter?", "What's my best performing asset?", "Show me my dividend income this year"
6. THE Supervisor_Agent SHALL extract intent and entities from natural language queries including: time periods (this quarter, last month, YTD), asset types (tech stocks, gold, crypto), metrics (profit, return, dividend), and actions (show, calculate, compare)
7. THE Supervisor_Agent SHALL query the Portfolio_Engine and Transaction_Ledger to answer natural language questions with specific data from the user's portfolio
8. THE Supervisor_Agent SHALL format responses in natural language with conversational tone matching the user's language preference (Vietnamese or English)
9. THE MyFi_Platform frontend SHALL support voice output (text-to-speech) for AI responses, allowing users to hear answers hands-free
10. THE MyFi_Platform frontend SHALL provide voice command shortcuts for common actions: "Show my portfolio", "Check VNM price", "What's the market doing today", "Set alert for FPT at 100,000"
11. THE Supervisor_Agent SHALL handle ambiguous queries by asking clarifying questions (e.g., "Which quarter did you mean - Q1 2025 or Q1 2026?")
12. THE MyFi_Platform backend SHALL log voice interactions for quality improvement and error analysis while respecting user privacy

### Requirement 47: Advanced Tax Optimization and Reporting

**User Story:** As a taxpayer, I want comprehensive tax optimization suggestions and automated tax report generation for VN securities trading, so that I can minimize my tax liability and file accurately.

#### Acceptance Criteria

1. THE MyFi_Platform backend SHALL compute VN personal income tax liability for securities trading based on current tax regulations (0.1% on sell value for stocks, capital gains tax for other assets)
2. THE Export_Service SHALL generate an annual tax report for VN securities trading containing: total sell value, total tax paid (0.1% withholding), realized capital gains/losses by asset type, and dividend income received
3. THE Portfolio_Engine SHALL identify tax-loss harvesting opportunities by finding holdings with unrealized losses that can offset realized gains
4. THE Portfolio_Engine SHALL compute the optimal timing for selling winning positions to minimize tax impact, considering: holding period, current year realized gains, and projected future gains
5. THE MyFi_Platform frontend SHALL display a tax dashboard showing: YTD tax paid, projected year-end tax liability, available tax-loss harvesting opportunities, and tax-efficient withdrawal strategies
6. THE Portfolio_Engine SHALL suggest tax-efficient withdrawal strategies for users needing to liquidate positions, prioritizing: holdings with losses first, then long-term holdings, then low-gain positions
7. THE Export_Service SHALL generate tax reports in formats suitable for VN tax filing including: detailed transaction ledger, capital gains summary by asset type, and dividend income summary
8. THE MyFi_Platform backend SHALL track cost basis adjustments for corporate actions (stock splits, bonus shares, rights issues) to ensure accurate capital gains calculations
9. THE Portfolio_Engine SHALL compute wash sale violations (selling at a loss and repurchasing within 30 days) and flag them in tax reports
10. THE MyFi_Platform frontend SHALL provide tax scenario modeling allowing users to simulate the tax impact of planned trades before execution
11. THE Export_Service SHALL support exporting tax reports to PDF with Vietnamese language formatting suitable for submission to VN tax authorities
12. THE MyFi_Platform backend SHALL maintain a tax lot tracking system using FIFO, LIFO, or specific identification methods as selected by the user


### Requirement 48: Educational Content and Onboarding

**User Story:** As a new investor, I want interactive tutorials, a financial glossary, and investment strategy templates, so that I can learn how to use the platform and improve my investment knowledge.

#### Acceptance Criteria

1. THE MyFi_Platform frontend SHALL provide an interactive onboarding flow for new users covering: account setup, adding first asset, understanding the dashboard, and using the AI chat
2. THE MyFi_Platform frontend SHALL implement a step-by-step tutorial system using tooltips and guided tours for key features: portfolio tracking, chart analysis, screener, and AI recommendations
3. THE MyFi_Platform frontend SHALL provide a searchable glossary of financial terms in both Vietnamese and English covering: asset types, technical indicators, fundamental metrics, and VN market terminology
4. THE MyFi_Platform frontend SHALL display contextual help tooltips on complex UI elements explaining: what the metric means, how it's calculated, and how to interpret it
5. THE MyFi_Platform backend SHALL provide investment strategy templates including: value investing, growth investing, dividend investing, index tracking, and sector rotation, each with predefined screener filters and allocation targets
6. THE MyFi_Platform frontend SHALL provide video tutorials or animated guides for: setting up portfolio tracking, using the chart engine, interpreting AI recommendations, and understanding risk metrics
7. THE MyFi_Platform frontend SHALL implement a "Learn" section with educational articles covering: VN stock market basics, technical analysis fundamentals, fundamental analysis, portfolio management, and risk management
8. THE MyFi_Platform frontend SHALL provide example portfolios (demo data) that new users can explore before adding their own assets
9. THE MyFi_Platform frontend SHALL implement progressive disclosure, showing basic features first and gradually introducing advanced features as users gain experience
10. THE MyFi_Platform frontend SHALL track user progress through tutorials and educational content, displaying completion badges and unlocking advanced features as milestones are reached
11. THE MyFi_Platform frontend SHALL provide a "Tips & Tricks" section with practical advice for: optimizing portfolio allocation, using technical indicators effectively, and interpreting AI recommendations
12. THE MyFi_Platform frontend SHALL allow users to replay tutorials at any time from the help menu

### Requirement 49: Data Quality and Validation

**User Story:** As an investor, I want the platform to detect and alert me about data anomalies and quality issues, so that I can trust the accuracy of my portfolio valuations and market data.

#### Acceptance Criteria

1. THE Data_Source_Router SHALL implement data anomaly detection for price data, flagging: sudden price jumps > 20% without corresponding volume increase, prices outside ceiling/floor range, and zero or negative prices
2. WHEN a data anomaly is detected, THE Data_Source_Router SHALL log the anomaly with details (symbol, anomaly type, detected value, expected range) and attempt to fetch from alternative source
3. THE MyFi_Platform frontend SHALL display data quality indicators on the dashboard showing: data source reliability score (percentage of successful fetches in past 24 hours), last successful update timestamp, and any active data quality warnings
4. THE Price_Service SHALL validate OHLCV data for logical consistency: high >= low, high >= open, high >= close, low <= open, low <= close, and volume >= 0
5. THE Price_Service SHALL detect and flag gaps in time series data (missing trading days) and interpolate or mark as unavailable
6. THE MyFi_Platform backend SHALL maintain a data source reliability scoring system tracking: uptime percentage, average response time, error rate, and data completeness for each source (VCI, KBS, CoinGecko, Doji)
7. THE MyFi_Platform frontend SHALL allow users to report data issues directly from the UI with a "Report Data Issue" button on stock detail pages
8. THE MyFi_Platform backend SHALL track user-reported data issues in a database with: symbol, issue type (wrong price, missing data, incorrect fundamentals), reporter, timestamp, and resolution status
9. THE Data_Source_Router SHALL implement historical data accuracy verification by comparing current data against previously cached values and flagging significant discrepancies
10. THE MyFi_Platform backend SHALL generate daily data quality reports for administrators showing: anomaly count by source, user-reported issues, data gaps, and source reliability scores
11. THE MyFi_Platform frontend SHALL display a data quality badge on each price display indicating: verified (green), unverified (yellow), or anomaly detected (red)
12. THE Price_Service SHALL implement cross-source validation for critical data, fetching the same data point from multiple sources and flagging discrepancies > 5%

### Requirement 50: Performance Optimization and Caching

**User Story:** As a user, I want the platform to load quickly and respond instantly to interactions, so that I can efficiently manage my portfolio without waiting for slow page loads or API calls.

#### Acceptance Criteria

1. THE MyFi_Platform frontend SHALL implement code splitting to load only the JavaScript required for the current page, with lazy loading for routes and heavy components
2. THE MyFi_Platform frontend SHALL use a service worker to cache static assets (JS, CSS, fonts, images) with a cache-first strategy and background updates
3. THE MyFi_Platform frontend SHALL implement optimistic UI updates, showing expected results immediately while API calls complete in the background
4. THE MyFi_Platform backend SHALL implement Redis or in-memory caching for frequently accessed data with appropriate TTLs: price data (15 min), sector data (30 min), company info (6 hours)
5. THE MyFi_Platform backend SHALL implement database query optimization with proper indexing on: user_id, symbol, transaction_date, and asset_type columns
6. THE MyFi_Platform backend SHALL use database connection pooling to reuse connections and reduce connection overhead
7. THE MyFi_Platform frontend SHALL implement virtual scrolling for long lists (transaction history, screener results) to render only visible items
8. THE Chart_Engine SHALL implement canvas-based rendering for charts instead of SVG to improve performance with large datasets
9. THE MyFi_Platform backend SHALL implement API response compression using gzip or brotli to reduce payload sizes
10. THE MyFi_Platform frontend SHALL use image optimization techniques: lazy loading, responsive images with srcset, and WebP format with fallbacks
11. THE MyFi_Platform backend SHALL implement database read replicas for read-heavy operations (portfolio queries, historical data) to distribute load
12. THE MyFi_Platform frontend SHALL implement request deduplication to prevent multiple identical API calls from being sent simultaneously
13. THE MyFi_Platform backend SHALL use CDN for serving static assets with edge caching to reduce latency for users across different regions
14. THE MyFi_Platform frontend SHALL implement skeleton screens and progressive loading to show content structure immediately while data loads

### Requirement 51: Compliance and Audit Trail

**User Story:** As a platform operator, I want comprehensive activity logging and audit trails for all user actions, so that I can ensure compliance, investigate issues, and maintain data integrity.

#### Acceptance Criteria

1. THE MyFi_Platform backend SHALL log all user actions in an audit trail including: login/logout, portfolio changes (add/edit/delete assets), transaction creation/modification, settings changes, and data exports
2. THE audit trail SHALL record for each action: user ID, action type, timestamp, IP address, user agent, affected resources (asset IDs, transaction IDs), old values, new values, and action result (success/failure)
3. THE MyFi_Platform backend SHALL implement immutable audit logs that cannot be modified or deleted by users or administrators
4. THE MyFi_Platform backend SHALL retain audit logs for a configurable period (default: 7 years) to meet financial record-keeping requirements
5. THE MyFi_Platform backend SHALL provide an audit log query interface for administrators with filters by: user, date range, action type, and resource
6. THE MyFi_Platform backend SHALL implement data retention policies allowing users to request data deletion after account closure, while preserving audit logs for compliance
7. THE MyFi_Platform backend SHALL support GDPR-style data export allowing users to download all their data in machine-readable format (JSON)
8. THE MyFi_Platform backend SHALL implement data anonymization for deleted accounts, removing PII while preserving anonymized data for analytics
9. THE MyFi_Platform backend SHALL log all API calls with: endpoint, method, request parameters, response status, response time, and user ID
10. THE MyFi_Platform backend SHALL implement security event logging for: failed login attempts, password changes, session expirations, and suspicious activity (multiple failed logins, unusual access patterns)
11. THE MyFi_Platform backend SHALL provide compliance reports showing: total users, active users, data storage size, API usage statistics, and security events
12. THE MyFi_Platform backend SHALL implement automated backup of audit logs to separate storage with encryption at rest

### Requirement 52: Advanced Charting Features

**User Story:** As a trader, I want to save chart templates, use multi-timeframe analysis, replay historical trading sessions, and get pattern recognition alerts, so that I can perform sophisticated technical analysis.

#### Acceptance Criteria

1. THE Chart_Engine SHALL allow users to save chart templates containing: selected indicators with parameters, drawing tools, time interval, and chart style (candlestick, line, bar)
2. THE Chart_Engine SHALL allow users to name, save, and load chart templates, with templates persisted in the backend database
3. THE Chart_Engine SHALL support multi-timeframe (MTF) indicators, displaying indicator values from higher timeframes on lower timeframe charts (e.g., daily SMA on 1-hour chart)
4. THE Chart_Engine SHALL provide a chart replay mode allowing users to step through historical price action bar-by-bar to simulate live trading conditions
5. WHEN chart replay mode is active, THE Chart_Engine SHALL hide future price data and allow users to place simulated trades to practice strategy execution
6. THE Chart_Engine SHALL implement automated pattern recognition for common chart patterns: head and shoulders, inverse head and shoulders, double top, double bottom, ascending triangle, descending triangle, symmetrical triangle, flag, and pennant
7. WHEN a chart pattern is detected, THE Chart_Engine SHALL draw the pattern on the chart with labels and provide a pattern analysis (bullish/bearish, target price, stop loss level)
8. THE Alert_Service SHALL allow users to set pattern recognition alerts, receiving notifications when specific patterns are detected on watched symbols
9. THE Chart_Engine SHALL implement volume profile analysis showing volume distribution at different price levels as a horizontal histogram overlaid on the chart
10. THE Chart_Engine SHALL support custom time ranges allowing users to select any arbitrary start and end date for chart display
11. THE Chart_Engine SHALL provide a comparison mode overlaying multiple symbols on the same chart with normalized percentage scale
12. THE Chart_Engine SHALL implement chart annotations allowing users to add text notes, arrows, and shapes to mark important levels or events
13. THE Chart_Engine SHALL sync chart state across devices for logged-in users, preserving: active indicators, drawings, annotations, and zoom level
14. THE Chart_Engine SHALL provide keyboard shortcuts for common actions: add indicator (I), draw trend line (T), toggle crosshair (C), zoom in/out (+/-), and reset zoom (R)


### Requirement 53: Portfolio Stress Testing and Scenario Analysis

**User Story:** As an investor, I want to stress test my portfolio against historical crises and custom scenarios, so that I can understand how my portfolio would perform under adverse conditions and prepare accordingly.

#### Acceptance Criteria

1. THE Risk_Service SHALL implement historical stress testing by replaying portfolio performance during past market crises: 2008 global financial crisis, 2020 COVID crash, and VN-specific events (2011 banking crisis)
2. WHEN historical stress testing is run, THE Risk_Service SHALL compute: portfolio drawdown during the crisis period, recovery time to pre-crisis NAV, and comparison against VN-Index performance during the same period
3. THE Risk_Service SHALL support custom scenario analysis allowing users to define hypothetical market conditions: percentage change in VN-Index, sector-specific shocks, currency devaluation, and interest rate changes
4. THE Risk_Service SHALL compute portfolio impact for custom scenarios showing: projected NAV change, affected holdings, and risk contribution by asset type
5. THE MyFi_Platform frontend SHALL display stress test results with visualizations: NAV drawdown chart, recovery timeline, and worst-case scenario projections
6. THE Risk_Service SHALL implement Monte Carlo simulation for portfolio projections, running 10,000 simulations with randomized returns based on historical volatility
7. THE MyFi_Platform frontend SHALL display Monte Carlo results showing: probability distribution of future NAV, confidence intervals (10th, 50th, 90th percentile), and probability of reaching financial goals
8. THE Risk_Service SHALL compute correlation breakdown scenarios showing how portfolio risk changes if correlations between assets increase during market stress
9. THE Risk_Service SHALL identify portfolio concentration risks by flagging: single holdings exceeding 20% of NAV, sector concentration exceeding 40%, and asset type concentration exceeding 60%
10. THE MyFi_Platform frontend SHALL provide a "What-If" calculator allowing users to simulate adding or removing positions and see the impact on portfolio risk metrics
11. THE Risk_Service SHALL compute tail risk metrics including: conditional Value at Risk (CVaR), expected shortfall, and maximum loss at 99% confidence level
12. THE MyFi_Platform frontend SHALL display stress test recommendations suggesting portfolio adjustments to improve resilience: diversification suggestions, hedging strategies, and safe-haven asset allocation

### Requirement 54: Public API and Webhooks

**User Story:** As a developer, I want access to a public API and webhook system, so that I can integrate MyFi data with external tools and receive real-time notifications for custom workflows.

#### Acceptance Criteria

1. THE MyFi_Platform backend SHALL provide a public REST API with endpoints for: portfolio data, transaction history, price data, market data, and AI recommendations
2. THE MyFi_Platform backend SHALL implement API key authentication for public API access with rate limiting per API key
3. THE MyFi_Platform backend SHALL provide API documentation using OpenAPI/Swagger specification with interactive API explorer
4. THE MyFi_Platform backend SHALL implement webhook support allowing users to register webhook URLs for specific events: price alerts triggered, portfolio NAV milestones reached, and AI pattern detections
5. WHEN a webhook event occurs, THE MyFi_Platform backend SHALL send an HTTP POST request to the registered webhook URL with event payload in JSON format
6. THE MyFi_Platform backend SHALL implement webhook retry logic with exponential backoff for failed deliveries (3 retries: 1s, 5s, 15s)
7. THE MyFi_Platform backend SHALL provide webhook signature verification using HMAC-SHA256 to ensure webhook authenticity
8. THE MyFi_Platform frontend SHALL provide a webhook management UI allowing users to: register webhooks, test webhooks, view delivery logs, and disable webhooks
9. THE MyFi_Platform backend SHALL implement API rate limiting: 100 requests per minute for authenticated users, 10 requests per minute for unauthenticated access
10. THE MyFi_Platform backend SHALL provide API usage analytics showing: request count, error rate, most-used endpoints, and rate limit violations
11. THE MyFi_Platform backend SHALL support API versioning (v1, v2) to maintain backward compatibility when introducing breaking changes
12. THE MyFi_Platform backend SHALL provide SDKs or client libraries for popular languages (JavaScript, Python, Go) to simplify API integration

### Requirement 55: Integration with External Services

**User Story:** As a user, I want to export my portfolio data to popular portfolio trackers and integrate with tax software, so that I can use MyFi alongside my existing financial tools.

#### Acceptance Criteria

1. THE Export_Service SHALL support exporting portfolio data in formats compatible with popular portfolio trackers: Yahoo Finance CSV, Google Sheets format, and Personal Capital format
2. THE Export_Service SHALL provide one-click export to Google Sheets with automatic sheet creation and data population using Google Sheets API
3. THE MyFi_Platform backend SHALL support OAuth integration with Google account for Google Sheets export and Google Drive backup
4. THE Export_Service SHALL generate tax reports compatible with VN tax software formats if available, or provide detailed CSV suitable for manual entry
5. THE MyFi_Platform backend SHALL support exporting transaction history to accounting software formats: QuickBooks CSV, Xero format, and generic accounting CSV
6. THE MyFi_Platform frontend SHALL provide a "Connect" section in settings showing available integrations with status indicators (connected, not connected, error)
7. THE MyFi_Platform backend SHALL implement scheduled exports allowing users to automatically export portfolio snapshots to Google Drive or Dropbox on a daily, weekly, or monthly schedule
8. THE MyFi_Platform backend SHALL support importing watchlists from external sources: TradingView watchlist export, Yahoo Finance portfolio CSV, and generic symbol lists
9. THE Export_Service SHALL provide portfolio performance reports in formats suitable for sharing with financial advisors: PDF with charts, Excel with detailed breakdowns
10. THE MyFi_Platform backend SHALL support calendar integration (Google Calendar, Outlook) for syncing corporate action events, ex-dividend dates, and earnings dates
11. THE MyFi_Platform backend SHALL implement Zapier integration allowing users to create custom automation workflows triggered by MyFi events
12. THE MyFi_Platform backend SHALL provide IFTTT integration for simple automation rules: "If portfolio NAV increases by 10%, then send email"

### Requirement 56: User Experience Enhancements

**User Story:** As a power user, I want keyboard shortcuts, customizable dashboard layouts, a command palette, and undo/redo functionality, so that I can work efficiently and personalize my workflow.

#### Acceptance Criteria

1. THE MyFi_Platform frontend SHALL implement a command palette (Cmd+K or Ctrl+K) providing quick access to all features, navigation, and actions with fuzzy search
2. THE command palette SHALL support action execution including: navigate to pages, add transaction, create alert, search symbols, and change settings
3. THE MyFi_Platform frontend SHALL provide keyboard shortcuts for common actions: add transaction (A), open chart (C), open screener (S), open chat (H), and refresh data (R)
4. THE MyFi_Platform frontend SHALL display a keyboard shortcuts help modal (? key) showing all available shortcuts organized by category
5. THE MyFi_Platform frontend SHALL implement customizable dashboard layouts using a drag-and-drop grid system allowing users to rearrange, resize, and hide widgets
6. THE MyFi_Platform frontend SHALL provide a widget library with available dashboard widgets: NAV summary, allocation chart, watchlist, recent transactions, alerts, market indices, sector heatmap, and AI insights
7. THE MyFi_Platform frontend SHALL persist custom dashboard layouts in the backend database per user
8. THE MyFi_Platform frontend SHALL implement undo/redo functionality for transaction operations with keyboard shortcuts (Cmd+Z / Cmd+Shift+Z)
9. THE MyFi_Platform frontend SHALL maintain an undo history of the last 20 actions per session
10. THE MyFi_Platform frontend SHALL implement bulk transaction editing allowing users to select multiple transactions and perform batch operations: delete, edit date, edit category, or export
11. THE MyFi_Platform frontend SHALL provide quick actions menu accessible via right-click or long-press on holdings, transactions, and watchlist items
12. THE MyFi_Platform frontend SHALL implement smart search in the header with symbol lookup, company name search, and action suggestions
13. THE MyFi_Platform frontend SHALL provide dashboard presets (Trader, Long-term Investor, Dividend Investor) with pre-configured widget layouts
14. THE MyFi_Platform frontend SHALL implement dark mode auto-switching based on system preference or time of day (dark mode after sunset)

### Requirement 57: Enhanced Security Features

**User Story:** As a security-conscious user, I want two-factor authentication, biometric login, session management, and suspicious activity detection, so that my financial data is protected against unauthorized access.

#### Acceptance Criteria

1. THE Auth_Service SHALL support two-factor authentication (2FA) using TOTP (Time-based One-Time Password) compatible with Google Authenticator, Authy, and similar apps
2. THE Auth_Service SHALL require 2FA setup during account creation or allow users to enable it in security settings
3. WHEN 2FA is enabled, THE Auth_Service SHALL require both password and TOTP code for login
4. THE Auth_Service SHALL provide backup codes (10 single-use codes) during 2FA setup for account recovery if TOTP device is lost
5. THE MyFi_Platform frontend SHALL support biometric authentication (fingerprint, Face ID) on supported devices using Web Authentication API (WebAuthn)
6. THE Auth_Service SHALL implement session management allowing users to view all active sessions with: device type, browser, location (IP-based), last activity timestamp, and current session indicator
7. THE MyFi_Platform frontend SHALL allow users to revoke individual sessions or revoke all sessions except current from the security settings page
8. THE Auth_Service SHALL implement suspicious activity detection flagging: login from new device, login from unusual location (different country), multiple failed login attempts, and unusual API usage patterns
9. WHEN suspicious activity is detected, THE Auth_Service SHALL send an email alert to the user and optionally require additional verification (2FA or email confirmation) for the suspicious session
10. THE Auth_Service SHALL implement device fingerprinting to recognize trusted devices and reduce 2FA prompts for known devices
11. THE Auth_Service SHALL support security questions as an additional recovery method for account access if 2FA device is lost
12. THE MyFi_Platform backend SHALL implement API request signing for sensitive operations (transaction creation, settings changes) to prevent CSRF attacks
13. THE Auth_Service SHALL enforce password complexity requirements: minimum 12 characters, at least one uppercase, one lowercase, one number, and one special character
14. THE Auth_Service SHALL implement password breach detection by checking passwords against known breach databases (Have I Been Pwned API) and forcing password change if compromised
15. THE MyFi_Platform backend SHALL implement rate limiting on authentication endpoints: maximum 5 login attempts per 15 minutes per IP address


### Requirement 58: Market Calendar and Events

**User Story:** As an investor, I want to see a comprehensive calendar of market events including holidays, earnings dates, IPOs, and economic releases, so that I can plan my trading and investment decisions around important dates.

#### Acceptance Criteria

1. THE MyFi_Platform backend SHALL maintain a calendar of VN market holidays and trading schedule including: public holidays, early close days, and special trading sessions
2. THE MyFi_Platform frontend SHALL display a market calendar view showing: upcoming market holidays, earnings announcement dates for holdings, ex-dividend dates, AGM dates, and IPO listings
3. THE Market_Data_Service SHALL fetch earnings calendar data from available sources showing: company symbol, earnings date, estimated EPS, and actual EPS (after announcement)
4. THE MyFi_Platform frontend SHALL highlight earnings dates for holdings in the user's portfolio on the calendar
5. THE Alert_Service SHALL send reminders for upcoming earnings announcements for holdings: 1 week before, 1 day before, and on the day of announcement
6. THE Market_Data_Service SHALL fetch IPO calendar data showing: company name, IPO date, price range, lot size, and subscription period
7. THE MyFi_Platform frontend SHALL display an economic calendar showing VN macroeconomic releases: GDP, CPI, interest rate decisions, trade balance, and FDI data
8. THE MyFi_Platform frontend SHALL allow users to filter calendar events by type: market holidays, earnings, dividends, IPOs, economic releases, and corporate actions
9. THE MyFi_Platform frontend SHALL provide calendar export functionality allowing users to export events to Google Calendar, Outlook, or iCal format
10. THE Alert_Service SHALL allow users to set custom reminders for calendar events with configurable lead time (1 day, 1 week, 1 month before)
11. THE MyFi_Platform frontend SHALL display a "Today's Events" widget on the dashboard showing all events scheduled for the current day
12. THE Market_Data_Service SHALL fetch corporate action calendar data including: stock splits, bonus share distributions, rights issues, and merger/acquisition dates

### Requirement 59: Vietnamese Financial News Sentiment Analysis

**User Story:** As an investor, I want the AI system to analyze sentiment from Vietnamese financial news sources and forums, so that I can gauge market sentiment and incorporate it into investment decisions.

#### Acceptance Criteria

1. THE News_Agent SHALL fetch news articles from major VN financial news sources: CafeF, VietStock, Đầu tư Chứng khoán, Nhịp sống kinh tế, and Báo Đầu tư
2. THE News_Agent SHALL implement Vietnamese language sentiment analysis using NLP models trained on Vietnamese financial text to classify articles as: positive, negative, or neutral
3. THE News_Agent SHALL extract key entities from Vietnamese news articles including: company names, stock symbols, sector mentions, and key figures (executives, government officials)
4. THE News_Agent SHALL aggregate sentiment scores per symbol over rolling time windows: 24 hours, 7 days, and 30 days
5. THE MyFi_Platform frontend SHALL display sentiment indicators on stock detail pages showing: current sentiment (bullish/bearish/neutral), sentiment trend (improving/declining), and recent news count
6. THE News_Agent SHALL monitor popular VN investor forums and social media: CafeF forum, VietStock forum, and relevant Facebook groups for community sentiment
7. THE News_Agent SHALL compute community sentiment scores based on: post frequency, positive/negative keyword analysis, and engagement metrics (likes, comments)
8. THE Analysis_Agent SHALL incorporate news sentiment and community sentiment into stock analysis, factoring sentiment into confidence scores
9. THE Supervisor_Agent SHALL reference significant news events and sentiment shifts in recommendations (e.g., "Recent positive news about VNM's expansion plans has improved sentiment")
10. THE MyFi_Platform frontend SHALL display a sentiment timeline chart showing how sentiment for a symbol has evolved over time
11. THE News_Agent SHALL detect breaking news and significant sentiment shifts, triggering alerts for holdings when sentiment changes dramatically (> 30 point shift in 24 hours)
12. THE News_Agent SHALL provide news summaries in both Vietnamese and English based on user language preference

### Requirement 60: Customizable Alert Scheduling and Grouping

**User Story:** As an investor, I want to configure when I receive alerts and have them grouped intelligently, so that I'm not overwhelmed by notifications and only receive alerts at appropriate times.

#### Acceptance Criteria

1. THE Alert_Service SHALL allow users to configure alert schedules per alert type: trading hours only (9:00-15:00 ICT), business hours (8:00-18:00), or 24/7
2. THE Alert_Service SHALL support quiet hours configuration allowing users to mute all alerts during specified time ranges (e.g., 22:00-07:00)
3. THE Alert_Service SHALL support day-of-week filtering allowing users to disable alerts on weekends or specific weekdays
4. THE Alert_Service SHALL implement alert grouping to prevent notification spam: group multiple alerts of the same type within a 1-hour window into a single notification
5. WHEN alerts are grouped, THE Alert_Service SHALL send a summary notification showing: number of alerts, affected symbols, and alert types
6. THE Alert_Service SHALL implement alert priority levels (critical, high, medium, low) with different delivery rules per priority
7. THE Alert_Service SHALL allow users to configure delivery channels per priority level: critical alerts via SMS + push + email, high via push + email, medium via push only, low via in-app only
8. THE Alert_Service SHALL implement alert digest mode sending a daily summary email of all alerts instead of individual notifications
9. THE Alert_Service SHALL support alert snoozing allowing users to temporarily mute specific alert types for a configurable duration (1 hour, 1 day, 1 week)
10. THE MyFi_Platform frontend SHALL provide an alert management UI showing: active alerts, snoozed alerts, alert history, and alert configuration per type
11. THE Alert_Service SHALL implement smart alert throttling: reduce alert frequency for repeated similar alerts (e.g., price oscillating around alert threshold)
12. THE Alert_Service SHALL allow users to create alert rules with complex conditions: "Alert me if VNM price > 100,000 AND volume > 2M AND RSI < 30"

### Requirement 61: Social Sentiment from VN Investor Community

**User Story:** As an investor, I want to see aggregated sentiment and discussion trends from the VN investor community, so that I can understand what other investors are thinking and identify emerging trends.

#### Acceptance Criteria

1. THE News_Agent SHALL monitor popular VN investor forums: CafeF forum, VietStock forum, and relevant subreddits/Facebook groups for stock discussions
2. THE News_Agent SHALL extract discussion metrics per symbol: post count, comment count, sentiment distribution (bullish/bearish/neutral), and trending status
3. THE MyFi_Platform frontend SHALL display community sentiment indicators on stock detail pages showing: discussion volume (high/medium/low), sentiment breakdown (% bullish/bearish/neutral), and trending status
4. THE News_Agent SHALL identify trending stocks based on discussion volume increase: stocks with 3x or more discussion volume compared to 7-day average
5. THE MyFi_Platform frontend SHALL display a "Community Trending" section on the dashboard showing top 10 most-discussed stocks with sentiment indicators
6. THE News_Agent SHALL extract key discussion topics and themes from forum posts using topic modeling: earnings expectations, technical analysis discussions, news reactions, and rumor/speculation
7. THE MyFi_Platform frontend SHALL display a word cloud or topic summary for each symbol showing common discussion themes
8. THE News_Agent SHALL identify influential community members (high engagement, accurate predictions) and weight their sentiment more heavily
9. THE Supervisor_Agent SHALL incorporate community sentiment into recommendations, noting when community sentiment diverges significantly from technical/fundamental analysis
10. THE MyFi_Platform frontend SHALL provide a community sentiment timeline showing how discussion volume and sentiment have evolved over time
11. THE Alert_Service SHALL allow users to set alerts for community sentiment changes: "Alert me when VNM becomes trending" or "Alert me when community sentiment shifts from bearish to bullish"
12. THE News_Agent SHALL implement spam and bot detection to filter out low-quality posts and artificial sentiment manipulation

### Requirement 62: Portfolio Performance Attribution

**User Story:** As an investor, I want detailed performance attribution analysis showing which holdings, sectors, and decisions contributed most to my returns, so that I can understand what's working and what's not in my strategy.

#### Acceptance Criteria

1. THE Performance_Engine SHALL compute performance attribution breaking down total portfolio return into contributions from: individual holdings, asset type allocation, sector allocation, and timing of trades
2. THE Performance_Engine SHALL compute holding-level attribution showing: each holding's contribution to total return (in VND and percentage points), weight in portfolio, and return of the holding itself
3. THE MyFi_Platform frontend SHALL display a performance attribution chart showing top contributors and top detractors to portfolio performance over selectable time periods
4. THE Performance_Engine SHALL compute sector attribution showing: return contribution from each ICB sector, sector weight vs benchmark, and sector selection effect vs allocation effect
5. THE Performance_Engine SHALL compute asset type attribution showing: return contribution from stocks, gold, crypto, savings, and bonds
6. THE Performance_Engine SHALL compute timing attribution by analyzing the impact of buy/sell timing decisions: comparing actual returns against buy-and-hold returns for the same holdings
7. THE MyFi_Platform frontend SHALL display attribution analysis with waterfall charts showing: starting NAV, contributions from each source, and ending NAV
8. THE Performance_Engine SHALL compute currency attribution showing the impact of USD/VND exchange rate changes on crypto and USD-denominated holdings
9. THE Performance_Engine SHALL identify best and worst investment decisions over a time period based on realized returns and opportunity cost
10. THE MyFi_Platform frontend SHALL provide a "What Went Right / What Went Wrong" summary highlighting: best performing holdings, worst performing holdings, and missed opportunities (symbols on watchlist that outperformed holdings)
11. THE Performance_Engine SHALL compute skill vs luck analysis using statistical methods to estimate how much of portfolio performance is attributable to investment skill vs market conditions
12. THE Supervisor_Agent SHALL incorporate performance attribution insights into recommendations, suggesting: doubling down on successful strategies, cutting losses on underperformers, and learning from past decisions

### Requirement 39: Mobile Responsiveness and Progressive Web App

**User Story:** As a mobile user, I want the platform to work seamlessly on my smartphone with touch-optimized interactions and offline capabilities, so that I can manage my portfolio on the go.

#### Acceptance Criteria

1. THE MyFi_Platform frontend SHALL implement responsive design with breakpoints for mobile (< 768px), tablet (768px - 1024px), and desktop (> 1024px) viewports
2. THE MyFi_Platform frontend SHALL use mobile-first CSS approach with progressive enhancement for larger screens
3. THE Chart_Engine SHALL support touch gestures including: pinch-to-zoom, two-finger pan, tap to show crosshair, and long-press for context menu
4. THE MyFi_Platform frontend SHALL implement a mobile-optimized navigation pattern using a hamburger menu or bottom navigation bar on mobile devices
5. THE MyFi_Platform frontend SHALL register a service worker to enable Progressive Web App (PWA) capabilities
6. THE MyFi_Platform frontend SHALL provide a web app manifest file enabling "Add to Home Screen" functionality on iOS and Android
7. WHEN installed as a PWA, THE MyFi_Platform SHALL cache critical assets (HTML, CSS, JS, fonts) for offline access
8. THE MyFi_Platform frontend SHALL optimize touch target sizes to minimum 44x44 pixels for all interactive elements on mobile
9. THE MyFi_Platform frontend SHALL implement swipe gestures for navigation between tabs and dismissing modals on mobile
10. THE Chart_Engine SHALL adapt chart controls for mobile with larger touch-friendly buttons and simplified indicator configuration panels
11. THE MyFi_Platform frontend SHALL use viewport-relative units and flexible layouts to ensure content fits without horizontal scrolling on all screen sizes
12. THE MyFi_Platform frontend SHALL lazy-load images and heavy components to improve mobile performance on slower networks

### Requirement 40: Real-Time WebSocket Price Streaming

**User Story:** As an investor, I want real-time price updates pushed to my browser via WebSocket instead of polling, so that I see instant price changes with lower latency and reduced server load.

#### Acceptance Criteria

1. THE MyFi_Platform backend SHALL implement a WebSocket server endpoint at /ws/prices for real-time price streaming
2. WHEN a client connects to the WebSocket endpoint, THE MyFi_Platform backend SHALL authenticate the connection using the JWT token passed in the connection handshake
3. THE MyFi_Platform frontend SHALL establish a WebSocket connection on dashboard load and subscribe to price updates for all symbols in the user's portfolio and watchlists
4. THE Price_Service SHALL publish price updates to all connected WebSocket clients whenever new price data is fetched from upstream sources
5. THE MyFi_Platform backend SHALL support subscription management allowing clients to subscribe and unsubscribe from specific symbols dynamically
6. WHEN a price update is received via WebSocket, THE MyFi_Platform frontend SHALL update the UI immediately without waiting for the next polling interval
7. THE MyFi_Platform backend SHALL implement WebSocket heartbeat/ping-pong to detect and close stale connections
8. IF the WebSocket connection is lost, THEN THE MyFi_Platform frontend SHALL automatically attempt to reconnect with exponential backoff (1s, 2s, 4s, 8s, max 30s)
9. THE MyFi_Platform frontend SHALL fall back to HTTP polling if WebSocket connection fails after 5 reconnection attempts
10. THE MyFi_Platform backend SHALL limit each WebSocket connection to subscribing to a maximum of 100 symbols to prevent resource exhaustion
11. THE MyFi_Platform backend SHALL broadcast market status changes (market open, market close, trading halt) to all connected clients via WebSocket

### Requirement 41: Push Notifications

**User Story:** As an investor, I want to receive push notifications on my browser and mobile device for critical alerts, so that I never miss important market events even when the app is not open.

#### Acceptance Criteria

1. THE MyFi_Platform frontend SHALL request push notification permission from the user on first login or when enabling alerts
2. THE MyFi_Platform frontend SHALL register for push notifications using the Web Push API and store the push subscription in the backend
3. THE Alert_Service SHALL send push notifications for high-priority alerts including: price alerts triggered, Monitor_Agent pattern detections with confidence > 80, and upcoming ex-dividend dates within 24 hours
4. THE MyFi_Platform backend SHALL implement Web Push protocol to send notifications to subscribed clients even when the browser is closed
5. THE MyFi_Platform frontend SHALL display push notifications with: title, body text, icon, and action buttons (View, Dismiss)
6. WHEN a user clicks a push notification, THE MyFi_Platform frontend SHALL open the app and navigate to the relevant view (chart for price alerts, alert detail for pattern detections)
7. THE MyFi_Platform frontend SHALL allow users to configure notification preferences per alert type: push, in-app, or disabled
8. THE Alert_Service SHALL respect quiet hours configured by the user (default: no push notifications between 22:00 and 07:00 local time)
9. THE Alert_Service SHALL implement notification grouping to prevent spam: maximum 5 push notifications per hour, with additional alerts queued and sent as a summary
10. THE MyFi_Platform backend SHALL clean up expired push subscriptions (failed delivery for 7 consecutive days) from the database

### Requirement 42: Email and SMS Alerts

**User Story:** As an investor, I want to receive critical alerts via email and SMS in addition to in-app notifications, so that I can stay informed through my preferred communication channels.

#### Acceptance Criteria

1. THE Alert_Service SHALL support three delivery channels: in-app, email, and SMS, with per-alert-type configuration
2. THE MyFi_Platform frontend SHALL allow users to configure email and SMS preferences in the settings panel, including: enabled channels per alert type, email address, and phone number
3. THE Alert_Service SHALL send email alerts using an SMTP service or email API (SendGrid, AWS SES, or similar) for high-priority events
4. THE Alert_Service SHALL send SMS alerts using an SMS gateway API (Twilio, AWS SNS, or similar) for critical events only
5. THE Alert_Service SHALL define alert priority levels: critical (price alerts, large portfolio losses > 5%), high (Monitor_Agent patterns > 80 confidence), medium (ex-dividend reminders), and low (general notifications)
6. THE Alert_Service SHALL send SMS only for critical-priority alerts to minimize SMS costs
7. THE Alert_Service SHALL send email for critical and high-priority alerts with HTML formatting including: alert summary, relevant data, and a link to view details in the app
8. THE Alert_Service SHALL implement rate limiting for email (max 20 per day) and SMS (max 5 per day) per user to prevent notification fatigue and control costs
9. THE Alert_Service SHALL allow users to configure quiet hours separately for each channel (e.g., no SMS between 22:00-07:00, but email allowed anytime)
10. THE Alert_Service SHALL provide an unsubscribe link in all email alerts allowing users to disable email notifications for specific alert types

### Requirement 43: Social Features and Community

**User Story:** As an investor, I want to share my portfolio performance anonymously with the community, follow other investors' public watchlists, and see aggregated sentiment, so that I can learn from the community and benchmark my performance.

#### Acceptance Criteria

1. THE MyFi_Platform backend SHALL allow users to opt-in to sharing their portfolio performance anonymously with the community
2. WHEN a user enables public sharing, THE MyFi_Platform backend SHALL publish anonymized performance metrics (total return %, Sharpe ratio, max drawdown, asset allocation percentages) without revealing absolute NAV values or specific holdings
3. THE MyFi_Platform frontend SHALL display a community leaderboard showing top-performing anonymized portfolios ranked by total return, Sharpe ratio, and risk-adjusted return over selectable time periods (1M, 3M, 6M, 1Y)
4. THE MyFi_Platform backend SHALL allow users to publish watchlists as public with a shareable link
5. THE MyFi_Platform frontend SHALL allow users to follow public watchlists from other users, with followed watchlists appearing in a "Community Watchlists" section
6. THE MyFi_Platform backend SHALL track the number of followers for each public watchlist and display popularity metrics
7. THE MyFi_Platform backend SHALL aggregate sentiment data from public watchlists by counting how many users have added each symbol to their watchlists in the past 7 days
8. THE MyFi_Platform frontend SHALL display community sentiment indicators on stock detail pages showing: number of users watching, trending status (increasing/decreasing watchlist adds), and sentiment score (bullish/bearish based on watchlist activity)
9. THE MyFi_Platform backend SHALL allow users to compare their portfolio performance against anonymized peer benchmarks grouped by portfolio size ranges (< 100M VND, 100M-500M, 500M-1B, > 1B)
10. THE MyFi_Platform frontend SHALL display peer comparison charts showing the user's performance percentile within their portfolio size cohort
11. THE MyFi_Platform backend SHALL implement privacy controls allowing users to hide their profile from leaderboards while still accessing community features
12. THE MyFi_Platform backend SHALL moderate public watchlists and comments to prevent spam and inappropriate content


### Requirement 44: Transaction Import and Broker Integration

**User Story:** As an investor, I want to import my transaction history from CSV files exported by my broker or sync directly with broker APIs, so that I can avoid manual data entry and ensure my portfolio is always up-to-date.

#### Acceptance Criteria

1. THE MyFi_Platform backend SHALL support CSV import for transaction history with a flexible column mapping interface
2. THE MyFi_Platform backend SHALL provide preset CSV templates for major VN brokers: SSI, VPS, HSC, VCBS, Vietcombank Securities, and MB Securities
3. WHEN a user uploads a CSV file, THE MyFi_Platform backend SHALL parse the file, detect the broker format automatically if possible, and display a preview of parsed transactions before import
4. THE MyFi_Platform backend SHALL validate imported transactions for: required fields (symbol, quantity, price, date, type), data type correctness, and logical consistency (sell quantity not exceeding holdings)
5. THE MyFi_Platform backend SHALL support importing the following transaction types from CSV: buy, sell, dividend received, stock split, bonus shares, rights issue, and cash deposit/withdrawal
6. THE MyFi_Platform frontend SHALL provide a reconciliation view showing imported transactions alongside existing transactions, with conflict detection and resolution options (skip, overwrite, merge)
7. THE MyFi_Platform backend SHALL support broker API integration for automated transaction sync where APIs are available (SSI iBoard API, VPS API if accessible)
8. WHEN broker API integration is enabled, THE MyFi_Platform backend SHALL sync transactions daily at a configurable time (default: 16:00 ICT after market close)
9. THE MyFi_Platform backend SHALL store the last sync timestamp per broker connection and display sync status in the frontend
10. THE MyFi_Platform backend SHALL support bank statement parsing for cash flow transactions, detecting deposits and withdrawals from common VN bank statement formats (Vietcombank, Techcombank, VPBank)
11. THE MyFi_Platform frontend SHALL provide a transaction import history log showing all import operations with: timestamp, source (CSV file name or broker API), number of transactions imported, and any errors or warnings
12. THE MyFi_Platform backend SHALL implement duplicate detection to prevent importing the same transaction multiple times based on: symbol, date, quantity, price, and transaction type matching

### Requirement 45: Portfolio Rebalancing Tools

**User Story:** As an investor, I want to see my target allocation vs actual allocation with automatic rebalancing suggestions, so that I can maintain my desired portfolio balance and optimize for my investment strategy.

#### Acceptance Criteria

1. THE MyFi_Platform backend SHALL allow users to define target allocation percentages for asset types (stocks, gold, crypto, savings, bonds, cash) and for individual holdings or sectors
2. THE MyFi_Platform frontend SHALL display a target vs actual allocation comparison chart showing deviations from target percentages
3. THE Portfolio_Engine SHALL compute rebalancing suggestions when actual allocation deviates from target by more than a configurable threshold (default: 5 percentage points)
4. WHEN rebalancing is needed, THE Portfolio_Engine SHALL generate specific buy/sell recommendations with quantities and estimated costs to bring the portfolio back to target allocation
5. THE MyFi_Platform frontend SHALL provide a rebalancing simulator allowing users to preview the portfolio state after executing suggested trades before committing
6. THE Portfolio_Engine SHALL support multiple rebalancing strategies: threshold-based (rebalance when deviation exceeds threshold), calendar-based (rebalance monthly/quarterly), and opportunistic (rebalance when market conditions favor it)
7. THE Portfolio_Engine SHALL factor in transaction costs (broker fees, taxes) when computing rebalancing recommendations to ensure rebalancing is cost-effective
8. THE MyFi_Platform backend SHALL support tax-loss harvesting by identifying holdings with unrealized losses that can be sold to offset capital gains for tax purposes
9. WHEN tax-loss harvesting opportunities are detected, THE Portfolio_Engine SHALL suggest selling losing positions and optionally replacing them with similar assets to maintain allocation
10. THE MyFi_Platform frontend SHALL display the estimated tax savings from tax-loss harvesting recommendations
11. THE Portfolio_Engine SHALL track rebalancing history showing past rebalancing actions, their outcomes, and performance impact
12. THE MyFi_Platform backend SHALL allow users to set rebalancing constraints such as: minimum trade size, maximum number of trades per rebalancing, and excluded holdings (do not sell)

### Requirement 46: Voice Input and Natural Language Queries

**User Story:** As a user, I want to interact with the AI chat using voice input in Vietnamese and ask natural language questions about my portfolio, so that I can get information hands-free and in a conversational manner.

#### Acceptance Criteria

1. THE MyFi_Platform frontend SHALL integrate Web Speech API for voice input with Vietnamese language support (vi-VN)
2. THE MyFi_Platform frontend SHALL provide a microphone button in the chat widget that activates voice recording when pressed
3. WHEN voice recording is active, THE MyFi_Platform frontend SHALL display a visual indicator (animated waveform or pulsing icon) and transcribe speech to text in real-time
4. THE MyFi_Platform frontend SHALL support both push-to-talk (hold button to record) and toggle mode (click to start, click to stop) for voice input
5. THE Supervisor_Agent SHALL parse natural language portfolio queries such as: "How much did I make on tech stocks this quarter?", "What's my best performing asset?", "Show me my dividend income this year"
6. THE Supervisor_Agent SHALL extract intent and entities from natural language queries including: time periods (this quarter, last month, YTD), asset types (tech stocks, gold, crypto), metrics (profit, return, dividend), and actions (show, calculate, compare)
7. THE Supervisor_Agent SHALL query the Portfolio_Engine and Transaction_Ledger to answer natural language questions with specific data from the user's portfolio
8. THE Supervisor_Agent SHALL format responses in natural language with conversational tone matching the user's language preference (Vietnamese or English)
9. THE MyFi_Platform frontend SHALL support voice output (text-to-speech) for AI responses, allowing users to hear answers hands-free
10. THE MyFi_Platform frontend SHALL provide voice command shortcuts for common actions: "Show my portfolio", "Check VNM price", "What's the market doing today", "Set alert for FPT at 100,000"
11. THE Supervisor_Agent SHALL handle ambiguous queries by asking clarifying questions (e.g., "Which quarter did you mean - Q1 2025 or Q1 2026?")
12. THE MyFi_Platform backend SHALL log voice interactions for quality improvement and error analysis while respecting user privacy

### Requirement 47: Advanced Tax Optimization and Reporting

**User Story:** As a taxpayer, I want comprehensive tax optimization suggestions and automated tax report generation for VN securities trading, so that I can minimize my tax liability and file accurately.

#### Acceptance Criteria

1. THE MyFi_Platform backend SHALL compute VN personal income tax liability for securities trading based on current tax regulations (0.1% on sell value for stocks, capital gains tax for other assets)
2. THE Export_Service SHALL generate an annual tax report for VN securities trading containing: total sell value, total tax paid (0.1% withholding), realized capital gains/losses by asset type, and dividend income received
3. THE Portfolio_Engine SHALL identify tax-loss harvesting opportunities by finding holdings with unrealized losses that can offset realized gains
4. THE Portfolio_Engine SHALL compute the optimal timing for selling winning positions to minimize tax impact, considering: holding period, current year realized gains, and projected future gains
5. THE MyFi_Platform frontend SHALL display a tax dashboard showing: YTD tax paid, projected year-end tax liability, available tax-loss harvesting opportunities, and tax-efficient withdrawal strategies
6. THE Portfolio_Engine SHALL suggest tax-efficient withdrawal strategies for users needing to liquidate positions, prioritizing: holdings with losses first, then long-term holdings, then low-gain positions
7. THE Export_Service SHALL generate tax reports in formats suitable for VN tax filing including: detailed transaction ledger, capital gains summary by asset type, and dividend income summary
8. THE MyFi_Platform backend SHALL track cost basis adjustments for corporate actions (stock splits, bonus shares, rights issues) to ensure accurate capital gains calculations
9. THE Portfolio_Engine SHALL compute wash sale violations (selling at a loss and repurchasing within 30 days) and flag them in tax reports
10. THE MyFi_Platform frontend SHALL provide tax scenario modeling allowing users to simulate the tax impact of planned trades before execution
11. THE Export_Service SHALL support exporting tax reports to PDF with Vietnamese language formatting suitable for submission to VN tax authorities
12. THE MyFi_Platform backend SHALL maintain a tax lot tracking system using FIFO, LIFO, or specific identification methods as selected by the user


### Requirement 48: Educational Content and Onboarding

**User Story:** As a new investor, I want interactive tutorials, a financial glossary, and investment strategy templates, so that I can learn how to use the platform and improve my investment knowledge.

#### Acceptance Criteria

1. THE MyFi_Platform frontend SHALL provide an interactive onboarding flow for new users covering: account setup, adding first asset, understanding the dashboard, and using the AI chat
2. THE MyFi_Platform frontend SHALL implement a step-by-step tutorial system using tooltips and guided tours for key features: portfolio tracking, chart analysis, screener, and AI recommendations
3. THE MyFi_Platform frontend SHALL provide a searchable glossary of financial terms in both Vietnamese and English covering: asset types, technical indicators, fundamental metrics, and VN market terminology
4. THE MyFi_Platform frontend SHALL display contextual help tooltips on complex UI elements explaining: what the metric means, how it's calculated, and how to interpret it
5. THE MyFi_Platform backend SHALL provide investment strategy templates including: value investing, growth investing, dividend investing, index tracking, and sector rotation, each with predefined screener filters and allocation targets
6. THE MyFi_Platform frontend SHALL provide video tutorials or animated guides for: setting up portfolio tracking, using the chart engine, interpreting AI recommendations, and understanding risk metrics
7. THE MyFi_Platform frontend SHALL implement a "Learn" section with educational articles covering: VN stock market basics, technical analysis fundamentals, fundamental analysis, portfolio management, and risk management
8. THE MyFi_Platform frontend SHALL provide example portfolios (demo data) that new users can explore before adding their own assets
9. THE MyFi_Platform frontend SHALL implement progressive disclosure, showing basic features first and gradually introducing advanced features as users gain experience
10. THE MyFi_Platform frontend SHALL track user progress through tutorials and educational content, displaying completion badges and unlocking advanced features as milestones are reached
11. THE MyFi_Platform frontend SHALL provide a "Tips & Tricks" section with practical advice for: optimizing portfolio allocation, using technical indicators effectively, and interpreting AI recommendations
12. THE MyFi_Platform frontend SHALL allow users to replay tutorials at any time from the help menu

### Requirement 49: Data Quality and Validation

**User Story:** As an investor, I want the platform to detect and alert me about data anomalies and quality issues, so that I can trust the accuracy of my portfolio valuations and market data.

#### Acceptance Criteria

1. THE Data_Source_Router SHALL implement data anomaly detection for price data, flagging: sudden price jumps > 20% without corresponding volume increase, prices outside ceiling/floor range, and zero or negative prices
2. WHEN a data anomaly is detected, THE Data_Source_Router SHALL log the anomaly with details (symbol, anomaly type, detected value, expected range) and attempt to fetch from alternative source
3. THE MyFi_Platform frontend SHALL display data quality indicators on the dashboard showing: data source reliability score (percentage of successful fetches in past 24 hours), last successful update timestamp, and any active data quality warnings
4. THE Price_Service SHALL validate OHLCV data for logical consistency: high >= low, high >= open, high >= close, low <= open, low <= close, and volume >= 0
5. THE Price_Service SHALL detect and flag gaps in time series data (missing trading days) and interpolate or mark as unavailable
6. THE MyFi_Platform backend SHALL maintain a data source reliability scoring system tracking: uptime percentage, average response time, error rate, and data completeness for each source (VCI, KBS, CoinGecko, Doji)
7. THE MyFi_Platform frontend SHALL allow users to report data issues directly from the UI with a "Report Data Issue" button on stock detail pages
8. THE MyFi_Platform backend SHALL track user-reported data issues in a database with: symbol, issue type (wrong price, missing data, incorrect fundamentals), reporter, timestamp, and resolution status
9. THE Data_Source_Router SHALL implement historical data accuracy verification by comparing current data against previously cached values and flagging significant discrepancies
10. THE MyFi_Platform backend SHALL generate daily data quality reports for administrators showing: anomaly count by source, user-reported issues, data gaps, and source reliability scores
11. THE MyFi_Platform frontend SHALL display a data quality badge on each price display indicating: verified (green), unverified (yellow), or anomaly detected (red)
12. THE Price_Service SHALL implement cross-source validation for critical data, fetching the same data point from multiple sources and flagging discrepancies > 5%

### Requirement 50: Performance Optimization and Caching

**User Story:** As a user, I want the platform to load quickly and respond instantly to interactions, so that I can efficiently manage my portfolio without waiting for slow page loads or API calls.

#### Acceptance Criteria

1. THE MyFi_Platform frontend SHALL implement code splitting to load only the JavaScript required for the current page, with lazy loading for routes and heavy components
2. THE MyFi_Platform frontend SHALL use a service worker to cache static assets (JS, CSS, fonts, images) with a cache-first strategy and background updates
3. THE MyFi_Platform frontend SHALL implement optimistic UI updates, showing expected results immediately while API calls complete in the background
4. THE MyFi_Platform backend SHALL implement Redis or in-memory caching for frequently accessed data with appropriate TTLs: price data (15 min), sector data (30 min), company info (6 hours)
5. THE MyFi_Platform backend SHALL implement database query optimization with proper indexing on: user_id, symbol, transaction_date, and asset_type columns
6. THE MyFi_Platform backend SHALL use database connection pooling to reuse connections and reduce connection overhead
7. THE MyFi_Platform frontend SHALL implement virtual scrolling for long lists (transaction history, screener results) to render only visible items
8. THE Chart_Engine SHALL implement canvas-based rendering for charts instead of SVG to improve performance with large datasets
9. THE MyFi_Platform backend SHALL implement API response compression using gzip or brotli to reduce payload sizes
10. THE MyFi_Platform frontend SHALL use image optimization techniques: lazy loading, responsive images with srcset, and WebP format with fallbacks
11. THE MyFi_Platform backend SHALL implement database read replicas for read-heavy operations (portfolio queries, historical data) to distribute load
12. THE MyFi_Platform frontend SHALL implement request deduplication to prevent multiple identical API calls from being sent simultaneously
13. THE MyFi_Platform backend SHALL use CDN for serving static assets with edge caching to reduce latency for users across different regions
14. THE MyFi_Platform frontend SHALL implement skeleton screens and progressive loading to show content structure immediately while data loads

### Requirement 51: Compliance and Audit Trail

**User Story:** As a platform operator, I want comprehensive activity logging and audit trails for all user actions, so that I can ensure compliance, investigate issues, and maintain data integrity.

#### Acceptance Criteria

1. THE MyFi_Platform backend SHALL log all user actions in an audit trail including: login/logout, portfolio changes (add/edit/delete assets), transaction creation/modification, settings changes, and data exports
2. THE audit trail SHALL record for each action: user ID, action type, timestamp, IP address, user agent, affected resources (asset IDs, transaction IDs), old values, new values, and action result (success/failure)
3. THE MyFi_Platform backend SHALL implement immutable audit logs that cannot be modified or deleted by users or administrators
4. THE MyFi_Platform backend SHALL retain audit logs for a configurable period (default: 7 years) to meet financial record-keeping requirements
5. THE MyFi_Platform backend SHALL provide an audit log query interface for administrators with filters by: user, date range, action type, and resource
6. THE MyFi_Platform backend SHALL implement data retention policies allowing users to request data deletion after account closure, while preserving audit logs for compliance
7. THE MyFi_Platform backend SHALL support GDPR-style data export allowing users to download all their data in machine-readable format (JSON)
8. THE MyFi_Platform backend SHALL implement data anonymization for deleted accounts, removing PII while preserving anonymized data for analytics
9. THE MyFi_Platform backend SHALL log all API calls with: endpoint, method, request parameters, response status, response time, and user ID
10. THE MyFi_Platform backend SHALL implement security event logging for: failed login attempts, password changes, session expirations, and suspicious activity (multiple failed logins, unusual access patterns)
11. THE MyFi_Platform backend SHALL provide compliance reports showing: total users, active users, data storage size, API usage statistics, and security events
12. THE MyFi_Platform backend SHALL implement automated backup of audit logs to separate storage with encryption at rest

### Requirement 52: Advanced Charting Features

**User Story:** As a trader, I want to save chart templates, use multi-timeframe analysis, replay historical trading sessions, and get pattern recognition alerts, so that I can perform sophisticated technical analysis.

#### Acceptance Criteria

1. THE Chart_Engine SHALL allow users to save chart templates containing: selected indicators with parameters, drawing tools, time interval, and chart style (candlestick, line, bar)
2. THE Chart_Engine SHALL allow users to name, save, and load chart templates, with templates persisted in the backend database
3. THE Chart_Engine SHALL support multi-timeframe (MTF) indicators, displaying indicator values from higher timeframes on lower timeframe charts (e.g., daily SMA on 1-hour chart)
4. THE Chart_Engine SHALL provide a chart replay mode allowing users to step through historical price action bar-by-bar to simulate live trading conditions
5. WHEN chart replay mode is active, THE Chart_Engine SHALL hide future price data and allow users to place simulated trades to practice strategy execution
6. THE Chart_Engine SHALL implement automated pattern recognition for common chart patterns: head and shoulders, inverse head and shoulders, double top, double bottom, ascending triangle, descending triangle, symmetrical triangle, flag, and pennant
7. WHEN a chart pattern is detected, THE Chart_Engine SHALL draw the pattern on the chart with labels and provide a pattern analysis (bullish/bearish, target price, stop loss level)
8. THE Alert_Service SHALL allow users to set pattern recognition alerts, receiving notifications when specific patterns are detected on watched symbols
9. THE Chart_Engine SHALL implement volume profile analysis showing volume distribution at different price levels as a horizontal histogram overlaid on the chart
10. THE Chart_Engine SHALL support custom time ranges allowing users to select any arbitrary start and end date for chart display
11. THE Chart_Engine SHALL provide a comparison mode overlaying multiple symbols on the same chart with normalized percentage scale
12. THE Chart_Engine SHALL implement chart annotations allowing users to add text notes, arrows, and shapes to mark important levels or events
13. THE Chart_Engine SHALL sync chart state across devices for logged-in users, preserving: active indicators, drawings, annotations, and zoom level
14. THE Chart_Engine SHALL provide keyboard shortcuts for common actions: add indicator (I), draw trend line (T), toggle crosshair (C), zoom in/out (+/-), and reset zoom (R)


### Requirement 53: Portfolio Stress Testing and Scenario Analysis

**User Story:** As an investor, I want to stress test my portfolio against historical crises and custom scenarios, so that I can understand how my portfolio would perform under adverse conditions and prepare accordingly.

#### Acceptance Criteria

1. THE Risk_Service SHALL implement historical stress testing by replaying portfolio performance during past market crises: 2008 global financial crisis, 2020 COVID crash, and VN-specific events (2011 banking crisis)
2. WHEN historical stress testing is run, THE Risk_Service SHALL compute: portfolio drawdown during the crisis period, recovery time to pre-crisis NAV, and comparison against VN-Index performance during the same period
3. THE Risk_Service SHALL support custom scenario analysis allowing users to define hypothetical market conditions: percentage change in VN-Index, sector-specific shocks, currency devaluation, and interest rate changes
4. THE Risk_Service SHALL compute portfolio impact for custom scenarios showing: projected NAV change, affected holdings, and risk contribution by asset type
5. THE MyFi_Platform frontend SHALL display stress test results with visualizations: NAV drawdown chart, recovery timeline, and worst-case scenario projections
6. THE Risk_Service SHALL implement Monte Carlo simulation for portfolio projections, running 10,000 simulations with randomized returns based on historical volatility
7. THE MyFi_Platform frontend SHALL display Monte Carlo results showing: probability distribution of future NAV, confidence intervals (10th, 50th, 90th percentile), and probability of reaching financial goals
8. THE Risk_Service SHALL compute correlation breakdown scenarios showing how portfolio risk changes if correlations between assets increase during market stress
9. THE Risk_Service SHALL identify portfolio concentration risks by flagging: single holdings exceeding 20% of NAV, sector concentration exceeding 40%, and asset type concentration exceeding 60%
10. THE MyFi_Platform frontend SHALL provide a "What-If" calculator allowing users to simulate adding or removing positions and see the impact on portfolio risk metrics
11. THE Risk_Service SHALL compute tail risk metrics including: conditional Value at Risk (CVaR), expected shortfall, and maximum loss at 99% confidence level
12. THE MyFi_Platform frontend SHALL display stress test recommendations suggesting portfolio adjustments to improve resilience: diversification suggestions, hedging strategies, and safe-haven asset allocation

### Requirement 54: Public API and Webhooks

**User Story:** As a developer, I want access to a public API and webhook system, so that I can integrate MyFi data with external tools and receive real-time notifications for custom workflows.

#### Acceptance Criteria

1. THE MyFi_Platform backend SHALL provide a public REST API with endpoints for: portfolio data, transaction history, price data, market data, and AI recommendations
2. THE MyFi_Platform backend SHALL implement API key authentication for public API access with rate limiting per API key
3. THE MyFi_Platform backend SHALL provide API documentation using OpenAPI/Swagger specification with interactive API explorer
4. THE MyFi_Platform backend SHALL implement webhook support allowing users to register webhook URLs for specific events: price alerts triggered, portfolio NAV milestones reached, and AI pattern detections
5. WHEN a webhook event occurs, THE MyFi_Platform backend SHALL send an HTTP POST request to the registered webhook URL with event payload in JSON format
6. THE MyFi_Platform backend SHALL implement webhook retry logic with exponential backoff for failed deliveries (3 retries: 1s, 5s, 15s)
7. THE MyFi_Platform backend SHALL provide webhook signature verification using HMAC-SHA256 to ensure webhook authenticity
8. THE MyFi_Platform frontend SHALL provide a webhook management UI allowing users to: register webhooks, test webhooks, view delivery logs, and disable webhooks
9. THE MyFi_Platform backend SHALL implement API rate limiting: 100 requests per minute for authenticated users, 10 requests per minute for unauthenticated access
10. THE MyFi_Platform backend SHALL provide API usage analytics showing: request count, error rate, most-used endpoints, and rate limit violations
11. THE MyFi_Platform backend SHALL support API versioning (v1, v2) to maintain backward compatibility when introducing breaking changes
12. THE MyFi_Platform backend SHALL provide SDKs or client libraries for popular languages (JavaScript, Python, Go) to simplify API integration

### Requirement 55: Integration with External Services

**User Story:** As a user, I want to export my portfolio data to popular portfolio trackers and integrate with tax software, so that I can use MyFi alongside my existing financial tools.

#### Acceptance Criteria

1. THE Export_Service SHALL support exporting portfolio data in formats compatible with popular portfolio trackers: Yahoo Finance CSV, Google Sheets format, and Personal Capital format
2. THE Export_Service SHALL provide one-click export to Google Sheets with automatic sheet creation and data population using Google Sheets API
3. THE MyFi_Platform backend SHALL support OAuth integration with Google account for Google Sheets export and Google Drive backup
4. THE Export_Service SHALL generate tax reports compatible with VN tax software formats if available, or provide detailed CSV suitable for manual entry
5. THE MyFi_Platform backend SHALL support exporting transaction history to accounting software formats: QuickBooks CSV, Xero format, and generic accounting CSV
6. THE MyFi_Platform frontend SHALL provide a "Connect" section in settings showing available integrations with status indicators (connected, not connected, error)
7. THE MyFi_Platform backend SHALL implement scheduled exports allowing users to automatically export portfolio snapshots to Google Drive or Dropbox on a daily, weekly, or monthly schedule
8. THE MyFi_Platform backend SHALL support importing watchlists from external sources: TradingView watchlist export, Yahoo Finance portfolio CSV, and generic symbol lists
9. THE Export_Service SHALL provide portfolio performance reports in formats suitable for sharing with financial advisors: PDF with charts, Excel with detailed breakdowns
10. THE MyFi_Platform backend SHALL support calendar integration (Google Calendar, Outlook) for syncing corporate action events, ex-dividend dates, and earnings dates
11. THE MyFi_Platform backend SHALL implement Zapier integration allowing users to create custom automation workflows triggered by MyFi events
12. THE MyFi_Platform backend SHALL provide IFTTT integration for simple automation rules: "If portfolio NAV increases by 10%, then send email"

### Requirement 56: User Experience Enhancements

**User Story:** As a power user, I want keyboard shortcuts, customizable dashboard layouts, a command palette, and undo/redo functionality, so that I can work efficiently and personalize my workflow.

#### Acceptance Criteria

1. THE MyFi_Platform frontend SHALL implement a command palette (Cmd+K or Ctrl+K) providing quick access to all features, navigation, and actions with fuzzy search
2. THE command palette SHALL support action execution including: navigate to pages, add transaction, create alert, search symbols, and change settings
3. THE MyFi_Platform frontend SHALL provide keyboard shortcuts for common actions: add transaction (A), open chart (C), open screener (S), open chat (H), and refresh data (R)
4. THE MyFi_Platform frontend SHALL display a keyboard shortcuts help modal (? key) showing all available shortcuts organized by category
5. THE MyFi_Platform frontend SHALL implement customizable dashboard layouts using a drag-and-drop grid system allowing users to rearrange, resize, and hide widgets
6. THE MyFi_Platform frontend SHALL provide a widget library with available dashboard widgets: NAV summary, allocation chart, watchlist, recent transactions, alerts, market indices, sector heatmap, and AI insights
7. THE MyFi_Platform frontend SHALL persist custom dashboard layouts in the backend database per user
8. THE MyFi_Platform frontend SHALL implement undo/redo functionality for transaction operations with keyboard shortcuts (Cmd+Z / Cmd+Shift+Z)
9. THE MyFi_Platform frontend SHALL maintain an undo history of the last 20 actions per session
10. THE MyFi_Platform frontend SHALL implement bulk transaction editing allowing users to select multiple transactions and perform batch operations: delete, edit date, edit category, or export
11. THE MyFi_Platform frontend SHALL provide quick actions menu accessible via right-click or long-press on holdings, transactions, and watchlist items
12. THE MyFi_Platform frontend SHALL implement smart search in the header with symbol lookup, company name search, and action suggestions
13. THE MyFi_Platform frontend SHALL provide dashboard presets (Trader, Long-term Investor, Dividend Investor) with pre-configured widget layouts
14. THE MyFi_Platform frontend SHALL implement dark mode auto-switching based on system preference or time of day (dark mode after sunset)

### Requirement 57: Enhanced Security Features

**User Story:** As a security-conscious user, I want two-factor authentication, biometric login, session management, and suspicious activity detection, so that my financial data is protected against unauthorized access.

#### Acceptance Criteria

1. THE Auth_Service SHALL support two-factor authentication (2FA) using TOTP (Time-based One-Time Password) compatible with Google Authenticator, Authy, and similar apps
2. THE Auth_Service SHALL require 2FA setup during account creation or allow users to enable it in security settings
3. WHEN 2FA is enabled, THE Auth_Service SHALL require both password and TOTP code for login
4. THE Auth_Service SHALL provide backup codes (10 single-use codes) during 2FA setup for account recovery if TOTP device is lost
5. THE MyFi_Platform frontend SHALL support biometric authentication (fingerprint, Face ID) on supported devices using Web Authentication API (WebAuthn)
6. THE Auth_Service SHALL implement session management allowing users to view all active sessions with: device type, browser, location (IP-based), last activity timestamp, and current session indicator
7. THE MyFi_Platform frontend SHALL allow users to revoke individual sessions or revoke all sessions except current from the security settings page
8. THE Auth_Service SHALL implement suspicious activity detection flagging: login from new device, login from unusual location (different country), multiple failed login attempts, and unusual API usage patterns
9. WHEN suspicious activity is detected, THE Auth_Service SHALL send an email alert to the user and optionally require additional verification (2FA or email confirmation) for the suspicious session
10. THE Auth_Service SHALL implement device fingerprinting to recognize trusted devices and reduce 2FA prompts for known devices
11. THE Auth_Service SHALL support security questions as an additional recovery method for account access if 2FA device is lost
12. THE MyFi_Platform backend SHALL implement API request signing for sensitive operations (transaction creation, settings changes) to prevent CSRF attacks
13. THE Auth_Service SHALL enforce password complexity requirements: minimum 12 characters, at least one uppercase, one lowercase, one number, and one special character
14. THE Auth_Service SHALL implement password breach detection by checking passwords against known breach databases (Have I Been Pwned API) and forcing password change if compromised
15. THE MyFi_Platform backend SHALL implement rate limiting on authentication endpoints: maximum 5 login attempts per 15 minutes per IP address


### Requirement 58: Market Calendar and Events

**User Story:** As an investor, I want to see a comprehensive calendar of market events including holidays, earnings dates, IPOs, and economic releases, so that I can plan my trading and investment decisions around important dates.

#### Acceptance Criteria

1. THE MyFi_Platform backend SHALL maintain a calendar of VN market holidays and trading schedule including: public holidays, early close days, and special trading sessions
2. THE MyFi_Platform frontend SHALL display a market calendar view showing: upcoming market holidays, earnings announcement dates for holdings, ex-dividend dates, AGM dates, and IPO listings
3. THE Market_Data_Service SHALL fetch earnings calendar data from available sources showing: company symbol, earnings date, estimated EPS, and actual EPS (after announcement)
4. THE MyFi_Platform frontend SHALL highlight earnings dates for holdings in the user's portfolio on the calendar
5. THE Alert_Service SHALL send reminders for upcoming earnings announcements for holdings: 1 week before, 1 day before, and on the day of announcement
6. THE Market_Data_Service SHALL fetch IPO calendar data showing: company name, IPO date, price range, lot size, and subscription period
7. THE MyFi_Platform frontend SHALL display an economic calendar showing VN macroeconomic releases: GDP, CPI, interest rate decisions, trade balance, and FDI data
8. THE MyFi_Platform frontend SHALL allow users to filter calendar events by type: market holidays, earnings, dividends, IPOs, economic releases, and corporate actions
9. THE MyFi_Platform frontend SHALL provide calendar export functionality allowing users to export events to Google Calendar, Outlook, or iCal format
10. THE Alert_Service SHALL allow users to set custom reminders for calendar events with configurable lead time (1 day, 1 week, 1 month before)
11. THE MyFi_Platform frontend SHALL display a "Today's Events" widget on the dashboard showing all events scheduled for the current day
12. THE Market_Data_Service SHALL fetch corporate action calendar data including: stock splits, bonus share distributions, rights issues, and merger/acquisition dates

### Requirement 59: Vietnamese Financial News Sentiment Analysis

**User Story:** As an investor, I want the AI system to analyze sentiment from Vietnamese financial news sources and forums, so that I can gauge market sentiment and incorporate it into investment decisions.

#### Acceptance Criteria

1. THE News_Agent SHALL fetch news articles from major VN financial news sources: CafeF, VietStock, Đầu tư Chứng khoán, Nhịp sống kinh tế, and Báo Đầu tư
2. THE News_Agent SHALL implement Vietnamese language sentiment analysis using NLP models trained on Vietnamese financial text to classify articles as: positive, negative, or neutral
3. THE News_Agent SHALL extract key entities from Vietnamese news articles including: company names, stock symbols, sector mentions, and key figures (executives, government officials)
4. THE News_Agent SHALL aggregate sentiment scores per symbol over rolling time windows: 24 hours, 7 days, and 30 days
5. THE MyFi_Platform frontend SHALL display sentiment indicators on stock detail pages showing: current sentiment (bullish/bearish/neutral), sentiment trend (improving/declining), and recent news count
6. THE News_Agent SHALL monitor popular VN investor forums and social media: CafeF forum, VietStock forum, and relevant Facebook groups for community sentiment
7. THE News_Agent SHALL compute community sentiment scores based on: post frequency, positive/negative keyword analysis, and engagement metrics (likes, comments)
8. THE Analysis_Agent SHALL incorporate news sentiment and community sentiment into stock analysis, factoring sentiment into confidence scores
9. THE Supervisor_Agent SHALL reference significant news events and sentiment shifts in recommendations (e.g., "Recent positive news about VNM's expansion plans has improved sentiment")
10. THE MyFi_Platform frontend SHALL display a sentiment timeline chart showing how sentiment for a symbol has evolved over time
11. THE News_Agent SHALL detect breaking news and significant sentiment shifts, triggering alerts for holdings when sentiment changes dramatically (> 30 point shift in 24 hours)
12. THE News_Agent SHALL provide news summaries in both Vietnamese and English based on user language preference

### Requirement 60: Customizable Alert Scheduling and Grouping

**User Story:** As an investor, I want to configure when I receive alerts and have them grouped intelligently, so that I'm not overwhelmed by notifications and only receive alerts at appropriate times.

#### Acceptance Criteria

1. THE Alert_Service SHALL allow users to configure alert schedules per alert type: trading hours only (9:00-15:00 ICT), business hours (8:00-18:00), or 24/7
2. THE Alert_Service SHALL support quiet hours configuration allowing users to mute all alerts during specified time ranges (e.g., 22:00-07:00)
3. THE Alert_Service SHALL support day-of-week filtering allowing users to disable alerts on weekends or specific weekdays
4. THE Alert_Service SHALL implement alert grouping to prevent notification spam: group multiple alerts of the same type within a 1-hour window into a single notification
5. WHEN alerts are grouped, THE Alert_Service SHALL send a summary notification showing: number of alerts, affected symbols, and alert types
6. THE Alert_Service SHALL implement alert priority levels (critical, high, medium, low) with different delivery rules per priority
7. THE Alert_Service SHALL allow users to configure delivery channels per priority level: critical alerts via SMS + push + email, high via push + email, medium via push only, low via in-app only
8. THE Alert_Service SHALL implement alert digest mode sending a daily summary email of all alerts instead of individual notifications
9. THE Alert_Service SHALL support alert snoozing allowing users to temporarily mute specific alert types for a configurable duration (1 hour, 1 day, 1 week)
10. THE MyFi_Platform frontend SHALL provide an alert management UI showing: active alerts, snoozed alerts, alert history, and alert configuration per type
11. THE Alert_Service SHALL implement smart alert throttling: reduce alert frequency for repeated similar alerts (e.g., price oscillating around alert threshold)
12. THE Alert_Service SHALL allow users to create alert rules with complex conditions: "Alert me if VNM price > 100,000 AND volume > 2M AND RSI < 30"

### Requirement 61: Social Sentiment from VN Investor Community

**User Story:** As an investor, I want to see aggregated sentiment and discussion trends from the VN investor community, so that I can understand what other investors are thinking and identify emerging trends.

#### Acceptance Criteria

1. THE News_Agent SHALL monitor popular VN investor forums: CafeF forum, VietStock forum, and relevant subreddits/Facebook groups for stock discussions
2. THE News_Agent SHALL extract discussion metrics per symbol: post count, comment count, sentiment distribution (bullish/bearish/neutral), and trending status
3. THE MyFi_Platform frontend SHALL display community sentiment indicators on stock detail pages showing: discussion volume (high/medium/low), sentiment breakdown (% bullish/bearish/neutral), and trending status
4. THE News_Agent SHALL identify trending stocks based on discussion volume increase: stocks with 3x or more discussion volume compared to 7-day average
5. THE MyFi_Platform frontend SHALL display a "Community Trending" section on the dashboard showing top 10 most-discussed stocks with sentiment indicators
6. THE News_Agent SHALL extract key discussion topics and themes from forum posts using topic modeling: earnings expectations, technical analysis discussions, news reactions, and rumor/speculation
7. THE MyFi_Platform frontend SHALL display a word cloud or topic summary for each symbol showing common discussion themes
8. THE News_Agent SHALL identify influential community members (high engagement, accurate predictions) and weight their sentiment more heavily
9. THE Supervisor_Agent SHALL incorporate community sentiment into recommendations, noting when community sentiment diverges significantly from technical/fundamental analysis
10. THE MyFi_Platform frontend SHALL provide a community sentiment timeline showing how discussion volume and sentiment have evolved over time
11. THE Alert_Service SHALL allow users to set alerts for community sentiment changes: "Alert me when VNM becomes trending" or "Alert me when community sentiment shifts from bearish to bullish"
12. THE News_Agent SHALL implement spam and bot detection to filter out low-quality posts and artificial sentiment manipulation

### Requirement 62: Portfolio Performance Attribution

**User Story:** As an investor, I want detailed performance attribution analysis showing which holdings, sectors, and decisions contributed most to my returns, so that I can understand what's working and what's not in my strategy.

#### Acceptance Criteria

1. THE Performance_Engine SHALL compute performance attribution breaking down total portfolio return into contributions from: individual holdings, asset type allocation, sector allocation, and timing of trades
2. THE Performance_Engine SHALL compute holding-level attribution showing: each holding's contribution to total return (in VND and percentage points), weight in portfolio, and return of the holding itself
3. THE MyFi_Platform frontend SHALL display a performance attribution chart showing top contributors and top detractors to portfolio performance over selectable time periods
4. THE Performance_Engine SHALL compute sector attribution showing: return contribution from each ICB sector, sector weight vs benchmark, and sector selection effect vs allocation effect
5. THE Performance_Engine SHALL compute asset type attribution showing: return contribution from stocks, gold, crypto, savings, and bonds
6. THE Performance_Engine SHALL compute timing attribution by analyzing the impact of buy/sell timing decisions: comparing actual returns against buy-and-hold returns for the same holdings
7. THE MyFi_Platform frontend SHALL display attribution analysis with waterfall charts showing: starting NAV, contributions from each source, and ending NAV
8. THE Performance_Engine SHALL compute currency attribution showing the impact of USD/VND exchange rate changes on crypto and USD-denominated holdings
9. THE Performance_Engine SHALL identify best and worst investment decisions over a time period based on realized returns and opportunity cost
10. THE MyFi_Platform frontend SHALL provide a "What Went Right / What Went Wrong" summary highlighting: best performing holdings, worst performing holdings, and missed opportunities (symbols on watchlist that outperformed holdings)
11. THE Performance_Engine SHALL compute skill vs luck analysis using statistical methods to estimate how much of portfolio performance is attributable to investment skill vs market conditions
12. THE Supervisor_Agent SHALL incorporate performance attribution insights into recommendations, suggesting: doubling down on successful strategies, cutting losses on underperformers, and learning from past decisions



### Requirement 63: Advanced Chart Types

**User Story:** As a trader, I want access to advanced chart types beyond basic candlesticks including Hollow Candles, Heikin Ashi, HLC Area, Line with markers, High-Low, and Step Line, so that I can visualize price action in different ways to identify trends and patterns more effectively.

#### Acceptance Criteria

1. THE Chart_Engine SHALL support Hollow Candles chart type where bullish candles (close > open) are rendered hollow and bearish candles (close < open) are filled, making trend visualization clearer
2. THE Chart_Engine SHALL support Heikin Ashi chart type which smooths price data using modified OHLC calculations: HA_Close = (Open + High + Low + Close) / 4, HA_Open = (Previous HA_Open + Previous HA_Close) / 2, HA_High = max(High, HA_Open, HA_Close), HA_Low = min(Low, HA_Open, HA_Close)
3. THE Chart_Engine SHALL support HLC Area chart type displaying high, low, and close prices as a filled area chart without open prices
4. THE Chart_Engine SHALL support Line with markers chart type rendering a line chart with circular markers at each data point for highlighting specific events or price levels
5. THE Chart_Engine SHALL support High-Low chart type displaying only high and low prices as vertical bars without open and close prices
6. THE Chart_Engine SHALL support Step Line chart type rendering price changes as horizontal steps rather than diagonal lines, useful for visualizing discrete price movements
7. THE Chart_Engine SHALL provide a chart type selector in the chart toolbar allowing users to switch between all supported chart types: Candlestick, Hollow Candles, Heikin Ashi, Area, Line, Baseline, Columns, HLC Area, Line with markers, High-Low, and Step Line
8. WHEN a user switches chart types, THE Chart_Engine SHALL preserve all active indicators, drawings, and zoom level while re-rendering the price data in the new chart style
9. THE Chart_Engine SHALL persist the selected chart type preference per symbol in local storage so the user's preferred chart type is restored when viewing the same symbol again
10. THE Chart_Engine SHALL adapt chart colors based on the current theme (light/dark) for all chart types ensuring readability in both modes
11. WHEN Heikin Ashi chart type is active, THE Chart_Engine SHALL display a visual indicator in the chart header noting that smoothed candles are being displayed rather than actual OHLC data
12. THE Chart_Engine SHALL support exporting charts in any chart type to PNG or SVG format preserving the visual appearance including colors, indicators, and drawings

### Requirement 64: 110+ Drawing Tools

**User Story:** As a technical analyst, I want access to 110+ professional drawing tools including channels, pitchforks, Gann tools, Elliott Wave tools, geometric shapes, text annotations, price projections, time cycles, and measurement tools, so that I can perform comprehensive chart analysis and mark important levels and patterns.

#### Acceptance Criteria

1. THE Chart_Engine SHALL support channel drawing tools including: Parallel Channel (two parallel trend lines), Regression Channel (linear regression with standard deviation bands), and Fibonacci Channel (parallel lines at Fibonacci ratios)
2. THE Chart_Engine SHALL support pitchfork tools including: Andrews Pitchfork (three parallel trend lines from three pivot points) and Schiff Pitchfork (modified Andrews Pitchfork with adjusted starting point)
3. THE Chart_Engine SHALL support Gann tools including: Gann Fan (diagonal lines at Gann angles: 1x1, 1x2, 2x1, 1x3, 3x1, 1x4, 4x1, 1x8, 8x1), Gann Grid (grid of horizontal and vertical lines at Gann intervals), and Gann Square (square overlay with Gann angles)
4. THE Chart_Engine SHALL support Elliott Wave tools including: Wave Pattern drawing (labeling 5-wave impulse and 3-wave corrective patterns) and Wave Labels (customizable labels for wave counts: 1, 2, 3, 4, 5, A, B, C)
5. THE Chart_Engine SHALL support geometric shape tools including: Circle, Ellipse, Triangle, Rectangle (already supported), Polygon (multi-point custom shape), and Arc
6. THE Chart_Engine SHALL support text annotation tools including: Text Label (simple text at a point), Callout (text with arrow pointer), Arrow (directional arrow between two points), and Price Note (text anchored to a specific price level)
7. THE Chart_Engine SHALL support price projection tools including: XABCD Pattern (harmonic pattern with 5 points: X, A, B, C, D with Fibonacci ratios), Head and Shoulders Pattern (marking left shoulder, head, right shoulder, and neckline), and Price Target Projection (projecting future price based on measured move)
8. THE Chart_Engine SHALL support time cycle tools including: Cycle Lines (vertical lines at regular time intervals), Time Zones (shaded vertical zones marking time periods), and Periodic Markers (recurring markers at specified intervals)
9. THE Chart_Engine SHALL support measurement tools including: Price Range (measuring vertical distance between two price levels with percentage and VND display), Date Range (measuring horizontal distance between two dates with day count), and Distance Tool (measuring diagonal distance with price and time components)
10. THE Chart_Engine SHALL provide a drawing tools panel accessible from the chart toolbar organized by category: Trend Lines, Channels, Fibonacci, Pitchforks, Gann, Elliott Wave, Shapes, Annotations, Projections, Time, and Measurements
11. WHEN a drawing tool is selected, THE Chart_Engine SHALL display usage instructions in a tooltip (e.g., "Click three points to draw Andrews Pitchfork") and change the cursor to indicate drawing mode
12. THE Chart_Engine SHALL allow users to edit existing drawings by clicking to select, then dragging anchor points to adjust position, or using handles to resize and rotate
13. THE Chart_Engine SHALL provide drawing properties panel allowing users to customize: line color, line width, line style (solid, dashed, dotted), fill color, fill opacity, and text properties (font, size, color)
14. THE Chart_Engine SHALL support drawing templates allowing users to save frequently used drawing configurations (e.g., "Blue Trend Line - 2px") and apply them to new drawings with one click
15. THE Chart_Engine SHALL implement drawing layers with z-index control allowing users to bring drawings to front or send to back when overlapping
16. THE Chart_Engine SHALL support drawing groups allowing users to select multiple drawings and group them together for moving, copying, or deleting as a unit
17. THE Chart_Engine SHALL provide drawing alignment tools: align left, align right, align top, align bottom, distribute horizontally, and distribute vertically for multiple selected drawings
18. THE Chart_Engine SHALL implement drawing snapping to OHLC data points, indicator lines, and other drawings when enabled, making precise placement easier
19. THE Chart_Engine SHALL support drawing duplication allowing users to copy a drawing and paste it at a different location on the chart
20. THE Chart_Engine SHALL persist all drawings per symbol in the backend database so drawings are synchronized across devices and sessions for logged-in users
21. THE Chart_Engine SHALL provide a drawings list panel showing all drawings on the current chart with visibility toggles, allowing users to show/hide individual drawings without deleting them
22. THE Chart_Engine SHALL support drawing alerts allowing users to set alerts when price crosses a drawn trend line, touches a Fibonacci level, or enters a drawn shape
23. THE Chart_Engine SHALL implement drawing magnetism where trend lines and channels automatically adjust to nearby swing highs and lows when drawn close to them
24. THE Chart_Engine SHALL support drawing cloning across symbols allowing users to copy all drawings from one symbol's chart and apply them to another symbol's chart with automatic scaling

### Requirement 65: Advanced Price Scaling

**User Story:** As a trader, I want advanced price scaling options including multiple simultaneous price scales, percent scale, indexed to 100 scale, locked scale, and inverted scale, so that I can compare multiple assets, normalize price movements, and customize chart display for different analysis needs.

#### Acceptance Criteria

1. THE Chart_Engine SHALL support multiple price scales (up to 8 simultaneous) allowing different indicators or overlays to use independent Y-axis scales displayed on the left or right side of the chart
2. THE Chart_Engine SHALL support Percent scale mode where all prices are displayed as percentage change from a reference point (first visible candle, selected date, or custom reference price)
3. THE Chart_Engine SHALL support Indexed to 100 scale mode where all prices are normalized to 100 at a reference point, enabling direct comparison of multiple assets with different absolute price levels
4. THE Chart_Engine SHALL support Locked scale mode preventing automatic Y-axis rescaling when zooming or panning, maintaining a fixed price range for consistent visual analysis
5. THE Chart_Engine SHALL support Inverted scale mode flipping the Y-axis so higher prices are displayed at the bottom and lower prices at the top, useful for analyzing inverse relationships
6. THE Chart_Engine SHALL provide a scale settings panel accessible from the chart toolbar with options to: select scale mode (Auto, Percent, Indexed, Locked, Inverted), set reference point for Percent and Indexed modes, and configure scale range for Locked mode
7. WHEN Percent scale mode is active, THE Chart_Engine SHALL display percentage values on the Y-axis (e.g., +5.2%, -3.1%) and show percentage change in tooltips instead of absolute prices
8. WHEN Indexed to 100 scale mode is active, THE Chart_Engine SHALL display indexed values on the Y-axis (e.g., 95, 100, 105) and show indexed values in tooltips with the reference date clearly indicated
9. WHEN multiple price scales are active, THE Chart_Engine SHALL color-code each Y-axis to match the corresponding indicator or overlay, and display scale labels clearly identifying which data series uses which scale
10. THE Chart_Engine SHALL allow users to assign indicators to specific price scales by right-clicking an indicator and selecting "Move to Scale" with options: Main (price scale), Scale 2, Scale 3, etc., or "New Scale"
11. THE Chart_Engine SHALL support logarithmic scale mode where the Y-axis uses logarithmic spacing, useful for analyzing assets with large price movements or long-term trends
12. THE Chart_Engine SHALL persist scale settings per symbol in local storage so the user's preferred scale configuration is restored when viewing the same symbol again
13. WHEN comparing multiple symbols on the same chart, THE Chart_Engine SHALL automatically apply Indexed to 100 scale mode by default to normalize price levels for meaningful comparison
14. THE Chart_Engine SHALL provide quick scale presets accessible via keyboard shortcuts: Auto scale (A), Percent scale (%), Indexed scale (I), Locked scale (L), and Logarithmic scale (G)
15. THE Chart_Engine SHALL display the current scale mode in the chart header with a visual indicator (e.g., "% Scale" badge) so users always know which scale mode is active

### Requirement 66: Advanced Time Features

**User Story:** As a day trader, I want second-based timeframes, date range selector, go to date feature, and bar countdown timer, so that I can analyze intraday price action with precision and navigate historical data efficiently.

#### Acceptance Criteria

1. THE Chart_Engine SHALL support second-based timeframes: 1 second, 5 seconds, 15 seconds, and 30 seconds for ultra-short-term intraday analysis
2. THE Chart_Engine SHALL fetch second-based OHLCV data from the Data_Source_Router when second-based timeframes are selected, falling back to minute data if second data is unavailable
3. THE Chart_Engine SHALL provide a date range selector allowing users to specify custom start and end dates for chart display with a calendar picker interface
4. THE Chart_Engine SHALL implement a "Go to Date" feature accessible via keyboard shortcut (G) or toolbar button, opening a date picker that immediately jumps the chart to the selected date when confirmed
5. THE Chart_Engine SHALL display a bar countdown timer in the chart header showing the time remaining until the next candle closes for the current timeframe (e.g., "Next candle in 00:45" for 1-minute chart)
6. WHEN second-based timeframes are active during trading hours (9:00-15:00 ICT), THE Chart_Engine SHALL update the chart in real-time as new second-level data arrives via WebSocket or polling
7. THE Chart_Engine SHALL provide quick date navigation buttons: Today, Yesterday, 1 Week Ago, 1 Month Ago, 1 Year Ago, and All Data, allowing instant jumps to common date ranges
8. THE Chart_Engine SHALL support date range presets in the date range selector: Last 7 Days, Last 30 Days, Last 3 Months, Last 6 Months, Last 1 Year, Year to Date, and Custom
9. WHEN the bar countdown timer reaches zero, THE Chart_Engine SHALL visually indicate the new candle formation with a brief animation or highlight
10. THE Chart_Engine SHALL display the current visible date range in the chart header (e.g., "Jan 15, 2025 - Feb 15, 2025") so users always know what time period they are viewing
11. THE Chart_Engine SHALL support session breaks visualization showing gaps between trading sessions (overnight, weekends) with shaded areas or vertical lines marking session boundaries
12. THE Chart_Engine SHALL provide a time zone selector allowing users to display chart times in different time zones (ICT, UTC, EST, PST) with automatic conversion of all timestamps
13. WHEN second-based timeframes are selected outside trading hours, THE Chart_Engine SHALL display a warning message indicating that second-level data is only available during market hours and show the most recent available data
14. THE Chart_Engine SHALL implement smart date range adjustment: when zooming out beyond available data, automatically extend the date range to show all available historical data
15. THE Chart_Engine SHALL persist the selected timeframe and date range per symbol in local storage so the user's preferred view is restored when returning to the same symbol

### Requirement 67: Trading Visualization on Chart

**User Story:** As a trader, I want to see my buy/sell orders, trading history, position tracking, cost basis lines, profit/loss zones, and trade statistics overlaid directly on the chart, so that I can visualize my trading activity in the context of price action and make better informed decisions.

#### Acceptance Criteria

1. THE Chart_Engine SHALL display buy order markers on the chart as green upward-pointing triangles at the entry price level with the entry date, showing all executed buy transactions from the Transaction_Ledger
2. THE Chart_Engine SHALL display sell order markers on the chart as red downward-pointing triangles at the exit price level with the exit date, showing all executed sell transactions from the Transaction_Ledger
3. THE Chart_Engine SHALL draw horizontal cost basis lines for each active holding showing the weighted average purchase price in a distinct color (e.g., blue dashed line) with a label displaying the cost basis value
4. THE Chart_Engine SHALL display current position tracking for active holdings with a floating panel showing: symbol, quantity held, average cost, current price, unrealized P&L (VND and percentage), and position duration
5. THE Chart_Engine SHALL visualize profit/loss zones by shading the area between the cost basis line and current price: green shading for profit zones (current price > cost basis), red shading for loss zones (current price < cost basis)
6. THE Chart_Engine SHALL connect buy and sell markers with a line when they represent a completed round-trip trade, color-coded green for profitable trades and red for losing trades, with the P&L percentage displayed on the line
7. THE Chart_Engine SHALL display a trade statistics overlay panel showing: total trades on the visible chart, win rate, average profit per trade, average loss per trade, largest win, largest loss, and total P&L for the visible period
8. WHEN a user hovers over a buy or sell marker, THE Chart_Engine SHALL display a detailed tooltip showing: transaction date and time, quantity, price, total value, transaction fees, and notes if any
9. WHEN a user clicks on a trade line connecting buy and sell markers, THE Chart_Engine SHALL highlight the trade and display detailed trade metrics: entry date, exit date, holding period, entry price, exit price, quantity, P&L (VND and percentage), and return on investment
10. THE Chart_Engine SHALL support filtering trade visualizations by date range, trade outcome (profitable, losing, breakeven), and position size, allowing users to focus on specific subsets of trades
11. THE Chart_Engine SHALL display pending limit orders (if order management is implemented) as horizontal dashed lines at the order price level with order type labels (Buy Limit, Sell Limit) and quantity
12. THE Chart_Engine SHALL provide a toggle in the chart toolbar to show/hide trading visualizations: "Show Trades", "Show Cost Basis", "Show P&L Zones", and "Show Statistics", allowing users to declutter the chart when needed
13. THE Chart_Engine SHALL support trade annotations allowing users to add notes to specific trades (e.g., "Bought on breakout signal", "Sold due to stop loss") which are displayed when hovering over trade markers
14. THE Chart_Engine SHALL compute and display trade performance metrics relative to buy-and-hold strategy: showing what the P&L would have been if the user had simply held from the first buy to the current date, comparing actual trading performance against passive holding
15. THE Chart_Engine SHALL synchronize trading visualizations with the Portfolio_Engine in real-time, updating cost basis lines, P&L zones, and statistics immediately when new transactions are recorded
16. THE Chart_Engine SHALL support exporting trade visualizations as annotated chart images (PNG/SVG) including all markers, lines, zones, and statistics for record-keeping or sharing

### Requirement 68: Quick Trading Actions

**User Story:** As an active trader, I want quick trading actions including buy/sell floating buttons on the chart, quick order entry from chart clicks, bracket order visualization, and one-click trading from watchlist, so that I can execute trades rapidly without leaving the chart view or navigating through multiple screens.

#### Acceptance Criteria

1. THE Chart_Engine SHALL display floating Buy and Sell buttons in the bottom-right corner of the chart that remain visible while scrolling or zooming
2. WHEN the Buy button is clicked, THE Chart_Engine SHALL open a quick order entry modal pre-filled with: the current symbol, current market price, and a quantity input field, allowing the user to confirm and record a buy transaction with minimal clicks
3. WHEN the Sell button is clicked, THE Chart_Engine SHALL open a quick order entry modal pre-filled with: the current symbol, current market price, available quantity from holdings, and a quantity input field for partial or full position exit
4. THE Chart_Engine SHALL support click-to-trade functionality where right-clicking on the chart at any price level opens a context menu with options: "Buy at [price]", "Sell at [price]", "Set Alert at [price]", and "Add Note at [price]"
5. THE Chart_Engine SHALL implement bracket order visualization allowing users to define entry price, stop-loss price, and take-profit price by clicking three points on the chart, then displaying all three levels as horizontal lines with labels
6. WHEN a bracket order is visualized, THE Chart_Engine SHALL calculate and display the risk/reward ratio based on the distance between entry and stop-loss vs entry and take-profit
7. THE MyFi_Platform frontend SHALL provide one-click trading actions in the Watchlist component with Buy and Sell icon buttons next to each symbol that open the quick order entry modal pre-filled with the symbol and current price
8. THE quick order entry modal SHALL support keyboard shortcuts for rapid entry: Enter to confirm, Escape to cancel, Tab to move between fields, and arrow keys to adjust quantity
9. THE quick order entry modal SHALL display a transaction preview showing: total cost (for buys), total proceeds (for sells), estimated fees, and impact on portfolio allocation before confirmation
10. THE Chart_Engine SHALL support order templates allowing users to save frequently used order configurations (e.g., "Standard Position: 100 shares, Stop Loss -5%, Take Profit +10%") and apply them with one click
11. THE Chart_Engine SHALL implement position sizing calculator in the quick order entry modal: users input risk amount (VND) and stop-loss percentage, and the calculator suggests the appropriate quantity to risk only the specified amount
12. THE MyFi_Platform frontend SHALL provide a trading hotkey system allowing users to configure custom keyboard shortcuts for common actions: Quick Buy (default: B), Quick Sell (default: S), Close Position (default: X), and Reverse Position (default: R)
13. THE Chart_Engine SHALL support drag-and-drop order placement where users can drag a "Buy" or "Sell" chip from a toolbar and drop it at the desired price level on the chart to create an order at that price
14. THE quick order entry modal SHALL remember the user's last used quantity and order type per symbol, pre-filling these values for faster repeat orders
15. THE Chart_Engine SHALL display a confirmation toast notification after each trade is recorded showing: action (bought/sold), symbol, quantity, price, and a link to view the transaction in the portfolio

### Requirement 69: Watchlist Enhancements

**User Story:** As an investor, I want enhanced watchlist features including multi-column sorting, quick actions from watchlist rows, and drag-and-drop reordering, so that I can organize and interact with my watchlists more efficiently.

#### Acceptance Criteria

1. THE Watchlist component SHALL support multi-column sorting allowing users to sort by: symbol (alphabetical), price (ascending/descending), change percentage (ascending/descending), volume (ascending/descending), and market cap (ascending/descending)
2. THE Watchlist component SHALL display sort indicators (up/down arrows) in column headers showing the current sort column and direction
3. THE Watchlist component SHALL support secondary sorting: when clicking a column header while holding Shift, add that column as a secondary sort criterion (e.g., sort by sector, then by change% within each sector)
4. THE Watchlist component SHALL provide quick action buttons on each watchlist row: Chart icon (opens chart view for the symbol), Trade icon (opens quick order entry), Alert icon (opens alert configuration), and Remove icon (removes symbol from watchlist)
5. THE Watchlist component SHALL support drag-and-drop reordering allowing users to click and drag symbols to rearrange their custom order within the watchlist
6. WHEN a symbol is dragged, THE Watchlist component SHALL display a visual indicator (ghost image or placeholder) showing where the symbol will be dropped
7. THE Watchlist component SHALL persist the custom sort order and reordering in the backend database via the Watchlist_Service so the order is maintained across sessions
8. THE Watchlist component SHALL support bulk actions: users can select multiple symbols using checkboxes and perform actions on all selected symbols: Add Alert, Remove from Watchlist, Move to Another Watchlist, or Export to CSV
9. THE Watchlist component SHALL display additional columns (configurable by user): P/E ratio, P/B ratio, dividend yield, 52-week high/low, average volume, and sector
10. THE Watchlist component SHALL provide column visibility controls allowing users to show/hide specific columns based on their preferences, with settings persisted per user
11. THE Watchlist component SHALL support filtering within the watchlist: users can type in a search box to filter symbols by name or symbol code, or use dropdown filters for sector, exchange, or custom tags
12. THE Watchlist component SHALL display color-coded change indicators: green for positive change, red for negative change, with intensity based on magnitude (darker colors for larger changes)
13. THE Watchlist component SHALL support custom tags allowing users to add labels to symbols (e.g., "High Conviction", "Watch for Breakout", "Dividend Play") and filter by tags
14. THE Watchlist component SHALL provide a compact view mode and a detailed view mode: compact shows only symbol, price, and change; detailed shows all configured columns
15. THE Watchlist component SHALL support exporting the current watchlist view to CSV including all visible columns and current prices for external analysis

### Requirement 70: Indicator Templates

**User Story:** As a technical analyst, I want to save indicator configurations as templates, load saved templates, share templates with others, and access a built-in template library, so that I can quickly apply my preferred indicator setups across different charts and learn from community-shared configurations.

#### Acceptance Criteria

1. THE Chart_Engine SHALL allow users to save the current chart's indicator configuration as a named template including: all active indicators with their parameters, indicator colors, line styles, and pane assignments
2. THE Chart_Engine SHALL provide a "Save as Template" button in the chart toolbar that opens a dialog for naming the template and optionally adding a description
3. THE Chart_Engine SHALL store indicator templates in the backend database per user, making them accessible across all devices and sessions
4. THE Chart_Engine SHALL provide a "Load Template" button in the chart toolbar that displays a list of saved templates with preview thumbnails showing the indicator layout
5. WHEN a template is loaded, THE Chart_Engine SHALL remove all existing indicators and apply the template's indicator configuration, preserving the chart type, timeframe, and drawings
6. THE Chart_Engine SHALL support template categories allowing users to organize templates by strategy type: Trend Following, Mean Reversion, Momentum, Volatility, Volume Analysis, and Custom
7. THE Chart_Engine SHALL provide a built-in template library with preset configurations including: "Trend Trader" (SMA 20/50/200, MACD, ADX), "Momentum Scalper" (RSI, Stochastic, Volume), "Volatility Breakout" (Bollinger Bands, ATR, Volume), "Swing Trader" (EMA 9/21, RSI, MACD), and "Multi-Timeframe" (Daily SMA on hourly chart, Weekly EMA on daily chart)
8. THE Chart_Engine SHALL allow users to edit existing templates: modify the indicator configuration, update the name/description, or delete the template
9. THE Chart_Engine SHALL support template sharing: users can generate a shareable link or export a template file (JSON format) that other users can import
10. THE Chart_Engine SHALL provide a template import function allowing users to import templates from JSON files or shareable links, adding them to their personal template library
11. THE Chart_Engine SHALL display template metadata in the template list: template name, description, number of indicators, creation date, and last modified date
12. THE Chart_Engine SHALL support template favoriting allowing users to mark frequently used templates as favorites for quick access at the top of the template list
13. THE Chart_Engine SHALL provide template preview: when hovering over a template in the list, display a tooltip showing which indicators are included with their parameters
14. THE Chart_Engine SHALL implement template versioning: when a user modifies a loaded template, offer to save as a new version or overwrite the existing template
15. THE Chart_Engine SHALL support community template sharing (if social features are enabled): users can publish templates to a community library where others can browse, rate, and import popular templates

### Requirement 71: Custom Event Marks

**User Story:** As an investor, I want to mark important events on the chart including earnings dates, dividend dates, corporate actions, and custom user annotations, so that I can correlate price movements with significant company events and maintain a visual record of important dates.

#### Acceptance Criteria

1. THE Chart_Engine SHALL automatically display earnings date markers on the chart as vertical lines or icons at the date of earnings announcements, fetched from the Corporate_Action_Service
2. THE Chart_Engine SHALL automatically display ex-dividend date markers on the chart as vertical lines or icons with dividend amount labels, fetched from the Corporate_Action_Service
3. THE Chart_Engine SHALL automatically display corporate action markers on the chart for stock splits, bonus shares, rights issues, and mergers/acquisitions, fetched from the Corporate_Action_Service
4. THE Chart_Engine SHALL allow users to add custom event marks by right-clicking on the chart at any date and selecting "Add Event Mark", then entering event name, description, and optional icon
5. THE Chart_Engine SHALL display event marks as vertical lines with icons at the top of the chart, color-coded by event type: blue for earnings, green for dividends, orange for corporate actions, and purple for custom events
6. WHEN a user hovers over an event mark, THE Chart_Engine SHALL display a tooltip showing: event type, event name, date, description, and relevant details (e.g., dividend amount, split ratio)
7. THE Chart_Engine SHALL provide an event marks panel accessible from the chart toolbar showing a chronological list of all events on the current chart with filters by event type
8. THE Chart_Engine SHALL allow users to show/hide event marks by category: toggle earnings marks, toggle dividend marks, toggle corporate action marks, and toggle custom marks independently
9. THE Chart_Engine SHALL support event mark templates for recurring events: users can create templates like "Quarterly Earnings" that automatically generate marks at regular intervals
10. THE Chart_Engine SHALL persist custom event marks per symbol in the backend database so they are synchronized across devices and sessions
11. THE Chart_Engine SHALL support event mark alerts: users can set alerts to be notified X days before an upcoming event (e.g., "Alert me 3 days before earnings")
12. THE Chart_Engine SHALL display upcoming events in a dedicated "Upcoming Events" panel on the dashboard showing the next 10 events across all watched symbols with countdown timers
13. THE Chart_Engine SHALL support bulk event import allowing users to import a CSV file with event dates and descriptions to create multiple event marks at once
14. THE Chart_Engine SHALL provide event mark statistics showing: number of earnings beats/misses, average price movement on earnings days, and average price movement on ex-dividend days
15. THE Chart_Engine SHALL support event mark linking to news: when an event mark is clicked, display related news articles from the News_Agent for that date and symbol

### Requirement 72: Synchronized Multi-Chart Layout

**User Story:** As a professional trader, I want to display multiple charts in a grid layout with cursor sync, symbol sync, time sync, and drawings sync, so that I can analyze multiple symbols or timeframes simultaneously and maintain consistent analysis across all charts.

#### Acceptance Criteria

1. THE Chart_Engine SHALL support multi-chart layouts with configurable grid arrangements: 1x1 (single chart), 1x2 (two charts side-by-side), 2x1 (two charts stacked), 2x2 (four charts in grid), 1x3, 3x1, 2x3, and 3x2
2. THE MyFi_Platform frontend SHALL provide a layout selector in the chart toolbar allowing users to choose the desired grid arrangement with visual previews of each layout option
3. THE Chart_Engine SHALL implement cursor synchronization across all charts in the layout: when the user moves the crosshair cursor on one chart, the same date/time position is highlighted on all other charts with synchronized vertical lines
4. THE Chart_Engine SHALL support symbol synchronization mode: when enabled, changing the symbol on one chart automatically changes the symbol on all other charts in the layout to the same symbol
5. THE Chart_Engine SHALL support time synchronization mode: when enabled, zooming or panning on one chart automatically applies the same time range and zoom level to all other charts in the layout
6. THE Chart_Engine SHALL support drawings synchronization mode: when enabled, creating, modifying, or deleting a drawing on one chart automatically replicates the action on all other charts in the layout with automatic scaling to fit each chart's price range
7. THE Chart_Engine SHALL allow users to configure each chart in the layout independently: different symbols, different timeframes, different chart types, and different indicator sets when sync modes are disabled
8. THE Chart_Engine SHALL provide sync toggle buttons in the chart toolbar: "Sync Cursor", "Sync Symbol", "Sync Time", and "Sync Drawings", allowing users to enable/disable each sync mode independently
9. THE Chart_Engine SHALL support layout templates allowing users to save multi-chart configurations including: grid arrangement, symbols assigned to each chart, timeframes, indicators, and sync settings
10. THE Chart_Engine SHALL persist the current layout configuration in local storage so the user's multi-chart setup is restored when returning to the chart view
11. THE Chart_Engine SHALL support maximizing individual charts within the layout: clicking a maximize button on any chart temporarily expands it to full screen, with a restore button to return to the grid layout
12. THE Chart_Engine SHALL implement smart layout resizing: when the browser window is resized, automatically adjust chart sizes proportionally to maintain the grid arrangement without overlapping
13. THE Chart_Engine SHALL support chart swapping within the layout: users can drag a chart's header and drop it on another chart's position to swap their locations in the grid
14. THE Chart_Engine SHALL provide layout presets for common use cases: "Multi-Timeframe Analysis" (same symbol on 4 charts with 1H, 4H, 1D, 1W timeframes), "Sector Comparison" (4 charts with different symbols from the same sector), "Correlation Analysis" (2 charts with correlated symbols), and "Custom"
15. THE Chart_Engine SHALL support exporting the entire multi-chart layout as a single image (PNG) with all charts arranged in the grid, useful for sharing analysis or creating reports
16. THE Chart_Engine SHALL display chart identifiers (Chart 1, Chart 2, etc.) in each chart's header when multiple charts are visible, making it easy to reference specific charts
17. THE Chart_Engine SHALL support independent indicator management per chart: adding an indicator to one chart does not affect other charts unless drawings sync is enabled
18. THE Chart_Engine SHALL implement performance optimization for multi-chart layouts: lazy rendering of off-screen charts, throttled updates during rapid panning/zooming, and shared data caching across charts displaying the same symbol

### Requirement 73: Dual-Mode Recommendation Engine

**User Story:** As a Vietnamese stock investor, I want the platform to provide both short-term trading signals and long-term investment recommendations with structured outputs (entry, SL, TP, confidence, reasoning), so that I can make informed decisions for both trading and investing strategies.

#### Acceptance Criteria

1. THE Recommendation_Engine SHALL support two distinct modes: Trading mode (short-term, days to weeks) and Investment mode (long-term, weeks to months)
2. WHEN the Recommendation_Engine operates in Trading mode, THE Recommendation_Engine SHALL produce a Trading_Signal containing: symbol, signal direction (long or short), exact entry price, stop-loss price, take-profit price, risk/reward ratio, confidence score (0–100), expected holding period (days), and a structured reasoning section
3. WHEN the Recommendation_Engine operates in Investment mode, THE Recommendation_Engine SHALL produce an Investment_Signal containing: symbol, entry price zone (low price–high price), target price, suggested holding period (weeks or months), confidence score (0–100), key fundamental metrics (P/E, P/B, ROE, revenue growth, profit growth), and a structured reasoning section combining fundamental and technical factors
4. THE Recommendation_Engine SHALL compute stop-loss prices for Trading_Signal outputs using ATR-based calculation: stop-loss = entry price minus (ATR(14) multiplied by a configurable multiplier, default 1.5 for long positions)
5. THE Recommendation_Engine SHALL compute take-profit prices for Trading_Signal outputs using risk-reward ratio: take-profit = entry price plus ((entry price minus stop-loss price) multiplied by the target risk-reward ratio, default 2.0)
6. THE Recommendation_Engine SHALL include a risk/reward ratio in each Trading_Signal, computed as the distance from entry to take-profit divided by the distance from entry to stop-loss
7. THE Recommendation_Engine SHALL assign a confidence score (0–100) to each recommendation based on the weighted combination of: technical signal alignment (30%), volume confirmation (20%), money flow direction (20%), sector momentum (15%), and fundamental strength (15% for Investment mode) or price action clarity (15% for Trading mode)
8. THE Recommendation_Engine SHALL include a structured reasoning section in each signal containing: primary signal trigger, supporting indicators, risk factors, and sector context
9. IF the Recommendation_Engine cannot produce a recommendation with a confidence score of 40 or above, THEN THE Recommendation_Engine SHALL omit that stock from the output and log the reason
10. THE Recommendation_Engine SHALL validate that the computed stop-loss price is within a reasonable range (no more than 10% from entry for Trading mode, no more than 20% from entry zone midpoint for Investment mode) and flag anomalies

### Requirement 74: Stock Universe Scanner and Ranking

**User Story:** As a Vietnamese stock investor, I want the platform to proactively scan the entire VN stock universe and rank stocks by signal strength, so that I automatically discover the best trading and investment opportunities without manually analyzing each stock.

#### Acceptance Criteria

1. THE Stock_Scanner SHALL scan all actively traded stocks on HOSE, HNX, and UPCOM exchanges by fetching the full listing from the Data_Source_Router
2. THE Stock_Scanner SHALL run on a configurable schedule (default: once per day at 16:00 ICT after market close for full scan, and every 2 hours during trading hours 9:00–15:00 ICT for incremental scan of top 100 stocks by volume)
3. THE Stock_Scanner SHALL compute a composite signal strength score (0–100) for each stock based on: technical indicator alignment, volume anomaly detection, money flow direction, sector momentum, and price action quality
4. THE Stock_Scanner SHALL rank all scanned stocks by composite signal strength score in descending order and retain the top 50 stocks for Trading mode and top 30 stocks for Investment mode
5. THE Stock_Scanner SHALL filter out stocks with average daily trading volume below 100,000 shares (20-day average) to ensure sufficient liquidity for actionable recommendations
6. THE Stock_Scanner SHALL filter out stocks with market capitalization below a configurable threshold (default: 500 billion VND) for Investment mode to focus on fundamentally significant companies
7. WHEN the Stock_Scanner completes a scan cycle, THE Stock_Scanner SHALL pass the ranked stock lists to the Recommendation_Engine for detailed signal generation in both Trading and Investment modes
8. THE Stock_Scanner SHALL persist scan results (symbol, composite score, component scores, scan timestamp) in the database for historical tracking and performance evaluation
9. IF the Stock_Scanner fails to fetch data for a symbol during a scan cycle, THEN THE Stock_Scanner SHALL skip that symbol, log the failure, and continue scanning the remaining symbols
10. THE Stock_Scanner SHALL expose a REST API endpoint returning the latest scan results with filtering by mode (trading/investment), minimum score, sector, and exchange

### Requirement 75: Money Flow Index (MFI) Indicator

**User Story:** As a technical analyst, I want the platform to compute and display the Money Flow Index (MFI) as an additional indicator combining price and volume, so that I can measure buying and selling pressure more accurately than price-only indicators.

#### Acceptance Criteria

1. THE Analysis_Agent SHALL compute the Money_Flow_Index using the standard 14-period formula: typical price = (high + low + close) / 3, raw money flow = typical price × volume, money ratio = positive money flow sum / negative money flow sum over the period, MFI = 100 − (100 / (1 + money ratio))
2. THE Chart_Engine SHALL support MFI as the 22nd technical indicator, rendered as an oscillator in a separate pane below the price chart with a range of 0–100
3. THE Chart_Engine SHALL display horizontal reference lines on the MFI pane at the 80 (overbought) and 20 (oversold) levels with configurable thresholds
4. THE Chart_Engine SHALL allow the user to configure the MFI period (default 14) and overbought/oversold thresholds (default 80/20)
5. THE Analysis_Agent SHALL classify MFI readings as: overbought (above 80), oversold (below 20), bullish (rising from below 20), bearish (falling from above 80), or neutral (between 20 and 80 with no divergence)
6. THE Analysis_Agent SHALL detect MFI-price divergence: bullish divergence when price makes a lower low but MFI makes a higher low, and bearish divergence when price makes a higher high but MFI makes a lower high
7. WHEN MFI divergence is detected, THE Analysis_Agent SHALL include the divergence type and magnitude in the analysis summary as a potential reversal signal
8. THE Analysis_Agent SHALL include MFI signals in the composite signal summary alongside the other 21 indicators when computing the overall technical outlook

### Requirement 76: Institutional and Foreign Flow Tracking

**User Story:** As a Vietnamese stock investor, I want to track net foreign and institutional buying/selling activity for each stock, so that I can identify smart money movement and align my trades with informed capital flows.

#### Acceptance Criteria

1. THE Analysis_Agent SHALL fetch net foreign investor buy/sell volume and net institutional investor buy/sell volume for each analyzed stock from the price board data via the Data_Source_Router
2. THE Analysis_Agent SHALL compute Smart_Money_Flow as the sum of net foreign buy volume and net institutional buy volume over configurable periods: 1-day, 5-day, 10-day, and 20-day rolling windows
3. THE Analysis_Agent SHALL classify Smart_Money_Flow for each period as: strong inflow (net buy exceeding 2x average daily volume contribution), moderate inflow (net buy exceeding 1x average), neutral (net buy/sell within 0.5x average), moderate outflow (net sell exceeding 1x average), or strong outflow (net sell exceeding 2x average)
4. THE Analysis_Agent SHALL detect Smart_Money_Flow trend changes: when the 5-day flow direction reverses from outflow to inflow (or vice versa), flag the reversal as a potential signal
5. THE MyFi_Platform frontend SHALL display a Smart Money Flow panel on the stock detail view showing: net foreign buy/sell (shares and VND value), net institutional buy/sell (shares and VND value), flow classification for each period, and a 20-day flow trend chart
6. THE Stock_Scanner SHALL incorporate Smart_Money_Flow direction as a component of the composite signal strength score, weighting strong inflow positively and strong outflow negatively
7. IF foreign or institutional flow data is unavailable for a symbol (e.g., UPCOM stocks with limited reporting), THEN THE Analysis_Agent SHALL omit the Smart_Money_Flow component from the analysis and note the data gap
8. THE Analysis_Agent SHALL detect accumulation signals by identifying the combination of: Smart_Money_Flow showing sustained inflow over 10 or more trading days while price remains in a consolidation range (less than 5% movement)

### Requirement 77: Volume Anomaly Detection

**User Story:** As a trader, I want the platform to automatically detect unusual volume spikes and patterns, so that I can identify potential breakout or breakdown setups before they fully develop.

#### Acceptance Criteria

1. THE Analysis_Agent SHALL compute volume anomaly scores by comparing each day's volume against the 20-day simple moving average of volume, expressed as a multiple (e.g., 2.5x average)
2. THE Analysis_Agent SHALL classify volume anomalies into tiers: notable (1.5x–2x average), significant (2x–3x average), and extreme (above 3x average)
3. THE Analysis_Agent SHALL correlate volume anomalies with price action: volume spike on an up day is classified as bullish volume, volume spike on a down day is classified as bearish volume, and volume spike with a small price change (less than 1%) is classified as churning
4. THE Analysis_Agent SHALL detect volume dry-up patterns: 5 or more consecutive days with volume below 0.5x the 20-day average, which may indicate an impending volatility expansion
5. THE Analysis_Agent SHALL detect volume climax patterns: a single day with volume exceeding 3x the 20-day average accompanied by a price reversal (close in the opposite direction of the open), indicating potential exhaustion
6. WHEN a volume anomaly of significant or extreme tier is detected, THE Analysis_Agent SHALL include the anomaly in the analysis summary with: anomaly tier, volume multiple, price action correlation, and interpretation
7. THE Stock_Scanner SHALL use volume anomaly detection as a component of the composite signal strength score, weighting significant bullish volume anomalies positively for Trading mode candidates
8. THE Chart_Engine SHALL visually highlight volume bars that exceed 1.5x the 20-day average with a distinct color (default: bright orange for bullish volume anomaly, bright purple for bearish volume anomaly)

### Requirement 78: Price Action Analysis

**User Story:** As a technical trader, I want the platform to analyze price action patterns including support/resistance levels, trend structure, and candlestick patterns, so that the recommendation engine can identify high-probability entry and exit points.

#### Acceptance Criteria

1. THE Analysis_Agent SHALL identify support levels by detecting price zones where the stock has bounced upward at least 2 times within the last 120 trading days, with touches within a 2% price tolerance band
2. THE Analysis_Agent SHALL identify resistance levels by detecting price zones where the stock has reversed downward at least 2 times within the last 120 trading days, with touches within a 2% price tolerance band
3. THE Analysis_Agent SHALL classify trend structure as: uptrend (higher highs and higher lows over the last 20 days), downtrend (lower highs and lower lows over the last 20 days), or sideways (price contained within a range with no clear higher highs/lows pattern)
4. THE Analysis_Agent SHALL detect the following candlestick patterns on daily timeframe: hammer, inverted hammer, bullish engulfing, bearish engulfing, doji, morning star, evening star, three white soldiers, and three black crows
5. THE Analysis_Agent SHALL assign a significance score (1–5) to each detected candlestick pattern based on: pattern location relative to support/resistance (higher score near key levels), volume confirmation (higher score with above-average volume), and trend context (higher score for reversal patterns at trend extremes)
6. THE Analysis_Agent SHALL include detected price action patterns in the analysis summary containing: identified support/resistance levels with price values, trend structure classification, detected candlestick patterns with significance scores, and whether price is near a key level
7. THE Recommendation_Engine SHALL use support/resistance levels from the Analysis_Agent to refine entry prices (near support for long entries, near resistance for short entries) and validate stop-loss placement (below support for longs, above resistance for shorts)
8. THE Chart_Engine SHALL optionally display auto-detected support and resistance levels as horizontal dashed lines on the price chart with labels showing the price level and number of touches

### Requirement 79: Sector Rotation and Momentum Analysis for Recommendations

**User Story:** As an investor, I want the recommendation engine to factor in sector rotation dynamics and sector momentum, so that recommendations favor stocks in sectors with improving capital flows and avoid sectors losing momentum.

#### Acceptance Criteria

1. THE Analysis_Agent SHALL compute sector momentum scores for each ICB sector by measuring the sector index performance over 1-week, 1-month, and 3-month periods and ranking sectors from strongest to weakest
2. THE Analysis_Agent SHALL detect sector rotation by comparing the change in sector rankings between the current period and the prior period: sectors moving up 2 or more rank positions are classified as rotation-in, sectors moving down 2 or more positions are classified as rotation-out
3. THE Recommendation_Engine SHALL boost the confidence score of Trading_Signal and Investment_Signal outputs for stocks belonging to sectors classified as rotation-in by a configurable amount (default: +10 points)
4. THE Recommendation_Engine SHALL reduce the confidence score of Trading_Signal and Investment_Signal outputs for stocks belonging to sectors classified as rotation-out by a configurable amount (default: -10 points)
5. THE Stock_Scanner SHALL prioritize stocks from sectors with positive momentum (top 3 ranked sectors) when generating the candidate list for both Trading and Investment modes
6. THE MyFi_Platform frontend SHALL display a Sector Rotation Dashboard showing: sector rankings by momentum score, rotation-in and rotation-out classifications, sector performance heatmap over 1-week/1-month/3-month periods, and top stocks within each sector by signal strength
7. THE Analysis_Agent SHALL include sector rotation context in each stock analysis: whether the stock's sector is in a rotation-in or rotation-out phase, the sector's momentum rank, and the sector's performance relative to the VN-Index over the same period

### Requirement 80: Fundamental Analysis for Investment Mode

**User Story:** As a long-term investor, I want the recommendation engine to perform deep fundamental analysis when generating investment recommendations, so that I can identify undervalued stocks with strong growth potential for long-term holding.

#### Acceptance Criteria

1. WHEN the Recommendation_Engine operates in Investment mode, THE Analysis_Agent SHALL evaluate the following fundamental metrics for each candidate stock: P/E ratio, P/B ratio, EV/EBITDA, ROE, ROA, revenue growth (YoY and QoQ), net profit growth (YoY and QoQ), dividend yield, debt-to-equity ratio, current ratio, and free cash flow margin
2. THE Analysis_Agent SHALL compare each fundamental metric against the median value of stocks within the same ICB_Sector and classify each metric as: significantly undervalued (below 25th percentile for valuation metrics), undervalued (25th–50th percentile), fair value (50th–75th percentile), or overvalued (above 75th percentile)
3. THE Analysis_Agent SHALL compute a fundamental quality score (0–100) based on the weighted combination of: valuation attractiveness (P/E, P/B, EV/EBITDA relative to sector — 30%), profitability (ROE, ROA — 25%), growth trajectory (revenue and profit growth — 25%), and financial health (debt-to-equity, current ratio — 20%)
4. THE Recommendation_Engine SHALL use the fundamental quality score as the primary ranking factor for Investment mode candidates, with technical and money flow signals as secondary confirmation
5. THE Investment_Signal reasoning section SHALL include: key fundamental strengths and weaknesses, comparison against sector medians for each metric, growth trajectory assessment, and valuation conclusion (undervalued, fair, or overvalued)
6. THE Recommendation_Engine SHALL suggest a holding period for each Investment_Signal based on the investment thesis: growth-driven recommendations suggest 3–6 months, value-driven recommendations suggest 6–12 months, and dividend-driven recommendations suggest 12+ months
7. IF fundamental data (financial statements) is unavailable or older than 6 months for a candidate stock, THEN THE Recommendation_Engine SHALL exclude that stock from Investment mode recommendations and log the data gap

### Requirement 81: Recommendation Output and Display

**User Story:** As a user, I want trading and investment recommendations displayed in a clear, actionable format on the frontend with all relevant details, so that I can quickly evaluate and act on each recommendation.

#### Acceptance Criteria

1. THE MyFi_Platform frontend SHALL display a Recommendations tab with two sub-views: Trading Signals and Investment Signals, each showing the latest recommendations from the Recommendation_Engine
2. THE Trading Signals view SHALL display each Trading_Signal as a card containing: symbol, signal direction (long/short with color coding), entry price, stop-loss price, take-profit price, risk/reward ratio, confidence score (with color gradient: green above 70, yellow 50–70, orange 40–50), expected holding period, and a collapsible reasoning section
3. THE Investment Signals view SHALL display each Investment_Signal as a card containing: symbol, entry price zone (low–high), target price, suggested holding period, confidence score, key fundamental metrics (P/E, P/B, ROE, growth rates), and a collapsible reasoning section with fundamental analysis
4. THE MyFi_Platform frontend SHALL allow sorting recommendations by: confidence score, risk/reward ratio (Trading), fundamental quality score (Investment), sector, and signal date
5. THE MyFi_Platform frontend SHALL allow filtering recommendations by: minimum confidence score, sector, exchange, mode (trading/investment), and signal direction
6. WHEN the user clicks a recommendation card, THE MyFi_Platform frontend SHALL navigate to the stock's chart view with the entry price, stop-loss, and take-profit levels pre-drawn as horizontal lines
7. THE MyFi_Platform frontend SHALL display a recommendation summary banner at the top of the Recommendations tab showing: total active signals count, average confidence score, sector distribution of current signals, and timestamp of the last scan
8. THE MyFi_Platform backend SHALL expose REST API endpoints for: listing recommendations with pagination and filters, fetching a single recommendation detail, and fetching recommendation performance history
9. THE Recommendation_Engine SHALL persist each generated Trading_Signal and Investment_Signal in the database with a unique identifier, generation timestamp, and status (active, expired, stopped-out, target-hit)
10. THE Recommendation_Engine SHALL automatically update signal status: mark as stopped-out when the stock price reaches the stop-loss level, mark as target-hit when the stock price reaches the take-profit or target price, and mark as expired when the expected holding period elapses without either level being reached

### Requirement 82: Recommendation Performance Tracking

**User Story:** As an investor, I want to track the historical performance of the recommendation engine's signals, so that I can evaluate the system's accuracy and adjust my trust in its recommendations over time.

#### Acceptance Criteria

1. THE Recommendation_Engine SHALL track the outcome of each Trading_Signal by recording the stock's price at: 1-day, 3-day, 5-day, and at signal close (stop-loss hit, take-profit hit, or expiry) after signal generation
2. THE Recommendation_Engine SHALL track the outcome of each Investment_Signal by recording the stock's price at: 1-week, 2-week, 1-month, 2-month, and 3-month intervals after signal generation
3. THE Recommendation_Engine SHALL compute aggregate performance metrics for Trading mode: total signals generated, win rate (percentage of signals hitting take-profit before stop-loss), average risk/reward achieved, average holding period, and profit factor (total gains / total losses)
4. THE Recommendation_Engine SHALL compute aggregate performance metrics for Investment mode: total signals generated, percentage of signals reaching target price within the suggested holding period, average return at holding period end, and average confidence score of winning vs losing signals
5. THE MyFi_Platform frontend SHALL display a Recommendation Performance dashboard showing: win rate over time (line chart), cumulative return if all signals were followed, performance breakdown by sector, performance breakdown by confidence score range, and comparison of high-confidence (above 70) vs low-confidence (40–70) signal outcomes
6. THE Recommendation_Engine SHALL use historical performance data to calibrate confidence scores: if signals in a particular sector or pattern type consistently underperform, THE Recommendation_Engine SHALL adjust the confidence weighting for that category
7. THE Recommendation_Audit_Log SHALL record each recommendation signal with full inputs (technical indicators, money flow data, fundamental data, sector context) and outputs (signal details) for transparency and debugging
8. THE MyFi_Platform frontend SHALL display a performance summary on each recommendation card showing the engine's historical accuracy for similar signals (same sector, similar confidence range)

### Requirement 83: SwingMax Signal

**User Story:** As a swing trader, I want AI-generated swing trading signals with staged exit guidance and configurable stop-loss, so that I can act on high-conviction opportunities with a clear risk management plan.

#### Acceptance Criteria

1. THE SwingMax_Engine SHALL generate SwingMax_Signal records for symbols on HOSE, HNX, and UPCOM with average daily trading value (ADTV) ≥ 5 billion VND over the prior 20 sessions, using the existing `SignalEngine.ScanMarket` pipeline extended with swing-specific parameters.
2. THE SwingMax_Engine SHALL compute per signal: entry price (VND), stop-loss price (ATR-based, default −15% from entry, configurable −5% to −25%), Sell_Target_1 (T1), Sell_Target_2 (T2), unrealized return (%), potential gain to T2 (%), Signal_Status (Opening / Closed), and generation timestamp.
3. THE SwingMax_Engine SHALL provide staged exit guidance per signal: sell 30% of position at T1, sell 30% at T2, and trail the remaining 40% with a trailing stop equal to the ATR value at T2 hit.
4. WHEN a user views the SwingMax Signal page, THE SwingMax_Engine SHALL return signals separated into two tabs: Opening (Signal_Status = opening or partial_exit) and Closed (Signal_Status = closed), ordered by Conviction_Score descending within each tab.
5. THE SwingMax_Engine SHALL display a header stats row containing: overall win rate (% of closed signals where exit price > entry price), annualized return (CAGR from closed signals), and count of signals generated in the current calendar day (ICT).
6. WHEN a user customises the stop-loss percentage, THE SwingMax_Engine SHALL accept a value between −5% and −25% and recompute stop-loss price for all opening signals without regenerating the underlying scan.
7. WHEN the current market price of an opening signal crosses below the stop-loss price, THE SwingMax_Engine SHALL automatically update Signal_Status to closed and record the exit price and exit timestamp.
8. WHEN the current market price of an opening signal reaches T1, THE SwingMax_Engine SHALL update Signal_Status to partial_exit and record the T1 hit timestamp.
9. IF a SwingMax_Signal has been in opening status for more than 30 calendar days without hitting T1 or stop-loss, THEN THE SwingMax_Engine SHALL update Signal_Status to closed with exit reason `timeout`.
10. WHEN a symbol's foreign ownership approaches the FOL limit such that (fol_limit − current_ownership) ≤ 5%, THE SwingMax_Engine SHALL display a FOL warning badge on the signal card.
11. WHEN a symbol's current foreign ownership ≥ fol_limit, THE SwingMax_Engine SHALL exclude that symbol from new SwingMax_Signal generation.
12. THE SwingMax_Engine SHALL support a maximum of 5 simultaneously active Opening signals at any time.

---

### Requirement 84: SwingMax Portfolio

**User Story:** As an investor, I want an AI-curated portfolio of up to 5 swing positions with automatic replacement and portfolio-level metrics, so that I can follow a managed set of positions without manually selecting from the full signal list.

#### Acceptance Criteria

1. THE SwingMax_Engine SHALL maintain a SwingMax_Portfolio containing at most 5 simultaneously active SwingMax_Signal positions selected by highest Conviction_Score.
2. WHEN selecting candidates for the SwingMax_Portfolio, THE SwingMax_Engine SHALL require: ADTV ≥ 5 billion VND over the prior 20 sessions, and at least one positive fundamental indicator (ROE > 0 or profit growth > 0), using the existing `LiquidityFilter` whitelist check.
3. THE SwingMax_Engine SHALL track portfolio-level performance metrics: total unrealized return (%), total realised return (%), portfolio win rate (% of closed positions with positive return), and average holding days.
4. WHEN a SwingMax_Portfolio position is closed (stop-loss, T2 hit, or timeout), THE SwingMax_Engine SHALL automatically evaluate the next highest-Conviction_Score signal not already in the portfolio and add it if it passes liquidity and fundamental checks.
5. THE SwingMax_Engine SHALL NOT add a new position to the SwingMax_Portfolio if the symbol already has an active position in the portfolio.
6. WHEN a user views the SwingMax Portfolio page, THE SwingMax_Engine SHALL display each position with: symbol, entry date, entry price (VND), current price (VND), unrealized return (%), Signal_Status, T1, T2, stop-loss, and staged exit guidance.

---

### Requirement 85: AI Stock Picker

**User Story:** As a day trader, I want a pre-market list of up to 5 intraday buy candidates with AI reasoning, so that I can execute a simple intraday strategy with AI-backed selection each trading day.

#### Acceptance Criteria

1. THE AI_Stock_Picker SHALL generate a daily pick list of up to 5 symbols by 08:45 ICT on each trading day, using overnight price action, pre-market foreign flow data, and the existing `SignalEngine` momentum and volume factors.
2. THE AI_Stock_Picker SHALL only select symbols with ADTV ≥ 10 billion VND and Conviction_Score ≥ 60.
3. THE AI_Stock_Picker SHALL include per pick: symbol, exchange, suggested entry price range (low–high VND), intraday target price (VND), intraday stop-loss price (VND), expected return (%), and AI reasoning summary (≤ 100 words).
4. WHEN the trading session ends at 15:00 ICT, THE AI_Stock_Picker SHALL mark all open picks as closed, record the ATC (At-The-Close) price as the exit price, and compute the realised return for each pick.
5. THE AI_Stock_Picker SHALL maintain a historical accuracy log recording per pick: entry price, exit price, realised return, and whether the pick was profitable.
6. THE AI_Stock_Picker SHALL display a rolling accuracy metric: win rate (%) and average return (%) over the trailing 30 trading days.
7. IF the AI_Stock_Picker cannot generate 5 qualifying picks due to insufficient liquidity or no signals above minimum Conviction_Score, THEN THE AI_Stock_Picker SHALL generate as many qualifying picks as available and display the count clearly.

---

### Requirement 86: Whale Auto Tracker

**User Story:** As a trader, I want to track unusual volume events, block trades, and net foreign flows across Vietnamese exchanges, so that I can follow institutional and large investor activity in real time.

#### Acceptance Criteria

1. THE Whale_Tracker SHALL detect unusual volume events by comparing each symbol's current session volume against its 20-session average volume; a volume ratio ≥ 2.0 SHALL be classified as an unusual volume event.
2. THE Whale_Tracker SHALL detect Block_Trade events (single matched trade value ≥ 1 billion VND) from intraday order flow data via the vnstock-go VCI connector.
3. THE Whale_Tracker SHALL fetch Net_Foreign_Flow data for all HOSE symbols from the vnstock-go VCI connector and display per symbol: net foreign buy value (VND), net foreign sell value (VND), net flow (VND), and flow direction (net buy / net sell / neutral).
4. WHEN a Whale_Tracker event is detected (unusual volume or block trade), THE Whale_Tracker SHALL create an alert record containing: symbol, event type, event value, timestamp, and current price.
5. THE Whale_Tracker SHALL display a ranked list of symbols sorted by absolute Net_Foreign_Flow descending, filterable by exchange (HOSE / HNX / UPCOM) and flow direction (net buy / net sell).
6. THE Whale_Tracker SHALL display a ranked list of unusual volume events for the current session, sorted by volume ratio descending.
7. THE Whale_Tracker SHALL display a block trade feed showing the 50 most recent Block_Trade events for the current session with: symbol, trade value (VND), price, time, and buyer/seller side if available.
8. WHILE the trading session is active (09:00–15:00 ICT), THE Whale_Tracker SHALL refresh unusual volume and block trade data every 5 minutes.
9. IF the vnstock-go VCI connector does not provide intraday block trade data for a symbol, THEN THE Whale_Tracker SHALL omit that symbol from the block trade feed without returning an error to the user.
10. IF the data source returns stale data, THEN THE Whale_Tracker SHALL display the last known data with a stale indicator and the timestamp of the last successful refresh.

---

### Requirement 87: Daytrading Center

**User Story:** As a day trader, I want a dedicated intraday workspace with real-time signals filtered to high-momentum setups and a session summary, so that I can manage my intraday trades in one place during Vietnamese trading hours.

#### Acceptance Criteria

1. THE Daytrading_Center SHALL display intraday signals generated by the existing `SignalEngine` filtered to signals with momentum factor score ≥ 70 AND volume factor score ≥ 60, refreshed every 5 minutes during trading session hours (09:00–15:00 ICT).
2. THE Daytrading_Center SHALL display the current trading session phase: Pre-open (before 09:00 ICT), Continuous trading (09:00–14:30 ICT), ATC (14:30–15:00 ICT), or Closed (after 15:00 ICT).
3. THE Daytrading_Center SHALL display a session summary panel showing: number of intraday signals generated today, number of signals that hit their intraday target, number that hit stop-loss, and session win rate.
4. WHILE the trading session is active, THE Daytrading_Center SHALL highlight symbols where the current price has moved more than 3% from the session open price.
5. WHEN the trading session is outside 09:00–15:00 ICT, THE Daytrading_Center SHALL display the page in a disabled/read-only state indicating the session is closed, showing the last session's summary.

---

### Requirement 88: Patterns Detection

**User Story:** As a technical analyst, I want to see detected chart patterns with confidence scores and historical accuracy, so that I can quickly identify actionable setups across the market.

#### Acceptance Criteria

1. THE Pattern_Detection_Page SHALL expose all pattern observations produced by the existing `PatternDetector` (accumulation, distribution, breakout) for the current scan cycle, sorted by confidence score descending.
2. THE Pattern_Detection_Page SHALL display per pattern card: symbol, exchange, pattern type, confidence score (0–100), price at detection (VND), detection timestamp, and a brief supporting data summary.
3. THE Pattern_Detection_Page SHALL support filtering by: pattern type (accumulation / distribution / breakout), minimum confidence score (slider 0–100), and exchange (HOSE / HNX / UPCOM).
4. WHEN a pattern was detected less than 60 minutes ago and has confidence score ≥ 70, THE Pattern_Detection_Page SHALL display a "New" badge on the pattern card.
5. THE Pattern_Detection_Page SHALL display historical pattern accuracy metrics sourced from the existing `KnowledgeBase`: for each pattern type, show total observations, success rate, and average 7-day return.
6. WHEN a user clicks a pattern card, THE Pattern_Detection_Page SHALL open the symbol's daily chart in the existing `Chart_Engine` with a visual annotation marking the detection date and price level.

---

### Requirement 89: AI Screener with Conviction Score

**User Story:** As an investor, I want the stock screener to layer an AI conviction score on top of fundamental filters, so that I can find stocks that are both fundamentally sound and technically ready.

#### Acceptance Criteria

1. THE AI_Screener SHALL expose all existing `ScreenerService` fundamental filters (P/E, P/B, ROE, ROA, revenue growth, profit growth, dividend yield, debt/equity, market cap, EV/EBITDA, sector, exchange) unchanged.
2. THE AI_Screener SHALL compute an AI_Conviction_Score (0–100, clamped) for each screener result as: (SignalEngine composite score × 0.50) + (PatternDetector highest-confidence pattern score × 0.30) + (fundamental quality score derived from ROE and profit growth × 0.20).
3. WHEN a user runs a screener query, THE AI_Screener SHALL return results sorted by AI_Conviction_Score descending by default, with the option to sort by any individual fundamental metric.
4. THE AI_Screener SHALL display per result: symbol, exchange, sector, AI_Conviction_Score, P/E, P/B, ROE, profit growth, market cap (VND), and the primary signal reason (top factor driving the AI_Conviction_Score).
5. THE AI_Screener SHALL allow users to filter results by minimum AI_Conviction_Score (slider 0–100).
6. THE AI_Screener SHALL display a "Top AI Picks" section showing the 10 highest AI_Conviction_Score stocks that pass the current filter set, regardless of pagination.
7. WHEN a user saves a screener preset, THE AI_Screener SHALL persist the full filter set including the minimum AI_Conviction_Score threshold using the existing `ScreenerService.SavePreset` mechanism.

---

### Requirement 90: Stock Monitor

**User Story:** As an investor, I want a dedicated monitoring page that shows all active price alerts, pattern alerts, whale alerts, and signal alerts in one unified feed, so that I never miss an actionable event.

#### Acceptance Criteria

1. THE Stock_Monitor SHALL expose the existing `AlertService` through a dedicated UI page, displaying all active alerts in a unified feed categorised by type: price alert, pattern alert (from `PatternDetector`), whale alert (from `Whale_Tracker`), and signal alert (from `SwingMax_Engine`).
2. THE Stock_Monitor SHALL display per alert: symbol, alert type, trigger condition, trigger value, current price (VND), timestamp, and status (active / acknowledged / expired).
3. WHEN a user acknowledges an alert, THE Stock_Monitor SHALL mark it as acknowledged and move it to a separate "Acknowledged" section, retaining it for 7 days before automatic deletion.
4. THE Stock_Monitor SHALL allow users to create price alerts by specifying: symbol, condition (price above / price below / % change above / % change below), threshold value, and optional notification note.
5. WHEN a price alert condition is met, THE Stock_Monitor SHALL trigger the existing `AlertService` notification pipeline and update the alert status to `triggered`.
6. THE Stock_Monitor SHALL display a summary header showing: total active alerts, alerts triggered today, and alerts triggered this week.
7. THE Stock_Monitor SHALL support filtering the alert feed by alert type and by symbol.
8. IF a price alert has not been triggered within 30 days of creation, THEN THE Stock_Monitor SHALL mark it as `expired` and move it to the Acknowledged section.

---

### Requirement 91: QuantAI Alpha

**User Story:** As a long-term investor, I want a quantitative factor model that ranks stocks by momentum, value, quality, and growth with adjustable weights, so that I can identify alpha-generating opportunities beyond pure technical signals.

#### Acceptance Criteria

1. THE QuantAI_Alpha SHALL implement a Factor_Model scoring each stock on four factors: momentum (3-month and 12-month price return, earnings momentum), value (P/E, P/B, EV/EBITDA percentile within sector), quality (ROE, debt/equity, profit growth consistency), and growth (revenue growth YoY, EPS growth YoY).
2. THE QuantAI_Alpha SHALL compute each factor score using percentile ranking within the universe of symbols from the `LiquidityFilter` whitelist.
3. THE QuantAI_Alpha SHALL compute a composite alpha score (0–100) as a weighted sum: momentum 30%, value 25%, quality 25%, growth 20% (default weights).
4. THE QuantAI_Alpha SHALL produce a weekly alpha pick list of the top 10 stocks by composite alpha score, refreshed every Monday before 08:00 ICT.
5. THE QuantAI_Alpha SHALL display per pick: symbol, exchange, sector, composite alpha score, individual factor scores, current price (VND), and a one-sentence AI-generated rationale.
6. THE QuantAI_Alpha SHALL allow users to adjust factor weights (momentum, value, quality, growth) within 0–100% with the constraint that all four weights sum to 100%, and recompute the alpha score in real time.
7. THE QuantAI_Alpha SHALL track historical pick performance: for each past weekly pick list, record the 1-week, 1-month, and 3-month returns for each symbol and compute the average return of the pick list.
8. IF fundamental data (P/E, ROE, revenue growth) is unavailable for a symbol, THEN THE QuantAI_Alpha SHALL exclude that symbol from the Factor_Model computation and omit it from the pick list.

---

### Requirement 92: Backtesting Playground

**User Story:** As a strategy developer, I want a visual backtesting interface with an equity curve, trade log, and stop-loss sensitivity panel, so that I can validate strategies against historical Vietnamese market data before trading them live.

#### Acceptance Criteria

1. THE Backtest_Playground SHALL expose the existing `BacktestEngine.RunBacktest` and `BacktestEngine.GetPresetStrategies` functions through a dedicated frontend page with a visual strategy builder.
2. THE Backtest_Playground SHALL render the backtest Equity_Curve as a line chart using lightweight-charts, with the VN-Index benchmark overlaid for comparison over the selected backtest date range.
3. THE Backtest_Playground SHALL display result metrics: total return (%), annualized return (%), win rate (%), max drawdown (%), Sharpe ratio, number of trades, average holding days, average win (%), and average loss (%).
4. THE Backtest_Playground SHALL provide a trade log table listing each simulated trade with: entry date, exit date, entry price (VND), exit price (VND), return (%), holding days, and exit reason (signal / stop_loss / take_profit / end_of_data / timeout).
5. THE Backtest_Playground SHALL display a stop-loss sensitivity panel showing how total return changes as the stop-loss percentage varies from −3% to −15% in 1% increments, using the current strategy and symbol.
6. THE Backtest_Playground SHALL include the three preset strategies from `BacktestEngine.GetPresetStrategies`: RSI Oversold Bounce, MACD Crossover, and Bollinger Band Squeeze, selectable from a dropdown.
7. IF the selected symbol has fewer than 60 trading days of historical data available, THEN THE Backtest_Playground SHALL display an error message and prevent the backtest from running.
8. WHEN a user runs a backtest, THE Backtest_Playground SHALL display a progress indicator and return results within 30 seconds for date ranges up to 3 years of daily data.

---

### Requirement 93: Market Movers

**User Story:** As a trader, I want to see the top 20 gainers, losers, and most-active stocks across all Vietnamese exchanges with a sector heatmap, so that I can quickly identify where market momentum is concentrated.

#### Acceptance Criteria

1. THE Market_Movers service SHALL fetch current session price data for all listed symbols on HOSE, HNX, and UPCOM via the existing `Data_Source_Router` and compute: price change (VND), price change percent (%), and session value (VND) for each symbol.
2. THE Market_Movers service SHALL produce three ranked lists: top 20 gainers (highest % change), top 20 losers (lowest % change), and top 20 most active (highest session value in VND).
3. THE Market_Movers service SHALL display per symbol in each list: symbol, exchange, current price (VND), price change (VND), price change percent (%), session volume (shares), and session value (VND).
4. WHILE the trading session is active (09:00–15:00 ICT), THE Market_Movers service SHALL refresh all three lists every 5 minutes.
5. THE Market_Movers service SHALL support filtering each list by exchange (HOSE / HNX / UPCOM / All) and by ICB sector.
6. THE Market_Movers service SHALL display a mini heatmap of ICB sector performance (today's % change per sector) above the three lists.
7. IF the `Data_Source_Router` returns stale data (stale indicator flag set), THEN THE Market_Movers service SHALL set the response header `X-Data-Stale: true` and display a stale data warning banner with the timestamp of the last successful refresh.

---

### Requirement 94: Economic Calendar

**User Story:** As an investor, I want a calendar of upcoming Vietnamese macro events, earnings releases, dividend ex-dates, and IPO schedules, so that I can plan my trades around market-moving events.

#### Acceptance Criteria

1. THE Economic_Calendar SHALL aggregate the following event types: VN macro events (CPI, GDP, trade balance, interest rate decisions from SBV), corporate earnings release dates, dividend ex-dates and payment dates, stock split and bonus share events, and IPO listing dates on HOSE/HNX/UPCOM.
2. THE Economic_Calendar SHALL source dividend ex-dates, stock splits, and bonus share events from the existing `Corporate_Action_Service`.
3. THE Economic_Calendar SHALL display events in both a monthly calendar grid view and a chronological list view, switchable by the user.
4. WHEN a calendar event is within 3 calendar days of the current date, THE Economic_Calendar SHALL display an "Upcoming" badge on the event card.
5. THE Economic_Calendar SHALL highlight events with expected high market impact in amber to distinguish them from medium and low impact events.
6. THE Economic_Calendar SHALL allow users to filter events by type (macro / earnings / dividend / corporate action / IPO) and by symbol.
7. WHEN a user clicks an event, THE Economic_Calendar SHALL display event details: event name, date, affected symbol (if applicable), expected impact level, and a brief description.
8. THE Economic_Calendar SHALL allow users to subscribe to event reminders: WHEN a subscribed event is 1 day away, THE Economic_Calendar SHALL trigger the existing `AlertService` to deliver a notification.

---

### Requirement 95: Daily Market Brief

**User Story:** As an investor, I want an AI-generated daily market summary delivered before market open with a Vietnamese/English toggle, so that I can start each trading day with a clear picture of the macro environment, sector rotation, and top opportunities.

#### Acceptance Criteria

1. THE Daily_Brief service SHALL generate a structured daily market summary by 08:45 ICT on each trading day, using the existing `Multi_Agent_System` (Supervisor_Agent + Analysis_Agent + News_Agent).
2. THE Daily_Brief SHALL be limited to 500 words maximum and contain: VN-Index overnight context, sector rotation summary, top 3 opportunities from SwingMax_Engine opening signals, top whale activity from the prior session, and key events from the Economic_Calendar for the current day.
3. THE Daily_Brief SHALL be written in Vietnamese by default, with an English toggle, using the existing `I18n_Service` language preference.
4. THE Daily_Brief SHALL display a freshness timestamp showing when it was generated, and a "Regenerate" button that triggers a new generation on demand, rate-limited to once per 30 minutes per user (HTTP 429 when limit exceeded).
5. WHEN the Daily_Brief is generated, THE Daily_Brief service SHALL persist the brief text and generation timestamp so that users who open the app after 08:45 ICT see the same brief for the current day.
6. IF the `Multi_Agent_System` fails to generate the Daily_Brief within 60 seconds, THEN THE Daily_Brief service SHALL display the most recently persisted brief with a warning that the current day's brief is unavailable.
7. THE Daily_Brief SHALL include a disclaimer that the content is AI-generated and does not constitute financial advice.
