package com.obliviousorion.kawaiiwify.data.auth

import android.net.Network
import com.obliviousorion.kawaiiwify.core.Constants
import com.obliviousorion.kawaiiwify.core.LogLevel
import com.obliviousorion.kawaiiwify.core.Logger
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext
import okhttp3.FormBody
import okhttp3.OkHttpClient
import okhttp3.Request
import java.io.IOException
import java.util.concurrent.TimeUnit
import java.util.regex.Pattern

class FortiGateAuth {

    private val fgtAuthPattern = Pattern.compile("fgtauth\\?([a-f0-9]+)")
    private val keepalivePattern = Pattern.compile("keepalive\\?([a-f0-9]+)")

    private fun buildClient(network: Network?, gatewayHost: String): OkHttpClient {
        val trustManager = HostVerifier.createInsecureTrustManager()
        val sslSocketFactory = HostVerifier.createSSLSocketFactory(trustManager)
        val hostnameVerifier = HostVerifier.createHostnameVerifier(gatewayHost)

        val builder = OkHttpClient.Builder()
            .connectTimeout(8, TimeUnit.SECONDS)
            .readTimeout(8, TimeUnit.SECONDS)
            .writeTimeout(8, TimeUnit.SECONDS)
            .followRedirects(true)
            .sslSocketFactory(sslSocketFactory, trustManager)
            .hostnameVerifier(hostnameVerifier)

        // Bind directly to Network SocketFactory if provided (Solves Socket Routing Trap)
        if (network != null) {
            builder.socketFactory(network.socketFactory)
        }

        return builder.build()
    }

    suspend fun probe(network: Network?, gatewayHost: String = Constants.DEFAULT_GATEWAY): ProbeResult =
        withContext(Dispatchers.IO) {
            val client = buildClient(network, gatewayHost)
            val request = Request.Builder()
                .url(Constants.PROBE_URL)
                .header("User-Agent", "Mozilla/5.0")
                .get()
                .build()

            try {
                client.newCall(request).execute().use { response ->
                    if (response.code == 204) {
                        Logger.log("NET", "Probe returned HTTP 204 (Online)", LogLevel.SUCCESS)
                        return@withContext ProbeResult.Online
                    }

                    val body = response.body?.string() ?: ""
                    val matcher = fgtAuthPattern.matcher(body)
                    if (matcher.find()) {
                        val magicToken = matcher.group(1) ?: ""
                        Logger.log("NET", "Captive portal detected! Magic challenge: $magicToken", LogLevel.WARN)
                        return@withContext ProbeResult.Captive(magicToken)
                    }

                    Logger.log("NET", "Probe intercepted but no FortiGate challenge found", LogLevel.WARN)
                    return@withContext ProbeResult.Offline("No FortiGate magic challenge detected in response")
                }
            } catch (e: Exception) {
                Logger.log("NET", "Probe failed: ${e.message}", LogLevel.ERROR)
                return@withContext ProbeResult.Offline(e.message ?: "Network unreachable")
            }
        }

    suspend fun prime(network: Network?, gatewayHost: String, challengeToken: String): Result<Unit> =
        withContext(Dispatchers.IO) {
            val client = buildClient(network, gatewayHost)
            val primeUrl = "https://$gatewayHost/fgtauth?$challengeToken"

            val request = Request.Builder()
                .url(primeUrl)
                .header("User-Agent", "Mozilla/5.0")
                .get()
                .build()

            try {
                Logger.log("AUTH", "Priming challenge on FortiOS: $primeUrl", LogLevel.INFO)
                client.newCall(request).execute().use { response ->
                    if (!response.isSuccessful && response.code != 302) {
                        return@withContext Result.failure(IOException("Prime returned unexpected code: ${response.code}"))
                    }
                    return@withContext Result.success(Unit)
                }
            } catch (e: Exception) {
                Logger.log("AUTH", "Prime request failed: ${e.message}", LogLevel.ERROR)
                return@withContext Result.failure(e)
            }
        }

    suspend fun login(
        network: Network?,
        gatewayHost: String,
        challengeToken: String,
        username: String,
        password: String
    ): Result<String> = withContext(Dispatchers.IO) {
        val client = buildClient(network, gatewayHost)
        val loginUrl = "https://$gatewayHost/"
        val primeUrl = "https://$gatewayHost/fgtauth?$challengeToken"

        val formBody = FormBody.Builder()
            .add("4Tredir", Constants.PROBE_URL)
            .add("magic", challengeToken)
            .add("username", username)
            .add("password", password)
            .build()

        val request = Request.Builder()
            .url(loginUrl)
            .post(formBody)
            .header("Content-Type", "application/x-www-form-urlencoded")
            .header("User-Agent", "Mozilla/5.0")
            .header("Referer", primeUrl)
            .build()

        try {
            Logger.log("AUTH", "Submitting login credentials for user $username...", LogLevel.INFO)
            client.newCall(request).execute().use { response ->
                val body = response.body?.string() ?: ""
                val matcher = keepalivePattern.matcher(body)
                if (matcher.find()) {
                    val sessionToken = matcher.group(1) ?: ""
                    Logger.log("AUTH", "Login successful! Session token: $sessionToken", LogLevel.SUCCESS)
                    return@withContext Result.success(sessionToken)
                }

                Logger.log("AUTH", "Authentication rejected: invalid credentials or session declined", LogLevel.ERROR)
                return@withContext Result.failure(IOException("Invalid credentials or rejection by FortiGate"))
            }
        } catch (e: Exception) {
            Logger.log("AUTH", "Login request failed: ${e.message}", LogLevel.ERROR)
            return@withContext Result.failure(e)
        }
    }

    suspend fun keepalive(network: Network?, gatewayHost: String, sessionToken: String): Result<Unit> =
        withContext(Dispatchers.IO) {
            val client = buildClient(network, gatewayHost)
            val keepaliveUrl = "https://$gatewayHost/keepalive?$sessionToken"

            val request = Request.Builder()
                .url(keepaliveUrl)
                .header("User-Agent", "Mozilla/5.0")
                .get()
                .build()

            try {
                client.newCall(request).execute().use { response ->
                    if (response.code == 200) {
                        return@withContext Result.success(Unit)
                    }
                    Logger.log("AUTH", "Keepalive ping returned code ${response.code}", LogLevel.WARN)
                    return@withContext Result.failure(IOException("Keepalive returned HTTP ${response.code}"))
                }
            } catch (e: Exception) {
                Logger.log("AUTH", "Keepalive ping error: ${e.message}", LogLevel.ERROR)
                return@withContext Result.failure(e)
            }
        }

    suspend fun logout(network: Network?, gatewayHost: String, sessionToken: String): Result<Unit> =
        withContext(Dispatchers.IO) {
            val client = buildClient(network, gatewayHost)
            val logoutUrl = "https://$gatewayHost/logout?$sessionToken"

            val request = Request.Builder()
                .url(logoutUrl)
                .header("User-Agent", "Mozilla/5.0")
                .get()
                .build()

            try {
                Logger.log("AUTH", "Terminating session on FortiOS ($logoutUrl)", LogLevel.INFO)
                client.newCall(request).execute().use { response ->
                    if (response.code == 200) {
                        Logger.log("AUTH", "Session terminated successfully on gateway", LogLevel.SUCCESS)
                        return@withContext Result.success(Unit)
                    }
                    return@withContext Result.failure(IOException("Logout failed with code: ${response.code}"))
                }
            } catch (e: Exception) {
                Logger.log("AUTH", "Logout failed: ${e.message}", LogLevel.ERROR)
                return@withContext Result.failure(e)
            }
        }
}
