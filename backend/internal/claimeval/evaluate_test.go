package claimeval

import (
	"context"
	"fmt"
	"strings"
	"testing"
)

func TestEvaluateDetailed_Participation(t *testing.T) {
	// minLoginDuration = 0.02 (approx 1.2 mins)
	
	tests := []struct {
		name             string
		activity         WorkerActivity
		expectedDecision Decision
		containsReason   string
		containsSignal   string
	}{
		{
			name: "Case A: Good Worker - Online - Active - Historical",
			activity: WorkerActivity{
				WorkerID:         1,
				IsOnline:         true,
				ActiveBefore:     true,
				ActiveDuring:     true,
				LoginDuration:    0.05,
				OrdersAttempted:  1,
				EarningsActual:   50,
				EarningsExpected: 150,
			},
			expectedDecision: DecisionApprove,
		},
		{
			name: "Case B: Bad Worker - Offline - No Activity - No History",
			activity: WorkerActivity{
				WorkerID:         2,
				IsOnline:         false,
				ActiveBefore:     false,
				ActiveDuring:     false,
				LoginDuration:    0,
				OrdersAttempted:  0,
				EarningsActual:   0,
				EarningsExpected: 100,
			},
			expectedDecision: DecisionReject,
			containsReason:   "participation fail: worker was offline and showed no evidence of activity",
		},
		{
			name: "Case D: Idle Frauder - Online - No Activity - No History",
			activity: WorkerActivity{
				WorkerID:         3,
				IsOnline:         true,
				ActiveBefore:     false,
				ActiveDuring:     false,
				LoginDuration:    0,
				OrdersAttempted:  0,
				EarningsActual:   0,
				EarningsExpected: 100,
			},
			expectedDecision: DecisionReject, // 0.45 + 0.50 = 0.95 result
			containsSignal:   "idle_online_presence",
		},
		{
			name: "Storm Safe Net: Offline but Active - Historical",
			activity: WorkerActivity{
				WorkerID:         4,
				IsOnline:         false, // Network drop
				ActiveBefore:     true,
				ActiveDuring:     true,
				LoginDuration:    0.05,
				OrdersAttempted:  1,
				EarningsActual:   50,
				EarningsExpected: 150,
			},
			expectedDecision: DecisionApprove,
		},
		{
			name: "No Loss Check: Active but Over-Earning",
			activity: WorkerActivity{
				WorkerID:         5,
				IsOnline:         true,
				ActiveBefore:     true,
				ActiveDuring:     true,
				LoginDuration:    0.05,
				OrdersAttempted:  5,
				EarningsActual:   200,
				EarningsExpected: 150,
			},
			expectedDecision: DecisionReject,
			containsReason:   "audit fail: claim has no positive income loss",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			outcome := EvaluateDetailed(context.Background(), tt.activity)
			if outcome.Decision != tt.expectedDecision {
				t.Errorf("%s: expected decision %s, got %s (Score: %.2f)", tt.name, tt.expectedDecision, outcome.Decision, outcome.FraudScore)
			}
			if tt.containsReason != "" {
				found := false
				for _, r := range outcome.Reasons {
					if strings.Contains(r, tt.containsReason) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("%s: expected reason containing %q, got %v", tt.name, tt.containsReason, outcome.Reasons)
				}
			}
			if tt.containsSignal != "" {
				found := false
				for _, s := range outcome.Signals {
					if strings.Contains(s.Name, tt.containsSignal) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("%s: expected signal name containing %q, got %v", tt.name, tt.containsSignal, outcome.Signals)
				}
			}
			fmt.Printf("--- Test %s ---\nDecision: %s | Score: %.2f | Reasons: %v\n", tt.name, outcome.Decision, outcome.FraudScore, outcome.Reasons)
		})
	}
}
