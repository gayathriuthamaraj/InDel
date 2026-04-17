package claimeval

import (
	"context"
	"fmt"
	"strings"
	"testing"
)

func TestEvaluateDetailed_Participation(t *testing.T) {
	const approxGoodSession = 50.0 / 3600.0
	const approxIdleFraudSession = 65.0 / 3600.0

	tests := []struct {
		name             string
		activity         WorkerActivity
		expectedDecision Decision
		expectedScore    float64
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
			name: "Case A2: Good Worker - One Real Order Before Disruption",
			activity: WorkerActivity{
				WorkerID:         11,
				IsOnline:         true,
				ActiveBefore:     true,
				ActiveDuring:     true,
				LoginDuration:    approxGoodSession,
				OrdersAttempted:  0,
				OrdersCompleted:  0,
				EarningsActual:   0,
				EarningsExpected: 150,
			},
			expectedDecision: DecisionApprove,
			expectedScore:    0.10,
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
				ActiveDuring:     true,
				LoginDuration:    approxIdleFraudSession,
				OrdersAttempted:  0,
				EarningsActual:   0,
				EarningsExpected: 100,
			},
			expectedDecision: DecisionReject,
			expectedScore:    0.95,
			containsSignal:   "idle_online_presence",
		},
		{
			name: "Case E: Online Historical Worker - No Live Window Evidence",
			activity: WorkerActivity{
				WorkerID:         6,
				IsOnline:         true,
				ActiveBefore:     true,
				ActiveDuring:     false,
				LoginDuration:    0,
				OrdersAttempted:  0,
				OrdersCompleted:  0,
				EarningsActual:   0,
				EarningsExpected: 100,
			},
			expectedDecision: DecisionDelay,
			containsReason:   "manual review required: worker is online but has no live order/login evidence in the disruption window",
			containsSignal:   "no_live_window_activity",
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
			if tt.expectedScore > 0 && outcome.FraudScore != tt.expectedScore {
				t.Errorf("%s: expected fraud score %.2f, got %.2f", tt.name, tt.expectedScore, outcome.FraudScore)
			}
			fmt.Printf("--- Test %s ---\nDecision: %s | Score: %.2f | Reasons: %v\n", tt.name, outcome.Decision, outcome.FraudScore, outcome.Reasons)
		})
	}
}
