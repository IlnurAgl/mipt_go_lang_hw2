package service

import (
	"context"
	"errors"
	"testing"

	"ledger/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockBudgetRepository struct {
	mock.Mock
}

func (m *MockBudgetRepository) SetBudget(budget *domain.Budget, ctx context.Context) error {
	args := m.Called(budget, ctx)
	return args.Error(0)
}

func (m *MockBudgetRepository) GetBudget(category string, ctx context.Context) (*domain.Budget, error) {
	args := m.Called(category, ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Budget), args.Error(1)
}

func (m *MockBudgetRepository) GetBudgets(ctx context.Context) ([]domain.Budget, error) {
	args := m.Called(ctx)
	return args.Get(0).([]domain.Budget), args.Error(1)
}

type MockTransactionRepository struct {
	mock.Mock
}

func (m *MockTransactionRepository) GetAmountTransactionByCategory(category string, ctx context.Context) (float64, error) {
	args := m.Called(category, ctx)
	return args.Get(0).(float64), args.Error(1)
}

func (m *MockTransactionRepository) GetAmountTransactionByCategoryAndMonth(ctx context.Context, category string, date string) (float64, error) {
	args := m.Called(ctx, category, date)
	return args.Get(0).(float64), args.Error(1)
}

func (m *MockTransactionRepository) AddTransaction(transaction *domain.Transaction, ctx context.Context) (int64, error) {
	args := m.Called(transaction, ctx)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockTransactionRepository) GetTransaction(id int64, ctx context.Context) (*domain.Transaction, error) {
	args := m.Called(id, ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Transaction), args.Error(1)
}

func (m *MockTransactionRepository) ListTransactions(ctx context.Context) ([]domain.Transaction, error) {
	args := m.Called(ctx)
	return args.Get(0).([]domain.Transaction), args.Error(1)
}

type MockSummaryRepository struct {
	mock.Mock
}

func (m *MockSummaryRepository) GetSummary(ctx context.Context, from string, to string) (*domain.Summary, error) {
	args := m.Called(ctx, from, to)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Summary), args.Error(1)
}

func TestLedgerServiceImpl_BudgetAdd(t *testing.T) {
	mockBudgetRepo := &MockBudgetRepository{}
	mockTransactionRepo := &MockTransactionRepository{}
	mockSummaryRepo := &MockSummaryRepository{}

	service := NewLedgerService(mockBudgetRepo, mockTransactionRepo, mockSummaryRepo)
	ctx := context.Background()

	tests := []struct {
		name        string
		budget      *domain.Budget
		mockSetup   func()
		expectedErr bool
	}{
		{
			name: "successful budget add",
			budget: &domain.Budget{
				Category: "Food",
				Limit:    1000.0,
			},
			mockSetup: func() {
				mockBudgetRepo.On("SetBudget", mock.AnythingOfType("*domain.Budget"), ctx).Return(nil)
			},
			expectedErr: false,
		},
		{
			name: "invalid budget",
			budget: &domain.Budget{
				Category: "",
				Limit:    1000.0,
			},
			mockSetup:   func() {},
			expectedErr: true,
		},
		{
			name: "repository error",
			budget: &domain.Budget{
				Category: "Food",
				Limit:    1000.0,
			},
			mockSetup: func() {
				mockBudgetRepo.On("SetBudget", mock.AnythingOfType("*domain.Budget"), ctx).Return(errors.New("db error"))
			},
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockBudgetRepo.ExpectedCalls = nil
			tt.mockSetup()

			err := service.BudgetAdd(ctx, tt.budget)

			if tt.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			mockBudgetRepo.AssertExpectations(t)
		})
	}
}

func TestLedgerServiceImpl_BudgetGet(t *testing.T) {
	mockBudgetRepo := &MockBudgetRepository{}
	mockTransactionRepo := &MockTransactionRepository{}
	mockSummaryRepo := &MockSummaryRepository{}

	service := NewLedgerService(mockBudgetRepo, mockTransactionRepo, mockSummaryRepo)
	ctx := context.Background()

	tests := []struct {
		name        string
		category    string
		mockSetup   func()
		expected    *domain.Budget
		expectedErr bool
	}{
		{
			name:     "successful budget get",
			category: "Food",
			mockSetup: func() {
				budget := &domain.Budget{Category: "Food", Limit: 1000.0}
				mockBudgetRepo.On("GetBudget", "Food", ctx).Return(budget, nil)
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
				mockBudgetRepo.On("GetBudget", "NonExistent", ctx).Return(nil, errors.New("not found"))
			},
			expected:    nil,
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockBudgetRepo.ExpectedCalls = nil
			tt.mockSetup()

			result, err := service.BudgetGet(ctx, tt.category)

			if tt.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
			mockBudgetRepo.AssertExpectations(t)
		})
	}
}

func TestLedgerServiceImpl_BudgetsList(t *testing.T) {
	mockBudgetRepo := &MockBudgetRepository{}
	mockTransactionRepo := &MockTransactionRepository{}
	mockSummaryRepo := &MockSummaryRepository{}

	service := NewLedgerService(mockBudgetRepo, mockTransactionRepo, mockSummaryRepo)
	ctx := context.Background()

	tests := []struct {
		name        string
		mockSetup   func()
		expected    map[string]domain.Budget
		expectedErr bool
	}{
		{
			name: "successful budgets list",
			mockSetup: func() {
				budgets := []domain.Budget{
					{Category: "Food", Limit: 1000.0},
					{Category: "Transport", Limit: 500.0},
				}
				mockBudgetRepo.On("GetBudgets", ctx).Return(budgets, nil)
			},
			expected: map[string]domain.Budget{
				"Food":      {Category: "Food", Limit: 1000.0},
				"Transport": {Category: "Transport", Limit: 500.0},
			},
			expectedErr: false,
		},
		{
			name: "repository error",
			mockSetup: func() {
				mockBudgetRepo.On("GetBudgets", ctx).Return([]domain.Budget{}, errors.New("db error"))
			},
			expected:    nil,
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockBudgetRepo.ExpectedCalls = nil
			tt.mockSetup()

			result, err := service.BudgetsList(ctx)

			if tt.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
			mockBudgetRepo.AssertExpectations(t)
		})
	}
}

func TestLedgerServiceImpl_TransactionAdd(t *testing.T) {
	mockBudgetRepo := &MockBudgetRepository{}
	mockTransactionRepo := &MockTransactionRepository{}
	mockSummaryRepo := &MockSummaryRepository{}

	service := NewLedgerService(mockBudgetRepo, mockTransactionRepo, mockSummaryRepo)
	ctx := context.Background()

	tests := []struct {
		name        string
		transaction *domain.Transaction
		mockSetup   func()
		expected    int64
		expectedErr bool
	}{
		{
			name: "successful transaction add",
			transaction: &domain.Transaction{
				Amount:      100.0,
				Category:    "Food",
				Description: "Lunch",
				Date:        "2025-12-01",
			},
			mockSetup: func() {
				budget := &domain.Budget{Category: "Food", Limit: 1000.0}
				mockBudgetRepo.On("GetBudget", "Food", ctx).Return(budget, nil)
				mockTransactionRepo.On("GetAmountTransactionByCategoryAndMonth", ctx, "Food", "2025-12-01").Return(200.0, nil)
				mockTransactionRepo.On("AddTransaction", mock.AnythingOfType("*domain.Transaction"), ctx).Return(int64(1), nil)
			},
			expected:    1,
			expectedErr: false,
		},
		{
			name: "invalid transaction",
			transaction: &domain.Transaction{
				Amount:      -100.0,
				Category:    "Food",
				Description: "Lunch",
				Date:        "2025-12-01",
			},
			mockSetup:   func() {},
			expected:    0,
			expectedErr: true,
		},
		{
			name: "budget exceeded",
			transaction: &domain.Transaction{
				Amount:      900.0,
				Category:    "Food",
				Description: "Lunch",
				Date:        "2025-12-01",
			},
			mockSetup: func() {
				budget := &domain.Budget{Category: "Food", Limit: 1000.0}
				mockBudgetRepo.On("GetBudget", "Food", ctx).Return(budget, nil)
				mockTransactionRepo.On("GetAmountTransactionByCategoryAndMonth", ctx, "Food", "2025-12-01").Return(200.0, nil)
			},
			expected:    0,
			expectedErr: true,
		},
		{
			name: "budget not found",
			transaction: &domain.Transaction{
				Amount:      100.0,
				Category:    "NonExistent",
				Description: "Lunch",
				Date:        "2025-12-01",
			},
			mockSetup: func() {
				mockBudgetRepo.On("GetBudget", "NonExistent", ctx).Return(nil, errors.New("not found"))
			},
			expected:    0,
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockBudgetRepo.ExpectedCalls = nil
			mockTransactionRepo.ExpectedCalls = nil
			tt.mockSetup()

			result, err := service.TransactionAdd(ctx, tt.transaction)

			if tt.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
			mockBudgetRepo.AssertExpectations(t)
			mockTransactionRepo.AssertExpectations(t)
		})
	}
}

func TestLedgerServiceImpl_TransactionGet(t *testing.T) {
	mockBudgetRepo := &MockBudgetRepository{}
	mockTransactionRepo := &MockTransactionRepository{}
	mockSummaryRepo := &MockSummaryRepository{}

	service := NewLedgerService(mockBudgetRepo, mockTransactionRepo, mockSummaryRepo)
	ctx := context.Background()

	tests := []struct {
		name        string
		id          int64
		mockSetup   func()
		expected    *domain.Transaction
		expectedErr bool
	}{
		{
			name: "successful transaction get",
			id:   1,
			mockSetup: func() {
				transaction := &domain.Transaction{
					ID:          1,
					Amount:      100.0,
					Category:    "Food",
					Description: "Lunch",
					Date:        "2025-12-01",
				}
				mockTransactionRepo.On("GetTransaction", int64(1), ctx).Return(transaction, nil)
			},
			expected: &domain.Transaction{
				ID:          1,
				Amount:      100.0,
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
				mockTransactionRepo.On("GetTransaction", int64(999), ctx).Return(nil, errors.New("not found"))
			},
			expected:    nil,
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockTransactionRepo.ExpectedCalls = nil
			tt.mockSetup()

			result, err := service.TransactionGet(ctx, tt.id)

			if tt.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
			mockTransactionRepo.AssertExpectations(t)
		})
	}
}

func TestLedgerServiceImpl_TransactionsList(t *testing.T) {
	mockBudgetRepo := &MockBudgetRepository{}
	mockTransactionRepo := &MockTransactionRepository{}
	mockSummaryRepo := &MockSummaryRepository{}

	service := NewLedgerService(mockBudgetRepo, mockTransactionRepo, mockSummaryRepo)
	ctx := context.Background()

	tests := []struct {
		name        string
		mockSetup   func()
		expected    []domain.Transaction
		expectedErr bool
	}{
		{
			name: "successful transactions list",
			mockSetup: func() {
				transactions := []domain.Transaction{
					{
						ID:          1,
						Amount:      100.0,
						Category:    "Food",
						Description: "Lunch",
						Date:        "2025-12-01",
					},
					{
						ID:          2,
						Amount:      50.0,
						Category:    "Transport",
						Description: "Bus",
						Date:        "2025-12-02",
					},
				}
				mockTransactionRepo.On("ListTransactions", ctx).Return(transactions, nil)
			},
			expected: []domain.Transaction{
				{
					ID:          1,
					Amount:      100.0,
					Category:    "Food",
					Description: "Lunch",
					Date:        "2025-12-01",
				},
				{
					ID:          2,
					Amount:      50.0,
					Category:    "Transport",
					Description: "Bus",
					Date:        "2025-12-02",
				},
			},
			expectedErr: false,
		},
		{
			name: "repository error",
			mockSetup: func() {
				mockTransactionRepo.On("ListTransactions", ctx).Return([]domain.Transaction{}, errors.New("db error"))
			},
			expected:    nil,
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockTransactionRepo.ExpectedCalls = nil
			tt.mockSetup()

			result, err := service.TransactionsList(ctx)

			if tt.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
			mockTransactionRepo.AssertExpectations(t)
		})
	}
}

func TestLedgerServiceImpl_GetReportSummary(t *testing.T) {
	mockBudgetRepo := &MockBudgetRepository{}
	mockTransactionRepo := &MockTransactionRepository{}
	mockSummaryRepo := &MockSummaryRepository{}

	service := NewLedgerService(mockBudgetRepo, mockTransactionRepo, mockSummaryRepo)
	ctx := context.Background()

	tests := []struct {
		name        string
		from        string
		to          string
		mockSetup   func()
		expected    *domain.Summary
		expectedErr bool
	}{
		{
			name: "successful summary get",
			from: "2025-12-01",
			to:   "2025-12-31",
			mockSetup: func() {
				summary := &domain.Summary{
					Categories:  map[string]float64{"Food": 500.0, "Transport": 200.0},
					CacheResult: false,
				}
				mockSummaryRepo.On("GetSummary", ctx, "2025-12-01", "2025-12-31").Return(summary, nil)
			},
			expected: &domain.Summary{
				Categories:  map[string]float64{"Food": 500.0, "Transport": 200.0},
				CacheResult: false,
			},
			expectedErr: false,
		},
		{
			name: "repository error",
			from: "2025-12-01",
			to:   "2025-12-31",
			mockSetup: func() {
				mockSummaryRepo.On("GetSummary", ctx, "2025-12-01", "2025-12-31").Return(nil, errors.New("db error"))
			},
			expected:    nil,
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSummaryRepo.ExpectedCalls = nil
			tt.mockSetup()

			result, err := service.GetReportSummary(ctx, tt.from, tt.to)

			if tt.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
			mockSummaryRepo.AssertExpectations(t)
		})
	}
}
