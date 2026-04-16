package claimeval

import (
	"context"
	"testing"
	"time"
)

func TestAdaptClaimActivityAppliesStaleness(t *testing.T) {
	// Mock a claim source where worker is "Online" but LastActiveAt is very old
	isOnline := true
	lastActiveAt := time.Now().Add(-20 * time.Minute)
	
	source := ClaimSource{
		WorkerID:     1,
		ZoneID:       1,
		IsOnline:     &isOnline,
		LastActiveAt: &lastActiveAt,
	}

	// In a real environment, AdaptClaimActivity might query DB.
	// Since we provided IsOnline and LastActiveAt in the source, it skips the query.
	// But it SHOULD apply the staleness check regardless of source.
	
	activity, err := AdaptClaimActivity(context.Background(), nil, source)
	if err != nil {
		// Mock DB is nil, but it should skip query if pointers are provided
		t.Fatalf("AdaptClaimActivity failed: %v", err)
	}

	if activity.IsOnline {
		t.Errorf("expected worker to be marked offline due to 20m staleness, but was online")
	}
}

func TestAdaptClaimActivityAcceptsFreshActivity(t *testing.T) {
	isOnline := true
	lastActiveAt := time.Now().Add(-5 * time.Minute)
	
	source := ClaimSource{
		WorkerID:     1,
		ZoneID:       1,
		IsOnline:     &isOnline,
		LastActiveAt: &lastActiveAt,
	}
	
	activity, err := AdaptClaimActivity(context.Background(), nil, source)
	if err != nil {
		t.Fatalf("AdaptClaimActivity failed: %v", err)
	}

	if !activity.IsOnline {
		t.Errorf("expected worker to be online with 5m staleness, but was offline")
	}
}
