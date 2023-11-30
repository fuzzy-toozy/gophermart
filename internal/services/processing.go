package services

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/fuzzy-toozy/gophermart/internal/database/repo"
	serviceErrs "github.com/fuzzy-toozy/gophermart/internal/errors"
	"github.com/fuzzy-toozy/gophermart/internal/models"
	"go.uber.org/zap"
)

type ProcessingService struct {
	accural *AccrualService
	repo    repo.ProcessServiceRepo
	logger  *zap.SugaredLogger
}

func NewProcessingService(repo repo.ProcessServiceRepo, accural *AccrualService, logger *zap.SugaredLogger) *ProcessingService {
	return &ProcessingService{
		repo:    repo,
		accural: accural,
		logger:  logger,
	}
}

func (s *ProcessingService) Withdraw(ctx context.Context, wd *models.Withdraw, username string) serviceErrs.ServiceError {
	if len(wd.Order) == 0 || !LuhnCheck(wd.Order) {
		return serviceErrs.NewServiceError(http.StatusUnprocessableEntity,
			"invalid order number: '%v'", wd.Order)
	}

	err := s.repo.WithdrawBalance(ctx, wd, username)

	if err != nil && !errors.Is(err, repo.ErrWithdrawUnavailable) {
		return serviceErrs.NewServiceError(http.StatusInternalServerError,
			"failed to withdraw balance for order '%v': %w", wd.Order, err)
	}

	if errors.Is(err, repo.ErrWithdrawUnavailable) {
		return serviceErrs.NewServiceError(http.StatusPaymentRequired, "not enough funds")
	}

	return nil
}

func (s *ProcessingService) processOrder(ctx context.Context, order *models.Order, accural float64) error {
	err := s.repo.ProcessOrder(ctx, order, accural)
	if err != nil {
		return err
	}

	return nil
}

func (s *ProcessingService) processAccural(ctx context.Context, order *models.Order) error {
	orderInfo, err := s.accural.GetOrderInfo(order.Number)
	if err != nil {
		return fmt.Errorf("failed to get order '%v' info for user '%v' from accural: %w",
			order.Number, order.Username, err)
	}

	if orderInfo.Order != order.Number {
		return fmt.Errorf("invalid order number from accural: expected '%v', got '%v'", order.Number, orderInfo.Order)
	}

	switch orderInfo.Status {
	case models.OrderINVALID:
		order.Status = models.OrderINVALID
		if err = s.repo.UpdateOrderStatus(ctx, order); err != nil {
			return fmt.Errorf("failed to update order '%v' status to '%v' for user '%v': %w",
				order.Number, order.Status, order.Username, err)
		}
	case models.OrderPROCESSED:
		order.Status = models.OrderPROCESSED
		err := s.repo.ProcessOrder(ctx, order, orderInfo.Accrual)
		if err != nil {
			return fmt.Errorf("falied to finalize processed order '%v' for user '%v': %w", order.Number, order.Username, err)
		}
	}

	return nil
}

func (s *ProcessingService) ProcessOrders(ctx context.Context) serviceErrs.ServiceError {
	orders, err := s.repo.GetAllUnprocessedOrders(ctx)
	if err != nil {
		return serviceErrs.NewServiceError(http.StatusInternalServerError, "%w", err)
	}

	for _, order := range orders {
		s.logger.Debugf("Received uprocessed order '%v' from user '%v' with status '%v'",
			order.Number, order.Username, order.Status)
		switch order.Status {
		case models.OrderNEW:
			order.Status = models.OrderPROCESSING
			if err = s.repo.UpdateOrderStatus(ctx, &order); err != nil {
				s.logger.Errorf("Failed to update order '%v' status to '%v' for user '%v': %w",
					order.Number, order.Status, order.Username, err)

				continue
			}
		case models.OrderPROCESSING:
			if err := s.processAccural(ctx, &order); err != nil {
				s.logger.Errorf("Failed to process accural order: %v", err)

				continue
			}
		}
	}

	return nil
}
