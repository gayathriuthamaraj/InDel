package claimeval

import (
	"context"
	"strconv"
	"testing"
)

func TestAdaptSyntheticActivityDerivesMissingFields(t *testing.T) {
	completed := 3
	activity, err := AdaptWorkerActivity(context.Background(), AdaptSyntheticActivity, SyntheticSource{
		WorkerID:        77,
		Zone:            "Test Zone",
		OrdersCompleted: &completed,
	})
	if err != nil {
		t.Fatalf("adapt synthetic activity: %v", err)
	}

	if activity.OrdersAttempted != 3 {
		t.Fatalf("orders attempted should be derived from completions, got %d", activity.OrdersAttempted)
	}
	if activity.LoginDuration <= 0 {
		t.Fatalf("login duration should be derived, got %.2f", activity.LoginDuration)
	}
	if activity.EarningsExpected <= 0 {
		t.Fatalf("earnings expected should be derived, got %.2f", activity.EarningsExpected)
	}
	if !activity.ActiveDuring {
		t.Fatal("active_during should be derived as true")
	}
}

func TestEvaluateDetailedRejectsInactiveWorker(t *testing.T) {
	raw, err := SyntheticScenario("lazy_fraud")
	if err != nil {
		t.Fatalf("synthetic scenario: %v", err)
	}
	activity, err := AdaptWorkerActivity(context.Background(), AdaptSyntheticActivity, raw)
	if err != nil {
		t.Fatalf("adapt activity: %v", err)
	}

	outcome := EvaluateDetailed(context.Background(), activity)
	if outcome.Decision != DecisionReject {
		t.Fatalf("expected reject, got %+v", outcome)
	}
	if outcome.Eligible {
		t.Fatalf("inactive worker should not be eligible: %+v", outcome)
	}
}

func TestEvaluateDetailedUsesFraudThresholds(t *testing.T) {
	tests := []struct {
		name     string
		score    float64
		expected Decision
	}{
		{name: "approve", score: 0.20, expected: DecisionApprove},
		{name: "delay", score: 0.45, expected: DecisionDelay},
		{name: "reject", score: 0.85, expected: DecisionReject},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("FRAUD_ML_MOCK_SCORE", formatScore(tt.score))
			t.Setenv("FRAUD_ML_MOCK_ERROR", "")

			raw, err := SyntheticScenario("genuine_worker")
			if err != nil {
				t.Fatalf("synthetic scenario: %v", err)
			}
			activity, err := AdaptWorkerActivity(context.Background(), AdaptSyntheticActivity, raw)
			if err != nil {
				t.Fatalf("adapt activity: %v", err)
			}

			outcome := EvaluateDetailed(context.Background(), activity)
			if outcome.Decision != tt.expected {
				t.Fatalf("expected %s, got %+v", tt.expected, outcome)
			}
			if outcome.FraudScore != tt.score {
				t.Fatalf("expected fraud score %.2f, got %.2f", tt.score, outcome.FraudScore)
			}
		})
	}
}

func TestEvaluateDetailedFallsBackToLowRiskWhenFraudServiceFails(t *testing.T) {
	t.Setenv("FRAUD_ML_MOCK_ERROR", "true")
	t.Setenv("FRAUD_ML_MOCK_SCORE", "")

	raw, err := SyntheticScenario("genuine_worker")
	if err != nil {
		t.Fatalf("synthetic scenario: %v", err)
	}
	activity, err := AdaptWorkerActivity(context.Background(), AdaptSyntheticActivity, raw)
	if err != nil {
		t.Fatalf("adapt activity: %v", err)
	}

	outcome := EvaluateDetailed(context.Background(), activity)
	if outcome.Decision != DecisionApprove {
		t.Fatalf("fallback should fail open to approve, got %+v", outcome)
	}
	if !outcome.Fallback {
		t.Fatalf("expected fallback marker, got %+v", outcome)
	}
	if outcome.FraudScore != fallbackFraudScore {
		t.Fatalf("expected fallback score %.2f, got %.2f", fallbackFraudScore, outcome.FraudScore)
	}
}

func formatScore(value float64) string {
	return strconv.FormatFloat(value, 'f', 2, 64)
}
