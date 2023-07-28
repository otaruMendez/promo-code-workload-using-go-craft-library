package main

import (
	"example.com/main/internal/database"
)

// Config represents the DSN, and options used to connect to a databases.
type Config struct {
	Database     database.Config `envconfig:"SERVICE"`
	PromoCSVPath string          `envconfig:"PROMO_CSVPATH" default:"promotions.csv"`
	BatchSize    int             `envconfig:"BATCH_SIZE" default:"50"`
	RedisHost    string          `envconfig:"REDIS_HOST"`
}
