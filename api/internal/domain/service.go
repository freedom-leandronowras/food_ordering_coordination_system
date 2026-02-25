package domain

import (
	"github.com/google/uuid"
)

type Service struct {
	placeOrderUseCase *PlaceOrderUseCase
	getCreditsUseCase *GetCreditsUseCase
}

func NewFoodOrderingService(creditRepo CreditRepository, orderEventRepo OrderEventRepository) *Service {
	return &Service{
		placeOrderUseCase: NewPlaceOrderUseCase(creditRepo, orderEventRepo),
		getCreditsUseCase: NewGetCreditsUseCase(creditRepo),
	}
}

func (s *Service) PlaceOrder(input PlaceOrderInput) (PlaceOrderResult, error) {
	return s.placeOrderUseCase.Execute(input)
}

func (s *Service) GetCredits(memberID uuid.UUID) (float64, bool, error) {
	return s.getCreditsUseCase.Execute(memberID)
}
