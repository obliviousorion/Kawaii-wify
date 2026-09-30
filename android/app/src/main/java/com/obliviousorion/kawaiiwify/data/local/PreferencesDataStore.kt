package com.obliviousorion.kawaiiwify.data.local

import android.content.Context
import androidx.datastore.core.DataStore
import androidx.datastore.preferences.core.*
import androidx.datastore.preferences.preferencesDataStore
import com.obliviousorion.kawaiiwify.core.Constants
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.catch
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.flow.map
import java.io.IOException

val Context.dataStore: DataStore<Preferences> by preferencesDataStore(name = "kawaii_wify_settings")

data class AppConfig(
    val gateway: String = Constants.DEFAULT_GATEWAY,
    val checkIntervalSeconds: Int = Constants.DEFAULT_CHECK_INTERVAL_SECONDS,
    val keepaliveEnabled: Boolean = true,
    val autoConnectEnabled: Boolean = true,
    val ssidWhitelist: Set<String> = setOf("BITS-STAFF", "BITS-STUDENT"),
    val enforceSsidWhitelist: Boolean = false, // Explicit opt-in so users without Location permission aren't locked out
    val skipHostMismatch: Boolean = true,
    val verifyTls: Boolean = true,
    val certPins: Map<String, List<String>> = emptyMap() // "endpoint" -> ["SHA256:..."]
) {
    fun getPinsForGateway(gw: String = gateway): List<String> {
        val cleanEndpoint = gw.trim().removePrefix("https://").removePrefix("http://").removeSuffix("/")
        val withPort = if (!cleanEndpoint.contains(":")) "$cleanEndpoint:8090" else cleanEndpoint
        return certPins[withPort] ?: certPins[cleanEndpoint] ?: emptyList()
    }
}

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
            val rawWhitelist = prefs[KEY_SSID_WHITELIST]
            val sanitizedWhitelist = when {
                rawWhitelist == null -> setOf("BITS-STAFF", "BITS-STUDENT")
                rawWhitelist.contains("BITS-Hostel") || rawWhitelist.contains("BITS-Pilani") -> {
                    (rawWhitelist - setOf("BITS-Hostel", "BITS-Pilani") + setOf("BITS-STAFF", "BITS-STUDENT")).toSet()
                }
                else -> rawWhitelist
            }

            val rawPins = prefs[KEY_CERT_PINS_SET] ?: emptySet()
            val parsedPins = mutableMapOf<String, MutableList<String>>()
            for (entry in rawPins) {
                val parts = entry.split("|", limit = 2)
                if (parts.size == 2) {
                    parsedPins.getOrPut(parts[0]) { mutableListOf() }.add(parts[1])
                }
            }

            AppConfig(
                gateway = prefs[KEY_GATEWAY] ?: Constants.DEFAULT_GATEWAY,
                checkIntervalSeconds = prefs[KEY_INTERVAL] ?: Constants.DEFAULT_CHECK_INTERVAL_SECONDS,
                keepaliveEnabled = prefs[KEY_KEEPALIVE] ?: true,
                autoConnectEnabled = prefs[KEY_AUTOCONNECT] ?: true,
                ssidWhitelist = sanitizedWhitelist,
                enforceSsidWhitelist = prefs[KEY_ENFORCE_SSID_WHITELIST] ?: false,
                skipHostMismatch = prefs[KEY_SKIP_HOST] ?: true,
                verifyTls = prefs[KEY_VERIFY_TLS] ?: true,
                certPins = parsedPins
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

    suspend fun updateEnforceSsidWhitelist(enforce: Boolean) {
        context.dataStore.edit { it[KEY_ENFORCE_SSID_WHITELIST] = enforce }
    }

    suspend fun updateSkipHostMismatch(skip: Boolean) {
        context.dataStore.edit { it[KEY_SKIP_HOST] = skip }
    }

    suspend fun updateVerifyTls(verify: Boolean) {
        context.dataStore.edit { it[KEY_VERIFY_TLS] = verify }
    }

    suspend fun addCertPin(endpoint: String, fingerprint: String) {
        val cleanEndpoint = endpoint.trim().removePrefix("https://").removePrefix("http://").removeSuffix("/")
        val withPort = if (!cleanEndpoint.contains(":")) "$cleanEndpoint:8090" else cleanEndpoint

        context.dataStore.edit { prefs ->
            val current = prefs[KEY_CERT_PINS_SET]?.toMutableSet() ?: mutableSetOf()
            val entry = "$withPort|$fingerprint"
            if (!current.contains(entry)) {
                current.add(entry)
                prefs[KEY_CERT_PINS_SET] = current
            }
        }
    }

    suspend fun clearCertPins(endpoint: String? = null) {
        context.dataStore.edit { prefs ->
            if (endpoint == null) {
                prefs.remove(KEY_CERT_PINS_SET)
            } else {
                val cleanEndpoint = endpoint.trim().removePrefix("https://").removePrefix("http://").removeSuffix("/")
                val withPort = if (!cleanEndpoint.contains(":")) "$cleanEndpoint:8090" else cleanEndpoint
                val current = prefs[KEY_CERT_PINS_SET]?.toMutableSet() ?: mutableSetOf()
                current.removeAll { it.startsWith("$withPort|") || it.startsWith("$cleanEndpoint|") }
                prefs[KEY_CERT_PINS_SET] = current
            }
        }
    }

    suspend fun getLastUpdateCheckTime(): Long {
        val prefs = context.dataStore.data.first()
        return prefs[KEY_LAST_UPDATE_CHECK_TIME] ?: 0L
    }

    suspend fun updateLastUpdateCheckTime(timestamp: Long) {
        context.dataStore.edit { it[KEY_LAST_UPDATE_CHECK_TIME] = timestamp }
    }

    suspend fun getLastNotifiedUpdateVersion(): String {
        val prefs = context.dataStore.data.first()
        return prefs[KEY_LAST_NOTIFIED_UPDATE_VERSION] ?: ""
    }

    suspend fun updateLastNotifiedUpdateVersion(version: String) {
        context.dataStore.edit { it[KEY_LAST_NOTIFIED_UPDATE_VERSION] = version }
    }

    companion object {
        private val KEY_GATEWAY = stringPreferencesKey("gateway_endpoint")
        private val KEY_INTERVAL = intPreferencesKey("check_interval_seconds")
        private val KEY_KEEPALIVE = booleanPreferencesKey("keepalive_enabled")
        private val KEY_AUTOCONNECT = booleanPreferencesKey("autoconnect_enabled")
        private val KEY_SSID_WHITELIST = stringSetPreferencesKey("ssid_whitelist")
        private val KEY_ENFORCE_SSID_WHITELIST = booleanPreferencesKey("enforce_ssid_whitelist")
        private val KEY_SKIP_HOST = booleanPreferencesKey("skip_host_mismatch")
        private val KEY_VERIFY_TLS = booleanPreferencesKey("verify_tls_enabled")
        private val KEY_CERT_PINS_SET = stringSetPreferencesKey("cert_pins_set")
        private val KEY_LAST_UPDATE_CHECK_TIME = longPreferencesKey("last_update_check_time")
        private val KEY_LAST_NOTIFIED_UPDATE_VERSION = stringPreferencesKey("last_notified_update_version")
    }
}
