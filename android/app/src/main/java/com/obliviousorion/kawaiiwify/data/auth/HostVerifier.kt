package com.obliviousorion.kawaiiwify.data.auth

import java.security.MessageDigest
import java.security.SecureRandom
import java.security.cert.CertificateException
import java.security.cert.X509Certificate
import javax.net.ssl.*

object HostVerifier {

    class CertificatePinMismatchException(message: String) : CertificateException(message)

    fun createPinningTrustManager(
        endpoint: String,
        trustedPins: List<String> = emptyList(),
        verifyTls: Boolean = true,
        onRecordPin: ((fingerprint: String) -> Unit)? = null
    ): X509TrustManager {
        return object : X509TrustManager {
            override fun checkClientTrusted(chain: Array<out X509Certificate>?, authType: String?) {}

            override fun checkServerTrusted(chain: Array<out X509Certificate>?, authType: String?) {
                if (!verifyTls) return

                val leaf = chain?.firstOrNull()
                    ?: throw CertificateException("No certificates presented by gateway server")

                val digest = MessageDigest.getInstance("SHA-256")
                val hashBytes = digest.digest(leaf.encoded)
                val hexString = hashBytes.joinToString("") { "%02X".format(it) }
                val presentedPin = "SHA256:$hexString"

                // 1. Enforce pinned certificates if present
                if (trustedPins.isNotEmpty()) {
                    val match = trustedPins.any { it.equals(presentedPin, ignoreCase = true) }
                    if (!match) {
                        throw CertificatePinMismatchException(
                            "Gateway certificate mismatch for $endpoint! Presented: $presentedPin, Trusted: $trustedPins"
                        )
                    }
                    return
                }

                // 2. TOFU Mode: Record the observed fingerprint to commit after successful auth
                onRecordPin?.invoke(presentedPin)
            }

            override fun getAcceptedIssuers(): Array<X509Certificate> = arrayOf()
        }
    }

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
        return HostnameVerifier { hostname, _ ->
            if (allowedHost.isNullOrBlank()) return@HostnameVerifier true
            val cleanAllowed = allowedHost.split(":").first().trim()
            val cleanHost = hostname.split(":").first().trim()
            cleanHost.equals(cleanAllowed, ignoreCase = true)
        }
    }
}
