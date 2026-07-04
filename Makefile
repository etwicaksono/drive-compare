.PHONY: all build run clean \
	fe-install fe-build fe-run fe-lint fe-fmt \
	be-build be-run be-fmt be-test \
	wails-dev wails-build

# --------------------------------------------------
# Variables
# --------------------------------------------------
BE_OUT_DIR   := ./temp
BE_BINARY    := $(BE_OUT_DIR)/personal-digital-inventory.exe
FE_DIR       := ./frontend

# --------------------------------------------------
# All
# --------------------------------------------------
all: build

build: be-build fe-build

clean:
	rm -rf $(BE_OUT_DIR)
	rm -rf $(FE_DIR)/dist
	rm -rf $(FE_DIR)/node_modules/.cache

# --------------------------------------------------
# Frontend
# --------------------------------------------------
fe-install:
	cd $(FE_DIR) && npm install

fe-build:
	cd $(FE_DIR) && npm run build

fe-run:
	cd $(FE_DIR) && npm run dev

fe-lint:
	cd $(FE_DIR) && npx eslint .

fe-lint-fix:
	cd $(FE_DIR) && npx eslint . --fix

fe-fmt:
	cd $(FE_DIR) && npx prettier --write "src/**/*.{ts,tsx,css,json}"

fe-fmt-check:
	cd $(FE_DIR) && npx prettier --check "src/**/*.{ts,tsx,css,json}"

# --------------------------------------------------
# Backend
# --------------------------------------------------
be-build:
	@mkdir -p $(BE_OUT_DIR)
	go build -o $(BE_BINARY) ./cmd/...

be-run:
	go run ./cmd/...

be-wire:
	wire ./internal/infrastructure/di/...

be-fmt:
	go fmt ./...

be-fmt-check:
	@test -z "$$(gofmt -l .)" || (echo "Files not formatted:" && gofmt -l . && exit 1)

be-test:
	go test ./... -v -race -count=1

be-test-cover:
	go test ./... -v -race -count=1 -coverprofile=$(BE_OUT_DIR)/coverage.out
	go tool cover -html=$(BE_OUT_DIR)/coverage.out -o $(BE_OUT_DIR)/coverage.html

be-lint:
	golangci-lint run ./...

# --------------------------------------------------
# Wails (full app)
# --------------------------------------------------
wails-dev:
	wails dev

wails-build:
	wails build
