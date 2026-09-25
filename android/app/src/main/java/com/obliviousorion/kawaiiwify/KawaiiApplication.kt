package com.obliviousorion.kawaiiwify

import android.app.Application
import android.app.NotificationChannel
import android.app.NotificationManager
import android.os.Build
import com.obliviousorion.kawaiiwify.core.Constants
import com.obliviousorion.kawaiiwify.core.Logger
import com.obliviousorion.kawaiiwify.data.local.PreferencesManager
import com.obliviousorion.kawaiiwify.data.local.SecurityManager
import com.obliviousorion.kawaiiwify.domain.SessionEngine

class KawaiiApplication : Application() {

    lateinit var securityManager: SecurityManager
        private set

    lateinit var preferencesManager: PreferencesManager
        private set

    lateinit var sessionEngine: SessionEngine
        private set

    override fun onCreate() {
        super.onCreate()
        instance = this

        Logger.init(this)
        securityManager = SecurityManager(this)
        preferencesManager = PreferencesManager(this)
        sessionEngine = SessionEngine()

        createNotificationChannel()
    }

    private fun createNotificationChannel() {
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.O) {
            val channel = NotificationChannel(
                Constants.NOTIFICATION_CHANNEL_ID,
                getString(R.string.channel_name),
                NotificationManager.IMPORTANCE_LOW
            ).apply {
                description = getString(R.string.channel_desc)
                setShowBadge(false)
            }
            val manager = getSystemService(NotificationManager::class.java)
            manager.createNotificationChannel(channel)
        }
    }

    companion object {
        lateinit var instance: KawaiiApplication
            private set
    }
}
