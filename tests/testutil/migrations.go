package testutil

import (
	"errors"
	"fmt"
	"net/url"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func RunMigrations(dsn, migrationsPath string) (err error) {
	// Resolve migrations directory to an absolute path.
	absPath, err := filepath.Abs(migrationsPath)
	if err != nil {
		return fmt.Errorf("resolve migrations path: %w", err)
	}

	// Construct a file:// URL safely.
	sourceURL := (&url.URL{
		Scheme: "file",
		Path:   absPath,
	}).String()

	m, err := migrate.New(sourceURL, dsn)
	if err != nil {
		return fmt.Errorf("initialize migrations: %w", err)
	}

	// Close migration resources.
	defer func() {
		sourceErr, databaseErr := m.Close()
		if sourceErr != nil {
			sourceErr = fmt.Errorf("close migration source: %w", sourceErr)
		}
		if databaseErr != nil {
			databaseErr = fmt.Errorf("close migration database: %w", databaseErr)
		}
		err = errors.Join(err, sourceErr, databaseErr)
	}()

	// Apply all pending migrations.
	if err := m.Up(); err != nil {
		if err == migrate.ErrNoChange {
			return nil
		}

		return fmt.Errorf("apply migrations: %w", err)
	}

	return nil
}
