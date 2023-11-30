package services

import (
	"context"
	"net/http"

	"github.com/fuzzy-toozy/gophermart/internal/database/repo"
	serviceErrs "github.com/fuzzy-toozy/gophermart/internal/errors"
	"github.com/fuzzy-toozy/gophermart/internal/models"
)

type BalanceService struct {
	repo repo.BalanceServiceRepo
}

func NewBalanceService(repo repo.BalanceServiceRepo) *BalanceService {
	return &BalanceService{
		repo: repo,
	}
}

func (s *BalanceService) GetUserBalance(ctx context.Context, username string) (*models.Balance, serviceErrs.ServiceError) {
	balance, err := s.repo.GetBanaceData(ctx, username)

	if err != nil {
		return nil, serviceErrs.NewServiceError(http.StatusInternalServerError,
			"failed to get balance from db: %w", err)
	}

	return balance, nil
}

func (s *BalanceService) GetAllUserWithdrawals(ctx context.Context, username string) ([]models.Withdrawals, serviceErrs.ServiceError) {
	result, err := s.repo.GetWithdrawals(ctx, username)
	if err != nil {
		return nil, serviceErrs.NewServiceError(http.StatusInternalServerError,
			"failed to get withdrawals data: %w", err)
	}

	if len(result) == 0 {
		return nil, serviceErrs.NewServiceError(http.StatusNoContent, "no withdrawals found")
	}

	return result, nil
}
