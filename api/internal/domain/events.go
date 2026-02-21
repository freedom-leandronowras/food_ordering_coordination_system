// should be for requests/orders
// receive a entity to instate the event

// - [ ] dto json for response

package main

import (
	"time"

	"github.com/google/uuid"
)

type EventType string

const (
	FoodOrderCreatedEvt   EventType = "food-order.created.v1"
	FoodOrderSubmittedEvt EventType = "food-order.submitted.v1"
)

type Event struct {
	ID            uuid.UUID
	Type          EventType
	AggregateID   uuid.UUID
	CorrelationID uuid.UUID
	CausationID   uuid.UUID
	OccurredAt    time.Time
	Payload       any
}

type FoodOrderPlaced struct {
	OrderID       uuid.UUID
	MemberID      uuid.UUID
	Items         []FoodItem
	TotalPrice    float32
	DeliveryNotes string
	Status        string
}
