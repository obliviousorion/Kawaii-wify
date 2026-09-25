package com.obliviousorion.kawaiiwify.ui.screens

import android.content.ClipData
import android.content.ClipboardManager
import android.content.Context
import android.widget.Toast
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.verticalScroll
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.ui.Modifier
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.unit.dp
import com.obliviousorion.kawaiiwify.domain.EngineState
import com.obliviousorion.kawaiiwify.domain.SessionEngine
import com.obliviousorion.kawaiiwify.ui.components.GlowingButton
import com.obliviousorion.kawaiiwify.ui.components.MascotBanner
import com.obliviousorion.kawaiiwify.ui.components.TelemetryCard
import com.obliviousorion.kawaiiwify.ui.theme.*

@Composable
fun DashboardScreen(
    engine: SessionEngine,
    onConnectClick: () -> Unit,
    onDisconnectClick: () -> Unit,
    modifier: Modifier = Modifier
) {
    val telemetry by engine.telemetry.collectAsState()
    val context = LocalContext.current

    val hours = telemetry.uptimeSeconds / 3600
    val mins = (telemetry.uptimeSeconds % 3600) / 60
    val secs = telemetry.uptimeSeconds % 60
    val formattedUptime = "%02dh %02dm %02ds".format(hours, mins, secs)

    Column(
        modifier = modifier
            .fillMaxSize()
            .verticalScroll(rememberScrollState())
            .padding(16.dp),
        verticalArrangement = Arrangement.spacedBy(16.dp)
    ) {
        // Hero Mascot Banner (Modular slot)
        MascotBanner(state = telemetry.state)

        // Glowing Action Button
        GlowingButton(
            state = telemetry.state,
            onClick = {
                if (telemetry.state is EngineState.Online) {
                    onDisconnectClick()
                } else {
                    onConnectClick()
                }
            }
        )

        // Telemetry Metrics Grid
        Text(
            text = "LIVE TELEMETRY",
            style = Typography.labelSmall.copy(color = TextSecondary)
        )

        Row(
            modifier = Modifier.fillMaxWidth(),
            horizontalArrangement = Arrangement.spacedBy(12.dp)
        ) {
            TelemetryCard(
                title = "Session Uptime",
                value = formattedUptime,
                icon = "⏱",
                modifier = Modifier.weight(1f),
                accentColor = CyberCyan
            )
            TelemetryCard(
                title = "Gateway Latency",
                value = if (telemetry.latencyMs > 0) "${telemetry.latencyMs} ms" else "--",
                icon = "⚡",
                modifier = Modifier.weight(1f),
                accentColor = if (telemetry.latencyMs < 50) MintGreen else AlertOrange
            )
        }

        Row(
            modifier = Modifier.fillMaxWidth(),
            horizontalArrangement = Arrangement.spacedBy(12.dp)
        ) {
            TelemetryCard(
                title = "Active SSID",
                value = telemetry.activeSsid ?: "Not Connected",
                icon = "📶",
                modifier = Modifier.weight(1f),
                accentColor = NeonLavender
            )
            TelemetryCard(
                title = "Account ID",
                value = telemetry.username.ifEmpty { "None" },
                icon = "🛡",
                modifier = Modifier.weight(1f),
                accentColor = SakuraPink
            )
        }

        Row(
            modifier = Modifier.fillMaxWidth(),
            horizontalArrangement = Arrangement.spacedBy(12.dp)
        ) {
            TelemetryCard(
                title = "Last Probe",
                value = telemetry.lastProbeTime,
                icon = "🔍",
                modifier = Modifier.weight(1f),
                accentColor = TextSecondary
            )
            TelemetryCard(
                title = "Session Token",
                value = if (telemetry.sessionToken.isNotEmpty()) {
                    "${telemetry.sessionToken.take(6)}... (tap)"
                } else {
                    "None"
                },
                icon = "🔑",
                modifier = Modifier.weight(1f),
                accentColor = CyberCyan,
                onCardClick = if (telemetry.sessionToken.isNotEmpty()) {
                    {
                        val clipboard = context.getSystemService(Context.CLIPBOARD_SERVICE) as ClipboardManager
                        clipboard.setPrimaryClip(ClipData.newPlainText("Session Token", telemetry.sessionToken))
                        Toast.makeText(context, "Session token copied!", Toast.LENGTH_SHORT).show()
                    }
                } else null
            )
        }
    }
}
