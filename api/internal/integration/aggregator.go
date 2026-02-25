package integration

import (
	"context"
	"sync"

	"handler/internal/domain"
)

// ---------- result envelopes ----------

// MenuResult wraps the outcome of a single FetchMenu fan-out call.
// Err is non-nil when the adapter failed; the remaining fields still
// identify which service was called so callers can handle partial failures.
type MenuResult struct {
	ServiceID   string
	ServiceName string
	Items       []domain.MenuItem
	Err         error
}

// SubmitResult wraps the outcome of a single SubmitOrder fan-out call.
type SubmitResult struct {
	ServiceID    string
	ServiceName  string
	Confirmation OrderConfirmation
	Err          error
}

// ---------- aggregator ----------

// Aggregator is the fan-in / fan-out orchestrator.  It holds a registry of
// ExternalFoodService adapters and provides methods that query or command all
// (or a subset of) them concurrently, collecting results through a single
// channel.
//
//	┌──────────┐         ┌────────────┐
//	│ Adapter A ├──────┐  │            │
//	└──────────┘      │  │            │
//	┌──────────┐      ▼  │  fan-in    │
//	│ Adapter B ├──────►──┤  channel   ├───► []Result
//	└──────────┘      ▲  │            │
//	┌──────────┐      │  │            │
//	│ Adapter C ├──────┘  │            │
//	└──────────┘         └────────────┘
type Aggregator struct {
	mu       sync.RWMutex
	adapters map[string]ExternalFoodService
}

// NewAggregator creates an empty Aggregator.  Use Register to add adapters.
func NewAggregator() *Aggregator {
	return &Aggregator{adapters: make(map[string]ExternalFoodService)}
}

// Register adds (or replaces) an adapter.  Safe for concurrent use.
func (a *Aggregator) Register(svc ExternalFoodService) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.adapters[svc.ServiceID()] = svc
}

// Adapters returns a snapshot of the currently registered service IDs.
func (a *Aggregator) Adapters() []string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	ids := make([]string, 0, len(a.adapters))
	for id := range a.adapters {
		ids = append(ids, id)
	}
	return ids
}

// ---------- fan-out / fan-in: menus ----------

// FetchAllMenus fans out a FetchMenu call to every registered adapter and
// fans the results back into a single slice.
//
// Each adapter runs in its own goroutine.  If a single adapter fails the
// others are unaffected — the caller receives one MenuResult per adapter and
// can inspect Err to decide how to handle partial failures.
//
// The supplied context is forwarded to every adapter so a timeout or
// cancellation propagates to all in-flight calls.
func (a *Aggregator) FetchAllMenus(ctx context.Context) []MenuResult {
	adapters := a.snapshot()

	ch := make(chan MenuResult, len(adapters))

	// Fan-out: one goroutine per adapter.
	var wg sync.WaitGroup
	for _, svc := range adapters {
		wg.Add(1)
		go func(s ExternalFoodService) {
			defer wg.Done()
			items, err := s.FetchMenu(ctx)
			ch <- MenuResult{
				ServiceID:   s.ServiceID(),
				ServiceName: s.ServiceName(),
				Items:       items,
				Err:         err,
			}
		}(svc)
	}

	// Close the channel once every goroutine has sent its result.
	go func() {
		wg.Wait()
		close(ch)
	}()

	// Fan-in: drain the channel into a slice.
	collected := make([]MenuResult, 0, len(adapters))
	for r := range ch {
		collected = append(collected, r)
	}
	return collected
}

// ---------- fan-out / fan-in: order submission ----------

// SubmitToVendors fans out order submissions to the adapters whose service
// IDs appear as keys in vendorOrders.  Unknown IDs are silently skipped.
//
// This lets the caller split a multi-vendor cart into per-vendor submissions
// and fire them all concurrently.
func (a *Aggregator) SubmitToVendors(ctx context.Context, vendorOrders map[string]OrderSubmission) []SubmitResult {
	a.mu.RLock()
	type work struct {
		svc   ExternalFoodService
		order OrderSubmission
	}
	tasks := make([]work, 0, len(vendorOrders))
	for id, order := range vendorOrders {
		if svc, ok := a.adapters[id]; ok {
			tasks = append(tasks, work{svc: svc, order: order})
		}
	}
	a.mu.RUnlock()

	ch := make(chan SubmitResult, len(tasks))

	var wg sync.WaitGroup
	for _, t := range tasks {
		wg.Add(1)
		go func(s ExternalFoodService, o OrderSubmission) {
			defer wg.Done()
			conf, err := s.SubmitOrder(ctx, o)
			ch <- SubmitResult{
				ServiceID:    s.ServiceID(),
				ServiceName:  s.ServiceName(),
				Confirmation: conf,
				Err:          err,
			}
		}(t.svc, t.order)
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	collected := make([]SubmitResult, 0, len(tasks))
	for r := range ch {
		collected = append(collected, r)
	}
	return collected
}

// snapshot returns a point-in-time copy of the adapter map values so callers
// can iterate without holding the lock for the entire fan-out.
func (a *Aggregator) snapshot() []ExternalFoodService {
	a.mu.RLock()
	defer a.mu.RUnlock()
	out := make([]ExternalFoodService, 0, len(a.adapters))
	for _, svc := range a.adapters {
		out = append(out, svc)
	}
	return out
}
