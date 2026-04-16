package claimeval

import (
	"context"
	"fmt"
	"math"
)

const (
	minLoginEvidenceHours          = 45.0 / 3600.0
	idleOnlinePresenceImpact       = 0.45
	zeroHistoryParticipationImpact = 0.40
	presenceOnlyFraudImpact        = 0.85
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

	hasSessionEvidence := normalized.LoginDuration >= minLoginEvidenceHours
	hasLiveOrderEvidence := normalized.OrdersAttempted >= 1 || normalized.OrdersCompleted >= 1
	hasEvidence := hasSessionEvidence || hasLiveOrderEvidence

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

	// Workers with historical participation but no live disruption-window evidence
	// should be reviewed manually before any automatic payout is queued.
	if normalized.IsOnline && !hasEvidence && normalized.ActiveBefore {
		outcome.Decision = DecisionDelay
		outcome.FraudScore = 0.55
		outcome.Reasons = append(outcome.Reasons, "manual review required: worker is online but has no live order/login evidence in the disruption window")
		outcome.Signals = append(outcome.Signals, FraudSignal{Name: "no_live_window_activity", Impact: 0.55})
		return outcome
	}

	// 4. Signal Stacking (Static Evaluation)
	var stackedImpact float64
	// Presence-only evidence from a fresh account is the core "idle frauder" case:
	// an online session exists, but there is no real pre-disruption work history and
	// no order attempts to support the claim.
	if normalized.IsOnline && hasSessionEvidence && !hasLiveOrderEvidence && !normalized.ActiveBefore {
		signal := FraudSignal{Name: "idle_online_presence", Impact: presenceOnlyFraudImpact}
		outcome.Signals = append(outcome.Signals, signal)
		stackedImpact += signal.Impact
	} else {
		if normalized.IsOnline && !hasEvidence && !normalized.ActiveBefore {
			signal := FraudSignal{Name: "idle_online_presence", Impact: idleOnlinePresenceImpact}
			outcome.Signals = append(outcome.Signals, signal)
			stackedImpact += signal.Impact
		}
		if !normalized.ActiveBefore && !hasEvidence {
			signal := FraudSignal{Name: "zero_historical_participation", Impact: zeroHistoryParticipationImpact}
			outcome.Signals = append(outcome.Signals, signal)
			stackedImpact += signal.Impact
		}
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
	hasEvidence := activity.LoginDuration >= minLoginEvidenceHours || activity.OrdersAttempted >= 1
	return hasEvidence && activity.ActiveBefore
}

func ComputeLoss(activity WorkerActivity) float64 {
	return round2(maxFloat(activity.EarningsExpected-activity.EarningsActual, 0))
}
