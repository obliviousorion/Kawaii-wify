# Kawaii-Wify

A lightweight background session manager and automated login daemon for FortiOS (FortiGate) captive portals, specifically designed for campus WLAN environments like BITS Pilani.

## Overview

Campus Wi-Fi networks protected by FortiGate firewalls require users to authenticate through a web captive portal. In practice, leases expire frequently, connection drops occur when roaming across access points, and devices are abruptly disconnected from the WAN.

Kawaii-Wify runs quietly in the background to handle this lifecycle automatically:
- Probes WAN status using lightweight HTTP endpoints.
- Detects captive portal interception and extracts gateway challenge tokens.
- Authenticates with the gateway and captures keepalive session tokens.
- Periodically pings the firewall keepalive endpoint to prevent idle timeouts.
- Incorporates a circuit breaker state machine with automated cooldowns to prevent account lockouts.
- Gracefully handles termination signals (SIGINT, SIGTERM).

## Architecture

The system is organized into an authentication library, a thread-safe state engine, and an entrypoint daemon:

```
kawaii-wify/
├── cmd/
│   └── kawaii-wify/
│       ├── main.go            # Daemon entrypoint and signal handling
│       └── main_script.go     # Procedural test script
├── internal/
│   ├── auth/
│   │   ├── client.go          # HTTP transport with custom TLS verification
│   │   └── portal.go          # FortiOS handshake primitives (Probe, Prime, Login, Keepalive, Logout)
│   ├── credentials/
│   │   └── store.go           # OS keyring credentials management
│   └── engine/
│       ├── engine.go          # State machine, ticker loop, circuit breaker
│       └── state.go           # State types and transitions
├── Makefile
└── go.mod
```

### State Machine

The core `Engine` operates as a finite state machine:

1. `Offline`: No route to the internet or gateway probe failed.
2. `Captive Portal Detected`: WAN probe intercepted by FortiGate. The engine extracts the challenge token, primes the gateway session, and submits login credentials.
3. `Online`: WAN access confirmed. The engine sends periodic keepalive requests using the active session token to maintain the lease.
4. `Cooldown`: Triggered after consecutive authentication failures (default: 3). Halts login attempts for a defined duration (10s) before resetting to prevent firewall lockouts, then self-heals back to `Offline` for a fresh probe.

### Security and TLS Handling

FortiGate captive portals often serve self-signed certificates or certificates issued by campus internal CAs. Rather than blindly disabling TLS checks with `InsecureSkipVerify: true`, Kawaii-Wify configures a `VerifyConnection` callback that inspects the peer certificate chain presented by the remote server to verify that the Common Name or DNS matches `fw.bits-pilani.ac.in`.

## Requirements

- Go 1.22 or newer
- Linux with NetworkManager or systemd (cross-compilation supported via Go)

## Building

A `Makefile` is included for development and release builds:

```bash
# Debug build (outputs to bin/debug/kawaii-wify-linux)
make build-linux

# Optimized release build (stripped debug symbols)
make build-linux-release

# Build and run immediately
make run

# Clean build artifacts
make clean
```

## Running as a Service

Kawaii-Wify can be run in the background or supervised via a systemd user unit:

```ini
# ~/.config/systemd/user/kawaii-wify.service
[Unit]
Description=Kawaii-Wify FortiGate WLAN Session Manager
After=network.target

[Service]
ExecStart=%h/bin/kawaii-wify-linux
Restart=always
RestartSec=5

[Install]
WantedBy=default.target
```

Enable and start the service:

```bash
systemctl --user daemon-reload
systemctl --user enable --now kawaii-wify.service
```

## Roadmap

- OS Keyring integration (`github.com/zalando/go-keyring`) for secure, non-plaintext credential storage.
- CLI subcommands (`login`, `logout`, `status`, `daemon`).
- Network interface / SSID binding so probes only run when connected to targeted campus SSIDs.
- System tray status indicator.

## License

MIT
