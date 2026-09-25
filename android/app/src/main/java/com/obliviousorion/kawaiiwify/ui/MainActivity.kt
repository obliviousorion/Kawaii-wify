package com.obliviousorion.kawaiiwify.ui

import android.Manifest
import android.content.Intent
import android.content.pm.PackageManager
import android.os.Build
import android.os.Bundle
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.activity.result.contract.ActivityResultContracts
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.padding
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Code
import androidx.compose.material.icons.filled.Home
import androidx.compose.material.icons.filled.Settings
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.vector.ImageVector
import androidx.core.content.ContextCompat
import com.obliviousorion.kawaiiwify.KawaiiApplication
import com.obliviousorion.kawaiiwify.core.Constants
import com.obliviousorion.kawaiiwify.service.KeepaliveForegroundService
import com.obliviousorion.kawaiiwify.ui.screens.DashboardScreen
import com.obliviousorion.kawaiiwify.ui.screens.LogsScreen
import com.obliviousorion.kawaiiwify.ui.screens.SettingsScreen
import com.obliviousorion.kawaiiwify.ui.theme.*

sealed class NavItem(val route: String, val title: String, val icon: ImageVector) {
    data object Dashboard : NavItem("dashboard", "Status", Icons.Default.Home)
    data object Logs : NavItem("logs", "Console", Icons.Default.Code)
    data object Settings : NavItem("settings", "Config", Icons.Default.Settings)
}

class MainActivity : ComponentActivity() {

    private val requestPermissionLauncher = registerForActivityResult(
        ActivityResultContracts.RequestPermission()
    ) { _ -> }

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)

        checkNotificationPermission()
        // Automatically start the foreground service on app launch
        KeepaliveForegroundService.start(this)

        setContent {
            KawaiiWifyTheme {
                MainAppHost(
                    onConnect = {
                        val intent = Intent(this, KeepaliveForegroundService::class.java).apply {
                            action = Constants.ACTION_CONNECT
                        }
                        startService(intent)
                    },
                    onDisconnect = {
                        val intent = Intent(this, KeepaliveForegroundService::class.java).apply {
                            action = Constants.ACTION_DISCONNECT
                        }
                        startService(intent)
                    },
                    onPause = {
                        val intent = Intent(this, KeepaliveForegroundService::class.java).apply {
                            action = Constants.ACTION_PAUSE
                        }
                        startService(intent)
                    },
                    onResume = {
                        val intent = Intent(this, KeepaliveForegroundService::class.java).apply {
                            action = Constants.ACTION_RESUME
                        }
                        startService(intent)
                    }
                )
            }
        }
    }

    private fun checkNotificationPermission() {
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.TIRAMISU) {
            if (ContextCompat.checkSelfPermission(this, Manifest.permission.POST_NOTIFICATIONS)
                != PackageManager.PERMISSION_GRANTED
            ) {
                requestPermissionLauncher.launch(Manifest.permission.POST_NOTIFICATIONS)
            }
        }
    }
}

@Composable
fun MainAppHost(
    onConnect: () -> Unit,
    onDisconnect: () -> Unit,
    onPause: () -> Unit,
    onResume: () -> Unit
) {
    var selectedItem by remember { mutableStateOf<NavItem>(NavItem.Dashboard) }
    val items = listOf(NavItem.Dashboard, NavItem.Logs, NavItem.Settings)

    Scaffold(
        modifier = Modifier.fillMaxSize(),
        containerColor = BgDark,
        bottomBar = {
            NavigationBar(
                containerColor = SurfaceDark,
                contentColor = TextPrimary
            ) {
                items.forEach { item ->
                    val isSelected = selectedItem == item
                    NavigationBarItem(
                        selected = isSelected,
                        onClick = { selectedItem = item },
                        icon = {
                            Icon(
                                imageVector = item.icon,
                                contentDescription = item.title,
                                tint = if (isSelected) SakuraPink else TextSecondary
                            )
                        },
                        label = {
                            Text(
                                text = item.title,
                                color = if (isSelected) SakuraPink else TextSecondary,
                                style = Typography.labelSmall
                            )
                        },
                        colors = NavigationBarItemDefaults.colors(
                            indicatorColor = SurfaceGlass
                        )
                    )
                }
            }
        }
    ) { paddingValues ->
        val modifier = Modifier.padding(paddingValues)
        when (selectedItem) {
            is NavItem.Dashboard -> DashboardScreen(
                engine = KawaiiApplication.instance.sessionEngine,
                onConnectClick = onConnect,
                onDisconnectClick = onDisconnect,
                onPauseClick = onPause,
                onResumeClick = onResume,
                modifier = modifier
            )
            is NavItem.Logs -> LogsScreen(modifier = modifier)
            is NavItem.Settings -> SettingsScreen(modifier = modifier)
        }
    }
}
