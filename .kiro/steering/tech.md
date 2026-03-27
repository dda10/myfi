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
- VNStock Go client (Vietnamese market data)
- LangChain Go (AI/LLM integration)
- Go Server project layout (`cmd/server/` + `internal/` sub-packages)
- Package dependency direction: `cmd → handler → service → infra → model`

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
- `GET /api/crypto/quote` - Cryptocurrency quotes
- `GET /api/news` - Financial news
- `POST /api/chat` - AI chat assistant
