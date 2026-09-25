package com.obliviousorion.kawaiiwify.data.local

import android.content.Context
import android.content.SharedPreferences
import androidx.security.crypto.EncryptedSharedPreferences
import androidx.security.crypto.MasterKey
import com.obliviousorion.kawaiiwify.core.LogLevel
import com.obliviousorion.kawaiiwify.core.Logger

data class Credentials(
    val username: String,
    val password: String
)

class SecurityManager(context: Context) {

    private val masterKey = MasterKey.Builder(context)
        .setKeyScheme(MasterKey.KeyScheme.AES256_GCM)
        .build()

    private val prefs: SharedPreferences = EncryptedSharedPreferences.create(
        context,
        "kawaii_wify_secure_vault",
        masterKey,
        EncryptedSharedPreferences.PrefKeyEncryptionScheme.AES256_SIV,
        EncryptedSharedPreferences.PrefValueEncryptionScheme.AES256_GCM
    )

    fun saveCredentials(credentials: Credentials) {
        prefs.edit()
            .putString(KEY_USERNAME, credentials.username.trim())
            .putString(KEY_PASSWORD, credentials.password)
            .apply()
        Logger.log("SECURITY", "Encrypted credentials updated in Android Hardware KeyStore", LogLevel.SUCCESS)
    }

    fun getCredentials(): Credentials? {
        val user = prefs.getString(KEY_USERNAME, null)
        val pass = prefs.getString(KEY_PASSWORD, null)
        return if (!user.isNullOrBlank() && !pass.isNullOrBlank()) {
            Credentials(user, pass)
        } else {
            null
        }
    }

    fun clearCredentials() {
        prefs.edit().clear().apply()
        Logger.log("SECURITY", "Encrypted credentials purged from KeyStore", LogLevel.WARN)
    }

    fun hasCredentials(): Boolean {
        return getCredentials() != null
    }

    companion object {
        private const val KEY_USERNAME = "sec_username"
        private const val KEY_PASSWORD = "sec_password"
    }
}
