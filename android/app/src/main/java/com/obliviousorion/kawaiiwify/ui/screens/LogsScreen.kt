package com.obliviousorion.kawaiiwify.ui.screens

import android.content.Intent
import androidx.compose.foundation.layout.*
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Delete
import androidx.compose.material.icons.filled.Share
import androidx.compose.material3.*
import androidx.compose.runtime.Composable
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.unit.dp
import com.obliviousorion.kawaiiwify.core.Logger
import com.obliviousorion.kawaiiwify.ui.components.ConsoleLogViewer
import com.obliviousorion.kawaiiwify.ui.theme.*

@Composable
fun LogsScreen(
    modifier: Modifier = Modifier
) {
    val logs by Logger.logs.collectAsState()
    val context = LocalContext.current

    Column(
        modifier = modifier
            .fillMaxSize()
            .padding(16.dp)
    ) {
        // Header with Actions
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .padding(bottom = 12.dp),
            horizontalArrangement = Arrangement.SpaceBetween,
            verticalAlignment = Alignment.CenterVertically
        ) {
            Column {
                Text(
                    text = "CYBERPUNK CONSOLE",
                    style = Typography.titleLarge.copy(color = SakuraPink)
                )
                Text(
                    text = "Live daemon telemetry & audit trail",
                    style = Typography.bodyMedium.copy(color = TextSecondary)
                )
            }

            Row(horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                IconButton(
                    onClick = {
                        val allLogsText = logs.joinToString("\n") { it.toString() }
                        val sendIntent = Intent().apply {
                            action = Intent.ACTION_SEND
                            putExtra(Intent.EXTRA_TEXT, allLogsText)
                            type = "text/plain"
                        }
                        context.startActivity(Intent.createChooser(sendIntent, "Export Daemon Logs"))
                    }
                ) {
                    Icon(
                        imageVector = Icons.Default.Share,
                        contentDescription = "Share Logs",
                        tint = CyberCyan
                    )
                }

                IconButton(
                    onClick = { Logger.clear() }
                ) {
                    Icon(
                        imageVector = Icons.Default.Delete,
                        contentDescription = "Clear Logs",
                        tint = CrimsonRed
                    )
                }
            }
        }

        // Live Monospace Terminal
        ConsoleLogViewer(
            logs = logs,
            modifier = Modifier.weight(1f)
        )
    }
}
