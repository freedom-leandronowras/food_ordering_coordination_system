package httpapi

import (
	"encoding/json"
	"net/http"
)

type placeOrderRequest struct {
	MemberID      string                  `json:"member_id"`
	Items         []placeOrderRequestItem `json:"items"`
	DeliveryNotes string                  `json:"delivery_notes"`
}

type placeOrderRequestItem struct {
	ID       string  `json:"id"`
	Quantity int     `json:"quantity"`
	Price    float64 `json:"price"`
}

type placeOrderResponse struct {
	OrderID          string  `json:"order_id"`
	Status           string  `json:"status"`
	TotalPrice       float64 `json:"total_price"`
	RemainingCredits float64 `json:"remaining_credits"`
}

type creditsResponse struct {
	MemberID string  `json:"member_id"`
	Credits  float64 `json:"credits"`
}

type errorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func writeError(w http.ResponseWriter, status int, code string, message string) {
	writeJSON(w, status, errorResponse{
		Code:    code,
		Message: message,
	})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
