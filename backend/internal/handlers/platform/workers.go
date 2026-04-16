package platform

import (
	"time"

	"github.com/Shravanthi20/InDel/backend/internal/models"
	"github.com/gin-gonic/gin"
)

// GetWorkers returns worker list
func GetWorkers(c *gin.Context) {
	if hasDB() {
		type row struct {
			WorkerID     uint       `gorm:"column:worker_id"`
			Name         string     `gorm:"column:name"`
			Phone        string     `gorm:"column:phone"`
			Zone         string     `gorm:"column:zone"`
			IsOnline     bool       `gorm:"column:is_online"`
			LastActiveAt *time.Time `gorm:"column:last_active_at"`
		}

		rows := make([]row, 0)
		_ = platformDB.Raw(`
			SELECT u.id AS worker_id,
			       wp.name,
			       u.phone,
			       COALESCE(z.name || ', ' || z.city, 'Unknown Zone') AS zone,
			       COALESCE(wp.is_online, false) AS is_online,
			       wp.last_active_at AS last_active_at
			FROM users u
			JOIN worker_profiles wp ON wp.worker_id = u.id
			LEFT JOIN zones z ON z.id = wp.zone_id
			WHERE u.role = 'worker'
			ORDER BY u.id
		`).Scan(&rows).Error

		workers := make([]gin.H, 0, len(rows))
		now := time.Now()
		for _, r := range rows {
			lastActiveAt := time.Time{}
			if r.LastActiveAt != nil {
				lastActiveAt = *r.LastActiveAt
			}
			effectiveOnline := models.EffectiveWorkerOnlineStatus(r.IsOnline, lastActiveAt, now)

			workers = append(workers, gin.H{
				"worker_id":      r.WorkerID,
				"name":           r.Name,
				"phone":          r.Phone,
				"zone":           r.Zone,
				"is_online":      effectiveOnline,
				"last_active_at": lastActiveAt.Format(time.RFC3339),
				"status": func() string {
					if effectiveOnline {
						return "live"
					}
					return "offline"
				}(),
			})
		}

		c.JSON(200, gin.H{"workers": workers})
		return
	}

	c.JSON(200, gin.H{"workers": []gin.H{{
		"worker_id": 1,
		"name":      "Gayathri Worker",
		"phone":     "+919999999999",
		"zone":      "Tambaram, Chennai",
	}}})
}
