package com.obliviousorion.kawaiiwify.domain

data class Telemetry(
    val state: EngineState = EngineState.Offline,
    val isPaused: Boolean = false,
    val username: String = "",
    val gateway: String = "",
    val uptimeSeconds: Long = 0L,
    val lastProbeTime: String = "--:--:--",
    val latencyMs: Long = 0L,
    val sessionToken: String = "",
    val authFailures: Int = 0,
    val activeSsid: String? = null
)
