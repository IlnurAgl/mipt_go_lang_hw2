package domain

import (
	"context"
	"errors"
)

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
	AddTransaction(transaction *Transaction, ctx context.Context) (int64, error)
	GetTransaction(id int64, ctx context.Context) (*Transaction, error)
	ListTransactions(ctx context.Context) ([]Transaction, error)
	GetAmountTransactionByCategory(category string, ctx context.Context) (float64, error)
}
