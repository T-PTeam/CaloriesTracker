#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

if [[ -f .env ]]; then
  set -a
  # shellcheck disable=SC1091
  source .env
  set +a
fi

LOG="$ROOT/.cursor/debug-77e188.log"
mkdir -p "$ROOT/.cursor"

# #region agent log
python3 - <<'PY' || true
import json, time, os
from pathlib import Path
p = Path('/Users/root1/Projects/CaloriesTracker/.cursor/debug-77e188.log')
token = os.environ.get('TELEGRAM_BOT_TOKEN', '')
entry = {
  'sessionId': '77e188',
  'runId': 'set-webhook',
  'hypothesisId': 'H1',
  'location': 'scripts/set-webhook.sh',
  'message': 'token loaded for setWebhook',
  'data': {
    'present': bool(token),
    'length': len(token),
    'looks_placeholder': token in ('', '123456:replace-me') or token.startswith('123456:'),
    'has_webhook_url': bool(os.environ.get('WEBHOOK_URL', '')),
  },
  'timestamp': int(time.time()*1000),
}
p.parent.mkdir(parents=True, exist_ok=True)
with p.open('a') as f:
  f.write(json.dumps(entry) + '\n')
PY
# #endregion

WEBHOOK_URL="${1:-${WEBHOOK_URL:-}}"
if [[ -z "${TELEGRAM_BOT_TOKEN:-}" ]]; then
  echo "TELEGRAM_BOT_TOKEN is empty. Put your BotFather token in .env"
  exit 1
fi
if [[ "$TELEGRAM_BOT_TOKEN" == "123456:replace-me" || "$TELEGRAM_BOT_TOKEN" == 123456:* ]]; then
  echo "TELEGRAM_BOT_TOKEN in .env is still a placeholder. Replace it with the real BotFather token."
  exit 1
fi
if [[ -z "$WEBHOOK_URL" ]]; then
  echo "Usage: make set-webhook URL=https://abcd1234.ngrok-free.app/telegram/webhook"
  echo "Copy the real https URL from the ngrok terminal (not a placeholder)."
  exit 1
fi
if [[ "$WEBHOOK_URL" == *"xxxx.ngrok"* || "$WEBHOOK_URL" == *"PASTE_YOUR"* || "$WEBHOOK_URL" == *"YOUR_REAL_HOST"* || "$WEBHOOK_URL" == *"YOUR_TUNNEL_HOST"* ]]; then
  echo "That URL is still a placeholder."
  echo "Open the ngrok terminal, copy the https://....ngrok-free.app address, then run:"
  echo "  make set-webhook URL=https://YOUR_COPIED_HOST/telegram/webhook"
  exit 1
fi
case "$WEBHOOK_URL" in
  https://*) ;;
  *)
    echo "Webhook URL must start with https://"
    exit 1
    ;;
esac

echo "Checking bot token via getMe..."
ME="$(curl -sS "https://api.telegram.org/bot${TELEGRAM_BOT_TOKEN}/getMe")"
echo "$ME" | python3 -c 'import sys,json; d=json.load(sys.stdin); print("ok=", d.get("ok"), "username=", (d.get("result") or {}).get("username")); raise SystemExit(0 if d.get("ok") else 1)'

echo "Setting webhook to: $WEBHOOK_URL"
RESP="$(curl -sS -X POST "https://api.telegram.org/bot${TELEGRAM_BOT_TOKEN}/setWebhook" -d "url=${WEBHOOK_URL}")"
echo "$RESP"

# #region agent log
python3 - <<PY
import json, time, os
from pathlib import Path
resp = '''$RESP'''
try:
  d = json.loads(resp)
except Exception:
  d = {'raw': resp[:200]}
entry = {
  'sessionId': '77e188',
  'runId': 'set-webhook',
  'hypothesisId': 'H4',
  'location': 'scripts/set-webhook.sh:result',
  'message': 'setWebhook response',
  'data': {
    'ok': d.get('ok'),
    'error_code': d.get('error_code'),
    'description': d.get('description'),
  },
  'timestamp': int(time.time()*1000),
}
p = Path('/Users/root1/Projects/CaloriesTracker/.cursor/debug-77e188.log')
with p.open('a') as f:
  f.write(json.dumps(entry) + '\n')
PY
# #endregion

echo "$RESP" | python3 -c 'import sys,json; d=json.load(sys.stdin); raise SystemExit(0 if d.get("ok") else 1)'
