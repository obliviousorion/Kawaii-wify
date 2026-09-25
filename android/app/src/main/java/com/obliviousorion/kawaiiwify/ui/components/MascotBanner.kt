package com.obliviousorion.kawaiiwify.ui.components

import androidx.compose.animation.animateColorAsState
import androidx.compose.animation.core.*
import androidx.compose.foundation.Image
import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.draw.shadow
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.res.painterResource
import androidx.compose.ui.text.font.FontStyle
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.obliviousorion.kawaiiwify.domain.EngineState
import com.obliviousorion.kawaiiwify.ui.theme.*

@Composable
fun MascotBanner(
    state: EngineState,
    modifier: Modifier = Modifier,
    provider: MascotThemeProvider = MascotRegistry.currentProvider
) {
    val visual = provider.resolve(state)

    val animatedAura by animateColorAsState(
        targetValue = visual.auraColor,
        animationSpec = tween(durationMillis = 600),
        label = "auraColor"
    )

    // Pulsing aura animation
    val infiniteTransition = rememberInfiniteTransition(label = "pulse")
    val pulseScale by infiniteTransition.animateFloat(
        initialValue = 0.95f,
        targetValue = 1.05f,
        animationSpec = infiniteRepeatable(
            animation = tween(1500, easing = FastOutSlowInEasing),
            repeatMode = RepeatMode.Reverse
        ),
        label = "pulseScale"
    )

    Column(
        modifier = modifier
            .fillMaxWidth()
            .clip(RoundedCornerShape(24.dp))
            .background(
                Brush.verticalGradient(
                    colors = listOf(
                        SurfaceGlass,
                        SurfaceDark
                    )
                )
            )
            .border(1.dp, SurfaceBorder, RoundedCornerShape(24.dp))
            .padding(20.dp),
        horizontalAlignment = Alignment.CenterHorizontally
    ) {
        // Mascot Avatar with Glowing Neon Ring
        Box(
            contentAlignment = Alignment.Center,
            modifier = Modifier.size(160.dp)
        ) {
            // Glow shadow aura
            Box(
                modifier = Modifier
                    .size((148 * pulseScale).dp)
                    .clip(CircleShape)
                    .background(animatedAura.copy(alpha = 0.25f))
            )

            // Mascot Image Container
            Image(
                painter = painterResource(id = visual.drawableRes),
                contentDescription = "Wify-chan Mascot",
                contentScale = ContentScale.Crop,
                modifier = Modifier
                    .size(136.dp)
                    .clip(CircleShape)
                    .border(2.dp, animatedAura, CircleShape)
                    .shadow(12.dp, CircleShape)
            )
        }

        Spacer(modifier = Modifier.height(14.dp))

        // State Badge
        Box(
            modifier = Modifier
                .clip(RoundedCornerShape(50))
                .background(visual.badgeColor.copy(alpha = 0.15f))
                .border(1.dp, visual.badgeColor.copy(alpha = 0.6f), RoundedCornerShape(50))
                .padding(horizontal = 14.dp, vertical = 4.dp)
        ) {
            Text(
                text = visual.badgeText,
                color = visual.badgeColor,
                style = Typography.labelSmall.copy(fontWeight = FontWeight.Bold, letterSpacing = 1.2.sp)
            )
        }

        Spacer(modifier = Modifier.height(10.dp))

        // Mascot Speech Bubble / Quote
        Text(
            text = visual.statusQuote,
            style = Typography.bodyMedium.copy(
                fontStyle = FontStyle.Italic,
                color = TextPrimary.copy(alpha = 0.9f)
            ),
            textAlign = TextAlign.Center,
            modifier = Modifier.padding(horizontal = 12.dp)
        )
    }
}
