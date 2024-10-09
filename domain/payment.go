package domain

import "time"

type Payment struct {
	ID        string    `json:"id"`
	UserEmail string    `json:"user_email"`
	Amount    float64   `json:"amount"`
	Currency  string    `json:"currency"` // USDT
	Network   string    `json:"network"`  // TRX, BSC, or TON
	Status    string    `json:"status"`   // pending, completed, failed
	TxHash    string    `json:"tx_hash"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
