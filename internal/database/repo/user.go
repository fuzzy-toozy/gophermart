package repo

import (
	"context"
	"database/sql"
	"errors"

	"github.com/fuzzy-toozy/gophermart/internal/common"
	"github.com/fuzzy-toozy/gophermart/internal/database"
	"github.com/fuzzy-toozy/gophermart/internal/models"
)

type UserServiceRepo interface {
	GetUserByName(ctx context.Context, username string) (models.User, error)
	AddUser(ctx context.Context, user *models.User) error
}

type queryConfig struct {
	addUserQuery string
	getUserQuery string
}

type userServiceRepo struct {
	storage *database.ServiceStorage
	queries queryConfig
}

func getQueries() queryConfig {
	c := queryConfig{}

	c.addUserQuery = "INSERT INTO users (username, user_password) values ($1, $2)"
	c.getUserQuery = "SELECT username, user_password FROM users WHERE username = $1"

	return c
}

func (r *userServiceRepo) GetUserByName(ctx context.Context, username string) (models.User, error) {
	res := r.storage.DB.QueryRowContext(ctx, r.queries.getUserQuery, username)

	user := models.User{}

	err := res.Scan(&user.Username, &user.Password)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return user, err
	}

	return user, nil
}

func (r *userServiceRepo) AddUser(ctx context.Context, user *models.User) error {
	_, err := r.storage.DB.ExecContext(ctx, r.queries.addUserQuery, user.Username, common.EncryptStringMD5(user.Password))
	return err
}

func NewUserServiceRepo(storage database.ServiceStorage) UserServiceRepo {
	return &userServiceRepo{
		storage: &storage,
		queries: getQueries(),
	}
}
