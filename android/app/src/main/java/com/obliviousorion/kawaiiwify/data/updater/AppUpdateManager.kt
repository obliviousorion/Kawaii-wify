package com.obliviousorion.kawaiiwify.data.updater

import android.content.Context
import com.obliviousorion.kawaiiwify.BuildConfig
import com.obliviousorion.kawaiiwify.data.local.PreferencesManager
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.withContext
import okhttp3.OkHttpClient
import okhttp3.Request
import org.json.JSONObject
import java.util.concurrent.TimeUnit

data class UpdateInfo(
    val hasUpdate: Boolean = false,
    val latestVersion: String = "",
    val apkDownloadUrl: String = "",
    val releasePageUrl: String = "",
    val releaseNotes: String = ""
)

class AppUpdateManager(
    private val context: Context,
    private val prefManager: PreferencesManager
) {
    private val client = OkHttpClient.Builder()
        .connectTimeout(15, TimeUnit.SECONDS)
        .readTimeout(15, TimeUnit.SECONDS)
        .build()

    private val _updateState = MutableStateFlow(UpdateInfo())
    val updateState: StateFlow<UpdateInfo> = _updateState.asStateFlow()

    suspend fun checkForUpdates(force: Boolean = false): UpdateInfo = withContext(Dispatchers.IO) {
        val now = System.currentTimeMillis()
        val lastCheck = prefManager.getLastUpdateCheckTime()
        val oneDayMillis = 24 * 60 * 60 * 1000L

        if (!force && (now - lastCheck) < oneDayMillis) {
            return@withContext _updateState.value
        }

        try {
            // 1. Direct CDN Check via versions.json (Immune to GitHub API 60 req/hr rate limits)
            val cdnReq = Request.Builder()
                .url("https://github.com/obliviousorion/Kawaii-wify/releases/latest/download/versions.json")
                .header("User-Agent", "Kawaii-Wify-Android/${BuildConfig.VERSION_NAME}")
                .build()

            val cdnResp = client.newCall(cdnReq).execute()
            if (cdnResp.isSuccessful) {
                val cdnBody = cdnResp.body?.string()
                if (!cdnBody.isNullOrBlank()) {
                    val mJson = JSONObject(cdnBody)
                    val targetAndroidVersion = mJson.optString("android", "").trim().removePrefix("v")
                    val releaseUrl = mJson.optString("release_url", "https://github.com/obliviousorion/Kawaii-wify/releases/latest")
                    val apkName = mJson.optString("android_apk_name", "").trim()
                    val apkUrl = "https://github.com/obliviousorion/Kawaii-wify/releases/latest/download/" +
                            (if (apkName.isNotBlank()) apkName else "kawaii-wify-android.apk")

                    val isNewer = targetAndroidVersion.isNotBlank() && isNewerSemver(targetAndroidVersion, BuildConfig.VERSION_NAME)
                    prefManager.updateLastUpdateCheckTime(now)

                    val info = if (isNewer) {
                        UpdateInfo(
                            hasUpdate = true,
                            latestVersion = targetAndroidVersion,
                            apkDownloadUrl = apkUrl,
                            releasePageUrl = releaseUrl,
                            releaseNotes = "A newer version of Kawaii-Wify ($targetAndroidVersion) is available on GitHub."
                        )
                    } else {
                        UpdateInfo(hasUpdate = false)
                    }
                    _updateState.value = info
                    return@withContext info
                }
            }
        } catch (_: Exception) {
            // If CDN is unavailable, fall through to REST API fallback below
        }

        try {
            // 2. Fallback to GitHub REST API
            val req = Request.Builder()
                .url("https://api.github.com/repos/obliviousorion/Kawaii-wify/releases/latest")
                .header("User-Agent", "Kawaii-Wify-Android/${BuildConfig.VERSION_NAME}")
                .header("Accept", "application/vnd.github.v3+json")
                .build()

            val resp = client.newCall(req).execute()
            if (!resp.isSuccessful) {
                return@withContext _updateState.value
            }

            val bodyStr = resp.body?.string() ?: return@withContext _updateState.value
            val json = JSONObject(bodyStr)

            val tagName = json.optString("tag_name", "").trim().removePrefix("v")
            val htmlUrl = json.optString("html_url", "")
            val bodyText = json.optString("body", "")
            val assets = json.optJSONArray("assets")

            var apkUrl = ""
            var versionsManifestUrl = ""

            if (assets != null) {
                for (i in 0 until assets.length()) {
                    val asset = assets.getJSONObject(i)
                    val name = asset.optString("name", "")
                    if (name.startsWith("kawaii-wify-android") && name.endsWith(".apk")) {
                        apkUrl = asset.optString("browser_download_url", "")
                    } else if (name == "versions.json") {
                        versionsManifestUrl = asset.optString("browser_download_url", "")
                    }
                }
            }

            // Check versions.json manifest if available (Solution A)
            var targetAndroidVersion = ""
            var manifestFound = false

            if (versionsManifestUrl.isNotBlank()) {
                val manifestReq = Request.Builder()
                    .url(versionsManifestUrl)
                    .header("User-Agent", "Kawaii-Wify-Android/${BuildConfig.VERSION_NAME}")
                    .build()
                val manifestResp = client.newCall(manifestReq).execute()
                if (manifestResp.isSuccessful) {
                    val mBody = manifestResp.body?.string()
                    if (!mBody.isNullOrBlank()) {
                        manifestFound = true
                        val mJson = JSONObject(mBody)
                        val manifestAndroid = mJson.optString("android", "").trim().removePrefix("v")
                        if (manifestAndroid.isNotBlank()) {
                            targetAndroidVersion = manifestAndroid
                        }
                        val manifestApkName = mJson.optString("android_apk_name", "").trim()
                        if (manifestApkName.isNotBlank() && assets != null) {
                            for (i in 0 until assets.length()) {
                                val asset = assets.getJSONObject(i)
                                if (asset.optString("name", "") == manifestApkName) {
                                    apkUrl = asset.optString("browser_download_url", "")
                                    break
                                }
                            }
                        }
                    }
                }
            }

            // Fallback: If no versions.json was attached to the release, fall back to release tag_name
            if (!manifestFound) {
                targetAndroidVersion = tagName
            }

            val isNewer = targetAndroidVersion.isNotBlank() && isNewerSemver(targetAndroidVersion, BuildConfig.VERSION_NAME)
            prefManager.updateLastUpdateCheckTime(now)

            if (isNewer && apkUrl.isNotBlank()) {
                val info = UpdateInfo(
                    hasUpdate = true,
                    latestVersion = targetAndroidVersion,
                    apkDownloadUrl = apkUrl,
                    releasePageUrl = htmlUrl,
                    releaseNotes = bodyText
                )
                _updateState.value = info
                return@withContext info
            } else {
                val info = UpdateInfo(hasUpdate = false)
                _updateState.value = info
                return@withContext info
            }
        } catch (e: Exception) {
            return@withContext _updateState.value
        }
    }

    suspend fun shouldNotifyUpdate(info: UpdateInfo): Boolean {
        if (!info.hasUpdate || info.latestVersion.isBlank()) return false
        val lastNotified = prefManager.getLastNotifiedUpdateVersion()
        return lastNotified != info.latestVersion
    }

    suspend fun markUpdateNotified(version: String) {
        prefManager.updateLastNotifiedUpdateVersion(version)
    }

    private fun isNewerSemver(remote: String, local: String): Boolean {
        fun cleanPart(p: String): String {
            val idx = p.indexOfAny(charArrayOf('-', '+'))
            return if (idx != -1) p.substring(0, idx) else p
        }

        val rParts = remote.trim().removePrefix("v").split(".")
        val lParts = local.trim().removePrefix("v").split(".")
        val maxLen = maxOf(rParts.size, lParts.size)

        for (i in 0 until maxLen) {
            val rVal = rParts.getOrNull(i)?.let { cleanPart(it).toIntOrNull() } ?: 0
            val lVal = lParts.getOrNull(i)?.let { cleanPart(it).toIntOrNull() } ?: 0
            if (rVal > lVal) return true
            if (rVal < lVal) return false
        }
        return false
    }
}
