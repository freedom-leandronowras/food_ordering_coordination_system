package domain

import "github.com/google/uuid"

type CreditRepository interface {
	Get(memberID uuid.UUID) (float64, bool, error)
	Set(memberID uuid.UUID, amount float64) error
}

type OrderEventRepository interface {
	Save(order FoodOrder) error
	Append(event Event) error
}
