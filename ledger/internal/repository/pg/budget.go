package pg

import (
	"context"
	"database/sql"
)
import "ledger/internal/domain"

type BudgetPgRepository struct {
	db *sql.DB
}

func NewBudgetPgRepository(db *sql.DB) *BudgetPgRepository {
	return &BudgetPgRepository{
		db: db,
	}
}

func (r *BudgetPgRepository) SetBudget(b *domain.Budget, ctx context.Context) error {
	_, err := r.db.ExecContext(ctx, "INSERT INTO budgets(category, limit_amount) VALUES($1,$2) ON CONFLICT(category) DO UPDATE SET limit_amount =EXCLUDED.limit_amount", &b.Category, &b.Limit)
	return err
}

func (r *BudgetPgRepository) GetBudgets(ctx context.Context) ([]domain.Budget, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT category, limit_amount FROM budgets ORDER BY category")
	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			println(err.Error())
		}
	}(rows)
	var dbBudgets []domain.Budget
	for rows.Next() {
		var b domain.Budget
		if err := rows.Scan(&b.Category, &b.Limit); err != nil {
			return dbBudgets, err
		}
		dbBudgets = append(dbBudgets, b)
	}
	if err = rows.Err(); err != nil {
		return dbBudgets, err
	}
	return dbBudgets, nil
}

func (r *BudgetPgRepository) GetBudget(category string, ctx context.Context) (*domain.Budget, error) {
	var budget domain.Budget
	err := r.db.QueryRowContext(ctx, "SELECT category, limit_amount FROM budgets WHERE category = $1", category).Scan(&budget.Category, &budget.Limit)
	if err != nil {
		return &budget, err
	}
	return &budget, nil
}
