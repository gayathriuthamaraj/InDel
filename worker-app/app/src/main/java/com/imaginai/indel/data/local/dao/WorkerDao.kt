package com.imaginai.indel.data.local.dao

import androidx.room.Dao
import androidx.room.Insert
import androidx.room.OnConflictStrategy
import androidx.room.Query
import com.imaginai.indel.data.local.entity.WorkerEntity
import kotlinx.coroutines.flow.Flow

@Dao
interface WorkerDao {
    @Query("SELECT * FROM worker_profile LIMIT 1")
    fun getWorkerProfileFlow(): Flow<List<WorkerEntity>>

    @Query("SELECT * FROM worker_profile LIMIT 1")
    suspend fun getWorkerProfile(): List<WorkerEntity>

    @Insert(onConflict = OnConflictStrategy.REPLACE)
    suspend fun insertWorkerProfile(worker: WorkerEntity)

    @Query("DELETE FROM worker_profile")
    suspend fun clearProfile()
}
