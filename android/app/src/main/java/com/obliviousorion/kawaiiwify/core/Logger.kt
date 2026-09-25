package com.obliviousorion.kawaiiwify.core

import android.content.Context
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import java.io.File
import java.text.SimpleDateFormat
import java.util.Date
import java.util.Locale

data class LogEntry(
    val timestamp: String,
    val tag: String,
    val message: String,
    val level: LogLevel = LogLevel.INFO
) {
    override fun toString(): String = "[$timestamp] [$tag] $message"
}

enum class LogLevel {
    INFO, WARN, ERROR, SUCCESS
}

object Logger {
    private const val MAX_LOGS = 500
    private const val MAX_FILE_SIZE_BYTES = 512 * 1024L // 512 KB cap on disk
    private val timeFormat = SimpleDateFormat("HH:mm:ss", Locale.getDefault())

    private val _logs = MutableStateFlow<List<LogEntry>>(emptyList())
    val logs: StateFlow<List<LogEntry>> = _logs.asStateFlow()

    private var logFile: File? = null

    fun init(context: Context) {
        logFile = File(context.filesDir, "daemon.log")
        log("BOOT", "Kawaii-Wify logging engine initialized (◕‿◕)✌", LogLevel.SUCCESS)
    }

    @Synchronized
    fun log(tag: String, message: String, level: LogLevel = LogLevel.INFO) {
        val entry = LogEntry(
            timestamp = timeFormat.format(Date()),
            tag = tag.uppercase(),
            message = message,
            level = level
        )

        // Memory buffer is strictly capped at MAX_LOGS (500 items, ~40 KB)
        val updated = (_logs.value + entry).takeLast(MAX_LOGS)
        _logs.value = updated

        // Safely write to disk with file rotation to prevent disk bloat
        try {
            logFile?.let { file ->
                if (file.exists() && file.length() > MAX_FILE_SIZE_BYTES) {
                    val lines = file.readLines()
                    val retained = lines.takeLast(MAX_LOGS)
                    file.writeText(retained.joinToString("\n") + "\n")
                }
                file.appendText("${entry}\n")
            }
        } catch (_: Exception) {}
    }

    fun clear() {
        _logs.value = emptyList()
        try {
            logFile?.writeText("")
        } catch (_: Exception) {}
    }

    fun getLogFile(): File? = logFile
}
