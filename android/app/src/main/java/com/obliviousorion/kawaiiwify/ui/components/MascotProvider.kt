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

object DefaultKawaiiWifyProvider : MascotThemeProvider {
    override fun resolve(state: EngineState): MascotVisualState {
        return when (state) {
            is EngineState.Online -> MascotVisualState(
                drawableRes = R.drawable.wify_mascot_online,
                statusTitle = "ONLINE & SECURED",
                statusQuote = "\"Connection active! Campus Wi-Fi secured and unblocked. (◕‿◕)✌\"",
                auraColor = CyberCyan,
                badgeText = "ONLINE",
                badgeColor = CyberCyan
            )

            is EngineState.Paused -> MascotVisualState(
                drawableRes = R.drawable.wify_mascot_offline,
                statusTitle = "DAEMON PAUSED",
                statusQuote = "\"Kawaii-Wify is paused. Your firewall session is untouched and active. (ᴗ˳ᴗ)\"",
                auraColor = NeonLavender,
                badgeText = "PAUSED",
                badgeColor = NeonLavender
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
                statusQuote = "\"Auth failed 3 times! Cooling down (${state.remainingSeconds}s) to protect account from lockout.\"",
                auraColor = CrimsonRed,
                badgeText = "COOLDOWN",
                badgeColor = CrimsonRed
            )

            is EngineState.Probing -> MascotVisualState(
                drawableRes = R.drawable.wify_mascot_online,
                statusTitle = "PROBING GATEWAY",
                statusQuote = "\"Testing connectivity to campus gateway...\"",
                auraColor = SakuraPink,
                badgeText = "PROBING",
                badgeColor = SakuraPink
            )

            is EngineState.Offline -> MascotVisualState(
                drawableRes = R.drawable.wify_mascot_offline,
                statusTitle = "DAEMON OFFLINE",
                statusQuote = "\"Kawaii-Wify is resting. Tap CONNECT to authenticate. (ᴗ˳ᴗ)\"",
                auraColor = NeonLavender,
                badgeText = "OFFLINE",
                badgeColor = NeonLavender
            )
        }
    }
}

object MascotRegistry {
    var currentProvider: MascotThemeProvider = DefaultKawaiiWifyProvider
}
