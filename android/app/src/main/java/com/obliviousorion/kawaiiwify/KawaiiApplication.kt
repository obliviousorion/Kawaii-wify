package com.obliviousorion.kawaiiwify

import android.app.Application
import android.app.NotificationChannel
import android.app.NotificationManager
import android.os.Build
import com.obliviousorion.kawaiiwify.core.Constants
import com.obliviousorion.kawaiiwify.core.LogLevel
import com.obliviousorion.kawaiiwify.core.Logger
import com.obliviousorion.kawaiiwify.data.local.PreferencesManager
import com.obliviousorion.kawaiiwify.data.local.SecurityManager
import com.obliviousorion.kawaiiwify.domain.SessionEngine
import com.obliviousorion.kawaiiwify.data.updater.AppUpdateManager
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.launch

class KawaiiApplication : Application() {

    lateinit var securityManager: SecurityManager
        private set

    lateinit var preferencesManager: PreferencesManager
        private set

    lateinit var sessionEngine: SessionEngine
        private set

    lateinit var updateManager: AppUpdateManager
        private set

    override fun onCreate() {
        super.onCreate()
        instance = this

        Logger.init(this)
        securityManager = SecurityManager(this)
        preferencesManager = PreferencesManager(this)
        sessionEngine = SessionEngine()
        updateManager = AppUpdateManager(this, preferencesManager)

        // Persist TOFU certificate pins when authentic gateway proves identity
        sessionEngine.auth.onCommitPin = { endpoint, pin ->
            CoroutineScope(Dispatchers.IO).launch {
                preferencesManager.addCertPin(endpoint, pin)
                Logger.log("SECURITY", "Pinned trusted gateway certificate: $endpoint -> $pin", LogLevel.SUCCESS)
            }
        }

        createNotificationChannels()
    }

    private fun createNotificationChannels() {
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.O) {
            val manager = getSystemService(NotificationManager::class.java)

            // Low-importance channel for continuous background telemetry
            val statusChannel = NotificationChannel(
                Constants.NOTIFICATION_CHANNEL_ID,
                getString(R.string.channel_name),
                NotificationManager.IMPORTANCE_LOW
            ).apply {
                description = getString(R.string.channel_desc)
                setShowBadge(false)
            }

            // High-importance channel for security warnings and certificate alerts
            val securityChannel = NotificationChannel(
                Constants.SECURITY_ALERT_CHANNEL_ID,
                "Security & Gating Alerts",
                NotificationManager.IMPORTANCE_HIGH
            ).apply {
                description = "Urgent notifications regarding certificate mismatches, rogue portals, and SSID gating"
                setShowBadge(true)
            }

            // Software update notifications
            val updateChannel = NotificationChannel(
                Constants.UPDATE_ALERT_CHANNEL_ID,
                "Software Updates",
                NotificationManager.IMPORTANCE_DEFAULT
            ).apply {
                description = "Notifications when newer releases of Kawaii-Wify are available"
                setShowBadge(true)
            }

            manager.createNotificationChannel(statusChannel)
            manager.createNotificationChannel(securityChannel)
            manager.createNotificationChannel(updateChannel)
        }
    }

    companion object {
        lateinit var instance: KawaiiApplication
            private set
    }
}
