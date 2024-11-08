package model

import "time"

type Transaction struct {
    TransactionID         int64      `json:"transaction_id" db:"transaction_id" gorm:"primaryKey;autoIncrement"`
    TransactionCategoryID *int64     `json:"transaction_category_id,omitempty" db:"transaction_category_id"`
    AccountID             int64      `json:"account_id" db:"account_id"`
    FromAccountID         *int64     `json:"from_account_id,omitempty" db:"from_account_id"`
    ToAccountID           *int64     `json:"to_account_id,omitempty" db:"to_account_id"`
    Amount                int64      `json:"amount" db:"amount"`
    TransactionDate       time.Time  `json:"transaction_date" db:"transaction_date"`
}

func (Transaction) TableName() string {
    return "transaction"
}