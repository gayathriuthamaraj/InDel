package platform

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupWorkersTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "workers_test.db")
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open sqlite db: %v", err)
	}

	statements := []string{
		`CREATE TABLE users (id INTEGER PRIMARY KEY, phone TEXT, role TEXT)`,
		`CREATE TABLE zones (id INTEGER PRIMARY KEY, name TEXT, city TEXT)`,
		`CREATE TABLE worker_profiles (
			id INTEGER PRIMARY KEY,
			worker_id INTEGER,
			name TEXT,
			zone_id INTEGER,
			is_online BOOLEAN,
			last_active_at DATETIME
		)`,
	}
	for _, stmt := range statements {
		if err := db.Exec(stmt).Error; err != nil {
			t.Fatalf("failed to create schema: %v", err)
		}
	}

	return db
}

func TestGetWorkersAppliesSharedStalenessThreshold(t *testing.T) {
	db := setupWorkersTestDB(t)
	SetDB(db)

	now := time.Now().UTC()

	inserts := []struct {
		query string
		args  []any
	}{
		{query: `INSERT INTO zones (id, name, city) VALUES (?, ?, ?)`, args: []any{1, "Tambaram", "Chennai"}},
		{query: `INSERT INTO users (id, phone, role) VALUES (?, ?, ?)`, args: []any{1, "+911111111111", "worker"}},
		{query: `INSERT INTO users (id, phone, role) VALUES (?, ?, ?)`, args: []any{2, "+922222222222", "worker"}},
		{query: `INSERT INTO worker_profiles (id, worker_id, name, zone_id, is_online, last_active_at) VALUES (?, ?, ?, ?, ?, ?)`, args: []any{1, 1, "Stale Worker", 1, true, now.Add(-20 * time.Minute)}},
		{query: `INSERT INTO worker_profiles (id, worker_id, name, zone_id, is_online, last_active_at) VALUES (?, ?, ?, ?, ?, ?)`, args: []any{2, 2, "Fresh Worker", 1, true, now.Add(-5 * time.Minute)}},
	}
	for _, insert := range inserts {
		if err := db.Exec(insert.query, insert.args...).Error; err != nil {
			t.Fatalf("failed to seed worker data: %v", err)
		}
	}

	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	req, err := http.NewRequest(http.MethodGet, "/api/v1/platform/workers", nil)
	if err != nil {
		t.Fatalf("failed to build request: %v", err)
	}
	ctx.Request = req

	GetWorkers(ctx)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}

	var response struct {
		Workers []struct {
			WorkerID uint   `json:"worker_id"`
			IsOnline bool   `json:"is_online"`
			Status   string `json:"status"`
		} `json:"workers"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	statusByWorkerID := make(map[uint]struct {
		IsOnline bool
		Status   string
	}, len(response.Workers))
	for _, worker := range response.Workers {
		statusByWorkerID[worker.WorkerID] = struct {
			IsOnline bool
			Status   string
		}{
			IsOnline: worker.IsOnline,
			Status:   worker.Status,
		}
	}

	if stale := statusByWorkerID[1]; stale.IsOnline || stale.Status != "offline" {
		t.Fatalf("expected stale worker to be offline, got online=%v status=%q", stale.IsOnline, stale.Status)
	}

	if fresh := statusByWorkerID[2]; !fresh.IsOnline || fresh.Status != "live" {
		t.Fatalf("expected fresh worker to be live, got online=%v status=%q", fresh.IsOnline, fresh.Status)
	}
}
