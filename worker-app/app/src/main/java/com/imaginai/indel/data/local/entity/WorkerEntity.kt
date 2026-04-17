package com.imaginai.indel.data.local.entity

import androidx.room.Entity
import androidx.room.PrimaryKey
import com.imaginai.indel.data.model.WorkerProfile

@Entity(tableName = "worker_profile")
data class WorkerEntity(
    @PrimaryKey val workerId: String,
    val name: String,
    val phone: String,
    val zone: String?,
    val zoneLevel: String,
    val zoneName: String,
    val zoneId: Int?,
    val city: String?,
    val fromCity: String?,
    val toCity: String?,
    val vehicleType: String,
    val vehicleName: String?,
    val upiId: String,
    val coverageStatus: String,
    val enrolled: Boolean,
    val isOnline: Boolean?,
    val ordersCompleted: Int?,
    val todayEarnings: Int?
) {
    fun toNetworkModel(): WorkerProfile {
        return WorkerProfile(
            workerId = workerId,
            name = name,
            phone = phone,
            zone = zone,
            zoneLevel = zoneLevel,
            zoneName = zoneName,
            zoneId = zoneId,
            city = city,
            fromCity = fromCity,
            toCity = toCity,
            vehicleType = vehicleType,
            vehicleName = vehicleName,
            upiId = upiId,
            coverageStatus = coverageStatus,
            enrolled = enrolled,
            isOnline = isOnline,
            ordersCompleted = ordersCompleted,
            todayEarnings = todayEarnings
        )
    }

    companion object {
        fun fromNetworkModel(model: WorkerProfile): WorkerEntity {
            return WorkerEntity(
                workerId = model.workerId,
                name = model.name,
                phone = model.phone,
                zone = model.zone,
                zoneLevel = model.zoneLevel,
                zoneName = model.zoneName,
                zoneId = model.zoneId,
                city = model.city,
                fromCity = model.fromCity,
                toCity = model.toCity,
                vehicleType = model.vehicleType,
                vehicleName = model.vehicleName,
                upiId = model.upiId,
                coverageStatus = model.coverageStatus,
                enrolled = model.enrolled,
                isOnline = model.isOnline,
                ordersCompleted = model.ordersCompleted,
                todayEarnings = model.todayEarnings
            )
        }
    }
}
