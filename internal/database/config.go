package database

import (
	"time"
)

// Config represents the DSN, and options used to connect to a databases.
type Config struct {
	DSN          string        `envconfig:"SQL_DSN" default:"postgresql://postgres:databasePassword@localhost?sslmode=disable"`
	MaxLifetime  time.Duration `envconfig:"MAX_LIFETIME" default:"10m"`
	MaxIdleConns int           `envconfig:"MAX_IDLE_CONNECTIONS" default:"50"`
	MaxOpenConns int           `envconfig:"MAX_OPEN_CONNECTIONS" default:"100"`
}
