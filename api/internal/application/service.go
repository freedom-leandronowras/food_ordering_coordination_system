package application

import (
	"food_ordering_coordination_system/internal/domain"
	"food_ordering_coordination_system/internal/usecase/get_credits"
	"food_ordering_coordination_system/internal/usecase/place_order"
	"github.com/google/uuid"
)

type Service struct {
	placeOrderUseCase *place_order.PlaceOrderUseCase
	getCreditsUseCase *get_credits.GetCreditsUseCase
}

func NewFoodOrderingService(creditRepo domain.CreditRepository, orderEventRepo domain.OrderEventRepository) *Service {
	return &Service{
		placeOrderUseCase: place_order.NewPlaceOrderUseCase(creditRepo, orderEventRepo),
		getCreditsUseCase: get_credits.NewGetCreditsUseCase(creditRepo),
	}
}

func (s *Service) PlaceOrder(input place_order.PlaceOrderInput) (place_order.PlaceOrderResult, error) {
	return s.placeOrderUseCase.Execute(input)
}

func (s *Service) GetCredits(memberID uuid.UUID) (float64, bool, error) {
	return s.getCreditsUseCase.Execute(memberID)
}
