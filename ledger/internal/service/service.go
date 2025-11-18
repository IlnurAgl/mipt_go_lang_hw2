package service

import (
	"errors"
	"ledger/internal/domain"
)

type LedgerServiceImpl struct {
	BudgetRepository      domain.BudgetRepository
	TransactionRepository domain.TransactionRepository
}

func (l *LedgerServiceImpl) SetBudget(category string, limit float64) error {
	budget := domain.Budget{
		Category: category,
		Limit:    limit,
	}
	err := budget.Validate()
	if err != nil {
		return err
	}
	err = l.BudgetRepository.SetBudget(&budget)
	if err != nil {
		return err
	}
	return nil
}

func (l *LedgerServiceImpl) GetBudgets() (map[string]domain.Budget, error) {
	budgets, err := l.BudgetRepository.GetBudgets()
	if err != nil {
		return nil, err
	}
	budgetMap := make(map[string]domain.Budget)
	for _, budget := range budgets {
		budgetMap[budget.Category] = budget
	}
	return budgetMap, nil
}

func (l *LedgerServiceImpl) AddTransaction(Amount float64, Category string, Description string, Date string) (int64, error) {
	transaction := domain.Transaction{
		Amount:      Amount,
		Category:    Category,
		Description: Description,
		Date:        Date,
	}
	err := transaction.Validate()
	if err != nil {
		return 0, errors.New("invalid transaction")
	}
	budget, err := l.BudgetRepository.GetBudget(transaction.Category)
	if err != nil {
		return 0, err
	}
	amount, err := l.TransactionRepository.GetAmountTransactionByCategory(transaction.Category)
	if err != nil {
		return 0, err
	}
	if amount+transaction.Amount > budget.Limit {
		return 0, errors.New("budget exceeded")
	}
	id, err := l.TransactionRepository.AddTransaction(&transaction)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (l *LedgerServiceImpl) ListTransactions() ([]domain.Transaction, error) {
	transactions, err := l.TransactionRepository.ListTransactions()
	if err != nil {
		return nil, err
	}
	return transactions, nil
}
