package domain_test

import (
	"food_ordering_coordination_system/internal/domain"
	"github.com/google/uuid"
	"testing"
)

func TestPlaceOrderUseCase_Execute_Validation(t *testing.T) {
	tests := []struct {
		name    string
		input   domain.PlaceOrderInput
		wantErr error
	}{
		{
			name: "invalid member ID",
			input: domain.PlaceOrderInput{
				MemberID: uuid.Nil,
				Items: []domain.PlaceOrderItem{
					{ID: uuid.New(), Quantity: 1, Price: 10.0},
				},
			},
			wantErr: domain.ErrInvalidOrder,
		},
		{
			name: "empty items",
			input: domain.PlaceOrderInput{
				MemberID: uuid.New(),
				Items:    []domain.PlaceOrderItem{},
			},
			wantErr: domain.ErrInvalidOrder,
		},
		{
			name: "invalid item ID",
			input: domain.PlaceOrderInput{
				MemberID: uuid.New(),
				Items: []domain.PlaceOrderItem{
					{ID: uuid.Nil, Quantity: 1, Price: 10.0},
				},
			},
			wantErr: domain.ErrInvalidOrder,
		},
		{
			name: "invalid item quantity (zero)",
			input: domain.PlaceOrderInput{
				MemberID: uuid.New(),
				Items: []domain.PlaceOrderItem{
					{ID: uuid.New(), Quantity: 0, Price: 10.0},
				},
			},
			wantErr: domain.ErrInvalidOrder,
		},
		{
			name: "invalid item quantity (negative)",
			input: domain.PlaceOrderInput{
				MemberID: uuid.New(),
				Items: []domain.PlaceOrderItem{
					{ID: uuid.New(), Quantity: -1, Price: 10.0},
				},
			},
			wantErr: domain.ErrInvalidOrder,
		},
		{
			name: "invalid item price (negative)",
			input: domain.PlaceOrderInput{
				MemberID: uuid.New(),
				Items: []domain.PlaceOrderItem{
					{ID: uuid.New(), Quantity: 1, Price: -10.0},
				},
			},
			wantErr: domain.ErrInvalidOrder,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Since we are testing validation logic that happens before any repository calls,
			// we can pass nil for the repositories.
			uc := domain.NewPlaceOrderUseCase(nil, nil)
			_, err := uc.Execute(tt.input)

			if err != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
