package repository

import (
	"bs"

	"github.com/jmoiron/sqlx"
)

type Transactions interface {
	AddSum(user bs.Request) error
	TakeOff(user bs.Request) error
	GetBalance() ([]bs.Answer, error)
	GetBalanceByID(walletID uint64, currency string) (float64, error)
	UpdateStatus(status string, id int) error
	CreateTransaction(wallet_id uint64, currency string, sum float64) (int, error) 
}

type Repository struct {
	Transactions
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{
		Transactions: NewTransactionPostgres(db),
	}
}
