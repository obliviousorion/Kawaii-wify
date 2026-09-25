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

                // Check SSID whitelist if configured
                if (ssid != null && config.ssidWhitelist.isNotEmpty() && !config.ssidWhitelist.contains(ssid)) {
                    Logger.log("NET", "Current SSID '$ssid' is not in whitelist. Ignoring.", LogLevel.WARN)
                    return@launch
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
                    "Kawaii-Wify: Connection Online! (◕‿◕)✌",
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
                    R.drawable.wify_mascot_offline
                )
            }
            else -> {
                Triple(
                    "Kawaii-Wify: Resting (ᴗ˳ᴗ)",
                    "Engine is offline or waiting for Wi-Fi",
                    R.drawable.wify_mascot_offline
                )
            }
        }

        val avatar = BitmapFactory.decodeResource(resources, mascotRes)

        val builder = NotificationCompat.Builder(this, Constants.NOTIFICATION_CHANNEL_ID)
            .setContentTitle(title)
            .setContentText(content)
            .setSmallIcon(R.drawable.wify_mascot_online)
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
