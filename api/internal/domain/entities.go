package domain

import (
	"github.com/google/uuid"
)

type ManagerType string

const (
	HiveManager ManagerType = "HIVE_MANAGER"
)

type FacilityManager struct {
	ID          uuid.UUID
	Name        string
	ManagerType ManagerType
	Email       string
}

type Member struct {
	ID      uuid.UUID
	Name    string
	Role    string
	Email   string
	Credits Credit
}

type Company struct {
	ID   uuid.UUID
	Name string
}

type Vendor struct {
	ID          uuid.UUID
	Name        string
	Description string
	Active      bool
}

type Menu struct {
	ID       uuid.UUID
	VendorID uuid.UUID
	Name     string
	Items    []MenuItem
	Active   bool
}

type MenuItem struct {
	ID          uuid.UUID
	MenuID      uuid.UUID
	Name        string
	Description string
	Price       float64
	Available   bool
}

type OrderStatus string

const (
	OrderStatusConfirmed OrderStatus = "CONFIRMED"
)

type FoodOrder struct {
	ID            uuid.UUID
	MemberID      uuid.UUID
	Items         []FoodItem
	Status        OrderStatus
	TotalPrice    float64
	DeliveryNotes string
}

type FoodItem struct {
	ID       uuid.UUID
	Name     string
	Quantity int
	Price    float64
}

type Credit struct {
	ID       uuid.UUID
	MemberID uuid.UUID
	Amount   float64
}

const MaxMemberCredits float64 = 1000
