package internal

import (
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

func CreateTransaction(s ledger.LedgerService, r CreateTransactionRequest) (*TransactionResponse, error) {
	id, err := s.AddTransaction(r.Amount, r.Category, r.Description, r.Date)
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

func CreateBudget(s ledger.LedgerService, r CreateBudgetRequest) (*BudgetResponse, error) {
	err := s.SetBudget(
		r.Category,
		r.Limit,
	)
	if err != nil {
		return nil, err
	}
	return &BudgetResponse{Category: r.Category, Limit: r.Limit}, nil
}
