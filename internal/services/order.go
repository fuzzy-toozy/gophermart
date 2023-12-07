package services

import (
	"context"
	"net/http"

	"github.com/fuzzy-toozy/gophermart/internal/database/repo"
	"github.com/fuzzy-toozy/gophermart/internal/errors"
	"github.com/fuzzy-toozy/gophermart/internal/models"
)

type OrderService struct {
	repo repo.OrderServiceRepo
}

func NewOrderService(repo repo.OrderServiceRepo) *OrderService {
	return &OrderService{
		repo: repo,
	}
}

func (s *OrderService) CheckOrderNumber(ctx context.Context, username string, number string) errors.ServiceError {
	if len(number) == 0 || !LuhnCheck(number) {
		return errors.NewServiceError(http.StatusUnprocessableEntity, "wrong or empty order number: '%v'", number)
	}

	order, err := s.repo.GetOrderByNumber(ctx, number)
	if err != nil {
		return errors.NewServiceError(http.StatusInternalServerError,
			"failed to get order with number '%v' from db: %w", number, err)
	}

	if len(order.Number) == 0 {
		return nil
	}

	if username != order.Username {
		return errors.NewServiceError(http.StatusConflict, "order with number '%v' doesn't belong to user", number)
	}

	return errors.NewServiceError(http.StatusOK, "order with number '%v' already exists", number)
}

func (s *OrderService) AddNewOrder(ctx context.Context, username string, number string) errors.ServiceError {
	order := models.NewOrder(username, number)
	err := s.repo.AddNewOrder(ctx, order)

	if err != nil {
		return errors.NewServiceError(http.StatusInternalServerError,
			"failed to add new order with number '%v': %w", number, err)
	}

	return nil
}

func (s *OrderService) GetAllOrders(ctx context.Context, username string) ([]models.Order, errors.ServiceError) {
	orders, err := s.repo.GetAllUserOrders(ctx, username)
	if err != nil {
		return nil, errors.NewServiceError(http.StatusBadRequest, "failed to get all orders: %w", err)
	}

	if len(orders) == 0 {
		return nil, errors.NewServiceError(http.StatusNoContent, "no orders found")
	}

	return orders, nil
}
