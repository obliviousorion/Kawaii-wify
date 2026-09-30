package com.obliviousorion.kawaiiwify.ui.screens

import android.content.ClipData
import android.content.ClipboardManager
import android.content.Context
import android.content.Intent
import android.net.Uri
import android.widget.Toast
import androidx.compose.animation.core.*
import androidx.compose.foundation.BorderStroke
import androidx.compose.foundation.Image
import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.verticalScroll
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.*
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.graphicsLayer
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.res.painterResource
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.obliviousorion.kawaiiwify.BuildConfig
import com.obliviousorion.kawaiiwify.KawaiiApplication
import com.obliviousorion.kawaiiwify.R
import com.obliviousorion.kawaiiwify.domain.EngineState
import com.obliviousorion.kawaiiwify.domain.SessionEngine
import com.obliviousorion.kawaiiwify.ui.components.GlowingButton
import com.obliviousorion.kawaiiwify.ui.components.MascotBanner
import com.obliviousorion.kawaiiwify.ui.components.TelemetryCard
import com.obliviousorion.kawaiiwify.ui.theme.*

@Composable
fun DashboardScreen(
    engine: SessionEngine,
    openUpdateDialog: Boolean = false,
    onDismissUpdateDialog: () -> Unit = {},
    onConnectClick: () -> Unit,
    onDisconnectClick: () -> Unit,
    onPauseClick: () -> Unit,
    onResumeClick: () -> Unit,
    modifier: Modifier = Modifier
) {
    val telemetry by engine.telemetry.collectAsState()
    val context = LocalContext.current

    val updateManager = remember { KawaiiApplication.instance.updateManager }
    val updateInfo by updateManager.updateState.collectAsState()
    var showDialog by remember { mutableStateOf(openUpdateDialog) }

    LaunchedEffect(openUpdateDialog) {
        if (openUpdateDialog) {
            showDialog = true
        }
    }

    LaunchedEffect(Unit) {
        updateManager.checkForUpdates(force = false)
    }

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
        // Top Header Row with Subtle Pulsing Update Download Icon
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .padding(horizontal = 4.dp, vertical = 2.dp),
            horizontalArrangement = Arrangement.SpaceBetween,
            verticalAlignment = Alignment.CenterVertically
        ) {
            Row(verticalAlignment = Alignment.CenterVertically) {
                Text(
                    text = "KAWAII-WIFY",
                    style = Typography.titleMedium.copy(
                        fontWeight = FontWeight.Black,
                        letterSpacing = 2.sp
                    ),
                    color = SakuraPink
                )
                Spacer(modifier = Modifier.width(6.dp))
                Text(
                    text = "v${BuildConfig.VERSION_NAME}",
                    style = Typography.labelSmall.copy(fontWeight = FontWeight.Bold),
                    color = TextSecondary
                )
            }

            if (updateInfo.hasUpdate) {
                val infiniteTransition = rememberInfiniteTransition(label = "pulse")
                val alpha by infiniteTransition.animateFloat(
                    initialValue = 0.35f,
                    targetValue = 1f,
                    animationSpec = infiniteRepeatable(
                        animation = tween(1000, easing = FastOutSlowInEasing),
                        repeatMode = RepeatMode.Reverse
                    ),
                    label = "updatePulseAlpha"
                )
                val scale by infiniteTransition.animateFloat(
                    initialValue = 0.95f,
                    targetValue = 1.05f,
                    animationSpec = infiniteRepeatable(
                        animation = tween(1000, easing = FastOutSlowInEasing),
                        repeatMode = RepeatMode.Reverse
                    ),
                    label = "updatePulseScale"
                )

                Box(
                    modifier = Modifier
                        .size(38.dp)
                        .graphicsLayer(scaleX = scale, scaleY = scale)
                        .background(NeonLavender.copy(alpha = 0.2f * alpha), CircleShape)
                        .border(1.5.dp, NeonLavender.copy(alpha = alpha), CircleShape)
                        .clickable { showDialog = true },
                    contentAlignment = Alignment.Center
                ) {
                    Icon(
                        imageVector = Icons.Default.Download,
                        contentDescription = "New version v${updateInfo.latestVersion} available",
                        tint = NeonLavender,
                        modifier = Modifier.size(20.dp)
                    )
                }
            }
        }

        // Cyber-Glass Update Dialog
        if (showDialog && updateInfo.hasUpdate) {
            AlertDialog(
                onDismissRequest = {
                    showDialog = false
                    onDismissUpdateDialog()
                },
                containerColor = SurfaceDark,
                titleContentColor = TextPrimary,
                shape = RoundedCornerShape(24.dp),
                modifier = Modifier.border(1.dp, SurfaceBorder, RoundedCornerShape(24.dp)),
                title = {
                    Row(
                        verticalAlignment = Alignment.CenterVertically,
                        modifier = Modifier.fillMaxWidth()
                    ) {
                        Image(
                            painter = painterResource(id = R.drawable.wify_mascot_update),
                            contentDescription = "Rocket Wify-Chan",
                            contentScale = ContentScale.Crop,
                            modifier = Modifier
                                .size(48.dp)
                                .clip(CircleShape)
                                .border(1.5.dp, NeonLavender, CircleShape)
                        )
                        Spacer(modifier = Modifier.width(12.dp))
                        Column {
                            Text(
                                text = "UPDATE AVAILABLE",
                                style = Typography.titleMedium.copy(
                                    fontWeight = FontWeight.Black,
                                    letterSpacing = 1.sp
                                ),
                                color = NeonLavender
                            )
                            Text(
                                text = "v${BuildConfig.VERSION_NAME} -> v${updateInfo.latestVersion}",
                                style = Typography.labelSmall.copy(fontWeight = FontWeight.Bold),
                                color = MintGreen
                            )
                        }
                    }
                },
                text = {
                    Column(
                        modifier = Modifier.fillMaxWidth(),
                        verticalArrangement = Arrangement.spacedBy(10.dp)
                    ) {
                        Text(
                            text = "A new release of Kawaii-Wify is ready with security improvements and connectivity stability. (ᴗ˳ᴗ)",
                            style = Typography.bodyMedium,
                            color = TextPrimary
                        )
                        if (updateInfo.releaseNotes.isNotBlank()) {
                            Surface(
                                color = SurfaceGlass,
                                shape = RoundedCornerShape(12.dp),
                                border = BorderStroke(1.dp, SurfaceBorder),
                                modifier = Modifier.fillMaxWidth()
                            ) {
                                Text(
                                    text = updateInfo.releaseNotes.take(280) + if (updateInfo.releaseNotes.length > 280) "..." else "",
                                    style = Typography.bodySmall.copy(fontFamily = androidx.compose.ui.text.font.FontFamily.Monospace),
                                    color = TextSecondary,
                                    modifier = Modifier.padding(10.dp)
                                )
                            }
                        }
                    }
                },
                confirmButton = {
                    Button(
                        onClick = {
                            try {
                                val intent = Intent(Intent.ACTION_VIEW, Uri.parse(updateInfo.apkDownloadUrl))
                                context.startActivity(intent)
                            } catch (e: Exception) {
                                Toast.makeText(context, "Could not open browser for download", Toast.LENGTH_SHORT).show()
                            }
                        },
                        colors = ButtonDefaults.buttonColors(containerColor = NeonLavender),
                        shape = RoundedCornerShape(12.dp)
                    ) {
                        Icon(
                            imageVector = Icons.Default.Download,
                            contentDescription = null,
                            modifier = Modifier.size(16.dp),
                            tint = SurfaceDark
                        )
                        Spacer(modifier = Modifier.width(6.dp))
                        Text("Download APK", color = SurfaceDark, fontWeight = FontWeight.Bold)
                    }
                },
                dismissButton = {
                    Row(horizontalArrangement = Arrangement.spacedBy(4.dp)) {
                        TextButton(
                            onClick = {
                                showDialog = false
                                onDismissUpdateDialog()
                            }
                        ) {
                            Text("Later", color = TextSecondary)
                        }
                        if (updateInfo.releasePageUrl.isNotBlank()) {
                            TextButton(
                                onClick = {
                                    try {
                                        val intent = Intent(Intent.ACTION_VIEW, Uri.parse(updateInfo.releasePageUrl))
                                        context.startActivity(intent)
                                    } catch (e: Exception) {
                                        Toast.makeText(context, "Could not open browser", Toast.LENGTH_SHORT).show()
                                    }
                                }
                            ) {
                                Text("Changelog", color = CyberCyan)
                            }
                        }
                    }
                }
            )
        }

        // Hero Mascot Banner (Modular slot)
        MascotBanner(state = telemetry.state)

        // Security Alert Banner
        if (telemetry.state is EngineState.SecurityHalted) {
            val reason = (telemetry.state as EngineState.SecurityHalted).reason
            Surface(
                modifier = Modifier.fillMaxWidth(),
                shape = RoundedCornerShape(16.dp),
                color = CrimsonRed.copy(alpha = 0.15f),
                border = BorderStroke(1.dp, CrimsonRed.copy(alpha = 0.8f))
            ) {
                Column(modifier = Modifier.padding(16.dp), verticalArrangement = Arrangement.spacedBy(8.dp)) {
                    Row(verticalAlignment = Alignment.CenterVertically) {
                        Icon(
                            imageVector = Icons.Default.Warning,
                            contentDescription = null,
                            tint = CrimsonRed,
                            modifier = Modifier.size(24.dp)
                        )
                        Spacer(modifier = Modifier.width(8.dp))
                        Text(
                            text = "SECURITY ALERT: Connection Halted",
                            style = Typography.titleMedium.copy(fontWeight = FontWeight.Bold),
                            color = CrimsonRed
                        )
                    }
                    Text(
                        text = reason,
                        style = Typography.bodySmall,
                        color = TextPrimary
                    )
                    Text(
                        text = "Authentication was suspended to protect credentials. If campus IT changed certificates, clear pins in Settings.",
                        style = Typography.labelSmall,
                        color = TextSecondary
                    )
                    Button(
                        onClick = onConnectClick,
                        colors = ButtonDefaults.buttonColors(containerColor = CrimsonRed),
                        modifier = Modifier.align(Alignment.End),
                        shape = RoundedCornerShape(12.dp)
                    ) {
                        Text("Reset & Retry", color = androidx.compose.ui.graphics.Color.White)
                    }
                }
            }
        } else if (telemetry.state is EngineState.BlockedByWhitelist) {
            val ssid = (telemetry.state as EngineState.BlockedByWhitelist).currentSsid ?: "Unknown"
            Surface(
                modifier = Modifier.fillMaxWidth(),
                shape = RoundedCornerShape(16.dp),
                color = WarningAmber.copy(alpha = 0.15f),
                border = BorderStroke(1.dp, WarningAmber.copy(alpha = 0.8f))
            ) {
                Column(modifier = Modifier.padding(16.dp), verticalArrangement = Arrangement.spacedBy(6.dp)) {
                    Row(verticalAlignment = Alignment.CenterVertically) {
                        Icon(
                            imageVector = Icons.Default.Lock,
                            contentDescription = null,
                            tint = WarningAmber,
                            modifier = Modifier.size(20.dp)
                        )
                        Spacer(modifier = Modifier.width(8.dp))
                        Text(
                            text = "Wi-Fi Whitelist Active",
                            style = Typography.titleSmall.copy(fontWeight = FontWeight.Bold),
                            color = WarningAmber
                        )
                    }
                    Text(
                        text = "Connected to '$ssid', which is not in permitted Wi-Fi networks. Auto-login paused.",
                        style = Typography.bodySmall,
                        color = TextPrimary
                    )
                }
            }
        }

        // Main Action & Secondary Controls Container
        Column(verticalArrangement = Arrangement.spacedBy(10.dp)) {
            // Main Glowing Button
            GlowingButton(
                state = telemetry.state,
                onClick = {
                    when (telemetry.state) {
                        is EngineState.Online -> onDisconnectClick()
                        is EngineState.Paused -> onResumeClick()
                        else -> onConnectClick()
                    }
                }
            )

            // Secondary Action Row (Pause or Logout)
            when (telemetry.state) {
                is EngineState.Online -> {
                    OutlinedButton(
                        onClick = onPauseClick,
                        modifier = Modifier
                            .fillMaxWidth()
                            .height(46.dp),
                        shape = RoundedCornerShape(23.dp),
                        colors = ButtonDefaults.outlinedButtonColors(contentColor = NeonLavender),
                        border = BorderStroke(1.dp, NeonLavender.copy(alpha = 0.5f))
                    ) {
                        Icon(
                            imageVector = Icons.Default.Pause,
                            contentDescription = null,
                            tint = NeonLavender,
                            modifier = Modifier.size(18.dp)
                        )
                        Spacer(modifier = Modifier.width(8.dp))
                        Text(
                            text = "Pause Daemon",
                            style = Typography.bodyMedium.copy(fontWeight = FontWeight.SemiBold)
                        )
                    }
                }

                is EngineState.Paused -> {
                    OutlinedButton(
                        onClick = onDisconnectClick,
                        modifier = Modifier
                            .fillMaxWidth()
                            .height(46.dp),
                        shape = RoundedCornerShape(23.dp),
                        colors = ButtonDefaults.outlinedButtonColors(contentColor = CrimsonRed),
                        border = BorderStroke(1.dp, CrimsonRed.copy(alpha = 0.5f))
                    ) {
                        Icon(
                            imageVector = Icons.Default.PowerSettingsNew,
                            contentDescription = null,
                            tint = CrimsonRed,
                            modifier = Modifier.size(18.dp)
                        )
                        Spacer(modifier = Modifier.width(8.dp))
                        Text(
                            text = "Disconnect",
                            style = Typography.bodyMedium.copy(fontWeight = FontWeight.SemiBold)
                        )
                    }
                }

                else -> {}
            }
        }

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
                icon = Icons.Default.Timer,
                modifier = Modifier.weight(1f),
                accentColor = CyberCyan
            )
            TelemetryCard(
                title = "Gateway Latency",
                value = if (telemetry.latencyMs > 0) "${telemetry.latencyMs} ms" else "--",
                icon = Icons.Default.Bolt,
                modifier = Modifier.weight(1f),
                accentColor = if (telemetry.latencyMs < 50) MintGreen else AlertOrange
            )
        }

        Row(
            modifier = Modifier.fillMaxWidth(),
            horizontalArrangement = Arrangement.spacedBy(12.dp)
        ) {
            val activeSsidDisplay: String = when {
                !telemetry.activeSsid.isNullOrBlank() -> telemetry.activeSsid!!
                telemetry.state is EngineState.Online || telemetry.state is EngineState.Captive -> "Connected (Wi-Fi)"
                else -> "Not Connected"
            }

            TelemetryCard(
                title = "Active SSID",
                value = activeSsidDisplay,
                icon = Icons.Default.Wifi,
                modifier = Modifier.weight(1f),
                accentColor = if (activeSsidDisplay != "Not Connected") NeonLavender else TextMuted
            )
            TelemetryCard(
                title = "Account ID",
                value = telemetry.username.ifEmpty { "None" },
                icon = Icons.Default.AccountCircle,
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
                icon = Icons.Default.Schedule,
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
                icon = Icons.Default.VpnKey,
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

        // Real-Time App Device Resource Footprint Section
        Text(
            text = "APP RESOURCE FOOTPRINT",
            style = Typography.labelSmall.copy(color = TextSecondary)
        )

        Row(
            modifier = Modifier.fillMaxWidth(),
            horizontalArrangement = Arrangement.spacedBy(12.dp)
        ) {
            TelemetryCard(
                title = "RAM Heap Usage",
                value = telemetry.memoryUsageMb,
                icon = Icons.Default.Memory,
                modifier = Modifier.weight(1f),
                accentColor = MintGreen
            )
            TelemetryCard(
                title = "CPU Thread State",
                value = telemetry.cpuStatus,
                icon = Icons.Default.Speed,
                modifier = Modifier.weight(1f),
                accentColor = CyberCyan
            )
        }
    }
}
