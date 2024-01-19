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

// Типы ошибок
const(
	NegativeOnTakeoff = "negative value transaction (to add value use /invoice instead)"
	NegativeOnWithdraw = "negative value transaction (to take off use /withdraw instead)"
	NotEnoughMoney = "not enough money on balance"
)

func (s *TransactionService) AddSum(user bs.Request) error {

	// Проверяем что переданное число положительно
	if user.Sum < 0 {
		return errors.New("negative value transaction (to take off use /withdraw instead)")
	}

	// Проверяем что имеем валюту на счету
	hasCurrency, err := s.repo.HasCurrency(user.WalletID, user.Currency)
	if err != nil{
		s.CreateErrorTransaction(user, bs.TypeInvoice)
		return err
	}

	// Если нет валюты то создаем кошелек
	if !hasCurrency{
		return s.repo.AddWallet(user)
	}

	// Если валюта есть то добавляем
	return s.repo.AddSum(user)
}

func (s *TransactionService) TransferTo(transf bs.Transfer) error {

	// ну тут и ежу понятно
	if transf.Sum < 0 {
		return errors.New("negative value transaction")
	}

	// Создаем временного пользователя для ошибочной транзакции
	user_from := bs.Request{
		WalletID: transf.WalletID_from,
		Currency: transf.Currency,
		Sum: transf.Sum,
	}

	// Проверяем что в первом кошельке есть валюта
	hasCurrency_from, err := s.repo.HasCurrency(transf.WalletID_from, transf.Currency)
	if err != nil{
		s.CreateErrorTransaction(user_from, bs.TypeTransaction)
		log.Printf("service err: %s", err.Error())
		return err
	}

	// Если нет то сразу ошибка
	if !hasCurrency_from{
		s.CreateErrorTransaction(user_from, bs.TypeTransaction)
		return errors.New("no such value on account")
	}

	// Проверяем чтобы не взять больше чем есть на счету
	balance_from, err := s.repo.GetBalanceByID(transf.WalletID_from, transf.Currency)
	if err != nil{
		s.CreateErrorTransaction(user_from, bs.TypeTransaction)
		log.Printf("service err: %s", err.Error())
		return err
	}
	if balance_from - transf.Sum < 0{
		s.CreateErrorTransaction(user_from, bs.TypeTransaction)
		return errors.New("no enough money on first balance")
	}

	// Проверяем что у второго кошелька есть валюта, если нет то добавляем 0.0 новой валюты
	hasCurrency_to, err := s.repo.HasCurrency(transf.WalletID_to, transf.Currency)
	if err != nil{
		s.CreateErrorTransaction(user_from, bs.TypeTransaction)
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
			}, bs.TypeTransaction)
			log.Printf("service err: %s", err.Error())
			return err
		}
	}


	// Если валюта есть то просто переводим
	return s.repo.TransferTo(transf)
}

func (s *TransactionService) TakeOff(user bs.Request) error {

	// Узнаем сколько есть на счету
	var sum float64
	sum, err := s.repo.GetBalanceByID(user.WalletID, user.Currency)

	if err != nil{
		s.CreateErrorTransaction(user, bs.TypeWithdraw)
		return err
	}

	// Проверяем что не заеберем больше чем есть
	if sum-user.Sum < 0 {
		s.CreateErrorTransaction(user, bs.TypeWithdraw)
		return errors.New(NotEnoughMoney)
	}

	// Да.
	if user.Sum < 0 {
		s.CreateErrorTransaction(user, bs.TypeWithdraw)
		return errors.New(NegativeOnTakeoff)
	}
	return s.repo.TakeOff(user)
}

// Гет запросы поэтому бизнес логики нет, чисто запрашиваем из БД 

func (s *TransactionService) GetBalanceByID(walletID uint64, currency string) (float64, error){
	return s.repo.GetBalanceByID(walletID, currency)
}

func (s *TransactionService) GetAllBalancesByID(walletID uint64) ([]bs.WalletCurrency, error){
	return s.repo.GetAllBalancesByID(walletID)
}

// Для удобства вынес в отдельную функцию создание ошибочной транзакции
func (s *TransactionService) CreateErrorTransaction(user bs.Request, typeOfTransaction string) error{
	var id int
	id, err := s.repo.CreateTransaction(user.WalletID, user.Currency, user.Sum, typeOfTransaction)
	s.repo.UpdateStatus("Error", id)
	return err
}
