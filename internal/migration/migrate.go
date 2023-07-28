package migration

import (
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/mattes/migrate/source/file"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

func newMigrate(logger *zap.Logger, path, dsn string) (*migrate.Migrate, error) {
	m, err := migrate.New(
		"file://"+path,
		dsn,
	)
	if err != nil {
		return nil, err
	}

	m.Log = NewLogger(logger, true)

	return m, nil
}

// PerformUp performs migrations all the way up
// you will need to import the migration drivers for the given database anonymously
//
//	for example: github.com/golang-migrate/migrate/v4/database/postgres
func PerformUp(logger *zap.Logger, path, dsn string) error {
	m, err := newMigrate(logger, path, dsn)
	if err != nil {
		return errors.Wrap(err, "failed to initialise migrate")
	}
	defer m.Close()

	err = runUp(m)
	if err != nil {
		return errors.Wrap(err, "failed to run migrations up")
	}

	err = logVersion(logger, m)
	if err != nil {
		return errors.Wrap(err, "failed to query version")
	}

	return nil
}

func logVersion(logger *zap.Logger, m *migrate.Migrate) error {
	migrateVersion, migrateDirty, err := m.Version()
	if err != nil {
		return err
	}

	logger.Info(
		"migrations completed successfully",
		zap.Uint("version", migrateVersion),
		zap.Bool("dirty", migrateDirty),
	)

	return nil
}

func runUp(m *migrate.Migrate) error {
	err := m.Up()
	if err != nil && err != migrate.ErrNoChange {
		return err
	}

	return nil
}
