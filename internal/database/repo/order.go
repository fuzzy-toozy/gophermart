package repo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/fuzzy-toozy/gophermart/internal/database"
	"github.com/fuzzy-toozy/gophermart/internal/models"
)

type OrderServiceRepo interface {
	GetOrderByNumber(ctx context.Context, number string) (*models.Order, error)
	GetAllUserOrders(ctx context.Context, username string) ([]models.Order, error)
	GetAllUnprocessedOrders(ctx context.Context) ([]models.Order, error)

	AddNewOrder(ctx context.Context, order *models.Order) error

	UpdateStatus(ctx context.Context, order *models.Order) error
	UpdateAccural(ctx context.Context, order *models.Order, accural float64) error
}

type orderQueryConfig struct {
	getOrderByUsername     string
	addNewOrder            string
	updateStatus           string
	updateAccural          string
	getAllUserOrders       string
	getAllUnprocessedOders string
}

type orderServiceRepo struct {
	storage *database.ServiceStorage
	queries orderQueryConfig
}

func getOrderQueries() orderQueryConfig {
	c := orderQueryConfig{}

	c.getOrderByUsername = "SELECT number, username, uploaded_at, status, accrual FROM orders WHERE number = $1"

	c.addNewOrder = "INSERT INTO orders(number, username, uploaded_at, status) " +
		"VALUES ($1, $2, $3, $4) ON CONFLICT (number) DO NOTHING"

	c.updateStatus = "UPDATE orders SET status = $1 WHERE number = $2"

	c.updateAccural = "UPDATE orders SET accrual = $1 WHERE number = $2"

	c.getAllUserOrders = "SELECT number, username, uploaded_at, status, accrual FROM orders WHERE username = $1"

	c.getAllUnprocessedOders = "SELECT * FROM orders WHERE status in ('NEW', 'PROCESSING')"

	return c
}

func (r *orderServiceRepo) GetOrderByNumber(ctx context.Context, number string) (*models.Order, error) {
	order := models.Order{}

	row := r.storage.DB.QueryRowContext(ctx, r.queries.getOrderByUsername, number)

	err := row.Scan(&order.Number, &order.Username, &order.UploadedAt, &order.Status, &order.Accrual)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return &order, err
	}

	return &order, nil
}

func (r *orderServiceRepo) AddNewOrder(ctx context.Context, order *models.Order) error {
	_, err := r.storage.DB.ExecContext(ctx, r.queries.addNewOrder, order.Number, order.Username, order.UploadedAt, order.Status)
	return err
}

func (r *orderServiceRepo) updateOrder(ctx context.Context, order *models.Order, query string, args ...any) error {
	res, err := r.storage.DB.ExecContext(ctx, query, args...)

	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()

	if err != nil {
		return err
	}

	if rowsAffected < 1 {
		return fmt.Errorf("order with number '%v' doesn't exist", order.Number)
	}

	return nil
}

func (r *orderServiceRepo) UpdateStatus(ctx context.Context, order *models.Order) error {
	return r.updateOrder(ctx, order, r.queries.updateStatus, order.Status, order.Number)
}

func (r *orderServiceRepo) UpdateAccural(ctx context.Context, order *models.Order, accural float64) error {
	return r.updateOrder(ctx, order, r.queries.updateAccural, accural, order.Number)
}

func (r *orderServiceRepo) getOrders(ctx context.Context, query string, args ...any) ([]models.Order, error) {
	row, err := r.storage.DB.QueryContext(ctx, query, args...)

	if err != nil {
		return nil, err
	}

	defer row.Close()

	result := make([]models.Order, 0)

	for row.Next() {
		order := models.Order{}
		err := row.Scan(&order.Number, &order.Username, &order.UploadedAt, &order.Status, &order.Accrual)

		if err != nil {
			return nil, fmt.Errorf("row scan error: %w", err)
		}

		result = append(result, order)
	}

	if err = row.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate all orders: %w", err)
	}

	return result, nil
}

func (r *orderServiceRepo) GetAllUserOrders(ctx context.Context, username string) ([]models.Order, error) {
	return r.getOrders(ctx, r.queries.getAllUserOrders, username)
}

func (r *orderServiceRepo) GetAllUnprocessedOrders(ctx context.Context) ([]models.Order, error) {
	return r.getOrders(ctx, r.queries.getAllUnprocessedOders)
}

func NewOrderServiceRepo(storage *database.ServiceStorage) OrderServiceRepo {
	r := orderServiceRepo{
		storage: storage,
		queries: getOrderQueries(),
	}

	return &r
}
