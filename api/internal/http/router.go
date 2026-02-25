package httpapi

import (
	"net/http"

	"food_ordering_coordination_system/internal/domain"
	"food_ordering_coordination_system/internal/integration"
)

func NewFoodOrderingRouter(service *domain.Service, aggregator *integration.Aggregator) http.Handler {
	controller := NewFoodOrderingController(service, aggregator)
	authenticator := NewAuthenticator()

	mux := http.NewServeMux()

	// Orders
	mux.HandleFunc("POST /api/orders",
		authenticator.RequireRoles(RoleMember, RoleHiveManager, RoleInnovationLead)(controller.PlaceOrder),
	)
	mux.HandleFunc("GET /api/members/{memberId}/orders",
		authenticator.RequireRoles(RoleMember, RoleHiveManager, RoleInnovationLead)(controller.GetMemberOrders),
	)

	// Credits
	mux.HandleFunc("GET /api/members/{memberId}/credits",
		authenticator.RequireRoles(RoleMember, RoleHiveManager, RoleInnovationLead)(controller.GetCredits),
	)
	mux.HandleFunc("POST /api/members/{memberId}/credits",
		authenticator.RequireRoles(RoleHiveManager, RoleInnovationLead)(controller.GrantCredits),
	)

	// Menus (fan-out / fan-in from external vendors)
	mux.HandleFunc("GET /api/menus", controller.GetAllMenus)

	// Vendors
	mux.HandleFunc("GET /api/vendors", controller.GetVendors)

	return mux
}
