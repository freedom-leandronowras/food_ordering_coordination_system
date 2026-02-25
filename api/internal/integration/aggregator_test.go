package integration_test

import (
	"context"
	"errors"
	"sort"
	"testing"
	"time"

	"handler/internal/domain"
	"handler/internal/integration"
	"handler/internal/integration/adapters"

	"github.com/google/uuid"
)

// ---------------------------------------------------------------------------
// FetchAllMenus
// ---------------------------------------------------------------------------

func TestFetchAllMenus_FansOutToAllAdapters(t *testing.T) {
	agg := integration.NewAggregator()
	agg.Register(adapters.NewStubAdapter("pizza", "Pizza Place", adapters.QuickMenu("Margherita", 12.5)))
	agg.Register(adapters.NewStubAdapter("sushi", "Sushi Bar", adapters.QuickMenu("Salmon Roll", 15.0)))
	agg.Register(adapters.NewStubAdapter("tacos", "Taco Truck", adapters.QuickMenu("Al Pastor", 8.0)))

	results := agg.FetchAllMenus(context.Background())

	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}

	totalItems := 0
	for _, r := range results {
		if r.Err != nil {
			t.Fatalf("unexpected error from %s: %v", r.ServiceID, r.Err)
		}
		totalItems += len(r.Items)
	}
	if totalItems != 3 {
		t.Fatalf("expected 3 total items across all vendors, got %d", totalItems)
	}
}

func TestFetchAllMenus_PartialFailureDoesNotBlockOthers(t *testing.T) {
	agg := integration.NewAggregator()
	agg.Register(adapters.NewStubAdapter("ok", "Healthy Vendor", adapters.QuickMenu("Salad", 9.0)))
	agg.Register(adapters.NewStubAdapter("broken", "Down Vendor", nil).WithError(errors.New("connection refused")))

	results := agg.FetchAllMenus(context.Background())

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}

	var successes, failures int
	for _, r := range results {
		if r.Err != nil {
			failures++
			if r.ServiceID != "broken" {
				t.Fatalf("expected the failure to come from 'broken', got %s", r.ServiceID)
			}
		} else {
			successes++
		}
	}
	if successes != 1 || failures != 1 {
		t.Fatalf("expected 1 success + 1 failure, got %d + %d", successes, failures)
	}
}

func TestFetchAllMenus_RunsConcurrently(t *testing.T) {
	delay := 200 * time.Millisecond
	agg := integration.NewAggregator()
	agg.Register(adapters.NewStubAdapter("a", "A", adapters.QuickMenu("A", 1)).WithDelay(delay))
	agg.Register(adapters.NewStubAdapter("b", "B", adapters.QuickMenu("B", 2)).WithDelay(delay))
	agg.Register(adapters.NewStubAdapter("c", "C", adapters.QuickMenu("C", 3)).WithDelay(delay))

	start := time.Now()
	results := agg.FetchAllMenus(context.Background())
	elapsed := time.Since(start)

	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}
	for _, r := range results {
		if r.Err != nil {
			t.Fatalf("unexpected error from %s: %v", r.ServiceID, r.Err)
		}
	}

	// If they ran sequentially the total would be ≥ 3*delay = 600ms.
	// Concurrent execution should finish in roughly 1*delay.
	maxAcceptable := delay + 150*time.Millisecond
	if elapsed > maxAcceptable {
		t.Fatalf("expected completion within %v (concurrent), took %v (sequential?)", maxAcceptable, elapsed)
	}
}

func TestFetchAllMenus_RespectsContextCancellation(t *testing.T) {
	agg := integration.NewAggregator()
	agg.Register(adapters.NewStubAdapter("slow", "Slow Vendor", adapters.QuickMenu("x", 1)).WithDelay(5 * time.Second))

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	results := agg.FetchAllMenus(ctx)

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if !errors.Is(results[0].Err, context.DeadlineExceeded) {
		t.Fatalf("expected DeadlineExceeded, got %v", results[0].Err)
	}
}

func TestFetchAllMenus_EmptyRegistry(t *testing.T) {
	agg := integration.NewAggregator()
	results := agg.FetchAllMenus(context.Background())
	if len(results) != 0 {
		t.Fatalf("expected 0 results for empty registry, got %d", len(results))
	}
}

// ---------------------------------------------------------------------------
// SubmitToVendors
// ---------------------------------------------------------------------------

func TestSubmitToVendors_FansOutToSelectedVendors(t *testing.T) {
	agg := integration.NewAggregator()
	agg.Register(adapters.NewStubAdapter("pizza", "Pizza Place", nil))
	agg.Register(adapters.NewStubAdapter("sushi", "Sushi Bar", nil))
	agg.Register(adapters.NewStubAdapter("tacos", "Taco Truck", nil))

	orderID := uuid.New()
	vendorOrders := map[string]integration.OrderSubmission{
		"pizza": {OrderID: orderID, Items: []integration.OrderSubmissionItem{{ItemID: uuid.New(), Quantity: 2}}, Notes: "extra cheese"},
		"tacos": {OrderID: orderID, Items: []integration.OrderSubmissionItem{{ItemID: uuid.New(), Quantity: 1}}, Notes: "no cilantro"},
	}

	results := agg.SubmitToVendors(context.Background(), vendorOrders)

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}

	ids := make([]string, 0, len(results))
	for _, r := range results {
		if r.Err != nil {
			t.Fatalf("unexpected error from %s: %v", r.ServiceID, r.Err)
		}
		if !r.Confirmation.Confirmed {
			t.Fatalf("expected confirmed from %s", r.ServiceID)
		}
		ids = append(ids, r.ServiceID)
	}

	sort.Strings(ids)
	if ids[0] != "pizza" || ids[1] != "tacos" {
		t.Fatalf("expected [pizza tacos], got %v", ids)
	}
}

func TestSubmitToVendors_SkipsUnknownServiceIDs(t *testing.T) {
	agg := integration.NewAggregator()
	agg.Register(adapters.NewStubAdapter("pizza", "Pizza Place", nil))

	vendorOrders := map[string]integration.OrderSubmission{
		"pizza":   {OrderID: uuid.New()},
		"unknown": {OrderID: uuid.New()},
	}

	results := agg.SubmitToVendors(context.Background(), vendorOrders)

	if len(results) != 1 {
		t.Fatalf("expected 1 result (unknown skipped), got %d", len(results))
	}
	if results[0].ServiceID != "pizza" {
		t.Fatalf("expected 'pizza', got %s", results[0].ServiceID)
	}
}

func TestSubmitToVendors_PartialFailure(t *testing.T) {
	agg := integration.NewAggregator()
	agg.Register(adapters.NewStubAdapter("ok", "OK Vendor", nil))
	agg.Register(adapters.NewStubAdapter("fail", "Fail Vendor", nil).WithError(errors.New("500 Internal Server Error")))

	vendorOrders := map[string]integration.OrderSubmission{
		"ok":   {OrderID: uuid.New(), Items: []integration.OrderSubmissionItem{{ItemID: uuid.New(), Quantity: 1}}},
		"fail": {OrderID: uuid.New(), Items: []integration.OrderSubmissionItem{{ItemID: uuid.New(), Quantity: 1}}},
	}

	results := agg.SubmitToVendors(context.Background(), vendorOrders)

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}

	for _, r := range results {
		switch r.ServiceID {
		case "ok":
			if r.Err != nil {
				t.Fatalf("expected no error from ok, got %v", r.Err)
			}
			if !r.Confirmation.Confirmed {
				t.Fatal("expected confirmed from ok")
			}
		case "fail":
			if r.Err == nil {
				t.Fatal("expected error from fail, got nil")
			}
		}
	}
}

func TestSubmitToVendors_RunsConcurrently(t *testing.T) {
	delay := 200 * time.Millisecond
	agg := integration.NewAggregator()
	agg.Register(adapters.NewStubAdapter("a", "A", nil).WithDelay(delay))
	agg.Register(adapters.NewStubAdapter("b", "B", nil).WithDelay(delay))

	vendorOrders := map[string]integration.OrderSubmission{
		"a": {OrderID: uuid.New(), Items: []integration.OrderSubmissionItem{{ItemID: uuid.New(), Quantity: 1}}},
		"b": {OrderID: uuid.New(), Items: []integration.OrderSubmissionItem{{ItemID: uuid.New(), Quantity: 1}}},
	}

	start := time.Now()
	results := agg.SubmitToVendors(context.Background(), vendorOrders)
	elapsed := time.Since(start)

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}

	maxAcceptable := delay + 150*time.Millisecond
	if elapsed > maxAcceptable {
		t.Fatalf("expected completion within %v (concurrent), took %v", maxAcceptable, elapsed)
	}
}

// ---------------------------------------------------------------------------
// Register / Adapters
// ---------------------------------------------------------------------------

func TestRegister_ReplacesExistingAdapter(t *testing.T) {
	agg := integration.NewAggregator()
	agg.Register(adapters.NewStubAdapter("v1", "Version 1", adapters.QuickMenu("Old", 1)))
	agg.Register(adapters.NewStubAdapter("v1", "Version 2", adapters.QuickMenu("New", 2)))

	results := agg.FetchAllMenus(context.Background())
	if len(results) != 1 {
		t.Fatalf("expected 1 result after replacement, got %d", len(results))
	}
	if results[0].ServiceName != "Version 2" {
		t.Fatalf("expected replaced adapter, got %s", results[0].ServiceName)
	}
}

func TestAdapters_ReturnsRegisteredIDs(t *testing.T) {
	agg := integration.NewAggregator()
	agg.Register(adapters.NewStubAdapter("a", "A", nil))
	agg.Register(adapters.NewStubAdapter("b", "B", nil))

	ids := agg.Adapters()
	sort.Strings(ids)
	if len(ids) != 2 || ids[0] != "a" || ids[1] != "b" {
		t.Fatalf("expected [a b], got %v", ids)
	}
}

// ---------------------------------------------------------------------------
// Adapter interface compliance (compile-time check)
// ---------------------------------------------------------------------------

var _ integration.ExternalFoodService = (*adapters.StubAdapter)(nil)

// ---------------------------------------------------------------------------
// Multi-item menu aggregation
// ---------------------------------------------------------------------------

func TestFetchAllMenus_AggregatesItemsFromMultipleVendors(t *testing.T) {
	agg := integration.NewAggregator()
	agg.Register(adapters.NewStubAdapter("v1", "V1", []domain.MenuItem{
		{ID: uuid.New(), Name: "A", Price: 1, Available: true},
		{ID: uuid.New(), Name: "B", Price: 2, Available: true},
	}))
	agg.Register(adapters.NewStubAdapter("v2", "V2", []domain.MenuItem{
		{ID: uuid.New(), Name: "C", Price: 3, Available: true},
	}))

	results := agg.FetchAllMenus(context.Background())

	total := 0
	for _, r := range results {
		total += len(r.Items)
	}
	if total != 3 {
		t.Fatalf("expected 3 items across vendors, got %d", total)
	}
}
