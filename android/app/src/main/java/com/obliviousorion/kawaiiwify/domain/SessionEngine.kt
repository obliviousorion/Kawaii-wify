package com.obliviousorion.kawaiiwify.domain

import android.net.Network
import com.obliviousorion.kawaiiwify.core.Constants
import com.obliviousorion.kawaiiwify.core.LogLevel
import com.obliviousorion.kawaiiwify.core.Logger
import com.obliviousorion.kawaiiwify.data.auth.FortiGateAuth
import com.obliviousorion.kawaiiwify.data.auth.ProbeResult
import com.obliviousorion.kawaiiwify.data.local.AppConfig
import com.obliviousorion.kawaiiwify.data.local.Credentials
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.sync.Mutex
import kotlinx.coroutines.sync.withLock
import java.text.SimpleDateFormat
import java.util.Date
import java.util.Locale

class SessionEngine(
    private val auth: FortiGateAuth = FortiGateAuth()
) {
    private val mutex = Mutex()
    private val timeFormat = SimpleDateFormat("HH:mm:ss", Locale.getDefault())

    private var engineState: EngineState = EngineState.Offline
    private var sessionToken: String = ""
    private var failCount: Int = 0
    private var cooldownStartTime: Long = 0L
    private var isPaused: Boolean = false
    private val startTime: Long = System.currentTimeMillis()

    private val _telemetry = MutableStateFlow(Telemetry())
    val telemetry: StateFlow<Telemetry> = _telemetry.asStateFlow()

    private fun updateTelemetry(
        username: String = "",
        gateway: String = "",
        latency: Long = 0L,
        activeSsid: String? = null
    ) {
        val uptime = (System.currentTimeMillis() - startTime) / 1000
        _telemetry.value = Telemetry(
            state = engineState,
            isPaused = isPaused,
            username = username.ifEmpty { _telemetry.value.username },
            gateway = gateway.ifEmpty { _telemetry.value.gateway },
            uptimeSeconds = uptime,
            lastProbeTime = timeFormat.format(Date()),
            latencyMs = if (latency > 0) latency else _telemetry.value.latencyMs,
            sessionToken = sessionToken,
            authFailures = failCount,
            activeSsid = activeSsid ?: _telemetry.value.activeSsid
        )
    }

    suspend fun tick(
        network: Network?,
        config: AppConfig,
        credentials: Credentials?,
        activeSsid: String?
    ) = mutex.withLock {
        if (isPaused) {
            return@withLock
        }

        // Circuit breaker check
        if (inCooldownLocked()) {
            return@withLock
        }

        val startMs = System.currentTimeMillis()
        val probeRes = auth.probe(network, config.gateway)
        val latency = System.currentTimeMillis() - startMs

        when (probeRes) {
            is ProbeResult.Online -> {
                transitionLocked(EngineState.Online, "Probe OK")
                failCount = 0

                if (config.keepaliveEnabled && sessionToken.isNotEmpty()) {
                    val keepaliveRes = auth.keepalive(network, config.gateway, sessionToken)
                    if (keepaliveRes.isFailure) {
                        Logger.log("AUTH", "Keepalive ping failed: ${keepaliveRes.exceptionOrNull()?.message}", LogLevel.WARN)
                    }
                }
            }

            is ProbeResult.Offline -> {
                if (engineState != EngineState.Offline) {
                    Logger.log("NET", "Network unreachable: ${probeRes.reason}", LogLevel.ERROR)
                    transitionLocked(EngineState.Offline, "Unreachable")
                }
            }

            is ProbeResult.Captive -> {
                transitionLocked(EngineState.Captive(probeRes.magicToken), "CaptiveDetected")

                if (failCount >= Constants.MAX_AUTH_FAILURES) {
                    enterCooldownLocked()
                    return@withLock
                }

                if (credentials == null || credentials.username.isBlank()) {
                    Logger.log("AUTH", "No campus credentials saved! Please configure in Settings.", LogLevel.WARN)
                    return@withLock
                }

                // 1. Prime Challenge Session
                val primeRes = auth.prime(network, config.gateway, probeRes.magicToken)
                if (primeRes.isFailure) {
                    Logger.log("AUTH", "Gateway priming failed: ${primeRes.exceptionOrNull()?.message}", LogLevel.WARN)
                    return@withLock
                }

                // 2. Submit Credentials
                val loginRes = auth.login(
                    network = network,
                    gatewayHost = config.gateway,
                    challengeToken = probeRes.magicToken,
                    username = credentials.username,
                    password = credentials.password
                )

                if (loginRes.isSuccess) {
                    sessionToken = loginRes.getOrNull() ?: ""
                    failCount = 0
                    transitionLocked(EngineState.Online, "LoginSuccess")
                } else {
                    failCount++
                    Logger.log("AUTH", "Login failed (attempt $failCount/${Constants.MAX_AUTH_FAILURES})", LogLevel.ERROR)
                    if (failCount >= Constants.MAX_AUTH_FAILURES) {
                        enterCooldownLocked()
                    }
                }
            }
        }

        updateTelemetry(
            username = credentials?.username ?: "",
            gateway = config.gateway,
            latency = latency,
            activeSsid = activeSsid
        )
    }

    suspend fun connect(
        network: Network?,
        config: AppConfig,
        credentials: Credentials?,
        activeSsid: String?
    ) = mutex.withLock {
        isPaused = false
        failCount = 0
        if (engineState is EngineState.Cooldown) {
            transitionLocked(EngineState.Offline, "UserConnectReset")
        }
        Logger.log("STATE", "Engine manually activated (◕‿◕)✌", LogLevel.INFO)
    }

    suspend fun disconnect(
        network: Network?,
        config: AppConfig
    ) = mutex.withLock {
        val currentToken = sessionToken
        isPaused = true
        sessionToken = ""
        transitionLocked(EngineState.Offline, "UserPaused")

        if (currentToken.isNotEmpty()) {
            auth.logout(network, config.gateway, currentToken)
        }
        Logger.log("STATE", "Engine paused and offline by user request", LogLevel.WARN)
        updateTelemetry()
    }

    private fun transitionLocked(next: EngineState, reason: String) {
        if (engineState == next) return
        Logger.log("STATE", "Transition: ${engineState::class.simpleName} -> ${next::class.simpleName} ($reason)", LogLevel.INFO)
        engineState = next
        updateTelemetry()
    }

    private fun enterCooldownLocked() {
        cooldownStartTime = System.currentTimeMillis()
        isPaused = true
        transitionLocked(EngineState.Cooldown(Constants.COOLDOWN_DURATION_SECONDS), "MaxFailuresExceeded")
        Logger.log("WARN", "Max auth failures reached. Cooldown initiated to protect account!", LogLevel.WARN)
    }

    private fun inCooldownLocked(): Boolean {
        if (engineState !is EngineState.Cooldown) return false

        val elapsed = (System.currentTimeMillis() - cooldownStartTime) / 1000
        if (elapsed >= Constants.COOLDOWN_DURATION_SECONDS) {
            Logger.log("STATE", "Cooldown elapsed, resetting circuit breaker", LogLevel.INFO)
            failCount = 0
            isPaused = false
            transitionLocked(EngineState.Offline, "CooldownElapsed")
            return false
        }

        val remaining = (Constants.COOLDOWN_DURATION_SECONDS - elapsed).toInt()
        engineState = EngineState.Cooldown(remaining)
        updateTelemetry()
        return true
    }
}
