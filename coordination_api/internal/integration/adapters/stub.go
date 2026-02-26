package adapters

import (
	"context"
	"time"

	"food_ordering_coordination_system/internal/domain"
	"food_ordering_coordination_system/internal/integration"

	"github.com/google/uuid"
)

// StubAdapter is a configurable in-memory adapter for development and testing.
// It satisfies integration.ExternalFoodService so it can be registered with an
// Aggregator like any real vendor adapter.
type StubAdapter struct {
	id    string
	name  string
	menu  []domain.MenuItem
	err   error
	delay time.Duration
}

// NewStubAdapter returns an adapter that always succeeds with the given menu.
func NewStubAdapter(id, name string, menu []domain.MenuItem) *StubAdapter {
	return &StubAdapter{id: id, name: name, menu: menu}
}

func (s *StubAdapter) ServiceID() string   { return s.id }
func (s *StubAdapter) ServiceName() string { return s.name }

func (s *StubAdapter) FetchMenu(ctx context.Context) ([]domain.MenuItem, error) {
	if err := s.wait(ctx); err != nil {
		return nil, err
	}
	if s.err != nil {
		return nil, s.err
	}
	return s.menu, nil
}

func (s *StubAdapter) SubmitOrder(ctx context.Context, req integration.OrderSubmission) (integration.OrderConfirmation, error) {
	if err := s.wait(ctx); err != nil {
		return integration.OrderConfirmation{}, err
	}
	if s.err != nil {
		return integration.OrderConfirmation{}, s.err
	}
	return integration.OrderConfirmation{
		ExternalRef: "STUB-" + req.OrderID.String()[:8],
		Confirmed:   true,
	}, nil
}

// WithError makes every call to this adapter return err.
func (s *StubAdapter) WithError(err error) *StubAdapter {
	s.err = err
	return s
}

// WithDelay makes the adapter sleep before responding — useful for verifying
// that the fan-out actually runs concurrently.
func (s *StubAdapter) WithDelay(d time.Duration) *StubAdapter {
	s.delay = d
	return s
}

// wait respects both the configured delay and context cancellation.
func (s *StubAdapter) wait(ctx context.Context) error {
	if s.delay == 0 {
		return nil
	}
	select {
	case <-time.After(s.delay):
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// QuickMenu is a helper that builds a single-item MenuItem slice for tests.
func QuickMenu(name string, price float64) []domain.MenuItem {
	return []domain.MenuItem{{
		ID:        uuid.New(),
		Name:      name,
		Price:     price,
		Available: true,
	}}
}
