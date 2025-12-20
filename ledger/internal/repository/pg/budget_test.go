package pg

import (
	"context"
	"database/sql"
	"testing"

	"ledger/internal/domain"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBudgetPgRepository_SetBudget(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewBudgetPgRepository(db)
	ctx := context.Background()
	budget := &domain.Budget{
		Category: "Food",
		Limit:    1000.0,
	}

	tests := []struct {
		name        string
		budget      *domain.Budget
		mockSetup   func()
		expectedErr bool
	}{
		{
			name:   "successful insert",
			budget: budget,
			mockSetup: func() {
				mock.ExpectExec(`INSERT INTO budgets\(category, limit_amount\) VALUES\(\$1,\$2\) ON CONFLICT\(category\) DO UPDATE SET limit_amount =EXCLUDED\.limit_amount`).
					WithArgs("Food", 1000.0).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			expectedErr: false,
		},
		{
			name:   "database error",
			budget: budget,
			mockSetup: func() {
				mock.ExpectExec(`INSERT INTO budgets\(category, limit_amount\) VALUES\(\$1,\$2\) ON CONFLICT\(category\) DO UPDATE SET limit_amount =EXCLUDED\.limit_amount`).
					WithArgs("Food", 1000.0).
					WillReturnError(sql.ErrConnDone)
			},
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			err := repo.SetBudget(tt.budget, ctx)

			if tt.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestBudgetPgRepository_GetBudgets(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewBudgetPgRepository(db)
	ctx := context.Background()

	tests := []struct {
		name        string
		mockSetup   func()
		expected    []domain.Budget
		expectedErr bool
	}{
		{
			name: "successful get budgets",
			mockSetup: func() {
				rows := sqlmock.NewRows([]string{"category", "limit_amount"}).
					AddRow("Food", 1000.0).
					AddRow("Transport", 500.0)
				mock.ExpectQuery(`SELECT category, limit_amount FROM budgets ORDER BY category`).
					WillReturnRows(rows)
			},
			expected: []domain.Budget{
				{Category: "Food", Limit: 1000.0},
				{Category: "Transport", Limit: 500.0},
			},
			expectedErr: false,
		},
		{
			name: "empty result",
			mockSetup: func() {
				rows := sqlmock.NewRows([]string{"category", "limit_amount"})
				mock.ExpectQuery(`SELECT category, limit_amount FROM budgets ORDER BY category`).
					WillReturnRows(rows)
			},
			expected:    nil,
			expectedErr: false,
		},
		{
			name: "database error",
			mockSetup: func() {
				mock.ExpectQuery(`SELECT category, limit_amount FROM budgets ORDER BY category`).
					WillReturnError(sql.ErrConnDone)
			},
			expected:    nil,
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			result, err := repo.GetBudgets(ctx)

			if tt.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestBudgetPgRepository_GetBudget(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewBudgetPgRepository(db)
	ctx := context.Background()

	tests := []struct {
		name        string
		category    string
		mockSetup   func()
		expected    *domain.Budget
		expectedErr bool
	}{
		{
			name:     "successful get budget",
			category: "Food",
			mockSetup: func() {
				rows := sqlmock.NewRows([]string{"category", "limit_amount"}).
					AddRow("Food", 1000.0)
				mock.ExpectQuery(`SELECT category, limit_amount FROM budgets WHERE category = \$1`).
					WithArgs("Food").
					WillReturnRows(rows)
			},
			expected: &domain.Budget{
				Category: "Food",
				Limit:    1000.0,
			},
			expectedErr: false,
		},
		{
			name:     "budget not found",
			category: "NonExistent",
			mockSetup: func() {
				mock.ExpectQuery(`SELECT category, limit_amount FROM budgets WHERE category = \$1`).
					WithArgs("NonExistent").
					WillReturnError(sql.ErrNoRows)
			},
			expected: &domain.Budget{
				Category: "",
				Limit:    0,
			},
			expectedErr: true,
		},
		{
			name:     "database error",
			category: "Food",
			mockSetup: func() {
				mock.ExpectQuery(`SELECT category, limit_amount FROM budgets WHERE category = \$1`).
					WithArgs("Food").
					WillReturnError(sql.ErrConnDone)
			},
			expected: &domain.Budget{
				Category: "",
				Limit:    0,
			},
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			result, err := repo.GetBudget(tt.category, ctx)

			if tt.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
