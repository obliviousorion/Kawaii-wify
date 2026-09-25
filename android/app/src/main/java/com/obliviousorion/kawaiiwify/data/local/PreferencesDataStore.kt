package com.obliviousorion.kawaiiwify.data.local

import android.content.Context
import androidx.datastore.core.DataStore
import androidx.datastore.preferences.core.*
import androidx.datastore.preferences.preferencesDataStore
import com.obliviousorion.kawaiiwify.core.Constants
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.catch
import kotlinx.coroutines.flow.map
import java.io.IOException

val Context.dataStore: DataStore<Preferences> by preferencesDataStore(name = "kawaii_wify_settings")

data class AppConfig(
    val gateway: String = Constants.DEFAULT_GATEWAY,
    val checkIntervalSeconds: Int = Constants.DEFAULT_CHECK_INTERVAL_SECONDS,
    val keepaliveEnabled: Boolean = true,
    val autoConnectEnabled: Boolean = true,
    val ssidWhitelist: Set<String> = setOf("BITS-Pilani", "BITS-Hostel"),
    val skipHostMismatch: Boolean = true
)

class PreferencesManager(private val context: Context) {

    val configFlow: Flow<AppConfig> = context.dataStore.data
        .catch { exception ->
            if (exception is IOException) {
                emit(emptyPreferences())
            } else {
                throw exception
            }
        }
        .map { prefs ->
            AppConfig(
                gateway = prefs[KEY_GATEWAY] ?: Constants.DEFAULT_GATEWAY,
                checkIntervalSeconds = prefs[KEY_INTERVAL] ?: Constants.DEFAULT_CHECK_INTERVAL_SECONDS,
                keepaliveEnabled = prefs[KEY_KEEPALIVE] ?: true,
                autoConnectEnabled = prefs[KEY_AUTOCONNECT] ?: true,
                ssidWhitelist = prefs[KEY_SSID_WHITELIST] ?: setOf("BITS-Pilani", "BITS-Hostel"),
                skipHostMismatch = prefs[KEY_SKIP_HOST] ?: true
            )
        }

    suspend fun updateGateway(gateway: String) {
        context.dataStore.edit { it[KEY_GATEWAY] = gateway.trim() }
    }

    suspend fun updateInterval(seconds: Int) {
        context.dataStore.edit { it[KEY_INTERVAL] = seconds.coerceIn(2, 60) }
    }

    suspend fun updateKeepalive(enabled: Boolean) {
        context.dataStore.edit { it[KEY_KEEPALIVE] = enabled }
    }

    suspend fun updateAutoConnect(enabled: Boolean) {
        context.dataStore.edit { it[KEY_AUTOCONNECT] = enabled }
    }

    suspend fun updateSsidWhitelist(ssids: Set<String>) {
        context.dataStore.edit { it[KEY_SSID_WHITELIST] = ssids }
    }

    suspend fun updateSkipHostMismatch(skip: Boolean) {
        context.dataStore.edit { it[KEY_SKIP_HOST] = skip }
    }

    companion object {
        private val KEY_GATEWAY = stringPreferencesKey("gateway_endpoint")
        private val KEY_INTERVAL = intPreferencesKey("check_interval_seconds")
        private val KEY_KEEPALIVE = booleanPreferencesKey("keepalive_enabled")
        private val KEY_AUTOCONNECT = booleanPreferencesKey("autoconnect_enabled")
        private val KEY_SSID_WHITELIST = stringSetPreferencesKey("ssid_whitelist")
        private val KEY_SKIP_HOST = booleanPreferencesKey("skip_host_mismatch")
    }
}
