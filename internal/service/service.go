package service

import (
	"bs"
	"bs/internal/repository"
) 

type Transactions interface {
	AddSum(user bs.Request) error
	TakeOff(user bs.Request) error
	GetBalance() ([]bs.Answer, error)
}

type Service struct {
	Transactions
}

func NewService(repos *repository.Repository) *Service {
	return &Service{
		Transactions: NewTransactionService(repos),
	}
}
