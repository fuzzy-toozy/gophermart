package database

import (
	"errors"

	"github.com/golang-migrate/migrate"
)

func Migrate(dbURL string) error {
	m, err := migrate.New("file://migration/dbscripts/", dbURL)

	defer m.Close()

	if err != nil {
		return err
	}

	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}

	err = m.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}

	return nil
}
