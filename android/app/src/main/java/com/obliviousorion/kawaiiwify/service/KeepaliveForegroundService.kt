package com.obliviousorion.kawaiiwify.service

import android.app.Notification
import android.app.PendingIntent
import android.app.Service
import android.content.Context
import android.content.Intent
import android.content.pm.ServiceInfo
import android.graphics.BitmapFactory
import android.net.ConnectivityManager
import android.net.NetworkRequest
import android.os.Build
import android.os.IBinder
import androidx.core.app.NotificationCompat
import androidx.core.app.NotificationManagerCompat
import androidx.core.app.ServiceCompat
import com.obliviousorion.kawaiiwify.KawaiiApplication
import com.obliviousorion.kawaiiwify.R
import com.obliviousorion.kawaiiwify.core.Constants
import com.obliviousorion.kawaiiwify.core.LogLevel
import com.obliviousorion.kawaiiwify.core.Logger
import com.obliviousorion.kawaiiwify.domain.EngineState
import com.obliviousorion.kawaiiwify.ui.MainActivity
import kotlinx.coroutines.*
import kotlinx.coroutines.flow.first

class KeepaliveForegroundService : Service() {

    private val serviceScope = CoroutineScope(Dispatchers.IO + SupervisorJob())
    private var loopJob: Job? = null
    private var updateJob: Job? = null

    private lateinit var connectivityManager: ConnectivityManager
    private lateinit var networkCallback: CaptivePortalCallback

    override fun onCreate() {
        super.onCreate()
        Logger.log("BOOT", "KeepaliveForegroundService starting...", LogLevel.INFO)

        connectivityManager = getSystemService(Context.CONNECTIVITY_SERVICE) as ConnectivityManager
        networkCallback = CaptivePortalCallback(this) { network, ssid ->
            serviceScope.launch {
                val config = KawaiiApplication.instance.preferencesManager.configFlow.first()
                val credentials = KawaiiApplication.instance.securityManager.getCredentials()

                // Check SSID whitelist only if explicitly enforced
                if (config.enforceSsidWhitelist && config.ssidWhitelist.isNotEmpty()) {
                    if (ssid == null || !config.ssidWhitelist.contains(ssid)) {
                        Logger.log("NET", "Current SSID '$ssid' is not permitted by whitelist. Ignoring.", LogLevel.WARN)
                        KawaiiApplication.instance.sessionEngine.tick(network, config, credentials, ssid)
                        return@launch
                    }
                }

                KawaiiApplication.instance.sessionEngine.tick(network, config, credentials, ssid)
            }
        }

        val request = NetworkRequest.Builder()
            .addTransportType(android.net.NetworkCapabilities.TRANSPORT_WIFI)
            .build()

        connectivityManager.registerNetworkCallback(request, networkCallback)

        startForegroundNotification()
        observeTelemetry()
        startDaemonLoop()
        startPeriodicUpdateCheck()
    }

    override fun onStartCommand(intent: Intent?, flags: Int, startId: Int): Int {
        when (intent?.action) {
            Constants.ACTION_CONNECT -> {
                serviceScope.launch {
                    val config = KawaiiApplication.instance.preferencesManager.configFlow.first()
                    val credentials = KawaiiApplication.instance.securityManager.getCredentials()
                    KawaiiApplication.instance.sessionEngine.connect(
                        networkCallback.currentNetwork,
                        config,
                        credentials,
                        networkCallback.currentSsid
                    )
                    KawaiiApplication.instance.sessionEngine.tick(
                        networkCallback.currentNetwork,
                        config,
                        credentials,
                        networkCallback.currentSsid
                    )
                }
            }

            Constants.ACTION_DISCONNECT -> {
                serviceScope.launch {
                    val config = KawaiiApplication.instance.preferencesManager.configFlow.first()
                    KawaiiApplication.instance.sessionEngine.disconnect(networkCallback.currentNetwork, config)
                }
            }

            Constants.ACTION_PAUSE -> {
                serviceScope.launch {
                    KawaiiApplication.instance.sessionEngine.pause()
                }
            }

            Constants.ACTION_RESUME -> {
                serviceScope.launch {
                    val config = KawaiiApplication.instance.preferencesManager.configFlow.first()
                    val credentials = KawaiiApplication.instance.securityManager.getCredentials()
                    KawaiiApplication.instance.sessionEngine.resume(
                        networkCallback.currentNetwork,
                        config,
                        credentials,
                        networkCallback.currentSsid
                    )
                    KawaiiApplication.instance.sessionEngine.tick(
                        networkCallback.currentNetwork,
                        config,
                        credentials,
                        networkCallback.currentSsid
                    )
                }
            }

            Constants.ACTION_STOP_SERVICE -> {
                stopSelf()
            }
        }
        return START_STICKY
    }

    private fun startDaemonLoop() {
        loopJob?.cancel()
        loopJob = serviceScope.launch {
            while (isActive) {
                val config = KawaiiApplication.instance.preferencesManager.configFlow.first()
                val credentials = KawaiiApplication.instance.securityManager.getCredentials()

                KawaiiApplication.instance.sessionEngine.tick(
                    network = networkCallback.currentNetwork,
                    config = config,
                    credentials = credentials,
                    activeSsid = networkCallback.currentSsid
                )

                delay(config.checkIntervalSeconds * 1000L)
            }
        }
    }

    private fun observeTelemetry() {
        serviceScope.launch {
            KawaiiApplication.instance.sessionEngine.telemetry.collect { telemetry ->
                updateNotification(telemetry.state, telemetry.uptimeSeconds)
            }
        }
    }

    private fun startForegroundNotification() {
        val notification = buildNotification(EngineState.Offline, 0L)
        val serviceType = if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.Q) {
            ServiceInfo.FOREGROUND_SERVICE_TYPE_DATA_SYNC
        } else {
            0
        }
        ServiceCompat.startForeground(this, Constants.NOTIFICATION_ID, notification, serviceType)
    }

    private fun updateNotification(state: EngineState, uptimeSeconds: Long) {
        val notification = buildNotification(state, uptimeSeconds)
        val manager = getSystemService(Context.NOTIFICATION_SERVICE) as android.app.NotificationManager
        manager.notify(Constants.NOTIFICATION_ID, notification)

        if (state is EngineState.SecurityHalted) {
            notifySecurityAlert(state.reason)
        }
    }

    private fun notifySecurityAlert(reason: String) {
        val openIntent = Intent(this, MainActivity::class.java).apply {
            flags = Intent.FLAG_ACTIVITY_SINGLE_TOP
        }
        val openPendingIntent = PendingIntent.getActivity(
            this, 99, openIntent, PendingIntent.FLAG_IMMUTABLE or PendingIntent.FLAG_UPDATE_CURRENT
        )

        val avatar = BitmapFactory.decodeResource(resources, R.drawable.wify_mascot_security)

        val alertNotification = NotificationCompat.Builder(this, Constants.SECURITY_ALERT_CHANNEL_ID)
            .setContentTitle("Security Alert: Connection Halted")
            .setContentText(reason)
            .setStyle(NotificationCompat.BigTextStyle().bigText("Kawaii-Wify suspended automatic authentication to protect your credentials. Reason: $reason"))
            .setSmallIcon(R.drawable.ic_stat_kawaii_wifi)
            .setLargeIcon(avatar)
            .setContentIntent(openPendingIntent)
            .setAutoCancel(true)
            .setPriority(NotificationCompat.PRIORITY_HIGH)
            .build()

        val manager = getSystemService(Context.NOTIFICATION_SERVICE) as android.app.NotificationManager
        manager.notify(Constants.SECURITY_NOTIFICATION_ID, alertNotification)
    }

    private fun startPeriodicUpdateCheck() {
        updateJob?.cancel()
        updateJob = serviceScope.launch {
            // Initial 30-second delay so critical Wi-Fi keepalive starts uninterrupted
            delay(30_000L)
            while (isActive) {
                try {
                    val updateManager = KawaiiApplication.instance.updateManager
                    val updateInfo = updateManager.checkForUpdates(force = false)
                    if (updateInfo.hasUpdate && updateManager.shouldNotifyUpdate(updateInfo)) {
                        notifyUpdateAvailable(updateInfo.latestVersion)
                        updateManager.markUpdateNotified(updateInfo.latestVersion)
                    }
                } catch (e: Exception) {
                    Logger.log("UPDATE", "Periodic update check encountered error: ${e.message}", LogLevel.DEBUG)
                }
                // Check once every 24 hours
                delay(24 * 60 * 60 * 1000L)
            }
        }
    }

    private fun notifyUpdateAvailable(version: String) {
        try {
            if (!NotificationManagerCompat.from(this).areNotificationsEnabled()) {
                return
            }
            val openIntent = Intent(this, MainActivity::class.java).apply {
                flags = Intent.FLAG_ACTIVITY_SINGLE_TOP or Intent.FLAG_ACTIVITY_CLEAR_TOP
                putExtra(Constants.EXTRA_OPEN_UPDATE_DIALOG, true)
            }
            val openPendingIntent = PendingIntent.getActivity(
                this, 100, openIntent, PendingIntent.FLAG_IMMUTABLE or PendingIntent.FLAG_UPDATE_CURRENT
            )

            val avatar = BitmapFactory.decodeResource(resources, R.drawable.wify_mascot_update)

            val notification = NotificationCompat.Builder(this, Constants.UPDATE_ALERT_CHANNEL_ID)
                .setContentTitle("Kawaii-Wify: Update Available")
                .setContentText("Version v$version is available. Tap to view.")
                .setStyle(NotificationCompat.BigTextStyle().bigText("Version v$version is available with security enhancements and improvements. Tap to view or download."))
                .setSmallIcon(R.drawable.ic_stat_kawaii_wifi)
                .setLargeIcon(avatar)
                .setContentIntent(openPendingIntent)
                .setAutoCancel(true)
                .setPriority(NotificationCompat.PRIORITY_DEFAULT)
                .build()

            val manager = getSystemService(Context.NOTIFICATION_SERVICE) as android.app.NotificationManager
            manager.notify(Constants.UPDATE_NOTIFICATION_ID, notification)
        } catch (e: Exception) {
            Logger.log("UPDATE", "Failed to dispatch update notification: ${e.message}", LogLevel.WARN)
        }
    }

    private fun buildNotification(state: EngineState, uptimeSeconds: Long): Notification {
        val openIntent = Intent(this, MainActivity::class.java).apply {
            flags = Intent.FLAG_ACTIVITY_SINGLE_TOP
        }
        val openPendingIntent = PendingIntent.getActivity(
            this, 0, openIntent, PendingIntent.FLAG_IMMUTABLE or PendingIntent.FLAG_UPDATE_CURRENT
        )

        val connectIntent = Intent(this, KeepaliveForegroundService::class.java).apply {
            action = Constants.ACTION_CONNECT
        }
        val connectPending = PendingIntent.getService(
            this, 1, connectIntent, PendingIntent.FLAG_IMMUTABLE or PendingIntent.FLAG_UPDATE_CURRENT
        )

        val disconnectIntent = Intent(this, KeepaliveForegroundService::class.java).apply {
            action = Constants.ACTION_DISCONNECT
        }
        val disconnectPending = PendingIntent.getService(
            this, 2, disconnectIntent, PendingIntent.FLAG_IMMUTABLE or PendingIntent.FLAG_UPDATE_CURRENT
        )

        val pauseIntent = Intent(this, KeepaliveForegroundService::class.java).apply {
            action = Constants.ACTION_PAUSE
        }
        val pausePending = PendingIntent.getService(
            this, 3, pauseIntent, PendingIntent.FLAG_IMMUTABLE or PendingIntent.FLAG_UPDATE_CURRENT
        )

        val resumeIntent = Intent(this, KeepaliveForegroundService::class.java).apply {
            action = Constants.ACTION_RESUME
        }
        val resumePending = PendingIntent.getService(
            this, 4, resumeIntent, PendingIntent.FLAG_IMMUTABLE or PendingIntent.FLAG_UPDATE_CURRENT
        )

        val (title, content, mascotRes) = when (state) {
            is EngineState.Online -> {
                val hours = uptimeSeconds / 3600
                val mins = (uptimeSeconds % 3600) / 60
                val secs = uptimeSeconds % 60
                Triple(
                    "Kawaii-Wify: Connection Online",
                    "Session active • Uptime %02d:%02d:%02d".format(hours, mins, secs),
                    R.drawable.wify_mascot_online
                )
            }
            is EngineState.Paused -> {
                Triple(
                    "Kawaii-Wify: Paused (Firewall Untouched)",
                    "Daemon paused • Background checks frozen",
                    R.drawable.wify_mascot_offline
                )
            }
            is EngineState.Captive -> {
                Triple(
                    "Kawaii-Wify: Captive Portal Intercepted",
                    "Authenticating with FortiGate...",
                    R.drawable.wify_mascot_online
                )
            }
            is EngineState.Cooldown -> {
                Triple(
                    "Kawaii-Wify: Circuit Breaker Cooldown",
                    "Pausing to protect account (${state.remainingSeconds}s remaining)",
                    R.drawable.wify_mascot_auth_failed
                )
            }
            is EngineState.SecurityHalted -> {
                Triple(
                    "Kawaii-Wify: Security Alert",
                    "Halted: ${state.reason}",
                    R.drawable.wify_mascot_security
                )
            }
            is EngineState.BlockedByWhitelist -> {
                Triple(
                    "Kawaii-Wify: Wi-Fi Not Permitted",
                    "SSID '${state.currentSsid ?: "Unknown"}' is not whitelisted",
                    R.drawable.wify_mascot_offline
                )
            }
            else -> {
                Triple(
                    "Kawaii-Wify: Offline",
                    "Engine is offline or waiting for Wi-Fi",
                    R.drawable.wify_mascot_offline
                )
            }
        }

        val avatar = BitmapFactory.decodeResource(resources, mascotRes)

        val builder = NotificationCompat.Builder(this, Constants.NOTIFICATION_CHANNEL_ID)
            .setContentTitle(title)
            .setContentText(content)
            .setSmallIcon(R.drawable.ic_stat_kawaii_wifi)
            .setLargeIcon(avatar)
            .setContentIntent(openPendingIntent)
            .setOngoing(true)
            .setOnlyAlertOnce(true)
            .setPriority(NotificationCompat.PRIORITY_LOW)

        when (state) {
            is EngineState.Online -> {
                builder.addAction(android.R.drawable.ic_media_pause, "Pause", pausePending)
                builder.addAction(android.R.drawable.ic_menu_close_clear_cancel, "Disconnect", disconnectPending)
            }
            is EngineState.Paused -> {
                builder.addAction(android.R.drawable.ic_media_play, "Resume", resumePending)
                builder.addAction(android.R.drawable.ic_menu_close_clear_cancel, "Disconnect", disconnectPending)
            }
            is EngineState.SecurityHalted -> {
                builder.addAction(android.R.drawable.ic_menu_close_clear_cancel, "Dismiss", disconnectPending)
            }
            else -> {
                builder.addAction(android.R.drawable.ic_media_play, "Connect", connectPending)
            }
        }

        return builder.build()
    }

    override fun onDestroy() {
        super.onDestroy()
        Logger.log("BOOT", "KeepaliveForegroundService destroyed", LogLevel.WARN)
        try {
            connectivityManager.unregisterNetworkCallback(networkCallback)
        } catch (_: Exception) {}
        loopJob?.cancel()
        updateJob?.cancel()
        serviceScope.cancel()
    }

    override fun onBind(intent: Intent?): IBinder? = null

    companion object {
        fun start(context: Context) {
            val intent = Intent(context, KeepaliveForegroundService::class.java)
            if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.O) {
                context.startForegroundService(intent)
            } else {
                context.startService(intent)
            }
        }

        fun stop(context: Context) {
            val intent = Intent(context, KeepaliveForegroundService::class.java)
            context.stopService(intent)
        }
    }
}
