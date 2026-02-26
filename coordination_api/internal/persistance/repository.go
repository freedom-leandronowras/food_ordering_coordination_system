package persistence

import (
	"food_ordering_coordination_system/internal/domain"
	"github.com/google/uuid"
	"time"
)

type CreditRepository interface {
	Get(memberID uuid.UUID) (float64, bool, error)
	Set(memberID uuid.UUID, amount float64) error
}

type OrderEventRepository interface {
	Save(order domain.FoodOrder) error
	Append(event domain.Event) error
}

type User struct {
	UserID       uuid.UUID
	MemberID     uuid.UUID
	Email        string
	FullName     string
	Role         string
	PasswordHash string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type UserRepository interface {
	CreateUser(user User) error
	FindUserByEmail(email string) (User, bool, error)
	FindUserByMemberID(memberID uuid.UUID) (User, bool, error)
	ListUsersByEmailDomain(domain string) ([]User, error)
}
