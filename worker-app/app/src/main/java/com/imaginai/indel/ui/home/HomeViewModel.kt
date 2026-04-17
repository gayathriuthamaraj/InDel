package com.imaginai.indel.ui.home

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.imaginai.indel.data.model.*
import com.imaginai.indel.data.repository.EarningsRepository
import com.imaginai.indel.data.repository.PolicyRepository
import com.imaginai.indel.data.repository.WorkerRepository
import android.util.Log
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch
import javax.inject.Inject

@HiltViewModel
class HomeViewModel @Inject constructor(
    private val workerRepository: WorkerRepository,
    private val policyRepository: PolicyRepository,
    private val earningsRepository: EarningsRepository
) : ViewModel() {

    companion object {
        private const val TAG = "HomeViewModel"
    }

    private val _uiState = MutableStateFlow<HomeUiState>(HomeUiState.Loading)
    val uiState = _uiState.asStateFlow()

    private val _isRefreshing = MutableStateFlow(false)
    val isRefreshing = _isRefreshing.asStateFlow()

    private val _isOnline = MutableStateFlow(false)
    val isOnline = _isOnline.asStateFlow()

    private val _workerProfile = MutableStateFlow<WorkerProfile?>(null)

    init {
        viewModelScope.launch {
            workerRepository.getProfileFlow().collect { worker ->
                _workerProfile.value = worker
                updateUiState()
            }
        }
        
        loadDashboard()
        startAutoRefresh()
    }

    fun loadDashboard() {
        viewModelScope.launch {
            _uiState.value = HomeUiState.Loading
            fetchData()
        }
    }

    fun refresh() {
        viewModelScope.launch {
            _isRefreshing.value = true
            fetchData()
            delay(500) // Small delay for better UX
            _isRefreshing.value = false
        }
    }

    private var _lastPolicy: Policy? = null
    private var _lastEarnings: Earnings? = null
    private var _lastDisruptionMessage: String? = null
    private var _hasDisruptionAlert: Boolean = false

    private suspend fun fetchData() {
        try {
            // Attempt to fetch from network and sync caches
            try {
                workerRepository.getProfile() // Updates Room
            } catch (e: Exception) {
                Log.w(TAG, "Offline syncing profile: ${e.message}")
            }
            try {
                policyRepository.fetchAndCachePolicy() // Updates DataStore
            } catch (e: Exception) {
                Log.w(TAG, "Offline syncing policy: ${e.message}")
            }

            // For earnings and notifications without caches yet, just try to load
            val policyRes = policyRepository.getPolicyFromCache()
            _lastPolicy = policyRes

            try {
                val earningsRes = earningsRepository.getEarnings()
                if (earningsRes.isSuccessful) {
                    val summary = earningsRes.body()!!
                    _lastEarnings = Earnings(
                        thisWeekActual = summary.thisWeekActual.toDouble(),
                        thisWeekBaseline = summary.thisWeekBaseline.toDouble(),
                        todayEarnings = (summary.todayEarnings ?: 0).toDouble(),
                        protectedIncome = summary.protectedIncome.toDouble(),
                        history = summary.history.map { EarningRecord(it.week, it.actual.toDouble()) }
                    )
                }
            } catch (e: Exception) {
                Log.w(TAG, "Offline syncing earnings: ${e.message}")
            }

            try {
                val notificationsRes = workerRepository.getNotifications()
                if (notificationsRes.isSuccessful) {
                    val latestDisruption = notificationsRes.body()
                        ?.notifications
                        ?.firstOrNull { it.type.equals("disruption_alert", ignoreCase = true) }
                    _hasDisruptionAlert = latestDisruption != null
                    _lastDisruptionMessage = latestDisruption?.body
                }
            } catch (e: Exception) {
                 Log.w(TAG, "Offline syncing notifications: ${e.message}")
            }

            updateUiState()
        } catch (e: Exception) {
            _uiState.value = HomeUiState.Error(e.message ?: "Unknown error")
        }
    }

    private fun updateUiState() {
        val worker = _workerProfile.value
        val policy = _lastPolicy

        if (worker != null && policy != null && _lastEarnings != null) {
            _isOnline.value = worker.isOnline ?: false
            _uiState.value = HomeUiState.Success(
                worker = worker,
                policy = policy,
                earnings = _lastEarnings!!,
                hasDisruptionAlert = _hasDisruptionAlert,
                disruptionMessage = _lastDisruptionMessage,
            )
        } else if (worker != null || policy != null) {
            // We have partial offline data, we can still show a degraded success state
            // but for simplicity we'll only show success if we have earnings too
            if (_lastEarnings == null) {
                // If offline & no earnings cache, mock empty earnings for offline stability
                _lastEarnings = Earnings(0.0, 0.0, 0.0, 0.0, null, emptyList())
                updateUiState() // recursive call once to fill mock earnings
            }
        }
    }

    fun toggleOnlineStatus(online: Boolean) {
        viewModelScope.launch {
            val previous = _isOnline.value
            _isOnline.value = online

            try {
                val response = workerRepository.updateOnlineStatus(online)
                if (response.isSuccessful) {
                    _isOnline.value = response.body()?.online ?: online
                    fetchData()
                } else {
                    _isOnline.value = previous
                    Log.w(TAG, "toggleOnlineStatus failed status=${response.code()}")
                }
            } catch (e: Exception) {
                _isOnline.value = previous
                Log.w(TAG, "toggleOnlineStatus exception=${e.message}")
            }
        }
    }

    private fun startAutoRefresh() {
        viewModelScope.launch {
            while (true) {
                delay(12000)
                fetchData()
            }
        }
    }
}

sealed class HomeUiState {
    object Loading : HomeUiState()
    data class Success(
        val worker: WorkerProfile,
        val policy: Policy,
        val earnings: Earnings,
        val hasDisruptionAlert: Boolean,
        val disruptionMessage: String?,
    ) : HomeUiState()
    data class Error(val message: String) : HomeUiState()
}
