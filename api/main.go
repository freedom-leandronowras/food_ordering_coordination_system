package main

/*
1. fazer events
2. go channels? testar?
3. graphql
4. go mux endpoint? (by hand)
*/

/*
## flow
1. authenticate
2. use the returned data to instantiate structs
3. struct instantiate events
4. dto for interface

obs: i need to store the event and its status
*/

import (
	"fmt"

	"github.com/google/uuid"
)

func main() {
	fmt.Println("oi")
}

type FoodItemDTO struct {
	ID       uuid.UUID `json:"id"`
	Quantity int       `json:"quantity"`
	Price    float64   `json:"price"`
}

type FoodOrderPlacedDTO struct {
	Items         []FoodItemDTO `json:"items"`
	DeliveryNotes string        `json:"delivery_notes"`
}
