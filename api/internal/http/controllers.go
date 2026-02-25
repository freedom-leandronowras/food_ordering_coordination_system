package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"

	"food_ordering_coordination_system/internal/application"
	"food_ordering_coordination_system/internal/domain"
	"food_ordering_coordination_system/internal/usecase/place_order"
	"github.com/google/uuid"
)

type Controller struct {
	service *application.Service
}

func NewFoodOrderingController(service *application.Service) *Controller {
	return &Controller{service: service}
}

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

	items := make([]place_order.PlaceOrderItem, 0, len(req.Items))
	for _, item := range req.Items {
		itemID, parseErr := uuid.Parse(item.ID)
		if parseErr != nil {
			writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "item id must be a valid uuid")
			return
		}
		items = append(items, place_order.PlaceOrderItem{
			ID:       itemID,
			Quantity: item.Quantity,
			Price:    item.Price,
		})
	}

	result, err := c.service.PlaceOrder(place_order.PlaceOrderInput{
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
