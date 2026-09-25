# Kawaii-Wify

A lightweight background session manager and automated login daemon for FortiOS (FortiGate) captive portals, specifically designed for campus WLAN environments like BITS Pilani.

---

## 📥 Downloads (Latest Release)

Pre-built binaries and native packages are available directly from the [GitHub Releases](https://github.com/obliviousorion/Kawaii-wify/releases/latest) page:

| Platform | Package | Architecture | Direct Download |
| :--- | :--- | :--- | :--- |
| 📱 **Android** | Native App (`.apk`) | Android 8.0+ (ARM64 / x86_64) | [**kawaii-wify-android.apk**](https://github.com/obliviousorion/Kawaii-wify/releases/latest/download/kawaii-wify-android.apk) |
| 🐧 **Linux** | Standalone Binary | x86_64 / amd64 | [**kawaii-wify-linux-amd64**](https://github.com/obliviousorion/Kawaii-wify/releases/latest/download/kawaii-wify-linux-amd64) |
| 🐧 **Linux** | Standalone Binary | ARM64 / aarch64 | [**kawaii-wify-linux-arm64**](https://github.com/obliviousorion/Kawaii-wify/releases/latest/download/kawaii-wify-linux-arm64) |
| 🪟 **Windows** | Executable (`.exe`) | x86_64 / amd64 | [**kawaii-wify-windows-amd64.exe**](https://github.com/obliviousorion/Kawaii-wify/releases/latest/download/kawaii-wify-windows-amd64.exe) |
| 🍏 **macOS** | Universal / Apple Silicon | M1 / M2 / M3 / M4 (arm64) | [**kawaii-wify-darwin-arm64**](https://github.com/obliviousorion/Kawaii-wify/releases/latest/download/kawaii-wify-darwin-arm64) |
| 🍏 **macOS** | Intel Binary | x86_64 (amd64) | [**kawaii-wify-darwin-amd64**](https://github.com/obliviousorion/Kawaii-wify/releases/latest/download/kawaii-wify-darwin-amd64) |

---

## Overview

Campus Wi-Fi networks secured by FortiGate firewalls require users to authenticate via a captive portal web interface. In active campus environments, lease durations expire frequently, connections drop during access point roaming, and devices are disconnected silently.

**Kawaii-Wify** runs as a background daemon and CLI utility to automate this lifecycle end-to-end:
- **Probes WAN Connectivity**: Checks reachability using lightweight HTTP `204 No Content` probes.
- **Captive Portal Interception**: Detects captive portal redirection and extracts the dynamic challenge token (`fgtauth?<token>`) from the firewall HTML payload.
- **Atomic Authentication Sequence**: Primes the FortiOS session state, enforces HTTP/1.1 over TLS with pinned referers, and retrieves dynamic session keepalive tokens.
- **Configurable Keepalive Ping**: Periodically pings the firewall keepalive endpoint (`/keepalive?<token>`) to prevent idle lease timeouts, or can be toggled off to rely purely on automatic re-login on disconnect.
- **Circuit Breaker & Cooldown**: Protects against firewall bans and account lockouts by halting login attempts after consecutive failures and self-healing.
- **Persistent Configuration**: Automatically manages user settings (such as default username and polling intervals) in a local config file.
- **Secure Credential Storage**: Leverages OS Keyrings (Linux Secret Service/D-Bus, macOS Keychain, Windows Credential Manager) and masked interactive prompts so passwords are never stored in plaintext.
- **Modular CLI Architecture**: Powered by Cobra with dedicated subcommands (`daemon`, `login`, `logout`, `status`, `config`).
- **Cross-Platform IPC Layer**: Exposes live daemon telemetry and session management over Unix domain sockets (Linux/macOS) and Named Pipes (Windows).
- **Graceful Signal Handling**: Listens for termination signals (`SIGINT`, `SIGTERM`) to clean up engine routines and IPC sockets smoothly.

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
│   │   └── gateway.go         # FortiOS Gateway client (Probe, Prime, Login, Keepalive, Logout)
│   ├── cli/
│   │   ├── root.go            # Cobra root command definition and execution entrypoint
│   │   ├── start.go           # 'start' command (spawns daemon detached in background)
│   │   ├── daemon.go          # 'daemon' command (foreground authentication and keepalive)
│   │   ├── connect.go         # 'connect' command (triggers immediate probe/login via IPC)
│   │   ├── disconnect.go      # 'disconnect' command (pauses daemon & clears session via IPC)
│   │   ├── stop.go            # 'stop' command (gracefully shuts down daemon via IPC)
│   │   ├── logs.go            # 'logs' command (inspects daemon output & log storage path)
│   │   ├── login.go           # 'login' command (credentials enrollment into Keyring & config)
│   │   ├── logout.go          # 'logout' command (credential purge and active session reset)
│   │   ├── status.go          # 'status' command (queries running daemon over IPC)
│   │   └── config.go          # 'config' command ('get' and 'set' for local preferences)
│   ├── config/
│   │   ├── config.go          # JSON configuration loader and persistent storage
│   │   └── paths.go           # Platform standard configuration and log file paths
│   ├── credentials/
│   │   ├── prompt.go          # TTY-aware masked interactive credential & password prompts
│   │   ├── resolver.go        # Multi-tiered cascading credential resolution
│   │   └── store.go           # OS keyring storage wrapper (Set, Get, Delete)
│   ├── engine/
│   │   ├── engine.go          # Thread-safe state engine, ticker loop, circuit breaker
│   │   └── state.go           # State constants and transitions
│   ├── ipc/
│   │   ├── client.go          # RPC client wrapper for CLI communication
│   │   ├── server.go          # Net/RPC daemon service and concurrent server loop
│   │   ├── transport_unix.go  # Unix domain socket transport (Linux/macOS)
│   │   ├── transport_windows.go # Named pipe transport (Windows)
│   │   └── types.go           # Telemetry, status, and control request/response types
│   ├── logger/
│   │   └── logger.go          # Dual-write session logger with atomic sync & boot rotation
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

Kawaii-Wify supports persistent user configuration stored in your standard user configuration directory:

- **Linux**: `~/.config/kawaii-wify/config.json`
- **Windows**: `%APPDATA%\kawaii-wify\config.json`
- **macOS**: `~/Library/Application Support/kawaii-wify/config.json`
- **File Permissions**: Restricted to `0600` (read/write only by user)

### Configuration Schema

```json
{
  "username": "F20230814",
  "gateway": "fw.bits-pilani.ac.in:8090",
  "check_interval": "10s",
  "keepalive": true,
  "auto_connect": true
}
```

| Field | Type | Default | Description |
|---|---|---|---|
| `username` | string | `""` | Active student ID / campus login ID. Automatically saved on login/override. |
| `gateway` | string | `"fw.bits-pilani.ac.in:8090"` | FortiOS captive portal endpoint (automatically appends port `:8090` if omitted). |
| `check_interval` | string | `"10s"` | Frequency of probe and keepalive checks (parsed as a Go duration, e.g. `"5s"`, `"10s"`, `"1m"`). |
| `keepalive` | bool | `true` | When enabled, sends periodic keepalive pings while online. When disabled, relies purely on automatic re-login upon connection drops. |
| `auto_connect` | bool | `true` | When enabled, daemon automatically connects on startup. When false, daemon starts in paused state. |

### Persistent Logging & Rotation

Daemon session logs are automatically dual-written to standard output and a persistent file with atomic flush syncing:
- **Windows**: `%LOCALAPPDATA%\kawaii-wify\logs\daemon.log`
- **Linux/macOS**: `$XDG_STATE_HOME/kawaii-wify/logs/daemon.log` (or `~/.local/state/kawaii-wify/logs/daemon.log`)
- **Boot Rotation**: Upon launching a new daemon session, the previous log is atomically rotated to `daemon.prev.log`, preventing unbounded disk growth while preserving crash history.

---

## Credential Resolution Hierarchy

Credentials are resolved using a cascading hierarchy:

1. **Explicit CLI Flags**: `-u <username>` and `-p <password>` take highest precedence during login.
2. **Environment Variables**: `KAWAII_USER` and `KAWAII_PASS` (ideal for headless scripts or containers).
3. **Persistent Config File**: If no flag is given, the active username is loaded from `config.json`.
4. **System Keyring**: Using the resolved target username, Kawaii-Wify queries the OS Keyring (Linux Secret Service/D-Bus, macOS Keychain, Windows Credential Manager).
5. **Interactive Terminal Prompt**: If credentials are missing, Kawaii-Wify prompts for missing information using masked TTY inputs. If the username is already known, it prompts only for the password (`PromptPassword`). Newly entered credentials are automatically saved to the OS Keyring.

---

## CLI Usage & Commands

Kawaii-Wify provides commands to control the daemon, manage authentication, and configure settings:

### 1. Starting the Daemon (`start` & `daemon`)
Launch the background authentication and keepalive service:
```bash
# Start daemon detached in the background (survives closing the terminal)
kawaii-wify start

# Start in background with user or gateway override
kawaii-wify start -u F20230814 --gateway 172.16.100.1:8090

# Start in background in paused state
kawaii-wify start -p

# Run attached in the foreground (useful for watching live logs in terminal)
kawaii-wify daemon
```

### 2. Manual Connect, Disconnect & Stop (IPC Triggers)
Control a running background daemon on-demand over IPC:
```bash
# Trigger an immediate probe & login attempt (unpauses background daemon)
kawaii-wify connect

# Clear active session & pause automatic reconnect monitoring
kawaii-wify disconnect

# Gracefully terminate the background daemon process
kawaii-wify stop
```

### 3. Status & Telemetry Query
Query the running background daemon over IPC for real-time status and operational metrics:
```bash
# Formatted human-readable status
kawaii-wify status

# Structured JSON output (for scripts, GUI integrations, or piping into jq)
kawaii-wify status --json
```
Example output:
```text
kawaii-wify Daemon Status
─────────────────────────
  State:         Online
  User:          F20230814
  Paused:        false
  Uptime:        42m15s
  Last Probe:    15:04:05
  Session:       0a1b2c3d4e5f6a7b
```

### 4. Daemon Logs Inspection
View recent output from the background daemon without needing to manually hunt down file paths:
```bash
# View last 30 log lines (default)
kawaii-wify logs

# View a specific number of recent lines
kawaii-wify logs -n 50

# Print the resolved absolute path to the active log file
kawaii-wify logs -p
```

### 5. Login & Credential Enrollment
Store or update credentials in the OS keyring and set the active user:
```bash
# Interactive prompt for username and password
kawaii-wify login

# Specify user and prompt only for password
kawaii-wify login -u F20230814

# Specify user and password directly (non-interactive)
kawaii-wify login -u F20230814 -p "F20237057#"
```

### 6. Logout & Purge
Purge stored credentials from the keyring and clear the active user session:
```bash
# Logout the currently configured active user
kawaii-wify logout

# Logout a specific user from the keyring
kawaii-wify logout -u F20230814
```

### 7. Configuration Management (`get` & `set`)
Inspect and update settings without manually editing JSON files:
```bash
# View all configuration settings
kawaii-wify config get

# View a specific setting
kawaii-wify config get gateway
kawaii-wify config get check_interval
kawaii-wify config get username
kawaii-wify config get keepalive

# Update gateway endpoint (or use 'default' to reset)
kawaii-wify config set gateway fw.goa.bits-pilani.ac.in
kawaii-wify config set gateway default

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
- **Scoped Certificate Verification**: Instead of disabling TLS checks globally with `InsecureSkipVerify: true`, Kawaii-Wify uses a custom `VerifyConnection` callback to dynamically verify that the remote certificate Common Name or DNS name matches the configured gateway target host (defaulting to `fw.bits-pilani.ac.in`).
- **Referer Pinning & User-Agent**: The FortiOS gateway requires requests to include a browser `User-Agent` and a matching `Referer: https://<gateway>/fgtauth?<token>` header.
- **Isolated Transport Failure Retries**: Gateway priming errors retry on the next tick without consuming authentication attempts, preventing false-positive account lockouts during Wi-Fi drops.
- **Cross-Platform IPC Layer**: CLI-to-daemon communication uses standard Go `net/rpc` over native local IPC transports:
  - **Windows**: Named Pipes (`\\.\pipe\kawaii-wify`) via `github.com/Microsoft/go-winio`.
  - **Linux/macOS**: Unix Domain Sockets (`$XDG_RUNTIME_DIR/kawaii-wify.sock` or `$TMPDIR/kawaii-wify.sock`) enforced with user-only file permissions (`0600`).

---

## Building

A cross-platform `Makefile` is provided for common development and production tasks:

```bash
# Automatically build for your current host OS (Windows, Linux, or macOS)
make

# Compile and install globally to system Go binary path (~/go/bin)
make install
# (Ensure ~/go/bin is in your PATH: e.g. `fish_add_path ~/go/bin` for Fish, or `export PATH="$HOME/go/bin:$PATH"` for Bash/Zsh)

# Automatically build and run for current host OS with optional flags
make run ARGS="daemon"
make run ARGS='login -u F20230814 -p "MyPassword#"'

# Compile release binaries for all platforms (Linux, Windows, macOS)
make build-all

# Platform-specific builds (debug / stripped release)
make build-windows          # bin/debug/kawaii-wify-windows.exe
make build-windows-release  # bin/release/kawaii-wify-windows_0.1.0.exe
make build-linux            # bin/debug/kawaii-wify-linux
make build-linux-release    # bin/release/kawaii-wify-linux_0.1.0
make build-darwin           # bin/debug/kawaii-wify-darwin
make build-darwin-release   # bin/release/kawaii-wify-darwin_0.1.0

# Cross-platform cleanup of build artifacts
make clean
```

---

## Running in the Background

### Linux: Systemd User Service

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

### Windows: Task Scheduler

Run Kawaii-Wify silently in the background whenever you log into Windows:

1. Install the binary globally:
   ```powershell
   go install ./cmd/kawaii-wify
   ```

2. Register the scheduled task (PowerShell):
   ```powershell
   $action = New-ScheduledTaskAction -Execute "$env:USERPROFILE\go\bin\kawaii-wify.exe" -Argument "daemon"
   $trigger = New-ScheduledTaskTrigger -AtLogOn
   Register-ScheduledTask -TaskName "KawaiiWify" -Action $action -Trigger $trigger -Description "FortiGate WLAN Session Manager"
   ```

3. Start or stop the service at any time:
   ```powershell
   Start-ScheduledTask -TaskName "KawaiiWify"
   Stop-ScheduledTask -TaskName "KawaiiWify"
   ```

---

## Roadmap

- [x] Cobra CLI subcommands (`daemon`, `login`, `logout`, `status`, `config get`, `config set`).
- [x] Password-only prompt when target user is already known.
- [x] Cross-platform IPC Server & RPC control layer (Unix Domain Sockets & Windows Named Pipes).
- [x] `status` CLI command to inspect live daemon telemetry.
- [ ] Network interface & SSID binding (trigger probes only when connected to designated campus SSIDs).
- [ ] Desktop notifications via D-Bus (`notify-send` / `libnotify`) for connection drops and re-logins.
- [ ] System tray indicator with real-time status and quick actions.

---

## License

MIT License. See [LICENSE](LICENSE) for details.
