package database

import (
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"go.nhat.io/otelsql"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

const dbDriver = "postgres"

func Connect(cfg Config) (*sqlx.DB, error) {
	driverName, err := otelsql.Register(
		dbDriver,
		otelsql.TracePing(),
		otelsql.TraceQueryWithoutArgs(),
		otelsql.WithSystem(semconv.DBSystemPostgreSQL), // Optional.
	)
	if err != nil {
		return nil, fmt.Errorf("cound not register tracing with db driver: %w", err)
	}

	sqlDB, err := sql.Open(driverName, cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("could not open DB connection: %w", err)
	}

	sqlxDB := sqlx.NewDb(sqlDB, dbDriver)
	if err := sqlxDB.Ping(); err != nil {
		return nil, fmt.Errorf("could not ping DB after connecting: %w", err)
	}

	sqlxDB.SetConnMaxLifetime(cfg.MaxLifetime)
	sqlxDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlxDB.SetMaxOpenConns(cfg.MaxOpenConns)

	return sqlxDB, err
}
