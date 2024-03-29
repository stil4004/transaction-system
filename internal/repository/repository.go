package repository

import (
	"bs"

	"github.com/jmoiron/sqlx"
)

type Transactions interface {
	AddSum(user bs.Request) error
	AddWallet(user bs.Request) error
	TakeOff(user bs.Request) error
	TransferTo(transf bs.Transfer) error
	GetBalanceByID(walletID uint64, currency string) (float64, error)
	GetAllBalancesByID(walletID uint64) ([]bs.WalletCurrency, error)
	UpdateStatus(status string, id int) error
	CreateTransaction(wallet_id uint64, currency string, sum float64, operationType string) (int, error) 
	HasCurrency(walletID uint64, currency string) (bool, error)
}

type Repository struct {
	Transactions
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{
		Transactions: NewTransactionPostgres(db),
	}
}
