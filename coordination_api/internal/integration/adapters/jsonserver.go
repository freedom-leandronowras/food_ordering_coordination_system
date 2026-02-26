package adapters

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"food_ordering_coordination_system/internal/domain"
	"food_ordering_coordination_system/internal/integration"

	"github.com/google/uuid"
)

// JSONServerAdapter is an HTTP adapter that talks to a json-server-style REST
// API.  Each instance points at one vendor's base URL and translates HTTP
// responses into domain types.
//
// Expected endpoints:
//
//	GET  /menu   → JSON array of menu items
//	POST /orders → accepts an order, returns a confirmation
type JSONServerAdapter struct {
	id      string
	name    string
	baseURL string
	client  *http.Client
}

// NewJSONServerAdapter creates an adapter that will issue HTTP requests to the
// given base URL.  A zero-value http.Client with a 10s timeout is used when
// client is nil.
func NewJSONServerAdapter(id, name, baseURL string, client *http.Client) *JSONServerAdapter {
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}
	return &JSONServerAdapter{id: id, name: name, baseURL: baseURL, client: client}
}

func (a *JSONServerAdapter) ServiceID() string   { return a.id }
func (a *JSONServerAdapter) ServiceName() string { return a.name }

// FetchMenu calls GET /menu on the vendor's JSON server and converts the
// response into a slice of domain.MenuItem.
func (a *JSONServerAdapter) FetchMenu(ctx context.Context) ([]domain.MenuItem, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, a.baseURL+"/menu", nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("GET /menu: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GET /menu: unexpected status %d", resp.StatusCode)
	}

	var raw []menuItemJSON
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("decode menu: %w", err)
	}

	items := make([]domain.MenuItem, 0, len(raw))
	for _, r := range raw {
		id, parseErr := uuid.Parse(r.ID)
		if parseErr != nil {
			return nil, fmt.Errorf("parse item id %q: %w", r.ID, parseErr)
		}
		items = append(items, domain.MenuItem{
			ID:          id,
			Name:        r.Name,
			Description: r.Description,
			Price:       r.Price,
			Available:   r.Available,
		})
	}
	return items, nil
}

// SubmitOrder calls POST /orders on the vendor's JSON server.
func (a *JSONServerAdapter) SubmitOrder(ctx context.Context, sub integration.OrderSubmission) (integration.OrderConfirmation, error) {
	body := orderRequestJSON{
		OrderID: sub.OrderID.String(),
		Notes:   sub.Notes,
	}
	for _, item := range sub.Items {
		body.Items = append(body.Items, orderItemJSON{
			ItemID:   item.ItemID.String(),
			Quantity: item.Quantity,
		})
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return integration.OrderConfirmation{}, fmt.Errorf("marshal order: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.baseURL+"/orders", bytes.NewReader(payload))
	if err != nil {
		return integration.OrderConfirmation{}, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		return integration.OrderConfirmation{}, fmt.Errorf("POST /orders: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return integration.OrderConfirmation{}, fmt.Errorf("POST /orders: unexpected status %d", resp.StatusCode)
	}

	var confirmation orderConfirmationJSON
	if err := json.NewDecoder(resp.Body).Decode(&confirmation); err != nil {
		return integration.OrderConfirmation{}, fmt.Errorf("decode confirmation: %w", err)
	}

	return integration.OrderConfirmation{
		ExternalRef: confirmation.ExternalRef,
		Confirmed:   confirmation.Confirmed,
	}, nil
}

// ---------- JSON wire types ----------

type menuItemJSON struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Available   bool    `json:"available"`
}

type orderRequestJSON struct {
	OrderID string          `json:"order_id"`
	Items   []orderItemJSON `json:"items"`
	Notes   string          `json:"notes"`
}

type orderItemJSON struct {
	ItemID   string `json:"item_id"`
	Quantity int    `json:"quantity"`
}

type orderConfirmationJSON struct {
	ExternalRef string `json:"external_ref"`
	Confirmed   bool   `json:"confirmed"`
}

