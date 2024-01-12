package bs

type Request struct {
	WalletID uint64  `json:"wallet_id" db:"wallet_id"`
	Currency   string  `json:"currency" db:"currency"`
	Sum        float64 `json:"sum" db:"sum"`
}

type Answer struct {
	WalletID uint64  `json:"wallet_id"`
	Usdt       float64 `json:"USDT"`
	Rub        float64 `json:"RUS"`
	Eur        float64 `json:"EUR"`
}
