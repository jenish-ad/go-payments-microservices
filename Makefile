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
export WALLET_TEST_DATABASE_URL

test:
	@go test -count=1 $(TESTFLAGS) ./...
