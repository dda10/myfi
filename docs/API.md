# MyFi API Documentation

Base URL: `http://localhost:8080/api`

All responses are JSON unless otherwise noted. Protected endpoints require a `Bearer` token in the `Authorization` header.

## Table of Contents

- [Authentication](#authentication)
- [Market Data](#market-data)
- [Prices](#prices)
- [Portfolio](#portfolio)
- [Sectors](#sectors)
- [Screener](#screener)
- [Comparison](#comparison)
- [Watchlists](#watchlists)
- [AI Chat](#ai-chat)
- [Alerts](#alerts)
- [Knowledge Base](#knowledge-base)
- [Goals](#goals)
- [Recommendations](#recommendations)
- [Signals](#signals)
- [Backtest](#backtest)
- [Export](#export)
- [Health and Metrics](#health-and-metrics)
- [Liquidity Whitelist](#liquidity-whitelist)
- [Authentication Flow](#authentication-flow)
- [Rate Limits and Quotas](#rate-limits-and-quotas)
- [Error Codes and Messages](#error-codes-and-messages)

---

## Authentication

Public endpoints — no JWT required.

### Register

```
POST /api/auth/register
```

**Request:**
```json
{
  "username": "trader1",
  "password": "securepass8",
  "email": "trader@example.com"
}
```

| Field | Type | Required | Constraints |
|-------|------|----------|-------------|
| `username` | string | yes | 3–50 characters |
| `password` | string | yes | min 8 characters |
| `email` | string | no | valid email format |

**Response** `201 Created`:
```json
{
  "id": 1,
  "username": "trader1",
  "email": "trader@example.com",
  "createdAt": "2025-01-15T10:00:00Z",
  "themePreference": "light",
  "languagePreference": "vi-VN"
}
```

**Errors:**
- `409 Conflict` — `USER_EXISTS`: username already taken

### Login

```
POST /api/auth/login
```

**Request:**
```json
{ "username": "trader1", "password": "securepass8" }
```

**Response** `200 OK`:
```json
{
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "expiresAt": 1705420800,
  "user": {
    "id": 1,
    "username": "trader1",
    "themePreference": "dark",
    "languagePreference": "vi-VN"
  }
}
```

**Errors:**
- `401 Unauthorized` — `INVALID_CREDENTIALS`: wrong username or password
- `429 Too Many Requests` — `ACCOUNT_LOCKED`: account locked after 5 failed attempts (30-minute lockout)

### Logout

```
POST /api/auth/logout
Authorization: Bearer <token>
```

**Response** `200 OK`:
```json
{ "message": "logged out" }
```

### Change Password

```
POST /api/auth/change-password
Authorization: Bearer <token>
```

**Request:**
```json
{ "currentPassword": "oldpass", "newPassword": "newpass123" }
```

**Response** `200 OK`:
```json
{ "message": "password changed successfully" }
```

**Errors:**
- `400 Bad Request` — `PASSWORD_MISMATCH`: current password is incorrect

### Get Current User

```
GET /api/auth/me
Authorization: Bearer <token>
```

**Response** `200 OK`:
```json
{
  "id": 1,
  "username": "trader1",
  "email": "trader@example.com",
  "createdAt": "2025-01-15T10:00:00Z",
  "lastLogin": "2025-07-01T08:30:00Z",
  "themePreference": "dark",
  "languagePreference": "vi-VN"
}
```

---

## Market Data

All market data endpoints require authentication.

### Market Quote

Real-time stock quotes with intelligent VCI/KBS source selection.

```
GET /api/market/quote?symbols=SSI,FPT
Authorization: Bearer <token>
```

| Param | Type | Required | Description |
|-------|------|----------|-------------|
| `symbols` | string | yes | Comma-separated stock symbols |

**Response** `200 OK`:
```json
{
  "data": [
    {
      "symbol": "FPT",
      "close": 125000,
      "change": 1500,
      "changePercent": 1.21,
      "source": "VCI",
      "whitelisted": true
    }
  ],
  "source": "VCI"
}
```

### Market Chart (OHLCV)

Historical candlestick data.

```
GET /api/market/chart?symbol=SSI&interval=1d&start=2024-01-01&end=2024-12-31
Authorization: Bearer <token>
```

| Param | Type | Required | Description |
|-------|------|----------|-------------|
| `symbol` | string | yes | Stock symbol |
| `interval` | string | no | `1m`, `5m`, `15m`, `1h`, `1d` (default), `1w`, `1M` |
| `start` | string | no | Start date `YYYY-MM-DD` |
| `end` | string | no | End date `YYYY-MM-DD` |

**Response** `200 OK`:
```json
{
  "data": [
    {
      "time": "2024-06-01",
      "open": 120000,
      "high": 126000,
      "low": 119500,
      "close": 125000,
      "volume": 3500000
    }
  ],
  "source": "VCI"
}
```

### Unified Listing

All available stocks across HOSE, HNX, UPCOM.

```
GET /api/market/listing
Authorization: Bearer <token>
```

### Company Data

Company overview for a specific symbol.

```
GET /api/market/company/:symbol
Authorization: Bearer <token>
```

**Example:** `GET /api/market/company/FPT`

### Financial Reports

Financial statements for a symbol.

```
GET /api/market/finance/:symbol
Authorization: Bearer <token>
```

**Example:** `GET /api/market/finance/FPT`

### Trading Statistics

Trading data for a symbol. Supports optional date range.

```
GET /api/market/trading/:symbol?start=2024-01-01&end=2024-12-31
Authorization: Bearer <token>
```

### Batch Trading Quotes

Fetch trading quotes for multiple symbols in one request.

```
POST /api/market/trading/batch
Authorization: Bearer <token>
```

**Request:**
```json
{ "symbols": ["FPT", "SSI", "VNM"] }
```

### Market Statistics

Aggregate market statistics.

```
GET /api/market/statistics
Authorization: Bearer <token>
```

### Valuation Metrics

Market-wide valuation data.

```
GET /api/market/valuation
Authorization: Bearer <token>
```

### Funds

Open fund (mutual fund) data.

```
GET /api/market/funds
Authorization: Bearer <token>
```

### Commodities

Commodity market data (gold, oil, steel, agricultural).

```
GET /api/market/commodities
Authorization: Bearer <token>
```

### Macro

Macroeconomic indicators.

```
GET /api/market/macro
Authorization: Bearer <token>
```

### Market Screener

Server-side stock screener with query parameters.

```
GET /api/market/screener
Authorization: Bearer <token>
```

---

## Prices

Multi-asset price endpoints.

### Price Quotes

Fetch current prices for VN stocks.

```
GET /api/prices/quotes?symbols=SSI,FPT,VNM
Authorization: Bearer <token>
```

| Param | Type | Required | Description |
|-------|------|----------|-------------|
| `symbols` | string | yes | Comma-separated symbols |

### Price History

Historical OHLCV data for a single symbol.

```
GET /api/prices/history?symbol=SSI&start=2024-01-01&end=2024-12-31
Authorization: Bearer <token>
```

| Param | Type | Required | Description |
|-------|------|----------|-------------|
| `symbol` | string | yes | Stock symbol |
| `start` | string | no | Start date `YYYY-MM-DD` |
| `end` | string | no | End date `YYYY-MM-DD` |

### Gold Prices

Current gold buy/sell prices from Doji (primary) and SJC (fallback).

```
GET /api/prices/gold
Authorization: Bearer <token>
```

**Response** `200 OK`:
```json
{
  "prices": [
    { "type": "SJC", "buy": 82500000, "sell": 84000000 },
    { "type": "9999", "buy": 81000000, "sell": 82000000 }
  ],
  "source": "Doji",
  "isStale": false
}
```

Prices are in VND (Doji prices multiplied by 1000). Cache TTL: 1 hour.

### Crypto Prices

Cryptocurrency quotes from CoinGecko.

```
GET /api/prices/crypto
Authorization: Bearer <token>
```

Cache TTL: 5 minutes.

### FX Rate

USD/VND exchange rate from CoinGecko (USDT/VND pair). Fallback: 25,400 VND.

```
GET /api/prices/fx
Authorization: Bearer <token>
```

**Response** `200 OK`:
```json
{ "pair": "USD/VND", "rate": 25450.0, "source": "CoinGecko" }
```

### Crypto Quote

Single cryptocurrency quote.

```
GET /api/crypto/quote
Authorization: Bearer <token>
```

---

## Portfolio

All portfolio endpoints are scoped to the authenticated user.

### Get Summary

Full portfolio overview: NAV, allocation, holdings with unrealized P&L.

```
GET /api/portfolio/summary
Authorization: Bearer <token>
```

**Response** `200 OK`:
```json
{
  "nav": 1500000000,
  "navChange24h": 15000000,
  "navChangePercent": 1.01,
  "allocationByType": { "vn_stock": 1200000000, "crypto": 200000000, "gold": 100000000 },
  "allocationPercent": { "vn_stock": 80.0, "crypto": 13.3, "gold": 6.7 },
  "holdings": [
    {
      "asset": {
        "id": 1,
        "userId": 1,
        "assetType": "vn_stock",
        "symbol": "FPT",
        "quantity": 100,
        "averageCost": 85000,
        "acquisitionDate": "2024-06-01T00:00:00Z",
        "account": ""
      },
      "currentPrice": 125000,
      "marketValue": 12500000,
      "unrealizedPL": 4000000,
      "unrealizedPLPct": 47.06
    }
  ]
}
```

### Add Asset

```
POST /api/portfolio/assets
Authorization: Bearer <token>
```

**Request:**
```json
{
  "assetType": "vn_stock",
  "symbol": "FPT",
  "quantity": 100,
  "averageCost": 85000,
  "acquisitionDate": "2024-06-01T00:00:00Z"
}
```

| Field | Type | Required | Values |
|-------|------|----------|--------|
| `assetType` | string | yes | `vn_stock`, `crypto`, `gold`, `savings`, `term_deposit`, `bank_account`, `bond` |
| `symbol` | string | yes | Asset symbol/identifier |
| `quantity` | number | yes | Amount held |
| `averageCost` | number | yes | Average cost per unit (VND) |
| `acquisitionDate` | string | yes | ISO 8601 datetime |

**Response** `201 Created`:
```json
{ "id": 1 }
```

### Update Asset

```
PUT /api/portfolio/assets/:id
Authorization: Bearer <token>
```

**Request:**
```json
{ "quantity": 150, "averageCost": 87000 }
```

**Response** `200 OK`:
```json
{ "ok": true }
```

### Delete Asset

```
DELETE /api/portfolio/assets/:id
Authorization: Bearer <token>
```

**Response** `200 OK`:
```json
{ "ok": true }
```

### Record Transaction

```
POST /api/portfolio/transactions
Authorization: Bearer <token>
```

**Request:**
```json
{
  "assetType": "vn_stock",
  "symbol": "FPT",
  "quantity": 50,
  "unitPrice": 90000,
  "transactionDate": "2024-07-15T00:00:00Z",
  "transactionType": "buy"
}
```

| Field | Type | Values |
|-------|------|--------|
| `transactionType` | string | `buy`, `sell`, `deposit`, `withdrawal`, `interest`, `dividend` |

**Response** `201 Created`:
```json
{ "id": 1 }
```

### Get Transactions

```
GET /api/portfolio/transactions
Authorization: Bearer <token>
```

Returns an array of all transactions for the user.

### Performance

Portfolio performance metrics: TWR, MWRR/XIRR, equity curve, benchmark comparison.

```
GET /api/portfolio/performance?start=2024-01-01&end=2024-12-31
Authorization: Bearer <token>
```

| Param | Type | Required | Description |
|-------|------|----------|-------------|
| `start` | string | no | Start date `YYYY-MM-DD` (default: 1 year ago) |
| `end` | string | no | End date `YYYY-MM-DD` (default: today) |

**Response** `200 OK`:
```json
{
  "twr": 0.152,
  "mwrr": 0.148,
  "equityCurve": [
    { "date": "2024-01-01T00:00:00Z", "value": 1000000000 },
    { "date": "2024-12-31T00:00:00Z", "value": 1152000000 }
  ],
  "benchmarkComparison": { "vnIndex": 0.12, "vn30": 0.11 },
  "performanceByType": { "vn_stock": 0.18, "crypto": 0.05 }
}
```

### Risk Analysis

Portfolio risk metrics: Sharpe ratio, max drawdown, beta, volatility, VaR.

```
GET /api/portfolio/risk
Authorization: Bearer <token>
```

**Response** `200 OK`:
```json
{
  "sharpeRatio": 1.45,
  "maxDrawdown": -0.12,
  "beta": 0.95,
  "volatility": 0.18,
  "valueAtRisk": -25000000,
  "holdingRisks": [
    { "symbol": "FPT", "beta": 1.1, "volatility": 0.22, "contribution": 0.35 }
  ]
}
```

---

## Sectors

ICB sector classification and performance. Sector codes: `VNIT` (Technology), `VNIND` (Industrial), `VNCONS` (Consumer), `VNCOND` (Consumer Staples), `VNHEAL` (Healthcare), `VNENE` (Energy), `VNUTI` (Utilities), `VNREAL` (Real Estate), `VNFIN` (Finance), `VNMAT` (Materials).

### All Sector Performance

```
GET /api/sectors/performance
Authorization: Bearer <token>
```

**Response** `200 OK`:
```json
[
  {
    "sector": "VNFIN",
    "trend": "uptrend",
    "todayChange": 0.85,
    "oneWeekChange": 2.1,
    "oneMonthChange": 5.3,
    "threeMonthChange": 8.2,
    "sixMonthChange": 12.5,
    "oneYearChange": 18.0,
    "currentPrice": 1250.5,
    "sma20": 1230.0,
    "sma50": 1200.0
  }
]
```

### Sector Performance

```
GET /api/sectors/:sector/performance
Authorization: Bearer <token>
```

**Example:** `GET /api/sectors/VNFIN/performance`

### Symbol Sector Lookup

```
GET /api/sectors/symbol/:symbol
Authorization: Bearer <token>
```

**Example:** `GET /api/sectors/symbol/FPT`

**Response** `200 OK`:
```json
{ "symbol": "FPT", "sector": "VNIT" }
```

### Sector Averages

Fundamental averages for a sector.

```
GET /api/sectors/:sector/averages
Authorization: Bearer <token>
```

**Response** `200 OK`:
```json
{
  "sector": "VNFIN",
  "medianPE": 12.5,
  "medianPB": 1.8,
  "medianROE": 15.2,
  "medianROA": 2.1,
  "medianDivYield": 3.5,
  "medianDebtToEquity": 5.2
}
```

### Sector Stocks

All stocks in a given sector.

```
GET /api/sectors/:sector/stocks
Authorization: Bearer <token>
```

**Example:** `GET /api/sectors/VNFIN/stocks`

---

## Screener

Advanced stock filtering with fundamental, sector, and trend criteria.

### Run Screener

```
POST /api/screener
Authorization: Bearer <token>
```

**Request:**
```json
{
  "minPE": 5,
  "maxPE": 15,
  "minMarketCap": 1000000000000,
  "minROE": 10,
  "sectors": ["VNFIN", "VNIT"],
  "exchanges": ["HOSE"],
  "sectorTrends": ["uptrend"],
  "sortBy": "pe",
  "sortOrder": "asc",
  "page": 1,
  "pageSize": 20
}
```

| Field | Type | Description |
|-------|------|-------------|
| `minPE` / `maxPE` | number | P/E ratio range |
| `minPB` / `maxPB` | number | P/B ratio range |
| `minMarketCap` | number | Minimum market cap (VND) |
| `minEVEBITDA` / `maxEVEBITDA` | number | EV/EBITDA range |
| `minROE` / `maxROE` | number | ROE range (%) |
| `minROA` / `maxROA` | number | ROA range (%) |
| `minRevenueGrowth` / `maxRevenueGrowth` | number | Revenue growth range (%) |
| `minProfitGrowth` / `maxProfitGrowth` | number | Profit growth range (%) |
| `minDivYield` / `maxDivYield` | number | Dividend yield range (%) |
| `minDebtToEquity` / `maxDebtToEquity` | number | Debt-to-equity range |
| `sectors` | string[] | ICB sector codes to include |
| `exchanges` | string[] | `HOSE`, `HNX`, `UPCOM` |
| `sectorTrends` | string[] | `uptrend`, `downtrend`, `sideways` |
| `sortBy` | string | Field to sort by |
| `sortOrder` | string | `asc` or `desc` |
| `page` | int | Page number (default: 1) |
| `pageSize` | int | Results per page (default: 20) |

All filter fields are optional. Only non-null filters are applied.

**Response** `200 OK`:
```json
{
  "data": [
    {
      "symbol": "VCB",
      "exchange": "HOSE",
      "sector": "VNFIN",
      "marketCap": 450000000000000,
      "pe": 12.5,
      "pb": 2.8,
      "evEbitda": 8.5,
      "roe": 22.1,
      "roa": 1.8,
      "revenueGrowth": 15.2,
      "profitGrowth": 18.5,
      "divYield": 1.2,
      "debtToEquity": 8.5,
      "sectorTrend": "uptrend"
    }
  ],
  "total": 45,
  "page": 1,
  "pageSize": 20,
  "totalPages": 3
}
```

### Get Presets

Saved screener filter presets.

```
GET /api/screener/presets
Authorization: Bearer <token>
```

**Response** `200 OK`:
```json
[
  {
    "id": 1,
    "userId": 1,
    "name": "Value Stocks",
    "filters": { "maxPE": 10, "minDivYield": 3 }
  }
]
```

### Save Preset

```
POST /api/screener/presets
Authorization: Bearer <token>
```

**Request:**
```json
{ "name": "Value Stocks", "filters": { "maxPE": 10, "minDivYield": 3 } }
```

**Response** `201 Created`:
```json
{ "id": 1 }
```

### Delete Preset

```
DELETE /api/screener/presets/:id
Authorization: Bearer <token>
```

**Response** `200 OK`:
```json
{ "ok": true }
```

---

## Comparison

Side-by-side stock comparison. All endpoints accept `symbols` (comma-separated, max 10) and optional `period`.

| Period | Description |
|--------|-------------|
| `1M` | 1 month |
| `3M` | 3 months |
| `6M` | 6 months |
| `1Y` | 1 year (default) |
| `3Y` | 3 years |

### Valuation Comparison

```
GET /api/comparison/valuation?symbols=FPT,SSI,VNM&period=1Y
Authorization: Bearer <token>
```

### Performance Comparison

```
GET /api/comparison/performance?symbols=FPT,SSI&period=6M
Authorization: Bearer <token>
```

### Correlation Matrix

```
GET /api/comparison/correlation?symbols=FPT,SSI,VNM&period=1Y
Authorization: Bearer <token>
```

---

## Watchlists

Named watchlists with per-symbol price alerts. Scoped to the authenticated user.

### List Watchlists

```
GET /api/watchlists
Authorization: Bearer <token>
```

**Response** `200 OK`:
```json
[
  {
    "id": 1,
    "userId": 1,
    "name": "Tech Stocks",
    "symbols": [
      { "symbol": "FPT", "sortOrder": 0, "alertAbove": 130000, "alertBelow": null }
    ]
  }
]
```

### Create Watchlist

```
POST /api/watchlists
Authorization: Bearer <token>
```

**Request:**
```json
{ "name": "Tech Stocks" }
```

**Response** `201 Created`: returns the created watchlist object.

### Rename Watchlist

```
PUT /api/watchlists/:id
Authorization: Bearer <token>
```

**Request:**
```json
{ "name": "Tech & Finance" }
```

**Response** `200 OK`:
```json
{ "ok": true }
```

### Delete Watchlist

```
DELETE /api/watchlists/:id
Authorization: Bearer <token>
```

**Response** `200 OK`:
```json
{ "ok": true }
```

### Add Symbol

```
POST /api/watchlists/:id/symbols
Authorization: Bearer <token>
```

**Request:**
```json
{ "symbol": "FPT" }
```

**Response** `201 Created`:
```json
{ "ok": true }
```

### Remove Symbol

```
DELETE /api/watchlists/:id/symbols/:symbol
Authorization: Bearer <token>
```

**Response** `200 OK`:
```json
{ "ok": true }
```

### Set Price Alert

```
PUT /api/watchlists/:id/symbols/:symbol/alert
Authorization: Bearer <token>
```

**Request:**
```json
{ "alertAbove": 130000, "alertBelow": 80000 }
```

Both fields are optional. Set to `null` to clear.

**Response** `200 OK`:
```json
{ "ok": true }
```

### Reorder Symbols

```
PUT /api/watchlists/:id/reorder
Authorization: Bearer <token>
```

**Request:**
```json
{ "symbols": ["SSI", "FPT", "VNM"] }
```

**Response** `200 OK`:
```json
{ "ok": true }
```

---

## AI Chat

Multi-agent AI advisory system powered by langchaingo.

### Send Message

```
POST /api/chat
Authorization: Bearer <token>
```

**Request:**
```json
{ "message": "Phân tích cổ phiếu FPT" }
```

**Response** `200 OK`:
```json
{
  "summary": "FPT đang trong xu hướng tăng...",
  "assetRecommendations": [
    {
      "symbol": "FPT",
      "action": "buy",
      "positionSize": 5.0,
      "riskAssessment": "medium",
      "reasoning": "Strong uptrend with sector support..."
    }
  ],
  "portfolioSuggestions": ["Consider increasing tech allocation"],
  "identifiedOpportunities": ["FPT approaching breakout level"],
  "sectorContext": "VNIT sector in uptrend",
  "knowledgeBaseInsights": ["Similar accumulation pattern detected 30 days ago"]
}
```

### List Available Models

```
POST /api/models
Authorization: Bearer <token>
```

Returns the list of configured LLM providers (OpenAI, Anthropic, Google, Qwen, Bedrock).

---

## Alerts

Proactive notifications from the Monitor_Agent for detected market patterns.

### Get Alerts

```
GET /api/alerts
Authorization: Bearer <token>
```

### Update Alert Preferences

```
PUT /api/alerts/preferences
Authorization: Bearer <token>
```

### Mark Alert Viewed

```
PUT /api/alerts/:id/viewed
Authorization: Bearer <token>
```

**Response** `200 OK`:
```json
{ "ok": true }
```

---

## Knowledge Base

Pattern observations and accuracy metrics from the autonomous Monitor_Agent.

### Get Observations

```
GET /api/knowledge/observations
Authorization: Bearer <token>
```

**Response** `200 OK`:
```json
[
  {
    "id": 1,
    "symbol": "FPT",
    "patternType": "accumulation",
    "detectionDate": "2025-01-10T09:00:00Z",
    "confidenceScore": 75,
    "priceAtDetection": 120000,
    "supportingData": "{...}",
    "outcome1Day": 1.2,
    "outcome7Day": 3.5,
    "outcome14Day": null,
    "outcome30Day": null
  }
]
```

### Get Pattern Accuracy

```
GET /api/knowledge/accuracy/:patternType
Authorization: Bearer <token>
```

| Pattern Type | Description |
|-------------|-------------|
| `accumulation` | Accumulation phase detection |
| `distribution` | Distribution phase detection |
| `breakout` | Breakout pattern detection |

**Example:** `GET /api/knowledge/accuracy/breakout`

**Response** `200 OK`:
```json
{
  "patternType": "breakout",
  "totalObservations": 42,
  "successCount": 28,
  "failureCount": 14,
  "avgPriceChange": 3.2,
  "avgConfidence": 68.5
}
```

---

## Goals

Financial goal tracking with progress calculation against current NAV.

### List Goals

```
GET /api/goals
Authorization: Bearer <token>
```

**Response** `200 OK`:
```json
[
  {
    "id": 1,
    "userId": 1,
    "name": "Retirement Fund",
    "targetAmount": 5000000000,
    "targetDate": "2035-01-01T00:00:00Z",
    "category": "retirement"
  }
]
```

### Create Goal

```
POST /api/goals
Authorization: Bearer <token>
```

**Request:**
```json
{
  "name": "Retirement Fund",
  "targetAmount": 5000000000,
  "targetDate": "2035-01-01T00:00:00Z",
  "category": "retirement"
}
```

**Response** `201 Created`:
```json
{ "id": 1 }
```

### Update Goal

```
PUT /api/goals/:id
Authorization: Bearer <token>
```

**Request:**
```json
{ "targetAmount": 6000000000 }
```

**Response** `200 OK`:
```json
{ "ok": true }
```

### Delete Goal

```
DELETE /api/goals/:id
Authorization: Bearer <token>
```

**Response** `200 OK`:
```json
{ "ok": true }
```

### Goal Progress

```
GET /api/goals/:id/progress
Authorization: Bearer <token>
```

**Response** `200 OK`:
```json
{
  "goalId": 1,
  "currentNAV": 1500000000,
  "targetAmount": 5000000000,
  "progressPercent": 30.0,
  "remainingAmount": 3500000000,
  "monthsRemaining": 120,
  "requiredMonthlyContribution": 29166667
}
```

---

## Recommendations

AI recommendation audit trail with outcome tracking.

### Summary

Overall recommendation accuracy metrics.

```
GET /api/recommendations/summary
Authorization: Bearer <token>
```

**Response** `200 OK`:
```json
{
  "totalRecommendations": 150,
  "byAction": [
    {
      "action": "buy",
      "totalCount": 80,
      "winRate7Day": 0.65,
      "avgReturn7Day": 2.1,
      "avgConfidence": 72.0,
      "highConfWinRate": 0.78,
      "medConfWinRate": 0.55,
      "lowConfWinRate": 0.30
    }
  ],
  "bestPerformingSymbol": "FPT",
  "worstPerformingSymbol": "HBC",
  "overallWinRate7Day": 0.62,
  "overallAvgReturn7Day": 1.8
}
```

### Accuracy by Action

```
GET /api/recommendations/accuracy?action=buy
Authorization: Bearer <token>
```

| Param | Type | Values |
|-------|------|--------|
| `action` | string | `buy`, `sell`, `hold` (default: `buy`) |

### List Recommendations

```
GET /api/recommendations?symbol=FPT&action=buy&minConfidence=60&limit=50
Authorization: Bearer <token>
```

| Param | Type | Description |
|-------|------|-------------|
| `symbol` | string | Filter by symbol |
| `action` | string | Filter by action (`buy`/`sell`/`hold`) |
| `minConfidence` | int | Minimum confidence score (0–100) |
| `limit` | int | Max results (default: 100) |

**Response** `200 OK`:
```json
{
  "count": 5,
  "recommendations": [
    {
      "id": 42,
      "symbol": "FPT",
      "action": "buy",
      "positionSize": 5.0,
      "riskAssessment": "medium",
      "confidenceScore": 78,
      "reasoning": "Strong uptrend with accumulation pattern...",
      "priceAtSignal": 120000,
      "createdAt": "2025-01-15T10:00:00Z",
      "price1Day": 121500,
      "return1Day": 1.25,
      "price7Day": 125000,
      "return7Day": 4.17
    }
  ]
}
```

### Get Recommendation by ID

```
GET /api/recommendations/:id
Authorization: Bearer <token>
```

### Update Outcomes

Trigger a batch update of recommendation outcomes (price changes at 1d, 7d, 14d, 30d).

```
POST /api/recommendations/update-outcomes
Authorization: Bearer <token>
```

**Response** `200 OK`:
```json
{ "status": "outcomes updated" }
```

---

## Signals

Dual-mode trading and investment signal engine. Scans the VN stock universe for opportunities.

### Scan Signals

Run a full market scan for trading and investment signals.

```
GET /api/signals/scan
Authorization: Bearer <token>
```

**Response** `200 OK`:
```json
{
  "signals": [
    {
      "symbol": "FPT",
      "direction": "long",
      "entryPrice": 125000,
      "stopLoss": 120000,
      "takeProfit": 135000,
      "riskReward": 2.0,
      "confidence": 82,
      "reasoning": "Breakout above resistance with volume confirmation..."
    }
  ],
  "totalScanned": 500,
  "timestamp": "2025-07-01T10:00:00Z"
}
```

### Backtest Signals

Backtest signal engine performance over a historical period.

```
POST /api/signals/backtest
Authorization: Bearer <token>
```

**Request:**
```json
{ "period": "6M" }
```

### Optimize Signal Weights

Optimize signal scoring weights based on historical performance.

```
POST /api/signals/optimize
Authorization: Bearer <token>
```

---

## Backtest

Strategy backtesting against historical OHLCV data.

```
POST /api/backtest
Authorization: Bearer <token>
```

**Request:**
```json
{
  "symbol": "FPT",
  "startDate": "2023-01-01T00:00:00Z",
  "endDate": "2024-01-01T00:00:00Z",
  "strategy": {
    "name": "SMA Crossover",
    "entryConditions": [
      {
        "left": { "type": "indicator", "indicator": "SMA", "period": 20, "field": "value" },
        "operator": "CROSSES_ABOVE",
        "right": { "type": "indicator", "indicator": "SMA", "period": 50, "field": "value" }
      }
    ],
    "exitConditions": [
      {
        "left": { "type": "indicator", "indicator": "SMA", "period": 20, "field": "value" },
        "operator": "CROSSES_BELOW",
        "right": { "type": "indicator", "indicator": "SMA", "period": 50, "field": "value" }
      }
    ],
    "stopLossPct": 0.05,
    "takeProfitPct": 0.10
  }
}
```

**Supported indicators:** `SMA`, `EMA`, `RSI`, `MACD`, `BOLLINGER`, `STOCHASTIC`, `ADX`, `AROON`, `PARABOLIC_SAR`, `SUPERTREND`, `VWAP`, `VWMA`, `WILLIAMS_R`, `CMO`, `ROC`, `MOMENTUM`, `KELTNER`, `ATR`, `STDDEV`, `OBV`, `LINEAR_REG`

**Condition operators:** `LT`, `GT`, `LTE`, `GTE`, `CROSSES_ABOVE`, `CROSSES_BELOW`

**Operand types:** `indicator` (with period/field), `price`, `constant`

**Response** `200 OK`:
```json
{
  "totalReturn": 0.25,
  "winRate": 0.65,
  "maxDrawdown": -0.08,
  "sharpeRatio": 1.8,
  "trades": 12,
  "avgHoldingPeriod": 15.3,
  "equityCurve": [
    { "date": "2023-01-15T00:00:00Z", "value": 100000000 },
    { "date": "2024-01-01T00:00:00Z", "value": 125000000 }
  ],
  "tradeList": [
    {
      "entryDate": "2023-02-01T00:00:00Z",
      "exitDate": "2023-03-15T00:00:00Z",
      "entryPrice": 85000,
      "exitPrice": 92000,
      "returnPct": 0.082,
      "exitReason": "signal",
      "holdingDays": 42
    }
  ]
}
```

---

## Export

File download endpoints. All return binary data with `Content-Disposition` headers.

### Export Transactions (CSV)

```
GET /api/export/transactions?start=2024-01-01&end=2024-12-31
Authorization: Bearer <token>
```

| Param | Type | Required | Description |
|-------|------|----------|-------------|
| `start` | string | no | Start date `YYYY-MM-DD` (default: 1 year ago) |
| `end` | string | no | End date `YYYY-MM-DD` (default: today) |

Returns `text/csv` with filename `transactions_YYYYMMDD.csv`.

### Export Portfolio Snapshot (CSV)

```
GET /api/export/snapshot
Authorization: Bearer <token>
```

Returns `text/csv` with filename `snapshot_YYYYMMDD.csv`.

### Export Portfolio Report

```
GET /api/export/report
Authorization: Bearer <token>
```

Returns a text-based portfolio report with filename `portfolio_report_YYYYMMDD.txt`.

### Export Tax Report (CSV)

```
GET /api/export/tax?start=2024-01-01&end=2024-12-31
Authorization: Bearer <token>
```

Capital gains summary for tax reporting. Returns `text/csv` with filename `tax_report_YYYYMMDD.csv`.

---

## Health and Metrics

### Health Check

Public endpoint — no authentication required.

```
GET /api/health
```

**Response** `200 OK`:
```json
{ "status": "ok" }
```

### Rate Limit Metrics

Current rate limit status for all configured data sources.

```
GET /api/metrics/rate-limits
Authorization: Bearer <token>
```

**Response** `200 OK`:
```json
[
  { "source": "VCI", "maxRequests": 100, "windowDuration": "1m0s" },
  { "source": "KBS", "maxRequests": 100, "windowDuration": "1m0s" },
  { "source": "CoinGecko", "maxRequests": 50, "windowDuration": "1m0s" },
  { "source": "Doji", "maxRequests": 60, "windowDuration": "1m0s" }
]
```

---

## Liquidity Whitelist

Liquidity-filtered stock universe.

### Get Whitelist

```
GET /api/market/whitelist
Authorization: Bearer <token>
```

### Check Symbol

```
GET /api/market/whitelist/check?symbol=FPT
Authorization: Bearer <token>
```

### Refresh Whitelist

```
POST /api/market/whitelist/refresh
Authorization: Bearer <token>
```

---

## News

Financial news aggregation.

```
GET /api/news
Authorization: Bearer <token>
```

---

## Authentication Flow

1. **Register** — `POST /api/auth/register` with username, password, optional email. Returns user object.
2. **Login** — `POST /api/auth/login` with username and password. Returns a JWT token valid for 24 hours.
3. **Use token** — Include `Authorization: Bearer <token>` header on all protected endpoints.
4. **Session timeout** — Sessions expire after 4 hours of inactivity, independent of token expiry.
5. **Token refresh** — No refresh token mechanism. Re-login when the token expires.
6. **Logout** — `POST /api/auth/logout` invalidates the current session.
7. **Account lockout** — After 5 failed login attempts within 15 minutes, the account is locked for 30 minutes.

**Token structure (JWT claims):**

| Claim | Description |
|-------|-------------|
| `user_id` | Numeric user ID |
| `username` | Username string |
| `session_id` | Unique session identifier |
| `exp` | Expiration timestamp (24h from issuance) |
| `iat` | Issued-at timestamp |

**Password requirements:**
- Minimum 8 characters
- Hashed with bcrypt (cost factor 12)

---

## Rate Limits and Quotas

### HTTP Rate Limits

| Scope | Limit | Window | Applies To |
|-------|-------|--------|------------|
| Per IP | 200 requests | 1 minute | All endpoints (applied globally before auth) |
| Per user | 100 requests | 1 minute | Authenticated endpoints (applied after JWT validation) |

Both limits use token-bucket algorithms. Stale entries are cleaned up after 5 minutes of inactivity.

### Data Source Rate Limits

Internal rate limits for upstream API calls:

| Source | Limit | Window |
|--------|-------|--------|
| VCI (Vietcap) | 100 requests | 1 minute |
| KBS | 100 requests | 1 minute |
| CoinGecko | 50 requests | 1 minute |
| Doji (gold) | 60 requests | 1 minute |

When a data source rate limit is reached, requests are queued (max queue depth: 100) with a 5-second wait timeout.

### Circuit Breaker

Each data source has a circuit breaker:
- **Threshold:** 3 consecutive failures within 60 seconds triggers the breaker
- **Open state:** All requests routed to the fallback source for 60 seconds
- **Half-open:** After 60 seconds, a single test request is sent to check recovery
- **Recovery:** On success, the circuit closes and normal routing resumes

### Cache TTLs

| Data Type | TTL |
|-----------|-----|
| VN stock prices | 15 minutes |
| Crypto prices | 5 minutes |
| Gold prices | 1 hour |
| FX rates | 1 hour |

### Rate Limit Response

When any rate limit is exceeded, the API returns:

```
HTTP/1.1 429 Too Many Requests
```

```json
{
  "error": "rate limit exceeded — max 100 requests per minute",
  "code": "RATE_LIMIT_EXCEEDED"
}
```

---

## Error Codes and Messages

### Standard Error Response Format

All errors follow this structure:

```json
{
  "error": "human-readable message",
  "code": "ERROR_CODE"
}
```

### Authentication Errors

| Code | HTTP Status | Description |
|------|-------------|-------------|
| `INVALID_CREDENTIALS` | 401 | Wrong username or password |
| `ACCOUNT_LOCKED` | 429 | Too many failed login attempts (5 within 15 min, locked for 30 min) |
| `USER_NOT_FOUND` | 404 | User does not exist |
| `USER_EXISTS` | 409 | Username already taken during registration |
| `INVALID_TOKEN` | 401 | JWT token is malformed or signature is invalid |
| `TOKEN_EXPIRED` | 401 | JWT token has passed its expiration time |
| `SESSION_EXPIRED` | 401 | Session timed out after 4 hours of inactivity |
| `PASSWORD_MISMATCH` | 400 | Current password is incorrect during password change |

### Rate Limiting Errors

| Code | HTTP Status | Description |
|------|-------------|-------------|
| `RATE_LIMIT_EXCEEDED` | 429 | Per-user (100/min) or per-IP (200/min) limit exceeded |

### Validation Errors

| HTTP Status | When |
|-------------|------|
| 400 | Missing required fields, invalid JSON, invalid parameter format |

Validation errors return the binding error message in the `error` field without a `code`.

**Examples:**
```json
{ "error": "symbols parameter is required" }
{ "error": "invalid asset id" }
{ "error": "name is required" }
{ "error": "invalid action, must be buy/sell/hold" }
```

### Resource Errors

| HTTP Status | When |
|-------------|------|
| 404 | Resource not found (user, recommendation, goal, sector symbol) |
| 503 | Service unavailable (recommendation tracking not enabled) |
| 500 | Internal server error (database failure, upstream API failure) |

### Data Source Errors

When both primary and fallback data sources fail, the API returns cached data with a `isStale: true` flag in the response. If no cached data is available, a `500` error is returned.
