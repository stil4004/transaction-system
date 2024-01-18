package service

import (
	"bs"
	"bs/internal/repository"
	"errors"
)

type TransactionService struct {
	repo repository.Transactions
}

func NewTransactionService(repo repository.Transactions) *TransactionService {
	return &TransactionService{repo: repo}
}

const(
	NegativeOnTakeoff = "negative value transaction (to add value use /invoice instead)"
	NegativeOnWithdraw = "negative value transaction (to take off use /withdraw instead)"
	NotEnoughMoney = "not enough money on balance"
)

func (s *TransactionService) AddSum(user bs.Request) error {

	// ну тут и ежу понятно
	if user.Sum < 0 {
		return errors.New("negative value transaction (to take off use /withdraw instead)")
	}

	// Checking if we got currency in table
	hasCurrency, err := s.repo.HasCurrency(user.WalletID, user.Currency)
	if err != nil{
		s.CreateErrorTransaction(user, err.Error())
		return err
	}

	if !hasCurrency{
		return s.repo.AddWallet(user)
	}

	// If not adding to existing
	return s.repo.AddSum(user)
}

func (s *TransactionService) TakeOff(user bs.Request) error {

	var sum float64
	sum, err := s.repo.GetBalanceByID(user.WalletID, user.Currency)

	if err != nil{
		s.CreateErrorTransaction(user, err.Error())
		return err
	}

	// Checking for taking off more then we have
	if sum-user.Sum < 0 {
		//r.UpdateStatus(Status_neg, id)
		s.CreateErrorTransaction(user, NotEnoughMoney)
		return errors.New(NotEnoughMoney)
	}

	// Да.
	if user.Sum < 0 {
		s.CreateErrorTransaction(user, NegativeOnTakeoff)
		return errors.New(NegativeOnTakeoff)
	}
	return s.repo.TakeOff(user)
}

func (s *TransactionService) GetBalance() ([]bs.Answer, error) {
	return s.repo.GetBalance()
}

func (s *TransactionService) GetBalanceByID(walletID uint64, currency string) (float64, error){
	return s.repo.GetBalanceByID(walletID, currency)
}

// Для удобства вынес в отдельную функцию создание ошибочной транзакции
func (s *TransactionService) CreateErrorTransaction(user bs.Request, ErrorType string) error{
	var id int
	id, err := s.repo.CreateTransaction(user.WalletID, user.Currency, user.Sum)
	s.repo.UpdateStatus("Error", id)
	return err
}
