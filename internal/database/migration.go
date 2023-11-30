package database

import (
	"errors"

	"github.com/golang-migrate/migrate"
)

func Migrate(dbURL string) error {
	m, err := migrate.New("file://migration/dbscripts/", dbURL)

	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}

	err = m.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}

	m.Close()

	return nil
}
