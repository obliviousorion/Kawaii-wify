package com.obliviousorion.kawaiiwify.service

import android.content.Intent
import android.os.Build
import android.service.quicksettings.Tile
import android.service.quicksettings.TileService
import androidx.annotation.RequiresApi
import com.obliviousorion.kawaiiwify.KawaiiApplication
import com.obliviousorion.kawaiiwify.core.Constants
import com.obliviousorion.kawaiiwify.domain.EngineState
import kotlinx.coroutines.*
import kotlinx.coroutines.flow.first

@RequiresApi(Build.VERSION_CODES.N)
class KawaiiTileService : TileService() {

    private val serviceScope = CoroutineScope(Dispatchers.Main + SupervisorJob())
    private var observeJob: Job? = null

    override fun onStartListening() {
        super.onStartListening()
        observeJob = serviceScope.launch {
            KawaiiApplication.instance.sessionEngine.telemetry.collect { telemetry ->
                updateTile(telemetry.state)
            }
        }
    }

    override fun onStopListening() {
        super.onStopListening()
        observeJob?.cancel()
    }

    override fun onClick() {
        super.onClick()
        val current = qsTile ?: return
        val isOnline = current.state == Tile.STATE_ACTIVE

        val intent = Intent(this, KeepaliveForegroundService::class.java).apply {
            action = if (isOnline) Constants.ACTION_DISCONNECT else Constants.ACTION_CONNECT
        }

        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.O) {
            startForegroundService(intent)
        } else {
            startService(intent)
        }
    }

    private fun updateTile(state: EngineState) {
        val tile = qsTile ?: return
        when (state) {
            is EngineState.Online -> {
                tile.state = Tile.STATE_ACTIVE
                tile.subtitle = "Connected (◕‿◕)"
            }
            is EngineState.Captive -> {
                tile.state = Tile.STATE_ACTIVE
                tile.subtitle = "Authenticating..."
            }
            is EngineState.Cooldown -> {
                tile.state = Tile.STATE_UNAVAILABLE
                tile.subtitle = "Cooldown (${state.remainingSeconds}s)"
            }
            else -> {
                tile.state = Tile.STATE_INACTIVE
                tile.subtitle = "Disconnected"
            }
        }
        tile.updateTile()
    }

    override fun onDestroy() {
        super.onDestroy()
        serviceScope.cancel()
    }
}
