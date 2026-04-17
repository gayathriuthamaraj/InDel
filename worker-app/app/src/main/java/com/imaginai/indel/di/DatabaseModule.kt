package com.imaginai.indel.di

import android.content.Context
import androidx.room.Room
import com.imaginai.indel.data.local.InDelDatabase
import com.imaginai.indel.data.local.dao.WorkerDao
import dagger.Module
import dagger.Provides
import dagger.hilt.InstallIn
import dagger.hilt.android.qualifiers.ApplicationContext
import dagger.hilt.components.SingletonComponent
import javax.inject.Singleton

@Module
@InstallIn(SingletonComponent::class)
object DatabaseModule {

    @Provides
    @Singleton
    fun provideInDelDatabase(
        @ApplicationContext context: Context
    ): InDelDatabase {
        return Room.databaseBuilder(
            context,
            InDelDatabase::class.java,
            "indel_worker_db"
        ).fallbackToDestructiveMigration().build()
    }

    @Provides
    @Singleton
    fun provideWorkerDao(database: InDelDatabase): WorkerDao {
        return database.workerDao()
    }
}
