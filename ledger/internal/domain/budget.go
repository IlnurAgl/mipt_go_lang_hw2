package domain

import (
	"context"
	"errors"
)

type Budget struct {
	Category string  `json:"category"`
	Limit    float64 `json:"limit"`
}

func (budget *Budget) Validate() error {
	if budget.Limit <= 0 {
		return errors.New("invalid limit")
	}
	if budget.Category == "" {
		return errors.New("invalid category")
	}
	return nil
}

type BudgetRepository interface {
	SetBudget(budget *Budget, ctx context.Context) error
	GetBudgets(ctx context.Context) ([]Budget, error)
	GetBudget(category string, ctx context.Context) (*Budget, error)
}
