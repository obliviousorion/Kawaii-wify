BINARY_NAME=bits-wlan-manager
VERSION=0.1.0
LDFLAGS=-s -w -X main.Version=$(VERSION)

.PHONY: build-linux-release build-linux run clean

build-linux-release: # release build
	@GOOS=linux GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o "bin/release/$(BINARY_NAME)-linux_$(VERSION)" ./cmd/bits-wlan-manager

build-linux: # default is debug build
	@GOOS=linux GOARCH=amd64 go build -o "bin/debug/$(BINARY_NAME)-linux" ./cmd/bits-wlan-manager

run: build-linux # builds and runs the binary
	@./bin/debug/$(BINARY_NAME)-linux

clean:
	@rm -rf bin/


