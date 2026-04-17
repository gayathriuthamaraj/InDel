package com.imaginai.indel.data.local

import androidx.room.Database
import androidx.room.RoomDatabase
import com.imaginai.indel.data.local.dao.WorkerDao
import com.imaginai.indel.data.local.entity.WorkerEntity

@Database(entities = [WorkerEntity::class], version = 1, exportSchema = false)
abstract class InDelDatabase : RoomDatabase() {
    abstract fun workerDao(): WorkerDao
}
