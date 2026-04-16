package com.imaginai.indel.ui.orders

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material.icons.filled.LocationOn
import androidx.compose.material3.Button
import androidx.compose.material3.ButtonDefaults
import androidx.compose.material3.Card
import androidx.compose.material3.CardDefaults
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Surface
import androidx.compose.material3.Tab
import androidx.compose.material3.TabRow
import androidx.compose.material3.Text
import androidx.compose.material3.TopAppBar
import androidx.compose.material3.TopAppBarDefaults
import androidx.compose.material3.pulltorefresh.PullToRefreshBox
import androidx.compose.runtime.Composable
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.navigation.NavController
import com.imaginai.indel.data.model.Order
import com.imaginai.indel.ui.navigation.Screen
import com.imaginai.indel.ui.theme.BackgroundWarmWhite
import com.imaginai.indel.ui.theme.BlueSoft
import com.imaginai.indel.ui.theme.BrandBlue
import com.imaginai.indel.ui.theme.ErrorRed
import com.imaginai.indel.ui.theme.SuccessGreen
import com.imaginai.indel.ui.theme.TextSecondary

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun OrderPipelineScreen(
    navController: NavController,
    viewModel: OrdersViewModel = hiltViewModel()
) {
    val uiState by viewModel.uiState.collectAsState()
    val isRefreshing by viewModel.isRefreshing.collectAsState()
    val selectedTab by viewModel.selectedOrderTab.collectAsState()

    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text("Orders", fontWeight = FontWeight.Bold) },
                navigationIcon = {
                    IconButton(onClick = { navController.navigateUp() }) {
                        Icon(Icons.AutoMirrored.Filled.ArrowBack, contentDescription = "Back")
                    }
                },
                colors = TopAppBarDefaults.topAppBarColors(
                    containerColor = BrandBlue,
                    titleContentColor = Color.White,
                    navigationIconContentColor = Color.White
                )
            )
        }
    ) { padding ->
        PullToRefreshBox(
            isRefreshing = isRefreshing,
            onRefresh = { viewModel.refresh() },
            modifier = Modifier.padding(padding)
        ) {
            Box(
                modifier = Modifier
                    .fillMaxSize()
                    .background(BackgroundWarmWhite)
            ) {
                when (val state = uiState) {
                    is OrdersUiState.Loading -> CircularProgressIndicator(modifier = Modifier.align(Alignment.Center))
                    is OrdersUiState.Success -> OrderPipelineContent(
                        state = state,
                        selectedTab = selectedTab,
                        onTabSelected = viewModel::selectOrderTab,
                        onAcceptOrder = viewModel::acceptOrder,
                        onProgressOrder = { order ->
                            val destination = if (order.status.equals("picked_up", ignoreCase = true)) {
                                Screen.DeliveryCompletion.createRoute(order.orderId)
                            } else {
                                Screen.DeliveryExecution.createRoute(order.orderId)
                            }
                            navController.navigate(destination)
                        }
                    )
                    is OrdersUiState.Error -> OrderPipelineError(state.message) { viewModel.loadOrders() }
                }
            }
        }
    }
}

@Composable
private fun OrderPipelineContent(
    state: OrdersUiState.Success,
    selectedTab: OrderLifecycleTab,
    onTabSelected: (OrderLifecycleTab) -> Unit,
    onAcceptOrder: (String) -> Unit,
    onProgressOrder: (Order) -> Unit,
) {
    val tabs = OrderLifecycleTab.entries
    val visibleOrders = when (selectedTab) {
        OrderLifecycleTab.AVAILABLE -> state.availableOrders
        OrderLifecycleTab.ACTIVE -> state.activeOrders
        OrderLifecycleTab.COMPLETED -> state.completedOrders
    }

    LazyColumn(
        modifier = Modifier.fillMaxSize(),
        contentPadding = PaddingValues(16.dp),
        verticalArrangement = Arrangement.spacedBy(16.dp)
    ) {
        item {
            Text(
                "Order Pipeline",
                style = MaterialTheme.typography.titleMedium,
                fontWeight = FontWeight.Bold
            )
            Spacer(modifier = Modifier.height(8.dp))
            TabRow(selectedTabIndex = tabs.indexOf(selectedTab), containerColor = Color.White) {
                tabs.forEach { tab ->
                    val count = when (tab) {
                        OrderLifecycleTab.AVAILABLE -> state.availableOrders.size
                        OrderLifecycleTab.ACTIVE -> state.activeOrders.size
                        OrderLifecycleTab.COMPLETED -> state.completedOrders.size
                    }
                    Tab(
                        selected = selectedTab == tab,
                        onClick = { onTabSelected(tab) },
                        text = { Text("${tab.title} ($count)") }
                    )
                }
            }
        }

        if (visibleOrders.isEmpty()) {
            item {
                Box(modifier = Modifier.fillParentMaxSize(), contentAlignment = Alignment.Center) {
                    Text("No ${selectedTab.title.lowercase()} orders right now", color = TextSecondary)
                }
            }
        } else {
            items(visibleOrders, key = { it.orderId }) { order ->
                OrderPipelineCard(
                    order = order,
                    selectedTab = selectedTab,
                    onAcceptOrder = onAcceptOrder,
                    onProgressOrder = onProgressOrder,
                )
            }
        }
    }
}

@Composable
private fun OrderPipelineCard(
    order: Order,
    selectedTab: OrderLifecycleTab,
    onAcceptOrder: (String) -> Unit,
    onProgressOrder: (Order) -> Unit,
) {
    val routeLabel = when {
        order.pickupArea.isNotBlank() && order.dropArea.isNotBlank() -> "${order.pickupArea} -> ${order.dropArea}"
        !order.fromCity.isNullOrBlank() && !order.toCity.isNullOrBlank() -> "${order.fromCity} -> ${order.toCity}"
        else -> order.orderId
    }
    val statusLabel = when (order.status.trim().lowercase()) {
        "assigned" -> "Waiting for acceptance"
        "accepted" -> "Accepted"
        "picked_up" -> "Picked Up"
        "delivered" -> "Delivered"
        else -> order.status
    }
    val routeTypeLabel = when (order.routeType?.trim()?.lowercase()) {
        "local" -> "Local"
        "interstate" -> "Interstate"
        else -> "Other"
    }

    val activeActionLabel = if (order.status.trim().equals("picked_up", ignoreCase = true)) {
        "Complete Order"
    } else {
        "Start Pickup"
    }

    Card(
        modifier = Modifier.fillMaxWidth(),
        shape = RoundedCornerShape(16.dp),
        colors = CardDefaults.cardColors(containerColor = Color.White),
        elevation = CardDefaults.cardElevation(defaultElevation = 2.dp)
    ) {
        Column(modifier = Modifier.padding(16.dp), verticalArrangement = Arrangement.spacedBy(12.dp)) {
            Row(modifier = Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.SpaceBetween) {
                Column(modifier = Modifier.weight(1f)) {
                    Text(order.orderId, style = MaterialTheme.typography.labelSmall, color = TextSecondary)
                    Text(routeLabel, style = MaterialTheme.typography.titleMedium, fontWeight = FontWeight.Bold, color = BrandBlue)
                    Text(
                        "Zone ${order.zoneLevel ?: "-"} • $routeTypeLabel • ${order.zoneRouteDisplay ?: ""}",
                        style = MaterialTheme.typography.labelSmall,
                        color = TextSecondary
                    )
                }
                Surface(color = BlueSoft, shape = RoundedCornerShape(999.dp)) {
                    Text(
                        text = statusLabel,
                        modifier = Modifier.padding(horizontal = 10.dp, vertical = 6.dp),
                        style = MaterialTheme.typography.labelMedium,
                        fontWeight = FontWeight.Bold,
                        color = BrandBlue
                    )
                }
            }

            Row(verticalAlignment = Alignment.CenterVertically) {
                Icon(Icons.Default.LocationOn, contentDescription = null, tint = BrandBlue)
                Spacer(modifier = Modifier.width(8.dp))
                Column {
                    Text("Pickup: ${order.pickupArea}", style = MaterialTheme.typography.bodySmall, fontWeight = FontWeight.Medium)
                    Text("Drop: ${order.dropArea}", style = MaterialTheme.typography.bodySmall, color = TextSecondary)
                }
            }

            Row(modifier = Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.SpaceBetween) {
                Text("Order amount: Rs ${order.orderValue.toInt()}", style = MaterialTheme.typography.bodySmall, color = TextSecondary)
                Text("Earn: Rs ${order.earningInr.toInt()}", style = MaterialTheme.typography.bodySmall, color = SuccessGreen, fontWeight = FontWeight.Bold)
            }
            Row(modifier = Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.SpaceBetween) {
                Text("Fee: Rs ${order.deliveryFeeInr.toInt()}", style = MaterialTheme.typography.bodySmall, color = TextSecondary)
                Text("Distance: ${String.format("%.1f", order.distanceKm)} km", style = MaterialTheme.typography.bodySmall, color = TextSecondary)
            }

            when (selectedTab) {
                OrderLifecycleTab.AVAILABLE -> Button(
                    onClick = { onAcceptOrder(order.orderId) },
                    modifier = Modifier.fillMaxWidth(),
                    shape = RoundedCornerShape(12.dp),
                    colors = ButtonDefaults.buttonColors(containerColor = BrandBlue)
                ) {
                    Text("Accept Order")
                }
                OrderLifecycleTab.ACTIVE -> Button(
                    onClick = { onProgressOrder(order) },
                    modifier = Modifier.fillMaxWidth(),
                    shape = RoundedCornerShape(12.dp),
                    colors = ButtonDefaults.buttonColors(containerColor = SuccessGreen)
                ) {
                    Text(activeActionLabel)
                }
                OrderLifecycleTab.COMPLETED -> Text(
                    "Completed and posted to earnings",
                    color = SuccessGreen,
                    fontWeight = FontWeight.Bold,
                )
            }
        }
    }
}

@Composable
private fun OrderPipelineError(message: String, onRetry: () -> Unit) {
    Column(
        modifier = Modifier.fillMaxSize(),
        verticalArrangement = Arrangement.Center,
        horizontalAlignment = Alignment.CenterHorizontally
    ) {
        Text(message, color = ErrorRed)
        Button(onClick = onRetry, modifier = Modifier.padding(top = 16.dp)) {
            Text("Retry")
        }
    }
}
