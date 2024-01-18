package repository

import (
	"bs"
	"database/sql"
	"fmt"
	"log"
	"strings"

	"github.com/jmoiron/sqlx"
)

const (
	Status_neg = "Error"
	Status_pos = "Success"
	Status_ntrl = "Created"
)

type TransactionPostgres struct {
	db *sqlx.DB
}

type WalletCurrency struct{
	Currency string `json:"currency"`
	Value float64 `json:"value"`
}

func NewTransactionPostgres(db *sqlx.DB) *TransactionPostgres {
	return &TransactionPostgres{db: db}
}

func (r *TransactionPostgres) AddSum(user bs.Request) error {

	// Добавляем транзакцию в таблицу (Если потом пойдет не так поменяем на ошибочное) 
	var id int
	id, err := r.CreateTransaction(user.WalletID, user.Currency, user.Sum)
	if err != nil {
		return err
	}
	r.UpdateStatus(Status_ntrl, id)

	// Добавляем к уже существующему значению новую сумму
	st := fmt.Sprintf("UPDATE %s SET value = ROUND (CAST(value AS numeric) + $1, 2) WHERE wallet_id = $2 and currency ILIKE $3", WalletTable)
	_, err = r.db.Exec(st, user.Sum, user.WalletID, "%" + user.Currency + "%")
	if err != nil {
		r.UpdateStatus(Status_neg, id)
		return err
	}
	
	r.UpdateStatus(Status_pos, id)
	return err

}

func (r *TransactionPostgres) AddWallet(user bs.Request) error {

	var id int
	id, err := r.CreateTransaction(user.WalletID, user.Currency, user.Sum)
	if err != nil {
		log.Println("error occured while adding transaction to DB: ", err)
		return err
	}
	r.UpdateStatus(Status_ntrl, id)

	_, err = r.db.Exec(`INSERT INTO Wallets (wallet_id, currency, value) VALUES ($1, $2, $3)`, user.WalletID, user.Currency, user.Sum)
	if err != nil {
		r.UpdateStatus(Status_neg, id)
		log.Println("Addwallet on repo: ", err)
		return err
	}	
	r.UpdateStatus(Status_pos, id)
	return nil

}

func (r *TransactionPostgres) TakeOff(user bs.Request) error {
	var id int

	id, err := r.CreateTransaction(user.WalletID, user.Currency, user.Sum)
	if err != nil{
		return err
	}

	Take_money := fmt.Sprintf("UPDATE %s SET value = ROUND (CAST(value AS numeric) - $1, 2) WHERE wallet_id = $2 and currency ILIKE $3", WalletTable)
	_, err = r.db.Exec(Take_money, user.Sum, user.WalletID, "%" + user.Currency + "%")
	if err != nil {
		r.UpdateStatus(Status_neg, id)
		return err
	}

	r.UpdateStatus(Status_pos, id)
	return err
}

func (r *TransactionPostgres) TransferTo(transf bs.Transfer) error {
	id1, err := r.CreateTransaction(transf.WalletID_from, transf.Currency, transf.Sum)
 	if err != nil{
		log.Println("TransferTO creating transaction 1 eror: ", err)
 	 	return err
 	}
	r.UpdateStatus(Status_ntrl, id1)

	
 	id2, err := r.CreateTransaction(transf.WalletID_to, transf.Currency, transf.Sum)
 	if err != nil{
		r.UpdateStatus(Status_neg, id1)
		log.Println("TransferTO creating transaction 2 eror: ", err)
 	 	return err
 	}
	r.UpdateStatus(Status_ntrl, id2)


	tx, err := r.db.Begin()
    if err != nil {
		r.UpdateStatus(Status_neg, id1)
		r.UpdateStatus(Status_neg, id2)
		log.Println("TransferTO error on creating sql-transaction occured: ", err)
        return err
    }
	stmt, err := tx.Prepare(`
		UPDATE wallets SET value = ROUND (CAST(value AS numeric) - $1, 2)
	 	WHERE wallet_id = $2 and currency ILIKE $3;
	`)
	if err != nil {
		tx.Rollback()
		r.UpdateStatus(Status_neg, id1)
		r.UpdateStatus(Status_neg, id2)
		log.Println("TransferTO error on first prepare: ", err)
		return err
	}	
	defer stmt.Close()

	if _, err := stmt.Exec(transf.Sum, transf.WalletID_from, "%" + transf.Currency + "%"); err != nil {
        tx.Rollback() 
		r.UpdateStatus(Status_neg, id1)
		r.UpdateStatus(Status_neg, id2)
		log.Println("TransferTO error on exec of first upd: ", err)
        return err
    }

    stmt, err = tx.Prepare(`
		UPDATE wallets SET value = ROUND (CAST(value AS numeric) + $1, 2)
		WHERE wallet_id = $2 AND currency ILIKE $3;
	`)
    if err != nil {
        tx.Rollback()
		r.UpdateStatus(Status_neg, id1)
		r.UpdateStatus(Status_neg, id2)
		log.Println("TransferTO error on second prepare: ", err)
        return err
    }
    defer stmt.Close()

    if _, err := stmt.Exec(transf.Sum, transf.WalletID_to, "%" + transf.Currency + "%"); err != nil {
        tx.Rollback() 
		r.UpdateStatus(Status_neg, id1)
		r.UpdateStatus(Status_neg, id2)
		log.Println("TransferTO error on exec of second upd: ", err)

        return err
    }

	r.UpdateStatus(Status_pos, id1)
	r.UpdateStatus(Status_pos, id2)

    return tx.Commit()
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
	st := fmt.Sprintf("SELECT value FROM %s WHERE wallet_id = $1 and currency ILIKE $2", WalletTable)
	if err := r.db.Get(&reqBalance, st, walletID, "%" + currency + "%"); err != nil {
		fmt.Println("JA tut suka")
		return 0.0, err
	}
	return reqBalance, nil

}

func (r *TransactionPostgres) GetAllBalancesByID(walletID uint64) ([]bs.WalletCurrency, error){

	st := fmt.Sprintf("SELECT currency, value FROM %s WHERE wallet_id = $1", WalletTable)
	rows, err := r.db.Query(st, walletID)
	defer rows.Close()

	var ans_wallets []bs.WalletCurrency

	for rows.Next() {
		//var temp uint64
        var wal bs.WalletCurrency
        if err := rows.Scan(&wal.Currency, &wal.Value); err != nil {
            return ans_wallets, err
        }
		wal.Currency = strings.TrimSpace(wal.Currency)
        ans_wallets = append(ans_wallets, wal)
    }

	if err = rows.Err(); err != nil {
        return ans_wallets, err
    }

	return ans_wallets, nil

}

func (r *TransactionPostgres) HasCurrency(walletID uint64, currency string) (bool, error){

	st := fmt.Sprintf("SELECT value FROM %s WHERE wallet_id = $1 and currency ILIKE $2", WalletTable)
	row := r.db.QueryRow(st, walletID, "%" + currency + "%")

	var temp any
	if err := row.Scan(&temp); err == sql.ErrNoRows {
		return false, nil
	} else if err != nil{
		return false, err
	}
	return true, nil

}
