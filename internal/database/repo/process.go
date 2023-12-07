package repo

import (
	"context"
	"fmt"

	"github.com/fuzzy-toozy/gophermart/internal/database"
	"github.com/fuzzy-toozy/gophermart/internal/models"
)

type ProcessServiceRepo interface {
	ProcessOrder(ctx context.Context, order *models.Order, accural float64) error
	WithdrawBalance(ctx context.Context, wd *models.Withdraw, username string) error
	GetAllUnprocessedOrders(ctx context.Context) ([]models.Order, error)
	UpdateOrderStatus(ctx context.Context, order *models.Order) error
}

type processRepo struct {
	balanceRepo BalanceServiceRepo
	ordersRepo  OrderServiceRepo
	storage     *database.ServiceStorage
}

func (r *processRepo) ProcessOrder(ctx context.Context, order *models.Order, accural float64) error {
	callback := func() error {
		err := r.ordersRepo.UpdateStatus(ctx, order)
		if err != nil {
			return fmt.Errorf("failed to update order '%v' status: %w", order.Number, err)
		}

		err = r.ordersRepo.UpdateAccural(ctx, order, accural)
		if err != nil {
			return fmt.Errorf("failed to update order '%v' accural: %w", order.Number, err)
		}

		err = r.balanceRepo.AddIncomeRecord(ctx, order.Username, order.Number, accural)
		if err != nil {
			return fmt.Errorf("failed to add new balance record: %w", err)
		}

		return nil
	}

	return r.storage.RunInTransaction(callback)
}

func (r *processRepo) WithdrawBalance(ctx context.Context, wd *models.Withdraw, username string) error {
	callback := func() error {
		balance, err := r.balanceRepo.GetBanaceData(ctx, username)
		if err != nil {
			return fmt.Errorf("failed to get balance data: %w", err)
		}

		if wd.Sum > balance.Current {
			return ErrWithdrawUnavailable
		}

		err = r.balanceRepo.AddWithdrawRecord(ctx, username, wd.Order, wd.Sum)
		if err != nil {
			return fmt.Errorf("failed to add widthdraw record: %w", err)
		}

		order := models.NewOrder(username, wd.Order)

		err = r.ordersRepo.AddNewOrder(ctx, order)
		if err != nil {
			return fmt.Errorf("failed to add new withdraw order: %w", err)
		}

		return nil
	}

	return r.storage.RunInTransaction(callback)
}

func (r *processRepo) GetAllUnprocessedOrders(ctx context.Context) ([]models.Order, error) {
	orders, err := r.ordersRepo.GetAllUnprocessedOrders(ctx)
	if err != nil {
		return nil, err
	}

	return orders, nil
}

func (r *processRepo) UpdateOrderStatus(ctx context.Context, order *models.Order) error {
	return r.ordersRepo.UpdateStatus(ctx, order)
}

func NewProcessRepo(storage *database.ServiceStorage,
	balanceRepo BalanceServiceRepo,
	ordersRepo OrderServiceRepo) ProcessServiceRepo {
	return &processRepo{
		storage:     storage,
		balanceRepo: balanceRepo,
		ordersRepo:  ordersRepo,
	}
}
