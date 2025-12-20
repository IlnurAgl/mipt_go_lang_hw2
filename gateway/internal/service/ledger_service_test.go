package service

import (
	"context"
	"testing"

	"gateway/internal/model"
	ledgerv1 "gateway/internal/pb/ledger/v1"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockLedgerServiceClient struct {
	mock.Mock
}

func (m *MockLedgerServiceClient) BudgetAdd(ctx context.Context, in *ledgerv1.BudgetAddRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	args := m.Called(ctx, in, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*emptypb.Empty), args.Error(1)
}

func (m *MockLedgerServiceClient) BudgetGet(ctx context.Context, in *ledgerv1.BudgetGetRequest, opts ...grpc.CallOption) (*ledgerv1.BudgetGetResponse, error) {
	args := m.Called(ctx, in, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ledgerv1.BudgetGetResponse), args.Error(1)
}

func (m *MockLedgerServiceClient) BudgetsList(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*ledgerv1.BudgetGetListResponse, error) {
	args := m.Called(ctx, in, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ledgerv1.BudgetGetListResponse), args.Error(1)
}

func (m *MockLedgerServiceClient) TransactionAdd(ctx context.Context, in *ledgerv1.TransactionAddRequest, opts ...grpc.CallOption) (*ledgerv1.TransactionAddResponse, error) {
	args := m.Called(ctx, in, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ledgerv1.TransactionAddResponse), args.Error(1)
}

func (m *MockLedgerServiceClient) TransactionGet(ctx context.Context, in *ledgerv1.TransactionGetRequest, opts ...grpc.CallOption) (*ledgerv1.TransactionGetResponse, error) {
	args := m.Called(ctx, in, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ledgerv1.TransactionGetResponse), args.Error(1)
}

func (m *MockLedgerServiceClient) TransactionList(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*ledgerv1.TransactionGetListResponse, error) {
	args := m.Called(ctx, in, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ledgerv1.TransactionGetListResponse), args.Error(1)
}

func (m *MockLedgerServiceClient) ReportSummary(ctx context.Context, in *ledgerv1.SummaryRequest, opts ...grpc.CallOption) (*ledgerv1.SummaryResponse, error) {
	args := m.Called(ctx, in, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ledgerv1.SummaryResponse), args.Error(1)
}

func (m *MockLedgerServiceClient) BulkAddTransactions(ctx context.Context, in *ledgerv1.TransactionBulkAddRequest, opts ...grpc.CallOption) (*ledgerv1.TransactionBulkAddResponse, error) {
	args := m.Called(ctx, in, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ledgerv1.TransactionBulkAddResponse), args.Error(1)
}

func TestLedgerGatewayService_BudgetAdd(t *testing.T) {
	mockClient := &MockLedgerServiceClient{}
	service := NewLedgerGatewayService(mockClient)
	ctx := context.Background()

	req := model.BudgetAdd{
		Category: "Food",
		Limit:    1000.0,
	}

	tests := []struct {
		name        string
		req         model.BudgetAdd
		mockSetup   func()
		expectedErr bool
	}{
		{
			name: "successful budget add",
			req:  req,
			mockSetup: func() {
				mockClient.On("BudgetAdd", ctx, &ledgerv1.BudgetAddRequest{
					Category: "Food",
					Limit:    1000.0,
				}, mock.Anything).Return(&emptypb.Empty{}, nil)
			},
			expectedErr: false,
		},
		{
			name: "gRPC error",
			req:  req,
			mockSetup: func() {
				mockClient.On("BudgetAdd", ctx, &ledgerv1.BudgetAddRequest{
					Category: "Food",
					Limit:    1000.0,
				}, mock.Anything).Return(nil, assert.AnError)
			},
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient.ExpectedCalls = nil
			tt.mockSetup()

			err := service.BudgetAdd(ctx, tt.req)

			if tt.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			mockClient.AssertExpectations(t)
		})
	}
}

func TestLedgerGatewayService_BudgetGet(t *testing.T) {
	mockClient := &MockLedgerServiceClient{}
	service := NewLedgerGatewayService(mockClient)
	ctx := context.Background()

	req := model.BudgetGet{
		Category: "Food",
	}

	tests := []struct {
		name        string
		req         model.BudgetGet
		mockSetup   func()
		expected    *model.BudgetGetResponse
		expectedErr bool
	}{
		{
			name: "successful budget get",
			req:  req,
			mockSetup: func() {
				mockClient.On("BudgetGet", ctx, &ledgerv1.BudgetGetRequest{
					Category: "Food",
				}, mock.Anything).Return(&ledgerv1.BudgetGetResponse{
					Category: "Food",
					Limit:    1000.0,
				}, nil)
			},
			expected: &model.BudgetGetResponse{
				Category: "Food",
				Limit:    1000.0,
			},
			expectedErr: false,
		},
		{
			name: "gRPC error",
			req:  req,
			mockSetup: func() {
				mockClient.On("BudgetGet", ctx, &ledgerv1.BudgetGetRequest{
					Category: "Food",
				}, mock.Anything).Return(nil, assert.AnError)
			},
			expected:    nil,
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient.ExpectedCalls = nil
			tt.mockSetup()

			result, err := service.BudgetGet(ctx, tt.req)

			if tt.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
			mockClient.AssertExpectations(t)
		})
	}
}

func TestLedgerGatewayService_BudgetList(t *testing.T) {
	mockClient := &MockLedgerServiceClient{}
	service := NewLedgerGatewayService(mockClient)
	ctx := context.Background()

	tests := []struct {
		name        string
		mockSetup   func()
		expected    []model.BudgetGetResponse
		expectedErr bool
	}{
		{
			name: "successful budget list",
			mockSetup: func() {
				mockClient.On("BudgetsList", ctx, mock.AnythingOfType("*emptypb.Empty"), mock.Anything).Return(&ledgerv1.BudgetGetListResponse{
					Budgets: []*ledgerv1.BudgetGetResponse{
						{Category: "Food", Limit: 1000.0},
						{Category: "Transport", Limit: 500.0},
					},
				}, nil)
			},
			expected: []model.BudgetGetResponse{
				{Category: "Food", Limit: 1000.0},
				{Category: "Transport", Limit: 500.0},
			},
			expectedErr: false,
		},
		{
			name: "gRPC error",
			mockSetup: func() {
				mockClient.On("BudgetsList", ctx, mock.AnythingOfType("*emptypb.Empty"), mock.Anything).Return(nil, assert.AnError)
			},
			expected:    nil,
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient.ExpectedCalls = nil
			tt.mockSetup()

			result, err := service.BudgetList(ctx)

			if tt.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
			mockClient.AssertExpectations(t)
		})
	}
}

func TestLedgerGatewayService_TransactionAdd(t *testing.T) {
	mockClient := &MockLedgerServiceClient{}
	service := NewLedgerGatewayService(mockClient)
	ctx := context.Background()

	req := model.TrasnactionAdd{
		Amount:      100.0,
		Category:    "Food",
		Description: "Lunch",
		Date:        "2025-12-01",
	}

	tests := []struct {
		name        string
		req         model.TrasnactionAdd
		mockSetup   func()
		expected    *model.TransactionAddResponse
		expectedErr bool
	}{
		{
			name: "successful transaction add",
			req:  req,
			mockSetup: func() {
				mockClient.On("TransactionAdd", ctx, &ledgerv1.TransactionAddRequest{
					Amount:      100.0,
					Category:    "Food",
					Description: "Lunch",
					Date:        "2025-12-01",
				}, mock.Anything).Return(&ledgerv1.TransactionAddResponse{
					Id: 1,
				}, nil)
			},
			expected: &model.TransactionAddResponse{
				Id: 1,
			},
			expectedErr: false,
		},
		{
			name: "gRPC error",
			req:  req,
			mockSetup: func() {
				mockClient.On("TransactionAdd", ctx, &ledgerv1.TransactionAddRequest{
					Amount:      100.0,
					Category:    "Food",
					Description: "Lunch",
					Date:        "2025-12-01",
				}, mock.Anything).Return(nil, assert.AnError)
			},
			expected:    nil,
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient.ExpectedCalls = nil
			tt.mockSetup()

			result, err := service.TransactionAdd(ctx, tt.req)

			if tt.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
			mockClient.AssertExpectations(t)
		})
	}
}

func TestLedgerGatewayService_TransactionGet(t *testing.T) {
	mockClient := &MockLedgerServiceClient{}
	service := NewLedgerGatewayService(mockClient)
	ctx := context.Background()

	req := model.TransactionGet{
		Id: 1,
	}

	tests := []struct {
		name        string
		req         model.TransactionGet
		mockSetup   func()
		expected    *model.TransactionGetResponse
		expectedErr bool
	}{
		{
			name: "successful transaction get",
			req:  req,
			mockSetup: func() {
				mockClient.On("TransactionGet", ctx, &ledgerv1.TransactionGetRequest{
					Id: 1,
				}, mock.Anything).Return(&ledgerv1.TransactionGetResponse{
					Id:          1,
					Amount:      100.0,
					Category:    "Food",
					Description: "Lunch",
					Date:        "2025-12-01",
				}, nil)
			},
			expected: &model.TransactionGetResponse{
				Id:          1,
				Amount:      100.0,
				Category:    "Food",
				Description: "Lunch",
				Date:        "2025-12-01",
			},
			expectedErr: false,
		},
		{
			name: "gRPC error",
			req:  req,
			mockSetup: func() {
				mockClient.On("TransactionGet", ctx, &ledgerv1.TransactionGetRequest{
					Id: 1,
				}, mock.Anything).Return(nil, assert.AnError)
			},
			expected:    nil,
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient.ExpectedCalls = nil
			tt.mockSetup()

			result, err := service.TransactionGet(ctx, tt.req)

			if tt.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
			mockClient.AssertExpectations(t)
		})
	}
}

func TestLedgerGatewayService_TransactionList(t *testing.T) {
	mockClient := &MockLedgerServiceClient{}
	service := NewLedgerGatewayService(mockClient)
	ctx := context.Background()

	tests := []struct {
		name        string
		mockSetup   func()
		expected    []model.TransactionGetResponse
		expectedErr bool
	}{
		{
			name: "successful transaction list",
			mockSetup: func() {
				mockClient.On("TransactionList", ctx, mock.AnythingOfType("*emptypb.Empty"), mock.Anything).Return(&ledgerv1.TransactionGetListResponse{
					Transactions: []*ledgerv1.TransactionGetResponse{
						{
							Id:          1,
							Amount:      100.0,
							Category:    "Food",
							Description: "Lunch",
							Date:        "2025-12-01",
						},
						{
							Id:          2,
							Amount:      50.0,
							Category:    "Transport",
							Description: "Bus",
							Date:        "2025-12-02",
						},
					},
				}, nil)
			},
			expected: []model.TransactionGetResponse{
				{
					Id:          1,
					Amount:      100.0,
					Category:    "Food",
					Description: "Lunch",
					Date:        "2025-12-01",
				},
				{
					Id:          2,
					Amount:      50.0,
					Category:    "Transport",
					Description: "Bus",
					Date:        "2025-12-02",
				},
			},
			expectedErr: false,
		},
		{
			name: "gRPC error",
			mockSetup: func() {
				mockClient.On("TransactionList", ctx, mock.AnythingOfType("*emptypb.Empty"), mock.Anything).Return(nil, assert.AnError)
			},
			expected:    nil,
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient.ExpectedCalls = nil
			tt.mockSetup()

			result, err := service.TransactionList(ctx)

			if tt.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
			mockClient.AssertExpectations(t)
		})
	}
}

func TestLedgerGatewayService_ReportSummary(t *testing.T) {
	mockClient := &MockLedgerServiceClient{}
	service := NewLedgerGatewayService(mockClient)
	ctx := context.Background()

	req := model.ReportSummary{
		From: "2025-12-01",
		To:   "2025-12-31",
	}

	tests := []struct {
		name        string
		req         model.ReportSummary
		mockSetup   func()
		expected    *model.ReportSummaryResponse
		expectedErr bool
	}{
		{
			name: "successful report summary",
			req:  req,
			mockSetup: func() {
				mockClient.On("ReportSummary", ctx, &ledgerv1.SummaryRequest{
					From: "2025-12-01",
					To:   "2025-12-31",
				}, mock.Anything).Return(&ledgerv1.SummaryResponse{
					Report:      map[string]float64{"Food": 500.0},
					CacheResult: false,
				}, nil)
			},
			expected: &model.ReportSummaryResponse{
				Report:      map[string]float64{"Food": 500.0},
				CacheResult: false,
			},
			expectedErr: false,
		},
		{
			name: "gRPC error",
			req:  req,
			mockSetup: func() {
				mockClient.On("ReportSummary", ctx, &ledgerv1.SummaryRequest{
					From: "2025-12-01",
					To:   "2025-12-31",
				}, mock.Anything).Return(nil, assert.AnError)
			},
			expected:    nil,
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient.ExpectedCalls = nil
			tt.mockSetup()

			result, err := service.ReportSummary(ctx, tt.req)

			if tt.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
			mockClient.AssertExpectations(t)
		})
	}
}

func TestLedgerGatewayService_TransactionBulkAdd(t *testing.T) {
	mockClient := &MockLedgerServiceClient{}
	service := NewLedgerGatewayService(mockClient)
	ctx := context.Background()

	req := model.TransactionBulkAdd{
		Transactions: []model.TrasnactionAdd{
			{
				Amount:      100.0,
				Category:    "Food",
				Description: "Lunch",
				Date:        "2025-12-01",
			},
		},
	}

	tests := []struct {
		name        string
		req         model.TransactionBulkAdd
		mockSetup   func()
		expected    *model.TransactionBulkAddResponse
		expectedErr bool
	}{
		{
			name: "successful bulk add",
			req:  req,
			mockSetup: func() {
				mockClient.On("BulkAddTransactions", ctx, &ledgerv1.TransactionBulkAddRequest{
					Transactions: []*ledgerv1.TransactionAddRequest{
						{
							Amount:      100.0,
							Category:    "Food",
							Description: "Lunch",
							Date:        "2025-12-01",
						},
					},
				}, mock.Anything).Return(&ledgerv1.TransactionBulkAddResponse{
					Accepted: 1,
					Rejected: 0,
					Errors:   nil,
				}, nil)
			},
			expected: &model.TransactionBulkAddResponse{
				Accepted: 1,
				Rejected: 0,
				Errors:   nil,
			},
			expectedErr: false,
		},
		{
			name: "gRPC error",
			req:  req,
			mockSetup: func() {
				mockClient.On("BulkAddTransactions", ctx, &ledgerv1.TransactionBulkAddRequest{
					Transactions: []*ledgerv1.TransactionAddRequest{
						{
							Amount:      100.0,
							Category:    "Food",
							Description: "Lunch",
							Date:        "2025-12-01",
						},
					},
				}, mock.Anything).Return(nil, assert.AnError)
			},
			expected:    nil,
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient.ExpectedCalls = nil
			tt.mockSetup()

			result, err := service.TransactionBulkAdd(ctx, tt.req)

			if tt.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
			mockClient.AssertExpectations(t)
		})
	}
}
