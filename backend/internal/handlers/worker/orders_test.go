package worker

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func newAuthedOrderContext(orderID string) (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodPost, "/orders/"+orderID+"/deliver?customer_code=1234", nil)
	req.Header.Set("Authorization", "Bearer mock-jwt-token")
	c.Request = req
	c.Params = gin.Params{{Key: "order_id", Value: orderID}}
	return c, w
}

func TestDeliverOrderIsIdempotentAfterFirstDelivery(t *testing.T) {
	gin.SetMode(gin.TestMode)
	SetDB(nil)
	store.reset()
	t.Cleanup(func() {
		store.reset()
		SetDB(nil)
	})

	c1, w1 := newAuthedOrderContext("ord-001")
	DeliverOrder(c1)
	if w1.Code != 200 {
		t.Fatalf("first DeliverOrder status = %d, want 200", w1.Code)
	}

	store.mu.RLock()
	firstActual, _ := store.data.Earnings["this_week_actual"].(int)
	firstCompleted, _ := store.data.WorkerProfiles["worker-001"]["orders_completed"].(int)
	firstNotifications := len(store.data.Notifications)
	store.mu.RUnlock()

	if firstActual != 3180 {
		t.Fatalf("this_week_actual after first delivery = %d, want 3180", firstActual)
	}
	if firstCompleted != 1 {
		t.Fatalf("orders_completed after first delivery = %d, want 1", firstCompleted)
	}

	c2, w2 := newAuthedOrderContext("ord-001")
	DeliverOrder(c2)
	if w2.Code != 200 {
		t.Fatalf("second DeliverOrder status = %d, want 200", w2.Code)
	}

	store.mu.RLock()
	secondActual, _ := store.data.Earnings["this_week_actual"].(int)
	secondCompleted, _ := store.data.WorkerProfiles["worker-001"]["orders_completed"].(int)
	secondNotifications := len(store.data.Notifications)
	store.mu.RUnlock()

	if secondActual != firstActual {
		t.Fatalf("this_week_actual changed on repeated delivery: got %d, first %d", secondActual, firstActual)
	}
	if secondCompleted != firstCompleted {
		t.Fatalf("orders_completed changed on repeated delivery: got %d, first %d", secondCompleted, firstCompleted)
	}
	if secondNotifications != firstNotifications {
		t.Fatalf("notifications changed on repeated delivery: got %d, first %d", secondNotifications, firstNotifications)
	}
}

func TestGetAssignedOrdersRequiresBearerToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	SetDB(nil)
	store.reset()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/orders/assigned", nil)

	GetAssignedOrders(c)

	if w.Code != 401 {
		t.Fatalf("status = %d, want 401", w.Code)
	}
}

func TestDeliverOrderRejectsInvalidCustomerCode(t *testing.T) {
	gin.SetMode(gin.TestMode)
	SetDB(nil)
	store.reset()
	t.Cleanup(func() {
		store.reset()
		SetDB(nil)
	})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodPost, "/orders/ord-001/deliver?customer_code=9999", nil)
	req.Header.Set("Authorization", "Bearer mock-jwt-token")
	c.Request = req
	c.Params = gin.Params{{Key: "order_id", Value: "ord-001"}}

	DeliverOrder(c)

	if w.Code != 400 {
		t.Fatalf("status = %d, want 400", w.Code)
	}
}

func TestAcceptOrderStoresWorkerAssignmentInMemory(t *testing.T) {
	gin.SetMode(gin.TestMode)
	SetDB(nil)
	store.reset()
	t.Cleanup(func() {
		store.reset()
		SetDB(nil)
	})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodPut, "/orders/ord-001/accept", nil)
	req.Header.Set("Authorization", "Bearer mock-jwt-token")
	c.Request = req
	c.Params = gin.Params{{Key: "order_id", Value: "ord-001"}}

	AcceptOrder(c)

	if w.Code != 200 {
		t.Fatalf("status = %d, want 200", w.Code)
	}

	store.mu.RLock()
	defer store.mu.RUnlock()
	order := store.data.Orders[0]
	if got := order["worker_id"]; got != "worker-001" {
		t.Fatalf("worker_id = %v, want worker-001", got)
	}
	if _, ok := order["accepted_at"].(string); !ok {
		t.Fatalf("accepted_at missing after accept")
	}
}

func TestGetAvailableOrdersReturnsStablePaddedOrderIDs(t *testing.T) {
	gin.SetMode(gin.TestMode)
	SetDB(nil)
	store.reset()
	t.Cleanup(func() {
		store.reset()
		SetDB(nil)
	})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/orders/available", nil)

	GetAvailableOrders(c)

	if w.Code != 200 {
		t.Fatalf("status = %d, want 200", w.Code)
	}

	var payload struct {
		Orders []struct {
			OrderID string `json:"order_id"`
		} `json:"orders"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if len(payload.Orders) == 0 {
		t.Fatalf("expected at least one order")
	}
	if payload.Orders[0].OrderID != "ord-001" {
		t.Fatalf("order_id = %q, want ord-001", payload.Orders[0].OrderID)
	}
}

func TestGetAssignedOrdersExcludesFreshAssignedOrders(t *testing.T) {
	gin.SetMode(gin.TestMode)
	SetDB(nil)
	store.reset()
	t.Cleanup(func() {
		store.reset()
		SetDB(nil)
	})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodGet, "/orders/assigned", nil)
	req.Header.Set("Authorization", "Bearer mock-jwt-token")
	c.Request = req

	GetAssignedOrders(c)

	if w.Code != 200 {
		t.Fatalf("status = %d, want 200", w.Code)
	}

	var payload struct {
		Orders []struct {
			Status string `json:"status"`
		} `json:"orders"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if len(payload.Orders) != 0 {
		t.Fatalf("expected no active orders before acceptance, got %d", len(payload.Orders))
	}
}

func TestOrderRouteTypeSupportsIntercity(t *testing.T) {
	if got := orderRouteType("A"); got != "local" {
		t.Fatalf("orderRouteType(A) = %q, want local", got)
	}
	if got := orderRouteType("B"); got != "intercity" {
		t.Fatalf("orderRouteType(B) = %q, want intercity", got)
	}
	if got := orderRouteType("C"); got != "interstate" {
		t.Fatalf("orderRouteType(C) = %q, want interstate", got)
	}
}

func TestNormalizeZoneLevelAcceptsWorkerFriendlyLabels(t *testing.T) {
	if got := normalizeZoneLevel("local"); got != "A" {
		t.Fatalf("normalizeZoneLevel(local) = %q, want A", got)
	}
	if got := normalizeZoneLevel("intercity"); got != "B" {
		t.Fatalf("normalizeZoneLevel(intercity) = %q, want B", got)
	}
	if got := normalizeZoneLevel("interstate"); got != "C" {
		t.Fatalf("normalizeZoneLevel(interstate) = %q, want C", got)
	}
}
