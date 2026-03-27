# MyFi API Documentation

Base URL: `http://localhost:8080/api`

All responses are JSON. Protected endpoints require a `Bearer` token in the `Authorization` header.

---

## Authentication

### Register

```
POST /api/auth/register
```

```bash
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username": "trader1", "password": "securepass8", "email": "trader@example.com"}'
```

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

### Login

```
POST /api/auth/login
```

```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username": "trader1", "password": "securepass8"}'
```

**Response** `200 OK`:
```json
{
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "expiresAt": 1705420800,
  "user": {
    "id": 1,
    "username": "trader1",
    "themePreference": "light",
    "languagePreference": "vi-VN"
  }
}
```

Token expires after 24 hours. Account locks after 5 failed attempts for 30 minutes.

### Logout

```
POST /api/auth/logout
Authorization: Bearer <token>
```

### Change Password

```
POST /api/auth/change-password
Authorization: Bearer <token>
```

```json
{ "currentPassword": "oldpass", "newPassword": "newpass123" }
```

### Get Current User

```
GET /api/auth/me
Authorization: Bearer <token>
```

---

## Market Data

### Unified Listing

```bash
curl -H "Authorization: Bearer $TOKEN" http://localhost:8080/api/market/listing
```

### Company Data

```bash
curl -H "Authorization: Bearer $TOKEN" http://localhost:8080/api/market/company/SSI
```

### Financial Reports

```bash
curl -H "Authorization: Bearer $TOKEN" http://localhost:8080/api/market/finance/FPT
```

### Market Statistics

```bash
curl -H "Authorization: Bearer $TOKEN" http://localhost:8080/api/market/statistics
```

### Commodities

```bash
curl -H "Authorization: Bearer $TOKEN" http://localhost:8080/api/market/commodities
```

### Macro Data

```bash
curl -H "Authorization: Bearer $TOKEN" http://localhost:8080/api/market/macro
```

### Market Quote

```bash
curl -H "Authorization: Bearer $TOKEN" "http://localhost:8080/api/market/quote?symbols=SSI,FPT"
```

### Market Chart (OHLCV)

```bash
curl -H "Authorization: Bearer $TOKEN" "http://localhost:8080/api/market/chart?symbol=SSI&interval=1d"
```

---

## Prices

### Quotes

```bash
curl -H "Authorization: Bearer $TOKEN" "http://localhost:8080/api/prices/quotes?symbols=SSI,FPT,VNM"
```

### Price History

```bash
curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8080/api/prices/history?symbol=SSI&start=2024-01-01&end=2024-12-31"
```

### Gold Prices

```bash
curl -H "Authorization: Bearer $TOKEN" http://localhost:8080/api/prices/gold
```

### Crypto Prices

```bash
curl -H "Authorization: Bearer $TOKEN" http://localhost:8080/api/prices/crypto
```

### FX Rate

```bash
curl -H "Authorization: Bearer $TOKEN" http://localhost:8080/api/prices/fx
```

---

## Portfolio

### Get Summary

```bash
curl -H "Authorization: Bearer $TOKEN" http://localhost:8080/api/portfolio/summary
```

### Add Asset

```bash
curl -X POST -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  http://localhost:8080/api/portfolio/assets \
  -d '{
    "assetType": "vn_stock",
    "symbol": "FPT",
    "quantity": 100,
    "averageCost": 85000,
    "acquisitionDate": "2024-06-01T00:00:00Z"
  }'
```

### Update Asset

```bash
curl -X PUT -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  http://localhost:8080/api/portfolio/assets/1 \
  -d '{"quantity": 150, "averageCost": 87000}'
```

### Delete Asset

```bash
curl -X DELETE -H "Authorization: Bearer $TOKEN" http://localhost:8080/api/portfolio/assets/1
```

### Record Transaction

```bash
curl -X POST -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  http://localhost:8080/api/portfolio/transactions \
  -d '{
    "assetType": "vn_stock",
    "symbol": "FPT",
    "quantity": 50,
    "unitPrice": 90000,
    "transactionDate": "2024-07-15T00:00:00Z",
    "transactionType": "buy"
  }'
```

### Get Transactions

```bash
curl -H "Authorization: Bearer $TOKEN" http://localhost:8080/api/portfolio/transactions
```

### Performance

```bash
curl -H "Authorization: Bearer $TOKEN" "http://localhost:8080/api/portfolio/performance?period=1M"
```

### Risk Analysis

```bash
curl -H "Authorization: Bearer $TOKEN" http://localhost:8080/api/portfolio/risk
```

---

## Sectors

### All Sector Performance

```bash
curl -H "Authorization: Bearer $TOKEN" http://localhost:8080/api/sectors/performance
```

### Sector Performance

```bash
curl -H "Authorization: Bearer $TOKEN" http://localhost:8080/api/sectors/banking/performance
```

### Sector Stocks

```bash
curl -H "Authorization: Bearer $TOKEN" http://localhost:8080/api/sectors/banking/stocks
```

### Sector Averages

```bash
curl -H "Authorization: Bearer $TOKEN" http://localhost:8080/api/sectors/banking/averages
```

---

## Screener

### Run Screener

```bash
curl -X POST -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  http://localhost:8080/api/screener \
  -d '{
    "filters": {
      "pe": {"min": 5, "max": 15},
      "marketCap": {"min": 1000000000000}
    },
    "sort": "pe",
    "order": "asc",
    "limit": 20
  }'
```

### Get Presets

```bash
curl -H "Authorization: Bearer $TOKEN" http://localhost:8080/api/screener/presets
```

### Save Preset

```bash
curl -X POST -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  http://localhost:8080/api/screener/presets \
  -d '{"name": "Value Stocks", "filters": {"pe": {"max": 10}}}'
```

### Delete Preset

```bash
curl -X DELETE -H "Authorization: Bearer $TOKEN" http://localhost:8080/api/screener/presets/1
```

---

## Comparison

### Valuation Comparison

```bash
curl -H "Authorization: Bearer $TOKEN" "http://localhost:8080/api/comparison/valuation?symbols=FPT,SSI,VNM"
```

### Performance Comparison

```bash
curl -H "Authorization: Bearer $TOKEN" "http://localhost:8080/api/comparison/performance?symbols=FPT,SSI"
```

### Correlation

```bash
curl -H "Authorization: Bearer $TOKEN" "http://localhost:8080/api/comparison/correlation?symbols=FPT,SSI,VNM"
```

---

## Watchlists

### List Watchlists

```bash
curl -H "Authorization: Bearer $TOKEN" http://localhost:8080/api/watchlists
```

### Create Watchlist

```bash
curl -X POST -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  http://localhost:8080/api/watchlists \
  -d '{"name": "Tech Stocks"}'
```

### Rename Watchlist

```bash
curl -X PUT -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  http://localhost:8080/api/watchlists/1 \
  -d '{"name": "Tech & Finance"}'
```

### Delete Watchlist

```bash
curl -X DELETE -H "Authorization: Bearer $TOKEN" http://localhost:8080/api/watchlists/1
```

### Add Symbol

```bash
curl -X POST -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  http://localhost:8080/api/watchlists/1/symbols \
  -d '{"symbol": "FPT"}'
```

### Remove Symbol

```bash
curl -X DELETE -H "Authorization: Bearer $TOKEN" http://localhost:8080/api/watchlists/1/symbols/FPT
```

### Set Price Alert

```bash
curl -X PUT -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  http://localhost:8080/api/watchlists/1/symbols/FPT/alert \
  -d '{"above": 100000, "below": 80000}'
```

---

## AI Chat

### Send Message

```bash
curl -X POST -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  http://localhost:8080/api/chat \
  -d '{"message": "Phân tích cổ phiếu FPT"}'
```

### List Models

```bash
curl -X POST -H "Authorization: Bearer $TOKEN" http://localhost:8080/api/models
```

---

## Alerts

### Get Alerts

```bash
curl -H "Authorization: Bearer $TOKEN" http://localhost:8080/api/alerts
```

### Mark Alert Viewed

```bash
curl -X PUT -H "Authorization: Bearer $TOKEN" http://localhost:8080/api/alerts/5/viewed
```

---

## Knowledge Base

### Get Observations

```bash
curl -H "Authorization: Bearer $TOKEN" http://localhost:8080/api/knowledge/observations
```

### Get Pattern Accuracy

```bash
curl -H "Authorization: Bearer $TOKEN" http://localhost:8080/api/knowledge/accuracy/breakout
```

---

## Goals

### List Goals

```bash
curl -H "Authorization: Bearer $TOKEN" http://localhost:8080/api/goals
```

### Create Goal

```bash
curl -X POST -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  http://localhost:8080/api/goals \
  -d '{
    "name": "Retirement Fund",
    "targetAmount": 5000000000,
    "targetDate": "2035-01-01",
    "category": "retirement"
  }'
```

### Update Goal

```bash
curl -X PUT -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  http://localhost:8080/api/goals/1 \
  -d '{"targetAmount": 6000000000}'
```

### Delete Goal

```bash
curl -X DELETE -H "Authorization: Bearer $TOKEN" http://localhost:8080/api/goals/1
```

### Goal Progress

```bash
curl -H "Authorization: Bearer $TOKEN" http://localhost:8080/api/goals/1/progress
```

---

## Backtest

```bash
curl -X POST -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  http://localhost:8080/api/backtest \
  -d '{
    "symbol": "FPT",
    "strategy": "sma_crossover",
    "startDate": "2023-01-01",
    "endDate": "2024-01-01",
    "initialCapital": 100000000
  }'
```

---

## Export

### Export Transactions (CSV)

```bash
curl -H "Authorization: Bearer $TOKEN" -o transactions.csv \
  http://localhost:8080/api/export/transactions
```

### Export Portfolio Snapshot

```bash
curl -H "Authorization: Bearer $TOKEN" -o snapshot.json \
  http://localhost:8080/api/export/snapshot
```

### Export Report

```bash
curl -H "Authorization: Bearer $TOKEN" -o report.json \
  http://localhost:8080/api/export/report
```

### Export Tax Report

```bash
curl -H "Authorization: Bearer $TOKEN" -o tax.json \
  http://localhost:8080/api/export/tax
```

---

## Signals

### Scan Signals

```bash
curl -H "Authorization: Bearer $TOKEN" http://localhost:8080/api/signals/scan
```

### Backtest Signals

```bash
curl -X POST -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  http://localhost:8080/api/signals/backtest \
  -d '{"period": "6M"}'
```

---

## Health & Metrics

### Health Check (public)

```bash
curl http://localhost:8080/api/health
```

**Response**: `{"status": "ok"}`

### Rate Limit Metrics

```bash
curl -H "Authorization: Bearer $TOKEN" http://localhost:8080/api/metrics/rate-limits
```

---

## Rate Limits

| Scope | Limit | Window |
|-------|-------|--------|
| Per IP (all endpoints) | 200 requests | 1 minute |
| Per user (authenticated) | 100 requests | 1 minute |
| VCI data source | 100 requests | 1 minute |
| KBS data source | 100 requests | 1 minute |
| CoinGecko | 50 requests | 1 minute |
| Doji (gold) | 60 requests | 1 minute |

When exceeded, the API returns `429 Too Many Requests`:
```json
{ "error": "rate limit exceeded — max 100 requests per minute", "code": "RATE_LIMIT_EXCEEDED" }
```

---

## Error Codes

| Code | HTTP Status | Description |
|------|-------------|-------------|
| `INVALID_CREDENTIALS` | 401 | Wrong username or password |
| `ACCOUNT_LOCKED` | 429 | Too many failed login attempts |
| `USER_NOT_FOUND` | 404 | User does not exist |
| `USER_EXISTS` | 409 | Username already taken |
| `INVALID_TOKEN` | 401 | JWT token is invalid |
| `TOKEN_EXPIRED` | 401 | JWT token has expired |
| `SESSION_EXPIRED` | 401 | Session timed out (4h inactivity) |
| `PASSWORD_MISMATCH` | 400 | Current password is incorrect |
| `RATE_LIMIT_EXCEEDED` | 429 | Too many requests |

Standard error response format:
```json
{ "error": "human-readable message", "code": "ERROR_CODE" }
```
