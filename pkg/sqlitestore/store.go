package sqlitestore

import (
	"database/sql"
	"embed"
	"fmt"
	"io/fs"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:generate sqlc generate

//go:embed migrations/*.sql
var migrationsFS embed.FS

func Migrate(db *sql.DB) error {
	// Run migrations
	migrationFS, err := fs.Sub(migrationsFS, "migrations")
	if err != nil {
		return err
	}

	d, err := iofs.New(migrationFS, ".")
	if err != nil {
		return err
	}

	driver, err := sqlite3.WithInstance(db, &sqlite3.Config{})
	if err != nil {
		return err
	}

	m, err := migrate.NewWithInstance("iofs", d, "sqlite3", driver)
	if err != nil {
		return fmt.Errorf("failed to create migrator: %w", err)
	}

	err = m.Up()
	if err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}
