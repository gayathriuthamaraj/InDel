package claimeval

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupClaimActivityTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "claim_activity_test.db")
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open sqlite db: %v", err)
	}

	statements := []string{
		`CREATE TABLE orders (
			id INTEGER PRIMARY KEY,
			worker_id INTEGER,
			created_at DATETIME
		)`,
		`CREATE TABLE earnings_records (
			id INTEGER PRIMARY KEY,
			worker_id INTEGER,
			date DATE,
			hours_worked REAL
		)`,
	}
	for _, stmt := range statements {
		if err := db.Exec(stmt).Error; err != nil {
			t.Fatalf("failed to create test schema: %v", err)
		}
	}

	return db
}

func TestAdaptClaimActivityAppliesStaleness(t *testing.T) {
	db := setupClaimActivityTestDB(t)
	isOnline := true
	lastActiveAt := time.Now().Add(-20 * time.Minute)
	baselineAmount := 1000.0

	source := ClaimSource{
		WorkerID:       1,
		ZoneID:         1,
		IsOnline:       &isOnline,
		LastActiveAt:   &lastActiveAt,
		BaselineAmount: &baselineAmount,
	}

	activity, err := AdaptClaimActivity(context.Background(), db, source)
	if err != nil {
		t.Fatalf("AdaptClaimActivity failed: %v", err)
	}

	if activity.IsOnline {
		t.Errorf("expected worker to be marked offline due to 20m staleness, but was online")
	}
}

func TestAdaptClaimActivityAcceptsFreshActivity(t *testing.T) {
	db := setupClaimActivityTestDB(t)
	isOnline := true
	lastActiveAt := time.Now().Add(-5 * time.Minute)
	baselineAmount := 1000.0

	source := ClaimSource{
		WorkerID:       1,
		ZoneID:         1,
		IsOnline:       &isOnline,
		LastActiveAt:   &lastActiveAt,
		BaselineAmount: &baselineAmount,
	}

	activity, err := AdaptClaimActivity(context.Background(), db, source)
	if err != nil {
		t.Fatalf("AdaptClaimActivity failed: %v", err)
	}

	if !activity.IsOnline {
		t.Errorf("expected worker to be online with 5m staleness, but was offline")
	}
}

func TestAdaptClaimActivityDerivesSessionEvidenceFromRecentPresence(t *testing.T) {
	db := setupClaimActivityTestDB(t)
	now := time.Now().UTC()
	isOnline := true
	lastActiveAt := now.Add(-65 * time.Second)
	baselineAmount := 1000.0

	source := ClaimSource{
		WorkerID:       1,
		ZoneID:         1,
		IsOnline:       &isOnline,
		LastActiveAt:   &lastActiveAt,
		BaselineAmount: &baselineAmount,
		StartTime:      &now,
		Now:            now,
	}

	activity, err := AdaptClaimActivity(context.Background(), db, source)
	if err != nil {
		t.Fatalf("AdaptClaimActivity failed: %v", err)
	}

	if activity.ActiveBefore {
		t.Fatalf("recent presence alone should not count as historical participation: %+v", activity)
	}

	if activity.LoginDuration < minLoginEvidenceHours {
		t.Fatalf("expected login duration evidence from recent session, got %.4f", activity.LoginDuration)
	}
}
