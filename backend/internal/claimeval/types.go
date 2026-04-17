package claimeval

import "time"

// Decision is the deterministic outcome returned by the pre-payout gate.
type Decision string

const (
	DecisionApprove Decision = "APPROVE"
	DecisionDelay   Decision = "DELAY"
	DecisionReject  Decision = "REJECT"
)

// WorkerActivity is the strict contract consumed by the evaluation layer.
// Core decision logic depends only on these fields.
type WorkerActivity struct {
	WorkerID         uint    `json:"worker_id"`
	Zone             string  `json:"zone"`
	SegmentID        string  `json:"segment_id"`
	IsOnline         bool    `json:"is_online"`
	ActiveBefore     bool    `json:"active_before"`
	ActiveDuring     bool    `json:"active_during"`
	LoginDuration    float64 `json:"login_duration"`
	OrdersAttempted  int     `json:"orders_attempted"`
	OrdersCompleted  int     `json:"orders_completed"`
	EarningsActual   float64 `json:"earnings_actual"`
	EarningsExpected float64 `json:"earnings_expected"`
}

// EvaluationOutcome contains the detailed result used by the integration layer.
type EvaluationOutcome struct {
	Decision   Decision      `json:"decision"`
	Eligible   bool          `json:"eligible"`
	Loss       float64       `json:"loss"`
	FraudScore float64       `json:"fraud_score"`
	Reasons    []string      `json:"reasons"`
	Signals    []FraudSignal `json:"signals"`
	Fallback   bool          `json:"fallback"`
}

// FraudSignal is a compact explanation item returned by the fraud service
// or generated locally when eligibility/loss rules short-circuit.
type FraudSignal struct {
	Name        string  `json:"name"`
	Impact      float64 `json:"impact"`
	Description string  `json:"description,omitempty"`
}

// ClaimSource is the adapter input used by the live payout gate.
type ClaimSource struct {
	ClaimID        uint
	WorkerID       uint
	ClaimAmount    float64
	DisruptionID   uint
	DisruptionType string
	ZoneID         uint
	StartTime      *time.Time
	EndTime        *time.Time
	ConfirmedAt    *time.Time
	Now            time.Time

	// Optional pre-fetched data from batch queries
	IsOnline       *bool
	LastActiveAt   *time.Time
	BaselineAmount *float64
	ActualEarnings *float64
}

// SyntheticSource is a partially filled source used by synthetic examples/tests.
// Missing fields are derived by the adapter.
type SyntheticSource struct {
	WorkerID         uint
	Zone             string
	SegmentID        string
	ActiveBefore     *bool
	ActiveDuring     *bool
	LoginDuration    *float64
	OrdersAttempted  *int
	OrdersCompleted  *int
	EarningsActual   *float64
	EarningsExpected *float64
}
