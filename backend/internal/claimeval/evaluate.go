package claimeval

import (
	"context"
	"fmt"
	"math"
)

// EvaluateClaim is the single public function exposed by the evaluation layer.
// It returns only the decision, while integration code can use EvaluateDetailed.
func EvaluateClaim(activity WorkerActivity) Decision {
	return EvaluateDetailed(context.Background(), activity).Decision
}

// EvaluateDetailed runs the deterministic pre-payout gate and returns the
// supporting score/reasons needed by the existing claim pipeline.
func EvaluateDetailed(ctx context.Context, activity WorkerActivity) EvaluationOutcome {
	normalized := normalizeActivity(activity)
	outcome := EvaluationOutcome{
		Decision: DecisionApprove,
		Reasons:  make([]string, 0, 4),
		Signals:  make([]FraudSignal, 0, 4),
	}

	// 1. Evidence Threshold (Standardized)
	const minEvidenceThreshold = 0.02
	hasEvidence := normalized.LoginDuration >= minEvidenceThreshold || normalized.OrdersAttempted >= 1

	// 2. Participation Gate
	if !normalized.IsOnline {
		if !hasEvidence {
			outcome.Decision = DecisionReject
			outcome.Reasons = append(outcome.Reasons, "participation fail: worker was offline and showed no evidence of activity")
			return outcome
		}
		if !normalized.ActiveBefore && normalized.OrdersAttempted == 0 {
			outcome.Decision = DecisionReject
			outcome.Reasons = append(outcome.Reasons, "participation fail: offline worker with weak evidence and no historical activity")
			return outcome
		}
	}

	outcome.Eligible = true

	// 3. Financial Gate
	loss := ComputeLoss(normalized)
	outcome.Loss = loss
	if loss <= 0 {
		outcome.Decision = DecisionReject
		outcome.Reasons = append(outcome.Reasons, "audit fail: claim has no positive income loss")
		outcome.Signals = append(outcome.Signals, FraudSignal{Name: "no_income_loss", Impact: 1.0})
		return outcome
	}

	// 4. Signal Stacking (Static Evaluation)
	var stackedImpact float64
	if normalized.IsOnline && !hasEvidence {
		signal := FraudSignal{Name: "idle_online_presence", Impact: 0.45}
		outcome.Signals = append(outcome.Signals, signal)
		stackedImpact += signal.Impact
	}
	if !normalized.ActiveBefore && !hasEvidence {
		signal := FraudSignal{Name: "zero_historical_participation", Impact: 0.50}
		outcome.Signals = append(outcome.Signals, signal)
		stackedImpact += signal.Impact
	}

	// 5. ML Scoring (SINGLE CALL)
	baseMLScore, mlSignals, fallback := FetchFraudScore(ctx, normalized)
	
	// Final Combined Score
	combinedScore := math.Min(1.0, baseMLScore+stackedImpact)
	
	outcome.FraudScore = combinedScore
	outcome.Signals = append(outcome.Signals, mlSignals...)
	outcome.Fallback = fallback

	// 6. Decision logic (ONE SCORE ONLY)
	if combinedScore <= 0.3 {
		outcome.Decision = DecisionApprove
		outcome.Reasons = append(outcome.Reasons, "fraud score within auto-approve threshold")
	} else if combinedScore <= 0.7 {
		outcome.Decision = DecisionDelay
		outcome.Reasons = append(outcome.Reasons, "fraud score requires delayed/manual review")
	} else {
		outcome.Decision = DecisionReject
		outcome.Reasons = append(outcome.Reasons, "fraud score exceeds rejection threshold")
	}

	// 7. Story-Telling Logs
	fmt.Printf("\n[CLAIM EVAL] worker: %d | intent: %v | evidence: %v | history: %v | score: %.2f | decision: %s\n", 
		normalized.WorkerID, normalized.IsOnline, hasEvidence, normalized.ActiveBefore, combinedScore, outcome.Decision)

	return outcome
}

func IsEligible(activity WorkerActivity) bool {
	// minLoginDuration = 0.02 is the demo standard for "Participation"
	hasEvidence := activity.LoginDuration >= 0.02 || activity.OrdersAttempted >= 1
	return hasEvidence && activity.ActiveBefore
}

func ComputeLoss(activity WorkerActivity) float64 {
	return round2(maxFloat(activity.EarningsExpected-activity.EarningsActual, 0))
}
