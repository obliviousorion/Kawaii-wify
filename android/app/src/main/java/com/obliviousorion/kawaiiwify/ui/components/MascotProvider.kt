package com.obliviousorion.kawaiiwify.ui.components

import androidx.annotation.DrawableRes
import androidx.compose.ui.graphics.Color
import com.obliviousorion.kawaiiwify.R
import com.obliviousorion.kawaiiwify.domain.EngineState
import com.obliviousorion.kawaiiwify.ui.theme.*

data class MascotVisualState(
    @DrawableRes val drawableRes: Int,
    val imageUrl: String? = null,
    val statusTitle: String,
    val statusQuote: String,
    val auraColor: Color,
    val badgeText: String,
    val badgeColor: Color
)

interface MascotThemeProvider {
    fun resolve(state: EngineState): MascotVisualState
}

object DefaultWifyChanProvider : MascotThemeProvider {
    override fun resolve(state: EngineState): MascotVisualState {
        return when (state) {
            is EngineState.Online -> MascotVisualState(
                drawableRes = R.drawable.wify_mascot_online,
                statusTitle = "ONLINE & SECURED",
                statusQuote = "\"Connection active! You're unblocked and online, senpai! (◕‿◕)✌\"",
                auraColor = CyberCyan,
                badgeText = "ONLINE",
                badgeColor = CyberCyan
            )

            is EngineState.Captive -> MascotVisualState(
                drawableRes = R.drawable.wify_mascot_online,
                statusTitle = "CAPTIVE PORTAL DETECTED",
                statusQuote = "\"FortiGate challenge intercepted! Priming session token...\"",
                auraColor = AlertOrange,
                badgeText = "INTERCEPTED",
                badgeColor = AlertOrange
            )

            is EngineState.Cooldown -> MascotVisualState(
                drawableRes = R.drawable.wify_mascot_offline,
                statusTitle = "CIRCUIT BREAKER COOLDOWN",
                statusQuote = "\"Whoa, senpai! Auth failed 3 times! Cooling down (${state.remainingSeconds}s) so your account won't get locked!\"",
                auraColor = CrimsonRed,
                badgeText = "COOLDOWN",
                badgeColor = CrimsonRed
            )

            is EngineState.Probing -> MascotVisualState(
                drawableRes = R.drawable.wify_mascot_online,
                statusTitle = "PROBING GATEWAY",
                statusQuote = "\"Testing connectivity to BITS-Pilani gateway...\"",
                auraColor = SakuraPink,
                badgeText = "PROBING",
                badgeColor = SakuraPink
            )

            is EngineState.Offline -> MascotVisualState(
                drawableRes = R.drawable.wify_mascot_offline,
                statusTitle = "DAEMON SLEEPING",
                statusQuote = "\"Zzz... Engine is resting. Tap CONNECT to wake me up! (ᴗ˳ᴗ)\"",
                auraColor = NeonLavender,
                badgeText = "OFFLINE",
                badgeColor = NeonLavender
            )
        }
    }
}

object MascotRegistry {
    var currentProvider: MascotThemeProvider = DefaultWifyChanProvider
}
