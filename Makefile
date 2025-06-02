include .env
MIGRATION_PATH = ./cmd/migrate/migrations

.PHONY: migrate-create
migrate-create:
	@migrate create -seq -ext sql -dir $(MIGRATION_PATH) $(name)

.PHONY: migrate-up
migrate-up:
	@migrate -path=$(MIGRATION_PATH) -database=$(DB_ADDR) up

.PHONY: migrate-down
migrate-down:
	@migrate -path=$(MIGRATION_PATH) -database=$(DB_ADDR) down

.PHONY: migrate-down-steps
migrate-down-steps:
	@migrate -path=$(MIGRATION_PATH) -database=$(DB_ADDR) down $(steps)

.PHONY: migrate-version
migrate-version:
	@migrate -path=$(MIGRATION_PATH) -database=$(DB_ADDR) version

.PHONY: migrate-force
migrate-force:
	@migrate -path=$(MIGRATION_PATH) -database=$(DB_ADDR) force $(version)

.PHONY: seed
seed:
	@go run cmd/migrate/seed/main.go

.PHONY: gen-docs
gen-docs:
	@swag init -g ./api/main.go -d cmd,internal && swag fmt

.PHONY: test
test: 
	@go test -v ./...
