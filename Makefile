.PHONY: all build run clean \
	fe-install fe-build fe-run fe-lint fe-fmt \
	be-build be-run be-fmt be-test \
	wails-dev wails-build

# --------------------------------------------------
# Variables
# --------------------------------------------------
BE_OUT_DIR   := temp
BE_BINARY    := $(BE_OUT_DIR)/personal-digital-inventory.exe
FE_DIR       := frontend

# --------------------------------------------------
# Shell — use cmd.exe on Windows so mkdir/rmdir work
# --------------------------------------------------
ifeq ($(OS),Windows_NT)
SHELL  := cmd.exe
MKDIR   = if not exist "$(1)" mkdir "$(1)"
RMDIR   = if exist "$(1)" rmdir /s /q "$(1)"
else
MKDIR   = mkdir -p $(1)
RMDIR   = rm -rf $(1)
endif

# --------------------------------------------------
# All
# --------------------------------------------------
all: build

build: be-build fe-build

clean:
	$(call RMDIR,$(BE_OUT_DIR))
	$(call RMDIR,$(FE_DIR)/dist)
	$(call RMDIR,$(FE_DIR)/node_modules/.cache)

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
	$(call MKDIR,$(BE_OUT_DIR))
	go build -o $(BE_BINARY) .

be-run:
	go run .

be-wire:
	wire ./internal/infrastructure/di/...

be-fmt:
	go fmt ./...

be-fmt-check:
	@gofmt -l . && exit 0

be-test:
	go test ./... -v -count=1

be-test-cover:
	$(call MKDIR,$(BE_OUT_DIR))
	go test ./... -v -count=1 -coverprofile=$(BE_OUT_DIR)/coverage.out
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
