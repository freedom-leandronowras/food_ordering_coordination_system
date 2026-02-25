package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"

	"handler/internal/domain"
	"handler/internal/integration"

	"github.com/google/uuid"
)

type Controller struct {
	service    *domain.Service
	aggregator *integration.Aggregator
}

func NewFoodOrderingController(service *domain.Service, aggregator *integration.Aggregator) *Controller {
	return &Controller{service: service, aggregator: aggregator}
}

// ---------------------------------------------------------------------------
// POST /api/orders
// ---------------------------------------------------------------------------

func (c *Controller) PlaceOrder(w http.ResponseWriter, r *http.Request, auth AuthClaims) {
	var req placeOrderRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "request body is invalid")
		return
	}

	memberID, err := uuid.Parse(req.MemberID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "member_id must be a valid uuid")
		return
	}

	if !canAccessMemberData(auth, memberID) {
		writeError(w, http.StatusForbidden, "FORBIDDEN", "members can only access their own data")
		return
	}

	items := make([]domain.PlaceOrderItem, 0, len(req.Items))
	for _, item := range req.Items {
		itemID, parseErr := uuid.Parse(item.ID)
		if parseErr != nil {
			writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "item id must be a valid uuid")
			return
		}
		items = append(items, domain.PlaceOrderItem{
			ID:       itemID,
			Quantity: item.Quantity,
			Price:    item.Price,
		})
	}

	result, err := c.service.PlaceOrder(domain.PlaceOrderInput{
		MemberID:      memberID,
		Items:         items,
		DeliveryNotes: req.DeliveryNotes,
	})
	if errors.Is(err, domain.ErrInvalidOrder) {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "order payload is invalid")
		return
	}
	if errors.Is(err, domain.ErrInsufficientCredits) {
		writeError(w, http.StatusUnprocessableEntity, "INSUFFICIENT_CREDITS", "member does not have enough credits")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "unexpected error")
		return
	}

	writeJSON(w, http.StatusCreated, placeOrderResponse{
		OrderID:          result.Order.ID.String(),
		Status:           string(result.Order.Status),
		TotalPrice:       result.Order.TotalPrice,
		RemainingCredits: result.RemainingCredits,
	})
}

// ---------------------------------------------------------------------------
// GET /api/members/{memberId}/credits
// ---------------------------------------------------------------------------

func (c *Controller) GetCredits(w http.ResponseWriter, r *http.Request, auth AuthClaims) {
	memberID, err := uuid.Parse(r.PathValue("memberId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "member id must be a valid uuid")
		return
	}

	if !canAccessMemberData(auth, memberID) {
		writeError(w, http.StatusForbidden, "FORBIDDEN", "members can only access their own data")
		return
	}

	credits, hasAccount, err := c.service.GetCredits(memberID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "unexpected error")
		return
	}
	if !hasAccount {
		writeError(w, http.StatusNotFound, "MEMBER_NOT_FOUND", "member has no credit account")
		return
	}

	writeJSON(w, http.StatusOK, creditsResponse{
		MemberID: memberID.String(),
		Credits:  credits,
	})
}

// ---------------------------------------------------------------------------
// POST /api/members/{memberId}/credits — grant credits (managers only)
// ---------------------------------------------------------------------------

func (c *Controller) GrantCredits(w http.ResponseWriter, r *http.Request, auth AuthClaims) {
	memberID, err := uuid.Parse(r.PathValue("memberId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "member id must be a valid uuid")
		return
	}

	var req grantCreditsRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "request body is invalid")
		return
	}

	newBalance, err := c.service.AddCredits(memberID, req.Amount)
	if errors.Is(err, domain.ErrInvalidAmount) {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "amount must be positive")
		return
	}
	if errors.Is(err, domain.ErrCreditsExceedCap) {
		writeError(w, http.StatusUnprocessableEntity, "CREDITS_EXCEED_CAP", "credits would exceed the maximum of 1000")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "unexpected error")
		return
	}

	writeJSON(w, http.StatusOK, grantCreditsResponse{
		MemberID:   memberID.String(),
		NewBalance: newBalance,
	})
}

// ---------------------------------------------------------------------------
// GET /api/members/{memberId}/orders — member order history
// ---------------------------------------------------------------------------

func (c *Controller) GetMemberOrders(w http.ResponseWriter, r *http.Request, auth AuthClaims) {
	memberID, err := uuid.Parse(r.PathValue("memberId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "member id must be a valid uuid")
		return
	}

	if !canAccessMemberData(auth, memberID) {
		writeError(w, http.StatusForbidden, "FORBIDDEN", "members can only access their own data")
		return
	}

	orders, err := c.service.GetMemberOrders(memberID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "unexpected error")
		return
	}

	payload := make([]orderResponse, 0, len(orders))
	for _, o := range orders {
		items := make([]orderItemPayload, 0, len(o.Items))
		for _, it := range o.Items {
			items = append(items, orderItemPayload{
				ID:       it.ID.String(),
				Name:     it.Name,
				Quantity: it.Quantity,
				Price:    it.Price,
			})
		}
		payload = append(payload, orderResponse{
			OrderID:       o.ID.String(),
			Status:        string(o.Status),
			TotalPrice:    o.TotalPrice,
			DeliveryNotes: o.DeliveryNotes,
			Items:         items,
		})
	}

	writeJSON(w, http.StatusOK, payload)
}

// ---------------------------------------------------------------------------
// GET /api/menus — fan-out to all vendors, fan-in combined menu
// ---------------------------------------------------------------------------

func (c *Controller) GetAllMenus(w http.ResponseWriter, r *http.Request) {
	if c.aggregator == nil {
		writeJSON(w, http.StatusOK, []vendorMenuResponse{})
		return
	}

	results := c.aggregator.FetchAllMenus(r.Context())

	payload := make([]vendorMenuResponse, 0, len(results))
	for _, res := range results {
		vmr := vendorMenuResponse{
			ServiceID:   res.ServiceID,
			ServiceName: res.ServiceName,
		}
		if res.Err != nil {
			vmr.Error = res.Err.Error()
		} else {
			items := make([]menuItemResponse, 0, len(res.Items))
			for _, it := range res.Items {
				items = append(items, menuItemResponse{
					ID:          it.ID.String(),
					Name:        it.Name,
					Description: it.Description,
					Price:       it.Price,
					Available:   it.Available,
				})
			}
			vmr.Items = items
		}
		payload = append(payload, vmr)
	}

	writeJSON(w, http.StatusOK, payload)
}

// ---------------------------------------------------------------------------
// GET /api/vendors — list registered vendors
// ---------------------------------------------------------------------------

func (c *Controller) GetVendors(w http.ResponseWriter, r *http.Request) {
	if c.aggregator == nil {
		writeJSON(w, http.StatusOK, []vendorResponse{})
		return
	}

	ids := c.aggregator.Adapters()
	payload := make([]vendorResponse, 0, len(ids))
	for _, id := range ids {
		payload = append(payload, vendorResponse{ServiceID: id})
	}

	writeJSON(w, http.StatusOK, payload)
}
