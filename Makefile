ifneq (,$(wildcard .env))
    include .env
    export $(shell sed 's/=.*//' .env)
endif

MIGRATIONS_PATH=./migrations
GOOSE_BIN=goose

DB_DSN=${DATABASE_DSN}

ifeq ($(strip $(DB_DSN)),)
$(error DATABASE_DSN is not set. Please configure DATABASE_DSN in the .env file)
endif

.PHONY: migrate-up
migrate-up:
	@echo "Running migrations up..."
	$(GOOSE_BIN) -dir $(MIGRATIONS_PATH) postgres "$(DB_DSN)" up

.PHONY: migrate-down
migrate-down:
	@echo "Reverting migrations..."
	$(GOOSE_BIN) -dir $(MIGRATIONS_PATH) postgres "$(DB_DSN)" down

.PHONY: create-migration
create-migration:
	@read -p "Enter migration name: " name; \
	$(GOOSE_BIN) create "$$name" sql -dir $(MIGRATIONS_PATH)


.PHONY: lint
lint:
	@echo "Running golangci-lint..."
	golangci-lint run

.PHONY: test
test:
	@echo "Running tests..."
	go test -v ./...

.PHONY: vet
vet:
	@echo "Running go vet..."
	go vet ./...

.PHONY: coverage
coverage:
	@echo "Running coverage..." 
	go test ./... -v -coverprofile=coverage.txt -covermode=atomic && go tool cover -html=coverage.txt && rm -rf coverage.txt

.PHONY: build
build:
	@echo "Building binary..."
	cd cmd/shortener && go build .

