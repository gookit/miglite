# link https://github.com/humbug/box/blob/master/Makefile
#SHELL = /bin/sh
.DEFAULT_GOAL := help
# 每行命令之前必须有一个tab键。如果想用其他键，可以用内置变量.RECIPEPREFIX 声明
# mac 下这条声明 没起作用 !!
#.RECIPEPREFIX = >
.PHONY: all usage help clean

# 需要注意的是，每行命令在一个单独的shell中执行。这些Shell之间没有继承关系。
# - 解决办法是将两行命令写在一行，中间用分号分隔。
# - 或者在换行符前加反斜杠转义 \

# 接收命令行传入参数 make COMMAND tag=v2.0.4
# TAG=$(tag)

BIN_NAME=miglite
MAIN_SRC_FILE=cmd/miglite/main.go
#ROOT_PACKAGE := main
#VERSION=$(shell git for-each-ref refs/tags/ --count=1 --sort=-version:refname --format='%(refname:short)' 1 |  sed 's/^v//')
GO_VERSION := $(shell go version | sed -e 's/^[^0-9.]*\([0-9.]*\).*/\1/')

# git commit id
COMMIT_ID := $(shell git rev-parse HEAD 2> /dev/null || echo 'unknown')
SHORT_HASH := $(shell git rev-parse --short HEAD 2> /dev/null || echo 'unknown')
# set dev version unless VERSION is explicitly set via environment
# manual set: make VERSION=1.2.3
VERSION ?= $(shell echo "$$(git for-each-ref refs/tags/ --count=1 --sort=-version:refname --format='%(refname:short)' | echo 'dev' 2>/dev/null)-$(SHORT_HASH)" | sed 's/^v//')
BUILD_DATE := $(shell date +%Y/%m/%d-%H:%M:%S)

# Full build flags used when building binaries. Not used for test compilation/execution.
BUILD_FLAGS := -ldflags \
  " -s -w \
   -X main.Version=$(VERSION)\
   -X main.GoVersion=$(GO_VERSION)\
   -X main.BuildTime=$(BUILD_DATE)\
   -X main.GitCommit=$(COMMIT_ID)"

##there some make command for the project
##

help:
	@fgrep -h "##" $(MAKEFILE_LIST) | fgrep -v fgrep | sed -e 's/\\$$//' | sed -e 's/##//' | sed -e 's/: / /'

##Available Commands:

install: ## Install to GOPATH/bin(local dev)
	cd cmd/miglite && go install $(BUILD_FLAGS) .
	#chmod +x $(GOPATH)/bin/miglite

build-all: win linux linux-arm darwin darwin-arm ## Build for Linux,ARM,OSX,Windows

linux: ## Build for Linux AMD64
	mkdir -p build
	GOOS=linux GOARCH=amd64 go build $(BUILD_FLAGS) -o build/miglite-linux-amd64 $(MAIN_SRC_FILE)
	chmod +x build/miglite-linux-amd64

linux-arm: ## Build for ARM64
	mkdir -p build
	GOOS=linux GOARCH=arm64 go build $(BUILD_FLAGS) -o build/miglite-linux-arm $(MAIN_SRC_FILE)
	chmod +x build/miglite-linux-arm

darwin: ## Build for OSX AMD64
	mkdir -p build
	GOOS=darwin GOARCH=amd64 go build $(BUILD_FLAGS) -o build/miglite-darwin-amd64 $(MAIN_SRC_FILE)
	chmod +x build/miglite-darwin-amd64

darwin-arm: ## Build for OSX ARM64
	mkdir -p build
	GOOS=darwin GOARCH=arm64 go build $(BUILD_FLAGS) -o build/miglite-darwin-arm64 $(MAIN_SRC_FILE)
	chmod +x build/miglite-darwin-arm64

win: ## Build for Windows AMD64
	mkdir -p build
	GOOS=windows GOARCH=amd64 go build $(BUILD_FLAGS) -o build/miglite-windows-amd64.exe $(MAIN_SRC_FILE)

  clean:     ## Clean all created artifacts
clean:
	git clean --exclude=.idea/ -fdx

  cs-fix:        ## Fix code style for all files
cs-fix:
	gofmt -w ./

  cs-diff:        ## Display code style error files
cs-diff:
	gofmt -l ./
