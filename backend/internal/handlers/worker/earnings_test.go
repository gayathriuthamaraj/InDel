package worker

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

func setupWorkerEarningsTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "worker_earnings_test.db")
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open sqlite db: %v", err)
	}

	statements := []string{
		`CREATE TABLE auth_tokens (user_id INTEGER, token TEXT, expires_at DATETIME)`,
		`CREATE TABLE earnings_baseline (worker_id INTEGER, baseline_amount REAL)`,
		`CREATE TABLE weekly_earnings_summary (worker_id INTEGER, week_start DATE, week_end DATE, total_earnings REAL)`,
		`CREATE TABLE earnings_records (worker_id INTEGER, date DATE, amount_earned REAL)`,
		`CREATE TABLE claims (id INTEGER PRIMARY KEY, worker_id INTEGER, claim_amount REAL, status TEXT)`,
		`CREATE TABLE payouts (id INTEGER PRIMARY KEY, claim_id INTEGER, worker_id INTEGER, amount REAL, status TEXT)`,
	}
	for _, stmt := range statements {
		if err := db.Exec(stmt).Error; err != nil {
			t.Fatalf("failed to create schema: %v", err)
		}
	}

	return db
}

func TestGetEarningsUsesProcessedPayoutsForProtectedIncome(t *testing.T) {
	db := setupWorkerEarningsTestDB(t)
	SetDB(db)
	t.Cleanup(func() {
		SetDB(nil)
	})

	token := "mock-jwt-token"
	now := time.Now().UTC()
	weekStart := now.AddDate(0, 0, -1).Format("2006-01-02")
	weekEnd := now.AddDate(0, 0, 1).Format("2006-01-02")
	today := now.Format("2006-01-02")
	expiresAt := now.Add(24 * time.Hour).Format(time.RFC3339)

	inserts := []struct {
		query string
		args  []any
	}{
		{query: `INSERT INTO auth_tokens (user_id, token, expires_at) VALUES (?, ?, ?)`, args: []any{1, token, expiresAt}},
		{query: `INSERT INTO earnings_baseline (worker_id, baseline_amount) VALUES (?, ?)`, args: []any{1, 5000}},
		{query: `INSERT INTO weekly_earnings_summary (worker_id, week_start, week_end, total_earnings) VALUES (?, ?, ?, ?)`, args: []any{1, weekStart, weekEnd, 1200}},
		{query: `INSERT INTO earnings_records (worker_id, date, amount_earned) VALUES (?, ?, ?)`, args: []any{1, today, 0}},
		{query: `INSERT INTO claims (id, worker_id, claim_amount, status) VALUES (?, ?, ?, ?)`, args: []any{10, 1, 346, "approved"}},
		{query: `INSERT INTO payouts (id, claim_id, worker_id, amount, status) VALUES (?, ?, ?, ?, ?)`, args: []any{20, 10, 1, 0, "queued"}},
		{query: `INSERT INTO payouts (id, claim_id, worker_id, amount, status) VALUES (?, ?, ?, ?, ?)`, args: []any{21, 11, 1, 125, "processed"}},
	}
	for _, insert := range inserts {
		if err := db.Exec(insert.query, insert.args...).Error; err != nil {
			t.Fatalf("failed to seed earnings data: %v", err)
		}
	}

	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	req, err := http.NewRequest(http.MethodGet, "/api/v1/worker/earnings", nil)
	if err != nil {
		t.Fatalf("failed to build request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	ctx.Request = req

	GetEarnings(ctx)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}

	var response struct {
		ProtectedIncome int `json:"protected_income"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.ProtectedIncome != 125 {
		t.Fatalf("protected_income = %d, want 125 from processed payouts only", response.ProtectedIncome)
	}
}
