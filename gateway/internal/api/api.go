package api

import (
	"context"
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

func CreateTransaction(s ledger.LedgerService, r CreateTransactionRequest, ctx context.Context) (*TransactionResponse, error) {
	id, err := s.AddTransaction(r.Amount, r.Category, r.Description, r.Date, ctx)
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

func CreateBudget(s ledger.LedgerService, r CreateBudgetRequest, ctx context.Context) (*BudgetResponse, error) {
	err := s.SetBudget(
		r.Category,
		r.Limit,
		ctx,
	)
	if err != nil {
		return nil, err
	}
	return &BudgetResponse{Category: r.Category, Limit: r.Limit}, nil
}

type BulkTransactionsResult struct {
	Accepted int64          `json:"accepted"`
	Rejected int64          `json:"rejected"`
	Errors   map[int]string `json:"errors"`
}

func BulkCreateTransaction(s ledger.LedgerService, transatcions []CreateTransactionRequest, ctx context.Context) (BulkTransactionsResult, error) {
	trs := make([]ledger.Transaction, 0)
	for _, transatcion := range transatcions {
		trs = append(trs, ledger.Transaction{
			Amount:      transatcion.Amount,
			Category:    transatcion.Category,
			Date:        transatcion.Date,
			Description: transatcion.Description,
		})
	}
	result, err := s.BulkAddTransactions(trs, 4, ctx)
	return BulkTransactionsResult{result.Accepted, result.Rejected, result.Errors}, err
}
