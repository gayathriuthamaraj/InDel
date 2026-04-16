package worker

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"math/big"
	"os"
	"strings"
	"time"

	"github.com/Shravanthi20/InDel/backend/internal/models"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// SendOTP sends OTP to worker phone
func SendOTP(c *gin.Context) {
	body := parseBody(c)
	phone := bodyString(body, "phone", "")
	if phone == "" {
		c.JSON(400, gin.H{"error": "phone_required"})
		return
	}

	store.mu.Lock()
	otp, err := generateOTPCode()
	if err != nil {
		store.mu.Unlock()
		c.JSON(500, gin.H{"error": "otp_generation_failed"})
		return
	}
	store.data.PhoneToOTP[phone] = otp
	store.mu.Unlock()

	resp := gin.H{
		"message":            "otp_sent",
		"phone":              phone,
		"expires_in_seconds": 300,
	}
	if shouldExposeTestingOTP() {
		resp["otp_for_testing"] = otp
	}
	c.JSON(200, resp)
}

// VerifyOTP verifies OTP and returns JWT
func VerifyOTP(c *gin.Context) {
	body := parseBody(c)
	phone := bodyString(body, "phone", "")
	otp := bodyString(body, "otp", "")
	if phone == "" || otp == "" {
		c.JSON(400, gin.H{"error": "phone_and_otp_required"})
		return
	}

	store.mu.Lock()
	defer store.mu.Unlock()

	expectedOTP, ok := store.data.PhoneToOTP[phone]
	if !ok || expectedOTP != otp {
		c.JSON(401, gin.H{"error": "invalid_otp"})
		return
	}

	token, err := generateSessionToken()
	if err != nil {
		c.JSON(500, gin.H{"error": "token_generation_failed"})
		return
	}
	workerID := "worker-001"

	if HasDB() {
		var user models.User
		err := workerDB.Where("phone = ?", phone).First(&user).Error
		if err == gorm.ErrRecordNotFound {
			newUser := models.User{Phone: phone, Role: "worker"}
			if createErr := workerDB.Create(&newUser).Error; createErr == nil {
				user = newUser
			}
		} // else if err == nil { /* user loaded, nothing to do */ }

		if user.ID != 0 {
			workerID = fmt.Sprintf("%d", user.ID)
		}
	}
	store.data.TokenToWorkerID[token] = workerID

	if HasDB() {
		if workerIDUint, parseErr := parseWorkerID(workerID); parseErr == nil {
			_ = workerDB.Exec(
				`INSERT INTO auth_tokens (user_id, token, expires_at)
				 VALUES (?, ?, CURRENT_TIMESTAMP + INTERVAL '24 hour')
				 ON CONFLICT (user_id)
				 DO UPDATE SET token = EXCLUDED.token, expires_at = EXCLUDED.expires_at`,
				workerIDUint, token,
			).Error

			zoneID := ensureZoneIDByLevelAndName("A", "Tambaram")
			if zoneID != 0 {
				_ = workerDB.Exec(
					`INSERT INTO worker_profiles (worker_id, name, zone_id, vehicle_type, upi_id, aqi_zone, total_earnings_lifetime, is_online, last_active_at)
					 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
					 ON CONFLICT (worker_id) DO UPDATE SET
					 name = COALESCE(NULLIF(worker_profiles.name, ''), EXCLUDED.name),
					 zone_id = COALESCE(NULLIF(worker_profiles.zone_id, 0), EXCLUDED.zone_id),
					 vehicle_type = COALESCE(NULLIF(worker_profiles.vehicle_type, ''), EXCLUDED.vehicle_type),
					 upi_id = COALESCE(NULLIF(worker_profiles.upi_id, ''), EXCLUDED.upi_id),
					 aqi_zone = COALESCE(NULLIF(worker_profiles.aqi_zone, ''), EXCLUDED.aqi_zone)`,
					workerIDUint, "New Worker", zoneID, "bike", "new@upi", "AQI-Medium", 0, false, time.Now(),
				).Error
			}

			_ = workerDB.Exec(
				`INSERT INTO earnings_baseline (worker_id, baseline_amount)
				 VALUES (?, 4080)
				 ON CONFLICT (worker_id) DO NOTHING`,
				workerIDUint,
			).Error

			_ = workerDB.Exec(
				`INSERT INTO policies (worker_id, status, premium_amount, created_at, updated_at)
				 VALUES (?, 'active', 35.00, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
				 ON CONFLICT (worker_id) WHERE status = 'active'
				 DO UPDATE SET premium_amount = EXCLUDED.premium_amount, updated_at = CURRENT_TIMESTAMP`,
				workerIDUint,
			).Error

			ensureMinimumOrdersForWorker(workerIDUint)
			log.Printf("VerifyOTP: provisioned worker profile and seeded queue worker_id=%d phone=%s", workerIDUint, phone)
		}
	}

	if _, exists := store.data.WorkerProfiles[workerID]; !exists {
		store.data.WorkerProfiles[workerID] = map[string]any{
			"worker_id":       workerID,
			"name":            "New Worker",
			"phone":           phone,
			"zone":            "Tambaram, Chennai",
			"vehicle_type":    "bike",
			"upi_id":          "new@upi",
			"coverage_status": "inactive",
			"enrolled":        false,
			"online":          false,
			"is_online":       false,
			"last_active_at":  time.Now(),
		}
	} else {
		store.data.WorkerProfiles[workerID]["last_active_at"] = time.Now()
	}

	if HasDB() {
		workerIDUint, _ := parseWorkerID(workerID)
		_ = workerDB.Model(&models.WorkerProfile{}).Where("worker_id = ?", workerIDUint).Update("last_active_at", time.Now()).Error
	}
	log.Printf("VerifyOTP: login success worker_id=%s phone=%s has_db=%t", workerID, phone, HasDB())

	c.JSON(200, gin.H{
		"message":    "otp_verified",
		"token":      token,
		"token_type": "Bearer",
		"worker_id":  workerID,
	})
}

// Register registers a new worker
func Register(c *gin.Context) {
	body := parseBody(c)
	username := bodyString(body, "username", "")
	phone := bodyString(body, "phone", "")
	email := bodyString(body, "email", "")
	password := bodyString(body, "password", "")
	normalizedZone := normalizeZoneInput(bodyString(body, "zone_level", ""), bodyString(body, "zone_name", ""))
	zoneLevel := normalizedZone.Level
	zoneName := normalizedZone.Name

	if username == "" || phone == "" || email == "" || password == "" {
		c.JSON(400, gin.H{"error": "username_phone_email_password_required"})
		return
	}

	workerID := fmt.Sprintf("worker-%s", phone)
	if HasDB() {
		var existing models.User
		err := workerDB.Where("phone = ? OR email = ?", phone, email).First(&existing).Error
		if err == gorm.ErrRecordNotFound {
			hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
			newUser := models.User{Phone: phone, Email: email, Role: "worker", PasswordHash: string(hashedPassword)}
			if createErr := workerDB.Create(&newUser).Error; createErr == nil {
				existing = newUser
			}
		} else {
			c.JSON(400, gin.H{"error": "user_already_exists"})
			return
		}
		if existing.ID != 0 {
			workerID = fmt.Sprintf("%d", existing.ID)
		}
	}

	token, err := generateSessionToken()
	if err != nil {
		c.JSON(500, gin.H{"error": "token_generation_failed"})
		return
	}
	if HasDB() {
		workerIDUint, _ := parseWorkerID(workerID)
		_ = workerDB.Exec(
			`INSERT INTO auth_tokens (user_id, token, expires_at)
			 VALUES (?, ?, CURRENT_TIMESTAMP + INTERVAL '24 hour')
			 ON CONFLICT (user_id) DO UPDATE SET token = EXCLUDED.token, expires_at = EXCLUDED.expires_at`,
			workerIDUint, token,
		)
	}

	store.mu.Lock()
	defer store.mu.Unlock()

	zoneDisplay := zoneName
	if zoneDisplay == "" {
		zoneDisplay = "Tambaram"
	}

	if _, exists := store.data.WorkerProfiles[workerID]; !exists {
		store.data.WorkerProfiles[workerID] = map[string]any{
			"worker_id":       workerID,
			"name":            username,
			"phone":           phone,
			"zone":            zoneDisplay,
			"zone_level":      zoneLevel,
			"zone_name":       zoneName,
			"vehicle_type":    "bike",
			"upi_id":          "new@upi",
			"coverage_status": "active",
			"enrolled":        true,
			"online":          false,
			"is_online":       false,
			"last_active_at":  time.Now(),
		}
	}
	store.data.WorkerProfiles[workerID]["name"] = username
	store.data.WorkerProfiles[workerID]["phone"] = phone
	store.data.WorkerProfiles[workerID]["zone"] = zoneDisplay
	store.data.WorkerProfiles[workerID]["zone_level"] = zoneLevel
	store.data.WorkerProfiles[workerID]["zone_name"] = zoneName
	store.data.WorkerProfiles[workerID]["coverage_status"] = "active"
	store.data.WorkerProfiles[workerID]["enrolled"] = true
	store.data.WorkerProfiles[workerID]["last_active_at"] = time.Now()

	// ─── DB Seeding for Demo "Real Data" ──────────────────────────────
	if HasDB() {
		workerIDUint, _ := parseWorkerID(workerID)

		// 1. Ensure the selected zone exists.
		resolvedZoneLevel := zoneLevel
		resolvedZoneName := zoneName
		if resolvedZoneLevel == "" {
			resolvedZoneLevel = "A"
		}
		if resolvedZoneName == "" {
			resolvedZoneName = "Tambaram"
		}
		zoneID := ensureZoneIDByLevelAndName(resolvedZoneLevel, resolvedZoneName)
		var zone models.Zone
		if zoneID != 0 {
			_ = workerDB.Where("id = ?", zoneID).First(&zone).Error
		}
		if zone.ID == 0 {
			normalizedFallback := normalizeZoneInput(resolvedZoneLevel, resolvedZoneName)
			zone = models.Zone{
				Name:       normalizedFallback.Name,
				Level:      normalizedFallback.Level,
				City:       normalizedFallback.City,
				State:      normalizedFallback.State,
				RiskRating: 0.62,
			}
			_ = workerDB.Create(&zone).Error
		}
		if zone.City != "" {
			store.data.WorkerProfiles[workerID]["zone"] = formatZoneDisplay(zone.Name, zone.City)
		}
		store.data.WorkerProfiles[workerID]["zone_level"] = resolvedZoneLevel
		store.data.WorkerProfiles[workerID]["zone_name"] = zone.Name

		// 2. Create Worker Profile in DB (Starting at 0 earnings).
		_ = workerDB.Exec(
			`INSERT INTO worker_profiles (worker_id, name, zone_id, vehicle_type, upi_id, aqi_zone, total_earnings_lifetime, is_online, last_active_at)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
			 ON CONFLICT (worker_id) DO UPDATE SET
			 name = EXCLUDED.name,
			 zone_id = EXCLUDED.zone_id,
			 vehicle_type = EXCLUDED.vehicle_type,
			 upi_id = EXCLUDED.upi_id,
			 aqi_zone = EXCLUDED.aqi_zone,
			 last_active_at = EXCLUDED.last_active_at`,
			workerIDUint, username, zone.ID, "bike", "demo@upi", "AQI-Medium", 0, false, time.Now(),
		)

		// 3. Set Earnings Baseline.
		_ = workerDB.Exec(
			`INSERT INTO earnings_baseline (worker_id, baseline_amount)
			 VALUES (?, 4080)
			 ON CONFLICT (worker_id) DO NOTHING`,
			workerIDUint,
		)

		// 3.5 Set Active Policy (Initial baseline seeded at 35 INR)
		_ = workerDB.Exec(
			`INSERT INTO policies (worker_id, status, premium_amount, created_at, updated_at)
			 VALUES (?, 'active', 35.00, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
			 ON CONFLICT (worker_id) WHERE status = 'active'
			 DO UPDATE SET premium_amount = EXCLUDED.premium_amount, updated_at = CURRENT_TIMESTAMP`,
			workerIDUint,
		)

		// 4. Keep the worker's live queue topped up instead of flooding the zone.
		ensureMinimumOrdersForWorker(workerIDUint)
		log.Printf("Register: provisioned worker_id=%d zone_id=%d", workerIDUint, zone.ID)

		// No fake earnings history seeded for new users
	}
	// ──────────────────────────────────────────────────────────────────

	c.JSON(201, gin.H{
		"token":      token,
		"token_type": "Bearer",
		"worker_id":  workerID,
	})
}

func Login(c *gin.Context) {
	body := parseBody(c)
	phone := bodyString(body, "phone", "")
	email := bodyString(body, "email", "")
	password := bodyString(body, "password", "")

	if password == "" || (phone == "" && email == "") {
		c.JSON(400, gin.H{"error": "identifier_and_password_required"})
		return
	}

	var workerID string

	if HasDB() {
		var user models.User
		query := workerDB.Where("phone = ?", phone)
		if phone == "" && email != "" {
			query = workerDB.Where("email = ?", email)
		}
		err := query.First(&user).Error
		if err != nil {
			c.JSON(401, gin.H{"error": "user_not_found"})
			return
		}

		if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
			c.JSON(401, gin.H{"error": "invalid_credentials"})
			return
		}
		workerID = fmt.Sprintf("%d", user.ID)
	} else {
		c.JSON(500, gin.H{"error": "db_connection_failed"})
		return
	}

	token, err := generateSessionToken()
	if err != nil {
		c.JSON(500, gin.H{"error": "token_generation_failed"})
		return
	}
	if HasDB() {
		workerIDUint, _ := parseWorkerID(workerID)
		_ = workerDB.Exec(
			`INSERT INTO auth_tokens (user_id, token, expires_at)
			 VALUES (?, ?, CURRENT_TIMESTAMP + INTERVAL '24 hour')
			 ON CONFLICT (user_id) DO UPDATE SET token = EXCLUDED.token, expires_at = EXCLUDED.expires_at`,
			workerIDUint, token,
		)
	}

	store.mu.Lock()
	store.data.TokenToWorkerID[token] = workerID
	store.mu.Unlock()

	// ─── DB Seeding for Order Refill ──────────────────────────────
	if HasDB() {
		workerIDUint, _ := parseWorkerID(workerID)
		zoneSummary := getWorkerZoneSummary(workerIDUint)
		if zoneSummary.ZoneName != "" {
			store.mu.Lock()
			profile := store.data.WorkerProfiles[workerID]
			if profile == nil {
				profile = map[string]any{"worker_id": workerID}
			}
			if zoneSummary.City != "" {
				profile["zone"] = formatZoneDisplay(zoneSummary.ZoneName, zoneSummary.City)
			} else {
				profile["zone"] = zoneSummary.ZoneName
			}
			profile["zone_level"] = normalizeZoneLevel(zoneSummary.ZoneLevel)
			profile["zone_type"] = orderRouteType(normalizeZoneLevel(zoneSummary.ZoneLevel))
			profile["worker_type"] = orderRouteType(normalizeZoneLevel(zoneSummary.ZoneLevel))
			profile["zone_name"] = zoneSummary.ZoneName
			profile["last_active_at"] = time.Now()
			if _, exists := profile["is_online"]; !exists {
				if online, ok := profile["online"].(bool); ok {
					profile["is_online"] = online
				}
			}
			store.data.WorkerProfiles[workerID] = profile
			store.mu.Unlock()
		}

		_ = workerDB.Model(&models.WorkerProfile{}).Where("worker_id = ?", workerIDUint).Update("last_active_at", time.Now()).Error

		// 1. Get user's zone.
		var zoneID uint
		_ = workerDB.Raw("SELECT zone_id FROM worker_profiles WHERE worker_id = ?", workerIDUint).Scan(&zoneID).Error
		if zoneID == 0 {
			zoneID = zoneSummary.ZoneID
		}
		if zoneID == 0 {
			zoneID = 1 // final fallback only when no saved worker zone exists
		}

		// 2. Keep the worker's queue at the configured minimum.
		ensureMinimumOrdersForWorker(workerIDUint)
		log.Printf("Login: seeded queue for worker_id=%d zone_id=%d", workerIDUint, zoneID)
	}
	// ─────────────────────────────────────────────────────────────

	c.JSON(200, gin.H{
		"token":      token,
		"token_type": "Bearer",
		"worker_id":  workerID,
	})
}

func generateOTPCode() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}

func generateSessionToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func shouldExposeTestingOTP() bool {
	if strings.EqualFold(strings.TrimSpace(os.Getenv("INDEL_ENV")), "production") {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(os.Getenv("INDEL_EXPOSE_TEST_OTP")), "true")
}

// Logout invalidates the session and marks worker offline
func Logout(c *gin.Context) {
	workerID, ok := requireAuth(c)
	if !ok {
		return
	}

	authHeader := c.GetHeader("Authorization")
	token := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))

	if HasDB() {
		workerIDUint, parseErr := parseWorkerID(workerID)
		if parseErr == nil {
			// 1. Mark Offline
			_ = workerDB.Model(&models.WorkerProfile{}).Where("worker_id = ?", workerIDUint).Updates(map[string]any{
				"is_online":      false,
				"last_active_at": time.Now(),
			}).Error

			// 2. Invalidate Token
			_ = workerDB.Exec("DELETE FROM auth_tokens WHERE token = ?", token).Error
		}
	}

	store.mu.Lock()
	// 1. Mark Offline in store
	if profile, exists := store.data.WorkerProfiles[workerID]; exists {
		profile["online"] = false
		profile["is_online"] = false
		profile["last_active_at"] = time.Now()
	}
	// 2. Clear from token map
	delete(store.data.TokenToWorkerID, token)
	store.mu.Unlock()

	c.JSON(200, gin.H{"message": "logged_out"})
}
