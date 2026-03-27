# MyFi Deployment Guide

## Environment Variables

### Required

| Variable | Description | Example |
|----------|-------------|---------|
| `DATABASE_URL` | PostgreSQL connection string | `postgres://myfi:secret@db:5432/myfi?sslmode=require` |
| `JWT_SECRET` | Secret key for signing JWT tokens | Random 64-char string |
| `FRONTEND_ORIGIN` | Allowed CORS origin | `https://myfi.example.com` |

### Optional

| Variable | Description | Default |
|----------|-------------|---------|
| `ENV` | Environment mode (`production` enables HTTPS redirect) | — |
| `PORT` | Backend server port | `8080` |
| `OPENAI_API_KEY` | OpenAI API key for AI chat | — |
| `ANTHROPIC_API_KEY` | Anthropic API key (alternative LLM) | — |
| `FMP_API_KEY` | Financial Modeling Prep API key | — |
| `VNSTOCK_CONNECTOR` | Default vnstock connector | `VCI` |
| `VNSTOCK_PROXY_URL` | HTTP proxy for vnstock requests | — |

---

## Database Setup

### Prerequisites

- PostgreSQL 16+

### Create Database

```bash
# Connect to PostgreSQL
psql -U postgres

# Create user and database
CREATE USER myfi WITH PASSWORD 'your_secure_password';
CREATE DATABASE myfi OWNER myfi;
GRANT ALL PRIVILEGES ON DATABASE myfi TO myfi;
\q
```

### Migrations

Migrations run automatically on server startup via `infra.InitDB()`. The server creates all required tables if they don't exist:

- `users` — User accounts and preferences
- `assets` — Portfolio holdings
- `transactions` — Buy/sell/dividend records
- `savings_accounts` — Savings deposit tracking
- `nav_snapshots` — Daily portfolio NAV history
- `watchlists` / `watchlist_symbols` — User watchlists
- `filter_presets` — Saved screener presets
- `alerts` / `alert_preferences` — Pattern alerts
- `pattern_observations` — Knowledge base observations
- `recommendation_audit_log` — AI recommendation tracking
- `financial_goals` — Goal planning
- `stock_sector_mapping` — Sector classification cache
- `cache_entries` — Server-side cache

### Connection Pool Settings

The backend configures the pool as:
- Max open connections: 25
- Max idle connections: 5
- Connection max lifetime: 5 minutes

---

## Build

### Backend

```bash
cd backend
go build -o myfi-server ./cmd/server
```

### Frontend

```bash
cd frontend
npm ci
npm run build
```

---

## Production Deployment

### Option 1: Docker Compose

See `docker-compose.prod.yml` and `docs/PRODUCTION.md` for the full production Docker setup.

```bash
# Build and start all services
docker compose -f docker-compose.prod.yml up -d --build

# Check logs
docker compose -f docker-compose.prod.yml logs -f backend
```

### Option 2: Manual Deployment

1. Provision a PostgreSQL 16 instance
2. Set environment variables (see table above)
3. Build the backend binary:
   ```bash
   cd backend && CGO_ENABLED=0 GOOS=linux go build -o myfi-server ./cmd/server
   ```
4. Build the frontend:
   ```bash
   cd frontend && npm ci && npm run build
   ```
5. Run the backend:
   ```bash
   ENV=production DATABASE_URL="postgres://..." JWT_SECRET="..." ./myfi-server
   ```
6. Run the frontend:
   ```bash
   cd frontend && npm start
   ```
7. Place a reverse proxy (Nginx/Caddy) in front for TLS termination.

---

## Monitoring & Logging

### Structured Logging

The backend uses `log/slog` with JSON output to stdout:

```json
{"time":"2025-01-15T10:00:00Z","level":"INFO","msg":"starting server","port":8080}
```

Collect logs with any log aggregator (Loki, CloudWatch, ELK) by capturing container stdout.

### Health Check

```bash
curl http://localhost:8080/api/health
# {"status": "ok"}
```

### Rate Limit Metrics

```bash
curl -H "Authorization: Bearer $TOKEN" http://localhost:8080/api/metrics/rate-limits
```

Returns per-source rate limit usage (VCI, KBS, CoinGecko, Doji).

### Recommended Monitoring

- **Uptime**: Poll `/api/health` every 30s
- **Database**: Monitor PostgreSQL connection count and query latency
- **Memory/CPU**: Container resource metrics via Docker stats or Prometheus
- **Error rate**: Track 5xx responses in reverse proxy access logs
- **Rate limits**: Monitor `RATE_LIMIT_EXCEEDED` occurrences in logs

---

## Backup

### Database Backup

```bash
# Full dump
pg_dump -U myfi -h localhost myfi > backup_$(date +%Y%m%d).sql

# Compressed
pg_dump -U myfi -h localhost myfi | gzip > backup_$(date +%Y%m%d).sql.gz

# Restore
psql -U myfi -h localhost myfi < backup_20250115.sql
```

### Automated Backups

Add a cron job for daily backups:

```bash
# /etc/cron.d/myfi-backup
0 2 * * * pg_dump -U myfi myfi | gzip > /backups/myfi_$(date +\%Y\%m\%d).sql.gz
# Keep last 30 days
0 3 * * * find /backups -name "myfi_*.sql.gz" -mtime +30 -delete
```
