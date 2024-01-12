package repository

import (
	"bs"

	"github.com/jmoiron/sqlx"
)

type Transactions interface {
	AddSum(user bs.Request) error
	TakeOff(user bs.Request) error
	GetBalance() ([]bs.Answer, error)
}

type Repository struct {
	Transactions
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{
		Transactions: NewTransactionPostgres(db),
	}
}
