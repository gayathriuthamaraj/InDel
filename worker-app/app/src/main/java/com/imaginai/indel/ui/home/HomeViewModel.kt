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

    init {
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

    private suspend fun fetchData() {
        try {
            val profileRes = workerRepository.getProfile()
            val policyRes = policyRepository.getPolicy()
            val earningsRes = earningsRepository.getEarnings()
            val notificationsRes = workerRepository.getNotifications()

            if (profileRes.isSuccessful && policyRes.isSuccessful && earningsRes.isSuccessful && notificationsRes.isSuccessful) {
                val worker = profileRes.body()!!.worker
                val summary = earningsRes.body()!!
                val earnings = Earnings(
                    thisWeekActual = summary.thisWeekActual.toDouble(),
                    thisWeekBaseline = summary.thisWeekBaseline.toDouble(),
                    todayEarnings = (summary.todayEarnings ?: 0).toDouble(),
                    protectedIncome = summary.protectedIncome.toDouble(),
                    history = summary.history.map { EarningRecord(it.week, it.actual.toDouble()) }
                )

                val latestDisruption = notificationsRes.body()
                    ?.notifications
                    ?.firstOrNull { it.type.equals("disruption_alert", ignoreCase = true) }

                _isOnline.value = worker.isOnline ?: false
                
                _uiState.value = HomeUiState.Success(
                    worker = worker,
                    policy = policyRes.body()!!.policy,
                    earnings = earnings,
                    hasDisruptionAlert = latestDisruption != null,
                    disruptionMessage = latestDisruption?.body,
                )
            } else {
                _uiState.value = HomeUiState.Error("Failed to load dashboard data")
            }
        } catch (e: Exception) {
            _uiState.value = HomeUiState.Error(e.message ?: "Unknown error")
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
