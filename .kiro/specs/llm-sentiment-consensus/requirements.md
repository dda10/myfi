# Requirements Document

## Introduction

This feature adds LLM-powered sentiment analysis and market consensus capabilities to the EziStock platform. It covers three integrated components: (1) Vietnamese financial news sentiment analysis via RSS feed ingestion, (2) analyst report ingestion with LLM-powered extraction and accuracy tracking, and (3) a composite market consensus engine that aggregates signals from news sentiment, analyst recommendations, and technical analysis into a unified market opinion per symbol. The Go_Backend orchestrates scheduling, caching, and persistence while delegating all LLM inference to the Python_AI_Service via gRPC/REST. Existing code in `backend/internal/domain/sentiment/` and `backend/internal/domain/consensus/` provides the foundation; this spec formalizes the remaining requirements and extends the system with RSS ingestion, analyst report extraction, and cross-source consensus aggregation.

## Glossary

- **Sentiment_Service**: The Go_Backend domain service (`sentiment.SentimentService`) that orchestrates LLM-powered sentiment analysis of Vietnamese financial news articles, manages caching, and persists results to the `article_sentiments` table
- **Consensus_Service**: The Go_Backend domain service (`consensus.ConsensusService`) that aggregates sentiment signals from multiple sources (news, analyst, technical) into composite consensus scores, detects divergences, and computes market mood
- **RSS_Ingester**: The Go_Backend component that periodically fetches and parses RSS feeds from Vietnamese financial news sources (Vietstock, CafeF, VnExpress), extracts article metadata, and queues articles for LLM sentiment analysis
- **Analyst_Report_Service**: The Go_Backend domain service that ingests analyst research reports from Vietnamese brokerages (SSI, VNDirect, HSC, VCBS, MBS), delegates LLM extraction of structured data (target price, recommendation, reasoning), and tracks analyst accuracy over time
- **LLM_Analyzer**: The interface in the Sentiment_Service (`sentiment.LLMAnalyzer`) through which the Go_Backend delegates text analysis to the Python_AI_Service via gRPC (primary) or REST (fallback)
- **Article_Analysis**: The structured result of LLM sentiment analysis on a single article, containing sentiment classification, confidence score, summary, key topics, and impact score
- **Sentiment_Trend**: An aggregated sentiment metric over a time window (1d/7d/30d) for a symbol, including article counts, average scores, and trend direction
- **Consensus_Score**: The weighted composite market opinion for a symbol, combining signals from news sentiment, analyst recommendations, and technical analysis
- **Divergence**: A detected disagreement between source types (e.g., analysts bullish but news negative) for a given symbol, classified by significance
- **Market_Mood**: The overall market-wide sentiment indicator aggregating all signals across all symbols, labeled as fear/cautious/neutral/optimistic/greed
- **Signal_Record**: A single sentiment signal from any source type, stored in the `consensus_signals` table with symbol, source, score, confidence, and metadata
- **Source_Weight**: The configurable contribution weight each source type has in the composite consensus calculation (analyst: 0.35, news: 0.30, technical: 0.20, other: 0.15)
- **Python_AI_Service**: The Python microservice hosting LLM inference, accessed by the Go_Backend via gRPC (primary) or REST (fallback) through the GRPCClient
- **Scheduler**: The Go_Backend infrastructure component that manages periodic jobs with trading-hours-aware intervals (5min during 9:00-15:00 ICT, 30min off-hours, nightly archival at 23:00 ICT)
- **Trading_Hours**: Vietnamese market trading hours: 9:00-15:00 ICT (UTC+7), Monday through Friday

## Requirements

### Requirement 1: RSS Feed Ingestion for Vietnamese Financial News

**User Story:** As an investor, I want the platform to automatically ingest articles from major Vietnamese financial news sources, so that I have a continuous stream of market-relevant content for sentiment analysis.

#### Acceptance Criteria

1. THE RSS_Ingester SHALL support fetching RSS feeds from three Vietnamese financial news sources: Vietstock (vietstock.vn), CafeF (cafef.vn), and VnExpress Business (vnexpress.net/kinh-doanh)
2. WHILE Trading_Hours are active, THE Scheduler SHALL trigger the RSS_Ingester every 30 minutes to poll all configured RSS feeds
3. WHILE Trading_Hours are not active, THE Scheduler SHALL trigger the RSS_Ingester every 2 hours to poll all configured RSS feeds
4. WHEN the RSS_Ingester fetches a feed, THE RSS_Ingester SHALL parse each RSS item and extract the title, URL, publication date, and source identifier
5. WHEN the RSS_Ingester encounters an article URL that has already been processed, THE RSS_Ingester SHALL skip that article to avoid duplicate analysis
6. WHEN the RSS_Ingester successfully parses a new article, THE RSS_Ingester SHALL extract the article body text from the URL using an HTTP fetch with a 10-second timeout
7. IF the RSS_Ingester fails to fetch an article body within the timeout, THEN THE RSS_Ingester SHALL log the failure and continue processing the remaining articles in the feed
8. IF an RSS feed fails to respond after 3 consecutive polling attempts, THEN THE RSS_Ingester SHALL log a warning and skip that feed until the next scheduled poll cycle
9. THE RSS_Ingester SHALL extract stock symbols mentioned in article titles and bodies by matching against the known VN stock symbol list from the Market_Data_Service

### Requirement 2: LLM-Powered Article Sentiment Analysis

**User Story:** As an investor, I want each news article to be analyzed by an LLM for sentiment, key topics, and market impact, so that I can quickly understand the market implications of news events.

#### Acceptance Criteria

1. WHEN the RSS_Ingester queues a new article for analysis, THE Sentiment_Service SHALL invoke the LLM_Analyzer to classify the article sentiment as positive, negative, or neutral
2. WHEN the LLM_Analyzer returns a result, THE Sentiment_Service SHALL persist the Article_Analysis to the `article_sentiments` table with the sentiment classification, confidence score (0.0-1.0), a 1-2 sentence summary, key topics, and impact score (0.0-1.0)
3. THE LLM_Analyzer SHALL delegate all LLM inference to the Python_AI_Service via the GRPCClient (gRPC primary, REST fallback)
4. WHEN the Sentiment_Service receives an on-demand analysis request via POST /api/sentiment/analyze, THE Sentiment_Service SHALL analyze the provided text or fetch the article from the provided URL and return the Article_Analysis
5. THE Sentiment_Service SHALL cache each Article_Analysis result for 1 hour using the article text hash as the cache key
6. IF the LLM_Analyzer call fails, THEN THE Sentiment_Service SHALL return an error to the caller without persisting a partial result
7. THE LLM_Analyzer SHALL extract from each article: catalysts (positive drivers), risk factors (negative drivers), and up to 5 key topics relevant to the Vietnamese stock market

### Requirement 3: Sentiment Trend Tracking

**User Story:** As an investor, I want to see how sentiment for a stock has changed over time, so that I can identify shifts in market opinion before they are reflected in price.

#### Acceptance Criteria

1. WHEN a trend request is received for a symbol and period (1d/7d/30d/90d), THE Sentiment_Service SHALL compute the Sentiment_Trend by aggregating all Article_Analysis records for that symbol within the time window
2. THE Sentiment_Trend SHALL include: total article count, positive/negative/neutral counts, average confidence, average impact, a composite sentiment score (-1.0 to +1.0), and a trend direction (improving/deteriorating/stable)
3. THE Sentiment_Service SHALL determine trend direction by comparing the average sentiment score of the first half of the period to the second half, classifying a difference greater than 0.15 as improving or deteriorating
4. THE Sentiment_Service SHALL cache trend results for 15 minutes
5. WHEN a time-series request is received for a symbol and day count, THE Sentiment_Service SHALL return daily Sentiment_Snapshot records containing date, average sentiment score, and article count for frontend charting

### Requirement 4: Analyst Report Ingestion with LLM Extraction

**User Story:** As an investor, I want the platform to ingest analyst reports from Vietnamese brokerages and use LLM to extract structured recommendations, so that I can see what professional analysts think about specific stocks.

#### Acceptance Criteria

1. THE Analyst_Report_Service SHALL support ingesting research reports from Vietnamese brokerages: SSI, VNDirect, HSC, VCBS, and MBS
2. WHEN an analyst report is ingested, THE Analyst_Report_Service SHALL delegate to the LLM_Analyzer to extract: target price (VND), recommendation (buy/sell/hold/outperform/underperform), analyst name, brokerage name, key reasoning summary, and the stock symbol
3. THE Analyst_Report_Service SHALL persist each extracted report as an AnalystReport record with all extracted fields and the publication date
4. IF the LLM_Analyzer fails to extract a target price or recommendation from a report, THEN THE Analyst_Report_Service SHALL log the extraction failure with the report URL and skip that report
5. THE Analyst_Report_Service SHALL deduplicate reports by matching on (symbol, brokerage, published_at) to prevent storing the same report twice
6. WHEN a new analyst report is successfully ingested, THE Analyst_Report_Service SHALL store a corresponding Signal_Record in the `consensus_signals` table with source type "analyst" to feed the Consensus_Service

### Requirement 5: Analyst Accuracy Tracking

**User Story:** As an investor, I want to see how accurate each analyst and brokerage has been historically, so that I can weight their recommendations by track record.

#### Acceptance Criteria

1. THE Analyst_Report_Service SHALL compute accuracy for each analyst report at 1-month, 3-month, and 6-month intervals by comparing the target price to the actual stock price at those time offsets
2. THE Scheduler SHALL trigger accuracy computation as a nightly archival job at 23:00 ICT to update accuracy fields for reports that have reached their 1m/3m/6m milestones
3. WHEN accuracy is computed for a report, THE Analyst_Report_Service SHALL calculate accuracy as 1.0 minus the absolute percentage difference between target price and actual price, clamped to a minimum of 0.0
4. WHEN a consensus recommendation request is received for a symbol, THE Analyst_Report_Service SHALL aggregate all analyst reports for that symbol and return: consensus action (buy/sell/hold by majority), average target price, high/low target prices, total analyst count, buy/hold/sell counts, and average accuracy across contributing analysts
5. THE Analyst_Report_Service SHALL weight analyst recommendations by their historical accuracy when computing the consensus action, giving higher-accuracy analysts more influence

### Requirement 6: Composite Market Consensus Scoring

**User Story:** As an investor, I want a single composite score that combines news sentiment, analyst recommendations, and technical signals for each stock, so that I can quickly gauge overall market opinion.

#### Acceptance Criteria

1. WHEN a consensus score request is received for a symbol and period, THE Consensus_Service SHALL compute a weighted Consensus_Score by aggregating Signal_Records from all source types (news, analyst, technical) within the time window
2. THE Consensus_Service SHALL apply Source_Weights to compute the composite score: analyst signals weighted at 0.35, news signals at 0.30, technical signals at 0.20, and other sources at 0.15
3. THE Consensus_Score SHALL include: composite score (-1.0 to +1.0), confidence (0.0-1.0 based on data volume and source agreement), signal strength classification (strong_buy/buy/neutral/sell/strong_sell), per-source breakdown, total signal count, and period
4. THE Consensus_Service SHALL compute confidence based on three factors: data volume (capped at 50 signals), source diversity (capped at 4 source types), and source agreement (inverse of score variance across sources)
5. WHEN a previous period's consensus score is available, THE Consensus_Service SHALL include the previous score and the score change delta in the response for trend comparison
6. THE Consensus_Service SHALL cache consensus scores for 15 minutes

### Requirement 7: Divergence Detection

**User Story:** As an investor, I want to be alerted when different information sources disagree about a stock, so that I can investigate potential opportunities or risks that others may be missing.

#### Acceptance Criteria

1. WHEN a divergence request is received for a symbol, THE Consensus_Service SHALL compare sentiment scores across all source types and identify pairs where the absolute score difference exceeds 0.4
2. THE Consensus_Service SHALL classify each Divergence by significance: high (gap > 0.8), medium (gap > 0.6), or low (gap > 0.4)
3. EACH Divergence SHALL include: the two disagreeing source types, their respective scores, the gap magnitude, significance level, detection timestamp, and a human-readable description
4. THE Consensus_Service SHALL return all detected divergences for the requested symbol sorted by significance (high first)

### Requirement 8: Market-Wide Mood Indicator

**User Story:** As an investor, I want to see the overall market mood across all Vietnamese stocks, so that I can understand the prevailing market sentiment before making investment decisions.

#### Acceptance Criteria

1. WHEN a market mood request is received, THE Consensus_Service SHALL compute the Market_Mood by averaging all Signal_Records from the past 7 days across all symbols
2. THE Market_Mood SHALL include: overall score (-1.0 to +1.0), mood label (fear/cautious/neutral/optimistic/greed), top 5 most bullish symbols, top 5 most bearish symbols, sector-level mood breakdown, total signal count, and computation timestamp
3. THE Consensus_Service SHALL classify mood labels using these thresholds: greed (score > 0.5), optimistic (score > 0.2), neutral (score > -0.2), cautious (score > -0.5), fear (score <= -0.5)
4. THE Consensus_Service SHALL require a minimum of 3 signals per symbol before including that symbol in the top bullish/bearish rankings
5. THE Consensus_Service SHALL cache the Market_Mood result for 15 minutes

### Requirement 9: Technical Analysis Signal Integration

**User Story:** As an investor, I want technical analysis signals from the existing Technical_Analyst_Agent to be included in the consensus calculation, so that the composite score reflects both qualitative and quantitative analysis.

#### Acceptance Criteria

1. WHEN the Technical_Analyst_Agent produces a composite signal for a symbol, THE Consensus_Service SHALL receive a Signal_Record with source type "technical" containing the signal score mapped to the -1.0 to +1.0 range
2. THE Consensus_Service SHALL map Technical_Analyst_Agent composite signals to consensus scores: strongly_bullish = +1.0, bullish = +0.5, neutral = 0.0, bearish = -0.5, strongly_bearish = -1.0
3. THE Scheduler SHALL trigger technical signal ingestion as an evaluation job, running every 5 minutes during Trading_Hours and every 30 minutes outside Trading_Hours

### Requirement 10: News Sentiment to Consensus Pipeline

**User Story:** As an investor, I want news sentiment results to automatically feed into the consensus engine, so that the composite score stays current with the latest news.

#### Acceptance Criteria

1. WHEN the Sentiment_Service persists a new Article_Analysis, THE Sentiment_Service SHALL also store a corresponding Signal_Record in the `consensus_signals` table with source type "news", the article's sentiment score mapped to -1.0/0.0/+1.0, the confidence score, article URL, and extracted topics
2. THE Sentiment_Service SHALL map sentiment classifications to consensus signal scores: positive = +1.0, neutral = 0.0, negative = -1.0
3. THE Signal_Record SHALL include the article's key topics in the topics JSONB field for theme-based querying by the Consensus_Service

### Requirement 11: Scheduled Job Orchestration

**User Story:** As a developer, I want all periodic tasks (RSS polling, accuracy updates, signal ingestion) to be managed by the existing Scheduler with trading-hours-aware intervals, so that the system operates efficiently without manual intervention.

#### Acceptance Criteria

1. THE Scheduler SHALL register the RSS_Ingester polling job as an evaluation job that runs every 30 minutes during Trading_Hours and every 2 hours outside Trading_Hours
2. THE Scheduler SHALL register the analyst accuracy update job as a nightly archival job that runs at 23:00 ICT
3. THE Scheduler SHALL register the technical signal ingestion job as an evaluation job with the standard 5-minute/30-minute trading-hours-aware interval
4. IF a scheduled job fails, THEN THE Scheduler SHALL log the error with the job name, category, and elapsed time, and continue executing remaining jobs in the cycle
5. THE Scheduler SHALL execute jobs within a category sequentially to avoid resource contention

### Requirement 12: API Endpoints for Sentiment and Consensus

**User Story:** As a frontend developer, I want well-defined REST API endpoints for sentiment analysis and consensus data, so that I can build the sentiment dashboard and consensus views.

#### Acceptance Criteria

1. THE Go_Backend SHALL expose POST /api/sentiment/analyze for on-demand article sentiment analysis, accepting a JSON body with symbol (required), and either url or text
2. THE Go_Backend SHALL expose GET /api/sentiment/trend?symbol={symbol}&period={period} for sentiment trend data
3. THE Go_Backend SHALL expose GET /api/sentiment/timeseries?symbol={symbol}&days={days} for daily sentiment snapshots
4. THE Go_Backend SHALL expose GET /api/sentiment/articles?symbol={symbol}&limit={limit} for recently analyzed articles
5. THE Go_Backend SHALL expose GET /api/consensus/score?symbol={symbol}&period={period} for composite consensus scores
6. THE Go_Backend SHALL expose GET /api/consensus/divergences?symbol={symbol} for divergence detection results
7. THE Go_Backend SHALL expose GET /api/consensus/mood for the market-wide mood indicator
8. ALL sentiment and consensus API endpoints SHALL require JWT authentication via the existing auth middleware
9. IF a required query parameter (symbol) is missing, THEN THE Go_Backend SHALL return HTTP 400 with a descriptive error message

### Requirement 13: Caching Strategy

**User Story:** As a developer, I want consistent caching across all sentiment and consensus data, so that the system performs well under load without serving stale data.

#### Acceptance Criteria

1. THE Sentiment_Service SHALL cache individual Article_Analysis results for 1 hour using the cache key pattern "sentiment:article:{symbol}:{textHash}"
2. THE Sentiment_Service SHALL cache Sentiment_Trend results for 15 minutes using the cache key pattern "sentiment:trend:{symbol}:{period}"
3. THE Sentiment_Service SHALL cache time-series snapshots for 15 minutes using the cache key pattern "sentiment:ts:{symbol}:{days}"
4. THE Consensus_Service SHALL cache Consensus_Score results for 15 minutes using the cache key pattern "consensus:{symbol}:{period}"
5. THE Consensus_Service SHALL cache Market_Mood results for 15 minutes using the cache key pattern "consensus:market_mood"

### Requirement 14: Graceful Degradation

**User Story:** As an investor, I want the platform to continue functioning even when the LLM service is unavailable, so that I can still access cached sentiment data and other platform features.

#### Acceptance Criteria

1. IF the Python_AI_Service is unavailable when the Sentiment_Service attempts LLM analysis, THEN THE Sentiment_Service SHALL return an error for the specific analysis request without affecting other platform features
2. IF the Python_AI_Service is unavailable during scheduled RSS ingestion, THEN THE RSS_Ingester SHALL log the failure and retry on the next scheduled poll cycle without crashing
3. WHILE the Python_AI_Service is unavailable, THE Consensus_Service SHALL continue to serve cached consensus scores and market mood data from the Redis cache
4. WHILE the Python_AI_Service is unavailable, THE Go_Backend SHALL return previously cached sentiment trends and time-series data with a stale indicator flag when the cache has not yet expired
