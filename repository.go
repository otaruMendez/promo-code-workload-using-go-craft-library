package main

import (
	"context"
	"database/sql"

	"example.com/main/internal/database"
	"github.com/jmoiron/sqlx"
)

type Promo struct {
	ID    string `json:"id" db:"id"`
	Price string `json:"price" db:"price"`
	Date  string `json:"expiration_date" db:"expiration_date"`
}

type Repository interface {
	UpsertPromo(ctx context.Context, tableID int, pk int, p Promo) error
	GetPromo(ctx context.Context, promoId int) (Promo, error)
}

type Repo struct {
	db *sqlx.DB
}

// Creates a new promo repo
func NewRepo(db *sqlx.DB) *Repo {
	return &Repo{
		db: db,
	}
}

// GetPromo returns a promo from DB
func (repo *Repo) GetPromo(ctx context.Context, promoPK int) (promo Promo, err error) {
	err = repo.db.QueryRowContext(
		ctx,
		`SELECT id, price, expiration_date FROM promotions WHERE pk = $1`,
		promoPK,
	).Scan(&promo.ID, &promo.Price, &promo.Date)

	if err == sql.ErrNoRows {
		return promo, database.ErrRecordNotFound
	}

	if err != nil {
		return promo, err
	}

	return promo, nil
}

// UpsertPromo creates a promo record and updates the info if it already exists
func (repo *Repo) UpsertPromo(ctx context.Context, historyID int, pk int, p Promo) error {
	_, err := repo.db.NamedExec(`
			INSERT INTO promotions (pk,id,price,expiration_date,version)
			VALUES (:pk,:id,:price,:date,:version)
			ON CONFLICT (pk)
			DO UPDATE SET price = :price, expiration_date = :date, version = :version
		`,
		map[string]any{
			"pk":      pk,
			"id":      p.ID,
			"price":   p.Price,
			"date":    p.Date,
			"version": historyID,
		},
	)

	return err
}
