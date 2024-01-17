package repository

import (
	"bs"
	"fmt"
	"log"

	"github.com/jmoiron/sqlx"
)

const (
	Status_neg = "Error"
	Status_pos = "Success"
)

type TransactionPostgres struct {
	db *sqlx.DB
}

func NewTransactionPostgres(db *sqlx.DB) *TransactionPostgres {
	return &TransactionPostgres{db: db}
}

func (r *TransactionPostgres) AddSum(user bs.Request) error {

	var id int
	id, err := r.CreateTransaction(user.WalletID, user.Currency, user.Sum)
	if err != nil {
		return err
	}

	st := fmt.Sprintf("UPDATE %s SET ? = ? + ? WHERE wallet_id = ? ", WalletTable)
	_, err = r.db.Exec(st, user.Currency, user.Currency, user.Sum, user.WalletID)
	if err != nil {
		r.UpdateStatus(Status_neg, id)
		return err
	}
	
	r.UpdateStatus(Status_pos, id)
	return err

}

func (r *TransactionPostgres) TakeOff(user bs.Request) error {
	var id int

	id, err := r.CreateTransaction(user.WalletID, user.Currency, user.Sum)
	if err != nil{
		return err
	}

	// var sum float64
	// st := fmt.Sprintf("SELECT ? FROM %s WHERE wallet_id = ?", WalletTable)
	// if err := r.db.Get(&sum, st, user.Currency, user.WalletID); err != nil {
	// 	return err
	// }

	// if sum-user.Sum < 0 {
	// 	r.UpdateStatus(Status_neg, id)
	// 	return nil
	// }

	Take_money := fmt.Sprintf("UPDATE %s SET ? = ? - ? WHERE wallet_id = ? ", WalletTable)
	_, err = r.db.Exec(Take_money, user.Currency, user.Currency, user.Sum, user.WalletID)
	if err != nil {
		r.UpdateStatus(Status_neg, id)
		return err
	}

	r.UpdateStatus(Status_pos, id)
	return err
}

func (r *TransactionPostgres) GetBalance() ([]bs.Answer, error) {
	var list []bs.Answer

	query := fmt.Sprintf(`SELECT t2.wallet_id ,t2.usdt, t2.rub, t2.eur FROM %s t2 JOIN 
		(SELECT wallet_id, status, ROW_NUMBER() OVER (PARTITION BY wallet_id ORDER BY id DESC) 
		AS rn FROM %s) t1 ON t1.wallet_id = t2.wallet_id WHERE t1.rn = 1 AND t1.status != 'Error'`, WalletTable, TransacitonTable)
	err := r.db.Select(&list, query)

	return list, err
}

func (r *TransactionPostgres) UpdateStatus(status string, id int) error {
	st := fmt.Sprintf("UPDATE %s SET status = $1 WHERE id = $2", TransacitonTable)
	_, err := r.db.Exec(st, status, id)
	return err
}

func (r *TransactionPostgres) CreateTransaction(wallet_id uint64, currency string, sum float64) (int, error) {
	var id int
	st := fmt.Sprintf("INSERT INTO %s (wallet_id, currency, sum) VALUES ($1, $2, $3) RETURNING id", TransacitonTable)
	row := r.db.QueryRow(st, wallet_id, currency, sum)
	err := row.Scan(&id)
	
	if err != nil {
		log.Println(err.Error())
	}
	return id, err
}

func (r *TransactionPostgres) GetBalanceByID(walletID uint64, currency string) (float64, error){
	var reqBalance float64
	st := fmt.Sprintf("SELECT ? FROM %s WHERE wallet_id = ?", WalletTable)
	if err := r.db.Get(&reqBalance, st, currency, walletID); err != nil {
		return 0.0, err
	}
	return reqBalance, nil

}
