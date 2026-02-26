package httpapi_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"food_ordering_coordination_system/internal/domain"
	httpapi "food_ordering_coordination_system/internal/http"
	"food_ordering_coordination_system/internal/integration"
	"food_ordering_coordination_system/internal/integration/adapters"
	persistence "food_ordering_coordination_system/internal/persistance"

	"github.com/google/uuid"
	"github.com/testcontainers/testcontainers-go/modules/mongodb"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ===========================================================================
// Grant Credits — POST /api/members/{memberId}/credits
// ===========================================================================

func TestGrantCredits_ManagerCanGrantCredits(t *testing.T) {
	env := newCrudTestEnv(t)
	defer env.close()

	memberID := uuid.New()
	managerID := uuid.New()
	token := tokenFor(t, managerID, httpapi.RoleHiveManager)

	rec := executeJSONRequest(t, env.router, http.MethodPost,
		"/api/members/"+memberID.String()+"/credits",
		map[string]any{"amount": 75.0}, token)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}

	var resp struct {
		MemberID   string  `json:"member_id"`
		NewBalance float64 `json:"new_balance"`
	}
	decodeResponse(t, rec, &resp)
	if resp.NewBalance != 75.0 {
		t.Fatalf("expected new_balance 75, got %v", resp.NewBalance)
	}

	// Verify via GET credits.
	creditsRec := httptest.NewRecorder()
	creditsReq := httptest.NewRequest(http.MethodGet, "/api/members/"+memberID.String()+"/credits", nil)
	creditsReq.Header.Set("Authorization", "Bearer "+tokenFor(t, managerID, httpapi.RoleHiveManager))
	env.router.ServeHTTP(creditsRec, creditsReq)

	var credResp struct {
		Credits float64 `json:"credits"`
	}
	decodeResponse(t, creditsRec, &credResp)
	if credResp.Credits != 75 {
		t.Fatalf("expected credits 75, got %v", credResp.Credits)
	}
}

func TestGrantCredits_MemberCannotGrant(t *testing.T) {
	env := newCrudTestEnv(t)
	defer env.close()

	memberID := uuid.New()
	token := tokenFor(t, memberID, httpapi.RoleMember)

	rec := executeJSONRequest(t, env.router, http.MethodPost,
		"/api/members/"+memberID.String()+"/credits",
		map[string]any{"amount": 50.0}, token)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestGrantCredits_ExceedsCap(t *testing.T) {
	env := newCrudTestEnv(t)
	defer env.close()

	memberID := uuid.New()
	token := tokenFor(t, uuid.New(), httpapi.RoleHiveManager)

	// Grant 900.
	rec := executeJSONRequest(t, env.router, http.MethodPost,
		"/api/members/"+memberID.String()+"/credits",
		map[string]any{"amount": 900.0}, token)
	if rec.Code != http.StatusOK {
		t.Fatalf("first grant: expected 200, got %d", rec.Code)
	}

	// Grant 200 more → exceeds 1000 cap.
	rec2 := executeJSONRequest(t, env.router, http.MethodPost,
		"/api/members/"+memberID.String()+"/credits",
		map[string]any{"amount": 200.0}, token)
	if rec2.Code != http.StatusUnprocessableEntity {
		t.Fatalf("second grant: expected 422, got %d body=%s", rec2.Code, rec2.Body.String())
	}
}

func TestGrantCredits_InvalidAmount(t *testing.T) {
	env := newCrudTestEnv(t)
	defer env.close()

	token := tokenFor(t, uuid.New(), httpapi.RoleHiveManager)

	rec := executeJSONRequest(t, env.router, http.MethodPost,
		"/api/members/"+uuid.New().String()+"/credits",
		map[string]any{"amount": -10.0}, token)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", rec.Code, rec.Body.String())
	}
}

// ===========================================================================
// Get Member Orders — GET /api/members/{memberId}/orders
// ===========================================================================

func TestGetMemberOrders_ReturnsPlacedOrders(t *testing.T) {
	env := newCrudTestEnv(t)
	defer env.close()

	memberID := uuid.New()
	managerToken := tokenFor(t, uuid.New(), httpapi.RoleHiveManager)
	memberToken := tokenFor(t, memberID, httpapi.RoleMember)

	// Seed credits.
	executeJSONRequest(t, env.router, http.MethodPost,
		"/api/members/"+memberID.String()+"/credits",
		map[string]any{"amount": 200.0}, managerToken)

	// Place 2 orders.
	for i := 0; i < 2; i++ {
		rec := executeJSONRequest(t, env.router, http.MethodPost, "/api/orders",
			map[string]any{
				"member_id": memberID.String(),
				"items":     []map[string]any{{"id": uuid.New().String(), "quantity": 1, "price": 10.0}},
			}, memberToken)
		if rec.Code != http.StatusCreated {
			t.Fatalf("place order %d: expected 201, got %d body=%s", i, rec.Code, rec.Body.String())
		}
	}

	// Fetch orders.
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/members/"+memberID.String()+"/orders", nil)
	req.Header.Set("Authorization", "Bearer "+memberToken)
	env.router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}

	var orders []struct {
		OrderID    string  `json:"order_id"`
		Status     string  `json:"status"`
		TotalPrice float64 `json:"total_price"`
	}
	decodeResponse(t, rec, &orders)

	if len(orders) != 2 {
		t.Fatalf("expected 2 orders, got %d", len(orders))
	}
	for _, o := range orders {
		if o.Status != "CONFIRMED" {
			t.Fatalf("expected CONFIRMED, got %s", o.Status)
		}
		if o.TotalPrice != 10.0 {
			t.Fatalf("expected total_price 10, got %v", o.TotalPrice)
		}
	}
}

func TestGetMemberOrders_MemberCannotReadOtherMemberOrders(t *testing.T) {
	env := newCrudTestEnv(t)
	defer env.close()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/members/"+uuid.New().String()+"/orders", nil)
	req.Header.Set("Authorization", "Bearer "+tokenFor(t, uuid.New(), httpapi.RoleMember))
	env.router.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
}

func TestGetMemberOrders_EmptyList(t *testing.T) {
	env := newCrudTestEnv(t)
	defer env.close()

	memberID := uuid.New()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/members/"+memberID.String()+"/orders", nil)
	req.Header.Set("Authorization", "Bearer "+tokenFor(t, memberID, httpapi.RoleMember))
	env.router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var orders []json.RawMessage
	decodeResponse(t, rec, &orders)
	if len(orders) != 0 {
		t.Fatalf("expected 0 orders, got %d", len(orders))
	}
}

// ===========================================================================
// Get All Menus — GET /api/menus (fan-out / fan-in)
// ===========================================================================

func TestGetAllMenus_FanOutFanIn(t *testing.T) {
	env := newCrudTestEnvWithVendors(t)
	defer env.close()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/menus", nil)
	env.router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}

	var menus []struct {
		ServiceID   string `json:"service_id"`
		ServiceName string `json:"service_name"`
		Items       []struct {
			ID          string  `json:"id"`
			Name        string  `json:"name"`
			Description string  `json:"description"`
			Price       float64 `json:"price"`
			Available   bool    `json:"available"`
		} `json:"items"`
		Error string `json:"error,omitempty"`
	}
	decodeResponse(t, rec, &menus)

	if len(menus) != 3 {
		t.Fatalf("expected 3 vendor menus, got %d", len(menus))
	}

	totalItems := 0
	for _, m := range menus {
		if m.Error != "" {
			t.Fatalf("unexpected error from %s: %s", m.ServiceID, m.Error)
		}
		totalItems += len(m.Items)
	}
	if totalItems != 3 {
		t.Fatalf("expected 3 total items, got %d", totalItems)
	}
}

func TestGetAllMenus_NoVendors(t *testing.T) {
	env := newCrudTestEnv(t)
	defer env.close()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/menus", nil)
	env.router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

// ===========================================================================
// Get Vendors — GET /api/vendors
// ===========================================================================

func TestGetVendors_ListsRegistered(t *testing.T) {
	env := newCrudTestEnvWithVendors(t)
	defer env.close()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/vendors", nil)
	env.router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var vendors []struct {
		ServiceID string `json:"service_id"`
	}
	decodeResponse(t, rec, &vendors)

	if len(vendors) != 3 {
		t.Fatalf("expected 3 vendors, got %d", len(vendors))
	}
}

func TestGetVendors_EmptyWhenNoAggregator(t *testing.T) {
	env := newCrudTestEnv(t)
	defer env.close()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/vendors", nil)
	env.router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

// ===========================================================================
// Full flow: grant → order → history → menus
// ===========================================================================

func TestFullFlow_GrantOrderHistoryMenus(t *testing.T) {
	env := newCrudTestEnvWithVendors(t)
	defer env.close()

	memberID := uuid.New()
	managerToken := tokenFor(t, uuid.New(), httpapi.RoleHiveManager)
	memberToken := tokenFor(t, memberID, httpapi.RoleMember)

	// 1. Grant 100 credits.
	rec := executeJSONRequest(t, env.router, http.MethodPost,
		"/api/members/"+memberID.String()+"/credits",
		map[string]any{"amount": 100.0}, managerToken)
	if rec.Code != http.StatusOK {
		t.Fatalf("grant: expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}

	// 2. Place an order for 25.
	rec = executeJSONRequest(t, env.router, http.MethodPost, "/api/orders",
		map[string]any{
			"member_id":      memberID.String(),
			"items":          []map[string]any{{"id": uuid.New().String(), "quantity": 1, "price": 25.0}},
			"delivery_notes": "room 3A",
		}, memberToken)
	if rec.Code != http.StatusCreated {
		t.Fatalf("order: expected 201, got %d body=%s", rec.Code, rec.Body.String())
	}

	// 3. Check credits = 75.
	creditsRec := httptest.NewRecorder()
	creditsReq := httptest.NewRequest(http.MethodGet, "/api/members/"+memberID.String()+"/credits", nil)
	creditsReq.Header.Set("Authorization", "Bearer "+memberToken)
	env.router.ServeHTTP(creditsRec, creditsReq)
	var credResp struct {
		Credits float64 `json:"credits"`
	}
	decodeResponse(t, creditsRec, &credResp)
	if credResp.Credits != 75 {
		t.Fatalf("expected 75 credits, got %v", credResp.Credits)
	}

	// 4. Check order history = 1 order.
	ordersRec := httptest.NewRecorder()
	ordersReq := httptest.NewRequest(http.MethodGet, "/api/members/"+memberID.String()+"/orders", nil)
	ordersReq.Header.Set("Authorization", "Bearer "+memberToken)
	env.router.ServeHTTP(ordersRec, ordersReq)
	var orders []struct {
		OrderID string `json:"order_id"`
	}
	decodeResponse(t, ordersRec, &orders)
	if len(orders) != 1 {
		t.Fatalf("expected 1 order, got %d", len(orders))
	}

	// 5. Check menus (fan-out/fan-in still works alongside DB ops).
	menusRec := httptest.NewRecorder()
	menusReq := httptest.NewRequest(http.MethodGet, "/api/menus", nil)
	env.router.ServeHTTP(menusRec, menusReq)
	if menusRec.Code != http.StatusOK {
		t.Fatalf("menus: expected 200, got %d", menusRec.Code)
	}
}

// ===========================================================================
// Test environments
// ===========================================================================

type crudTestEnv struct {
	container  *mongodb.MongoDBContainer
	client     *mongo.Client
	db         *mongo.Database
	repo       *persistence.MongoRepository
	router     http.Handler
	aggregator *integration.Aggregator
}

func newCrudTestEnv(t *testing.T) *crudTestEnv {
	t.Helper()

	ctx := context.Background()
	container, err := mongodb.Run(ctx, "mongo:7")
	if err != nil {
		t.Fatalf("start mongodb container: %v", err)
	}

	connString, err := container.ConnectionString(ctx)
	if err != nil {
		t.Fatalf("get mongodb connection string: %v", err)
	}

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(connString))
	if err != nil {
		t.Fatalf("connect mongodb: %v", err)
	}

	db := client.Database("crud_test_" + uuid.NewString())
	repo := persistence.NewMongoRepository(db)
	if err := repo.EnsureSchema(ctx); err != nil {
		t.Fatalf("ensure schema: %v", err)
	}

	service := domain.NewFoodOrderingService(repo, repo, repo)
	router := httpapi.NewFoodOrderingRouter(service, nil)

	return &crudTestEnv{
		container: container,
		client:    client,
		db:        db,
		repo:      repo,
		router:    router,
	}
}

func newCrudTestEnvWithVendors(t *testing.T) *crudTestEnv {
	t.Helper()

	ctx := context.Background()
	container, err := mongodb.Run(ctx, "mongo:7")
	if err != nil {
		t.Fatalf("start mongodb container: %v", err)
	}

	connString, err := container.ConnectionString(ctx)
	if err != nil {
		t.Fatalf("get mongodb connection string: %v", err)
	}

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(connString))
	if err != nil {
		t.Fatalf("connect mongodb: %v", err)
	}

	db := client.Database("crud_vendor_test_" + uuid.NewString())
	repo := persistence.NewMongoRepository(db)
	if err := repo.EnsureSchema(ctx); err != nil {
		t.Fatalf("ensure schema: %v", err)
	}

	agg := integration.NewAggregator()
	agg.Register(adapters.NewStubAdapter("pizza", "Pizza Place", adapters.QuickMenu("Margherita", 12.5)))
	agg.Register(adapters.NewStubAdapter("sushi", "Sushi Bar", adapters.QuickMenu("Salmon Roll", 15.0)))
	agg.Register(adapters.NewStubAdapter("tacos", "Taco Truck", adapters.QuickMenu("Al Pastor", 8.0)))

	service := domain.NewFoodOrderingService(repo, repo, repo)
	router := httpapi.NewFoodOrderingRouter(service, agg)

	return &crudTestEnv{
		container:  container,
		client:     client,
		db:         db,
		repo:       repo,
		router:     router,
		aggregator: agg,
	}
}

func (e *crudTestEnv) close() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = e.client.Disconnect(ctx)
	_ = e.container.Terminate(ctx)
}
