package com.obliviousorion.kawaiiwify.service

import android.content.BroadcastReceiver
import android.content.Context
import android.content.Intent
import com.obliviousorion.kawaiiwify.KawaiiApplication
import com.obliviousorion.kawaiiwify.core.LogLevel
import com.obliviousorion.kawaiiwify.core.Logger
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.launch

class BootReceiver : BroadcastReceiver() {
    override fun onReceive(context: Context, intent: Intent) {
        if (intent.action == Intent.ACTION_BOOT_COMPLETED) {
            Logger.log("BOOT", "Device reboot detected", LogLevel.INFO)
            CoroutineScope(Dispatchers.IO).launch {
                val config = KawaiiApplication.instance.preferencesManager.configFlow.first()
                if (config.autoConnectEnabled) {
                    Logger.log("BOOT", "Auto-connect is enabled, launching KeepaliveForegroundService", LogLevel.SUCCESS)
                    KeepaliveForegroundService.start(context)
                }
            }
        }
    }
}
