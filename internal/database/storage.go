package database

import (
	"database/sql"
	"fmt"
)

type queryConfig struct {
	addUserQuery string
	getUserQuery string
}

type ServiceStorage struct {
	DB       *sql.DB
	queries  queryConfig
	dbConfig DBConfig
}

func (s *ServiceStorage) RunInTransaction(callback func() error) error {
	tx, err := s.DB.Begin()

	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer tx.Rollback()

	err = callback()
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func getQueries() queryConfig {
	c := queryConfig{}

	c.addUserQuery = "INSERT INTO users (username, user_password) values ($1, $2)"
	c.getUserQuery = "SELECT username, user_password FROM users WHERE username = $1"

	return c
}

func NewServiceStorage(dbconfig DBConfig) (*ServiceStorage, error) {
	r := ServiceStorage{}

	db, err := sql.Open(dbconfig.DriverName, dbconfig.ConnURI)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	r.DB = db
	r.dbConfig = dbconfig
	r.queries = getQueries()

	return &r, nil
}

func (s *ServiceStorage) Close() error {
	return s.DB.Close()
}
