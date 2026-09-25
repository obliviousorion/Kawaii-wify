package com.obliviousorion.kawaiiwify.ui.components

import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.horizontalScroll
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.lazy.rememberLazyListState
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontFamily
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.obliviousorion.kawaiiwify.core.LogEntry
import com.obliviousorion.kawaiiwify.core.LogLevel
import com.obliviousorion.kawaiiwify.ui.theme.*

@Composable
fun ConsoleLogViewer(
    logs: List<LogEntry>,
    modifier: Modifier = Modifier,
    onClearLogs: () -> Unit = {}
) {
    var selectedTag by remember { mutableStateOf("ALL") }
    val tags = listOf("ALL", "AUTH", "NET", "STATE", "SECURITY", "BOOT", "WARN", "ERROR")
    val listState = rememberLazyListState()

    val filteredLogs = remember(logs, selectedTag) {
        if (selectedTag == "ALL") logs
        else logs.filter { it.tag.equals(selectedTag, ignoreCase = true) }
    }

    LaunchedEffect(filteredLogs.size) {
        if (filteredLogs.isNotEmpty()) {
            listState.animateScrollToItem(filteredLogs.size - 1)
        }
    }

    Column(
        modifier = modifier
            .fillMaxSize()
            .clip(RoundedCornerShape(18.dp))
            .background(Color(0xFF07080C))
            .border(1.dp, SurfaceBorder, RoundedCornerShape(18.dp))
            .padding(12.dp)
    ) {
        // Tag filter chips row
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .horizontalScroll(rememberScrollState()),
            horizontalArrangement = Arrangement.spacedBy(8.dp),
            verticalAlignment = Alignment.CenterVertically
        ) {
            tags.forEach { tag ->
                val isSelected = selectedTag == tag
                val chipColor = when (tag) {
                    "AUTH" -> SakuraPink
                    "NET" -> CyberCyan
                    "STATE" -> MintGreen
                    "ERROR" -> CrimsonRed
                    "WARN" -> AlertOrange
                    else -> NeonLavender
                }

                FilterChip(
                    selected = isSelected,
                    onClick = { selectedTag = tag },
                    label = {
                        Text(
                            text = tag,
                            style = Typography.labelSmall.copy(
                                fontSize = 10.sp,
                                color = if (isSelected) BgDark else chipColor
                            )
                        )
                    },
                    colors = FilterChipDefaults.filterChipColors(
                        selectedContainerColor = chipColor,
                        containerColor = SurfaceDark
                    ),
                    border = FilterChipDefaults.filterChipBorder(
                        borderColor = chipColor.copy(alpha = 0.4f),
                        selectedBorderColor = chipColor,
                        enabled = true,
                        selected = isSelected
                    )
                )
            }
        }

        Spacer(modifier = Modifier.height(10.dp))

        // Terminal Log Output
        LazyColumn(
            state = listState,
            modifier = Modifier
                .weight(1f)
                .fillMaxWidth(),
            verticalArrangement = Arrangement.spacedBy(4.dp)
        ) {
            items(filteredLogs) { entry ->
                val tagColor = when (entry.level) {
                    LogLevel.SUCCESS -> MintGreen
                    LogLevel.WARN -> AlertOrange
                    LogLevel.ERROR -> CrimsonRed
                    LogLevel.INFO -> when (entry.tag) {
                        "AUTH" -> SakuraPink
                        "NET" -> CyberCyan
                        "STATE" -> NeonLavender
                        else -> TextSecondary
                    }
                }

                Row(modifier = Modifier.fillMaxWidth()) {
                    Text(
                        text = "${entry.timestamp} ",
                        fontFamily = FontFamily.Monospace,
                        fontSize = 11.sp,
                        color = TextMuted
                    )
                    Text(
                        text = "[${entry.tag}] ",
                        fontFamily = FontFamily.Monospace,
                        fontSize = 11.sp,
                        color = tagColor
                    )
                    Text(
                        text = entry.message,
                        fontFamily = FontFamily.Monospace,
                        fontSize = 11.sp,
                        color = TextPrimary
                    )
                }
            }
        }
    }
}
