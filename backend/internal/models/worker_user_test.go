package models

import (
	"testing"
	"time"
)

func TestEffectiveWorkerOnlineStatusUsesFifteenMinuteThreshold(t *testing.T) {
	now := time.Date(2026, time.April, 16, 12, 0, 0, 0, time.UTC)

	if !EffectiveWorkerOnlineStatus(true, now.Add(-14*time.Minute-59*time.Second), now) {
		t.Fatalf("expected worker to remain online within the 15-minute threshold")
	}

	if EffectiveWorkerOnlineStatus(true, now.Add(-16*time.Minute), now) {
		t.Fatalf("expected worker to be marked offline after the 15-minute threshold")
	}
}

func TestIsWorkerStatusStaleIgnoresOfflineAndZeroActivity(t *testing.T) {
	now := time.Date(2026, time.April, 16, 12, 0, 0, 0, time.UTC)

	if IsWorkerStatusStale(false, now.Add(-24*time.Hour), now) {
		t.Fatalf("expected offline workers to remain non-stale for status evaluation")
	}

	if IsWorkerStatusStale(true, time.Time{}, now) {
		t.Fatalf("expected zero last-active timestamps to avoid false staleness")
	}
}
