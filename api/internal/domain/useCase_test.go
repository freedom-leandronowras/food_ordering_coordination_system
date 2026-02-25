package domain

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockCreditRepository is a mock implementation of CreditRepository
type MockCreditRepository struct {
	mock.Mock
}

func (m *MockCreditRepository) Get(memberID uuid.UUID) (float64, bool, error) {
	args := m.Called(memberID)
	return args.Get(0).(float64), args.Bool(1), args.Error(2)
}

func (m *MockCreditRepository) Set(memberID uuid.UUID, amount float64) error {
	args := m.Called(memberID, amount)
	return args.Error(0)
}

// MockOrderEventRepository is a mock implementation of OrderEventRepository
type MockOrderEventRepository struct {
	mock.Mock
}

func (m *MockOrderEventRepository) Save(order FoodOrder) error {
	args := m.Called(order)
	return args.Error(0)
}

func (m *MockOrderEventRepository) Append(event Event) error {
	args := m.Called(event)
	return args.Error(0)
}

func TestPlaceOrderUseCase_Execute(t *testing.T) {
	memberID := uuid.New()
	itemID := uuid.New()
	mockErr := errors.New("db error")

	validItem := PlaceOrderItem{
		ID:       itemID,
		Quantity: 2,
		Price:    10.0,
	}
	// Total should be 20.0

	tests := []struct {
		name           string
		input          PlaceOrderInput
		setupMocks     func(*MockCreditRepository, *MockOrderEventRepository)
		expectedError  error
		validateResult func(*testing.T, PlaceOrderResult)
	}{
		{
			name: "Success",
			input: PlaceOrderInput{
				MemberID:      memberID,
				Items:         []PlaceOrderItem{validItem},
				DeliveryNotes: "Leave at door",
			},
			setupMocks: func(cr *MockCreditRepository, er *MockOrderEventRepository) {
				cr.On("Get", memberID).Return(100.0, true, nil)
				cr.On("Set", memberID, 80.0).Return(nil)
				er.On("Save", mock.MatchedBy(func(order FoodOrder) bool {
					return order.MemberID == memberID &&
						order.TotalPrice == 20.0 &&
						order.Status == OrderStatusConfirmed &&
						len(order.Items) == 1 &&
						order.Items[0].ID == itemID &&
						order.Items[0].Quantity == 2 &&
						order.Items[0].Price == 10.0
				})).Return(nil)
				er.On("Append", mock.MatchedBy(func(event Event) bool {
					payload, ok := event.Payload.(FoodOrderPlaced)
					if !ok {
						return false
					}
					return event.Type == FoodOrderCreatedEvt &&
						payload.MemberID == memberID &&
						payload.TotalPrice == 20.0 &&
						payload.Status == OrderStatusConfirmed &&
						len(payload.Items) == 1 &&
						payload.Items[0].ID == itemID
				})).Return(nil)
			},
			expectedError: nil,
			validateResult: func(t *testing.T, res PlaceOrderResult) {
				assert.NotEqual(t, uuid.Nil, res.Order.ID)
				assert.Equal(t, 80.0, res.RemainingCredits)
				assert.Equal(t, 20.0, res.Order.TotalPrice)
			},
		},
		{
			name: "Insufficient Credits",
			input: PlaceOrderInput{
				MemberID: memberID,
				Items:    []PlaceOrderItem{validItem},
			},
			setupMocks: func(cr *MockCreditRepository, er *MockOrderEventRepository) {
				cr.On("Get", memberID).Return(10.0, true, nil) // Only 10 credits, need 20
			},
			expectedError:  ErrInsufficientCredits,
			validateResult: nil,
		},
		{
			name: "Member Not Found (Get returns !ok)",
			input: PlaceOrderInput{
				MemberID: memberID,
				Items:    []PlaceOrderItem{validItem},
			},
			setupMocks: func(cr *MockCreditRepository, er *MockOrderEventRepository) {
				cr.On("Get", memberID).Return(0.0, false, nil)
			},
			expectedError:  ErrInsufficientCredits,
			validateResult: nil,
		},
		{
			name: "Invalid Input - No Items",
			input: PlaceOrderInput{
				MemberID: memberID,
				Items:    []PlaceOrderItem{},
			},
			setupMocks:     func(cr *MockCreditRepository, er *MockOrderEventRepository) {},
			expectedError:  ErrInvalidOrder,
			validateResult: nil,
		},
		{
			name: "Invalid Input - Invalid Item Quantity",
			input: PlaceOrderInput{
				MemberID: memberID,
				Items:    []PlaceOrderItem{{ID: itemID, Quantity: 0, Price: 10.0}},
			},
			setupMocks:     func(cr *MockCreditRepository, er *MockOrderEventRepository) {},
			expectedError:  ErrInvalidOrder,
			validateResult: nil,
		},
		{
			name: "Invalid Input - Invalid Item Price",
			input: PlaceOrderInput{
				MemberID: memberID,
				Items:    []PlaceOrderItem{{ID: itemID, Quantity: 1, Price: -10.0}},
			},
			setupMocks:     func(cr *MockCreditRepository, er *MockOrderEventRepository) {},
			expectedError:  ErrInvalidOrder,
			validateResult: nil,
		},
		{
			name: "Get Error",
			input: PlaceOrderInput{
				MemberID: memberID,
				Items:    []PlaceOrderItem{validItem},
			},
			setupMocks: func(cr *MockCreditRepository, er *MockOrderEventRepository) {
				cr.On("Get", memberID).Return(0.0, false, mockErr)
			},
			expectedError:  mockErr,
			validateResult: nil,
		},
		{
			name: "Set Error",
			input: PlaceOrderInput{
				MemberID: memberID,
				Items:    []PlaceOrderItem{validItem},
			},
			setupMocks: func(cr *MockCreditRepository, er *MockOrderEventRepository) {
				cr.On("Get", memberID).Return(100.0, true, nil)
				cr.On("Set", memberID, 80.0).Return(mockErr)
			},
			expectedError:  mockErr,
			validateResult: nil,
		},
		{
			name: "Save Error",
			input: PlaceOrderInput{
				MemberID: memberID,
				Items:    []PlaceOrderItem{validItem},
			},
			setupMocks: func(cr *MockCreditRepository, er *MockOrderEventRepository) {
				cr.On("Get", memberID).Return(100.0, true, nil)
				cr.On("Set", memberID, 80.0).Return(nil)
				er.On("Save", mock.Anything).Return(mockErr)
			},
			expectedError:  mockErr,
			validateResult: nil,
		},
		{
			name: "Append Error",
			input: PlaceOrderInput{
				MemberID: memberID,
				Items:    []PlaceOrderItem{validItem},
			},
			setupMocks: func(cr *MockCreditRepository, er *MockOrderEventRepository) {
				cr.On("Get", memberID).Return(100.0, true, nil)
				cr.On("Set", memberID, 80.0).Return(nil)
				er.On("Save", mock.Anything).Return(nil)
				er.On("Append", mock.Anything).Return(mockErr)
			},
			expectedError:  mockErr,
			validateResult: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCredit := new(MockCreditRepository)
			mockOrder := new(MockOrderEventRepository)
			tt.setupMocks(mockCredit, mockOrder)

			uc := NewPlaceOrderUseCase(mockCredit, mockOrder)
			// Ensure deterministic time by mocking now
			fixedTime := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
			uc.now = func() time.Time { return fixedTime }

			res, err := uc.Execute(tt.input)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
			} else {
				assert.NoError(t, err)
				if tt.validateResult != nil {
					tt.validateResult(t, res)
				}
			}

			mockCredit.AssertExpectations(t)
			mockOrder.AssertExpectations(t)
		})
	}
}
