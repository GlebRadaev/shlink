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

M = $(shell printf "\033[34;1m▶\033[0m")
PROTOC_VER = 3.12.4
OS = linux
ifeq ($(shell uname -s), Darwin)
    OS = osx
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

.PHONY: run
run:
	@echo "Running application..."
	cd cmd/shortener && go run -ldflags="-X main.buildVersion=$(BUILD_VERSION) -X main.buildDate=$(BUILD_DATE) -X main.buildCommit=$(BUILD_COMMIT)" .

.PHONY: build
build:
	@echo "Building binary with version: $(BUILD_VERSION), date: $(BUILD_DATE), commit: $(BUILD_COMMIT)"
	cd cmd/shortener && go build -ldflags "-X main.buildVersion=$(BUILD_VERSION) -X main.buildDate=$(BUILD_DATE) -X main.buildCommit=$(BUILD_COMMIT)" -o shortener .

.PHONY: bin
bin: $(info $(M) install bin)
	@GOBIN=$(CURDIR)/bin go install -mod=mod \
        github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@v2.25.1 \
        github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@v2.25.1; \
    GOBIN=$(CURDIR)/bin go install -mod=mod \
        google.golang.org/protobuf/cmd/protoc-gen-go@v1.35.2; \
    GOBIN=$(CURDIR)/bin go install -mod=mod \
        google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.5.1; \
    GOBIN=$(CURDIR)/bin go install -mod=mod \
        github.com/envoyproxy/protoc-gen-validate@v1.1.0; \
    curl -Ls -o $(CURDIR)/bin/protoc.zip https://github.com/protocolbuffers/protobuf/releases/download/v${PROTOC_VER}/protoc-${PROTOC_VER}-${OS}-x86_64.zip; \
    mkdir $(CURDIR)/bin/.protoc; \
    unzip -q $(CURDIR)/bin/protoc.zip -d $(CURDIR)/bin/.protoc; \
    mv $(CURDIR)/bin/.protoc/bin/protoc $(CURDIR)/bin; \
    rm -rf $(CURDIR)/bin/.protoc; rm -rf $(CURDIR)/bin/protoc.zip;

.PHONY: gen
gen: $(info $(M) protoc gen)
	$(Q) echo "Generating shlink" && \
	mkdir -p $(CURDIR)/pkg/shlink && \
    $(CURDIR)/bin/protoc \
        --plugin=protoc-gen-grpc-gateway=$(CURDIR)/bin/protoc-gen-grpc-gateway \
        --plugin=protoc-gen-go-grpc=$(CURDIR)/bin/protoc-gen-go-grpc \
        --plugin=protoc-gen-validate=$(CURDIR)/bin/protoc-gen-validate \
        --plugin=protoc-gen-go=$(CURDIR)/bin/protoc-gen-go \
        -I$(CURDIR)/api:$(CURDIR)/vendor.pb \
        --go_out=$(CURDIR)/pkg \
        --validate_out=lang=go:$(CURDIR)/pkg \
        --go-grpc_out=$(CURDIR)/pkg \
		--experimental_allow_proto3_optional \
        --grpc-gateway_out=$(CURDIR)/pkg \
        $(CURDIR)/api/shlink.proto
