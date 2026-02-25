package adapters_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"testing"
	"time"

	"handler/internal/integration"
	"handler/internal/integration/adapters"

	"github.com/google/uuid"
)

// ---------------------------------------------------------------------------
// mockDB represents the structure of our mock/*_db.json files and powers
// the httptest servers that simulate real external vendor APIs.
// ---------------------------------------------------------------------------

type mockDB struct {
	Vendor struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
	} `json:"vendor"`
	Menu []struct {
		ID          string  `json:"id"`
		Name        string  `json:"name"`
		Description string  `json:"description"`
		Price       float64 `json:"price"`
		Available   bool    `json:"available"`
	} `json:"menu"`
	Orders []json.RawMessage `json:"orders"`
}

// loadMockDB reads a JSON DB file from the mock/ directory at the project root.
func loadMockDB(t *testing.T, filename string) mockDB {
	t.Helper()
	path := filepath.Join("..", "..", "..", "..", "mock", filename)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", filename, err)
	}
	var db mockDB
	if err := json.Unmarshal(data, &db); err != nil {
		t.Fatalf("parse %s: %v", filename, err)
	}
	return db
}

// serveMockVendor creates an httptest server that behaves like a json-server:
//
//	GET  /menu   → returns the menu array
//	POST /orders → accepts an order body, returns a confirmation
func serveMockVendor(t *testing.T, db mockDB) *httptest.Server {
	t.Helper()

	var mu sync.Mutex
	orders := make([]json.RawMessage, 0)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /menu", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(db.Menu)
	})

	mux.HandleFunc("POST /orders", func(w http.ResponseWriter, r *http.Request) {
		var body json.RawMessage
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, `{"error":"bad request"}`, http.StatusBadRequest)
			return
		}

		mu.Lock()
		orders = append(orders, body)
		mu.Unlock()

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"external_ref": db.Vendor.ID + "-" + uuid.NewString()[:8],
			"confirmed":    true,
		})
	})

	mux.HandleFunc("GET /orders", func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(orders)
	})

	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

// ---------------------------------------------------------------------------
// FetchAllMenus — fan-out across 3 mock JSON servers, fan-in results
// ---------------------------------------------------------------------------

func TestJSONServer_FetchAllMenus_FanOutFanIn(t *testing.T) {
	pizzaDB := loadMockDB(t, "pizza_place_db.json")
	sushiDB := loadMockDB(t, "sushi_bar_db.json")
	tacoDB := loadMockDB(t, "taco_truck_db.json")

	pizzaSrv := serveMockVendor(t, pizzaDB)
	sushiSrv := serveMockVendor(t, sushiDB)
	tacoSrv := serveMockVendor(t, tacoDB)

	agg := integration.NewAggregator()
	agg.Register(adapters.NewJSONServerAdapter("pizza", pizzaDB.Vendor.Name, pizzaSrv.URL, nil))
	agg.Register(adapters.NewJSONServerAdapter("sushi", sushiDB.Vendor.Name, sushiSrv.URL, nil))
	agg.Register(adapters.NewJSONServerAdapter("tacos", tacoDB.Vendor.Name, tacoSrv.URL, nil))

	results := agg.FetchAllMenus(context.Background())

	if len(results) != 3 {
		t.Fatalf("expected 3 vendor results, got %d", len(results))
	}

	itemsByVendor := make(map[string]int)
	for _, r := range results {
		if r.Err != nil {
			t.Fatalf("unexpected error from %s: %v", r.ServiceID, r.Err)
		}
		itemsByVendor[r.ServiceID] = len(r.Items)
	}

	// pizza_place_db.json has 4 items, sushi_bar_db.json has 5, taco_truck_db.json has 5
	if itemsByVendor["pizza"] != 4 {
		t.Fatalf("pizza: expected 4 items, got %d", itemsByVendor["pizza"])
	}
	if itemsByVendor["sushi"] != 5 {
		t.Fatalf("sushi: expected 5 items, got %d", itemsByVendor["sushi"])
	}
	if itemsByVendor["tacos"] != 5 {
		t.Fatalf("tacos: expected 5 items, got %d", itemsByVendor["tacos"])
	}
}

func TestJSONServer_FetchAllMenus_ParsesMenuItemFieldsCorrectly(t *testing.T) {
	pizzaDB := loadMockDB(t, "pizza_place_db.json")
	srv := serveMockVendor(t, pizzaDB)

	adapter := adapters.NewJSONServerAdapter("pizza", "Pizza", srv.URL, nil)
	items, err := adapter.FetchMenu(context.Background())
	if err != nil {
		t.Fatalf("FetchMenu: %v", err)
	}

	// Find the Margherita (first item in the DB).
	found := false
	for _, item := range items {
		if item.Name == "Margherita" {
			found = true
			if item.Price != 12.50 {
				t.Fatalf("Margherita price: want 12.50, got %v", item.Price)
			}
			if !item.Available {
				t.Fatal("Margherita should be available")
			}
			if item.Description != "San Marzano tomato, mozzarella di bufala, fresh basil" {
				t.Fatalf("Margherita description mismatch: %s", item.Description)
			}
			if item.ID == uuid.Nil {
				t.Fatal("Margherita ID should not be nil")
			}
		}
	}
	if !found {
		t.Fatal("Margherita not found in parsed menu")
	}

	// Check that unavailable items are parsed correctly.
	for _, item := range items {
		if item.Name == "Capricciosa" && item.Available {
			t.Fatal("Capricciosa should be unavailable per the DB")
		}
	}
}

func TestJSONServer_FetchAllMenus_RunsConcurrently(t *testing.T) {
	delay := 150 * time.Millisecond

	// Create servers that sleep before responding.
	slowHandler := func(db mockDB) *httptest.Server {
		mux := http.NewServeMux()
		mux.HandleFunc("GET /menu", func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(delay)
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(db.Menu)
		})
		mux.HandleFunc("POST /orders", func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(delay)
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(map[string]any{"external_ref": "slow", "confirmed": true})
		})
		srv := httptest.NewServer(mux)
		t.Cleanup(srv.Close)
		return srv
	}

	pizzaDB := loadMockDB(t, "pizza_place_db.json")
	sushiDB := loadMockDB(t, "sushi_bar_db.json")
	tacoDB := loadMockDB(t, "taco_truck_db.json")

	agg := integration.NewAggregator()
	agg.Register(adapters.NewJSONServerAdapter("pizza", "Pizza", slowHandler(pizzaDB).URL, nil))
	agg.Register(adapters.NewJSONServerAdapter("sushi", "Sushi", slowHandler(sushiDB).URL, nil))
	agg.Register(adapters.NewJSONServerAdapter("tacos", "Tacos", slowHandler(tacoDB).URL, nil))

	start := time.Now()
	results := agg.FetchAllMenus(context.Background())
	elapsed := time.Since(start)

	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}
	for _, r := range results {
		if r.Err != nil {
			t.Fatalf("error from %s: %v", r.ServiceID, r.Err)
		}
	}

	// 3 sequential calls would take >= 450ms. Concurrent should be ~150ms.
	maxAcceptable := delay + 200*time.Millisecond
	if elapsed > maxAcceptable {
		t.Fatalf("expected completion within %v (concurrent), took %v", maxAcceptable, elapsed)
	}
}

// ---------------------------------------------------------------------------
// SubmitToVendors — fan-out orders to selected mock JSON servers
// ---------------------------------------------------------------------------

func TestJSONServer_SubmitToVendors_FanOutFanIn(t *testing.T) {
	pizzaDB := loadMockDB(t, "pizza_place_db.json")
	sushiDB := loadMockDB(t, "sushi_bar_db.json")
	tacoDB := loadMockDB(t, "taco_truck_db.json")

	pizzaSrv := serveMockVendor(t, pizzaDB)
	sushiSrv := serveMockVendor(t, sushiDB)
	tacoSrv := serveMockVendor(t, tacoDB)

	agg := integration.NewAggregator()
	agg.Register(adapters.NewJSONServerAdapter("pizza", pizzaDB.Vendor.Name, pizzaSrv.URL, nil))
	agg.Register(adapters.NewJSONServerAdapter("sushi", sushiDB.Vendor.Name, sushiSrv.URL, nil))
	agg.Register(adapters.NewJSONServerAdapter("tacos", tacoDB.Vendor.Name, tacoSrv.URL, nil))

	orderID := uuid.New()
	vendorOrders := map[string]integration.OrderSubmission{
		"pizza": {
			OrderID: orderID,
			Items: []integration.OrderSubmissionItem{
				{ItemID: uuid.MustParse("f0000001-aaaa-4000-a000-000000000001"), Quantity: 2},
				{ItemID: uuid.MustParse("f0000001-aaaa-4000-a000-000000000002"), Quantity: 1},
			},
			Notes: "extra crispy",
		},
		"tacos": {
			OrderID: orderID,
			Items: []integration.OrderSubmissionItem{
				{ItemID: uuid.MustParse("f0000003-cccc-4000-a000-000000000001"), Quantity: 3},
			},
			Notes: "no cilantro please",
		},
	}

	results := agg.SubmitToVendors(context.Background(), vendorOrders)

	if len(results) != 2 {
		t.Fatalf("expected 2 results (pizza + tacos), got %d", len(results))
	}

	ids := make([]string, 0, len(results))
	for _, r := range results {
		if r.Err != nil {
			t.Fatalf("error from %s: %v", r.ServiceID, r.Err)
		}
		if !r.Confirmation.Confirmed {
			t.Fatalf("expected confirmed from %s", r.ServiceID)
		}
		if r.Confirmation.ExternalRef == "" {
			t.Fatalf("expected non-empty external ref from %s", r.ServiceID)
		}
		ids = append(ids, r.ServiceID)
	}

	sort.Strings(ids)
	if ids[0] != "pizza" || ids[1] != "tacos" {
		t.Fatalf("expected [pizza tacos], got %v", ids)
	}
}

func TestJSONServer_SubmitToVendors_OrderPersistedOnServer(t *testing.T) {
	pizzaDB := loadMockDB(t, "pizza_place_db.json")
	pizzaSrv := serveMockVendor(t, pizzaDB)

	adapter := adapters.NewJSONServerAdapter("pizza", "Pizza", pizzaSrv.URL, nil)

	orderID := uuid.New()
	_, err := adapter.SubmitOrder(context.Background(), integration.OrderSubmission{
		OrderID: orderID,
		Items: []integration.OrderSubmissionItem{
			{ItemID: uuid.MustParse("f0000001-aaaa-4000-a000-000000000001"), Quantity: 1},
		},
		Notes: "test order",
	})
	if err != nil {
		t.Fatalf("SubmitOrder: %v", err)
	}

	// Verify the order was persisted on the mock server by querying GET /orders.
	resp, err := http.Get(pizzaSrv.URL + "/orders")
	if err != nil {
		t.Fatalf("GET /orders: %v", err)
	}
	defer resp.Body.Close()

	var orders []json.RawMessage
	if err := json.NewDecoder(resp.Body).Decode(&orders); err != nil {
		t.Fatalf("decode orders: %v", err)
	}
	if len(orders) != 1 {
		t.Fatalf("expected 1 order on server, got %d", len(orders))
	}
}

// ---------------------------------------------------------------------------
// Partial failure — one vendor down, others still return results
// ---------------------------------------------------------------------------

func TestJSONServer_FetchAllMenus_PartialFailure(t *testing.T) {
	pizzaDB := loadMockDB(t, "pizza_place_db.json")
	sushiDB := loadMockDB(t, "sushi_bar_db.json")

	pizzaSrv := serveMockVendor(t, pizzaDB)
	sushiSrv := serveMockVendor(t, sushiDB)

	// Create a server that's already closed — simulates a vendor that's down.
	deadSrv := httptest.NewServer(http.NotFoundHandler())
	deadSrv.Close()

	agg := integration.NewAggregator()
	agg.Register(adapters.NewJSONServerAdapter("pizza", "Pizza", pizzaSrv.URL, nil))
	agg.Register(adapters.NewJSONServerAdapter("sushi", "Sushi", sushiSrv.URL, nil))
	agg.Register(adapters.NewJSONServerAdapter("dead", "Dead Vendor", deadSrv.URL, nil))

	results := agg.FetchAllMenus(context.Background())

	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}

	var successes, failures int
	for _, r := range results {
		if r.Err != nil {
			failures++
			if r.ServiceID != "dead" {
				t.Fatalf("expected failure from 'dead', got from %s: %v", r.ServiceID, r.Err)
			}
		} else {
			successes++
		}
	}

	if successes != 2 || failures != 1 {
		t.Fatalf("expected 2 successes + 1 failure, got %d + %d", successes, failures)
	}
}

// ---------------------------------------------------------------------------
// Context cancellation propagates to HTTP calls
// ---------------------------------------------------------------------------

func TestJSONServer_FetchMenu_RespectsContextCancellation(t *testing.T) {
	slowMux := http.NewServeMux()
	slowMux.HandleFunc("GET /menu", func(w http.ResponseWriter, r *http.Request) {
		// Block until the request context is done (the client cancelled).
		<-r.Context().Done()
	})
	srv := httptest.NewServer(slowMux)
	t.Cleanup(srv.Close)

	adapter := adapters.NewJSONServerAdapter("slow", "Slow", srv.URL, nil)

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err := adapter.FetchMenu(ctx)
	if err == nil {
		t.Fatal("expected error from cancelled context, got nil")
	}
}

// ---------------------------------------------------------------------------
// Compile-time interface check
// ---------------------------------------------------------------------------

var _ integration.ExternalFoodService = (*adapters.JSONServerAdapter)(nil)
