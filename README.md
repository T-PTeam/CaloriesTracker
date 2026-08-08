# Calories Tracker

Telegram bot + Go backend + OpenAI meal parser + PostgreSQL + React analytics dashboard.

## Environments

| Environment | DB | API | Web |
|-------------|----|-----|-----|
| **Dev Mac** | Docker Postgres (`docker-compose`) on `localhost:5432` | `go run ./cmd/bot` on `:8080` | Vite on `:5173` |
| **Droplet** | Docker Postgres (internal network) | Docker `api` service | Docker `web` (nginx) on port 80 |

Mac and droplet each have **their own** Postgres volume. They do not share data.

## Multi-user flow

1. Anyone can message the Telegram bot (no allowlist).
2. Bot stores meals per Telegram user.
3. User sends `/start` → gets a 6-digit code (15 min).
4. Register on the website with email, password, and that code.
5. Login → JWT session → analytics only for that user.

## Prerequisites

- Go 1.22+
- Node.js 20+
- Docker
- OpenAI API key
- Telegram bot token from [@BotFather](https://t.me/BotFather)

## Dev Mac (Docker DB)

```bash
cp .env.example .env
cp web/.env.example web/.env
```

Fill `OPENAI_API_KEY`, `TELEGRAM_BOT_TOKEN`, `API_SECRET` (and optional `JWT_SECRET`).

```bash
make db-up
make run
make web-dev
```

Health: `GET http://localhost:8080/api/health`  
Dashboard: `http://localhost:5173` → register/login

For local Telegram webhook testing, tunnel the Mac (`ngrok http 8080`) and set webhook to that URL.

## Droplet (DigitalOcean)

From your Mac (SSH key must already work on the droplet):

```bash
# ensure .env has real OPENAI_API_KEY, TELEGRAM_BOT_TOKEN, strong API_SECRET/JWT_SECRET/POSTGRES_PASSWORD
make deploy HOST=68.183.77.29 USER=root SSH_KEY=~/.ssh/your_key
```

Defaults: `HOST=68.183.77.29`, `USER=root`, remote dir `/opt/calories-tracker`.

Stack: Postgres + API + nginx web on port **80** (`docker-compose.prod.yml`).

- Dashboard: `http://68.183.77.29/`
- Health: `http://68.183.77.29/api/health`

Telegram **requires HTTPS** for webhooks. Point a domain at the droplet, put TLS in front (Caddy/Nginx + Let’s Encrypt), then:

```bash
make set-webhook URL=https://YOUR_DOMAIN/telegram/webhook
```

Until HTTPS is ready, the website can work on HTTP, but the bot will not receive updates via webhook.
## API

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/api/health` | none | Liveness + DB ping |
| POST | `/telegram/webhook` | Telegram | Bot updates |
| POST | `/api/auth/register` | none | email + password + link_code |
| POST | `/api/auth/login` | none | email + password → JWT |
| GET | `/api/auth/me` | Bearer JWT | Current user |
| GET | `/api/stats/summary?from=&to=` | Bearer JWT | Totals + daily series |
| GET | `/api/meals?from=&to=&limit=` | Bearer JWT | Recent meals |

## Project layout

```
cmd/bot/                 process entry
internal/config/         env config
internal/domain/         pure domain types
internal/parser/         OpenAI client + validation
internal/service/        use-cases (meals + auth)
internal/storage/postgres/
internal/telegram/       webhook + sendMessage
internal/httpapi/        REST for auth + analytics
migrations/
web/                     React analytics + auth UI
Dockerfile               API image (droplet)
docker-compose.yml       Mac: Postgres only
docker-compose.prod.yml  Droplet: postgres + api + web
```

## Tests

```bash
go test ./...
cd web && npm run build
cd web && npm run prettier:fix
```
