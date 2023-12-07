package services

import (
	"context"
	"net/http"
	"time"

	"github.com/fuzzy-toozy/gophermart/internal/common"
	"github.com/fuzzy-toozy/gophermart/internal/database/repo"
	"github.com/fuzzy-toozy/gophermart/internal/errors"
	"github.com/fuzzy-toozy/gophermart/internal/models"
)

type UserService struct {
	repo repo.UserServiceRepo
	auth TokenService
}

func NewUserService(repo repo.UserServiceRepo, auth TokenService) *UserService {
	return &UserService{
		repo: repo,
		auth: auth,
	}
}

func (s *UserService) Authenticate(token string) (username string, serr errors.ServiceError) {
	claims, err := s.auth.Validate(token)
	if err != nil {
		return "", errors.NewServiceError(http.StatusUnauthorized, "failed to validate jwt token: %w", err)
	}

	if claims.Expiry.Time().Before(time.Now()) {
		return "", errors.NewServiceError(http.StatusUnauthorized, "token for user `%v` is expired", claims.Subject)
	}

	return claims.Subject, nil
}

func (s *UserService) AuthDuration() time.Duration {
	return s.auth.Duration()
}

func (s *UserService) Register(ctx context.Context, user *models.User) (token string, serr errors.ServiceError) {
	userDB, err := s.repo.GetUserByName(ctx, user.Username)

	if err != nil {
		return "", errors.NewServiceError(http.StatusInternalServerError,
			"failed to get user '%v' from db: %w", user.Username, err)
	}

	if len(userDB.Username) > 0 {
		return "", errors.NewServiceError(http.StatusConflict, "user with name '%v' already exists", user.Username)
	}

	token, err = s.auth.Generate(user)
	if err != nil {
		return "", errors.NewServiceError(http.StatusBadRequest, "failed to generate jwt token for user '%v': %w", user.Username, err)
	}

	err = s.repo.AddUser(ctx, user)
	if err != nil {
		return "", errors.NewServiceError(http.StatusBadRequest, "failed to add new user to db: %v", err)
	}

	return token, nil
}

func (s *UserService) Login(ctx context.Context, user *models.User) (token string, serr errors.ServiceError) {
	if user.Password == "" || user.Username == "" {
		return "", errors.NewServiceError(http.StatusBadRequest, "password or username is empty")

	}

	userDB, err := s.repo.GetUserByName(ctx, user.Username)
	if err != nil {
		return "", errors.NewServiceError(http.StatusInternalServerError, "failed to get user '%v' from db: %w", user.Username, err)
	}

	if len(userDB.Username) == 0 {
		return "", errors.NewServiceError(http.StatusUnauthorized, "user '%v' is not registered", user.Username)
	}

	if userDB.Password != common.EncryptStringMD5(user.Password) {
		return "", errors.NewServiceError(http.StatusUnauthorized, "user '%v' provided invalid password", user.Username)

	}

	token, err = s.auth.Generate(&userDB)
	if err != nil {
		return "", errors.NewServiceError(http.StatusInternalServerError,
			"failed to generate jwt token for user '%v': %w", userDB.Username, err)

	}

	return token, nil
}
