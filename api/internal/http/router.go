package httpapi

import (
	"net/http"

	"food_ordering_coordination_system/internal/domain"
)

func NewFoodOrderingRouter(service *domain.Service) http.Handler {
	controller := NewFoodOrderingController(service)
	authenticator := NewAuthenticator()

	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/orders",
		authenticator.RequireRoles(RoleMember, RoleHiveManager, RoleInnovationLead)(controller.PlaceOrder),
	)

	mux.HandleFunc("GET /api/members/{memberId}/credits",
		authenticator.RequireRoles(RoleMember, RoleHiveManager, RoleInnovationLead)(controller.GetCredits),
	)

	return mux
}
