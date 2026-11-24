# Database migrations. Pick another service with SERVICE=auth, for example.
SERVICE ?= wallet
DB_URL 	?= postgres://payments:payments@localhost:5433/$(SERVICE)?sslmode=disable
MIGRATE	?= migrate -path db/$(SERVICE)/migrations -database "$(DB_URL)"

.PHONY: migrate-up migrate-down migrate-version migrate-new

migrate-up:
	@$(MIGRATE) up

migrate-down:
	@$(MIGRATE) down

migrate-version:
	@$(MIGRATE) version

migrate-new:
ifeq ($(strip $(name)),)
	$(error Usage: make migrate-new name=add_something)
endif
	migrate create -ext sql -dir db/$(SERVICE)/migrations -seq "$(name)"

.PHONY: generate test

# Regenerate the Go code for every service's SQL queries.
generate:
	@sqlc generate

# Integration tests use local Docker Postgres by default. Set the URL to empty to skip them.
WALLET_TEST_DATABASE_URL ?= postgres://payments:payments@localhost:5433/wallet?sslmode=disable
AUTH_TEST_DATABASE_URL ?= postgres://payments:payments@localhost:5433/auth?sslmode=disable
export WALLET_TEST_DATABASE_URL AUTH_TEST_DATABASE_URL

# Run part of the suite with PKGS=./internal/wallet TESTFLAGS='-run Transfer'.
PKGS ?= ./...

test:
	@go test -count=1 $(TESTFLAGS) $(PKGS)


.PHONY: run-wallet run-auth keys

# Run the wallet service against the local database, stamped with a Git version.
VERSION ?= $(shell git describe --always --dirty)

run-wallet: export WALLET_DATABASE_URL := $(DB_URL)
run-wallet:
	@docker compose up -d --wait postgres
	@go build -ldflags "-X main.version=$(VERSION)" -o bin/wallet.exe ./cmd/wallet
	@bin\wallet.exe

AUTH_DATABASE_URL ?= postgres://payments:payments@localhost:5433/auth?sslmode=disable
AUTH_PRIVATE_KEY_FILE ?= secrets/auth_ed25519.pem

run-auth: export AUTH_DATABASE_URL := $(AUTH_DATABASE_URL)
run-auth: export AUTH_PRIVATE_KEY_FILE := $(AUTH_PRIVATE_KEY_FILE)
run-auth:
	@docker compose up -d --wait postgres
	@go build -ldflags "-X main.version=$(VERSION)" -o bin/auth.exe ./cmd/auth
	@bin\auth.exe

keys:
	@mkdir -p secrets
	@test ! -e secrets/auth_ed25519.pem || (echo "secrets/auth_ed25519.pem exists, keeping it" && exit 1)
	@umask 077 && openssl genpkey -algorithm ed25519 -out secrets/auth_ed25519.pem
	@openssl pkey -in secrets/auth_ed25519.pem -pubout -out secrets/auth_ed25519.pub.pem
