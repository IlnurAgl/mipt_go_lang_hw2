package service

import (
	"context"
	"errors"
	"ledger/internal/domain"
)

type LedgerServiceImpl struct {
	BudgetRepository      domain.BudgetRepository
	TransactionRepository domain.TransactionRepository
	SummaryRepository     domain.SummaryRepository
}

func (l *LedgerServiceImpl) SetBudget(category string, limit float64, ctx context.Context) error {
	budget := domain.Budget{
		Category: category,
		Limit:    limit,
	}
	err := budget.Validate()
	if err != nil {
		return err
	}
	err = l.BudgetRepository.SetBudget(&budget, ctx)
	if err != nil {
		return err
	}
	return nil
}

func (l *LedgerServiceImpl) GetBudgets(ctx context.Context) (map[string]domain.Budget, error) {
	budgets, err := l.BudgetRepository.GetBudgets(ctx)
	if err != nil {
		return nil, err
	}
	budgetMap := make(map[string]domain.Budget)
	for _, budget := range budgets {
		budgetMap[budget.Category] = budget
	}
	return budgetMap, nil
}

func (l *LedgerServiceImpl) AddTransaction(Amount float64, Category string, Description string, Date string, ctx context.Context) (int64, error) {
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
	budget, err := l.BudgetRepository.GetBudget(transaction.Category, ctx)
	if err != nil {
		return 0, err
	}
	amount, err := l.TransactionRepository.GetAmountTransactionByCategory(transaction.Category, ctx)
	if err != nil {
		return 0, err
	}
	if amount+transaction.Amount > budget.Limit {
		return 0, errors.New("budget exceeded")
	}
	id, err := l.TransactionRepository.AddTransaction(&transaction, ctx)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (l *LedgerServiceImpl) ListTransactions(ctx context.Context) ([]domain.Transaction, error) {
	transactions, err := l.TransactionRepository.ListTransactions(ctx)
	if err != nil {
		return nil, err
	}
	return transactions, nil
}

func (l *LedgerServiceImpl) GetReportSummary(from string, to string, ctx context.Context) (*domain.Summary, error) {
	return l.SummaryRepository.GetSummary(from, to, ctx)
}
