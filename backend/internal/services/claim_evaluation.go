package services

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"strings"
	"time"

	"github.com/Shravanthi20/InDel/backend/internal/claimeval"
	"github.com/Shravanthi20/InDel/backend/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func fraudLayerEnabled() bool {
	value := strings.TrimSpace(strings.ToLower(os.Getenv("FRAUD_LAYER_ENABLED")))
	return value == "1" || value == "true" || value == "yes" || value == "on"
}

func (s *CoreOpsService) runPrePayoutEvaluation(ctx context.Context, claim models.Claim, now time.Time) (*claimeval.EvaluationOutcome, error) {
	if s.DB == nil {
		return nil, gorm.ErrInvalidDB
	}

	var disruption models.Disruption
	if claim.DisruptionID != 0 {
		if err := s.DB.WithContext(ctx).First(&disruption, claim.DisruptionID).Error; err != nil && err != gorm.ErrRecordNotFound {
			return nil, err
		}
	}

	activity, err := claimeval.AdaptClaimActivity(ctx, s.DB, claimeval.ClaimSource{
		ClaimID:        claim.ID,
		WorkerID:       claim.WorkerID,
		ClaimAmount:    claim.ClaimAmount,
		DisruptionID:   claim.DisruptionID,
		DisruptionType: disruption.Type,
		ZoneID:         disruption.ZoneID,
		StartTime:      disruption.StartTime,
		EndTime:        disruption.EndTime,
		ConfirmedAt:    disruption.ConfirmedAt,
		Now:            now,
	})
	if err != nil {
		return nil, err
	}

	outcome := claimeval.EvaluateDetailed(ctx, activity)
	if err := s.persistClaimEvaluation(ctx, claim, outcome, now); err != nil {
		return nil, err
	}
	return &outcome, nil
}

func (s *CoreOpsService) persistClaimEvaluation(ctx context.Context, claim models.Claim, outcome claimeval.EvaluationOutcome, now time.Time) error {
	if s.DB == nil {
		return gorm.ErrInvalidDB
	}

	claimStatus, claimVerdict, fraudVerdict := mapDecisionToStatuses(outcome.Decision)
	signalsJSON, err := json.Marshal(signalsForPersistence(outcome))
	if err != nil {
		return err
	}

	scoreRecord := models.ClaimFraudScore{
		ClaimID:              claim.ID,
		IsolationForestScore: round2(outcome.FraudScore),
		DbscanScore:          round2(outcome.FraudScore),
		FinalVerdict:         fraudVerdict,
		RuleViolations:       string(signalsJSON),
		CreatedAt:            now.UTC(),
	}

	return s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.Claim{}).
			Where("id = ?", claim.ID).
			Updates(map[string]any{
				"status":        claimStatus,
				"fraud_verdict": claimVerdict,
				"updated_at":    now.UTC(),
			}).Error; err != nil {
			return err
		}

		if err := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "claim_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"isolation_forest_score", "dbscan_score", "final_verdict", "rule_violations"}),
		}).Create(&scoreRecord).Error; err != nil {
			return err
		}

		auditNotes, _ := json.Marshal(map[string]any{
			"decision":    outcome.Decision,
			"fraud_score": round2(outcome.FraudScore),
			"loss":        round2(outcome.Loss),
			"reasons":     outcome.Reasons,
			"fallback":    outcome.Fallback,
		})
		return tx.Create(&models.ClaimAuditLog{
			ClaimID:   claim.ID,
			Action:    "claim_evaluated",
			Notes:     string(auditNotes),
			Reviewer:  "system",
			CreatedAt: now.UTC(),
		}).Error
	})
}

func mapDecisionToStatuses(decision claimeval.Decision) (claimStatus string, claimVerdict string, scoreVerdict string) {
	switch decision {
	case claimeval.DecisionDelay:
		return "manual_review", "review", "manual_review"
	case claimeval.DecisionReject:
		return "rejected", "fraud", "flagged"
	default:
		return "approved", "clear", "clear"
	}
}

func signalsForPersistence(outcome claimeval.EvaluationOutcome) []claimeval.FraudSignal {
	if len(outcome.Signals) > 0 {
		return outcome.Signals
	}

	impact := 0.25
	switch outcome.Decision {
	case claimeval.DecisionDelay:
		impact = 0.55
	case claimeval.DecisionReject:
		impact = 0.9
	}
	return []claimeval.FraudSignal{{
		Name:        "claim_decision",
		Impact:      impact,
		Description: strings.Join(outcome.Reasons, "; "),
	}}
}

func decisionToPayoutStatus(decision claimeval.Decision) string {
	switch decision {
	case claimeval.DecisionDelay:
		return "manual_review"
	case claimeval.DecisionReject:
		return "rejected"
	default:
		return "approved"
	}
}

func logFailOpen(claimID uint, err error) {
	log.Printf("[CLAIM-EVAL] fail-open for claimID=%d: %v", claimID, err)
}
