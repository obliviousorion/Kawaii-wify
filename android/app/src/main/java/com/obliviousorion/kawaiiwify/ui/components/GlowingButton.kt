package com.obliviousorion.kawaiiwify.ui.components

import androidx.compose.animation.animateColorAsState
import androidx.compose.animation.core.*
import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.interaction.MutableInteractionSource
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.Text
import androidx.compose.material3.ripple
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.remember
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.draw.shadow
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.obliviousorion.kawaiiwify.domain.EngineState
import com.obliviousorion.kawaiiwify.ui.theme.*

@Composable
fun GlowingButton(
    state: EngineState,
    onClick: () -> Unit,
    modifier: Modifier = Modifier,
    enabled: Boolean = true
) {
    val isOnline = state is EngineState.Online
    val isCooldown = state is EngineState.Cooldown
    val isCaptive = state is EngineState.Captive

    val primaryGlowColor by animateColorAsState(
        targetValue = when {
            isCooldown -> CrimsonRed
            isOnline -> NeonLavender
            isCaptive -> AlertOrange
            else -> SakuraPink
        },
        animationSpec = tween(500),
        label = "buttonGlow"
    )

    val buttonText = when {
        isCooldown -> "IN COOLDOWN (${(state as EngineState.Cooldown).remainingSeconds}s)"
        isOnline -> "DISCONNECT"
        isCaptive -> "AUTHENTICATING..."
        else -> "CONNECT TO CAMPUS"
    }

    val infiniteTransition = rememberInfiniteTransition(label = "btnPulse")
    val glowAlpha by infiniteTransition.animateFloat(
        initialValue = 0.4f,
        targetValue = 0.9f,
        animationSpec = infiniteRepeatable(
            animation = tween(1200, easing = LinearEasing),
            repeatMode = RepeatMode.Reverse
        ),
        label = "glowAlpha"
    )

    Box(
        modifier = modifier
            .fillMaxWidth()
            .height(58.dp)
            .shadow(
                elevation = if (enabled) 12.dp else 0.dp,
                shape = RoundedCornerShape(29.dp),
                ambientColor = primaryGlowColor,
                spotColor = primaryGlowColor
            )
            .clip(RoundedCornerShape(29.dp))
            .background(
                Brush.horizontalGradient(
                    colors = if (isOnline) {
                        listOf(Color(0xFF2A1B4D), Color(0xFF1B2A4D))
                    } else {
                        listOf(Color(0xFF4A1535), Color(0xFF15354A))
                    }
                )
            )
            .border(
                width = 2.dp,
                brush = Brush.horizontalGradient(
                    colors = listOf(
                        primaryGlowColor.copy(alpha = glowAlpha),
                        CyberCyan.copy(alpha = glowAlpha)
                    )
                ),
                shape = RoundedCornerShape(29.dp)
            )
            .clickable(
                interactionSource = remember { MutableInteractionSource() },
                indication = ripple(color = primaryGlowColor),
                enabled = enabled && !isCooldown,
                onClick = onClick
            ),
        contentAlignment = Alignment.Center
    ) {
        Row(
            verticalAlignment = Alignment.CenterVertically,
            horizontalArrangement = Arrangement.Center
        ) {
            if (isCaptive) {
                CircularProgressIndicator(
                    modifier = Modifier.size(20.dp),
                    color = AlertOrange,
                    strokeWidth = 2.5.dp
                )
                Spacer(modifier = Modifier.width(12.dp))
            }

            Text(
                text = buttonText,
                style = Typography.titleMedium.copy(
                    fontWeight = FontWeight.Bold,
                    letterSpacing = 1.2.sp,
                    color = if (enabled && !isCooldown) TextPrimary else TextMuted
                )
            )
        }
    }
}
