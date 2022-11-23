package pg

import (
	"embed"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed migrations/*.sql
var fs embed.FS

// Migrate takes care of executing the migrations on the database identified
// by the given databaseURL.
func Migrate(databaseURL string) error {
	d, err := iofs.New(fs, "migrations")
	if err != nil {
		return fmt.Errorf("cannot initialize iofs: %w", err)
	}

	migrator, err := migrate.NewWithSourceInstance("iofs", d, databaseURL)
	if err != nil {
		return fmt.Errorf("cannot instantiate migrator: %w", err)
	}

	return migrator.Up()
}
