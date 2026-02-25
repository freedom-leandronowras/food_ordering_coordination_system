package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidOrder        = errors.New("invalid order")
	ErrInsufficientCredits = errors.New("insufficient credits")
	ErrInvalidAmount       = errors.New("invalid amount")
	ErrCreditsExceedCap    = errors.New("credits would exceed cap")
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
	Order            FoodOrder
	RemainingCredits float64
}

type CreditRepository interface {
	Get(memberID uuid.UUID) (float64, bool, error)
	Set(memberID uuid.UUID, amount float64) error
}

type OrderEventRepository interface {
	Save(order FoodOrder) error
	Append(event Event) error
}

type OrderReader interface {
	OrdersByMember(memberID uuid.UUID) ([]FoodOrder, error)
}

type PlaceOrderUseCase struct {
	creditRepo     CreditRepository
	orderEventRepo OrderEventRepository
	now            func() time.Time
}

func NewPlaceOrderUseCase(creditRepo CreditRepository, orderEventRepo OrderEventRepository) *PlaceOrderUseCase {
	return &PlaceOrderUseCase{
		creditRepo:     creditRepo,
		orderEventRepo: orderEventRepo,
		now:            time.Now,
	}
}

func (u *PlaceOrderUseCase) Execute(input PlaceOrderInput) (PlaceOrderResult, error) {
	if input.MemberID == uuid.Nil || len(input.Items) == 0 {
		return PlaceOrderResult{}, ErrInvalidOrder
	}

	total := 0.0
	orderItems := make([]FoodItem, 0, len(input.Items))
	for _, item := range input.Items {
		if item.ID == uuid.Nil || item.Quantity <= 0 || item.Price < 0 {
			return PlaceOrderResult{}, ErrInvalidOrder
		}
		total += float64(item.Quantity) * item.Price
		orderItems = append(orderItems, FoodItem{
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
		return PlaceOrderResult{}, ErrInsufficientCredits
	}

	order := FoodOrder{
		ID:            uuid.New(),
		MemberID:      input.MemberID,
		Items:         orderItems,
		Status:        OrderStatusConfirmed,
		TotalPrice:    total,
		DeliveryNotes: input.DeliveryNotes,
	}

	event := Event{
		ID:          uuid.New(),
		Type:        FoodOrderCreatedEvt,
		AggregateID: order.ID,
		OccurredAt:  u.now().UTC(),
		Payload: FoodOrderPlaced{
			OrderID:       order.ID,
			MemberID:      input.MemberID,
			Items:         orderItems,
			TotalPrice:    total,
			DeliveryNotes: input.DeliveryNotes,
			Status:        OrderStatusConfirmed,
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

type GetCreditsUseCase struct {
	creditRepo CreditRepository
}

func NewGetCreditsUseCase(creditRepo CreditRepository) *GetCreditsUseCase {
	return &GetCreditsUseCase{
		creditRepo: creditRepo,
	}
}

func (u *GetCreditsUseCase) Execute(memberID uuid.UUID) (float64, bool, error) {
	return u.creditRepo.Get(memberID)
}

// ---------------------------------------------------------------------------
// AddCreditsUseCase — managers grant credits to a member
// ---------------------------------------------------------------------------

type AddCreditsUseCase struct {
	creditRepo     CreditRepository
	orderEventRepo OrderEventRepository
	now            func() time.Time
}

func NewAddCreditsUseCase(creditRepo CreditRepository, orderEventRepo OrderEventRepository) *AddCreditsUseCase {
	return &AddCreditsUseCase{
		creditRepo:     creditRepo,
		orderEventRepo: orderEventRepo,
		now:            time.Now,
	}
}

func (u *AddCreditsUseCase) Execute(memberID uuid.UUID, amount float64) (float64, error) {
	if memberID == uuid.Nil || amount <= 0 {
		return 0, ErrInvalidAmount
	}

	current, _, err := u.creditRepo.Get(memberID)
	if err != nil {
		return 0, err
	}

	newBalance := current + amount
	if newBalance > MaxMemberCredits {
		return 0, ErrCreditsExceedCap
	}

	if err := u.creditRepo.Set(memberID, newBalance); err != nil {
		return 0, err
	}

	event := Event{
		ID:          uuid.New(),
		Type:        CreditsGrantedEvt,
		AggregateID: memberID,
		OccurredAt:  u.now().UTC(),
		Payload: CreditsGranted{
			MemberID:   memberID,
			Amount:     amount,
			NewBalance: newBalance,
		},
	}
	if err := u.orderEventRepo.Append(event); err != nil {
		return 0, err
	}

	return newBalance, nil
}

// ---------------------------------------------------------------------------
// GetMemberOrdersUseCase — retrieve a member's order history
// ---------------------------------------------------------------------------

type GetMemberOrdersUseCase struct {
	orderReader OrderReader
}

func NewGetMemberOrdersUseCase(orderReader OrderReader) *GetMemberOrdersUseCase {
	return &GetMemberOrdersUseCase{orderReader: orderReader}
}

func (u *GetMemberOrdersUseCase) Execute(memberID uuid.UUID) ([]FoodOrder, error) {
	return u.orderReader.OrdersByMember(memberID)
}
