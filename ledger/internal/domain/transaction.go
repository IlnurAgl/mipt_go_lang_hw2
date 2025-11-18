package domain

import "errors"

type Transaction struct {
	ID          int64
	Amount      float64
	Category    string
	Description string
	Date        string
}

func (tx *Transaction) Validate() error {
	if tx.Amount <= 0 {
		return errors.New("invalid amount")
	}
	if tx.Category == "" {
		return errors.New("invalid category")
	}
	return nil
}

type TransactionRepository interface {
	AddTransaction(transaction *Transaction) (int64, error)
	ListTransactions() ([]Transaction, error)
	GetAmountTransactionByCategory(category string) (float64, error)
}
