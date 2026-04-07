.PHONY: up down dev test verify smoke migrate

up:
	docker compose up -d mysql redis

down:
	docker compose down

dev:
	go run ./cmd/server -config configs/config.local.yaml

test:
	go test ./...

verify:
	bash scripts/verify.sh

smoke:
	bash scripts/smoke.sh

migrate:
	go run ./cmd/migrate -config configs/config.local.yaml
