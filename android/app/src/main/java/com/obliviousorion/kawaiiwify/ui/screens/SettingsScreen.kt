package com.obliviousorion.kawaiiwify.ui.screens

import android.annotation.SuppressLint
import android.content.Context
import android.content.Intent
import android.net.Uri
import android.os.PowerManager
import android.provider.Settings
import android.widget.Toast
import androidx.compose.animation.AnimatedVisibility
import androidx.compose.foundation.BorderStroke
import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.horizontalScroll
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
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.text.font.FontFamily
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.input.PasswordVisualTransformation
import androidx.compose.ui.text.input.VisualTransformation
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.obliviousorion.kawaiiwify.KawaiiApplication
import com.obliviousorion.kawaiiwify.data.local.AppConfig
import com.obliviousorion.kawaiiwify.data.local.Credentials
import com.obliviousorion.kawaiiwify.ui.theme.*
import kotlinx.coroutines.launch

@SuppressLint("BatteryLife")
@Composable
fun SettingsScreen(
    modifier: Modifier = Modifier
) {
    val context = LocalContext.current
    val coroutineScope = rememberCoroutineScope()
    val securityManager = remember { KawaiiApplication.instance.securityManager }
    val prefManager = remember { KawaiiApplication.instance.preferencesManager }

    val savedCreds = remember { securityManager.getCredentials() }
    var username by remember { mutableStateOf(savedCreds?.username ?: "") }
    var password by remember { mutableStateOf(savedCreds?.password ?: "") }
    var passwordVisible by remember { mutableStateOf(false) }
    var showWipeConfirmDialog by remember { mutableStateOf(false) }
    var isVaultSaved by remember { mutableStateOf(securityManager.hasCredentials()) }

    val config by prefManager.configFlow.collectAsState(initial = AppConfig())
    var gateway by remember(config.gateway) { mutableStateOf(config.gateway) }
    var interval by remember(config.checkIntervalSeconds) { mutableFloatStateOf(config.checkIntervalSeconds.toFloat()) }
    var newSsidInput by remember { mutableStateOf("") }

    val powerManager = remember { context.getSystemService(Context.POWER_SERVICE) as PowerManager }
    var isBatteryExempt by remember {
        mutableStateOf(powerManager.isIgnoringBatteryOptimizations(context.packageName))
    }

    // Confirmation Dialog for Purging Vault
    if (showWipeConfirmDialog) {
        AlertDialog(
            onDismissRequest = { showWipeConfirmDialog = false },
            title = {
                Text("Purge Hardware KeyStore?", color = CrimsonRed, fontWeight = FontWeight.Bold)
            },
            text = {
                Text(
                    "This will permanently erase your campus credentials from Android's encrypted vault. You will need to log in again.",
                    color = TextPrimary
                )
            },
            confirmButton = {
                Button(
                    onClick = {
                        securityManager.clearCredentials()
                        username = ""
                        password = ""
                        isVaultSaved = false
                        showWipeConfirmDialog = false
                        Toast.makeText(context, "Credentials vault wiped!", Toast.LENGTH_SHORT).show()
                    },
                    colors = ButtonDefaults.buttonColors(containerColor = CrimsonRed)
                ) {
                    Text("Purge Vault", color = Color.White)
                }
            },
            dismissButton = {
                TextButton(onClick = { showWipeConfirmDialog = false }) {
                    Text("Cancel", color = TextSecondary)
                }
            },
            containerColor = SurfaceDark
        )
    }

    Column(
        modifier = modifier
            .fillMaxSize()
            .verticalScroll(rememberScrollState())
            .padding(16.dp),
        verticalArrangement = Arrangement.spacedBy(20.dp)
    ) {
        // Page Header
        Column(verticalArrangement = Arrangement.spacedBy(4.dp)) {
            Row(
                verticalAlignment = Alignment.CenterVertically,
                horizontalArrangement = Arrangement.spacedBy(8.dp)
            ) {
                Icon(
                    imageVector = Icons.Default.Settings,
                    contentDescription = null,
                    tint = SakuraPink,
                    modifier = Modifier.size(24.dp)
                )
                Text(
                    text = "SETTINGS & VAULT",
                    style = Typography.titleLarge.copy(color = TextPrimary, letterSpacing = 1.sp)
                )
            }
            Text(
                text = "Manage KeyStore credentials, gateway routing & Wi-Fi rules",
                style = Typography.bodyMedium.copy(color = TextSecondary)
            )
        }

        // ==========================================
        // 1. HARDWARE KEYSTORE VAULT SECTION
        // ==========================================
        Card(
            shape = RoundedCornerShape(20.dp),
            colors = CardDefaults.cardColors(containerColor = SurfaceDark),
            modifier = Modifier
                .fillMaxWidth()
                .border(1.dp, SurfaceBorder, RoundedCornerShape(20.dp))
        ) {
            Column(
                modifier = Modifier.padding(18.dp),
                verticalArrangement = Arrangement.spacedBy(14.dp)
            ) {
                // Section Title + Vault Status Badge
                Row(
                    modifier = Modifier.fillMaxWidth(),
                    horizontalArrangement = Arrangement.SpaceBetween,
                    verticalAlignment = Alignment.CenterVertically
                ) {
                    Row(
                        verticalAlignment = Alignment.CenterVertically,
                        horizontalArrangement = Arrangement.spacedBy(8.dp)
                    ) {
                        Icon(
                            imageVector = Icons.Default.Lock,
                            contentDescription = null,
                            tint = CyberCyan,
                            modifier = Modifier.size(18.dp)
                        )
                        Text(
                            text = "CREDENTIALS VAULT",
                            style = Typography.labelSmall.copy(color = CyberCyan, fontWeight = FontWeight.Bold)
                        )
                    }

                    // Dynamic Vault Status Badge
                    Box(
                        modifier = Modifier
                            .clip(RoundedCornerShape(50))
                            .background(if (isVaultSaved) MintGreen.copy(alpha = 0.15f) else AlertOrange.copy(alpha = 0.15f))
                            .border(
                                1.dp,
                                if (isVaultSaved) MintGreen.copy(alpha = 0.5f) else AlertOrange.copy(alpha = 0.5f),
                                RoundedCornerShape(50)
                            )
                            .padding(horizontal = 10.dp, vertical = 4.dp)
                    ) {
                        Text(
                            text = if (isVaultSaved) "AES-256 LOCKED" else "EMPTY",
                            style = Typography.labelSmall.copy(
                                color = if (isVaultSaved) MintGreen else AlertOrange,
                                fontSize = 10.sp,
                                fontWeight = FontWeight.Bold
                            )
                        )
                    }
                }

                Text(
                    text = "Stored locally inside Android Hardware KeyStore (AES-256 GCM). Never uploaded to any external server.",
                    style = Typography.bodyMedium.copy(color = TextMuted, fontSize = 12.sp)
                )

                // Username input
                OutlinedTextField(
                    value = username,
                    onValueChange = { username = it },
                    label = { Text("Campus ID / Username") },
                    leadingIcon = {
                        Icon(Icons.Default.Person, contentDescription = null, tint = NeonLavender)
                    },
                    singleLine = true,
                    modifier = Modifier.fillMaxWidth(),
                    shape = RoundedCornerShape(14.dp),
                    colors = OutlinedTextFieldDefaults.colors(
                        focusedBorderColor = SakuraPink,
                        unfocusedBorderColor = SurfaceBorder,
                        focusedContainerColor = SurfaceGlass,
                        unfocusedContainerColor = SurfaceGlass
                    )
                )

                // Password input
                OutlinedTextField(
                    value = password,
                    onValueChange = { password = it },
                    label = { Text("Campus Wi-Fi Password") },
                    leadingIcon = {
                        Icon(Icons.Default.Lock, contentDescription = null, tint = NeonLavender)
                    },
                    singleLine = true,
                    visualTransformation = if (passwordVisible) VisualTransformation.None else PasswordVisualTransformation(),
                    trailingIcon = {
                        IconButton(onClick = { passwordVisible = !passwordVisible }) {
                            Text(
                                text = if (passwordVisible) "HIDE" else "SHOW",
                                style = Typography.labelSmall.copy(color = CyberCyan, fontSize = 10.sp)
                            )
                        }
                    },
                    modifier = Modifier.fillMaxWidth(),
                    shape = RoundedCornerShape(14.dp),
                    colors = OutlinedTextFieldDefaults.colors(
                        focusedBorderColor = SakuraPink,
                        unfocusedBorderColor = SurfaceBorder,
                        focusedContainerColor = SurfaceGlass,
                        unfocusedContainerColor = SurfaceGlass
                    )
                )

                // Vault Action Buttons
                Row(
                    modifier = Modifier.fillMaxWidth(),
                    horizontalArrangement = Arrangement.spacedBy(10.dp)
                ) {
                    Button(
                        onClick = {
                            if (username.isNotBlank() && password.isNotBlank()) {
                                securityManager.saveCredentials(Credentials(username, password))
                                isVaultSaved = true
                                Toast.makeText(context, "Encrypted in Hardware KeyStore! (◕‿◕)✌", Toast.LENGTH_SHORT).show()
                            } else {
                                Toast.makeText(context, "Username and password cannot be empty", Toast.LENGTH_SHORT).show()
                            }
                        },
                        modifier = Modifier.weight(1f),
                        shape = RoundedCornerShape(12.dp),
                        colors = ButtonDefaults.buttonColors(containerColor = SakuraPink)
                    ) {
                        Icon(Icons.Default.Check, contentDescription = null, tint = BgDark, modifier = Modifier.size(16.dp))
                        Spacer(modifier = Modifier.width(6.dp))
                        Text("Save Vault", color = BgDark, fontWeight = FontWeight.Bold)
                    }

                    OutlinedButton(
                        onClick = { showWipeConfirmDialog = true },
                        modifier = Modifier.weight(0.9f),
                        shape = RoundedCornerShape(12.dp),
                        colors = ButtonDefaults.outlinedButtonColors(contentColor = CrimsonRed),
                        border = BorderStroke(1.dp, CrimsonRed)
                    ) {
                        Icon(Icons.Default.Delete, contentDescription = null, tint = CrimsonRed, modifier = Modifier.size(16.dp))
                        Spacer(modifier = Modifier.width(6.dp))
                        Text("Wipe Vault")
                    }
                }
            }
        }

        // ==========================================
        // 2. GATEWAY & ENGINE CONFIGURATION
        // ==========================================
        Card(
            shape = RoundedCornerShape(20.dp),
            colors = CardDefaults.cardColors(containerColor = SurfaceDark),
            modifier = Modifier
                .fillMaxWidth()
                .border(1.dp, SurfaceBorder, RoundedCornerShape(20.dp))
        ) {
            Column(
                modifier = Modifier.padding(18.dp),
                verticalArrangement = Arrangement.spacedBy(16.dp)
            ) {
                Row(
                    verticalAlignment = Alignment.CenterVertically,
                    horizontalArrangement = Arrangement.spacedBy(8.dp)
                ) {
                    Icon(
                        imageVector = Icons.Default.Refresh,
                        contentDescription = null,
                        tint = MintGreen,
                        modifier = Modifier.size(18.dp)
                    )
                    Text(
                        text = "DAEMON ENGINE CONFIG",
                        style = Typography.labelSmall.copy(color = MintGreen, fontWeight = FontWeight.Bold)
                    )
                }

                // Gateway Endpoint Field
                Column(verticalArrangement = Arrangement.spacedBy(6.dp)) {
                    Text("Gateway Endpoint", style = Typography.bodyMedium.copy(color = TextPrimary))
                    OutlinedTextField(
                        value = gateway,
                        onValueChange = {
                            gateway = it
                            coroutineScope.launch { prefManager.updateGateway(it) }
                        },
                        singleLine = true,
                        modifier = Modifier.fillMaxWidth(),
                        shape = RoundedCornerShape(14.dp),
                        colors = OutlinedTextFieldDefaults.colors(
                            focusedBorderColor = MintGreen,
                            unfocusedBorderColor = SurfaceBorder,
                            focusedContainerColor = SurfaceGlass,
                            unfocusedContainerColor = SurfaceGlass
                        )
                    )

                    // Quick Gateway Preset Chips (with 'Other' chip)
                    val standardGateways = listOf(
                        "fw.bits-pilani.ac.in:8090" to "BITS Campus",
                        "172.16.100.1:8090" to "Direct IP"
                    )
                    val isOtherGateway = gateway.isNotBlank() && gateway !in standardGateways.map { it.first }

                    Row(
                        modifier = Modifier
                            .fillMaxWidth()
                            .horizontalScroll(rememberScrollState()),
                        horizontalArrangement = Arrangement.spacedBy(8.dp)
                    ) {
                        standardGateways.forEach { (presetUrl, label) ->
                            val isSelected = gateway == presetUrl
                            SuggestionChip(
                                onClick = {
                                    gateway = presetUrl
                                    coroutineScope.launch { prefManager.updateGateway(presetUrl) }
                                },
                                label = { Text(label, style = Typography.labelSmall.copy(fontSize = 10.sp)) },
                                colors = SuggestionChipDefaults.suggestionChipColors(
                                    containerColor = if (isSelected) CyberCyan.copy(alpha = 0.2f) else SurfaceGlass,
                                    labelColor = if (isSelected) CyberCyan else TextSecondary
                                ),
                                border = SuggestionChipDefaults.suggestionChipBorder(
                                    borderColor = if (isSelected) CyberCyan else SurfaceBorder,
                                    enabled = true
                                )
                            )
                        }

                        // "Other" chip - automatically selected if gateway is custom
                        SuggestionChip(
                            onClick = {
                                if (!isOtherGateway) {
                                    gateway = ""
                                }
                            },
                            label = { Text("Other", style = Typography.labelSmall.copy(fontSize = 10.sp)) },
                            colors = SuggestionChipDefaults.suggestionChipColors(
                                containerColor = if (isOtherGateway) CyberCyan.copy(alpha = 0.2f) else SurfaceGlass,
                                labelColor = if (isOtherGateway) CyberCyan else TextSecondary
                            ),
                            border = SuggestionChipDefaults.suggestionChipBorder(
                                borderColor = if (isOtherGateway) CyberCyan else SurfaceBorder,
                                enabled = true
                            )
                        )
                    }
                }

                // Check Interval Slider with Presets
                Column(verticalArrangement = Arrangement.spacedBy(6.dp)) {
                    Row(
                        modifier = Modifier.fillMaxWidth(),
                        horizontalArrangement = Arrangement.SpaceBetween,
                        verticalAlignment = Alignment.CenterVertically
                    ) {
                        Text("Probe Interval", style = Typography.bodyMedium.copy(color = TextPrimary))
                        Text(
                            "${interval.toInt()}s",
                            style = Typography.bodyMedium.copy(
                                color = CyberCyan,
                                fontWeight = FontWeight.Bold,
                                fontFamily = FontFamily.Monospace
                            )
                        )
                    }

                    Slider(
                        value = interval,
                        onValueChange = {
                            interval = it
                            coroutineScope.launch { prefManager.updateInterval(it.toInt()) }
                        },
                        valueRange = 2f..60f,
                        steps = 58,
                        colors = SliderDefaults.colors(
                            thumbColor = CyberCyan,
                            activeTrackColor = CyberCyan,
                            inactiveTrackColor = SurfaceGlass
                        )
                    )

                    // Interval Preset Chips
                    Row(
                        modifier = Modifier.fillMaxWidth(),
                        horizontalArrangement = Arrangement.spacedBy(8.dp)
                    ) {
                        listOf(5, 10, 15, 30).forEach { sec ->
                            val isSelected = interval.toInt() == sec
                            FilterChip(
                                selected = isSelected,
                                onClick = {
                                    interval = sec.toFloat()
                                    coroutineScope.launch { prefManager.updateInterval(sec) }
                                },
                                label = {
                                    Text(
                                        "${sec}s${if (sec == 10) " (Def)" else ""}",
                                        style = Typography.labelSmall.copy(
                                            fontSize = 10.sp,
                                            color = if (isSelected) BgDark else TextPrimary
                                        )
                                    )
                                },
                                colors = FilterChipDefaults.filterChipColors(
                                    selectedContainerColor = CyberCyan,
                                    containerColor = SurfaceGlass
                                ),
                                border = FilterChipDefaults.filterChipBorder(
                                    borderColor = SurfaceBorder,
                                    enabled = true,
                                    selected = isSelected
                                )
                            )
                        }
                    }
                }

                HorizontalDivider(color = SurfaceBorder)

                // Session Keepalive Switch
                Row(
                    modifier = Modifier.fillMaxWidth(),
                    horizontalArrangement = Arrangement.SpaceBetween,
                    verticalAlignment = Alignment.CenterVertically
                ) {
                    Column(modifier = Modifier.weight(1f)) {
                        Text("Silent Session Keepalive", style = Typography.bodyLarge)
                        Text(
                            "Periodically pings gateway to prevent session lease expiry",
                            style = Typography.bodyMedium.copy(color = TextSecondary, fontSize = 12.sp)
                        )
                    }
                    Spacer(modifier = Modifier.width(12.dp))
                    Switch(
                        checked = config.keepaliveEnabled,
                        onCheckedChange = { checked ->
                            coroutineScope.launch { prefManager.updateKeepalive(checked) }
                        },
                        colors = SwitchDefaults.colors(
                            checkedThumbColor = CyberCyan,
                            checkedTrackColor = CyberCyan.copy(alpha = 0.4f)
                        )
                    )
                }

                // Auto-Connect Switch
                Row(
                    modifier = Modifier.fillMaxWidth(),
                    horizontalArrangement = Arrangement.SpaceBetween,
                    verticalAlignment = Alignment.CenterVertically
                ) {
                    Column(modifier = Modifier.weight(1f)) {
                        Text("Auto-Connect on Wi-Fi", style = Typography.bodyLarge)
                        Text(
                            "Automatically prime and log in when captive portal is intercepted",
                            style = Typography.bodyMedium.copy(color = TextSecondary, fontSize = 12.sp)
                        )
                    }
                    Spacer(modifier = Modifier.width(12.dp))
                    Switch(
                        checked = config.autoConnectEnabled,
                        onCheckedChange = { checked ->
                            coroutineScope.launch { prefManager.updateAutoConnect(checked) }
                        },
                        colors = SwitchDefaults.colors(
                            checkedThumbColor = SakuraPink,
                            checkedTrackColor = SakuraPink.copy(alpha = 0.4f)
                        )
                    )
                }
            }
        }

        // ==========================================
        // 3. CAMPUS SSID WHITELIST SECTION
        // ==========================================
        Card(
            shape = RoundedCornerShape(20.dp),
            colors = CardDefaults.cardColors(containerColor = SurfaceDark),
            modifier = Modifier
                .fillMaxWidth()
                .border(1.dp, SurfaceBorder, RoundedCornerShape(20.dp))
        ) {
            Column(
                modifier = Modifier.padding(18.dp),
                verticalArrangement = Arrangement.spacedBy(14.dp)
            ) {
                Row(
                    verticalAlignment = Alignment.CenterVertically,
                    horizontalArrangement = Arrangement.spacedBy(8.dp)
                ) {
                    Icon(
                        imageVector = Icons.Default.Wifi,
                        contentDescription = null,
                        tint = NeonLavender,
                        modifier = Modifier.size(18.dp)
                    )
                    Text(
                        text = "CAMPUS SSID WHITELIST",
                        style = Typography.labelSmall.copy(color = NeonLavender, fontWeight = FontWeight.Bold)
                    )
                }

                Text(
                    text = "Daemon only probes and auto-connects when your device is connected to one of these SSIDs.",
                    style = Typography.bodyMedium.copy(color = TextMuted, fontSize = 12.sp)
                )

                // Active SSID Chips
                if (config.ssidWhitelist.isEmpty()) {
                    Text(
                        "No SSID filter active (all Wi-Fi networks allowed)",
                        style = Typography.bodyMedium.copy(color = AlertOrange, fontSize = 12.sp)
                    )
                } else {
                    Row(
                        modifier = Modifier
                            .fillMaxWidth()
                            .horizontalScroll(rememberScrollState()),
                        horizontalArrangement = Arrangement.spacedBy(8.dp)
                    ) {
                        config.ssidWhitelist.forEach { ssid ->
                            InputChip(
                                selected = true,
                                onClick = {},
                                label = { Text(ssid, style = Typography.labelSmall.copy(color = TextPrimary)) },
                                trailingIcon = {
                                    Icon(
                                        imageVector = Icons.Default.Close,
                                        contentDescription = "Remove SSID",
                                        tint = SakuraPink,
                                        modifier = Modifier
                                            .size(16.dp)
                                            .clickable {
                                                val updated = config.ssidWhitelist - ssid
                                                coroutineScope.launch { prefManager.updateSsidWhitelist(updated) }
                                            }
                                    )
                                },
                                colors = InputChipDefaults.inputChipColors(
                                    selectedContainerColor = SurfaceGlass
                                ),
                                border = InputChipDefaults.inputChipBorder(
                                    borderColor = NeonLavender.copy(alpha = 0.6f),
                                    enabled = true,
                                    selected = true
                                )
                            )
                        }
                    }
                }

                // Add Custom SSID Input Row
                Row(
                    modifier = Modifier.fillMaxWidth(),
                    verticalAlignment = Alignment.CenterVertically,
                    horizontalArrangement = Arrangement.spacedBy(8.dp)
                ) {
                    OutlinedTextField(
                        value = newSsidInput,
                        onValueChange = { newSsidInput = it },
                        placeholder = { Text("Add campus SSID...") },
                        singleLine = true,
                        modifier = Modifier.weight(1f),
                        shape = RoundedCornerShape(12.dp),
                        colors = OutlinedTextFieldDefaults.colors(
                            focusedBorderColor = NeonLavender,
                            unfocusedBorderColor = SurfaceBorder,
                            focusedContainerColor = SurfaceGlass,
                            unfocusedContainerColor = SurfaceGlass
                        )
                    )

                    IconButton(
                        onClick = {
                            if (newSsidInput.isNotBlank()) {
                                val updated = config.ssidWhitelist + newSsidInput.trim()
                                coroutineScope.launch { prefManager.updateSsidWhitelist(updated) }
                                newSsidInput = ""
                            }
                        },
                        modifier = Modifier
                            .clip(RoundedCornerShape(12.dp))
                            .background(NeonLavender)
                    ) {
                        Icon(Icons.Default.Add, contentDescription = "Add SSID", tint = BgDark)
                    }
                }

                // Dynamic Preset SSID suggestions (BITS-STAFF, BITS-STUDENT)
                val campusPresets = listOf("BITS-STAFF", "BITS-STUDENT")
                val availableRecommendations = campusPresets.filter { it !in config.ssidWhitelist }

                if (availableRecommendations.isNotEmpty()) {
                    Column(verticalArrangement = Arrangement.spacedBy(6.dp)) {
                        Text(
                            text = "Recommended (tap to add):",
                            style = Typography.labelSmall.copy(color = TextMuted, fontSize = 10.sp)
                        )
                        Row(
                            modifier = Modifier
                                .fillMaxWidth()
                                .horizontalScroll(rememberScrollState()),
                            horizontalArrangement = Arrangement.spacedBy(8.dp)
                        ) {
                            availableRecommendations.forEach { preset ->
                                SuggestionChip(
                                    onClick = {
                                        val updated = config.ssidWhitelist + preset
                                        coroutineScope.launch { prefManager.updateSsidWhitelist(updated) }
                                    },
                                    label = { Text("+ $preset", style = Typography.labelSmall.copy(fontSize = 10.sp, color = CyberCyan)) },
                                    colors = SuggestionChipDefaults.suggestionChipColors(
                                        containerColor = SurfaceGlass
                                    ),
                                    border = SuggestionChipDefaults.suggestionChipBorder(
                                        borderColor = CyberCyan.copy(alpha = 0.4f),
                                        enabled = true
                                    )
                                )
                            }
                        }
                    }
                }
            }
        }

        // ==========================================
        // 4. BATTERY OPTIMIZATION & BACKGROUND DOZE
        // ==========================================
        Card(
            shape = RoundedCornerShape(20.dp),
            colors = CardDefaults.cardColors(containerColor = SurfaceDark),
            modifier = Modifier
                .fillMaxWidth()
                .border(1.dp, SurfaceBorder, RoundedCornerShape(20.dp))
        ) {
            Column(
                modifier = Modifier.padding(18.dp),
                verticalArrangement = Arrangement.spacedBy(12.dp)
            ) {
                Row(
                    verticalAlignment = Alignment.CenterVertically,
                    horizontalArrangement = Arrangement.spacedBy(8.dp)
                ) {
                    Icon(
                        imageVector = Icons.Default.BatteryFull,
                        contentDescription = null,
                        tint = if (isBatteryExempt) MintGreen else AlertOrange,
                        modifier = Modifier.size(18.dp)
                    )
                    Text(
                        text = "BACKGROUND DOZE EXEMPTION",
                        style = Typography.labelSmall.copy(
                            color = if (isBatteryExempt) MintGreen else AlertOrange,
                            fontWeight = FontWeight.Bold
                        )
                    )
                }

                if (isBatteryExempt) {
                    Row(
                        modifier = Modifier
                            .fillMaxWidth()
                            .clip(RoundedCornerShape(12.dp))
                            .background(MintGreen.copy(alpha = 0.12f))
                            .padding(12.dp),
                        verticalAlignment = Alignment.CenterVertically,
                        horizontalArrangement = Arrangement.spacedBy(10.dp)
                    ) {
                        Icon(Icons.Default.Check, contentDescription = null, tint = MintGreen, modifier = Modifier.size(20.dp))
                        Text(
                            text = "Unrestricted background execution active. Wify-chan daemon will run reliably during sleep.",
                            style = Typography.bodyMedium.copy(color = TextPrimary, fontSize = 12.sp)
                        )
                    }
                } else {
                    Text(
                        text = "Android battery saver may freeze background network keepalives when your phone screen turns off.",
                        style = Typography.bodyMedium.copy(color = TextSecondary, fontSize = 12.sp)
                    )

                    Button(
                        onClick = {
                            try {
                                val intent = Intent(Settings.ACTION_REQUEST_IGNORE_BATTERY_OPTIMIZATIONS).apply {
                                    data = Uri.parse("package:${context.packageName}")
                                }
                                context.startActivity(intent)
                                isBatteryExempt = powerManager.isIgnoringBatteryOptimizations(context.packageName)
                            } catch (e: Exception) {
                                Toast.makeText(context, "Cannot open battery settings: ${e.message}", Toast.LENGTH_SHORT).show()
                            }
                        },
                        modifier = Modifier.fillMaxWidth(),
                        shape = RoundedCornerShape(12.dp),
                        colors = ButtonDefaults.buttonColors(containerColor = AlertOrange)
                    ) {
                        Text("Grant Exemption (Unrestricted)", color = BgDark, fontWeight = FontWeight.Bold)
                    }
                }
            }
        }
    }
}
