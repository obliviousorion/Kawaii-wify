# Kawaii-Wify Android: Comprehensive Conversation Context

This document captures the full context, user requirements, technical discoveries, and design decisions recorded during ideation.

---

## 1. Project Background

`Kawaii-wify` is a lightweight background session manager and automated captive portal login daemon originally written in Go for FortiOS (FortiGate) captive portals, specifically designed for campus WLAN environments like BITS Pilani (`fw.bits-pilani.ac.in:8090`).

### The Goal
Create a native Android application port with:
1. **Anime Girl Theme with Dark Aesthetics**: Centered around a charming cyber-catgirl mascot named **"Wify-chan"**, utilizing AMOLED dark backgrounds (`#0A0B10`), neon sakura pinks, and electric cyan highlights.
2. **100% CLI Feature Parity**: Everything possible in the Golang CLI (`login`, `logout`, `connect`, `disconnect`, `status`, `start`, `stop`, `logs`, `config get/set`) must be directly supported on Android.
3. **Android-Native Capabilities**:
   - Quick Settings Tile (`KawaiiTileService`) for 1-tap toggling without opening the app.
   - Event-driven captive portal detection using `ConnectivityManager.NetworkCallback`.
   - Android Hardware KeyStore integration (`EncryptedSharedPreferences`).
   - Campus SSID whitelist filtering.

---

## 2. FortiGate Captive Portal Protocol Specifics

Extracted from the Golang implementation (`internal/auth/gateway.go` and `internal/engine/engine.go`):

1. **Default Gateway**: `fw.bits-pilani.ac.in:8090` (or `172.16.100.1`).
2. **Connectivity Probe**:
   - HTTP GET to `http://connectivitycheck.gstatic.com/generate_204`.
   - If HTTP status is `204`, internet is unblocked (`StateOnline`).
   - If intercepted, response HTML contains magic challenge token matching regex:
     ```regex
     fgtauth\?([a-f0-9]+)
     ```
3. **Challenge Priming**:
   - HTTP GET to `https://<gateway>/fgtauth?<magicToken>`.
   - Sets up challenge session on FortiOS with `User-Agent: Mozilla/5.0`.
4. **Authentication POST**:
   - POST to `https://<gateway>/` with `Content-Type: application/x-www-form-urlencoded`.
   - Form fields:
     - `4Tredir`: `http://connectivitycheck.gstatic.com/generate_204`
     - `magic`: `<magicToken>`
     - `username`: `<username>`
     - `password`: `<password>`
   - Header `Referer`: `https://<gateway>/fgtauth?<magicToken>`.
   - Response contains session token matching regex:
     ```regex
     keepalive\?([a-f0-9]+)
     ```
5. **Keepalive Ping**:
   - GET `https://<gateway>/keepalive?<sessionToken>`. Expects HTTP `200`.
6. **Logout**:
   - GET `https://<gateway>/logout?<sessionToken>`. Revokes remote session. Expects HTTP `200`.
7. **Circuit Breaker**:
   - Max 3 consecutive authentication failures (`MaxAuthFailures = 3`).
   - Cooldown period: 10s. If failures exceed max, engine enters `StateCooldown` and pauses to prevent campus account lockout.

---

## 3. Critical Android-Specific Challenges & Solutions

### A. The Socket Routing Trap
* **Symptom:** Android marks captive portals as having no internet. If mobile data is enabled, Android automatically routes all HTTP traffic over cellular. Thus, probes to `172.16.100.1` or `fw.bits-pilani.ac.in` fail with timeout or host unreachable.
* **Fix:** When captive portal is detected, call:
  ```kotlin
  connectivityManager.bindProcessToNetwork(network)
  ```
  And use `network.socketFactory` in OkHttp.

### B. Event-Driven OS Callbacks vs. Polling
* **Symptom:** Periodic polling every 10s wakes the CPU and triggers Android Doze mode thread-freezing.
* **Fix:** Register `ConnectivityManager.NetworkCallback` with `NetworkCapabilities.NET_CAPABILITY_CAPTIVE_PORTAL`. The OS wakes the app immediately upon network state change.

### C. Android 14+ Foreground Service Requirements
* **Requirement:** Declare `android:foregroundServiceType="dataSync"` in manifest.
* **Implementation:** Ongoing low-priority notification with Wify-chan avatar, live uptime, and `[ Disconnect ]` action button.
* Request `REQUEST_IGNORE_BATTERY_OPTIMIZATIONS` for reliable background keepalive pings.

---

## 4. Visual Assets Saved

The generated anime assets are stored in `docs/android/assets/`:
1. `wify_mascot_online.jpg`: Wify-chan cheering with peace sign and cyber headphones (`ONLINE`).
2. `wify_mascot_offline.jpg`: Wify-chan curled up asleep in dark hoodie (`OFFLINE`).
3. `wify_android_ui_mockup.jpg`: Full-screen Android mobile UI layout and design mockup.
