# Kawaii-Wify

A lightweight background session manager and automated login daemon for FortiOS (FortiGate) captive portals, specifically designed for campus WLAN environments like BITS Pilani.

---

## Overview

Campus Wi-Fi networks secured by FortiGate firewalls require users to authenticate via a captive portal web interface. In active campus environments, lease durations expire frequently, connections drop during access point roaming, and devices are disconnected silently.

**Kawaii-Wify** runs as a background daemon and CLI utility to automate this lifecycle end-to-end:
- **Probes WAN Connectivity**: Checks reachability using lightweight HTTP `204 No Content` probes.
- **Captive Portal Interception**: Detects captive portal redirection and extracts the dynamic challenge token (`fgtauth?<token>`) from the firewall HTML payload.
- **Atomic Authentication Sequence**: Primes the FortiOS session state, enforces HTTP/1.1 over TLS with pinned referers, and retrieves dynamic session keepalive tokens.
- **Active Keepalive Ping**: Periodically pings the firewall keepalive endpoint (`/keepalive?<token>`) to prevent idle lease timeouts.
- **Circuit Breaker & Cooldown**: Protects against firewall bans and account lockouts by halting login attempts after consecutive failures and self-healing.
- **Persistent Configuration**: Automatically manages user settings (such as default username and polling intervals) in a local config file.
- **Secure Credential Storage**: Leverages OS Keyrings (Linux Secret Service/D-Bus, macOS Keychain, Windows Credential Manager) and masked interactive prompts so passwords are never stored in plaintext.
- **Modular CLI Architecture**: Powered by Cobra with dedicated subcommands (`daemon`, `login`, `logout`, `config`).
- **Graceful Signal Handling**: Listens for termination signals (`SIGINT`, `SIGTERM`) to clean up engine routines smoothly.

---

## Architecture

The project is structured into modular packages separating authentication protocols, configuration persistence, credential resolution, state machine orchestration, and CLI entrypoints:

```
kawaii-wify/
├── cmd/
│   └── kawaii-wify/
│       ├── main.go            # Application entrypoint delegating to cli.Execute()
│       ├── main_script.go     # Procedural test script
│       └── wifi.md            # Reverse-engineering notes and protocol analysis
├── internal/
│   ├── auth/
│   │   ├── client.go          # Custom HTTP transport (HTTP/1.1 enforcement, SNI checks)
│   │   └── portal.go          # FortiOS primitives (Probe, Prime, Login, Keepalive, Logout)
│   ├── cli/
│   │   ├── root.go            # Cobra root command definition and execution entrypoint
│   │   ├── daemon.go          # 'daemon' command (background authentication and keepalive)
│   │   ├── login.go           # 'login' command (credentials enrollment into Keyring & config)
│   │   ├── logout.go          # 'logout' command (credential purge and active session reset)
│   │   └── config.go          # 'config' command ('get' and 'set' for local preferences)
│   ├── config/
│   │   └── config.go          # JSON configuration loader and persistent storage
│   ├── credentials/
│   │   ├── prompt.go          # TTY-aware masked interactive credential & password prompts
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
   - Primes the challenge session via `GET https://fw.bits-pilani.ac.in:8090/fgtauth?<token>` (transient errors retry immediately next tick without burning auth attempts).
   - Submits credentials via `POST https://fw.bits-pilani.ac.in:8090/` with form values and pinned `Referer`.
   - Extracts the dynamic session token from the `keepalive?<token>` response redirect.
3. **Online**: WAN route confirmed. On each tick interval, the engine sends a keepalive ping (`/keepalive?<sessionToken>`) to keep the firewall lease active.
4. **Cooldown (Circuit Breaker)**: Triggered if authentication fails 3 consecutive times (`MaxAuthFailures = 3`). Halts authentication attempts for 10 seconds (`CooldownDuration = 10s`) to prevent campus account lockouts and firewall IP bans. Automatically resets failure counters and self-heals back to `Offline`.

---

## Configuration

Kawaii-Wify supports persistent user configuration stored in your user configuration directory (per the XDG Base Directory specification on Linux):

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
| `username` | string | `""` | Active student ID / campus login ID. Automatically saved on login/override. |
| `check_interval` | string | `"10s"` | Frequency of probe and keepalive checks (parsed as a Go duration, e.g. `"5s"`, `"10s"`, `"1m"`). |

---

## Credential Resolution Hierarchy

Credentials are resolved using a cascading hierarchy:

1. **Explicit CLI Flag**: The `-u <username>` flag always takes highest precedence.
2. **Environment Variables**: `KAWAII_USER` and `KAWAII_PASS` (ideal for headless scripts or containers).
3. **Persistent Config File**: If no flag is given, the username is loaded from `~/.config/kawaii-wify/config.json`.
4. **System Keyring**: Using the resolved target username, Kawaii-Wify queries the OS Keyring (`gnome-keyring`, macOS Keychain, Windows Credential Manager).
5. **Interactive Terminal Prompt**: If credentials are missing, Kawaii-Wify prompts for missing information using masked TTY inputs. If the username is already known, it prompts only for the password (`PromptPassword`). Newly entered credentials are saved to the OS Keyring.

---

## CLI Usage & Commands

Kawaii-Wify provides commands to control the daemon, manage authentication, and configure settings:

### 1. Running the Daemon
Run the background authentication and keepalive service:
```bash
# Start daemon with saved defaults
kawaii-wify daemon

# Override active user for this session
kawaii-wify daemon -u F20230814

# Run without keepalive pings (only auto-relies on disconnect detection)
kawaii-wify daemon --no-keepalive
```

### 2. Login & Credential Enrollment
Store or update credentials in the OS keyring and set the active user:
```bash
# Interactive prompt for username and password
kawaii-wify login

# Specify user and prompt only for password
kawaii-wify login -u F20230814

# Specify user and password directly (non-interactive)
kawaii-wify login -u F20230814 -p "F20237057#"
```

### 3. Logout & Purge
Purge stored credentials from the keyring and clear the active user session:
```bash
# Logout the currently configured active user
kawaii-wify logout

# Logout a specific user from the keyring
kawaii-wify logout -u F20230814
```

### 4. Configuration Management (`get` & `set`)
Inspect and update settings without manually editing JSON files:
```bash
# View all configuration settings
kawaii-wify config get

# View a specific setting
kawaii-wify config get check_interval
kawaii-wify config get username
kawaii-wify config get keepalive

# Update polling interval (validated against Go durations)
kawaii-wify config set check_interval 15s

# Toggle keepalive pings permanently (true / false)
kawaii-wify config set keepalive false

# Update active default user
kawaii-wify config set username F20230814
```

---

## Security & Protocol Hardening

FortiGate's embedded port 8090 web server has specific network quirks:

- **HTTP/1.1 Enforcement**: Standard HTTP clients negotiate HTTP/2 ALPN by default over TLS. The FortiOS micro-server resets connections when HTTP/2 is offered (`curl: (52) Empty reply from server`). Kawaii-Wify suppresses ALPN negotiation by setting `TLSNextProto` to an empty map, forcing reliable HTTP/1.1 communication.
- **Scoped Certificate Verification**: Instead of disabling TLS checks globally with `InsecureSkipVerify: true`, Kawaii-Wify uses a custom `VerifyConnection` callback to verify that the remote certificate Common Name or DNS name matches `fw.bits-pilani.ac.in`.
- **Referer Pinning & User-Agent**: The FortiOS gateway requires requests to include a browser `User-Agent` and a matching `Referer: https://fw.bits-pilani.ac.in:8090/fgtauth?<token>` header.
- **Isolated Transport Failure Retries**: Gateway priming errors retry on the next tick without consuming authentication attempts, preventing false-positive account lockouts during Wi-Fi drops.

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
   ExecStart=%h/.local/bin/kawaii-wify daemon
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

- [x] Cobra CLI subcommands (`daemon`, `login`, `logout`, `config get`, `config set`).
- [x] Password-only prompt when target user is already known.
- [ ] Network interface & SSID binding (trigger probes only when connected to designated campus SSIDs).
- [ ] Desktop notifications via D-Bus (`notify-send` / `libnotify`) for connection drops and re-logins.
- [ ] System tray indicator with real-time status and quick actions.

---

## License

MIT License. See [LICENSE](LICENSE) for details.
