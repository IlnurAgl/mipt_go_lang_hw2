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

func TestTransactionPgRepository_GetAmountTransactionByCategory(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewTransactionPgRepository(db)
	ctx := context.Background()

	tests := []struct {
		name        string
		category    string
		mockSetup   func()
		expected    float64
		expectedErr bool
	}{
		{
			name:     "successful get amount",
			category: "Food",
			mockSetup: func() {
				rows := sqlmock.NewRows([]string{"coalesce"}).AddRow(150.50)
				mock.ExpectQuery(`SELECT COALESCE\(SUM\(amount\),0\) FROM expenses WHERE category=\$1`).
					WithArgs("Food").
					WillReturnRows(rows)
			},
			expected:    150.50,
			expectedErr: false,
		},
		{
			name:     "no transactions",
			category: "Empty",
			mockSetup: func() {
				rows := sqlmock.NewRows([]string{"coalesce"}).AddRow(0.0)
				mock.ExpectQuery(`SELECT COALESCE\(SUM\(amount\),0\) FROM expenses WHERE category=\$1`).
					WithArgs("Empty").
					WillReturnRows(rows)
			},
			expected:    0.0,
			expectedErr: false,
		},
		{
			name:     "database error",
			category: "Food",
			mockSetup: func() {
				mock.ExpectQuery(`SELECT COALESCE\(SUM\(amount\),0\) FROM expenses WHERE category=\$1`).
					WithArgs("Food").
					WillReturnError(sql.ErrConnDone)
			},
			expected:    0,
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			result, err := repo.GetAmountTransactionByCategory(tt.category, ctx)

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

func TestTransactionPgRepository_GetAmountTransactionByCategoryAndMonth(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewTransactionPgRepository(db)
	ctx := context.Background()

	tests := []struct {
		name        string
		category    string
		date        string
		mockSetup   func()
		expected    float64
		expectedErr bool
	}{
		{
			name:     "successful get amount",
			category: "Food",
			date:     "2025-12-15",
			mockSetup: func() {
				rows := sqlmock.NewRows([]string{"coalesce"}).AddRow(200.75)
				mock.ExpectQuery(`SELECT COALESCE\(SUM\(amount\),0\) FROM expenses WHERE category=\$1 AND date between \$2 and \$3`).
					WithArgs("Food", "2025-12-01", "2026-01-01").
					WillReturnRows(rows)
			},
			expected:    200.75,
			expectedErr: false,
		},
		{
			name:        "invalid date",
			category:    "Food",
			date:        "invalid-date",
			mockSetup:   func() {},
			expected:    0,
			expectedErr: true,
		},
		{
			name:     "database error",
			category: "Food",
			date:     "2025-12-15",
			mockSetup: func() {
				mock.ExpectQuery(`SELECT COALESCE\(SUM\(amount\),0\) FROM expenses WHERE category=\$1 AND date between \$2 and \$3`).
					WithArgs("Food", "2025-12-01", "2026-01-01").
					WillReturnError(sql.ErrConnDone)
			},
			expected:    0,
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			result, err := repo.GetAmountTransactionByCategoryAndMonth(ctx, tt.category, tt.date)

			if tt.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
			if tt.name != "invalid date" {
				assert.NoError(t, mock.ExpectationsWereMet())
			}
		})
	}
}

func TestTransactionPgRepository_AddTransaction(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewTransactionPgRepository(db)
	ctx := context.Background()

	transaction := &domain.Transaction{
		Amount:      100.50,
		Category:    "Food",
		Description: "Lunch",
		Date:        "2025-12-01",
	}

	tests := []struct {
		name        string
		transaction *domain.Transaction
		mockSetup   func()
		expected    int64
		expectedErr bool
	}{
		{
			name:        "successful add",
			transaction: transaction,
			mockSetup: func() {
				mock.ExpectQuery(`INSERT INTO expenses\(amount, category, description, date\) VALUES\(\$1,\$2,\$3,\$4\) RETURNING id`).
					WithArgs(100.50, "Food", "Lunch", "2025-12-01").
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
			},
			expected:    1,
			expectedErr: false,
		},
		{
			name:        "database error",
			transaction: transaction,
			mockSetup: func() {
				mock.ExpectQuery(`INSERT INTO expenses\(amount, category, description, date\) VALUES\(\$1,\$2,\$3,\$4\) RETURNING id`).
					WithArgs(100.50, "Food", "Lunch", "2025-12-01").
					WillReturnError(sql.ErrConnDone)
			},
			expected:    0,
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			result, err := repo.AddTransaction(tt.transaction, ctx)

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

func TestTransactionPgRepository_GetTransaction(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewTransactionPgRepository(db)
	ctx := context.Background()

	tests := []struct {
		name        string
		id          int64
		mockSetup   func()
		expected    *domain.Transaction
		expectedErr bool
	}{
		{
			name: "successful get",
			id:   1,
			mockSetup: func() {
				rows := sqlmock.NewRows([]string{"id", "amount", "category", "description", "date"}).
					AddRow(1, 100.50, "Food", "Lunch", "2025-12-01")
				mock.ExpectQuery(`SELECT id, amount, category, description, date FROM expenses where id=\$1`).
					WithArgs(1).
					WillReturnRows(rows)
			},
			expected: &domain.Transaction{
				ID:          1,
				Amount:      100.50,
				Category:    "Food",
				Description: "Lunch",
				Date:        "2025-12-01",
			},
			expectedErr: false,
		},
		{
			name: "transaction not found",
			id:   999,
			mockSetup: func() {
				mock.ExpectQuery(`SELECT id, amount, category, description, date FROM expenses where id=\$1`).
					WithArgs(999).
					WillReturnError(sql.ErrNoRows)
			},
			expected:    nil,
			expectedErr: true,
		},
		{
			name: "database error",
			id:   1,
			mockSetup: func() {
				mock.ExpectQuery(`SELECT id, amount, category, description, date FROM expenses where id=\$1`).
					WithArgs(1).
					WillReturnError(sql.ErrConnDone)
			},
			expected:    nil,
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			result, err := repo.GetTransaction(tt.id, ctx)

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

func TestTransactionPgRepository_ListTransactions(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewTransactionPgRepository(db)
	ctx := context.Background()

	tests := []struct {
		name        string
		mockSetup   func()
		expected    []domain.Transaction
		expectedErr bool
	}{
		{
			name: "successful list",
			mockSetup: func() {
				rows := sqlmock.NewRows([]string{"id", "amount", "category", "description", "date"}).
					AddRow(1, 100.50, "Food", "Lunch", "2025-12-01").
					AddRow(2, 50.25, "Transport", "Bus", "2025-12-02")
				mock.ExpectQuery(`SELECT id, amount, category, description, date FROM expenses ORDER BY date DESC, id DESC`).
					WillReturnRows(rows)
			},
			expected: []domain.Transaction{
				{
					ID:          1,
					Amount:      100.50,
					Category:    "Food",
					Description: "Lunch",
					Date:        "2025-12-01",
				},
				{
					ID:          2,
					Amount:      50.25,
					Category:    "Transport",
					Description: "Bus",
					Date:        "2025-12-02",
				},
			},
			expectedErr: false,
		},
		{
			name: "empty result",
			mockSetup: func() {
				rows := sqlmock.NewRows([]string{"id", "amount", "category", "description", "date"})
				mock.ExpectQuery(`SELECT id, amount, category, description, date FROM expenses ORDER BY date DESC, id DESC`).
					WillReturnRows(rows)
			},
			expected:    nil,
			expectedErr: false,
		},
		{
			name: "database error",
			mockSetup: func() {
				mock.ExpectQuery(`SELECT id, amount, category, description, date FROM expenses ORDER BY date DESC, id DESC`).
					WillReturnError(sql.ErrConnDone)
			},
			expected:    nil,
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			result, err := repo.ListTransactions(ctx)

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
