package repository

import (
	"fmt"

	"github.com/jmoiron/sqlx"
)

const (
	TransacitonTable = "Transactions"
	WalletTable      = "Wallets"
	TransfersTable = "Transfers"
)

type Config struct {
	Host     string
	Port     string
	Username string
	Password string
	DBName   string
	SSLMode  string
}

func NewPostgresDB(cfg Config) (*sqlx.DB, error) {
	db, err := sqlx.Open("postgres", fmt.Sprintf("host=%s port=%s user=%s dbname =%s password=%s sslmode=%s", cfg.Host, cfg.Port, cfg.Username, cfg.DBName, cfg.Password, cfg.SSLMode))
	if err != nil {
		return nil, err
	}

	er := db.Ping()
	if er != nil {
		return nil, er
	}

	return db, nil
}
