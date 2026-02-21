package domain

import (
	"github.com/google/uuid"
)

type ManagerType string

const (
	HiveManager    ManagerType = "HIVE_MANAGER"
	InnovationLead ManagerType = "INNOVATION_LEAD"
)

type FacilityManager struct {
	id          uuid.UUID
	name        string
	managerType ManagerType
	email       string
}

type Member struct {
	id    uuid.UUID
	name  string
	role  string
	email string
}

type Company struct {
	id   uuid.UUID
	name string
}

type FoodOrder struct {
	id            uuid.UUID
	memberId      uuid.UUID
	items         []FoodItem
	status        string
	totalPrice    float32
	deliveryNotes string // example: remove something
}

type FoodItem struct {
	id       uuid.UUID
	name     string
	quantity int
	price    float64
}

type Credits struct {
	id       uuid.UUID
	memberId uuid.UUID
	amount   float32
}
