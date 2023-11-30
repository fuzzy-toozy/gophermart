package repo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/fuzzy-toozy/gophermart/internal/database"
	"github.com/fuzzy-toozy/gophermart/internal/models"
)

type BalanceServiceRepo interface {
	AddIncomeRecord(ctx context.Context, username string, orderNumber string, income float64) error
	AddWithdrawRecord(ctx context.Context, username string, orderNumber string, outcome float64) error
	GetBanaceData(ctx context.Context, username string) (*models.Balance, error)
	GetWithdrawals(ctx context.Context, username string) ([]models.Withdrawals, error)
}

type balanceQueryConfig struct {
	getBalanceData string
	addNewRecord   string
	getWithdrawals string
}

type balanceServiceRepo struct {
	storage *database.ServiceStorage
	queries balanceQueryConfig
}

type balanceRecord struct {
	username    string
	processedAt time.Time
	income      float64
	outcome     float64
	orderNumber string
}

func newBalanceRecord(username string, ordNumber string, income, outcome float64) *balanceRecord {
	return &balanceRecord{
		username:    username,
		orderNumber: ordNumber,
		processedAt: time.Now(),
		income:      income,
		outcome:     outcome,
	}
}

func newIncomeRecord(username string, ordNumber string, income float64) *balanceRecord {
	record := newBalanceRecord(username, ordNumber, income, 0)
	if record.outcome != 0 {
		panic("balance outcome can't be not 0 for income record")
	}

	return record
}

func newOutcomeRecord(username string, ordNumber string, outcome float64) *balanceRecord {
	record := newBalanceRecord(username, ordNumber, 0, outcome)
	if record.income != 0 {
		panic("balance income can't be not 0 for outcome record")
	}

	return record
}

func getBalanceQueries() balanceQueryConfig {
	c := balanceQueryConfig{}

	c.getBalanceData = "SELECT sum(income)-sum(outcome) as current, sum(outcome) as withdraw FROM balances WHERE username=$1"

	c.addNewRecord = "INSERT INTO balances(username, order_number, income, outcome, processed_at) " +
		"VALUES ($1, $2, $3, $4, $5)"

	c.getWithdrawals = "SELECT order_number, outcome, processed_at " +
		"FROM balances WHERE outcome != 0.0 AND username = $1;"

	return c
}

func (r *balanceServiceRepo) AddIncomeRecord(ctx context.Context, username string, orderNumber string, income float64) error {
	balance := newIncomeRecord(username, orderNumber, income)
	return r.addBalanceRecord(ctx, balance)
}

func (r *balanceServiceRepo) AddWithdrawRecord(ctx context.Context, username string, orderNumber string, outcome float64) error {
	balance := newOutcomeRecord(username, orderNumber, outcome)
	return r.addBalanceRecord(ctx, balance)
}

func (r *balanceServiceRepo) addBalanceRecord(ctx context.Context, record *balanceRecord) error {
	_, err := r.storage.DB.ExecContext(ctx,
		r.queries.addNewRecord, record.username, record.orderNumber, record.income, record.outcome, record.processedAt)
	return err
}

func (r *balanceServiceRepo) GetBanaceData(ctx context.Context, username string) (*models.Balance, error) {
	balance := models.Balance{}

	row := r.storage.DB.QueryRowContext(ctx, r.queries.getBalanceData, username)

	var c, w sql.NullFloat64
	err := row.Scan(&c, &w)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return &balance, err
	}

	balance.Current = c.Float64
	balance.Withdrawn = w.Float64

	return &balance, nil
}

func (r *balanceServiceRepo) GetWithdrawals(ctx context.Context, username string) ([]models.Withdrawals, error) {
	row, err := r.storage.DB.QueryContext(ctx, r.queries.getWithdrawals, username)

	if err != nil {
		return nil, err
	}

	defer row.Close()

	result := make([]models.Withdrawals, 0)

	for row.Next() {
		wd := models.Withdrawals{}
		err := row.Scan(&wd.Order, &wd.Sum, &wd.ProcessedAt)

		if err != nil {
			return nil, fmt.Errorf("row scan error: %w", err)
		}

		result = append(result, wd)
	}

	if err = row.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate all withdrawals: %w", err)
	}

	return result, nil
}

func NewBalanceServiceRepo(storage *database.ServiceStorage) BalanceServiceRepo {
	return &balanceServiceRepo{
		storage: storage,
		queries: getBalanceQueries(),
	}
}
