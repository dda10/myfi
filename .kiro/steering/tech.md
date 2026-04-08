# Technology Stack

## Frontend

- Next.js 16 (App Router)
- React 19
- TypeScript 5
- Tailwind CSS 4
- Lightweight Charts (TradingView) for candlestick/volume charts
- Recharts for additional visualizations
- Framer Motion for animations
- Lucide React for icons

## Backend

- Go 1.26
- Gin web framework
- vnstock-go v2 (Vietnamese market data — multi-connector: VCI, KBS, VND, ENTRADE, CAFEF, VND_FINFO, MSN, GOLD, FMARKET)
- LangChain Go (AI/LLM integration)
- gRPC client to Python AI Service (port 50051, REST fallback on 8000)
- Domain-driven layout: `cmd/server/` + `internal/domain/` + `internal/platform/` + `internal/infra/`
- Package dependency direction: `cmd → platform → domain → infra` (acyclic, domains never import each other)

## Architecture

- Monorepo with separate frontend/backend directories
- RESTful API communication
- CORS enabled for cross-origin requests
- Context-based state management (React Context API)

## Common Commands

### Frontend (from `frontend/` directory)
```bash
npm run dev      # Start development server (localhost:3000)
npm run build    # Production build
npm start        # Start production server
npm run lint     # Run ESLint
```

### Backend (from `backend/` directory)
```bash
go run ./cmd/server  # Start development server (localhost:8080)
go build ./...       # Compile all packages
go build ./cmd/server # Compile server binary
go vet ./...         # Static analysis
go test ./...        # Run all tests
go mod tidy          # Clean up dependencies
```

### API Endpoints

Backend runs on port 8080 with these main routes:
- `GET /api/health` - Health check
- `GET /api/market/quote?symbols=SSI,FPT` - Real-time quotes
- `GET /api/market/chart?symbol=SSI&interval=1d` - Historical OHLCV data
- `GET /api/market/listing` - Available stocks
- `GET /api/market/screener` - Stock screener with filters
- `GET /api/market/search?q=FPT` - Global symbol search
- `GET /api/market/macro` - Macro indicators (interbank, bonds, VCB FX, gold)
- `GET /api/market/priceboard?symbols=SSI,FPT` - Live price board
- `POST /api/chat` - AI chat assistant (gRPC proxy to Python AI Service)
- `POST /api/screener` - Advanced screener (delegates to vnstock-go Screen)
- `GET /api/ranking` - AI stock ranking (gRPC proxy to Python AI Service)
- `POST /api/ranking` - Compute ranking with custom factors (gRPC proxy)
- `POST /api/ranking/backtest` - Run strategy backtest (gRPC proxy)
- `GET /api/ranking/recommendations` - Recommendation history with filters
- `GET /api/ranking/accuracy` - Recommendation accuracy metrics
- `GET /api/ideas` - Proactive investment ideas (gRPC proxy)
- `GET /api/market/world-indices` - World market indices (S&P 500, NASDAQ, Nikkei, etc. via MSN connector)
- `GET /api/market/gold` - SJC + BTMC gold prices (via GOLD connector)
- `GET /api/market/exchange-rates` - Vietcombank official FX rates
- `GET /api/market/status` - Market trading session status (HOSE/HNX/UPCOM)
- `GET /api/market/price-board?symbols=VNM,FPT` - Real-time bid/ask depth (30s cache)
- `GET /api/market/price-depth?symbol=VNM` - 3-level order book depth
- `GET /api/market/shareholders?symbol=VNM` - Major shareholders
- `GET /api/market/subsidiaries?symbol=VNM` - Subsidiary companies
- `GET /api/funds` - List mutual funds (via FMARKET connector)
- `GET /api/funds/:code` - Fund details (holdings + allocation)
- `GET /api/funds/:code/holdings` - Fund top stock holdings
- `GET /api/funds/:code/nav` - Fund NAV history
- `GET /api/docs` - Swagger UI (OpenAPI documentation)
- `GET /api/docs/swagger.json` - OpenAPI 3.0 spec
- `GET /metrics` - Prometheus metrics endpoint

### Observability

- Structured JSON logging via `slog.NewJSONHandler` to stdout (configurable via `LOG_LEVEL` env var)
- Prometheus metrics at `GET /metrics` (request latency, error rates, cache hit ratios, agent response times, data source availability, circuit breaker state)
- OpenTelemetry tracing with trace ID propagation across gRPC/REST calls
- JSON logger middleware replaces Gin's default text logger

### Testing

```bash
# Go Backend integration tests (from backend/)
go test ./internal/platform/ -v          # Integration tests (httptest, no DB required)
go test ./... -race                       # All tests with race detector

# Python AI Service tests (from ai-service/)
uv run --extra dev pytest tests/ -v       # All tests (uses MockChatModel in test mode)
EZISTOCK_TEST_MODE=true uv run --extra dev pytest  # Explicit test mode

# Frontend E2E tests (from frontend/)
npx playwright install                    # Install browsers (first time)
npm run test:e2e                          # Run Playwright E2E tests
npm run test:e2e:ui                       # Run with Playwright UI
```
