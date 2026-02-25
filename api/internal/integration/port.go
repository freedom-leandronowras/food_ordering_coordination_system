package integration

import (
	"context"

	"handler/internal/domain"

	"github.com/google/uuid"
)

// ExternalFoodService is the port that every vendor adapter must satisfy.
// Each adapter encapsulates the protocol specifics of a single external food
// service (REST, gRPC, SOAP, …) and translates them into the domain types
// the coordination system understands.
//
// Adding a new vendor means implementing this interface and registering the
// adapter with the Aggregator — no existing code needs to change.
type ExternalFoodService interface {
	// ServiceID returns a stable, machine-friendly identifier (e.g. "uber-eats").
	ServiceID() string
	// ServiceName returns a human-readable label (e.g. "Uber Eats").
	ServiceName() string
	// FetchMenu retrieves the currently available menu items from the vendor.
	FetchMenu(ctx context.Context) ([]domain.MenuItem, error)
	// SubmitOrder forwards an order to the vendor and returns a confirmation.
	SubmitOrder(ctx context.Context, req OrderSubmission) (OrderConfirmation, error)
}

// OrderSubmission carries the data an adapter needs to place an order with its
// vendor.  It is deliberately decoupled from domain.FoodOrder so the
// integration boundary stays clean.
type OrderSubmission struct {
	OrderID uuid.UUID
	Items   []OrderSubmissionItem
	Notes   string
}

// OrderSubmissionItem is a single line-item inside an OrderSubmission.
type OrderSubmissionItem struct {
	ItemID   uuid.UUID
	Quantity int
}

// OrderConfirmation is what the vendor responds with after accepting (or
// rejecting) an order submission.
type OrderConfirmation struct {
	ExternalRef string
	Confirmed   bool
}
