package service

import (
	"bs"
	"bs/internal/repository"
)

type TransactionService struct {
	repo repository.Transactions
}

func NewTransactionService(repo repository.Transactions) *TransactionService {
	return &TransactionService{repo: repo}
}

func (s *TransactionService) AddSum(user bs.Request) error {
	return s.repo.AddSum(user)
}

func (s *TransactionService) TakeOff(user bs.Request) error {
	return s.repo.TakeOff(user)
}

func (s *TransactionService) GetBalance() ([]bs.Answer, error) {
	return s.repo.GetBalance()
}
