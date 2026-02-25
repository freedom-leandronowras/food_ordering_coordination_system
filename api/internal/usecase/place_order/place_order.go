package place_order

import (
	"time"

	"food_ordering_coordination_system/internal/domain"
	"github.com/google/uuid"
)

type PlaceOrderItem struct {
	ID       uuid.UUID
	Quantity int
	Price    float64
}

type PlaceOrderInput struct {
	MemberID      uuid.UUID
	Items         []PlaceOrderItem
	DeliveryNotes string
}

type PlaceOrderResult struct {
	Order            domain.FoodOrder
	RemainingCredits float64
}

type PlaceOrderUseCase struct {
	creditRepo     domain.CreditRepository
	orderEventRepo domain.OrderEventRepository
	now            func() time.Time
}

func NewPlaceOrderUseCase(creditRepo domain.CreditRepository, orderEventRepo domain.OrderEventRepository) *PlaceOrderUseCase {
	return &PlaceOrderUseCase{
		creditRepo:     creditRepo,
		orderEventRepo: orderEventRepo,
		now:            time.Now,
	}
}

func (u *PlaceOrderUseCase) Execute(input PlaceOrderInput) (PlaceOrderResult, error) {
	if input.MemberID == uuid.Nil || len(input.Items) == 0 {
		return PlaceOrderResult{}, domain.ErrInvalidOrder
	}

	total := 0.0
	orderItems := make([]domain.FoodItem, 0, len(input.Items))
	for _, item := range input.Items {
		if item.ID == uuid.Nil || item.Quantity <= 0 || item.Price < 0 {
			return PlaceOrderResult{}, domain.ErrInvalidOrder
		}
		total += float64(item.Quantity) * item.Price
		orderItems = append(orderItems, domain.FoodItem{
			ID:       item.ID,
			Quantity: item.Quantity,
			Price:    item.Price,
		})
	}

	currentCredits, ok, err := u.creditRepo.Get(input.MemberID)
	if err != nil {
		return PlaceOrderResult{}, err
	}
	if !ok || currentCredits < total {
		return PlaceOrderResult{}, domain.ErrInsufficientCredits
	}

	order := domain.FoodOrder{
		ID:            uuid.New(),
		MemberID:      input.MemberID,
		Items:         orderItems,
		Status:        domain.OrderStatusConfirmed,
		TotalPrice:    total,
		DeliveryNotes: input.DeliveryNotes,
	}

	event := domain.Event{
		ID:          uuid.New(),
		Type:        domain.FoodOrderCreatedEvt,
		AggregateID: order.ID,
		OccurredAt:  u.now().UTC(),
		Payload: domain.FoodOrderPlaced{
			OrderID:       order.ID,
			MemberID:      input.MemberID,
			Items:         orderItems,
			TotalPrice:    total,
			DeliveryNotes: input.DeliveryNotes,
			Status:        domain.OrderStatusConfirmed,
		},
	}

	if err := u.creditRepo.Set(input.MemberID, currentCredits-total); err != nil {
		return PlaceOrderResult{}, err
	}
	if err := u.orderEventRepo.Save(order); err != nil {
		return PlaceOrderResult{}, err
	}
	if err := u.orderEventRepo.Append(event); err != nil {
		return PlaceOrderResult{}, err
	}

	return PlaceOrderResult{
		Order:            order,
		RemainingCredits: currentCredits - total,
	}, nil
}
