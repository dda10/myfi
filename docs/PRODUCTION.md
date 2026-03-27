# Production Setup Guide

## Prerequisites

- Docker 24+ and Docker Compose v2
- Domain name with DNS configured
- SSL certificate (Let's Encrypt recommended)

## Environment Variables

Create a `.env` file in the `myfi/` directory:

```env
# Database
POSTGRES_DB=myfi
POSTGRES_USER=myfi
POSTGRES_PASSWORD=<strong-random-password>

# Auth
JWT_SECRET=<64-char-random-string>

# API URLs
NEXT_PUBLIC_API_URL=https://api.yourdomain.com

# AI providers (optional)
OPENAI_API_KEY=sk-...
ANTHROPIC_API_KEY=sk-ant-...
```

## Deployment

```bash
# Build and start all services
docker compose -f docker-compose.prod.yml up -d --build

# View logs
docker compose -f docker-compose.prod.yml logs -f

# Stop
docker compose -f docker-compose.prod.yml down
```

## Database Migrations

Migrations run automatically on backend startup. To run manually:

```bash
docker compose -f docker-compose.prod.yml exec backend ./server migrate
```

## SSL / HTTPS

Use a reverse proxy (nginx or Caddy) in front of the services:

```nginx
server {
    listen 443 ssl;
    server_name api.yourdomain.com;
    ssl_certificate /etc/letsencrypt/live/yourdomain.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/yourdomain.com/privkey.pem;
    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

## Monitoring & Logging

- Backend logs to stdout in JSON format — collect with Loki or CloudWatch
- Health endpoint: `GET /api/health`
- Rate limit metrics: `GET /api/metrics/rate-limits`

## Backup Strategy

```bash
# Daily PostgreSQL backup
docker compose -f docker-compose.prod.yml exec postgres \
  pg_dump -U myfi myfi | gzip > backup-$(date +%Y%m%d).sql.gz
```

Schedule with cron: `0 2 * * * /path/to/backup.sh`
