package get_credits

import (
	"food_ordering_coordination_system/internal/domain"
	"github.com/google/uuid"
)

type GetCreditsUseCase struct {
	creditRepo domain.CreditRepository
}

func NewGetCreditsUseCase(creditRepo domain.CreditRepository) *GetCreditsUseCase {
	return &GetCreditsUseCase{
		creditRepo: creditRepo,
	}
}

func (u *GetCreditsUseCase) Execute(memberID uuid.UUID) (float64, bool, error) {
	return u.creditRepo.Get(memberID)
}
