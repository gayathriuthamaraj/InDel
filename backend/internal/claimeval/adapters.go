package claimeval

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/Shravanthi20/InDel/backend/internal/models"
	"gorm.io/gorm"
)

// ActivityAdapter is a pluggable adapter that maps a source into WorkerActivity.
type ActivityAdapter func(context.Context, any) (WorkerActivity, error)

type orderSchemaProbe struct {
	AcceptedAt     *time.Time `gorm:"column:accepted_at"`
	DeliveredAt    *time.Time `gorm:"column:delivered_at"`
	Status         string     `gorm:"column:status"`
	UpdatedAt      *time.Time `gorm:"column:updated_at"`
	DeliveryFeeINR float64    `gorm:"column:delivery_fee_inr"`
}

func (orderSchemaProbe) TableName() string { return "orders" }

// AdaptWorkerActivity normalizes the output of any adapter into the strict contract.
func AdaptWorkerActivity(ctx context.Context, adapter ActivityAdapter, source any) (WorkerActivity, error) {
	if adapter == nil {
		return WorkerActivity{}, fmt.Errorf("activity adapter is required")
	}
	activity, err := adapter(ctx, source)
	if err != nil {
		return WorkerActivity{}, err
	}
	return normalizeActivity(activity), nil
}

// AdaptSyntheticActivity converts a partial synthetic source into WorkerActivity.
func AdaptSyntheticActivity(_ context.Context, source any) (WorkerActivity, error) {
	raw, ok := source.(SyntheticSource)
	if !ok {
		return WorkerActivity{}, fmt.Errorf("expected SyntheticSource, got %T", source)
	}

	activity := WorkerActivity{
		WorkerID:  raw.WorkerID,
		Zone:      strings.TrimSpace(raw.Zone),
		SegmentID: strings.TrimSpace(raw.SegmentID),
	}

	if raw.ActiveBefore != nil {
		activity.ActiveBefore = *raw.ActiveBefore
	}
	if raw.ActiveDuring != nil {
		activity.ActiveDuring = *raw.ActiveDuring
	}
	if raw.LoginDuration != nil {
		activity.LoginDuration = *raw.LoginDuration
	}
	if raw.OrdersAttempted != nil {
		activity.OrdersAttempted = *raw.OrdersAttempted
	}
	if raw.OrdersCompleted != nil {
		activity.OrdersCompleted = *raw.OrdersCompleted
	}
	if raw.EarningsActual != nil {
		activity.EarningsActual = *raw.EarningsActual
	}
	if raw.EarningsExpected != nil {
		activity.EarningsExpected = *raw.EarningsExpected
	}

	return activity, nil
}

// AdaptClaimActivity loads live worker/order data and derives the strict contract.
// The adapter touches raw tables so the core engines never need direct order access.
func AdaptClaimActivity(ctx context.Context, db *gorm.DB, source ClaimSource) (WorkerActivity, error) {
	windowStart, windowEnd := disruptionWindow(source)
	preStart := windowStart.Add(-7 * 24 * time.Hour)

	type workerRow struct {
		ZoneName       string     `gorm:"column:zone_name"`
		ZoneCity       string     `gorm:"column:zone_city"`
		ZoneLevel      string     `gorm:"column:zone_level"`
		VehicleType    string     `gorm:"column:vehicle_type"`
		BaselineAmount float64    `gorm:"column:baseline_amount"`
		IsOnline       bool       `gorm:"column:is_online"`
		LastActiveAt   *time.Time `gorm:"column:last_active_at"`
	}

	row := workerRow{
		ZoneName:       fmt.Sprintf("zone-%d", source.ZoneID),
		VehicleType:    "two_wheeler",
		BaselineAmount: maxFloat(source.ClaimAmount/0.85, source.ClaimAmount),
	}

	// Use pre-fetched data if available to avoid redundant queries
	if source.IsOnline != nil {
		row.IsOnline = *source.IsOnline
	}
	if source.LastActiveAt != nil {
		row.LastActiveAt = source.LastActiveAt
	}
	if source.BaselineAmount != nil {
		row.BaselineAmount = *source.BaselineAmount
	}

	needsQuery := source.IsOnline == nil || source.LastActiveAt == nil || source.BaselineAmount == nil
	if needsQuery {
		if db == nil {
			return WorkerActivity{}, fmt.Errorf("db unavailable")
		}
		_ = db.WithContext(ctx).
			Table("worker_profiles wp").
			Select(`
				COALESCE(z.name, ?) AS zone_name,
				COALESCE(z.city, '') AS zone_city,
				COALESCE(z.level, 'B') AS zone_level,
				COALESCE(wp.vehicle_type, 'two_wheeler') AS vehicle_type,
				COALESCE(eb.baseline_amount, 0) AS baseline_amount,
				COALESCE(wp.is_online, true) AS is_online,
				wp.last_active_at
			`, row.ZoneName).
			Joins("LEFT JOIN zones z ON z.id = wp.zone_id").
			Joins("LEFT JOIN earnings_baseline eb ON eb.worker_id = wp.worker_id").
			Where("wp.worker_id = ?", source.WorkerID).
			Scan(&row).Error
	}

	statusNow := source.Now
	if statusNow.IsZero() {
		statusNow = time.Now()
	}
	lastActiveAt := time.Time{}
	if row.LastActiveAt != nil {
		lastActiveAt = *row.LastActiveAt
	}
	row.IsOnline = models.EffectiveWorkerOnlineStatus(row.IsOnline, lastActiveAt, statusNow)

	attemptTimeExpr := "created_at"
	completeTimeExpr := "created_at"
	statusClause := ""
	if hasOrderColumn(db, "accepted_at") {
		attemptTimeExpr = "COALESCE(accepted_at, created_at)"
	}
	if hasOrderColumn(db, "delivered_at") {
		completeTimeExpr = "COALESCE(delivered_at, accepted_at, created_at)"
	}
	if hasOrderColumn(db, "status") {
		statusClause = " AND LOWER(COALESCE(status, '')) IN ('delivered', 'completed')"
	}

	var ordersBefore int64
	_ = db.WithContext(ctx).
		Table("orders").
		Where("worker_id = ? AND "+attemptTimeExpr+" >= ? AND "+attemptTimeExpr+" < ?", source.WorkerID, preStart, windowStart).
		Count(&ordersBefore).Error

	attemptClause := ""
	if hasOrderColumn(db, "accepted_at") {
		attemptClause = " AND (accepted_at IS NOT NULL OR LOWER(COALESCE(status, '')) NOT IN ('assigned', 'pending'))"
	}

	var ordersAttempted int64
	_ = db.WithContext(ctx).
		Table("orders").
		Where("worker_id = ? AND "+attemptTimeExpr+" >= ? AND "+attemptTimeExpr+" <= ?"+attemptClause, source.WorkerID, windowStart, windowEnd).
		Count(&ordersAttempted).Error

	var ordersCompleted int64
	_ = db.WithContext(ctx).
		Table("orders").
		Where("worker_id = ? AND "+completeTimeExpr+" >= ? AND "+completeTimeExpr+" <= ?"+statusClause, source.WorkerID, windowStart, windowEnd).
		Count(&ordersCompleted).Error

	var beforeHours float64
	_ = db.WithContext(ctx).
		Table("earnings_records").
		Select("COALESCE(SUM(hours_worked), 0)").
		Where("worker_id = ? AND date >= ? AND date < ?", source.WorkerID, preStart.Format("2006-01-02"), windowStart.Format("2006-01-02")).
		Scan(&beforeHours).Error

	var duringHours float64
	_ = db.WithContext(ctx).
		Table("earnings_records").
		Select("COALESCE(SUM(hours_worked), 0)").
		Where("worker_id = ? AND date >= ? AND date <= ?", source.WorkerID, windowStart.Format("2006-01-02"), windowEnd.Format("2006-01-02")).
		Scan(&duringHours).Error

	var actualFromOrders float64
	if hasOrderColumn(db, "delivery_fee_inr") {
		_ = db.WithContext(ctx).
			Table("orders").
			Select("COALESCE(SUM(delivery_fee_inr), 0)").
			Where("worker_id = ? AND "+completeTimeExpr+" >= ? AND "+completeTimeExpr+" <= ?"+statusClause, source.WorkerID, windowStart, windowEnd).
			Scan(&actualFromOrders).Error
	}

	durationHours := math.Max(windowEnd.Sub(windowStart).Hours(), 1.0)
	baseline := row.BaselineAmount
	if baseline <= 0 {
		// Fallback to the loss already embedded in the auto-generated claim amount.
		baseline = maxFloat(source.ClaimAmount/0.85, source.ClaimAmount) * 4
	}
	expected := round2((baseline / 40.0) * durationHours)

	// Actual is the sum of granular order fees earned specifically during the window
	actual := round2(actualFromOrders)
	// NOTE: Do NOT fall back to source.ActualEarnings (weekly total) here.
	// If no orders were completed during the disruption window, actual = 0.
	// This correctly triggers the loss calculation for the income guarantee.
	if actual <= 0 && source.ActualEarnings != nil && ordersAttempted > 0 {
		// Only use the passed actual earnings if there IS window activity
		// (protects against stale weekly totals masking genuine disruption losses)
		actual = *source.ActualEarnings
	}

	zoneLabel := strings.TrimSpace(row.ZoneName)
	if strings.TrimSpace(row.ZoneCity) != "" && !strings.Contains(zoneLabel, row.ZoneCity) {
		zoneLabel = fmt.Sprintf("%s, %s", zoneLabel, row.ZoneCity)
	}

	loginDuration := duringHours
	if loginDuration <= 0 && row.IsOnline && !lastActiveAt.IsZero() && lastActiveAt.Before(windowStart) {
		// Use the recent live-session lead time as participation evidence for
		// workers who were online before the disruption hit.
		loginDuration = math.Min(durationHours, windowStart.Sub(lastActiveAt).Hours())
	}
	if loginDuration <= 0 && ordersAttempted > 0 {
		loginDuration = math.Min(durationHours, math.Max(minLoginEvidenceHours, float64(ordersAttempted)*0.5))
	}

	activeBefore := ordersBefore > 0 || beforeHours > 0

	return WorkerActivity{
		WorkerID:         source.WorkerID,
		Zone:             zoneLabel,
		SegmentID:        deriveSegmentID(zoneLabel, row.ZoneLevel, row.VehicleType),
		IsOnline:         row.IsOnline,
		ActiveBefore:     activeBefore,
		ActiveDuring:     loginDuration >= 0.02 || ordersAttempted >= 1 || ordersCompleted > 0,
		LoginDuration:    loginDuration,
		OrdersAttempted:  int(ordersAttempted),
		OrdersCompleted:  int(ordersCompleted),
		EarningsActual:   actual,
		EarningsExpected: expected,
	}, nil
}

func disruptionWindow(source ClaimSource) (time.Time, time.Time) {
	now := source.Now.UTC()
	if now.IsZero() {
		now = time.Now().UTC()
	}

	start := now.Add(-4 * time.Hour)
	if source.StartTime != nil {
		start = source.StartTime.UTC()
	} else if source.ConfirmedAt != nil {
		start = source.ConfirmedAt.Add(-4 * time.Hour).UTC()
	}

	end := start.Add(4 * time.Hour)
	if source.EndTime != nil && source.EndTime.After(start) {
		end = source.EndTime.UTC()
	}
	if !end.After(start) {
		end = start.Add(4 * time.Hour)
	}
	return start, end
}

func normalizeActivity(activity WorkerActivity) WorkerActivity {
	if activity.WorkerID == 0 {
		activity.WorkerID = 1
	}
	if strings.TrimSpace(activity.Zone) == "" {
		activity.Zone = "unknown-zone"
	}
	if strings.TrimSpace(activity.SegmentID) == "" {
		activity.SegmentID = deriveSegmentID(activity.Zone, "", "")
	}
	if activity.OrdersAttempted < activity.OrdersCompleted {
		activity.OrdersAttempted = activity.OrdersCompleted
	}
	if activity.LoginDuration < 0 {
		activity.LoginDuration = 0
	}
	if activity.LoginDuration == 0 && (activity.OrdersAttempted > 0 || activity.OrdersCompleted > 0) {
		activity.LoginDuration = math.Max(minLoginEvidenceHours, float64(activity.OrdersAttempted)*0.5)
	}
	if activity.EarningsExpected < 0 {
		activity.EarningsExpected = 0
	}
	if activity.EarningsActual < 0 {
		activity.EarningsActual = 0
	}
	if activity.EarningsExpected == 0 {
		activity.EarningsExpected = round2(float64(maxInt(activity.OrdersAttempted, 1)) * 60)
	}
	if activity.EarningsActual == 0 && activity.ActiveDuring && activity.OrdersCompleted > 0 {
		activity.EarningsActual = round2(float64(activity.OrdersCompleted) * 40)
	}
	if !activity.ActiveDuring {
		activity.ActiveDuring = activity.LoginDuration >= minLoginEvidenceHours || activity.OrdersAttempted >= 1 || activity.OrdersCompleted > 0
	}
	if activity.EarningsActual > activity.EarningsExpected && activity.EarningsExpected > 0 {
		activity.EarningsActual = activity.EarningsExpected
	}
	return activity
}

func deriveSegmentID(zone, zoneLevel, vehicleType string) string {
	cleanZone := slug(zone)
	level := slug(zoneLevel)
	vehicle := slug(vehicleType)
	parts := []string{cleanZone}
	if level != "" {
		parts = append(parts, level)
	}
	if vehicle != "" {
		parts = append(parts, vehicle)
	}
	return strings.Join(parts, ":")
}

func slug(value string) string {
	value = strings.TrimSpace(strings.ToLower(value))
	value = strings.ReplaceAll(value, ",", "-")
	value = strings.ReplaceAll(value, " ", "-")
	value = strings.ReplaceAll(value, "--", "-")
	return strings.Trim(value, "-")
}

func hasOrderColumn(db *gorm.DB, name string) bool {
	return db != nil && db.Migrator().HasColumn(&orderSchemaProbe{}, name)
}

func round2(value float64) float64 {
	return math.Round(value*100) / 100
}

func maxFloat(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// Ensure the probe type is anchored to the existing table name even when the
// runtime model omits later columns in SQLite test mode.
var _ interface{ TableName() string } = orderSchemaProbe{}
var _ = models.Order{}
