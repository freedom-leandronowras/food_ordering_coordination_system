package domain

import (
	"github.com/google/uuid"
)

type Service struct {
	placeOrderUseCase      *PlaceOrderUseCase
	getCreditsUseCase      *GetCreditsUseCase
	addCreditsUseCase      *AddCreditsUseCase
	getMemberOrdersUseCase *GetMemberOrdersUseCase
}

func NewFoodOrderingService(
	creditRepo CreditRepository,
	orderEventRepo OrderEventRepository,
	orderReader OrderReader,
) *Service {
	return &Service{
		placeOrderUseCase:      NewPlaceOrderUseCase(creditRepo, orderEventRepo),
		getCreditsUseCase:      NewGetCreditsUseCase(creditRepo),
		addCreditsUseCase:      NewAddCreditsUseCase(creditRepo, orderEventRepo),
		getMemberOrdersUseCase: NewGetMemberOrdersUseCase(orderReader),
	}
}

func (s *Service) PlaceOrder(input PlaceOrderInput) (PlaceOrderResult, error) {
	return s.placeOrderUseCase.Execute(input)
}

func (s *Service) GetCredits(memberID uuid.UUID) (float64, bool, error) {
	return s.getCreditsUseCase.Execute(memberID)
}

func (s *Service) AddCredits(memberID uuid.UUID, amount float64) (float64, error) {
	return s.addCreditsUseCase.Execute(memberID, amount)
}

func (s *Service) GetMemberOrders(memberID uuid.UUID) ([]FoodOrder, error) {
	return s.getMemberOrdersUseCase.Execute(memberID)
}
