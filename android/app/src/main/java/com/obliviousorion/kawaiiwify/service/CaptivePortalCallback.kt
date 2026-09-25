package com.obliviousorion.kawaiiwify.service

import android.content.Context
import android.net.ConnectivityManager
import android.net.Network
import android.net.NetworkCapabilities
import android.net.wifi.WifiInfo
import android.net.wifi.WifiManager
import android.os.Build
import com.obliviousorion.kawaiiwify.core.LogLevel
import com.obliviousorion.kawaiiwify.core.Logger

class CaptivePortalCallback(
    private val context: Context,
    private val onNetworkEvent: (network: Network?, activeSsid: String?) -> Unit
) : ConnectivityManager.NetworkCallback(
    if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.S) FLAG_INCLUDE_LOCATION_INFO else 0
) {

    private val connectivityManager =
        context.getSystemService(Context.CONNECTIVITY_SERVICE) as ConnectivityManager

    var currentNetwork: Network? = null
        private set

    var currentSsid: String? = null
        private set

    private var boundNetwork: Network? = null
    private var lastCaptiveState: Boolean? = null
    private var lastValidatedState: Boolean? = null

    override fun onAvailable(network: Network) {
        super.onAvailable(network)
        val caps = connectivityManager.getNetworkCapabilities(network) ?: return
        if (caps.hasTransport(NetworkCapabilities.TRANSPORT_WIFI)) {
            val ssid = extractSsid(caps)
            currentNetwork = network
            currentSsid = ssid

            Logger.log("NET", "Wi-Fi network connected (SSID: ${ssid ?: "Unknown"})", LogLevel.INFO)
            bindNetwork(network)
            onNetworkEvent(network, ssid)
        }
    }

    override fun onCapabilitiesChanged(network: Network, networkCapabilities: NetworkCapabilities) {
        super.onCapabilitiesChanged(network, networkCapabilities)
        if (!networkCapabilities.hasTransport(NetworkCapabilities.TRANSPORT_WIFI)) return

        val hasCaptive = networkCapabilities.hasCapability(NetworkCapabilities.NET_CAPABILITY_CAPTIVE_PORTAL)
        val isValidated = networkCapabilities.hasCapability(NetworkCapabilities.NET_CAPABILITY_VALIDATED)
        val ssid = extractSsid(networkCapabilities)

        // Prevent spam: only trigger if network, captive state, or validation actually changed
        val stateChanged = hasCaptive != lastCaptiveState ||
                isValidated != lastValidatedState ||
                currentNetwork != network ||
                currentSsid != ssid

        if (stateChanged) {
            lastCaptiveState = hasCaptive
            lastValidatedState = isValidated
            currentNetwork = network
            currentSsid = ssid

            Logger.log("NET", "Network state: Captive=$hasCaptive, Validated=$isValidated", LogLevel.INFO)
            bindNetwork(network)
            onNetworkEvent(network, ssid)
        }
    }

    override fun onLost(network: Network) {
        super.onLost(network)
        if (currentNetwork == network) {
            Logger.log("NET", "Wi-Fi connection disconnected", LogLevel.WARN)
            currentNetwork = null
            currentSsid = null
            lastCaptiveState = null
            lastValidatedState = null
            unbindNetwork()
            onNetworkEvent(null, null)
        }
    }

    private fun bindNetwork(network: Network) {
        if (boundNetwork == network) return // Already bound to this network interface

        try {
            if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.M) {
                connectivityManager.bindProcessToNetwork(network)
            } else {
                @Suppress("DEPRECATION")
                ConnectivityManager.setProcessDefaultNetwork(network)
            }
            boundNetwork = network
            Logger.log("NET", "Process bound to Wi-Fi socket interface", LogLevel.INFO)
        } catch (e: Exception) {
            Logger.log("NET", "Failed to bind process to network: ${e.message}", LogLevel.ERROR)
        }
    }

    private fun unbindNetwork() {
        if (boundNetwork == null) return
        try {
            if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.M) {
                connectivityManager.bindProcessToNetwork(null)
            } else {
                @Suppress("DEPRECATION")
                ConnectivityManager.setProcessDefaultNetwork(null)
            }
            boundNetwork = null
        } catch (_: Exception) {}
    }

    private fun extractSsid(capabilities: NetworkCapabilities): String? {
        // 1. Try NetworkCapabilities transportInfo (Android 10+)
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.Q) {
            val wifiInfo = capabilities.transportInfo as? WifiInfo
            val raw = wifiInfo?.ssid
            if (!raw.isNullOrBlank() && raw != "<unknown ssid>" && raw != "0x") {
                return raw.removeSurrounding("\"")
            }
        }

        // 2. Fallback to WifiManager connectionInfo
        try {
            val wifiManager = context.applicationContext.getSystemService(Context.WIFI_SERVICE) as? WifiManager
            @Suppress("DEPRECATION")
            val raw = wifiManager?.connectionInfo?.ssid
            if (!raw.isNullOrBlank() && raw != "<unknown ssid>" && raw != "0x") {
                return raw.removeSurrounding("\"")
            }
        } catch (_: Exception) {}

        return null
    }
}
