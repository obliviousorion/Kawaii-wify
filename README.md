# Kawaii-Wify

A lightweight background session manager and automated login daemon for FortiOS (FortiGate) captive portals, specifically designed for campus WLAN environments like BITS Pilani.

---

## Overview

Campus Wi-Fi networks secured by FortiGate firewalls require users to authenticate via a captive portal web interface. In active campus environments, lease durations expire frequently, connections drop during access point roaming, and devices are disconnected silently.

**Kawaii-Wify** runs as a background daemon to automate this lifecycle end-to-end:
- **Probes WAN Connectivity**: Checks reachability using lightweight HTTP `204 No Content` probes.
- **Captive Portal Interception**: Detects captive portal redirection and extracts the dynamic challenge token (`fgtauth?<token>`) from the firewall HTML payload.
- **Atomic Authentication Sequence**: Primes the FortiOS session state, enforces HTTP/1.1 over TLS with pinned referers, and retrieves dynamic session keepalive tokens.
- **Active Keepalive Ping**: Periodically pings the firewall keepalive endpoint (`/keepalive?<token>`) to prevent idle lease timeouts.
- **Circuit Breaker & Cooldown**: Protects against firewall bans and account lockouts by halting authentication attempts after consecutive failures and self-healing.
- **Persistent Configuration**: Automatically manages user settings (such as default username and polling intervals) in a local config file.
- **Secure Credential Storage**: Leverages OS Keyrings (Linux Secret Service/D-Bus, macOS Keychain, Windows Credential Manager) and masked interactive prompts so passwords are never stored in plaintext.
- **Graceful Signal Handling**: Listens for termination signals (`SIGINT`, `SIGTERM`) to clean up engine routines smoothly.

---

## Architecture

The project is structured into modular packages separating authentication protocols, configuration persistence, credential resolution, and state machine orchestration:

```
kawaii-wify/
├── cmd/
│   └── kawaii-wify/
│       ├── main.go            # Application entrypoint, CLI flags, signal handling
│       ├── main_script.go     # Procedural test script
│       └── wifi.md            # Reverse-engineering notes and protocol analysis
├── internal/
│   ├── auth/
│   │   ├── client.go          # Custom HTTP transport (HTTP/1.1 enforcement, SNI checks)
│   │   └── portal.go          # FortiOS primitives (Probe, Prime, Login, Keepalive, Logout)
│   ├── config/
│   │   └── config.go          # JSON configuration loader and persistent storage
│   ├── credentials/
│   │   ├── prompt.go          # TTY-aware masked interactive credential prompts
│   │   ├── resolver.go        # Multi-tiered cascading credential resolution
│   │   └── store.go           # OS keyring storage wrapper (Set, Get, Delete)
│   ├── engine/
│   │   ├── engine.go          # Thread-safe state engine, ticker loop, circuit breaker
│   │   └── state.go           # State constants and transitions
│   └── tray/
│       └── assets/            # System tray icon assets (under development)
├── Makefile                   # Build, release, and run targets
├── go.mod                     # Go module definitions
└── go.sum                     # Checksums for Go dependencies
```

---

## State Machine & Daemon Lifecycle

The core `Engine` operates as a thread-safe finite state machine:

```
                  ┌──────────────┐
                  │   Offline    │ ◄────────────────┐
                  └──────┬───────┘                  │
                         │ Intercepted              │ Cooldown
                         ▼                          │ Elapsed
                  ┌──────────────┐                  │
                  │   Captive    │                  │
                  └──────┬───────┘                  │
         Success │      │ Fail >= 3                │
                 ▼      ▼                           │
        ┌────────────┐ ┌──────────────┐             │
        │   Online   │ │   Cooldown   │ ────────────┘
        └─────┬──────┘ └──────────────┘
              │ (Keepalive Ticker)
              └─► Ping /keepalive
```

1. **Offline**: No route to the internet or gateway probe failed. Transitions to `Captive` if redirected.
2. **Captive Portal Detected**: Interception confirmed. The engine:
   - Captures the dynamic 16-character hexadecimal token (`magicToken`).
   - Primes the challenge session via `GET https://fw.bits-pilani.ac.in:8090/fgtauth?<token>`.
   - Submits credentials via `POST https://fw.bits-pilani.ac.in:8090/` with form values and pinned `Referer`.
   - Extracts the dynamic session token from the `keepalive?<token>` response redirect.
3. **Online**: WAN route confirmed. On each tick interval, the engine sends a keepalive ping (`/keepalive?<sessionToken>`) to keep the firewall lease active. If keepalive fails, it clears the token and triggers re-authentication.
4. **Cooldown (Circuit Breaker)**: Triggered if login fails 3 consecutive times (`MaxAuthFailures = 3`). Halts authentication attempts for 10 seconds (`CooldownDuration = 10s`) to prevent account lockouts and firewall IP bans. Automatically resets failure counters and self-heals back to `Offline`.

---

## Configuration

Kawaii-Wify supports a persistent user configuration stored in your user configuration directory (per the XDG Base Directory specification on Linux):

- **Default Location**: `~/.config/kawaii-wify/config.json`
- **File Permissions**: Restricted to `0600` (read/write only by user)

### Configuration Schema

```json
{
  "username": "F20230814",
  "check_interval": "10s"
}
```

| Field | Type | Default | Description |
|---|---|---|---|
| `username` | string | `""` | Saved student ID / campus login ID. Automatically saved on first login. |
| `check_interval` | string | `"10s"` | Frequency of probe and keepalive checks (parsed as a Go duration, e.g. `"5s"`, `"10s"`, `"1m"`). |

---

## Credential Resolution Hierarchy

Credentials are resolved using a cascading hierarchy so sensitive secrets are never logged, passed via command-line process arguments, or stored in plaintext:

1. **Environment Variables**: If both `KAWAII_USER` and `KAWAII_PASS` are defined, they take immediate precedence (ideal for CI, headless scripts, or containers).
2. **CLI Flag Override**: The `-u <username>` flag overrides the stored config file username.
3. **Persistent Config File**: If no `-u` flag is given, the username is loaded from `~/.config/kawaii-wify/config.json`.
4. **System Keyring**: Using the target username, Kawaii-Wify queries the OS Keyring (via D-Bus Secret Service on Linux, macOS Keychain, or Windows Credential Manager).
5. **Interactive Terminal Prompt**: If credentials are missing or not found in the keyring, Kawaii-Wify prompts for the student ID and securely masks password entry via `golang.org/x/term`. Once verified, credentials are automatically saved to the OS Keyring and the username is persisted to `config.json` for subsequent unattended runs.

---

## Security & Protocol Hardening

FortiGate's embedded port 8090 web server has specific network quirks:

- **HTTP/1.1 Enforcement**: Standard HTTP clients negotiate HTTP/2 ALPN by default over TLS. The FortiOS micro-server resets connections when HTTP/2 is offered (`curl: (52) Empty reply from server`). Kawaii-Wify suppresses ALPN negotiation by setting `TLSNextProto` to an empty map, forcing reliable HTTP/1.1 communication.
- **Scoped Certificate Verification**: Instead of disabling TLS checks globally with `InsecureSkipVerify: true`, Kawaii-Wify uses a custom `VerifyConnection` callback to verify that the remote certificate Common Name or DNS name matches `fw.bits-pilani.ac.in`.
- **Referer Pinning & User-Agent**: The FortiOS gateway requires requests to include a browser `User-Agent` and a matching `Referer: https://fw.bits-pilani.ac.in:8090/fgtauth?<token>` header.
- **Atomic Handshake**: The engine synchronizes token extraction, session priming, and authentication within a single transaction to prevent race conditions from background OS captive portal assistants.

---

## Requirements

- **Go**: 1.22 or newer
- **Linux**: Supported desktop environment or daemon running D-Bus with a Secret Service provider (`gnome-keyring`, `ksecretsservice`, or `keepassxc`)
- **macOS / Windows**: Compatible via Keychain / Windows Credential Manager

---

## Building

A `Makefile` is provided for common development and production tasks:

```bash
# Compile debug build (bin/debug/kawaii-wify-linux)
make build-linux

# Compile stripped release build (bin/release/kawaii-wify-linux_0.1.0)
make build-linux-release

# Build and execute debug binary
make run

# Clean build artifacts
make clean
```

---

## Usage

### 1. Unattended Run (Configured)
Once your credentials and configuration are saved, run without arguments:
```bash
./bin/debug/kawaii-wify-linux
```

### 2. Override Username
Override the configured username using the `-u` flag:
```bash
./bin/debug/kawaii-wify-linux -u F20230814
```

### 3. Headless / Environment Variables
Run in environments without keyring access:
```bash
export KAWAII_USER="F20230814"
export KAWAII_PASS="your_password"
./bin/debug/kawaii-wify-linux
```

### 4. First-Time Setup
If no credentials exist, simply run the binary. It will prompt for your ID and password, securely store the secret in your keyring, save your ID to `~/.config/kawaii-wify/config.json`, and start the daemon:
```bash
./bin/debug/kawaii-wify-linux
Enter Username: F20230814
Enter Password: 
[INFO] Saved default username F20230814 to config
[INFO] Starting kawaii-wify for user: F20230814 (interval: 10s)
```

---

## Running as a Systemd User Service

You can run Kawaii-Wify as a background service managed by `systemd`:

1. Copy the compiled binary to your local user binary directory:
   ```bash
   mkdir -p ~/.local/bin
   cp bin/debug/kawaii-wify-linux ~/.local/bin/kawaii-wify
   ```

2. Create the user service unit file:
   ```ini
   # ~/.config/systemd/user/kawaii-wify.service
   [Unit]
   Description=Kawaii-Wify FortiGate WLAN Session Manager
   After=network.target

   [Service]
   ExecStart=%h/.local/bin/kawaii-wify
   Restart=always
   RestartSec=5

   [Install]
   WantedBy=default.target
   ```

3. Enable and start the service:
   ```bash
   systemctl --user daemon-reload
   systemctl --user enable --now kawaii-wify.service
   ```

4. Inspect live logs:
   ```bash
   journalctl --user -u kawaii-wify.service -f
   ```

---

## Roadmap

- [ ] CLI subcommands (`status`, `logout`, `config set`, `credentials clear`).
- [ ] Network interface & SSID binding (trigger probes only when connected to designated campus SSIDs).
- [ ] Desktop notifications via D-Bus (`notify-send` / `libnotify`) for connection drops and re-logins.
- [ ] System tray indicator with real-time status and quick actions.

---

## License

MIT License. See [LICENSE](LICENSE) for details.
