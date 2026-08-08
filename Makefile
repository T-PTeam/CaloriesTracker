.PHONY: db-up db-down run test web-dev web-build env-local prod-up prod-down prod-logs set-webhook deploy

db-up:
	docker-compose up -d

db-down:
	docker-compose down

run:
	go run ./cmd/bot

test:
	go test ./...

web-dev:
	cd web && npm run dev

web-build:
	cd web && npm run build

env-local:
	cp -n .env.example .env || true
	cp -n web/.env.example web/.env || true

prod-up:
	docker-compose -f docker-compose.prod.yml up -d --build

prod-down:
	docker-compose -f docker-compose.prod.yml down

prod-logs:
	docker-compose -f docker-compose.prod.yml logs -f

set-webhook:
	@chmod +x ./scripts/set-webhook.sh
	@./scripts/set-webhook.sh "$(URL)"

deploy:
	@chmod +x ./scripts/deploy-droplet.sh
	@DROPLET_HOST="$(HOST)" DROPLET_USER="$(USER)" DROPLET_SSH_KEY="$(SSH_KEY)" ./scripts/deploy-droplet.sh
