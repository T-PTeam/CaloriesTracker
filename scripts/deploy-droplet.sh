#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

HOST="${DROPLET_HOST:-68.183.77.29}"
USER="${DROPLET_USER:-root}"
REMOTE_DIR="${DROPLET_DIR:-/opt/calories-tracker}"
SSH_KEY="${DROPLET_SSH_KEY:-$HOME/.ssh/id_ed25519}"
ENV_FILE="${DROPLET_ENV_FILE:-}"

SSH_OPTS=(-o StrictHostKeyChecking=accept-new -o IdentitiesOnly=yes)
if [[ -n "$SSH_KEY" && -f "$SSH_KEY" ]]; then
  SSH_OPTS+=(-i "$SSH_KEY")
fi

ssh_cmd() {
  ssh "${SSH_OPTS[@]}" "${USER}@${HOST}" "$@"
}

echo "Deploying to ${USER}@${HOST}:${REMOTE_DIR}"
echo "Leaving other projects (e.g. cashnove) untouched."

if [[ -z "$ENV_FILE" ]]; then
  if [[ -f .env.droplet ]]; then
    ENV_FILE=.env.droplet
  elif [[ -f .env ]]; then
    ENV_FILE=.env
  else
    echo "Missing .env.droplet or .env — create from .env.prod.example first."
    exit 1
  fi
fi

if grep -qE 'sk-your-openai-key|123456:replace-me|replace-with-strong' "$ENV_FILE"; then
  echo "Refusing deploy: ${ENV_FILE} still has placeholder secrets."
  exit 1
fi

ssh_cmd "mkdir -p '${REMOTE_DIR}'"

RSYNC_SSH="ssh ${SSH_OPTS[*]}"
rsync -az --delete \
  --exclude '.git' \
  --exclude 'web/node_modules' \
  --exclude 'web/dist' \
  --exclude '.cursor' \
  --exclude '.env' \
  --exclude '.env.droplet' \
  --exclude '*.test' \
  -e "$RSYNC_SSH" \
  ./ "${USER}@${HOST}:${REMOTE_DIR}/"

scp "${SSH_OPTS[@]}" "$ENV_FILE" "${USER}@${HOST}:${REMOTE_DIR}/.env"

ssh_cmd "bash -s" <<EOF
set -euo pipefail
cd '${REMOTE_DIR}'

if ! command -v docker >/dev/null 2>&1; then
  curl -fsSL https://get.docker.com | sh
  systemctl enable --now docker
fi

if ! docker compose version >/dev/null 2>&1; then
  echo 'Docker Compose plugin is required'
  exit 1
fi

if [[ ! -f /swapfile ]]; then
  fallocate -l 2G /swapfile || dd if=/dev/zero of=/swapfile bs=1M count=2048
  chmod 600 /swapfile
  mkswap /swapfile
  swapon /swapfile
elif ! swapon --show | grep -q /swapfile; then
  swapon /swapfile || true
fi

docker compose -p calories-tracker -f docker-compose.prod.yml build web
docker compose -p calories-tracker -f docker-compose.prod.yml build api
docker compose -p calories-tracker -f docker-compose.prod.yml up -d --force-recreate
docker compose -p calories-tracker -f docker-compose.prod.yml ps
EOF

WEB_PORT="$(grep -E '^WEB_PORT=' "$ENV_FILE" | cut -d= -f2- || true)"
WEB_PORT="${WEB_PORT:-8088}"

echo
echo "App should be at http://${HOST}:${WEB_PORT}/"
echo "Set Telegram webhook (HTTPS required by Telegram):"
echo "  make set-webhook URL=https://YOUR_DOMAIN/telegram/webhook"
echo "If you only have the IP for now, put a domain on this droplet and terminate TLS (Caddy/Nginx + Let's Encrypt)."
