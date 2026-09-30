package com.obliviousorion.kawaiiwify.core

object Constants {
    const val DEFAULT_GATEWAY = "fw.bits-pilani.ac.in:8090"
    const val PROBE_URL = "http://connectivitycheck.gstatic.com/generate_204"
    const val DEFAULT_CHECK_INTERVAL_SECONDS = 10
    const val MAX_AUTH_FAILURES = 3
    const val COOLDOWN_DURATION_SECONDS = 10

    const val NOTIFICATION_CHANNEL_ID = "kawaii_wify_keepalive"
    const val NOTIFICATION_ID = 1337

    const val SECURITY_ALERT_CHANNEL_ID = "kawaii_wify_security_alerts"
    const val SECURITY_NOTIFICATION_ID = 1338

    const val UPDATE_ALERT_CHANNEL_ID = "kawaii_wify_updates"
    const val UPDATE_NOTIFICATION_ID = 1339
    const val EXTRA_OPEN_UPDATE_DIALOG = "extra_open_update_dialog"

    const val ACTION_CONNECT = "com.obliviousorion.kawaiiwify.ACTION_CONNECT"
    const val ACTION_DISCONNECT = "com.obliviousorion.kawaiiwify.ACTION_DISCONNECT"
    const val ACTION_PAUSE = "com.obliviousorion.kawaiiwify.ACTION_PAUSE"
    const val ACTION_RESUME = "com.obliviousorion.kawaiiwify.ACTION_RESUME"
    const val ACTION_STOP_SERVICE = "com.obliviousorion.kawaiiwify.ACTION_STOP_SERVICE"
}
