package com.imaginai.indel.ui.orders

import android.util.Log
import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.imaginai.indel.data.model.BatchOrderDto
import com.imaginai.indel.data.model.DeliveryBatchDto
import com.imaginai.indel.data.model.Order
import com.imaginai.indel.data.repository.WorkerRepository
import dagger.hilt.android.lifecycle.HiltViewModel
import okhttp3.ResponseBody
import java.util.Locale
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch
import javax.inject.Inject

@HiltViewModel
class OrdersViewModel @Inject constructor(
    private val workerRepository: WorkerRepository
) : ViewModel() {

    companion object {
        private const val TAG = "OrdersViewModel"
    }

    private val _uiState = MutableStateFlow<OrdersUiState>(OrdersUiState.Loading)
    val uiState = _uiState.asStateFlow()

    private val _isRefreshing = MutableStateFlow(false)
    val isRefreshing = _isRefreshing.asStateFlow()

    private val _selectedTab = MutableStateFlow(BatchLifecycleTab.AVAILABLE_NEAR)
    val selectedTab = _selectedTab.asStateFlow()

    private val _selectedOrderTab = MutableStateFlow(OrderLifecycleTab.AVAILABLE)
    val selectedOrderTab = _selectedOrderTab.asStateFlow()

    private var cachedBatches: List<DeliveryBatch> = emptyList()
    private var hasAttemptedAutoPopulate = false

    init {
        loadOrders()
    }

    fun loadOrders() {
        viewModelScope.launch {
            _uiState.value = OrdersUiState.Loading
            hasAttemptedAutoPopulate = false
            fetchOrders()
        }
    }

    fun refresh() {
        viewModelScope.launch {
            _isRefreshing.value = true
            fetchOrders()
            delay(500)
            _isRefreshing.value = false
        }
    }

    private suspend fun fetchOrders() {
        try {
            val availableRes = workerRepository.getAvailableOrders()
            val assignedRes = workerRepository.getAssignedOrders()
            val allRes = workerRepository.getAllOrders()

            val availableError = availableRes.errorBody().asDebugString()
            val assignedError = assignedRes.errorBody().asDebugString()
            val allError = allRes.errorBody().asDebugString()

            val availableOrders = if (availableRes.isSuccessful) {
                (availableRes.body()?.orders ?: emptyList())
                    .filter { it.status.equals("assigned", ignoreCase = true) }
                    .sortedByDescending { it.assignedAt }
            } else {
                emptyList()
            }
            val activeOrders = if (assignedRes.isSuccessful) {
                (assignedRes.body()?.orders ?: emptyList())
                    .filter {
                        it.status.equals("accepted", ignoreCase = true) ||
                            it.status.equals("picked_up", ignoreCase = true)
                    }
                    .sortedByDescending { it.assignedAt }
            } else {
                emptyList()
            }
            val completedOrders = if (allRes.isSuccessful) {
                (allRes.body()?.orders ?: emptyList())
                    .filter { it.status.equals("delivered", ignoreCase = true) }
                    .sortedByDescending { it.assignedAt }
            } else {
                emptyList()
            }

            val activeOrderIds = activeOrders.map { it.orderId }.toSet()
            val availableOrderIds = availableOrders.map { it.orderId }.toSet()
            val sanitizedAvailableOrders = availableOrders.filterNot { it.orderId in activeOrderIds }
            val sanitizedCompletedOrders = completedOrders.filterNot {
                it.orderId in activeOrderIds || it.orderId in availableOrderIds
            }

            if (
                availableOrders.isEmpty() &&
                activeOrders.isEmpty() &&
                completedOrders.isEmpty() &&
                !hasAttemptedAutoPopulate
            ) {
                hasAttemptedAutoPopulate = true
                val seeded = workerRepository.assignOrders(4)
                Log.d(TAG, "fetchOrders autoPopulate status=${seeded.code()}")
                if (seeded.isSuccessful) {
                    fetchOrders()
                    return
                }
            }

            cachedBatches = emptyList()

            Log.d(
                TAG,
                "fetchOrders availableStatus=${availableRes.code()} assignedStatus=${assignedRes.code()} allStatus=${allRes.code()} " +
                    "available=${sanitizedAvailableOrders.size} active=${activeOrders.size} completed=${sanitizedCompletedOrders.size} " +
                    "availableError=$availableError assignedError=$assignedError allError=$allError"
            )

            if (availableRes.isSuccessful || assignedRes.isSuccessful || allRes.isSuccessful) {
                _uiState.value = OrdersUiState.Success(
                    availableOrders = sanitizedAvailableOrders,
                    activeOrders = activeOrders,
                    completedOrders = sanitizedCompletedOrders,
                    availableNearBatches = emptyList(),
                    pickedUpBatches = emptyList(),
                    deliveryBatches = emptyList(),
                )
            } else {
                _uiState.value = OrdersUiState.Error(
                    "Failed to load orders (available=${availableRes.code()}, assigned=${assignedRes.code()}, all=${allRes.code()})"
                )
            }
        } catch (e: Exception) {
            Log.e(TAG, "fetchOrders exception", e)
            _uiState.value = OrdersUiState.Error(e.message ?: "Unknown error")
        }
    }

    fun acceptOrder(orderId: String) {
        viewModelScope.launch {
            val response = workerRepository.acceptOrder(orderId)
            Log.d(TAG, "acceptOrder orderId=$orderId status=${response.code()} error=${response.errorBody().asDebugString()}")
            if (response.isSuccessful) {
                fetchOrders()
            }
        }
    }

    fun pickedUpOrder(orderId: String) {
        viewModelScope.launch {
            val response = workerRepository.pickedUpOrder(orderId)
            Log.d(TAG, "pickedUpOrder orderId=$orderId status=${response.code()} error=${response.errorBody().asDebugString()}")
            if (response.isSuccessful) {
                fetchOrders()
            }
        }
    }

    fun deliveredOrder(orderId: String, customerCode: String) {
        viewModelScope.launch {
            val response = workerRepository.deliveredOrder(orderId, customerCode)
            Log.d(
                TAG,
                "deliveredOrder orderId=$orderId status=${response.code()} error=${response.errorBody().asDebugString()}"
            )
            if (response.isSuccessful) {
                fetchOrders()
            }
        }
    }

    fun selectTab(tab: BatchLifecycleTab) {
        _selectedTab.value = tab
    }

    fun selectOrderTab(tab: OrderLifecycleTab) {
        _selectedOrderTab.value = tab
    }

    fun getBatchById(batchId: String): DeliveryBatch? = cachedBatches.firstOrNull { it.batchId == batchId }

    fun isZoneASingleStop(batch: DeliveryBatch): Boolean {
        val from = batch.fromCity.trim().lowercase()
        val to = batch.toCity.trim().lowercase()
        return batch.zoneLevel.trim().equals("A", ignoreCase = true) && from.isNotBlank() && from == to
    }

    suspend fun acceptBatch(batch: DeliveryBatch, pickupCode: String): BatchActionResult {
        val response = workerRepository.acceptBatch(batch.batchId, batch.orders.map { it.orderId }, pickupCode)
        return if (response.isSuccessful) {
            fetchOrders()
            BatchActionResult(success = true, errorMessage = null)
        } else {
            BatchActionResult(success = false, errorMessage = response.errorBody()?.string()?.ifBlank { null } ?: "Unable to pick up this batch right now.")
        }
    }

    suspend fun deliverBatch(batch: DeliveryBatch, deliveryCode: String): BatchDeliveryResult {
        val response = workerRepository.deliverBatch(batch.batchId, deliveryCode)
        return if (response.isSuccessful) {
            fetchOrders()
            val body = response.body()
            BatchDeliveryResult(
                success = true,
                batchCompleted = body?.batchCompleted == true,
                remainingOrders = body?.remainingOrders,
                errorMessage = null,
            )
        } else {
            BatchDeliveryResult(
                success = false,
                batchCompleted = false,
                remainingOrders = null,
                errorMessage = response.errorBody()?.string()?.ifBlank { null } ?: "Unable to complete delivery right now.",
            )
        }
    }

    private fun DeliveryBatchDto.toUiBatch(): DeliveryBatch {
        return DeliveryBatch(
            batchId = batchId,
            batchKey = batchKey,
            batchGroupKey = batchGroupKey,
            batchIndex = batchIndex,
            zoneLevel = zoneLevel,
            fromCity = fromCity,
            toCity = toCity,
            totalWeight = totalWeight,
            targetWeight = targetWeight,
            maxWeight = maxWeight,
            orderCount = orderCount,
            status = status,
            pickupCode = pickupCode,
            deliveryCode = deliveryCode,
            pickupTime = pickupTime,
            deliveryTime = deliveryTime,
            batchEarningInr = batchEarningInr,
            orders = orders.map { it.toUiOrder() },
        )
    }

    private fun BatchOrderDto.toUiOrder(): BatchOrder {
        return BatchOrder(
            orderId = orderId,
            deliveryAddress = deliveryAddress,
            contactName = contactName,
            contactPhone = contactPhone,
            weight = weight,
            pickupArea = pickupArea,
            dropArea = dropArea,
            deliveryCode = deliveryCode,
            status = status,
            pickupTime = pickupTime,
            deliveryTime = deliveryTime,
        )
    }
}

private fun ResponseBody?.asDebugString(): String {
    return try {
        this?.string()?.ifBlank { "<empty>" } ?: "<none>"
    } catch (e: Exception) {
        "<error:${e.message}>"
    }
}

data class BatchActionResult(
    val success: Boolean,
    val errorMessage: String?,
)

data class BatchDeliveryResult(
    val success: Boolean,
    val batchCompleted: Boolean,
    val remainingOrders: Int?,
    val errorMessage: String?,
)

enum class BatchLifecycleTab(val title: String) {
    AVAILABLE_NEAR("Available / Near"),
    PICKED_UP("Picked Up"),
    DELIVERY("Delivery"),
}

enum class OrderLifecycleTab(val title: String) {
    AVAILABLE("Available"),
    ACTIVE("Accepted"),
    COMPLETED("Completed"),
}

sealed class OrdersUiState {
    object Loading : OrdersUiState()
    data class Success(
        val availableOrders: List<Order>,
        val activeOrders: List<Order>,
        val completedOrders: List<Order>,
        val availableNearBatches: List<DeliveryBatch>,
        val pickedUpBatches: List<DeliveryBatch>,
        val deliveryBatches: List<DeliveryBatch>,
    ) : OrdersUiState()
    data class Error(val message: String) : OrdersUiState()
}


