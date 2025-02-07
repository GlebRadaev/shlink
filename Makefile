ifneq (,$(wildcard .env))
    include .env
    export $(shell sed 's/=.*//' .env)
endif

MIGRATIONS_PATH ?= migrations
GOOSE_BIN=goose

DB_DSN=${DATABASE_DSN}

ifeq ($(strip $(DB_DSN)),)
$(error DATABASE_DSN is not set. Please configure DATABASE_DSN in the .env file)
endif

BUILD_VERSION ?= $(shell git describe --tags --abbrev=0 2>/dev/null || echo "N/A")
BUILD_DATE ?= $(shell date +'%Y-%m-%d_%H:%M:%S')
BUILD_COMMIT ?= $(shell git rev-parse HEAD 2>/dev/null || echo "N/A")

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

.PHONY: run
run:
	@echo "Running application..."
	cd cmd/shortener && go run -ldflags="-X main.buildVersion=$(BUILD_VERSION) -X main.buildDate=$(BUILD_DATE) -X main.buildCommit=$(BUILD_COMMIT)" .

.PHONY: build
build:
	@echo "Building binary with version: $(BUILD_VERSION), date: $(BUILD_DATE), commit: $(BUILD_COMMIT)"
	cd cmd/shortener && go build -ldflags "-X main.buildVersion=$(BUILD_VERSION) -X main.buildDate=$(BUILD_DATE) -X main.buildCommit=$(BUILD_COMMIT)" -o shortener .