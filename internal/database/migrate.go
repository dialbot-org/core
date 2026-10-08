package database

import (
	config "core/internal/config"
	logger "core/pkg/log"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func Migrate(cfg *config.DatabaseConfig) error {
	dsn := cfg.GetMigrationDSN()
	migration, err := migrate.New(
		"file://internal/migrations",
		dsn,
	)

	if err != nil {
		logger.Fatal(err, "Cant Run Migrations")
		return err
	}

	if err := migration.Up(); err != nil && err != migrate.ErrNoChange {
		logger.Fatal(err, "Cant Run Up Migrations")
		return err
	}

	return nil
}
