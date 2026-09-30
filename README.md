# Kawaii-Wify

A lightweight background session manager and automated captive portal login utility for FortiOS (FortiGate) networks.

---

## Downloads

Pre-built binaries and native packages are available from the [GitHub Releases](https://github.com/obliviousorion/Kawaii-wify/releases/latest) page:

| Platform | Package | Architecture | Download |
| :--- | :--- | :--- | :--- |
| **Android** | Native App (`.apk`) | Android 8.0+ (ARM64, x86_64) | [kawaii-wify-android-v0.2.2.apk](https://github.com/obliviousorion/Kawaii-wify/releases/latest/download/kawaii-wify-android-v0.2.2.apk) / [kawaii-wify-android.apk](https://github.com/obliviousorion/Kawaii-wify/releases/latest/download/kawaii-wify-android.apk) |
| **Windows** | Executable (`.exe`) | x86_64 / amd64 | [kawaii-wify-windows-amd64.exe](https://github.com/obliviousorion/Kawaii-wify/releases/latest/download/kawaii-wify-windows-amd64.exe) |
| **Linux** | Standalone Binary | x86_64 / amd64 | [kawaii-wify-linux-amd64](https://github.com/obliviousorion/Kawaii-wify/releases/latest/download/kawaii-wify-linux-amd64) |
| **Linux** | Standalone Binary | ARM64 / aarch64 | [kawaii-wify-linux-arm64](https://github.com/obliviousorion/Kawaii-wify/releases/latest/download/kawaii-wify-linux-arm64) |
| **macOS** | Apple Silicon Binary | M1 / M2 / M3 / M4 (arm64) | [kawaii-wify-darwin-arm64](https://github.com/obliviousorion/Kawaii-wify/releases/latest/download/kawaii-wify-darwin-arm64) |
| **macOS** | Intel Binary | x86_64 (amd64) | [kawaii-wify-darwin-amd64](https://github.com/obliviousorion/Kawaii-wify/releases/latest/download/kawaii-wify-darwin-amd64) |

---

## Installation

### Automated Install (Recommended for Desktop)

**Linux and macOS:**
```bash
curl -fsSL https://raw.githubusercontent.com/obliviousorion/Kawaii-wify/main/install.sh | bash
```

**Windows (PowerShell):**
```powershell
irm https://raw.githubusercontent.com/obliviousorion/Kawaii-wify/main/install.ps1 | iex
```

### Manual Installation

- **Android**: Download `kawaii-wify-android-v0.2.2.apk`, open the file to install, and allow installation from unknown sources if prompted.
- **Linux / macOS**:
  1. Download the binary matching your CPU architecture.
  2. Make it executable: `chmod +x kawaii-wify-*`
  3. Move to your PATH: `sudo mv kawaii-wify-* /usr/local/bin/kawaii-wify`
- **Windows**:
  1. Download `kawaii-wify-windows-amd64.exe`.
  2. Move it to `%LOCALAPPDATA%\kawaii-wify\kawaii-wify.exe`.
  3. Add `%LOCALAPPDATA%\kawaii-wify` to your User `PATH` environment variable.

### Updating

#### Native Desktop Self-Updater (Recommended)

Starting with **v0.2.2**, Kawaii-Wify includes a built-in atomic self-updater:

```bash
# Check if an update is available without downloading
kawaii-wify update --check

# Perform in-place update (auto-detects platform, stops daemon, swaps binary safely, restarts daemon)
kawaii-wify update

# Force re-download / reinstall even if already on latest version
kawaii-wify update --force
```

#### Via Automated Install Script

You can also re-run the automated install script anytime to update:

**Linux / macOS:**
```bash
curl -fsSL https://raw.githubusercontent.com/obliviousorion/Kawaii-wify/main/install.sh | bash
```

**Windows (PowerShell):**
```powershell
irm https://raw.githubusercontent.com/obliviousorion/Kawaii-wify/main/install.ps1 | iex
```

> [!TIP]
> **Smart Updates with Zero Downtime:**
> - The native `update` command and install scripts both inspect the cross-platform `versions.json` manifest. You will never receive false update prompts if a release only bumped the other platform.
> - If an active `kawaii-wify` daemon is running, it safely pauses over IPC, replaces the binary (preventing Windows file-locking errors via atomic renaming), and restarts automatically.
> - Your saved credentials, configuration, and autostart settings are 100% preserved.
> - **Android Users**: When an update is ready, a pulsing download icon appears on the Dashboard top bar and an OS notification is sent. Tapping it opens the Cyber-Glass Update Dialog to download the APK directly. Official releases use a consistent signing key, allowing seamless in-place updates.

---

## How It Works

Campus and corporate networks with FortiGate firewalls enforce session expirations, access point handoffs, and idle disconnects that kick you back to a web captive portal.

**Kawaii-Wify automates the entire session lifecycle:**
1. **Connectivity Probes**: Lightweight HTTP probes detect captive portal redirects without consuming significant bandwidth.
2. **Instant Authentication**: Securely acquires challenge tokens and handles captive login handshakes automatically.
3. **Session Re-establishment**: Responds instantly when access points roam or sessions drop.
4. **Hardware-Backed Credential Security**: Credentials are encrypted in native hardware keystores (Android KeyStore, Windows Credential Manager, macOS Keychain, Linux Secret Service).
5. **Circuit Breaker**: Stops retrying after consecutive authentication failures to protect your account against lockouts.
6. **TLS Certificate Pinning (TOFU)**: Authenticates self-signed gateway certificates via cryptographic SHA-256 fingerprints, protecting against rogue APs and MITM attacks.
7. **Host-Restricted Redirects**: Restricts HTTP redirects strictly to the configured gateway host, ensuring credentials and session tokens never leak to external domains.

---

## Android App Setup & Usage

The Android app provides a simple UI backed by an efficient foreground background service.

### Recommended Android Setup

1. **Install & Grant Permissions**:
   - Install `kawaii-wify-android.apk`.
   - **Notifications**: Allow notification access so the foreground service can display real-time connection status.
   - **Location (Optional)**: Android's privacy framework requires location permission solely to read and display the connected Wi-Fi SSID (network name).
     > [!NOTE]
     > The app does **not** track, record, or transmit your location. Captive portal detection and background authentication work completely fine even if you choose **Approximate Location** or decline this permission.

2. **Save Credentials**:
   - Tap the **Config / Settings** tab.
   - Enter your campus username and password, then tap **Save Credentials**.
   - Credentials are encrypted on-device via the Android KeyStore (AES-256 GCM).

3. **Configure Recommended Engine Settings**:
   - In the Config screen, verify the engine parameters:
     - **Keepalive**: Set to **OFF** *(avoids unnecessary background pings; the engine re-authenticates seamlessly when needed)*.
     - **Auto-Connect**: Set to **ON** *(starts monitoring automatically upon app launch)*.
     - **Check Interval**: Leave at default (`5s` or `10s`).
     - **Gateway**: Leave as default (`fw.bits-pilani.ac.in:8090`) unless your campus uses a different portal host.

4. **Enable Unrestricted Background Execution (Critical)**:
   - Modern Android versions aggressively put background apps to sleep. To keep Wi-Fi connected while the screen is locked:
     - Go to your phone's **Settings → Apps → Kawaii-Wify → Battery**.
     - Set battery usage to **Unrestricted** (or disable "Battery Optimization").

5. **Start Connection**:
   - Return to the **Dashboard** and tap **Connect**.
   - *(Optional)* Add the **Kawaii-Wify Quick Settings Tile** to your Android notification tray for instant one-tap control.

> [!TIP]
> **Pause vs. Disconnect:**
> - **Pause**: Suspends background probes and logins while keeping your current firewall session active.
> - **Disconnect**: Revokes your session directly from the firewall gateway and stops background monitoring.

---

## Desktop Setup (Linux, macOS, Windows)

### Recommended Setup: Zero-Touch / Set & Forget

For general users who want their laptop to automatically connect to campus Wi-Fi upon opening the lid or logging in, follow these 4 steps:

#### Step 1: Save Portal Credentials
Save your login details into your OS credential manager. Passwords are saved encrypted and masked:
```bash
# Interactive prompt (recommended to keep passwords out of shell history)
kawaii-wify login

# Or as a one-liner with flags:
kawaii-wify login -u YOUR_USERNAME -p "YOUR_PASSWORD"
```

#### Step 2: Apply Recommended Performance Settings
Tune detection frequency for snappy re-authentication while keeping CPU/network impact negligible:
```bash
# Check connection every 5 seconds
kawaii-wify config set check_interval 5s

# Let the daemon re-authenticate on demand rather than sending redundant keepalive pings
kawaii-wify config set keepalive false
```

#### Step 3: Enable Autostart on Boot/Login
Ensure the service launches silently in the background whenever you log into your computer:
```bash
kawaii-wify autostart enable
```
*(If you prefer manual control instead of automatic startup, simply skip this step.)*

#### Step 4: Start the Daemon
Start the background daemon for your current session:
```bash
kawaii-wify start
```

Verify that it is running and healthy:
```bash
kawaii-wify status
```

> [!NOTE]
> `kawaii-wify autostart enable` schedules the daemon for future logins and reboots. Running `kawaii-wify start` starts it immediately for your active session.

---

### Quick Copy-Paste One-Liners

**Linux / macOS (Bash/Zsh):**
```bash
kawaii-wify login -u YOUR_USERNAME -p "YOUR_PASSWORD" && \
kawaii-wify config set check_interval 5s && \
kawaii-wify config set keepalive false && \
kawaii-wify autostart enable && \
kawaii-wify start && \
kawaii-wify status
```

**Windows (PowerShell):**
```powershell
kawaii-wify login -u YOUR_USERNAME -p "YOUR_PASSWORD"; `
kawaii-wify config set check_interval 5s; `
kawaii-wify config set keepalive false; `
kawaii-wify autostart enable; `
kawaii-wify start; `
kawaii-wify status
```

---

## Everyday Desktop Controls

| Action | Command | Description |
| :--- | :--- | :--- |
| **Check Status** | `kawaii-wify status` | View connection state, active user, uptime, and last portal check. |
| **View Logs** | `kawaii-wify logs` | View recent background activity or troubleshoot failed logins. |
| **Force Login** | `kawaii-wify connect` | Immediately trigger authentication without waiting for the next check interval. |
| **Pause / Disconnect**| `kawaii-wify disconnect` | Log out from the firewall gateway and pause monitoring. |
| **Stop Daemon** | `kawaii-wify stop` | Gracefully terminate the background daemon. |
| **Update Password** | `kawaii-wify login` | Overwrite your saved credentials whenever your campus password expires. |
| **Notifications** | `kawaii-wify notify` | Manage and test OS desktop toast notifications and mascot alerts. |

---

## Power User & Advanced CLI Reference

### Daemon Execution Modes

```bash
# Start background daemon detached from terminal (survives closing the terminal)
kawaii-wify start

# Run daemon attached in foreground (useful for debugging and viewing live output)
kawaii-wify daemon

# Start with temporary gateway or username override
kawaii-wify start -u f20210001 --gateway fw.bits-pilani.ac.in:8090

# Start in paused state (monitors network, but skips automatic login until 'connect')
kawaii-wify start -p
```

### Viewing Logs

```bash
# View last 30 log lines
kawaii-wify logs

# View last 100 log lines
kawaii-wify logs -n 100

# Print the path to the log file on disk
kawaii-wify logs -p
```

### Configuration Options

Configuration files are saved in standard user config directories:
- **Linux**: `~/.config/kawaii-wify/config.json`
- **macOS**: `~/Library/Application Support/kawaii-wify/config.json`
- **Windows**: `%APPDATA%\kawaii-wify\config.json`

```bash
# View all settings
kawaii-wify config get

# View or change a specific setting
kawaii-wify config set check_interval 5s
kawaii-wify config set keepalive false
kawaii-wify config set gateway fw.bits-pilani.ac.in:8090
kawaii-wify config set verify_tls true

# Clear stored certificate pins (re-triggers TOFU on next connection)
kawaii-wify config clear-pins

# Hot-reload configuration into running daemon without restart
kawaii-wify config reload
```

> [!NOTE]
> Running background daemon processes automatically hot-reload configuration changes over IPC in real time without needing a restart.

| Setting | Default | Description |
| :--- | :--- | :--- |
| `check_interval` | `10s` | Polling and health-check interval (e.g., `3s`, `5s`, `15s`). Minimum enforced is `2s`. |
| `keepalive` | `true` | Periodic keepalive pings to gateway. Recommended: `false`. |
| `auto_connect` | `true` | Start monitoring and authenticating automatically upon launch. |
| `gateway` | `fw.bits-pilani.ac.in:8090` | Host and port of the FortiGate captive portal. |
| `verify_tls` | `true` | Enforces cryptographic SHA-256 certificate fingerprint validation. |
| `cert_pins` | `{}` | Map of trusted SHA-256 certificate fingerprints per gateway endpoint. |
| `notifications` | `true` | Enables or disables native desktop OS toast notifications and mascot alerts. |

### Desktop Toast Notifications & Mascot Alerts

Kawaii-Wify includes an event-driven desktop toast notification engine with embedded anime mascot avatars across Windows (WinRT), Linux (notify-send), and macOS (Notification Center):
- **Zero Periodic Noise**: The daemon remains silent during routine successful keepalives and normal connections.
- **Actionable Alerts**: Only triggers when user attention is required: security halts (gateway certificate mismatches / rogue portals), campus password rejections (circuit breaker cooldown), or background updates.

```bash
# Check notification status
kawaii-wify notify status

# Toggle desktop notifications (instantly hot-reloads running daemon)
kawaii-wify notify enable
kawaii-wify notify disable

# Test notification popups and mascot avatars
kawaii-wify notify test
kawaii-wify notify test auth_failed
kawaii-wify notify test update
```

### Complete CLI Command Reference

| Command | Flags & Options | Description |
| :--- | :--- | :--- |
| `kawaii-wify login` | `-u, --user <name>`<br>`-p, --password <pass>` | Enrolls credentials into the OS keyring. |
| `kawaii-wify logout` | `-u, --user <name>` | Clears credentials from OS keyring and disconnects session. |
| `kawaii-wify start` | `-u <name>`, `--gateway <endpoint>`<br>`-p, --paused`<br>`--[no-]keepalive`<br>`--[no-]auto-connect` | Spawns background daemon detached from terminal. |
| `kawaii-wify daemon` | *(same flags as start)* | Runs daemon in foreground for real-time console debugging. |
| `kawaii-wify status` | — | Queries running daemon over IPC for state, security alerts, and telemetry. |
| `kawaii-wify connect` | — | Commands running daemon to probe network and login immediately. |
| `kawaii-wify disconnect`| — | Revokes gateway lease and pauses daemon. |
| `kawaii-wify stop` | — | Gracefully stops the running daemon. |
| `kawaii-wify logs` | `-n, --lines <count>`<br>`-p, --path` | Displays recent logs or prints log file path. |
| `kawaii-wify autostart` | `enable` / `disable` / `status` | Configures OS login autostart (Registry, XDG, LaunchAgent). |
| `kawaii-wify notify` | `status` / `enable` / `disable`<br>`test [category]` | Manages and tests OS desktop toast notifications and mascot alerts. |
| `kawaii-wify config get` | `[key]` | Displays all or a single configuration value. |
| `kawaii-wify config set` | `<key> <val>` | Updates local config value (auto-reloaded into active daemon). |
| `kawaii-wify config clear-pins` | `[endpoint]` | Clears stored certificate pins and re-triggers TOFU pinning. |
| `kawaii-wify config reload` | — | Hot-reloads configuration from disk into the active daemon. |

---

## Building from Source

### Prerequisites
- **Go 1.22+** (for desktop CLI)
- **JDK 21** & **Android SDK** (for Android app)

```bash
# Build binary for current platform
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
