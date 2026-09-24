# Kawaii-Wify Android: Implementation Plan & Linux Transition Guide

This document outlines the step-by-step roadmap to pick up development on **Linux**, configure the toolchain, and implement the native Android application with Jetpack Compose.

---

## 1. Prerequisites on Linux

Ensure the following tools are installed on your Linux workstation:
- **JDK 17 or 21**: `sudo apt install openjdk-17-jdk` (or distro equivalent)
- **Android SDK / Command-line Tools**:
  - `ANDROID_HOME` or `ANDROID_SDK_ROOT` exported in `~/.bashrc` or `~/.zshrc`.
  - Platforms: `android-34` or `android-35`
  - Build-tools: `34.0.0` or `35.0.0`
- **Gradle** (or use the Gradle wrapper `gradlew` generated in the project).
- **Android Studio** (optional, recommended for Compose layout previews and device/emulator debugging).

---

## 2. Phased Development Roadmap

### Phase 1: Project Scaffolding & Build Configuration
- [ ] Initialize Android project structure under `./android/`:
  - `settings.gradle.kts`
  - Root `build.gradle.kts`
  - `app/build.gradle.kts`
- [ ] Add dependencies:
  - Compose BOM (`androidx.compose:compose-bom:2024.09.00` or latest)
  - Material 3 (`androidx.compose.material3:material3`)
  - Navigation Compose (`androidx.navigation:navigation-compose:2.8.0`)
  - OkHttp 4 (`com.squareup.okhttp3:okhttp:4.12.0`)
  - Jetpack Security (`androidx.security:security-crypto:1.1.0-alpha06`)
  - Jetpack DataStore (`androidx.datastore:datastore-preferences:1.1.1`)
  - Coroutines (`org.jetbrains.kotlinx:kotlinx-coroutines-android:1.8.1`)
  - Coil for Compose (`io.coil-kt:coil-compose:2.6.0`)
- [ ] Configure `AndroidManifest.xml` with permissions and foreground service declarations.

### Phase 2: Core Networking & Security Layer
- [ ] Port `internal/auth/gateway.go` to Kotlin: `com.obliviousorion.kawaiiwify.auth.FortiGateAuth`:
  - `probe(network: Network?)`: Check `generate_204`, extract `fgtauth?<magic>`.
  - `prime(network: Network?, gateway: String, magic: String)`: HTTP GET priming request.
  - `login(...)`: POST credentials with `4Tredir` and extract `keepalive?<token>`.
  - `keepalive(...)`: Periodic GET ping maintaining lease.
  - `logout(...)`: GET `/logout?<token>` terminating remote FortiGate session.
- [ ] Port certificate handling: TLS hostname verification matching `internal/auth/client.go`.
- [ ] Port credentials management: `SecurityManager.kt` using `EncryptedSharedPreferences` (AES-256 GCM) backed by Android Hardware KeyStore.

### Phase 3: Android System Services & Background Lifecycle
- [ ] Implement `CaptivePortalCallback.kt` extending `ConnectivityManager.NetworkCallback`:
  - Listen for `NET_CAPABILITY_CAPTIVE_PORTAL`.
  - **Execute `connectivityManager.bindProcessToNetwork(network)`** to avoid socket routing trap.
  - Bind OkHttp `SocketFactory` to `network.socketFactory`.
  - Implement SSID whitelist check (`BITS-Pilani`, `BITS-Hostel`, etc.).
- [ ] Implement `KeepaliveForegroundService.kt`:
  - Declare `foregroundServiceType="dataSync"`.
  - Display sticky notification showing Wify-chan avatar, live status, and action buttons (`Disconnect`, `Logs`).
  - Manage keepalive coroutine loop while connected to campus Wi-Fi.
  - Circuit-breaker cooldown on 3 consecutive failures (matching Go engine).
- [ ] Implement `KawaiiTileService.kt`:
  - Quick Settings Tile enabling 1-tap connect/disconnect from notification shade.
- [ ] Implement `BootReceiver.kt`:
  - Auto-start service on boot if `auto_connect` is enabled.

### Phase 4: Cyber-Kawaii Dark Theme & Jetpack Compose UI
- [ ] Design System (`ui/theme/`):
  - AMOLED dark palette: `#0A0B10`, `#12131F`, `#1A1B2D`.
  - Neon accents: Sakura Pink `#FF69B4`, Cyber Cyan `#00F0FF`, Mint `#50FA7B`.
  - Monospace typography for telemetry and logs.
- [ ] Mascot Banner Composable (`ui/components/MascotBanner.kt`):
  - Dynamic avatar rendering Wify-chan based on state (`Online`, `Offline`, `Probing`, `Cooldown`).
  - Status aura pulse animations.
- [ ] Dashboard Screen (`ui/screens/DashboardScreen.kt`):
  - Glowing Neon Connect/Disconnect button.
  - Telemetry grid (Uptime, Last probe, Session token copy, Latency).
  - Quick status card.
- [ ] Live Terminal Log Screen (`ui/screens/LogsScreen.kt`):
  - In-memory rolling log buffer (mirroring `kawaii-wify logs`).
  - Filter chips (`[BOOT]`, `[STATE]`, `[AUTH]`, `[NET]`, `[ERROR]`).
  - Share/Export log file intent.
- [ ] Settings Screen (`ui/screens/SettingsScreen.kt`):
  - Gateway endpoint field.
  - Check interval slider (2s – 60s).
  - Keepalive & Auto-connect switches.
  - Campus SSID whitelist editor.
  - Battery optimization bypass trigger.
  - Purge account & KeyStore wipe (`logout`).

---

## 3. Quick Start Verification Commands (on Linux)

```bash
# Checkout the branch
git checkout dev-android

# Ensure assets and specs are in place
ls -la docs/android/
ls -la docs/android/assets/

# Build Android project once scaffolding is created
cd android
./gradlew assembleDebug

# Deploy to connected Android device or emulator
./gradlew installDebug
adb shell am start -n com.obliviousorion.kawaiiwify/.ui.MainActivity
```
