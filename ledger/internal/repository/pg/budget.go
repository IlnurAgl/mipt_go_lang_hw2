package pg

import "database/sql"
import "ledger/internal/domain"

type BudgetPgRepository struct {
	db *sql.DB
}

func NewBudgetPgRepository(db *sql.DB) *BudgetPgRepository {
	return &BudgetPgRepository{
		db: db,
	}
}

func (r *BudgetPgRepository) SetBudget(b *domain.Budget) error {
	_, err := r.db.Exec("INSERT INTO budgets(category, limit_amount) VALUES($1,$2) ON CONFLICT(category) DO UPDATE SET limit_amount =EXCLUDED.limit_amount", &b.Category, &b.Limit)
	return err
}

func (r *BudgetPgRepository) GetBudgets() ([]domain.Budget, error) {
	rows, err := r.db.Query("SELECT category, limit_amount FROM budgets ORDER BY category")
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

func (r *BudgetPgRepository) GetBudget(category string) (*domain.Budget, error) {
	var budget domain.Budget
	err := r.db.QueryRow("SELECT category, limit_amount FROM budgets WHERE category = $1", category).Scan(&budget.Category, &budget.Limit)
	if err != nil {
		return &budget, err
	}
	return &budget, nil
}
