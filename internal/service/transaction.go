package service

import (
	"bs"
	"bs/internal/repository"
	"errors"
	"log"
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
		s.CreateErrorTransaction(user)
		return err
	}

	if !hasCurrency{
		return s.repo.AddWallet(user)
	}

	// If not adding to existing
	return s.repo.AddSum(user)
}

func (s *TransactionService) TransferTo(transf bs.Transfer) error {

	// ну тут и ежу понятно
	if transf.Sum < 0 {
		return errors.New("negative value transaction")
	}

	// Creating temp user for CreateError function
	user_from := bs.Request{
		WalletID: transf.WalletID_from,
		Currency: transf.Currency,
		Sum: transf.Sum,
	}

	// Checking if FROM wallet has currency
	hasCurrency_from, err := s.repo.HasCurrency(transf.WalletID_from, transf.Currency)
	if err != nil{
		s.CreateErrorTransaction(user_from)
		log.Printf("service err: %s", err.Error())
		return err
	}

	// If not throw err
	if !hasCurrency_from{
		s.CreateErrorTransaction(user_from)
		return errors.New("no such value on account")
	}

	// Checking for negative balance after take off
	balance_from, err := s.repo.GetBalanceByID(transf.WalletID_from, transf.Currency)
	if err != nil{
		s.CreateErrorTransaction(user_from)
		log.Printf("service err: %s", err.Error())
		return err
	}
	if balance_from - transf.Sum < 0{
		s.CreateErrorTransaction(user_from)
		return errors.New("no enough money on first balance")
	}

	hasCurrency_to, err := s.repo.HasCurrency(transf.WalletID_to, transf.Currency)
	if err != nil{
		s.CreateErrorTransaction(user_from)
		log.Printf("service err: %s", err.Error())
		return err
	}
	if !hasCurrency_to{
		err = s.repo.AddWallet(bs.Request{transf.WalletID_to, transf.Currency, 0.0})
		if err != nil{
			s.CreateErrorTransaction(bs.Request{
				WalletID: transf.WalletID_to,
				Currency: transf.Currency,
				Sum: 0.0,
			})
			log.Printf("service err: %s", err.Error())
			return err
		}
	}


	// If not adding to existing
	return s.repo.TransferTo(transf)
}

func (s *TransactionService) TakeOff(user bs.Request) error {

	var sum float64
	sum, err := s.repo.GetBalanceByID(user.WalletID, user.Currency)

	if err != nil{
		s.CreateErrorTransaction(user)
		return err
	}

	// Checking for taking off more then we have
	if sum-user.Sum < 0 {
		//r.UpdateStatus(Status_neg, id)
		s.CreateErrorTransaction(user)
		return errors.New(NotEnoughMoney)
	}

	// Да.
	if user.Sum < 0 {
		s.CreateErrorTransaction(user)
		return errors.New(NegativeOnTakeoff)
	}
	return s.repo.TakeOff(user)
}

func (s *TransactionService) GetBalanceByID(walletID uint64, currency string) (float64, error){
	return s.repo.GetBalanceByID(walletID, currency)
}

func (s *TransactionService) GetAllBalancesByID(walletID uint64) ([]bs.WalletCurrency, error){
	return s.repo.GetAllBalancesByID(walletID)
}

// Для удобства вынес в отдельную функцию создание ошибочной транзакции
func (s *TransactionService) CreateErrorTransaction(user bs.Request) error{
	var id int
	id, err := s.repo.CreateTransaction(user.WalletID, user.Currency, user.Sum)
	s.repo.UpdateStatus("Error", id)
	return err
}
