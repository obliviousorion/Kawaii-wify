# Kawaii-Wify

A lightweight background session manager and automated captive portal login utility for FortiOS (FortiGate) networks.

## Downloads

Pre-built binaries and native packages are available from the [GitHub Releases](https://github.com/obliviousorion/Kawaii-wify/releases/latest) page:

| Platform | Package | Architecture | Download |
| :--- | :--- | :--- | :--- |
| Android | Native App (.apk) | Android 8.0+ (ARM64, x86_64) | [kawaii-wify-android.apk](https://github.com/obliviousorion/Kawaii-wify/releases/latest/download/kawaii-wify-android.apk) |
| Linux | Standalone Binary | x86_64 / amd64 | [kawaii-wify-linux-amd64](https://github.com/obliviousorion/Kawaii-wify/releases/latest/download/kawaii-wify-linux-amd64) |
| Linux | Standalone Binary | ARM64 / aarch64 | [kawaii-wify-linux-arm64](https://github.com/obliviousorion/Kawaii-wify/releases/latest/download/kawaii-wify-linux-arm64) |
| Windows | Executable (.exe) | x86_64 / amd64 | [kawaii-wify-windows-amd64.exe](https://github.com/obliviousorion/Kawaii-wify/releases/latest/download/kawaii-wify-windows-amd64.exe) |
| macOS | Apple Silicon Binary | M1 / M2 / M3 / M4 (arm64) | [kawaii-wify-darwin-arm64](https://github.com/obliviousorion/Kawaii-wify/releases/latest/download/kawaii-wify-darwin-arm64) |
| macOS | Intel Binary | x86_64 (amd64) | [kawaii-wify-darwin-amd64](https://github.com/obliviousorion/Kawaii-wify/releases/latest/download/kawaii-wify-darwin-amd64) |

---

## Installation

### Automated Install (Recommended)

**Linux and macOS:**
Run the installation script in your terminal to automatically detect your CPU architecture, download the latest release binary, and install it to `/usr/local/bin`:
```bash
curl -fsSL https://raw.githubusercontent.com/obliviousorion/Kawaii-wify/main/install.sh | bash
```

**Windows:**
Run the following command in PowerShell to download `kawaii-wify.exe` to `%LOCALAPPDATA%\kawaii-wify` and configure your user `PATH`:
```powershell
irm https://raw.githubusercontent.com/obliviousorion/Kawaii-wify/main/install.ps1 | iex
```

### Manual Installation

- **Android**: Download `kawaii-wify-android.apk`, tap the downloaded file to install, and allow installation from unknown sources if prompted.
- **Linux and macOS**:
  1. Download the binary matching your CPU architecture from the Downloads table.
  2. Mark the binary as executable:
     ```bash
     chmod +x kawaii-wify-*
     ```
  3. Move it to a directory in your PATH (e.g., `/usr/local/bin` or `~/.local/bin`):
     ```bash
     sudo mv kawaii-wify-* /usr/local/bin/kawaii-wify
     ```
- **Windows**:
  1. Download `kawaii-wify-windows-amd64.exe`.
  2. Place it in a folder of your choice (for example, `%LOCALAPPDATA%\kawaii-wify\kawaii-wify.exe`).
  3. Add that directory to your User `PATH` environment variable so you can invoke `kawaii-wify` from Command Prompt or PowerShell.

---

## Overview and How It Works

Campus and enterprise Wi-Fi networks protected by FortiGate firewalls require users to authenticate through a captive portal web page. Sessions regularly expire due to lease timeouts, access point roaming, or idle disconnects.

Kawaii-Wify automates the entire session lifecycle:
1. **Connectivity Probes**: Periodically sends lightweight HTTP probes to verify WAN reachability.
2. **Captive Interception**: When a portal redirect is detected, extracts the authentication challenge token.
3. **Automated Authentication**: Primes the FortiOS session and securely submits login credentials.
4. **Session Keepalive**: Sends periodic keepalive pings to prevent idle lease timeouts without user intervention.
5. **Circuit Breaker**: Halts re-authentication attempts after consecutive failures to protect accounts from lockouts.
6. **Hardware-Backed Credential Security**: Stores credentials in system keyrings (Android KeyStore, Linux Secret Service, macOS Keychain, Windows Credential Manager).

---

## Android App Usage

1. **Install**: Download and install `kawaii-wify-android.apk` on your device (Android 8.0 or newer).
2. **Setup Credentials**: Enter your campus username and password on first launch. Credentials are encrypted via Android KeyStore (AES-256 GCM).
3. **Background Daemon**: Tap **Connect** to activate background monitoring. A persistent foreground service handles automated logins and keepalives.
4. **Pause vs. Disconnect**:
   - **Pause**: Halts background network checks and keepalives without terminating your active firewall lease.
   - **Disconnect**: Revokes your active session token on the firewall and stops background checks.
5. **Quick Settings Tile**: Add the Kawaii-Wify tile to your Android notification shade for one-tap toggling.

---

## CLI Usage (Linux, macOS, Windows)

### Quick Start

```bash
# 1. Save your credentials to the OS keyring
kawaii-wify login -u YOUR_USERNAME

# 2. Start the daemon in the background
kawaii-wify start

# 3. Check status and connection telemetry
kawaii-wify status

# 4. View daemon logs
kawaii-wify logs

# 5. (Optional) Enable automatic startup on system boot/login
kawaii-wify autostart enable

# 6. Stop the background daemon
kawaii-wify stop
```

### Command Reference

| Command | Description |
| :--- | :--- |
| `kawaii-wify login` | Enrolls credentials into the OS keyring interactively. |
| `kawaii-wify start` | Starts the daemon in the background (survives terminal exit). |
| `kawaii-wify daemon` | Runs the daemon attached in the foreground for real-time console logs. |
| `kawaii-wify status` | Queries the running daemon for telemetry (state, user, uptime, lease). |
| `kawaii-wify connect` | Triggers an immediate network probe and login over IPC. |
| `kawaii-wify disconnect` | Disconnects the active session and pauses background probing. |
| `kawaii-wify stop` | Gracefully shuts down the background daemon. |
| `kawaii-wify logs` | Displays the most recent daemon output lines (`-n <count>`, `-p` for path). |
| `kawaii-wify logout` | Purges stored credentials from the keyring. |
| `kawaii-wify autostart enable` | Enables background daemon autostart on system boot/login. |
| `kawaii-wify autostart disable` | Disables background daemon autostart on system boot/login. |
| `kawaii-wify autostart status` | Checks whether background autostart is currently enabled. |
| `kawaii-wify config get` | Displays current configuration parameters. |
| `kawaii-wify config set <key> <val>` | Updates configuration (`gateway`, `check_interval`, `keepalive`). |

---

## Configuration

Configuration is stored in user space:
- **Linux**: `~/.config/kawaii-wify/config.json`
- **macOS**: `~/Library/Application Support/kawaii-wify/config.json`
- **Windows**: `%APPDATA%\kawaii-wify\config.json`

| Setting | Default | Description |
| :--- | :--- | :--- |
| `gateway` | `fw.bits-pilani.ac.in:8090` | FortiOS captive portal host and port. |
| `check_interval` | `10s` | Polling and keepalive interval (e.g., `5s`, `15s`, `1m`). |
| `keepalive` | `true` | When true, pings the firewall keepalive endpoint periodically. |
| `auto_connect` | `true` | When true, daemon starts monitoring immediately on launch. |

---

## Building from Source

### Prerequisites
- Go 1.22 or newer (for CLI)
- JDK 21 and Android SDK (for Android app)

```bash
# Build binary for current operating system
make

# Build release packages for all platforms (Linux, Windows, macOS, Android)
make release-all

# Build Android release APK only
make android-release

# Install CLI binary to ~/go/bin
make install
```

---

## License

MIT License. See [LICENSE](LICENSE) for details.
