package com.obliviousorion.kawaiiwify.domain

sealed interface EngineState {
    data object Offline : EngineState
    data object Probing : EngineState
    data class Captive(val magicToken: String) : EngineState
    data object Online : EngineState
    data object Paused : EngineState
    data class Cooldown(val remainingSeconds: Int) : EngineState
    data class BlockedByWhitelist(val currentSsid: String?) : EngineState
    data class SecurityHalted(val reason: String) : EngineState
}
