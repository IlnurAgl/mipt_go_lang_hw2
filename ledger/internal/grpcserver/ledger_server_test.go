package grpcserver

import (
	"context"
	"testing"

	"ledger/internal/domain"
	"ledger/internal/service"

	pb "ledger/internal/pb/ledger/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type MockLedgerServiceImpl struct {
	mock.Mock
}

var _ service.LedgerService = (*MockLedgerServiceImpl)(nil)

func (m *MockLedgerServiceImpl) BudgetAdd(ctx context.Context, budget *domain.Budget) error {
	args := m.Called(ctx, budget)
	return args.Error(0)
}

func (m *MockLedgerServiceImpl) BudgetGet(ctx context.Context, category string) (*domain.Budget, error) {
	args := m.Called(ctx, category)
	return args.Get(0).(*domain.Budget), args.Error(1)
}

func (m *MockLedgerServiceImpl) BudgetsList(ctx context.Context) (map[string]domain.Budget, error) {
	args := m.Called(ctx)
	return args.Get(0).(map[string]domain.Budget), args.Error(1)
}

func (m *MockLedgerServiceImpl) TransactionAdd(ctx context.Context, transaction *domain.Transaction) (int64, error) {
	args := m.Called(ctx, transaction)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockLedgerServiceImpl) TransactionGet(ctx context.Context, id int64) (*domain.Transaction, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*domain.Transaction), args.Error(1)
}

func (m *MockLedgerServiceImpl) TransactionsList(ctx context.Context) ([]domain.Transaction, error) {
	args := m.Called(ctx)
	return args.Get(0).([]domain.Transaction), args.Error(1)
}

func (m *MockLedgerServiceImpl) BulkAddTransactions(ctx context.Context, transactions []domain.Transaction, numWorkers int) (*service.BulkTransactionResult, error) {
	args := m.Called(ctx, transactions, numWorkers)
	return args.Get(0).(*service.BulkTransactionResult), args.Error(1)
}

func (m *MockLedgerServiceImpl) GetReportSummary(ctx context.Context, from string, to string) (*domain.Summary, error) {
	args := m.Called(ctx, from, to)
	return args.Get(0).(*domain.Summary), args.Error(1)
}

func TestLedgerServer_BudgetAdd(t *testing.T) {
	mockService := &MockLedgerServiceImpl{}
	server := NewLedgerServer(mockService)

	tests := []struct {
		name        string
		req         *pb.BudgetAddRequest
		mockSetup   func(*MockLedgerServiceImpl)
		expectedErr error
	}{
		{
			name: "service error",
			req: &pb.BudgetAddRequest{
				Category: "Error",
				Limit:    1000.0,
			},
			mockSetup: func(m *MockLedgerServiceImpl) {
				m.On("BudgetAdd", mock.Anything, mock.MatchedBy(func(b *domain.Budget) bool {
					return b.Category == "Error"
				})).Return(assert.AnError)
			},
			expectedErr: status.Errorf(codes.Internal, "set budget: %v", assert.AnError),
		},
		{
			name: "successful budget add",
			req: &pb.BudgetAddRequest{
				Category: "Food",
				Limit:    1000.0,
			},
			mockSetup: func(m *MockLedgerServiceImpl) {
				m.On("BudgetAdd", mock.Anything, mock.MatchedBy(func(b *domain.Budget) bool {
					return b.Category == "Food" && b.Limit == 1000.0
				})).Return(nil)
			},
			expectedErr: nil,
		},
		{
			name: "empty category",
			req: &pb.BudgetAddRequest{
				Category: "",
				Limit:    1000.0,
			},
			mockSetup:   func(m *MockLedgerServiceImpl) {},
			expectedErr: status.Error(codes.InvalidArgument, "category is required"),
		},
		{
			name: "zero limit",
			req: &pb.BudgetAddRequest{
				Category: "Food",
				Limit:    0,
			},
			mockSetup:   func(m *MockLedgerServiceImpl) {},
			expectedErr: status.Error(codes.InvalidArgument, "limit is required"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup(mockService)

			_, err := server.BudgetAdd(context.Background(), tt.req)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}
			mockService.AssertExpectations(t)
		})
	}
}

func TestLedgerServer_BudgetGet(t *testing.T) {
	mockService := &MockLedgerServiceImpl{}
	server := NewLedgerServer(mockService)

	tests := []struct {
		name         string
		req          *pb.BudgetGetRequest
		mockSetup    func(*MockLedgerServiceImpl)
		expectedResp *pb.BudgetGetResponse
		expectedErr  error
	}{
		{
			name: "successful budget get",
			req: &pb.BudgetGetRequest{
				Category: "Food",
			},
			mockSetup: func(m *MockLedgerServiceImpl) {
				m.On("BudgetGet", mock.Anything, "Food").Return(&domain.Budget{
					Category: "Food",
					Limit:    1000.0,
				}, nil)
			},
			expectedResp: &pb.BudgetGetResponse{
				Category: "Food",
				Limit:    1000.0,
			},
			expectedErr: nil,
		},
		{
			name: "empty category",
			req: &pb.BudgetGetRequest{
				Category: "",
			},
			mockSetup:    func(m *MockLedgerServiceImpl) {},
			expectedResp: nil,
			expectedErr:  status.Error(codes.InvalidArgument, "category is required"),
		},
		{
			name: "service error",
			req: &pb.BudgetGetRequest{
				Category: "Error",
			},
			mockSetup: func(m *MockLedgerServiceImpl) {
				m.On("BudgetGet", mock.Anything, "Error").Return((*domain.Budget)(nil), assert.AnError)
			},
			expectedResp: nil,
			expectedErr:  status.Errorf(codes.Internal, "get budget: %v", assert.AnError),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup(mockService)

			resp, err := server.BudgetGet(context.Background(), tt.req)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErr.Error(), err.Error())
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResp, resp)
			}
			mockService.AssertExpectations(t)
		})
	}
}

func TestLedgerServer_BudgetsList(t *testing.T) {
	mockService := &MockLedgerServiceImpl{}
	server := NewLedgerServer(mockService)

	tests := []struct {
		name         string
		req          *emptypb.Empty
		mockSetup    func(*MockLedgerServiceImpl)
		expectedResp *pb.BudgetGetListResponse
		expectedErr  error
	}{
		{
			name: "successful budgets list",
			req:  &emptypb.Empty{},
			mockSetup: func(m *MockLedgerServiceImpl) {
				m.On("BudgetsList", mock.Anything).Return(map[string]domain.Budget{
					"Food":      {Category: "Food", Limit: 1000.0},
					"Transport": {Category: "Transport", Limit: 500.0},
				}, nil)
			},
			expectedResp: &pb.BudgetGetListResponse{
				Budgets: []*pb.BudgetGetResponse{
					{Category: "Food", Limit: 1000.0},
					{Category: "Transport", Limit: 500.0},
				},
			},
			expectedErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup(mockService)

			resp, err := server.BudgetsList(context.Background(), tt.req)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErr.Error(), err.Error())
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.Len(t, resp.Budgets, len(tt.expectedResp.Budgets))
				for _, expected := range tt.expectedResp.Budgets {
					found := false
					for _, actual := range resp.Budgets {
						if actual.Category == expected.Category && actual.Limit == expected.Limit {
							found = true
							break
						}
					}
					assert.True(t, found, "expected budget not found: %+v", expected)
				}
			}
			mockService.AssertExpectations(t)
		})
	}
}

func TestLedgerServer_TransactionAdd(t *testing.T) {
	mockService := &MockLedgerServiceImpl{}
	server := NewLedgerServer(mockService)

	req := &pb.TransactionAddRequest{
		Amount:      100.0,
		Category:    "Food",
		Description: "Lunch",
		Date:        "2025-12-01",
	}
	mockService.On("TransactionAdd", mock.Anything, mock.MatchedBy(func(tr *domain.Transaction) bool {
		return tr.Amount == 100.0 && tr.Category == "Food"
	})).Return(int64(1), nil)

	resp, err := server.TransactionAdd(context.Background(), req)

	assert.NoError(t, err)
	assert.Equal(t, &pb.TransactionAddResponse{Id: 1}, resp)
	mockService.AssertExpectations(t)
}

func TestLedgerServer_TransactionGet(t *testing.T) {
	mockService := &MockLedgerServiceImpl{}
	server := NewLedgerServer(mockService)

	req := &pb.TransactionGetRequest{Id: 1}
	mockService.On("TransactionGet", mock.Anything, int64(1)).Return(&domain.Transaction{
		ID:          1,
		Amount:      100.0,
		Category:    "Food",
		Description: "Lunch",
		Date:        "2025-12-01",
	}, nil)

	resp, err := server.TransactionGet(context.Background(), req)

	assert.NoError(t, err)
	assert.Equal(t, &pb.TransactionGetResponse{
		Id:          1,
		Amount:      100.0,
		Category:    "Food",
		Description: "Lunch",
		Date:        "2025-12-01",
	}, resp)
	mockService.AssertExpectations(t)
}

func TestLedgerServer_TransactionList(t *testing.T) {
	mockService := &MockLedgerServiceImpl{}
	server := NewLedgerServer(mockService)

	req := &emptypb.Empty{}
	mockService.On("TransactionsList", mock.Anything).Return([]domain.Transaction{
		{
			ID:          1,
			Amount:      100.0,
			Category:    "Food",
			Description: "Lunch",
			Date:        "2025-12-01",
		},
	}, nil)

	resp, err := server.TransactionList(context.Background(), req)

	assert.NoError(t, err)
	assert.Len(t, resp.Transactions, 1)
	assert.Equal(t, &pb.TransactionGetResponse{
		Id:          1,
		Amount:      100.0,
		Category:    "Food",
		Description: "Lunch",
		Date:        "2025-12-01",
	}, resp.Transactions[0])
	mockService.AssertExpectations(t)
}

func TestLedgerServer_ReportSummary(t *testing.T) {
	mockService := &MockLedgerServiceImpl{}
	server := NewLedgerServer(mockService)

	req := &pb.SummaryRequest{From: "2025-12-01", To: "2025-12-31"}
	mockService.On("GetReportSummary", mock.Anything, "2025-12-01", "2025-12-31").Return(&domain.Summary{
		Categories:  map[string]float64{"Food": 500.0},
		CacheResult: true,
	}, nil)

	resp, err := server.ReportSummary(context.Background(), req)

	assert.NoError(t, err)
	assert.Equal(t, &pb.SummaryResponse{
		Report:      map[string]float64{"Food": 500.0},
		CacheResult: true,
	}, resp)
	mockService.AssertExpectations(t)
}

func TestLedgerServer_BulkAddTransactions(t *testing.T) {
	mockService := &MockLedgerServiceImpl{}
	server := NewLedgerServer(mockService)

	req := &pb.TransactionBulkAddRequest{
		Transactions: []*pb.TransactionAddRequest{
			{
				Amount:      100.0,
				Category:    "Food",
				Description: "Lunch",
				Date:        "2025-12-01",
			},
			{
				Amount:      50.0,
				Category:    "Transport",
				Description: "Bus",
				Date:        "2025-12-02",
			},
		},
	}
	mockService.On("BulkAddTransactions", mock.Anything, mock.MatchedBy(func(trs []domain.Transaction) bool {
		return len(trs) == 2
	}), 4).Return(&service.BulkTransactionResult{
		Accepted: 2,
		Rejected: 0,
		Errors:   nil,
	}, nil)

	resp, err := server.BulkAddTransactions(context.Background(), req)

	assert.NoError(t, err)
	assert.Equal(t, &pb.TransactionBulkAddResponse{
		Accepted: 2,
		Rejected: 0,
		Errors:   nil,
	}, resp)
	mockService.AssertExpectations(t)
}
