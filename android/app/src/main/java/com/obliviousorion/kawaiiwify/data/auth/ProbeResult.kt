package com.obliviousorion.kawaiiwify.data.auth

sealed interface ProbeResult {
    data object Online : ProbeResult
    data class Captive(val magicToken: String) : ProbeResult
    data class Offline(val reason: String) : ProbeResult
}
