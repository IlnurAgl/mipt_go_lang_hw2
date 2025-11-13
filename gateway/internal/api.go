package internal

import (
	"errors"
	"ledger"
)

type CreateTransactionRequest struct {
	Amount      float64 `json:"amount"`
	Category    string  `json:"category"`
	Description string  `json:"description"`
	Date        string  `json:"date"`
}

type TransactionResponse struct {
	ID          int64   `json:"id"`
	Amount      float64 `json:"amount"`
	Category    string  `json:"category"`
	Description string  `json:"description"`
	Date        string  `json:"date"`
}

type CreateBudgetRequest struct {
	Category string
	Limit    float64
}

type BudgetResponse struct {
	Category string
	Limit    float64
}

func CreateTransaction(r CreateTransactionRequest) (*TransactionResponse, error) {
	tr := ledger.Transaction{
		Amount:      r.Amount,
		Category:    r.Category,
		Description: r.Description,
		Date:        r.Date,
	}
	err := tr.Validate()
	if err != nil {
		return nil, errors.New("invalid transaction")
	}
	id, err := ledger.AddTransaction(tr)
	if err != nil {
		return nil, err
	}
	return &TransactionResponse{
		ID:          id,
		Amount:      r.Amount,
		Category:    r.Category,
		Description: r.Description,
		Date:        r.Date,
	}, nil
}

func CreateBudget(r CreateBudgetRequest) (*BudgetResponse, error) {
	err := ledger.SetBudget(ledger.Budget{
		Category: r.Category,
		Limit:    r.Limit,
	})
	if err != nil {
		return nil, err
	}
	return &BudgetResponse{Category: r.Category, Limit: r.Limit}, nil
}
