BINARY_NAME=kawaii-wify
VERSION=0.1.0
LDFLAGS=-s -w -X main.Version=$(VERSION)

# Detect Host OS and set binary extension
ifeq ($(OS),Windows_NT)
HOST_OS ?= windows
EXT = .exe
RUN_PREFIX =
else
HOST_OS ?= $(shell go env GOOS 2>/dev/null || uname -s | tr '[:upper:]' '[:lower:]')
EXT =
RUN_PREFIX = ./
endif

# Target binary for the current host OS
CURRENT_BIN = bin/debug/$(BINARY_NAME)-$(HOST_OS)$(EXT)

.PHONY: all build run clean \
        build-linux build-linux-release \
        build-windows build-windows-release

all: build

# Automatically builds for your current OS
build:
	@go build -o "$(CURRENT_BIN)" ./cmd/kawaii-wify

# Automatically builds and runs for current OS with flags support (e.g. make run ARGS="...")
run: build
	@$(RUN_PREFIX)$(CURRENT_BIN) $(ARGS)

# --- Linux Targets ---
build-linux-release:
	@GOOS=linux GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o "bin/release/$(BINARY_NAME)-linux_$(VERSION)" ./cmd/kawaii-wify

build-linux:
	@GOOS=linux GOARCH=amd64 go build -o "bin/debug/$(BINARY_NAME)-linux" ./cmd/kawaii-wify

# --- Windows Targets ---
build-windows-release:
	@GOOS=windows GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o "bin/release/$(BINARY_NAME)-windows_$(VERSION).exe" ./cmd/kawaii-wify

build-windows:
	@GOOS=windows GOARCH=amd64 go build -o "bin/debug/$(BINARY_NAME)-windows.exe" ./cmd/kawaii-wify

clean:
	@rm -rf bin/
