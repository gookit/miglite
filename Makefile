## Miglite — Makefile

APP     := miglite
MAIN_DIR := ./cmd/miglite
GOEXE = $(shell go env GOEXE)
BINARY  := $(APP)$(GOEXE)

# Build metadata
BUILD_TIME := $(shell date +%Y-%m-%dT%H:%M:%S)
GIT_HASH  := $(shell git rev-parse --short=8 HEAD 2>/dev/null || echo "unknown")
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null | sed 's/^v//' || echo "dev-$(GIT_HASH)")
GO_VERSION := $(shell go version | sed -e 's/^[^0-9.]*\([0-9.]*\).*/\1/')

LDFLAGS := -s -w \
	-X main.Version=$(VERSION) \
	-X main.GitCommit=$(GIT_HASH) \
	-X main.GoVersion=$(GO_VERSION) \
	-X main.BuildTime=$(BUILD_TIME)

.PHONY: all build backend clean help latest

## all: build (default)
all: build

## build: build Go binary (current platform)
build:
	@echo "🐹 Building Go binary ($(VERSION) @ $(GIT_HASH))..."
	cd $(MAIN_DIR) && go build -ldflags "$(LDFLAGS)" -o $(BINARY) .
	@echo "📦 Compressing binary..."
	@upx -6 --no-progress $(BINARY)
	@echo "✅ Binary: $(BINARY) ($$(du -sh $(BINARY) | cut -f1))"

## install: install Go binary to $GOPATH/bin
install:
	cd $(MAIN_DIR) && go install -ldflags "$(LDFLAGS)" .
	upx -6 --no-progress $(GOPATH)/bin/$(BINARY)
	@echo "✅ Installed to GOPATH/bin"

## run: build and run with current directory
run: build
	./$(BINARY)

# ─── Cross Compilation ────────────────────────────────────────────────────────

DIST_DIR := dist
DIST_DIR_PATH := cmd/dist

## build-all: cross-compile for all platforms
build-all: dump-info build-linux build-linux-arm64 build-darwin build-darwin-arm64 build-windows latest-yaml
	ls -lh $(DIST_DIR_PATH)

## dump-info: dump build info
dump-info:
	@echo "Build Info:"
	@echo "  VERSION: $(VERSION)"
	@echo "  GIT_HASH: $(GIT_HASH)"
	@echo "  BUILD_TIME: $(BUILD_TIME)"

## latest-yaml: generate latest.yaml release metadata
latest-yaml:
	@mkdir -p $(DIST_DIR_PATH)
	@{ \
		echo "name: $(APP)"; \
		echo "version: $(VERSION)"; \
		echo "released_at: $(BUILD_TIME)"; \
	} > $(DIST_DIR_PATH)/latest.yaml
	@echo "   → $(DIST_DIR_PATH)/latest.yaml"

## build-linux: compile for Linux amd64
build-linux:
	@echo "🐧 linux/amd64..."
	@mkdir -p $(DIST_DIR_PATH)
	@cd $(MAIN_DIR) && GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o ../$(DIST_DIR)/$(APP)-linux-amd64 .
	@upx -6 --no-progress $(DIST_DIR_PATH)/$(APP)-linux-amd64
	@chmod +x $(DIST_DIR_PATH)/$(APP)-linux-amd64
	@echo "   → $(DIST_DIR_PATH)/$(APP)-linux-amd64"

## build-linux-arm64: compile for Linux arm64
build-linux-arm64:
	@echo "🐧 linux/arm64..."
	@mkdir -p $(DIST_DIR_PATH)
	@cd $(MAIN_DIR) && GOOS=linux GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o ../$(DIST_DIR)/$(APP)-linux-arm64 .
	@upx -6 --no-progress $(DIST_DIR_PATH)/$(APP)-linux-arm64
	@chmod +x $(DIST_DIR_PATH)/$(APP)-linux-arm64
	@echo "   → $(DIST_DIR_PATH)/$(APP)-linux-arm64"

## build-darwin: compile for macOS amd64
build-darwin:
	@echo "🍎 darwin/amd64..."
	@mkdir -p $(DIST_DIR_PATH)
	@cd $(MAIN_DIR) && GOOS=darwin GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o ../$(DIST_DIR)/$(APP)-darwin-amd64 .
	@chmod +x $(DIST_DIR_PATH)/$(APP)-darwin-amd64
	@echo "   → $(DIST_DIR_PATH)/$(APP)-darwin-amd64"

## build-darwin-arm64: compile for macOS Apple Silicon
build-darwin-arm64:
	@echo "🍎 darwin/arm64..."
	@mkdir -p $(DIST_DIR_PATH)
	@cd $(MAIN_DIR) && GOOS=darwin GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o ../$(DIST_DIR)/$(APP)-darwin-arm64 .
	# upx -6 --no-progress $(DIST_DIR_PATH)/$(APP)-darwin-arm64 # 压缩有问题在 macos 12+
	@echo "   → $(DIST_DIR_PATH)/$(APP)-darwin-arm64"

## build-windows: compile for Windows amd64
build-windows:
	@echo "🪟 windows/amd64..."
	@mkdir -p $(DIST_DIR_PATH)
	@cd $(MAIN_DIR) && GOOS=windows GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o ../$(DIST_DIR)/$(APP)-windows-amd64.exe .
	@upx -6 --no-progress $(DIST_DIR_PATH)/$(APP)-windows-amd64.exe
	@echo "   → $(DIST_DIR_PATH)/$(APP)-windows-amd64.exe"

.PHONY: release
release: build-all ## Create release archives for all platforms TODO 还未启用的
	@echo "Creating release archives..."
	@mkdir -p cmd/release
	@cd $(DIST_DIR) && \
	tar -czf ../release/$(APP)-linux-amd64.tar.gz $(APP)-linux-amd64; \
	tar -czf ../release/$(APP)-linux-arm64.tar.gz $(APP)-linux-arm64; \
	tar -czf ../release/$(APP)-darwin-amd64.tar.gz $(APP)-darwin-amd64; \
	tar -czf ../release/$(APP)-darwin-arm64.tar.gz $(APP)-darwin-arm64; \
	zip ../release/$(APP)-windows-amd64.zip $(APP)-windows-amd64.exe;
	@echo "Release archives created in cmd/release/"

## clean: remove build artifacts
clean:
	@rm -f $(BINARY)
	@rm -rf $(DIST_DIR)
	@echo "🧹 Cleaned"

## help: show this help
help:
	@echo "Skillc Build System"
	@echo ""
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/## /  /'
