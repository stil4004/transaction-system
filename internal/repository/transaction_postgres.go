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
	Status_ntrl = "Created" // По дефолту ставится в таблице
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
	id, err := r.CreateTransaction(user.WalletID, user.Currency, user.Sum, bs.TypeInvoice)
	if err != nil {
		return err
	}

	// Добавляем к уже существующему значению новую сумму
	st := fmt.Sprintf("UPDATE %s SET value = ROUND (CAST(value AS numeric) + $1, 2) WHERE wallet_id = $2 and currency ILIKE $3", WalletTable)
	_, err = r.db.Exec(st, user.Sum, user.WalletID, "%" + user.Currency + "%")
	if err != nil {
		r.UpdateStatus(Status_neg, id)
		return err
	}
	
	// Ставим в бд что транзакция успешна
	r.UpdateStatus(Status_pos, id)
	return err

}

func (r *TransactionPostgres) AddWallet(user bs.Request) error {

	// Cоздаем запись в транзакциях об операции 
	var id int
	id, err := r.CreateTransaction(user.WalletID, user.Currency, user.Sum, bs.TypeInvoice)
	if err != nil {
		log.Println("error occured while adding transaction to DB: ", err)
		return err
	}

	// задаем статус Created (UPD: Забыл что там дефолтно задается Created)
	//r.UpdateStatus(Status_ntrl, id)


	// Вставляем данные в таблицу кошельков
	_, err = r.db.Exec(`INSERT INTO Wallets (wallet_id, currency, value) VALUES ($1, $2, $3)`, user.WalletID,  strings.ToUpper(user.Currency), user.Sum)
	if err != nil {
		r.UpdateStatus(Status_neg, id)
		log.Println("Addwallet on repo: ", err)
		return err
	}
	
	// Если все успешно указываем это в истори транзакций
	r.UpdateStatus(Status_pos, id)
	return nil

}

func (r *TransactionPostgres) TakeOff(user bs.Request) error {
	
	//  Делаем запись в истории транзакций и забираем айдишник
	var id int
	id, err := r.CreateTransaction(user.WalletID, user.Currency, user.Sum, bs.TypeWithdraw)
	if err != nil{
		return err
	}

	// Изменяем таблицу с кошельками, с учетом что на балансе не больше двух цифр после запятой
	Take_money := fmt.Sprintf("UPDATE %s SET value = ROUND (CAST(value AS numeric) - $1, 2) WHERE wallet_id = $2 and currency ILIKE $3", WalletTable)
	_, err = r.db.Exec(Take_money, user.Sum, user.WalletID, "%" + user.Currency + "%")
	if err != nil {
		r.UpdateStatus(Status_neg, id)
		return err
	}

	// Если все ок подтверждаем все в истории
	r.UpdateStatus(Status_pos, id)
	return err
}

func (r *TransactionPostgres) TransferTo(transf bs.Transfer) error {

	// Создаем запись о снятии денег с аккаунта
	id1, err := r.CreateTransaction(transf.WalletID_from, transf.Currency, transf.Sum, bs.TypeTransaction)
 	if err != nil{
		log.Println("TransferTO creating transaction 1 eror: ", err)
 	 	return err
 	}
	r.UpdateStatus(Status_ntrl, id1)

	// Создаем запись о занесении денег на аккаунт
 	id2, err := r.CreateTransaction(transf.WalletID_to, transf.Currency, transf.Sum, bs.TypeTransaction)
 	if err != nil{
		r.UpdateStatus(Status_neg, id1)
		log.Println("TransferTO creating transaction 2 eror: ", err)
 	 	return err
 	}

	// Объявляем sql транзакцию
	tx, err := r.db.Begin()
    if err != nil {
		r.UpdateStatus(Status_neg, id1)
		r.UpdateStatus(Status_neg, id2)
		log.Println("TransferTO error on creating sql-transaction occured: ", err)
        return err
    }
	
	// Снимаем деньги с одного счета
	// Объявляем первую транзакцию снятия денег
	stmt, err := tx.Prepare(`
		UPDATE wallets SET value = ROUND (CAST(value AS numeric) - $1, 2)
	 	WHERE wallet_id = $2 and currency ILIKE $3;
	`)
	if err != nil {

		// Если ошибка от откат
		tx.Rollback()
		r.UpdateStatus(Status_neg, id1)
		r.UpdateStatus(Status_neg, id2)
		log.Println("TransferTO error on first prepare: ", err)
		return err
	}	
	defer stmt.Close()

	// Выполняем вторую команду
	if _, err := stmt.Exec(transf.Sum, transf.WalletID_from, "%" + transf.Currency + "%"); err != nil {

		// Если ошибка от откат
        tx.Rollback() 
		r.UpdateStatus(Status_neg, id1)
		r.UpdateStatus(Status_neg, id2)
		log.Println("TransferTO error on exec of first upd: ", err)
        return err
    }

	// Добавляем деньги на второй счет
	// Готовим вторую команду для добавления средств
    stmt, err = tx.Prepare(`
		UPDATE wallets SET value = ROUND (CAST(value AS numeric) + $1, 2)
		WHERE wallet_id = $2 AND currency ILIKE $3;
	`)
    if err != nil {

		// Если ошибка от откат
        tx.Rollback()
		r.UpdateStatus(Status_neg, id1)
		r.UpdateStatus(Status_neg, id2)
		log.Println("TransferTO error on second prepare: ", err)
        return err
    }
    defer stmt.Close()
	// Выполняем команду добавляем средства
    if _, err := stmt.Exec(transf.Sum, transf.WalletID_to, "%" + transf.Currency + "%"); err != nil {

		// Если ошибка от откат
        tx.Rollback() 
		r.UpdateStatus(Status_neg, id1)
		r.UpdateStatus(Status_neg, id2)
		log.Println("TransferTO error on exec of second upd: ", err)

        return err
    }
	err = r.CreateTransfer(transf.WalletID_from, transf.WalletID_to, strings.ToUpper(transf.Currency), transf.Sum)
	if err != nil {
		// Если ошибка от откат
        tx.Rollback() 
		r.UpdateStatus(Status_neg, id1)
		r.UpdateStatus(Status_neg, id2)
		log.Println("TransferTO error on adding to transfer table: ", err)

        return err
	}
	// Выставялем что транзакции прошли успешно
	r.UpdateStatus(Status_pos, id1)
	r.UpdateStatus(Status_pos, id2)

    return tx.Commit()
}


// Скрипт для удобного обновления статуса транзакции
func (r *TransactionPostgres) UpdateStatus(status string, id int) error {
	st := fmt.Sprintf("UPDATE %s SET status = $1 WHERE id = $2", TransacitonTable)
	_, err := r.db.Exec(st, status, id)
	return err
}

// Скрипт для создания транзакции (по дефолту стоит created)
func (r *TransactionPostgres) CreateTransaction(wallet_id uint64, currency string, sum float64, operationType string) (int, error) {

	// Выполнеям вставку о новой транзакции
	var id int
	st := fmt.Sprintf("INSERT INTO %s (wallet_id, currency, typeOF, sum) VALUES ($1, $2, $3, $4) RETURNING id", TransacitonTable)
	row := r.db.QueryRow(st, wallet_id, strings.ToUpper(currency), operationType, sum)

	// Возвращаем id новой операции
	err := row.Scan(&id)
	
	if err != nil {
		log.Println(err.Error())
	}
	return id, err
}

// Скрипт для записи перевода
func (r TransactionPostgres) CreateTransfer(wallet_id_from, wallet_id_to uint64, currency string, sum float64) (error) {

	// Выполнеям вставку о новом переводе
	st := fmt.Sprintf("INSERT INTO %s (wallet_id_from, wallet_id_to, currency, sum) VALUES ($1, $2, $3, $4)", TransfersTable)
	_, err := r.db.Exec(st, wallet_id_from, wallet_id_to, strings.ToUpper(currency), sum)

	if err != nil {
		log.Println(err.Error())
		return err
	}

	return err
}

// Скрипт для получения баланса по его UID и определенной валюте
func (r *TransactionPostgres) GetBalanceByID(walletID uint64, currency string) (float64, error){

	// Запрашиваем поле суммы на аккаунте, по UID и валюте (без учета регистра)
	var reqBalance float64
	st := fmt.Sprintf("SELECT value FROM %s WHERE wallet_id = $1 and currency ILIKE $2", WalletTable)
	if err := r.db.Get(&reqBalance, st, walletID, "%" + currency + "%"); err != nil {
		fmt.Println("JA tut suka")
		return 0.0, err
	}
	return reqBalance, nil
}

// Скрипт для получения всех валют и номинало по UID кошелька
func (r *TransactionPostgres) GetAllBalancesByID(walletID uint64) ([]bs.WalletCurrency, error){

	// Запрашиваем все строки соотв. UID
	st := fmt.Sprintf("SELECT currency, value FROM %s WHERE wallet_id = $1", WalletTable)
	rows, err := r.db.Query(st, walletID)
	if err != nil {
		log.Printf("error occured while getAllBalance: %s", err.Error())
		return nil, err
	}

	defer rows.Close()

	// Массив для распарсенных данных
	var ans_wallets []bs.WalletCurrency

	for rows.Next() {

		// Временная переменная для распарсенных данных
        var wal bs.WalletCurrency
		// Распарс
        if err := rows.Scan(&wal.Currency, &wal.Value); err != nil {
            return ans_wallets, err
        }

		// Убираем лишние пробелы
		wal.Currency = strings.TrimSpace(wal.Currency)

		// Да.
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
