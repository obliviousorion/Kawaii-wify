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
- Securely retrieves and caches credentials via the OS Keyring (Linux Secret Service, macOS Keychain, Windows Credential Manager).
- Gracefully handles termination signals (SIGINT, SIGTERM).

## Architecture

The system is organized into an authentication library, a credentials management layer, a thread-safe state engine, and an entrypoint daemon:

```
kawaii-wify/
├── cmd/
│   └── kawaii-wify/
│       ├── main.go            # Daemon entrypoint, flag parsing, signal handling
│       └── main_script.go     # Procedural test script
├── internal/
│   ├── auth/
│   │   ├── client.go          # HTTP transport with custom TLS verification
│   │   └── portal.go          # FortiOS handshake primitives (Probe, Prime, Login, Keepalive, Logout)
│   ├── credentials/
│   │   ├── store.go           # OS keyring storage wrapper (Set, Get, Delete)
│   │   ├── prompt.go          # TTY-aware masked terminal input
│   │   └── resolver.go        # Cascading credential resolution logic
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

### Credential Resolution

Credentials are resolved using a secure cascade so passwords never have to be passed via shell arguments or stored in plaintext files:

1. **Environment Variables**: If both `KAWAII_USER` and `KAWAII_PASS` are set, they take immediate precedence.
2. **CLI User Flag + Keyring**: If `-u <username>` or `KAWAII_USER` is specified, the system queries the OS Keyring (via D-Bus Secret Service on Linux) for the stored secret.
3. **Interactive Terminal Prompt**: If no credentials exist or none are found in the keyring, an interactive prompt requests the student ID and securely masks password input via `golang.org/x/term`. Once verified, the credentials are automatically saved to the keyring for future unattended executions.

### Security and TLS Handling

FortiGate captive portals often serve self-signed certificates or certificates issued by campus internal CAs. Rather than blindly disabling TLS checks with `InsecureSkipVerify: true`, Kawaii-Wify configures a `VerifyConnection` callback that inspects the peer certificate chain presented by the remote server to verify that the Common Name or DNS matches `fw.bits-pilani.ac.in`.

Additionally, `TLSNextProto` is initialized to an empty map to disable HTTP/2 ALPN negotiation, forcing clean HTTP/1.1 communication over TLS to prevent protocol hangs with FortiOS port 8090 web servers.

## Requirements

- Go 1.22 or newer
- Linux with NetworkManager or systemd (libsecret / GNOME Keyring / KWallet for keyring storage)

## Building

A `Makefile` is included for development and release builds:

```bash
# Debug build (outputs to bin/debug/kawaii-wify-linux)
make build-linux

# Optimized release build (stripped debug symbols)
make build-linux-release

# Run debug build
make run

# Clean build artifacts
make clean
```

## Usage

### Run with a Specified Username
```bash
# Uses stored keyring password, or prompts if running for the first time
./bin/debug/kawaii-wify-linux -u F20230814
```

### Run with Environment Variables
```bash
export KAWAII_USER="F20230814"
export KAWAII_PASS="your_password"
./bin/debug/kawaii-wify-linux
```

### First-Time Interactive Setup
Running without arguments prompts for credentials and caches them in your system keyring:
```bash
./bin/debug/kawaii-wify-linux
```

## Running as a Service

Kawaii-Wify can be run in the background or supervised via a systemd user unit:

```ini
# ~/.config/systemd/user/kawaii-wify.service
[Unit]
Description=Kawaii-Wify FortiGate WLAN Session Manager
After=network.target

[Service]
ExecStart=%h/bin/kawaii-wify-linux -u F20230814
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

- CLI subcommands (`login`, `logout`, `status`, `daemon`, `config`).
- Network interface / SSID binding so probes only run when connected to targeted campus SSIDs.
- Notification daemon hooks / desktop alert upon session drops or cooldown triggers.
- System tray status indicator.

## License

MIT
