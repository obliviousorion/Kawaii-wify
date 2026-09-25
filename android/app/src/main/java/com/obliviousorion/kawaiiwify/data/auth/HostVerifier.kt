package com.obliviousorion.kawaiiwify.data.auth

import java.security.SecureRandom
import java.security.cert.X509Certificate
import javax.net.ssl.*

object HostVerifier {

    fun createInsecureTrustManager(): X509TrustManager {
        return object : X509TrustManager {
            override fun checkClientTrusted(chain: Array<out X509Certificate>?, authType: String?) {}
            override fun checkServerTrusted(chain: Array<out X509Certificate>?, authType: String?) {}
            override fun getAcceptedIssuers(): Array<X509Certificate> = arrayOf()
        }
    }

    fun createSSLSocketFactory(trustManager: X509TrustManager): SSLSocketFactory {
        val sslContext = SSLContext.getInstance("TLS")
        sslContext.init(null, arrayOf<TrustManager>(trustManager), SecureRandom())
        return sslContext.socketFactory
    }

    fun createHostnameVerifier(allowedHost: String?): HostnameVerifier {
        return HostnameVerifier { hostname, session ->
            if (allowedHost.isNullOrBlank()) return@HostnameVerifier true
            val cleanAllowed = allowedHost.split(":").first()
            val cleanHost = hostname.split(":").first()
            cleanHost.equals(cleanAllowed, ignoreCase = true) || cleanHost.contains(cleanAllowed, ignoreCase = true)
        }
    }
}
