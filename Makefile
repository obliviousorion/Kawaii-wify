BINARY_NAME=kawaii-wify
VERSION=0.1.0
TARGET_ARCH ?= amd64
LDFLAGS=-s -w -X main.Version=$(VERSION)

# Detect Host OS and set binary extension
ifeq ($(OS),Windows_NT)
HOST_OS ?= windows
EXT = .exe
RUN_PREFIX =
CLEAN_CMD = cmd /c "if exist bin rmdir /s /q bin"
else
HOST_OS ?= $(shell go env GOOS 2>/dev/null || uname -s | tr '[:upper:]' '[:lower:]')
EXT =
RUN_PREFIX = ./
CLEAN_CMD = rm -rf bin/
endif

# Target binary for the current host OS
CURRENT_BIN = bin/debug/$(BINARY_NAME)-$(HOST_OS)$(EXT)

.PHONY: all build run install clean build-all \
        build-linux build-linux-release \
        build-windows build-windows-release \
        build-darwin build-darwin-release

all: build

# Automatically builds for your current OS
build:
	@go build -ldflags="$(LDFLAGS)" -o "$(CURRENT_BIN)" ./cmd/kawaii-wify

# Compiles and installs binary globally to Go binary directory
install:
	@go install -ldflags="$(LDFLAGS)" ./cmd/kawaii-wify
	@echo "Installed $(BINARY_NAME) v$(VERSION) to $$(GOBIN=$$(go env GOBIN 2>/dev/null); [ -n \"$$GOBIN\" ] && echo \"$$GOBIN\" || echo \"$$(go env GOPATH)/bin\")/$(BINARY_NAME)"

# Automatically builds and runs for current OS with flags support (e.g. make run ARGS="...")
run: build
	@$(RUN_PREFIX)$(CURRENT_BIN) $(ARGS)

# Compile release binaries for all platforms
build-all: build-linux-release build-windows-release build-darwin-release

# --- Linux Targets ---
build-linux-release: export GOOS = linux
build-linux-release: export GOARCH = $(TARGET_ARCH)
build-linux-release:
	@go build -ldflags="$(LDFLAGS)" -o "bin/release/$(BINARY_NAME)-linux_$(VERSION)" ./cmd/kawaii-wify

build-linux: export GOOS = linux
build-linux: export GOARCH = $(TARGET_ARCH)
build-linux:
	@go build -o "bin/debug/$(BINARY_NAME)-linux" ./cmd/kawaii-wify

# --- Windows Targets ---
build-windows-release: export GOOS = windows
build-windows-release: export GOARCH = $(TARGET_ARCH)
build-windows-release:
	@go build -ldflags="$(LDFLAGS)" -o "bin/release/$(BINARY_NAME)-windows_$(VERSION).exe" ./cmd/kawaii-wify

build-windows: export GOOS = windows
build-windows: export GOARCH = $(TARGET_ARCH)
build-windows:
	@go build -o "bin/debug/$(BINARY_NAME)-windows.exe" ./cmd/kawaii-wify

# --- macOS (Darwin) Targets ---
build-darwin-release: export GOOS = darwin
build-darwin-release: export GOARCH = $(TARGET_ARCH)
build-darwin-release:
	@go build -ldflags="$(LDFLAGS)" -o "bin/release/$(BINARY_NAME)-darwin_$(VERSION)" ./cmd/kawaii-wify

build-darwin: export GOOS = darwin
build-darwin: export GOARCH = $(TARGET_ARCH)
build-darwin:
	@go build -o "bin/debug/$(BINARY_NAME)-darwin" ./cmd/kawaii-wify

clean:
	@$(CLEAN_CMD)
