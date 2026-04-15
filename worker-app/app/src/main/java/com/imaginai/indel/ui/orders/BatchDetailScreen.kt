package com.imaginai.indel.ui.orders

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material3.Button
import androidx.compose.material3.ButtonDefaults
import androidx.compose.material3.Card
import androidx.compose.material3.CardDefaults
import androidx.compose.material3.Divider
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.OutlinedTextField
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Text
import androidx.compose.material3.TopAppBar
import androidx.compose.material3.TopAppBarDefaults
import androidx.compose.runtime.Composable
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.mutableStateMapOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.runtime.saveable.rememberSaveable
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.navigation.NavController
import com.imaginai.indel.R
import com.imaginai.indel.ui.navigation.Screen
import com.imaginai.indel.ui.theme.BackgroundWarmWhite
import com.imaginai.indel.ui.theme.BlueSoft
import com.imaginai.indel.ui.theme.BrandBlue
import com.imaginai.indel.ui.theme.ErrorRed
import com.imaginai.indel.ui.theme.SuccessGreen
import com.imaginai.indel.ui.theme.TextSecondary
import kotlinx.coroutines.launch
import java.util.Locale

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun BatchDetailScreen(
    navController: NavController,
    batchId: String,
) {
    val ordersEntry = remember(navController) {
        runCatching { navController.getBackStackEntry(Screen.Orders.route) }.getOrNull()
    }
    val viewModel: OrdersViewModel = ordersEntry?.let { hiltViewModel(it) } ?: hiltViewModel()
    val uiState by viewModel.uiState.collectAsState()
    val batch = viewModel.getBatchById(batchId)
    val coroutineScope = rememberCoroutineScope()
    val isZoneASingleStop = remember(batch?.batchId, batch?.zoneLevel, batch?.fromCity, batch?.toCity) {
        batch?.let { viewModel.isZoneASingleStop(it) } == true
    }
    val normalizedBatchStatus = remember(batch?.status) {
        batch?.status
            ?.trim()
            ?.lowercase(Locale.getDefault())
            ?.replace(" ", "_")
            .orEmpty()
    }
    val orderDeliveryCodes = remember(batchId) { mutableStateMapOf<String, String>() }
    var selectedZoneAOrderId by rememberSaveable(batchId) { mutableStateOf<String?>(null) }
    val deliveredOrderCount = remember(batch?.batchId, batch?.orders) {
        batch?.orders?.count { it.status.equals("delivered", ignoreCase = true) } ?: 0
    }
    var enteredCode by rememberSaveable(batchId) { mutableStateOf("") }
    var feedbackMessage by rememberSaveable(batchId) { mutableStateOf<String?>(null) }
    var isAccepting by rememberSaveable(batchId) { mutableStateOf(false) }
    var isDelivering by rememberSaveable(batchId) { mutableStateOf(false) }

    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text(stringResource(R.string.batch_details), fontWeight = FontWeight.Bold) },
                navigationIcon = {
                    IconButton(onClick = { navController.navigateUp() }) {
                        Icon(Icons.AutoMirrored.Filled.ArrowBack, contentDescription = stringResource(R.string.back))
                    }
                },
                colors = TopAppBarDefaults.topAppBarColors(
                    containerColor = BrandBlue,
                    titleContentColor = Color.White,
                    navigationIconContentColor = Color.White,
                )
            )
        }
    ) { padding ->
        when {
            batch != null -> {
                LazyColumn(
                    modifier = Modifier
                        .fillMaxSize()
                        .padding(padding)
                        .background(BackgroundWarmWhite),
                    contentPadding = PaddingValues(16.dp),
                    verticalArrangement = Arrangement.spacedBy(12.dp),
                ) {
                    item {
                        Card(
                            modifier = Modifier.fillMaxWidth(),
                            shape = RoundedCornerShape(14.dp),
                            colors = CardDefaults.cardColors(containerColor = Color.White),
                            border = androidx.compose.foundation.BorderStroke(1.dp, BlueSoft)
                        ) {
                            Column(modifier = Modifier.padding(14.dp)) {
                                Text(batch.batchId, style = MaterialTheme.typography.titleSmall, color = BrandBlue, fontWeight = FontWeight.Bold)
                                Spacer(modifier = Modifier.height(6.dp))
                                Text(stringResource(R.string.zone_value, batch.zoneLevel), style = MaterialTheme.typography.bodySmall, color = TextSecondary)
                                Text(compactRouteLabel(batch.zoneLevel, batch.fromCity, batch.toCity), style = MaterialTheme.typography.bodyMedium, fontWeight = FontWeight.SemiBold)
                                Text(stringResource(R.string.batch_earning_value, String.format(java.util.Locale.getDefault(), "%.0f", batch.batchEarningInr ?: 0.0)), style = MaterialTheme.typography.bodySmall, color = SuccessGreen)
                                if (batch.totalWeight < 10.0) {
                                    Text(stringResource(R.string.packing_below_target), style = MaterialTheme.typography.labelSmall, color = TextSecondary)
                                } else {
                                    Text(stringResource(R.string.packed_for_range), style = MaterialTheme.typography.labelSmall, color = SuccessGreen)
                                }
                            }
                        }
                    }

                    item {
                        Card(
                            modifier = Modifier.fillMaxWidth(),
                            shape = RoundedCornerShape(14.dp),
                            colors = CardDefaults.cardColors(containerColor = Color.White),
                            border = androidx.compose.foundation.BorderStroke(1.dp, BlueSoft)
                        ) {
                            Column(modifier = Modifier.padding(14.dp)) {
                                Text(stringResource(R.string.delivery_progress), style = MaterialTheme.typography.labelLarge, fontWeight = FontWeight.Bold)
                                Spacer(modifier = Modifier.height(6.dp))
                                Text(
                                    stringResource(R.string.orders_delivered, deliveredOrderCount, batch.orders.size),
                                    style = MaterialTheme.typography.bodyMedium,
                                    fontWeight = FontWeight.SemiBold,
                                    color = BrandBlue
                                )
                                Text(
                                    when {
                                        normalizedBatchStatus == "delivered" -> stringResource(R.string.batch_completed_earnings_released)
                                        normalizedBatchStatus == "picked_up" -> if (isZoneASingleStop) stringResource(R.string.enter_each_order_delivery_code) else stringResource(R.string.enter_delivery_code_move)
                                        else -> stringResource(R.string.pick_up_batch_first)
                                    },
                                    style = MaterialTheme.typography.bodySmall,
                                    color = TextSecondary
                                )
                            }
                        }
                    }

                    item {
                        when {
                            normalizedBatchStatus == "delivered" -> {
                                Card(
                                    modifier = Modifier.fillMaxWidth(),
                                    shape = RoundedCornerShape(14.dp),
                                    colors = CardDefaults.cardColors(containerColor = Color.White),
                                    border = androidx.compose.foundation.BorderStroke(1.dp, SuccessGreen)
                                ) {
                                    Column(modifier = Modifier.padding(14.dp)) {
                                        Text(stringResource(R.string.delivery_complete), style = MaterialTheme.typography.titleMedium, fontWeight = FontWeight.Bold, color = SuccessGreen)
                                        Spacer(modifier = Modifier.height(6.dp))
                                        Text(
                                            stringResource(R.string.all_orders_delivered_posted),
                                            style = MaterialTheme.typography.bodySmall,
                                            color = TextSecondary
                                        )
                                    }
                                }
                            }
                            normalizedBatchStatus == "picked_up" && isZoneASingleStop -> {
                                ZoneABatchOrdersSection(
                                    batch = batch,
                                    orderDeliveryCodes = orderDeliveryCodes,
                                    selectedOrderId = selectedZoneAOrderId,
                                    onSelectOrder = { orderId ->
                                        selectedZoneAOrderId = if (selectedZoneAOrderId == orderId) null else orderId
                                    },
                                    feedbackMessage = feedbackMessage,
                                    isDelivering = isDelivering,
                                    onDeliverOrder = { order, code ->
                                        isDelivering = true
                                        coroutineScope.launch {
                                            val deliveryResult = viewModel.deliverBatch(batch, code)
                                            isDelivering = false
                                            feedbackMessage = if (deliveryResult.success) {
                                                orderDeliveryCodes[order.orderId] = ""
                                                selectedZoneAOrderId = null
                                                if (deliveryResult.batchCompleted) {
                                                    stringResource(R.string.delivery_accepted_delivered)
                                                } else {
                                                    val remaining = deliveryResult.remainingOrders ?: (batch.orders.size - (deliveredOrderCount + 1))
                                                    stringResource(R.string.delivery_accepted_remaining, remaining)
                                                }
                                            } else {
                                                deliveryResult.errorMessage ?: stringResource(R.string.unable_complete_delivery_right_now)
                                            }
                                        }
                                    }
                                )
                            }
                            else -> {
                                BatchActionCard(
                                    batch = batch,
                                    enteredCode = enteredCode,
                                    onEnteredCodeChange = { enteredCode = it.take(4) },
                                    feedbackMessage = feedbackMessage,
                                    onConfirmPickup = {
                                        val code = enteredCode.trim()
                                        if (code.isBlank()) {
                                            feedbackMessage = stringResource(R.string.enter_pickup_code)
                                            return@BatchActionCard
                                        }
                                        isAccepting = true
                                        coroutineScope.launch {
                                            val accepted = viewModel.acceptBatch(batch, code)
                                            isAccepting = false
                                            feedbackMessage = if (accepted.success) {
                                                enteredCode = ""
                                                stringResource(R.string.batch_picked_up_successfully)
                                            } else {
                                                accepted.errorMessage ?: stringResource(R.string.unable_pick_up_batch_right_now)
                                            }
                                        }
                                    },
                                    onConfirmDelivery = {
                                        val code = enteredCode.trim()
                                        if (code.isBlank()) {
                                            feedbackMessage = stringResource(R.string.enter_delivery_code_move)
                                            return@BatchActionCard
                                        }

                                        isDelivering = true
                                        coroutineScope.launch {
                                            val deliveryResult = viewModel.deliverBatch(batch, code)
                                            isDelivering = false
                                            feedbackMessage = if (deliveryResult.success) {
                                                enteredCode = ""
                                                if (deliveryResult.batchCompleted) {
                                                    stringResource(R.string.delivery_accepted_delivered)
                                                } else {
                                                    val remaining = deliveryResult.remainingOrders ?: (batch.orders.size - (deliveredOrderCount + 1))
                                                    stringResource(R.string.delivery_accepted_remaining, remaining)
                                                }
                                            } else {
                                                deliveryResult.errorMessage ?: stringResource(R.string.unable_complete_delivery_right_now)
                                            }
                                        }
                                    },
                                    isAccepting = isAccepting,
                                    isDelivering = isDelivering,
                                )
                            }
                        }
                    }

                    item {
                        Text(stringResource(R.string.orders_in_this_batch), style = MaterialTheme.typography.titleMedium, fontWeight = FontWeight.Bold)
                    }

                    items(batch.orders) { nestedOrder ->
                        Card(
                            modifier = Modifier.fillMaxWidth(),
                            shape = RoundedCornerShape(12.dp),
                            colors = CardDefaults.cardColors(containerColor = Color.White),
                            elevation = CardDefaults.cardElevation(defaultElevation = 1.dp)
                        ) {
                            Column(modifier = Modifier.padding(12.dp)) {
                                Text(stringResource(R.string.order_value, nestedOrder.orderId), style = MaterialTheme.typography.labelLarge, fontWeight = FontWeight.SemiBold)
                                Spacer(modifier = Modifier.height(6.dp))
                                Text(stringResource(R.string.address_value, nestedOrder.deliveryAddress), style = MaterialTheme.typography.bodySmall)
                                Text(stringResource(R.string.contact_value, nestedOrder.contactName), style = MaterialTheme.typography.bodySmall, color = TextSecondary)
                                Text(stringResource(R.string.phone_value, nestedOrder.contactPhone), style = MaterialTheme.typography.bodySmall, color = TextSecondary)
                                Text(stringResource(R.string.pickup_value, compactRouteLabel(batch.zoneLevel, nestedOrder.pickupArea ?: "-", nestedOrder.dropArea ?: nestedOrder.deliveryAddress)), style = MaterialTheme.typography.bodySmall, color = TextSecondary)
                                if (!nestedOrder.pickupTime.isNullOrBlank()) {
                                    Text(stringResource(R.string.picked_up_value, nestedOrder.pickupTime), style = MaterialTheme.typography.bodySmall, color = TextSecondary)
                                }
                                if (!nestedOrder.deliveryTime.isNullOrBlank()) {
                                    Text(stringResource(R.string.delivered_value, nestedOrder.deliveryTime), style = MaterialTheme.typography.bodySmall, color = SuccessGreen)
                                }

                                Text(stringResource(R.string.weight_value, formatBatchWeight(nestedOrder.weight)), style = MaterialTheme.typography.bodySmall, color = TextSecondary)
                                Text(stringResource(R.string.status_value, nestedOrder.status ?: stringResource(R.string.assigned_status)), style = MaterialTheme.typography.bodySmall, color = TextSecondary)
                                Divider(modifier = Modifier.padding(top = 8.dp), color = Color(0xFFE9ECEF))
                            }
                        }
                    }
                }
            }
            uiState is OrdersUiState.Loading -> {
                Box(
                    modifier = Modifier
                        .fillMaxSize()
                        .padding(padding)
                        .background(BackgroundWarmWhite),
                    contentAlignment = Alignment.Center,
                ) {
                    Text(stringResource(R.string.loading_batch), color = TextSecondary)
                }
            }
            else -> {
                Box(
                    modifier = Modifier
                        .fillMaxSize()
                        .padding(padding)
                        .background(BackgroundWarmWhite),
                    contentAlignment = Alignment.Center,
                ) {
                    Text(stringResource(R.string.batch_not_found), color = TextSecondary)
                }
            }
        }
    }

}

@Composable
private fun ZoneABatchDeliverySection(
    batch: DeliveryBatch,
    orderDeliveryCodes: MutableMap<String, String>,
    selectedOrderId: String?,
    onSelectOrder: (String) -> Unit,
    feedbackMessage: String?,
    isDelivering: Boolean,
    onDeliverOrder: (BatchOrder, String) -> Unit,
) {
    Card(
        modifier = Modifier.fillMaxWidth(),
        shape = RoundedCornerShape(14.dp),
        colors = CardDefaults.cardColors(containerColor = Color.White),
        border = androidx.compose.foundation.BorderStroke(1.dp, BrandBlue.copy(alpha = 0.25f))
    ) {
        Column(modifier = Modifier.padding(14.dp), verticalArrangement = Arrangement.spacedBy(12.dp)) {
            Text(stringResource(R.string.zone_a_delivery), style = MaterialTheme.typography.titleMedium, fontWeight = FontWeight.Bold, color = BrandBlue)
            Text(
                stringResource(R.string.deliver_each_order),
                style = MaterialTheme.typography.bodySmall,
                color = TextSecondary
            )

            batch.orders.forEach { order ->
                val isDelivered = order.status.equals("delivered", ignoreCase = true)
                val isExpanded = selectedOrderId == order.orderId || batch.orders.count { it.status.equals("delivered", ignoreCase = true) } == batch.orders.size - 1
                Card(
                    modifier = Modifier.fillMaxWidth(),
                    shape = RoundedCornerShape(12.dp),
                    colors = CardDefaults.cardColors(containerColor = BackgroundWarmWhite),
                    border = androidx.compose.foundation.BorderStroke(1.dp, if (isDelivered) SuccessGreen else BlueSoft)
                ) {
                    Column(
                        modifier = Modifier
                            .fillMaxWidth()
                            .clickable(enabled = !isDelivered) { onSelectOrder(order.orderId) }
                            .padding(12.dp),
                        verticalArrangement = Arrangement.spacedBy(8.dp)
                    ) {
                        Text(order.orderId, style = MaterialTheme.typography.labelLarge, fontWeight = FontWeight.Bold)
                        Text("Address: ${order.deliveryAddress}", style = MaterialTheme.typography.bodySmall)
                        Text("Contact: ${order.contactName}", style = MaterialTheme.typography.bodySmall, color = TextSecondary)
                        Text("Phone: ${order.contactPhone}", style = MaterialTheme.typography.bodySmall, color = TextSecondary)
                        if (!order.deliveryTime.isNullOrBlank()) {
                            Text(stringResource(R.string.delivered_at_value, order.deliveryTime), style = MaterialTheme.typography.bodySmall, color = SuccessGreen)
                        }

                        if (isDelivered) {
                            Text(stringResource(R.string.delivery_complete), color = SuccessGreen, fontWeight = FontWeight.Bold)
                        } else {
                            Text(
                                if (isExpanded) stringResource(R.string.enter_code_for_order) else stringResource(R.string.tap_to_enter_delivery_code),
                                style = MaterialTheme.typography.bodySmall,
                                color = TextSecondary
                            )
                            if (isExpanded) {

                                OutlinedTextField(
                                    value = orderDeliveryCodes[order.orderId].orEmpty(),
                                    onValueChange = { orderDeliveryCodes[order.orderId] = it.take(4) },
                                    modifier = Modifier.fillMaxWidth(),
                                    singleLine = true,
                                    label = { Text(stringResource(R.string.delivery_code)) },
                                    placeholder = { Text(stringResource(R.string.enter_4_digit_code_short)) }
                                )
                                Button(
                                    onClick = {
                                        val code = orderDeliveryCodes[order.orderId].orEmpty().trim()
                                        if (code.isBlank()) {
                                            return@Button
                                        }
                                        onDeliverOrder(order, code)
                                    },
                                    modifier = Modifier.fillMaxWidth(),
                                    colors = ButtonDefaults.buttonColors(containerColor = SuccessGreen),
                                    shape = RoundedCornerShape(12.dp),
                                    enabled = !isDelivering
                                ) {
                                    Text(if (isDelivering) stringResource(R.string.checking) else stringResource(R.string.verify_move))
                                }
                                Text(
                                    stringResource(R.string.exact_order_code_hint),
                                    style = MaterialTheme.typography.bodySmall,
                                    color = TextSecondary
                                )
                            }
                        }
                    }
                }
            }

            feedbackMessage?.let {
                Text(
                    it,
                    color = if (it.contains("unable", ignoreCase = true) || it.contains("incorrect", ignoreCase = true)) ErrorRed else SuccessGreen,
                    style = MaterialTheme.typography.bodySmall,
                )
            }
        }
    }
}

@Composable
private fun ZoneABatchOrdersSection(
    batch: DeliveryBatch,
    orderDeliveryCodes: MutableMap<String, String>,
    selectedOrderId: String?,
    onSelectOrder: (String) -> Unit,
    feedbackMessage: String?,
    isDelivering: Boolean,
    onDeliverOrder: (BatchOrder, String) -> Unit,
) {
    Card(
        modifier = Modifier.fillMaxWidth(),
        shape = RoundedCornerShape(14.dp),
        colors = CardDefaults.cardColors(containerColor = Color.White),
        border = androidx.compose.foundation.BorderStroke(1.dp, BrandBlue.copy(alpha = 0.25f))
    ) {
        Column(modifier = Modifier.padding(14.dp), verticalArrangement = Arrangement.spacedBy(12.dp)) {
            Text("Zone A delivery", style = MaterialTheme.typography.titleMedium, fontWeight = FontWeight.Bold, color = BrandBlue)
            Text("Touch an order to reveal its delivery code field, then verify that order.", style = MaterialTheme.typography.bodySmall, color = TextSecondary)

            batch.orders.forEach { order ->
                val isDelivered = order.status.equals("delivered", ignoreCase = true)
                val isExpanded = selectedOrderId == order.orderId

                Card(
                    modifier = Modifier
                        .fillMaxWidth()
                        .clickable(enabled = !isDelivered) { onSelectOrder(order.orderId) },
                    shape = RoundedCornerShape(12.dp),
                    colors = CardDefaults.cardColors(containerColor = BackgroundWarmWhite),
                    border = androidx.compose.foundation.BorderStroke(1.dp, if (isDelivered) SuccessGreen else BlueSoft)
                ) {
                    Column(modifier = Modifier.padding(12.dp), verticalArrangement = Arrangement.spacedBy(8.dp)) {
                        Text(order.orderId, style = MaterialTheme.typography.labelLarge, fontWeight = FontWeight.Bold)
                        Text("Address: ${order.deliveryAddress}", style = MaterialTheme.typography.bodySmall)
                        Text("Contact: ${order.contactName}", style = MaterialTheme.typography.bodySmall, color = TextSecondary)
                        Text("Phone: ${order.contactPhone}", style = MaterialTheme.typography.bodySmall, color = TextSecondary)
                        if (!order.deliveryTime.isNullOrBlank()) {
                            Text("Delivered at: ${order.deliveryTime}", style = MaterialTheme.typography.bodySmall, color = SuccessGreen)
                        }

                        if (isDelivered) {
                            Text("Delivered", color = SuccessGreen, fontWeight = FontWeight.Bold)
                        } else {
                            Text(
                                if (isExpanded) "Enter the delivery code for this order" else "Tap to enter delivery code",
                                style = MaterialTheme.typography.bodySmall,
                                color = BrandBlue,
                                fontWeight = FontWeight.SemiBold,
                            )
                            if (isExpanded) {
                                OutlinedTextField(
                                    value = orderDeliveryCodes[order.orderId].orEmpty(),
                                    onValueChange = { orderDeliveryCodes[order.orderId] = it.take(4) },
                                    modifier = Modifier.fillMaxWidth(),
                                    singleLine = true,
                                    label = { Text("Delivery code") },
                                    placeholder = { Text("Enter 4-digit code") }
                                )
                                Button(
                                    onClick = {
                                        val code = orderDeliveryCodes[order.orderId].orEmpty().trim()
                                        if (code.isBlank()) {
                                            return@Button
                                        }
                                        onDeliverOrder(order, code)
                                    },
                                    modifier = Modifier.fillMaxWidth(),
                                    colors = ButtonDefaults.buttonColors(containerColor = SuccessGreen),
                                    shape = RoundedCornerShape(12.dp),
                                    enabled = !isDelivering
                                ) {
                                    Text(if (isDelivering) "Checking..." else "Verify & Move")
                                }
                                Text("Enter this exact order code to mark the order delivered.", style = MaterialTheme.typography.bodySmall, color = TextSecondary)
                            }
                        }
                    }
                }
            }

            feedbackMessage?.let {
                Text(
                    it,
                    color = if (it.contains("unable", ignoreCase = true) || it.contains("incorrect", ignoreCase = true)) ErrorRed else SuccessGreen,
                    style = MaterialTheme.typography.bodySmall,
                )
            }
        }
    }
}

@Composable
private fun BatchLevelDeliverySection(
    batch: DeliveryBatch,
    enteredCode: String,
    onEnteredCodeChange: (String) -> Unit,
    feedbackMessage: String?,
    isDelivering: Boolean,
    onDeliverBatch: () -> Unit,
) {
    Card(
        modifier = Modifier.fillMaxWidth(),
        shape = RoundedCornerShape(14.dp),
        colors = CardDefaults.cardColors(containerColor = Color.White),
        border = androidx.compose.foundation.BorderStroke(1.dp, SuccessGreen.copy(alpha = 0.35f))
    ) {
        Column(modifier = Modifier.padding(14.dp), verticalArrangement = Arrangement.spacedBy(10.dp)) {
            Text("Inter-city delivery", style = MaterialTheme.typography.titleMedium, fontWeight = FontWeight.Bold, color = SuccessGreen)
            Text(
                "Zone ${batch.zoneLevel} batches are delivered at batch level. Enter the batch code once to complete every order.",
                style = MaterialTheme.typography.bodySmall,
                color = TextSecondary
            )
            OutlinedTextField(
                value = enteredCode,
                onValueChange = onEnteredCodeChange,
                modifier = Modifier.fillMaxWidth(),
                singleLine = true,
                label = { Text("Delivery code") },
                placeholder = { Text("Enter 4-digit code") }
            )
            Text("Enter the delivery code you received to complete this batch.", style = MaterialTheme.typography.labelSmall, color = TextSecondary)
            Button(
                onClick = onDeliverBatch,
                modifier = Modifier.fillMaxWidth(),
                colors = ButtonDefaults.buttonColors(containerColor = SuccessGreen),
                shape = RoundedCornerShape(12.dp),
                enabled = !isDelivering
            ) {
                Text(if (isDelivering) "Checking..." else "Verify & Move")
            }
            feedbackMessage?.let {
                Text(
                    it,
                    color = if (it.contains("unable", ignoreCase = true) || it.contains("incorrect", ignoreCase = true)) ErrorRed else SuccessGreen,
                    style = MaterialTheme.typography.bodySmall,
                )
            }
        }
    }
}

@Composable
private fun BatchActionCard(
    batch: DeliveryBatch,
    enteredCode: String,
    onEnteredCodeChange: (String) -> Unit,
    feedbackMessage: String?,
    onConfirmPickup: () -> Unit,
    onConfirmDelivery: () -> Unit,
    isAccepting: Boolean,
    isDelivering: Boolean,
) {
    val statusLower = batch.status
        .trim()
        .lowercase(Locale.getDefault())
        .replace(" ", "_")
    val statusLabel = when (statusLower) {
        "assigned" -> "Assigned"
        "accepted" -> "Accepted"
        "picked_up" -> "Picked Up"
        "delivered" -> "Delivered"
        else -> batch.status.replace("_", " ").replaceFirstChar { it.uppercase() }
    }
    val isPickupStage = statusLower == "pending" || statusLower == "assigned" || statusLower == "accepted"
    val isDeliveryStage = statusLower == "picked_up"
    val actionLabel = when {
        isPickupStage -> "Verify & Move"
        isDeliveryStage -> "Verify & Move"
        else -> "Completed"
    }
    val inputLabel = if (isPickupStage) "Pickup code" else "Delivery code"

    Card(
        modifier = Modifier.fillMaxWidth(),
        shape = RoundedCornerShape(14.dp),
        colors = CardDefaults.cardColors(containerColor = Color.White),
        border = androidx.compose.foundation.BorderStroke(1.dp, BlueSoft)
    ) {
        Column(modifier = Modifier.padding(14.dp), verticalArrangement = Arrangement.spacedBy(10.dp)) {
            Text("Batch status", style = MaterialTheme.typography.labelLarge, fontWeight = FontWeight.Bold)
            Text(statusLabel, color = BrandBlue, fontWeight = FontWeight.SemiBold)

            if (isPickupStage || isDeliveryStage) {
                OutlinedTextField(
                    value = enteredCode,
                    onValueChange = { onEnteredCodeChange(it.take(8)) },
                    modifier = Modifier.fillMaxWidth(),
                    singleLine = true,
                    label = { Text(inputLabel) },
                    placeholder = { Text("Enter code") },
                )

                Text(
                    if (isPickupStage) {
                        "Enter the pickup code to move this batch from assigned to picked up."
                    } else {
                        "Enter the delivery code to move this batch from picked up to delivered."
                    },
                    style = MaterialTheme.typography.bodySmall,
                    color = TextSecondary,
                )
            }

            if (feedbackMessage != null) {
                Text(
                    feedbackMessage,
                    color = if (feedbackMessage.contains("unable", ignoreCase = true) || feedbackMessage.contains("incorrect", ignoreCase = true) || feedbackMessage.contains("error", ignoreCase = true)) ErrorRed else SuccessGreen,
                    style = MaterialTheme.typography.bodySmall,
                )
            }

            if (isPickupStage) {
                Button(
                    onClick = onConfirmPickup,
                    modifier = Modifier.fillMaxWidth(),
                    colors = ButtonDefaults.buttonColors(containerColor = BrandBlue),
                    shape = RoundedCornerShape(12.dp),
                    enabled = !isAccepting,
                ) {
                    Text(if (isAccepting) "Checking..." else actionLabel)
                }
            } else if (isDeliveryStage) {
                Button(
                    onClick = onConfirmDelivery,
                    modifier = Modifier.fillMaxWidth(),
                    colors = ButtonDefaults.buttonColors(containerColor = SuccessGreen),
                    shape = RoundedCornerShape(12.dp),
                    enabled = !isDelivering,
                ) {
                    Text(if (isDelivering) "Checking..." else actionLabel)
                }
            } else {
                Text(
                    "This batch is already delivered.",
                    style = MaterialTheme.typography.bodySmall,
                    color = TextSecondary,
                )
            }
        }
    }
}
private fun compactRouteLabel(zoneLevel: String, fromLabel: String, toLabel: String): String {
    val from = fromLabel.trim()
    val to = toLabel.trim()
    if (zoneLevel.equals("A", ignoreCase = true) && from.equals(to, ignoreCase = true)) {
        return if (from.isBlank()) "Unknown" else from
    }
    if (from.isBlank() && to.isBlank()) {
        return "Unknown"
    }
    if (from.isBlank()) {
        return to
    }
    if (to.isBlank()) {
        return from
    }
    return "$from -> $to"
}

private fun formatBatchWeight(weight: Double): String {
    return String.format(Locale.getDefault(), "%.1f", weight)
}


