package service

import (
	"bs"
	"bs/internal/repository"
) 

type Transactions interface {
	AddSum(user bs.Request) error
	TakeOff(user bs.Request) error
	TransferTo(transf bs.Transfer) error
	GetBalanceByID(walletID uint64, currency string) (float64, error)
	GetAllBalancesByID(walletID uint64) ([]bs.WalletCurrency, error)
}

type Service struct {
	Transactions
}

func NewService(repos *repository.Repository) *Service {
	return &Service{
		Transactions: NewTransactionService(repos),
	}
}
