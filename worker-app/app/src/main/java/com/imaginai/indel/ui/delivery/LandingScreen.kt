package com.imaginai.indel.ui.delivery

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.AccountCircle
import androidx.compose.material.icons.filled.Notifications
import androidx.compose.material.icons.filled.PlayArrow
import androidx.compose.material3.*
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.navigation.NavController
import com.imaginai.indel.R
import com.imaginai.indel.data.model.WorkerProfile
import com.imaginai.indel.ui.navigation.Screen
import com.imaginai.indel.ui.theme.*

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun LandingScreen(
    navController: NavController,
    viewModel: LandingViewModel = hiltViewModel()
) {
    val uiState by viewModel.uiState.collectAsState()

    LaunchedEffect(Unit) {
        viewModel.loadLandingData()
    }

    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text("InDel", fontWeight = FontWeight.Bold) },
                actions = {
                    IconButton(onClick = { navController.navigate(Screen.Notifications.route) }) {
                        Icon(Icons.Default.Notifications, contentDescription = stringResource(R.string.notifications))
                    }
                    IconButton(onClick = { navController.navigate(Screen.ProfileEdit.route) }) {
                        Icon(Icons.Default.AccountCircle, contentDescription = stringResource(R.string.profile))
                    }
                },
                colors = TopAppBarDefaults.topAppBarColors(
                    containerColor = BrandBlue,
                    titleContentColor = Color.White,
                    actionIconContentColor = Color.White
                )
            )
        }
    ) { padding ->
        Box(
            modifier = Modifier
                .padding(padding)
                .fillMaxSize()
                .background(BackgroundWarmWhite)
        ) {
            when (val state = uiState) {
                is LandingUiState.Loading -> CircularProgressIndicator(modifier = Modifier.align(Alignment.Center))
                is LandingUiState.Success -> LandingContent(state.worker, state.earningsToday, navController)
                is LandingUiState.Error -> Text(state.message, color = ErrorRed, modifier = Modifier.align(Alignment.Center))
            }
        }
    }
}

@Composable
fun LandingContent(
    worker: WorkerProfile,
    earningsToday: Double,
    navController: NavController
) {
    Column(
        modifier = Modifier
            .fillMaxSize()
            .padding(24.dp),
        horizontalAlignment = Alignment.CenterHorizontally,
        verticalArrangement = Arrangement.Top
    ) {
        Text(
            text = stringResource(R.string.hello_partner, worker.name ?: stringResource(R.string.partner)),
            style = MaterialTheme.typography.headlineMedium,
            fontWeight = FontWeight.Bold,
            modifier = Modifier.align(Alignment.Start)
        )
        Text(
            text = stringResource(R.string.ready_start_shift),
            style = MaterialTheme.typography.bodyLarge,
            color = TextSecondary,
            modifier = Modifier.align(Alignment.Start)
        )

        Spacer(modifier = Modifier.height(48.dp))

        // Start Delivery CTA
        Button(
            onClick = { navController.navigate(Screen.Orders.route) },
            modifier = Modifier
                .fillMaxWidth()
                .height(120.dp),
            shape = RoundedCornerShape(24.dp),
            colors = ButtonDefaults.buttonColors(containerColor = BrandBlue)
        ) {
            Row(verticalAlignment = Alignment.CenterVertically) {
                Icon(Icons.Default.PlayArrow, contentDescription = null, modifier = Modifier.size(48.dp))
                Spacer(modifier = Modifier.width(16.dp))
                Text(stringResource(R.string.start_delivery), fontSize = 24.sp, fontWeight = FontWeight.Bold)
            }
        }

        Spacer(modifier = Modifier.height(48.dp))

        // Quick Snapshot Cards
        Row(modifier = Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.spacedBy(16.dp)) {
            SnapshotCard(
                title = stringResource(R.string.earnings_today),
                value = "₹${earningsToday.toInt()}",
                modifier = Modifier.weight(1f)
            )
            SnapshotCard(
                title = stringResource(R.string.status),
                value = if (worker.enrolled == true) stringResource(R.string.protected) else stringResource(R.string.unprotected),
                color = if (worker.enrolled == true) SuccessGreen else ErrorRed,
                modifier = Modifier.weight(1f)
            )
        }

        Spacer(modifier = Modifier.weight(1f))

        // Bottom link to full dashboard
        TextButton(onClick = { navController.navigate(Screen.Home.route) }) {
            Text(stringResource(R.string.view_full_dashboard), color = BrandBlue, fontWeight = FontWeight.Bold)
        }
    }
}

@Composable
fun SnapshotCard(
    title: String,
    value: String,
    color: Color = BrandBlue,
    modifier: Modifier = Modifier
) {
    Card(
        modifier = modifier,
        shape = RoundedCornerShape(16.dp),
        colors = CardDefaults.cardColors(containerColor = Color.White),
        elevation = CardDefaults.cardElevation(defaultElevation = 2.dp)
    ) {
        Column(modifier = Modifier.padding(16.dp)) {
            Text(title, style = MaterialTheme.typography.labelSmall, color = TextSecondary)
            Text(value, style = MaterialTheme.typography.titleLarge, fontWeight = FontWeight.Bold, color = color)
        }
    }
}
