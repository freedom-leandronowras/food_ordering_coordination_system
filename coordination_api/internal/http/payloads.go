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

type grantCreditsRequest struct {
	Amount float64 `json:"amount"`
}

type grantCreditsResponse struct {
	MemberID   string  `json:"member_id"`
	NewBalance float64 `json:"new_balance"`
}

type orderResponse struct {
	OrderID       string             `json:"order_id"`
	Status        string             `json:"status"`
	TotalPrice    float64            `json:"total_price"`
	DeliveryNotes string             `json:"delivery_notes"`
	Items         []orderItemPayload `json:"items"`
}

type orderItemPayload struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Quantity int     `json:"quantity"`
	Price    float64 `json:"price"`
}

type menuItemResponse struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Available   bool    `json:"available"`
}

type vendorMenuResponse struct {
	ServiceID   string             `json:"service_id"`
	ServiceName string             `json:"service_name"`
	Items       []menuItemResponse `json:"items"`
	Error       string             `json:"error,omitempty"`
}

type vendorResponse struct {
	ServiceID string `json:"service_id"`
}

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	FullName string `json:"full_name"`
	Role     string `json:"role,omitempty"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type authUserResponse struct {
	UserID   string  `json:"user_id"`
	MemberID string  `json:"member_id"`
	Email    string  `json:"email"`
	FullName string  `json:"full_name"`
	Role     string  `json:"role"`
	Credits  float64 `json:"credits"`
}

type authSessionResponse struct {
	Token string           `json:"token"`
	User  authUserResponse `json:"user"`
}

type membersByDomainResponse struct {
	Domain  string             `json:"domain"`
	Members []authUserResponse `json:"members"`
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
