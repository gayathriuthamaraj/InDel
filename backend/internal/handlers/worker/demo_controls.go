package worker

import (
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/Shravanthi20/InDel/backend/internal/models"
	"github.com/gin-gonic/gin"
)

// ZonePair represents a from-to city pair for order generation
type ZonePair struct {
	ID         uint
	FromCity   string
	ToCity     string
	FromState  string
	ToState    string
	Distance   float64
	DistanceKm float64
	FromLat    float64
	FromLon    float64
	ToLat      float64
	ToLon      float64
}

type zoneBPair struct {
	From       string  `json:"from"`
	To         string  `json:"to"`
	State      string  `json:"state"`
	DistanceKm float64 `json:"distance_km"`
	FromLat    float64 `json:"from_lat"`
	FromLon    float64 `json:"from_lon"`
	ToLat      float64 `json:"to_lat"`
	ToLon      float64 `json:"to_lon"`
}

type zoneCPair struct {
	From       string  `json:"from"`
	To         string  `json:"to"`
	FromState  string  `json:"from_state"`
	ToState    string  `json:"to_state"`
	DistanceKm float64 `json:"distance_km"`
	FromLat    float64 `json:"from_lat"`
	FromLon    float64 `json:"from_lon"`
	ToLat      float64 `json:"to_lat"`
	ToLon      float64 `json:"to_lon"`
}

type zoneAEntry struct {
	City string `json:"city"`
}

type zoneIDRow struct {
	ID uint `gorm:"column:id"`
}

type zoneSeedScope struct {
	ZoneID    uint   `gorm:"column:zone_id"`
	ZoneCity  string `gorm:"column:zone_city"`
	ZoneLevel string `gorm:"column:zone_level"`
}

const minActiveOrdersPerWorker = 4

func readFirstExistingFile(paths []string) ([]byte, string, error) {
	for _, p := range paths {
		b, err := os.ReadFile(p)
		if err == nil {
			return b, p, nil
		}
	}
	return nil, "", fmt.Errorf("none of the candidate files exist: %v", paths)
}

// loadZonePairs loads zone A, B, and C pairs so the demo can generate all batch levels.
func loadZonePairs() ([]ZonePair, error) {
	var pairs []ZonePair

	zoneABytes, zoneAPath, err := readFirstExistingFile([]string{
		"/root/zone_a.json",
		"/app/zone_a.json",
		"../zone_a.json",
		"zone_a.json",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to load zone_a.json: %w", err)
	}

	zoneBBytes, zoneBPath, err := readFirstExistingFile([]string{
		"/root/zone_b.json",
		"/app/zone_b.json",
		"../zone_b.json",
		"zone_b.json",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to load zone_b.json: %w", err)
	}

	zoneCBytes, zoneCPath, err := readFirstExistingFile([]string{
		"/root/zone_c.json",
		"/app/zone_c.json",
		"../zone_c.json",
		"zone_c.json",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to load zone_c.json: %w", err)
	}

	var aEntries []string
	if err := json.Unmarshal(zoneABytes, &aEntries); err != nil {
		var fallbackEntries []zoneAEntry
		if fallbackErr := json.Unmarshal(zoneABytes, &fallbackEntries); fallbackErr == nil {
			for _, entry := range fallbackEntries {
				if strings.TrimSpace(entry.City) != "" {
					aEntries = append(aEntries, entry.City)
				}
			}
		} else {
			return nil, fmt.Errorf("failed to parse %s: %w", zoneAPath, err)
		}
	}

	for _, city := range aEntries {
		city = strings.TrimSpace(city)
		if city == "" {
			continue
		}
		pairs = append(pairs, ZonePair{
			ID:         uint(len(pairs) + 1),
			FromCity:   city,
			ToCity:     city,
			FromState:  "",
			ToState:    "",
			Distance:   1.0,
			DistanceKm: 1.0,
		})
	}

	var bPairs []zoneBPair
	if err := json.Unmarshal(zoneBBytes, &bPairs); err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", zoneBPath, err)
	}

	for _, p := range bPairs {
		if p.From == "" || p.To == "" {
			continue
		}
		pairs = append(pairs, ZonePair{
			ID:         uint(len(pairs) + 1),
			FromCity:   p.From,
			ToCity:     p.To,
			FromState:  p.State,
			ToState:    p.State,
			Distance:   p.DistanceKm,
			DistanceKm: p.DistanceKm,
			FromLat:    p.FromLat,
			FromLon:    p.FromLon,
			ToLat:      p.ToLat,
			ToLon:      p.ToLon,
		})
	}

	var cPairs []zoneCPair
	if err := json.Unmarshal(zoneCBytes, &cPairs); err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", zoneCPath, err)
	}

	for _, p := range cPairs {
		if p.From == "" || p.To == "" {
			continue
		}
		pairs = append(pairs, ZonePair{
			ID:         uint(len(pairs) + 1),
			FromCity:   p.From,
			ToCity:     p.To,
			FromState:  p.FromState,
			ToState:    p.ToState,
			Distance:   p.DistanceKm,
			DistanceKm: p.DistanceKm,
			FromLat:    p.FromLat,
			FromLon:    p.FromLon,
			ToLat:      p.ToLat,
			ToLon:      p.ToLon,
		})
	}

	if len(pairs) == 0 {
		return nil, fmt.Errorf("no usable zone pairs found in %s, %s, and %s", zoneAPath, zoneBPath, zoneCPath)
	}

	log.Printf("loadZonePairs: loaded %d pairs from %s, %s, and %s", len(pairs), zoneAPath, zoneBPath, zoneCPath)

	return pairs, nil
}

// calculateDeliveryFee calculates delivery fee based on distance and zone type
func calculateDeliveryFee(distanceKm float64, isInterState bool) float64 {
	if isInterState {
		return distanceKm * 2.0 // Inter-state: 2x multiplier
	}
	return distanceKm * 1.2 // Intra-state: 1.2x multiplier
}

type seededOrderSpec struct {
	ZoneLevel       string
	RouteType       string
	ZoneRoute       []string
	PickupArea      string
	DropArea        string
	PackageSize     string
	PackageWeightKg float64
	TipINR          float64
	OrderValue      float64
	DeliveryFee     float64
}

var realisticCustomerNames = []string{
	"Ananya Sharma",
	"Rohit Kumar",
	"Priya Nair",
	"Arjun Menon",
	"Sneha Reddy",
	"Karthik Iyer",
	"Meera Joshi",
	"Vikram Singh",
	"Divya Rao",
	"Nikhil Verma",
}

var cityAreaPools = map[string][]string{
	"hyderabad": {
		"Banjara Hills",
		"Gachibowli",
		"Madhapur",
		"Kondapur",
		"Jubilee Hills",
		"Hitech City",
		"Kukatpally",
		"Begumpet",
	},
	"chennai": {
		"Tambaram",
		"Chromepet",
		"Velachery",
		"Perungudi",
		"Guindy",
		"T Nagar",
		"Adyar",
		"Pallavaram",
	},
	"bengaluru": {
		"Indiranagar",
		"Koramangala",
		"Whitefield",
		"HSR Layout",
		"Jayanagar",
		"Marathahalli",
		"Electronic City",
		"BTM Layout",
	},
	"bangalore": {
		"Indiranagar",
		"Koramangala",
		"Whitefield",
		"HSR Layout",
		"Jayanagar",
		"Marathahalli",
		"Electronic City",
		"BTM Layout",
	},
	"mumbai": {
		"Andheri",
		"Bandra",
		"Powai",
		"Lower Parel",
		"Chembur",
		"Borivali",
		"Dadar",
		"Goregaon",
	},
	"delhi": {
		"Saket",
		"Dwarka",
		"Rohini",
		"Lajpat Nagar",
		"Karol Bagh",
		"Janakpuri",
		"Vasant Kunj",
		"Connaught Place",
	},
	"kolkata": {
		"Salt Lake",
		"Park Street",
		"New Town",
		"Garia",
		"Ballygunge",
		"Howrah",
		"Dum Dum",
		"Behala",
	},
	"pune": {
		"Hinjewadi",
		"Kothrud",
		"Viman Nagar",
		"Wakad",
		"Baner",
		"Hadapsar",
		"Aundh",
		"Kharadi",
	},
}

func realisticCustomerName(sequence int) string {
	if len(realisticCustomerNames) == 0 {
		return "Aarav Patel"
	}
	return realisticCustomerNames[sequence%len(realisticCustomerNames)]
}

func cityAreaPool(city string) []string {
	key := strings.ToLower(strings.TrimSpace(city))
	if areas, ok := cityAreaPools[key]; ok && len(areas) > 0 {
		return areas
	}
	return nil
}

func areaLabelForCity(city string, sequence int, suffix string) string {
	areas := cityAreaPool(city)
	if len(areas) > 0 {
		return areas[sequence%len(areas)]
	}
	base := firstNonEmpty(city, "Origin")
	if suffix == "" {
		return base
	}
	return fmt.Sprintf("%s %s", base, suffix)
}

func filterSeedPairsForZone(pairs []ZonePair, zoneCity, zoneLevel string) []ZonePair {
	normalizedCity := strings.ToLower(strings.TrimSpace(canonicalZoneCity(zoneCity, zoneCity)))
	normalizedLevel := strings.ToUpper(strings.TrimSpace(zoneLevel))
	if normalizedLevel == "" {
		normalizedLevel = "A"
	}

	filtered := make([]ZonePair, 0, len(pairs))
	for _, pair := range pairs {
		pairLevel := normalizeZoneLevelValue("", pair.FromCity, pair.ToCity, pair.FromState, pair.ToState)
		if pairLevel != normalizedLevel {
			continue
		}
		if normalizedLevel == "A" && normalizedCity != "" {
			fromLower := strings.ToLower(strings.TrimSpace(pair.FromCity))
			toLower := strings.ToLower(strings.TrimSpace(pair.ToCity))
			if fromLower != normalizedCity || toLower != normalizedCity {
				continue
			}
		}
		filtered = append(filtered, pair)
	}

	if len(filtered) > 0 {
		return filtered
	}
	return pairs
}

func hasGenericAreaLabel(area, city, fallbackSuffix string) bool {
	normalized := strings.ToLower(strings.TrimSpace(area))
	if normalized == "" {
		return true
	}

	genericLabels := []string{
		"pickup location",
		"drop location",
		"pickup",
		"drop",
	}
	for _, label := range genericLabels {
		if normalized == label {
			return true
		}
	}

	if strings.HasPrefix(normalized, "pickup hub ") || strings.HasPrefix(normalized, "drop point ") {
		return true
	}

	cityLower := strings.ToLower(strings.TrimSpace(city))
	if cityLower != "" {
		if normalized == cityLower {
			return true
		}
		if fallbackSuffix != "" && normalized == strings.ToLower(strings.TrimSpace(fmt.Sprintf("%s %s", city, fallbackSuffix))) {
			return true
		}
	}

	return false
}

func syncLocalOrderAreasForWorker(workerIDUint uint) {
	if !HasDB() || workerIDUint == 0 {
		return
	}

	workerScope := getWorkerOrderScope(workerIDUint)
	workerCity := strings.TrimSpace(canonicalZoneCity(workerScope.ZoneName, workerScope.ZoneCity))
	workerLevel := strings.ToUpper(strings.TrimSpace(workerScope.ZoneLevel))

	type orderAreaRow struct {
		ID         uint   `gorm:"column:id"`
		FromCity   string `gorm:"column:from_city"`
		ToCity     string `gorm:"column:to_city"`
		PickupArea string `gorm:"column:pickup_area"`
		DropArea   string `gorm:"column:drop_area"`
		Status     string `gorm:"column:status"`
	}

	rows := make([]orderAreaRow, 0)
	err := workerDB.Raw(`
		SELECT
			id,
			COALESCE(from_city, '') AS from_city,
			COALESCE(to_city, '') AS to_city,
			COALESCE(pickup_area, '') AS pickup_area,
			COALESCE(drop_area, '') AS drop_area,
			COALESCE(status, 'assigned') AS status
		FROM orders
		WHERE worker_id = ?
		  AND LOWER(TRIM(COALESCE(status, 'assigned'))) IN ('assigned', 'accepted', 'picked_up')
		ORDER BY id ASC
	`, workerIDUint).Scan(&rows).Error
	if err != nil {
		log.Printf("syncLocalOrderAreasForWorker: query failed worker_id=%d err=%v", workerIDUint, err)
		return
	}

	for _, row := range rows {
		if !strings.EqualFold(strings.TrimSpace(row.FromCity), strings.TrimSpace(row.ToCity)) {
			continue
		}
		targetCity := strings.TrimSpace(row.FromCity)
		if workerLevel == "A" && workerCity != "" {
			targetCity = workerCity
		}
		if len(cityAreaPool(targetCity)) == 0 {
			continue
		}

		cityNeedsUpdate := workerLevel == "A" &&
			workerCity != "" &&
			(!strings.EqualFold(strings.TrimSpace(row.FromCity), workerCity) || !strings.EqualFold(strings.TrimSpace(row.ToCity), workerCity))
		pickupNeedsUpdate := cityNeedsUpdate || hasGenericAreaLabel(row.PickupArea, targetCity, "Market Road")
		dropNeedsUpdate := cityNeedsUpdate || hasGenericAreaLabel(row.DropArea, targetCity, "Residency")
		if !pickupNeedsUpdate && !dropNeedsUpdate && !cityNeedsUpdate {
			continue
		}

		sequence := int(row.ID)
		newPickup := row.PickupArea
		if pickupNeedsUpdate {
			newPickup = areaLabelForCity(targetCity, sequence, "Market Road")
		}
		newDrop := row.DropArea
		if dropNeedsUpdate {
			newDrop = areaLabelForCity(targetCity, sequence+1, "Residency")
		}
		if strings.EqualFold(strings.TrimSpace(newPickup), strings.TrimSpace(newDrop)) {
			newDrop = areaLabelForCity(targetCity, sequence+2, "Residency")
		}

		if err := workerDB.Exec(
			"UPDATE orders SET from_city = ?, to_city = ?, pickup_area = ?, drop_area = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
			targetCity, targetCity, newPickup, newDrop, row.ID,
		).Error; err != nil {
			log.Printf("syncLocalOrderAreasForWorker: update failed worker_id=%d order_id=%d err=%v", workerIDUint, row.ID, err)
		}
	}
}

func buildSeededOrderSpec(pair ZonePair, sequence int) seededOrderSpec {
	zoneLevel := normalizeZoneLevelValue("", pair.FromCity, pair.ToCity, pair.FromState, pair.ToState)
	routeType := orderRouteType(zoneLevel)
	zoneRoute := deriveZoneRoutePath(zoneLevel)
	pickupArea := areaLabelForCity(pair.FromCity, sequence, "Market Road")
	dropArea := areaLabelForCity(pair.ToCity, sequence+1, "Residency")
	if strings.EqualFold(strings.TrimSpace(pickupArea), strings.TrimSpace(dropArea)) {
		dropArea = areaLabelForCity(pair.ToCity, sequence+2, "Residency")
	}

	spec := seededOrderSpec{
		ZoneLevel:       zoneLevel,
		RouteType:       routeType,
		ZoneRoute:       zoneRoute,
		PickupArea:      pickupArea,
		DropArea:        dropArea,
		PackageSize:     "small",
		PackageWeightKg: 0.8,
		TipINR:          6,
	}

	switch routeType {
	case "local":
		spec.PackageSize = "small"
		spec.PackageWeightKg = 0.8 + float64(sequence%3)*0.2
		spec.TipINR = 8 + float64(sequence%2)*2
		spec.OrderValue = 95 + float64(sequence%4)*20
	case "interstate":
		spec.PackageSize = "large"
		spec.PackageWeightKg = 2.5 + float64(sequence%3)*0.5
		spec.TipINR = 18 + float64(sequence%2)*4
		spec.OrderValue = 320 + pair.DistanceKm*1.6 + float64(sequence%5)*25
	default:
		spec.PackageSize = "medium"
		spec.PackageWeightKg = 1.4 + float64(sequence%3)*0.35
		spec.TipINR = 12 + float64(sequence%2)*3
		spec.OrderValue = 180 + pair.DistanceKm*0.9 + float64(sequence%4)*18
	}

	spec.DeliveryFee = float64(computeZoneRouteDeliveryFee(zoneRoute)) + calculateDeliveryFee(pair.DistanceKm, routeType == "interstate")
	return spec
}

func loadZoneIDs() ([]uint, error) {
	if !HasDB() {
		return nil, fmt.Errorf("no database connection available")
	}

	rows := make([]zoneIDRow, 0)
	if err := workerDB.Raw("SELECT id FROM zones ORDER BY id ASC").Scan(&rows).Error; err != nil {
		return nil, err
	}

	zoneIDs := make([]uint, 0, len(rows))
	for _, row := range rows {
		if row.ID != 0 {
			zoneIDs = append(zoneIDs, row.ID)
		}
	}
	if len(zoneIDs) == 0 {
		return nil, fmt.Errorf("no zones available in database")
	}
	return zoneIDs, nil
}

func countActiveOrdersForWorker(workerIDUint uint) int {
	if workerIDUint == 0 {
		return 0
	}

	if HasDB() {
		type countRow struct {
			Count int `gorm:"column:count"`
		}
		var row countRow
		_ = workerDB.Raw(`
			SELECT COUNT(*) AS count
			FROM orders
			WHERE worker_id = ?
			  AND LOWER(TRIM(COALESCE(status, 'assigned'))) IN ('assigned', 'accepted', 'picked_up')
		`, workerIDUint).Scan(&row).Error
		return row.Count
	}

	workerID := fmt.Sprintf("%d", workerIDUint)
	count := 0
	store.mu.RLock()
	for _, order := range store.data.Orders {
		if fmt.Sprintf("%v", order["worker_id"]) != workerID {
			continue
		}
		status := strings.ToLower(strings.TrimSpace(fmt.Sprintf("%v", order["status"])))
		if status == "assigned" || status == "accepted" || status == "picked_up" {
			count++
		}
	}
	store.mu.RUnlock()
	return count
}

func ensureMinimumOrdersForWorker(workerIDUint uint) {
	if workerIDUint == 0 {
		log.Printf("ensureMinimumOrdersForWorker: skipped because worker_id=0")
		return
	}

	activeCount := countActiveOrdersForWorker(workerIDUint)
	missing := minActiveOrdersPerWorker - activeCount
	log.Printf("ensureMinimumOrdersForWorker: worker_id=%d active_count=%d missing=%d has_db=%t", workerIDUint, activeCount, missing, HasDB())
	if missing <= 0 {
		return
	}

	if HasDB() {
		log.Printf("ensureMinimumOrdersForWorker: seeding db orders worker_id=%d count=%d", workerIDUint, missing)
		seedDemoOrdersForWorker(workerIDUint, missing)
		return
	}

	now := nowISO()
	workerID := fmt.Sprintf("%d", workerIDUint)
	store.mu.Lock()
	base := len(store.data.Orders)
	for i := 0; i < missing; i++ {
		order, _ := parseDemoOrderPayload(map[string]any{
			"order_id":                nextID("ord", base+i),
			"customer_name":           fmt.Sprintf("Demo Customer %02d", base+i+1),
			"customer_id":             fmt.Sprintf("cust-demo-%d-%d", workerIDUint, base+i+1),
			"customer_contact_number": fmt.Sprintf("+91%010d", 9300000000+base+i),
			"address":                 fmt.Sprintf("Drop Point %d", base+i+1),
			"payment_method":          "cod",
			"order_value":             110 + float64((base+i)%4)*20,
			"payment_amount":          110 + float64((base+i)%4)*20,
			"package_size":            "small",
			"package_weight_kg":       1.0,
			"zone_id":                 1,
			"from_city":               "Tambaram",
			"to_city":                 "Camp Road",
			"pickup_area":             fmt.Sprintf("Pickup Hub %d", base+i+1),
			"drop_area":               fmt.Sprintf("Drop Point %d", base+i+1),
			"distance_km":             2.5 + float64(i)*0.3,
			"tip_inr":                 8.0,
			"zone_level":              "A",
			"route_type":              "local",
			"zone_route_path":         []string{"A"},
			"delivery_fee_inr":        25.0,
			"status":                  "assigned",
			"assigned_at":             now,
			"source":                  "queue-refill",
		})
		order["worker_id"] = workerID
		order["created_at"] = now
		order["updated_at"] = now
		store.data.Orders = append(store.data.Orders, order)
	}
	store.mu.Unlock()
	log.Printf("ensureMinimumOrdersForWorker: seeded in-memory orders worker_id=%d count=%d", workerIDUint, missing)
}

func seedDemoOrdersForZones(workerIDUint uint, zoneIDs []uint, count int) {
	if len(zoneIDs) == 0 {
		log.Println("seedDemoOrdersForZones: no zone ids available, falling back to worker zone")
		seedDemoOrdersForWorker(workerIDUint, count)
		return
	}

	if count <= 0 {
		count = len(zoneIDs) * 2
	}

	pairs, err := loadZonePairs()
	if err != nil || len(pairs) == 0 {
		if err != nil {
			log.Printf("seedDemoOrdersForZones: failed to load zone pairs: %v\n", err)
		}
		seedDemoOrdersWithFallback(workerIDUint, zoneIDs[0], count)
		return
	}

	now := time.Now()
	successCount := 0
	for i := 0; i < count; i++ {
		zoneID := zoneIDs[i%len(zoneIDs)]
		scope := zoneSeedScope{}
		_ = workerDB.Raw("SELECT id AS zone_id, COALESCE(city, '') AS zone_city, COALESCE(level, '') AS zone_level FROM zones WHERE id = ? LIMIT 1", zoneID).Scan(&scope).Error
		zonePairs := filterSeedPairsForZone(pairs, scope.ZoneCity, scope.ZoneLevel)
		pair := zonePairs[i%len(zonePairs)]
		spec := buildSeededOrderSpec(pair, i)

		err := workerDB.Exec(`
			INSERT INTO orders (
				worker_id, zone_id, order_value, status,
				pickup_area, drop_area, distance_km, from_city, to_city,
				from_state, to_state, from_lat, from_lon, to_lat, to_lon,
				tip_inr, delivery_fee_inr, zone_route_path,
				package_size, package_weight_kg, customer_name, customer_id, customer_contact_number, address, payment_method,
				created_at, updated_at
			) VALUES (?, ?, ?, 'assigned', ?, ?, ?, ?, ?,
				  ?, ?, ?, ?, ?, ?,
				  ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			workerIDUint, zoneID, spec.OrderValue,
			spec.PickupArea, spec.DropArea, pair.DistanceKm,
			pair.FromCity, pair.ToCity,
			pair.FromState, pair.ToState,
			pair.FromLat, pair.FromLon, pair.ToLat, pair.ToLon,
			spec.TipINR, spec.DeliveryFee, encodeZonePath(spec.ZoneRoute),
			spec.PackageSize, spec.PackageWeightKg,
			fmt.Sprintf("Demo Customer %02d", i+1),
			fmt.Sprintf("cust-seed-%d-%d", workerIDUint, i+1),
			fmt.Sprintf("+91%010d", 9000000000+int(workerIDUint)+i),
			spec.DropArea,
			"cod",
			now, now,
		).Error

		if err != nil {
			log.Printf("seedDemoOrdersForZones: failed to insert order %d for zone %d: %v\n", i+1, zoneID, err)
			continue
		}
		successCount++
		log.Printf("seedDemoOrdersForZones: order %d -> zone %d [%s] %s -> %s | %.1f km | Rs %.2f\n", i+1, zoneID, spec.RouteType, spec.PickupArea, spec.DropArea, pair.DistanceKm, spec.DeliveryFee)
		seedOrder, _ := parseDemoOrderPayload(map[string]any{
			"order_id":                fmt.Sprintf("ord-seed-%d-%d", workerIDUint, i+1),
			"customer_name":           fmt.Sprintf("Demo Customer %02d", i+1),
			"customer_id":             fmt.Sprintf("cust-seed-%d-%d", workerIDUint, i+1),
			"customer_contact_number": fmt.Sprintf("+91%010d", 9000000000+int(workerIDUint)+i),
			"address":                 spec.DropArea,
			"payment_method":          "cod",
			"order_value":             spec.OrderValue,
			"payment_amount":          spec.OrderValue,
			"package_size":            spec.PackageSize,
			"package_weight_kg":       spec.PackageWeightKg,
			"zone_id":                 zoneID,
			"from_city":               pair.FromCity,
			"to_city":                 pair.ToCity,
			"from_state":              pair.FromState,
			"to_state":                pair.ToState,
			"pickup_area":             spec.PickupArea,
			"drop_area":               spec.DropArea,
			"distance_km":             pair.DistanceKm,
			"tip_inr":                 spec.TipINR,
			"zone_level":              spec.ZoneLevel,
			"route_type":              spec.RouteType,
			"zone_route_path":         spec.ZoneRoute,
			"delivery_fee_inr":        spec.DeliveryFee,
			"status":                  "assigned",
			"assigned_at":             nowISO(),
			"source":                  "demo-controls",
		})
		seedOrder["worker_id"] = fmt.Sprintf("%d", workerIDUint)
		seedOrder["created_at"] = nowISO()
		seedOrder["updated_at"] = nowISO()
		refreshBatchSnapshotsForOrder(seedOrder)
		scheduleBatchMaterialization(availableBatchCacheScope, seedOrder)
		scheduleBatchMaterialization(fmt.Sprintf("%d", workerIDUint), seedOrder)
	}
	log.Printf("seedDemoOrdersForZones: successfully seeded %d/%d orders across %d zones for worker %d\n", successCount, count, len(zoneIDs), workerIDUint)
}

// seedDemoOrdersForWorker creates realistic demo orders using zone pairs
func seedDemoOrdersForWorker(workerIDUint uint, count int) {
	if count <= 0 {
		count = 3 // Default to 3 orders
	}
	log.Printf("seedDemoOrdersForWorker: worker_id=%d requested_count=%d", workerIDUint, count)

	if !HasDB() {
		log.Println("seedDemoOrdersForWorker: No database connection available")
		return
	}

	// Get worker's zone_id
	var zoneID uint
	err := workerDB.Raw("SELECT zone_id FROM worker_profiles WHERE worker_id = ? LIMIT 1", workerIDUint).Scan(&zoneID).Error
	if err != nil {
		log.Printf("seedDemoOrdersForWorker: Failed to get worker zone_id: %v\n", err)
	}

	if zoneID == 0 {
		if zoneIDs, zoneErr := loadZoneIDs(); zoneErr == nil && len(zoneIDs) > 0 {
			zoneID = zoneIDs[0]
			_ = workerDB.Exec(
				`UPDATE worker_profiles
				 SET zone_id = ?, updated_at = CURRENT_TIMESTAMP
				 WHERE worker_id = ?`,
				zoneID, workerIDUint,
			).Error
			log.Printf("seedDemoOrdersForWorker: Worker %d had no zone_id, defaulted to zone %d\n", workerIDUint, zoneID)
		} else {
			log.Printf("seedDemoOrdersForWorker: Worker %d has no zone_id assigned and no zones available\n", workerIDUint)
			return
		}
	}

	seedDemoOrdersForZones(workerIDUint, []uint{zoneID}, count)
}

// seedDemoOrdersWithFallback creates demo orders using hardcoded areas (fallback)
func seedDemoOrdersWithFallback(workerIDUint, zoneID uint, count int) {
	now := time.Now()
	pickupAreas := []string{"Tambaram", "Camp Road", "Perungudi", "T Nagar"}
	dropAreas := []string{"Camp Road", "Perungudi", "T Nagar", "Nungambakkam"}

	for i := 0; i < count; i++ {
		pickupIdx := i % len(pickupAreas)
		dropIdx := (i + 1) % len(dropAreas)
		spec := seededOrderSpec{
			ZoneLevel:       "A",
			RouteType:       "local",
			ZoneRoute:       []string{"A"},
			PickupArea:      pickupAreas[pickupIdx] + " Main Road",
			DropArea:        dropAreas[dropIdx] + " Apartments",
			PackageSize:     "small",
			PackageWeightKg: 0.9,
			TipINR:          8,
			OrderValue:      95 + float64(i*14),
			DeliveryFee:     25.0 + float64(i*4),
		}
		customerName := realisticCustomerName(i)

		err := workerDB.Exec(`
			INSERT INTO orders (
				worker_id, zone_id, order_value, status, from_city, to_city,
				pickup_area, drop_area, distance_km, 
				tip_inr, delivery_fee_inr, zone_route_path,
				package_size, package_weight_kg, customer_name, customer_id, customer_contact_number, address, payment_method,
				created_at, updated_at
			) VALUES (?, ?, ?, 'assigned', ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			workerIDUint, zoneID, spec.OrderValue, "Chennai", "Chennai",
			spec.PickupArea, spec.DropArea, 2.5+float64(i)*0.4,
			spec.TipINR, spec.DeliveryFee, `["A"]`,
			spec.PackageSize, spec.PackageWeightKg,
			customerName,
			fmt.Sprintf("cust-fallback-%d-%d", workerIDUint, i+1),
			fmt.Sprintf("+91%010d", 9100000000+int(workerIDUint)+i),
			spec.DropArea,
			"cod",
			now, now,
		).Error

		if err != nil {
			log.Printf("seedDemoOrdersWithFallback: Failed to insert order %d: %v\n", i+1, err)
			continue
		}
		seedOrder, _ := parseDemoOrderPayload(map[string]any{
			"order_id":                fmt.Sprintf("ord-seed-%d-%d", workerIDUint, i+1),
			"customer_name":           customerName,
			"customer_id":             fmt.Sprintf("cust-fallback-%d-%d", workerIDUint, i+1),
			"customer_contact_number": fmt.Sprintf("+91%010d", 9100000000+int(workerIDUint)+i),
			"address":                 spec.DropArea,
			"payment_method":          "cod",
			"order_value":             spec.OrderValue,
			"payment_amount":          spec.OrderValue,
			"package_size":            spec.PackageSize,
			"package_weight_kg":       spec.PackageWeightKg,
			"zone_id":                 zoneID,
			"from_city":               pickupAreas[pickupIdx],
			"to_city":                 dropAreas[dropIdx],
			"pickup_area":             spec.PickupArea,
			"drop_area":               spec.DropArea,
			"distance_km":             2.5 + float64(i)*0.4,
			"tip_inr":                 spec.TipINR,
			"zone_level":              spec.ZoneLevel,
			"route_type":              spec.RouteType,
			"zone_route_path":         []string{"A"},
			"delivery_fee_inr":        spec.DeliveryFee,
			"status":                  "assigned",
			"assigned_at":             nowISO(),
			"source":                  "demo-controls",
		})
		seedOrder["worker_id"] = fmt.Sprintf("%d", workerIDUint)
		seedOrder["created_at"] = nowISO()
		seedOrder["updated_at"] = nowISO()
		refreshBatchSnapshotsForOrder(seedOrder)
		scheduleBatchMaterialization(fmt.Sprintf("%d", workerIDUint), seedOrder)
		scheduleBatchMaterialization(availableBatchCacheScope, seedOrder)
	}
}

// DemoReset resets all in-memory demo state and clears orders/batches (no auth required for demo).
func DemoReset(c *gin.Context) {
	if !requireDemoControlAuth(c) {
		return
	}

	body := parseBody(c)
	deleteDB := bodyBool(body, "delete_db", false)
	reason := strings.TrimSpace(bodyString(body, "reason", ""))
	confirm := strings.TrimSpace(bodyString(body, "confirm", ""))

	var resetLog []string
	resetLog = append(resetLog, "DemoReset initiated")

	clearBatchMaterializationTimers()
	store.batchMu.Lock()
	store.batchCache = map[string]map[string]gin.H{}
	store.batchMu.Unlock()

	store.mu.Lock()
	store.data.Orders = []map[string]any{}
	store.mu.Unlock()
	resetLog = append(resetLog, "Cleared in-memory orders and batches")

	if HasDB() && deleteDB {
		if !allowDestructiveDemoDelete() {
			c.JSON(403, gin.H{"error": "destructive_demo_delete_blocked", "message": "destructive demo delete is disabled in production unless INDEL_ALLOW_DESTRUCTIVE_OPS=true"})
			return
		}
		if len(reason) < 8 {
			c.JSON(400, gin.H{"error": "reason_required", "message": "reason must be provided and at least 8 characters", "field": "reason"})
			return
		}
		if !strings.EqualFold(confirm, "RESET_DEMO_DB") {
			c.JSON(400, gin.H{"error": "confirmation_required", "message": "set confirm to RESET_DEMO_DB to allow database deletion", "field": "confirm"})
			return
		}

		result1 := workerDB.Exec("DELETE FROM notifications")
		if result1.Error == nil {
			resetLog = append(resetLog, fmt.Sprintf("Deleted %d notifications", result1.RowsAffected))
		}

		result2 := workerDB.Exec("DELETE FROM auth_tokens")
		if result2.Error == nil {
			resetLog = append(resetLog, fmt.Sprintf("Deleted %d auth_tokens", result2.RowsAffected))
		}

		result3 := workerDB.Exec("DELETE FROM orders")
		if result3.Error == nil {
			resetLog = append(resetLog, fmt.Sprintf("Deleted %d orders", result3.RowsAffected))
		}
	} else if HasDB() {
		resetLog = append(resetLog, "Skipped database deletion (set delete_db=true with confirm and reason to enable)")
	}

	log.Println("DemoReset: " + fmt.Sprint(resetLog))
	c.JSON(200, gin.H{
		"message": "demo_reset",
		"time":    nowISO(),
		"details": resetLog,
	})
}

// DemoTriggerDisruption creates a disruption notification.
func DemoTriggerDisruption(c *gin.Context) {
	workerID, ok := requireDemoOperationRole(c)
	if !ok {
		return
	}
	body := parseBody(c)
	disruptionType := bodyString(body, "disruption_type", "heavy_rain")
	zoneLevel := normalizeZoneLevel(bodyString(body, "zone_level", ""))
	zoneName := bodyString(body, "zone_name", "")
	zone := bodyString(body, "zone", "")
	if zoneLevel == "" || zoneName == "" {
		if workerIDUint, parseErr := parseWorkerID(workerID); parseErr == nil {
			summary := getWorkerZoneSummary(workerIDUint)
			if zoneLevel == "" {
				zoneLevel = normalizeZoneLevel(summary.ZoneLevel)
			}
			if zoneName == "" {
				zoneName = summary.ZoneName
				if zoneName == "" {
					zoneName = summary.City
				}
			}
			if zone == "" {
				zone = formatZoneDisplay(summary.ZoneName, summary.City)
			}
		}
	}
	if zoneLevel == "" {
		zoneLevel = "A"
	}
	if zoneName == "" {
		zoneName = "Tambaram"
	}
	if zone == "" {
		zone = formatZoneDisplay(zoneName, "")
	}
	msg := disruptionType + " detected in " + zone + ". You are protected."

	if HasDB() {
		zoneID := ensureZoneIDByLevelAndName(zoneLevel, zoneName)
		if zoneID == 0 {
			c.JSON(400, gin.H{"error": "zone_resolution_failed"})
			return
		}

		now := time.Now().UTC()
		endTime := now.Add(4 * time.Hour)
		disruption := models.Disruption{
			ZoneID:          zoneID,
			Type:            disruptionType,
			Severity:        "high",
			Confidence:      0.92,
			Status:          "confirmed",
			SignalTimestamp: &now,
			ConfirmedAt:     &now,
			StartTime:       &now,
			EndTime:         &endTime,
		}
		if err := workerDB.Create(&disruption).Error; err != nil {
			c.JSON(500, gin.H{"error": "disruption_create_failed"})
			return
		}

		if workerCoreOps != nil {
			result, err := workerCoreOps.AutoProcessDisruption(disruption.ID, now)
			if err != nil {
				c.JSON(500, gin.H{"error": "disruption_auto_process_failed", "message": err.Error()})
				return
			}

			c.JSON(200, gin.H{
				"message":         "disruption_triggered",
				"disruption_id":   fmt.Sprintf("dis-%d", disruption.ID),
				"disruption_type": disruptionType,
				"zone":            zone,
				"zone_level":      zoneLevel,
				"zone_name":       zoneName,
				"time":            nowISO(),
				"result":          result,
			})
			return
		}
	}

	store.mu.Lock()
	store.data.Notifications = append([]map[string]any{{
		"id":         nextID("ntf", len(store.data.Notifications)),
		"type":       "disruption_alert",
		"title":      "Disruption detected",
		"body":       msg,
		"created_at": nowISO(),
		"read":       false,
	}}, store.data.Notifications...)
	store.mu.Unlock()

	c.JSON(200, gin.H{
		"message":         "disruption_triggered",
		"disruption_type": disruptionType,
		"zone":            zone,
		"time":            nowISO(),
	})
}

// DemoSimulateOrders appends assigned orders for demo.
func DemoSimulateOrders(c *gin.Context) {
	workerID, ok := requireDemoOperationRole(c)
	if !ok {
		return
	}
	body := parseBody(c)
	count := bodyInt(body, "count", 3)
	if count <= 0 {
		count = 1
	}

	if HasDB() {
		if workerIDUint, parseErr := parseWorkerID(workerID); parseErr == nil {
			seedDemoOrdersForWorker(workerIDUint, count)
		}
	}

	store.mu.Lock()
	base := len(store.data.Orders)
	for i := 0; i < count; i++ {
		order, _ := parseDemoOrderPayload(map[string]any{
			"order_id":                nextID("ord", base+i),
			"customer_name":           "Demo Customer",
			"customer_id":             fmt.Sprintf("cust-demo-%d", base+i),
			"customer_contact_number": fmt.Sprintf("+91%010d", 9200000000+base+i),
			"address":                 "Tambaram",
			"payment_method":          "cod",
			"order_value":             55 + float64(i*8),
			"payment_amount":          55 + float64(i*8),
			"package_size":            "small",
			"package_weight_kg":       1.0,
			"zone_id":                 1,
			"from_city":               "Tambaram",
			"to_city":                 "Camp Road",
			"pickup_area":             "Tambaram",
			"drop_area":               "Camp Road",
			"distance_km":             2.5 + float64(i)*0.4,
			"tip_inr":                 0.0,
			"zone_level":              "A",
			"zone_route_path":         []string{"A"},
			"delivery_fee_inr":        25.0,
			"status":                  "assigned",
			"assigned_at":             nowISO(),
			"source":                  "demo-controls",
		})
		order["created_at"] = nowISO()
		order["updated_at"] = nowISO()
		store.data.Orders = append(store.data.Orders, order)
	}
	store.mu.Unlock()

	for _, order := range store.data.Orders[base:] {
		refreshBatchSnapshotsForOrder(order)
		scheduleBatchMaterialization(availableBatchCacheScope, order)
		if workerID, ok := order["worker_id"].(string); ok && workerID != "" {
			scheduleBatchMaterialization(workerID, order)
		}
	}

	c.JSON(200, gin.H{"message": "orders_simulated", "count": count})
}

// DemoSettleEarnings settles demo earnings and triggers premium reminder.
func DemoSettleEarnings(c *gin.Context) {
	workerID, ok := requireDemoOperationRole(c)
	if !ok {
		return
	}

	if HasDB() {
		if workerIDUint, parseErr := parseWorkerID(workerID); parseErr == nil {
			_ = workerDB.Exec(
				`UPDATE weekly_earnings_summary
				 SET claim_eligible = TRUE
				 WHERE worker_id = ?
				   AND week_start = date_trunc('week', CURRENT_DATE)::date
				   AND week_end = (date_trunc('week', CURRENT_DATE)::date + INTERVAL '6 day')::date`,
				workerIDUint,
			).Error
			_ = workerDB.Exec(
				"INSERT INTO notifications (worker_id, type, message) VALUES (?, 'premium_due', 'Weekly earnings settled. Pay premium to keep coverage active.')",
				workerIDUint,
			).Error
		}
	}

	store.mu.Lock()
	store.data.Notifications = append([]map[string]any{{
		"id":         nextID("ntf", len(store.data.Notifications)),
		"type":       "premium_due",
		"title":      "Weekly settlement complete",
		"body":       "Weekly earnings settled. Pay premium to keep coverage active.",
		"created_at": nowISO(),
		"read":       false,
	}}, store.data.Notifications...)
	store.mu.Unlock()

	c.JSON(200, gin.H{"message": "earnings_settled", "time": nowISO()})
}

// DemoResetZone resets disruption and claim state for demo replay.
func DemoResetZone(c *gin.Context) {
	workerID, ok := requireDemoDestructiveRole(c)
	if !ok {
		return
	}
	body := parseBody(c)
	reason := strings.TrimSpace(bodyString(body, "reason", ""))

	if HasDB() {
		if !allowDestructiveDemoDelete() {
			c.JSON(403, gin.H{"error": "destructive_demo_delete_blocked", "message": "destructive demo delete is disabled in production unless INDEL_ALLOW_DESTRUCTIVE_OPS=true"})
			return
		}
		if len(reason) < 8 {
			c.JSON(400, gin.H{"error": "reason_required", "message": "reason must be provided and at least 8 characters", "field": "reason"})
			return
		}

		if workerIDUint, parseErr := parseWorkerID(workerID); parseErr == nil {
			_ = workerDB.Exec("DELETE FROM payouts WHERE worker_id = ?", workerIDUint).Error
			_ = workerDB.Exec("DELETE FROM claims WHERE worker_id = ?", workerIDUint).Error
			_ = workerDB.Exec("DELETE FROM notifications WHERE worker_id = ? AND type IN ('disruption_alert', 'payout_credited')", workerIDUint).Error
		}
	}

	store.mu.Lock()
	store.data.Claims = []map[string]any{}
	store.data.Payouts = []map[string]any{}
	store.data.Notifications = []map[string]any{}
	store.mu.Unlock()

	c.JSON(200, gin.H{"message": "zone_reset", "time": nowISO()})
}

func allowDestructiveDemoDelete() bool {
	if strings.EqualFold(strings.TrimSpace(os.Getenv("INDEL_ALLOW_DESTRUCTIVE_OPS")), "true") {
		return true
	}
	return !strings.EqualFold(strings.TrimSpace(os.Getenv("INDEL_ENV")), "production")
}

func requireDemoControlAuth(c *gin.Context) bool {
	if _, ok := requireDemoDestructiveRole(c); ok {
		return true
	}

	expected := strings.TrimSpace(os.Getenv("INDEL_DEMO_RESET_KEY"))
	provided := strings.TrimSpace(c.GetHeader("X-Demo-Reset-Key"))
	if expected != "" && provided != "" && subtle.ConstantTimeCompare([]byte(provided), []byte(expected)) == 1 {
		return true
	}

	c.JSON(401, gin.H{"error": "unauthorized_demo_control"})
	return false
}

func requireDemoOperationRole(c *gin.Context) (string, bool) {
	allowed := allowedRolesFromEnv(
		"INDEL_DEMO_ALLOWED_ROLES",
		[]string{"admin", "platform_admin", "ops_manager"},
		[]string{"worker", "admin", "platform_admin", "ops_manager"},
	)
	return requireRole(c, allowed, "forbidden_demo_operation")
}

func requireDemoDestructiveRole(c *gin.Context) (string, bool) {
	allowed := allowedRolesFromEnv(
		"INDEL_DEMO_DESTRUCTIVE_ROLES",
		[]string{"admin", "platform_admin"},
		[]string{"admin", "platform_admin"},
	)
	return requireRole(c, allowed, "forbidden_demo_destructive_operation")
}

func requireRole(c *gin.Context, allowedRoles []string, errorCode string) (string, bool) {
	workerID, ok := optionalAuthWorkerID(c)
	if !ok {
		c.JSON(401, gin.H{"error": "missing_or_invalid_bearer_token"})
		return "", false
	}

	role := getWorkerRole(workerID)
	if role == "" {
		c.JSON(403, gin.H{"error": errorCode, "message": "unable to determine role"})
		return "", false
	}

	if !containsRole(allowedRoles, role) {
		sortedRoles := append([]string(nil), allowedRoles...)
		sort.Strings(sortedRoles)
		c.JSON(403, gin.H{"error": errorCode, "role": role, "allowed_roles": sortedRoles})
		return "", false
	}

	return workerID, true
}

func getWorkerRole(workerID string) string {
	if HasDB() {
		if workerIDUint, err := parseWorkerID(workerID); err == nil {
			type userRoleRow struct {
				Role string `gorm:"column:role"`
			}
			var row userRoleRow
			if err := workerDB.Raw("SELECT COALESCE(role, '') AS role FROM users WHERE id = ? LIMIT 1", workerIDUint).Scan(&row).Error; err == nil {
				role := strings.ToLower(strings.TrimSpace(row.Role))
				if role != "" {
					return role
				}
			}
		}
	}

	if isProductionEnv() {
		return ""
	}

	defaultRole := strings.ToLower(strings.TrimSpace(os.Getenv("INDEL_DEFAULT_DEV_ROLE")))
	if defaultRole != "" {
		return defaultRole
	}
	return "worker"
}

func containsRole(roles []string, target string) bool {
	target = strings.ToLower(strings.TrimSpace(target))
	for _, role := range roles {
		if strings.ToLower(strings.TrimSpace(role)) == target {
			return true
		}
	}
	return false
}

func allowedRolesFromEnv(envKey string, prodDefault []string, nonProdDefault []string) []string {
	raw := strings.TrimSpace(os.Getenv(envKey))
	if raw == "" {
		if isProductionEnv() {
			return prodDefault
		}
		return nonProdDefault
	}

	parts := strings.Split(raw, ",")
	roles := make([]string, 0, len(parts))
	for _, p := range parts {
		role := strings.ToLower(strings.TrimSpace(p))
		if role != "" {
			roles = append(roles, role)
		}
	}
	if len(roles) == 0 {
		if isProductionEnv() {
			return prodDefault
		}
		return nonProdDefault
	}
	return roles
}

func isProductionEnv() bool {
	return strings.EqualFold(strings.TrimSpace(os.Getenv("INDEL_ENV")), "production")
}
