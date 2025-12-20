package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"gateway/internal/model"
	"gateway/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockLedgerGatewayService struct {
	mock.Mock
}

var _ service.LedgerGatewayService = (*MockLedgerGatewayService)(nil)

func (m *MockLedgerGatewayService) BudgetAdd(ctx context.Context, req model.BudgetAdd) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

func (m *MockLedgerGatewayService) BudgetGet(ctx context.Context, req model.BudgetGet) (*model.BudgetGetResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*model.BudgetGetResponse), args.Error(1)
}

func (m *MockLedgerGatewayService) BudgetList(ctx context.Context) ([]model.BudgetGetResponse, error) {
	args := m.Called(ctx)
	return args.Get(0).([]model.BudgetGetResponse), args.Error(1)
}

func (m *MockLedgerGatewayService) TransactionAdd(ctx context.Context, req model.TrasnactionAdd) (*model.TransactionAddResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*model.TransactionAddResponse), args.Error(1)
}

func (m *MockLedgerGatewayService) TransactionGet(ctx context.Context, req model.TransactionGet) (*model.TransactionGetResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*model.TransactionGetResponse), args.Error(1)
}

func (m *MockLedgerGatewayService) TransactionList(ctx context.Context) ([]model.TransactionGetResponse, error) {
	args := m.Called(ctx)
	return args.Get(0).([]model.TransactionGetResponse), args.Error(1)
}

func (m *MockLedgerGatewayService) TransactionBulkAdd(ctx context.Context, req model.TransactionBulkAdd) (*model.TransactionBulkAddResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*model.TransactionBulkAddResponse), args.Error(1)
}

func (m *MockLedgerGatewayService) ReportSummary(ctx context.Context, req model.ReportSummary) (*model.ReportSummaryResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*model.ReportSummaryResponse), args.Error(1)
}

func TestLedgerHandler_BudgetAdd(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		requestBody    model.BudgetAdd
		mockSetup      func(*MockLedgerGatewayService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "successful budget add",
			requestBody: model.BudgetAdd{
				Category: "Food",
				Limit:    1000.0,
			},
			mockSetup: func(m *MockLedgerGatewayService) {
				m.On("BudgetAdd", mock.Anything, mock.MatchedBy(func(req model.BudgetAdd) bool {
					return req.Category == "Food" && req.Limit == 1000.0
				})).Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "",
		},
		{
			name: "empty category",
			requestBody: model.BudgetAdd{
				Category: "",
				Limit:    1000.0,
			},
			mockSetup:      func(m *MockLedgerGatewayService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"category must not be empty"}`,
		},
		{
			name: "service error",
			requestBody: model.BudgetAdd{
				Category: "Food",
				Limit:    1000.0,
			},
			mockSetup: func(m *MockLedgerGatewayService) {
				m.On("BudgetAdd", mock.Anything, mock.Anything).Return(assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"assert.AnError general error for testing"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockLedgerGatewayService{}
			tt.mockSetup(mockService)

			handler := NewLedgerHandler(mockService)

			body, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest(http.MethodPost, "/budget/", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = req

			handler.BudgetAdd(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedBody != "" {
				assert.JSONEq(t, tt.expectedBody, w.Body.String())
			}
			mockService.AssertExpectations(t)
		})
	}
}

func TestLedgerHandler_BudgetList(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		mockSetup      func(*MockLedgerGatewayService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "successful budget list",
			mockSetup: func(m *MockLedgerGatewayService) {
				m.On("BudgetList", mock.Anything).Return([]model.BudgetGetResponse{
					{Category: "Food", Limit: 1000.0},
					{Category: "Transport", Limit: 500.0},
				}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `[{"category":"Food","limit":1000},{"category":"Transport","limit":500}]`,
		},
		{
			name: "service error",
			mockSetup: func(m *MockLedgerGatewayService) {
				m.On("BudgetList", mock.Anything).Return([]model.BudgetGetResponse(nil), assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"assert.AnError general error for testing"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockLedgerGatewayService{}
			tt.mockSetup(mockService)

			handler := NewLedgerHandler(mockService)

			req, _ := http.NewRequest(http.MethodGet, "/budget/list", nil)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = req

			handler.BudgetList(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedBody != "" {
				assert.JSONEq(t, tt.expectedBody, w.Body.String())
			}
			mockService.AssertExpectations(t)
		})
	}
}

func TestLedgerHandler_BudgetGet(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		query          string
		mockSetup      func(*MockLedgerGatewayService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:  "successful budget get",
			query: "category=Food",
			mockSetup: func(m *MockLedgerGatewayService) {
				m.On("BudgetGet", mock.Anything, model.BudgetGet{Category: "Food"}).Return(&model.BudgetGetResponse{
					Category: "Food",
					Limit:    1000.0,
				}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"category":"Food","limit":1000}`,
		},
		{
			name:           "empty category",
			query:          "",
			mockSetup:      func(m *MockLedgerGatewayService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"category must not be empty"}`,
		},
		{
			name:  "service error",
			query: "category=Food",
			mockSetup: func(m *MockLedgerGatewayService) {
				m.On("BudgetGet", mock.Anything, mock.Anything).Return((*model.BudgetGetResponse)(nil), assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"assert.AnError general error for testing"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockLedgerGatewayService{}
			tt.mockSetup(mockService)

			handler := NewLedgerHandler(mockService)

			req, _ := http.NewRequest(http.MethodGet, "/budget/?"+tt.query, nil)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = req

			handler.BudgetGet(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedBody != "" {
				assert.JSONEq(t, tt.expectedBody, w.Body.String())
			}
			mockService.AssertExpectations(t)
		})
	}
}

func TestLedgerHandler_TransactionAdd(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		requestBody    model.TrasnactionAdd
		mockSetup      func(*MockLedgerGatewayService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "successful transaction add",
			requestBody: model.TrasnactionAdd{
				Amount:      100.0,
				Category:    "Food",
				Description: "Lunch",
				Date:        "2025-12-01",
			},
			mockSetup: func(m *MockLedgerGatewayService) {
				m.On("TransactionAdd", mock.Anything, mock.MatchedBy(func(req model.TrasnactionAdd) bool {
					return req.Amount == 100.0 && req.Category == "Food"
				})).Return(&model.TransactionAddResponse{Id: 1}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"id":1}`,
		},
		{
			name: "budget exceeded",
			requestBody: model.TrasnactionAdd{
				Amount:      100.0,
				Category:    "Food",
				Description: "Lunch",
				Date:        "2025-12-01",
			},
			mockSetup: func(m *MockLedgerGatewayService) {
				m.On("TransactionAdd", mock.Anything, mock.Anything).Return((*model.TransactionAddResponse)(nil), errors.New("rpc error: code = Internal desc = add transaction: budget exceeded"))
			},
			expectedStatus: http.StatusConflict,
			expectedBody:   `{"error":"budget exceeded"}`,
		},
		{
			name: "zero amount",
			requestBody: model.TrasnactionAdd{
				Amount:      0,
				Category:    "Food",
				Description: "Lunch",
				Date:        "2025-12-01",
			},
			mockSetup:      func(m *MockLedgerGatewayService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"amount must not be empty"}`,
		},
		{
			name: "empty category",
			requestBody: model.TrasnactionAdd{
				Amount:      100.0,
				Category:    "",
				Description: "Lunch",
				Date:        "2025-12-01",
			},
			mockSetup:      func(m *MockLedgerGatewayService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"category must not be empty"}`,
		},
		{
			name: "empty date",
			requestBody: model.TrasnactionAdd{
				Amount:      100.0,
				Category:    "Food",
				Description: "Lunch",
				Date:        "",
			},
			mockSetup:      func(m *MockLedgerGatewayService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"date must not be empty"}`,
		},
		{
			name: "empty description",
			requestBody: model.TrasnactionAdd{
				Amount:      100.0,
				Category:    "Food",
				Description: "",
				Date:        "2025-12-01",
			},
			mockSetup:      func(m *MockLedgerGatewayService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"category must not be empty"}`,
		},
		{
			name: "service error",
			requestBody: model.TrasnactionAdd{
				Amount:      100.0,
				Category:    "Food",
				Description: "Lunch",
				Date:        "2025-12-01",
			},
			mockSetup: func(m *MockLedgerGatewayService) {
				m.On("TransactionAdd", mock.Anything, mock.Anything).Return((*model.TransactionAddResponse)(nil), assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"assert.AnError general error for testing"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockLedgerGatewayService{}
			tt.mockSetup(mockService)

			handler := NewLedgerHandler(mockService)

			body, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest(http.MethodPost, "/transactions/", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = req

			handler.TransactionAdd(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedBody != "" {
				assert.JSONEq(t, tt.expectedBody, w.Body.String())
			}
			mockService.AssertExpectations(t)
		})
	}
}

func TestLedgerHandler_TransactionGet(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		query          string
		mockSetup      func(*MockLedgerGatewayService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:  "successful transaction get",
			query: "id=1",
			mockSetup: func(m *MockLedgerGatewayService) {
				m.On("TransactionGet", mock.Anything, model.TransactionGet{Id: 1}).Return(&model.TransactionGetResponse{
					Id:          1,
					Amount:      100.0,
					Category:    "Food",
					Description: "Lunch",
					Date:        "2025-12-01",
				}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"id":1,"amount":100,"category":"Food","description":"Lunch","Date":"2025-12-01"}`,
		},
		{
			name:           "empty id",
			query:          "",
			mockSetup:      func(m *MockLedgerGatewayService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"id must not be empty"}`,
		},
		{
			name:           "invalid id",
			query:          "id=abc",
			mockSetup:      func(m *MockLedgerGatewayService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"id must be int64"}`,
		},
		{
			name:  "service error",
			query: "id=1",
			mockSetup: func(m *MockLedgerGatewayService) {
				m.On("TransactionGet", mock.Anything, mock.Anything).Return((*model.TransactionGetResponse)(nil), assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"assert.AnError general error for testing"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockLedgerGatewayService{}
			tt.mockSetup(mockService)

			handler := NewLedgerHandler(mockService)

			req, _ := http.NewRequest(http.MethodGet, "/transactions/?"+tt.query, nil)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = req

			handler.TransactionGet(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedBody != "" {
				assert.JSONEq(t, tt.expectedBody, w.Body.String())
			}
			mockService.AssertExpectations(t)
		})
	}
}

func TestLedgerHandler_TransactionList(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		mockSetup      func(*MockLedgerGatewayService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "successful transaction list",
			mockSetup: func(m *MockLedgerGatewayService) {
				m.On("TransactionList", mock.Anything).Return([]model.TransactionGetResponse{
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
				}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `[{"id":1,"amount":100,"category":"Food","description":"Lunch","Date":"2025-12-01"},{"id":2,"amount":50,"category":"Transport","description":"Bus","Date":"2025-12-02"}]`,
		},
		{
			name: "service error",
			mockSetup: func(m *MockLedgerGatewayService) {
				m.On("TransactionList", mock.Anything).Return([]model.TransactionGetResponse(nil), assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"assert.AnError general error for testing"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockLedgerGatewayService{}
			tt.mockSetup(mockService)

			handler := NewLedgerHandler(mockService)

			req, _ := http.NewRequest(http.MethodGet, "/transactions/list", nil)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = req

			handler.TransactionList(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedBody != "" {
				assert.JSONEq(t, tt.expectedBody, w.Body.String())
			}
			mockService.AssertExpectations(t)
		})
	}
}

func TestLedgerHandler_TransactionExportCSV(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		mockSetup      func(*MockLedgerGatewayService)
		expectedStatus int
		expectedHeader map[string]string
		expectedBody   string
	}{
		{
			name: "successful export",
			mockSetup: func(m *MockLedgerGatewayService) {
				m.On("TransactionList", mock.Anything).Return([]model.TransactionGetResponse{
					{
						Id:          1,
						Amount:      100.0,
						Category:    "Food",
						Description: "Lunch",
						Date:        "2025-12-01",
					},
				}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedHeader: map[string]string{
				"Content-Type":        "text/csv",
				"Content-Disposition": "attachment; filename=transactions.csv",
			},
			expectedBody: "ID,Amount,Category,Description,Date\n1,100.00,Food,Lunch,2025-12-01\n",
		},
		{
			name: "service error",
			mockSetup: func(m *MockLedgerGatewayService) {
				m.On("TransactionList", mock.Anything).Return([]model.TransactionGetResponse(nil), assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
			expectedHeader: nil,
			expectedBody:   `{"error":"assert.AnError general error for testing"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockLedgerGatewayService{}
			tt.mockSetup(mockService)

			handler := NewLedgerHandler(mockService)

			req, _ := http.NewRequest(http.MethodGet, "/transactions/export.csv", nil)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = req

			handler.TransactionExportCSV(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedHeader != nil {
				for k, v := range tt.expectedHeader {
					assert.Equal(t, v, w.Header().Get(k))
				}
			}
			if tt.expectedBody != "" {
				if tt.expectedStatus == http.StatusOK {
					assert.Equal(t, tt.expectedBody, w.Body.String())
				} else {
					assert.JSONEq(t, tt.expectedBody, w.Body.String())
				}
			}
			mockService.AssertExpectations(t)
		})
	}
}

func TestLedgerHandler_ReportSummary(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		query          string
		mockSetup      func(*MockLedgerGatewayService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:  "successful report",
			query: "from=2025-12-01&to=2025-12-31",
			mockSetup: func(m *MockLedgerGatewayService) {
				m.On("ReportSummary", mock.Anything, model.ReportSummary{From: "2025-12-01", To: "2025-12-31"}).Return(&model.ReportSummaryResponse{
					Report:      map[string]float64{"Food": 500.0},
					CacheResult: true,
				}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"report":{"Food":500},"cache_result":true}`,
		},
		{
			name:           "invalid from date",
			query:          "from=invalid&to=2025-12-31",
			mockSetup:      func(m *MockLedgerGatewayService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"invalid from"}`,
		},
		{
			name:           "invalid to date",
			query:          "from=2025-12-01&to=invalid",
			mockSetup:      func(m *MockLedgerGatewayService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"invalid to"}`,
		},
		{
			name:  "service error",
			query: "from=2025-12-01&to=2025-12-31",
			mockSetup: func(m *MockLedgerGatewayService) {
				m.On("ReportSummary", mock.Anything, mock.Anything).Return((*model.ReportSummaryResponse)(nil), assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"assert.AnError general error for testing"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockLedgerGatewayService{}
			tt.mockSetup(mockService)

			handler := NewLedgerHandler(mockService)

			req, _ := http.NewRequest(http.MethodGet, "/reports/summary?"+tt.query, nil)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = req

			handler.ReportSummary(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedBody != "" {
				assert.JSONEq(t, tt.expectedBody, w.Body.String())
			}
			mockService.AssertExpectations(t)
		})
	}
}

func TestLedgerHandler_TransactionBulkAdd(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		requestBody    model.TransactionBulkAdd
		mockSetup      func(*MockLedgerGatewayService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "successful bulk add",
			requestBody: model.TransactionBulkAdd{
				Transactions: []model.TrasnactionAdd{
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
			},
			mockSetup: func(m *MockLedgerGatewayService) {
				m.On("TransactionBulkAdd", mock.Anything, mock.MatchedBy(func(req model.TransactionBulkAdd) bool {
					return len(req.Transactions) == 2
				})).Return(&model.TransactionBulkAddResponse{
					Accepted: 2,
					Rejected: 0,
					Errors:   nil,
				}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"Accepted":2,"Rejected":0,"Errors":null}`,
		},
		{
			name: "partial success",
			requestBody: model.TransactionBulkAdd{
				Transactions: []model.TrasnactionAdd{
					{
						Amount:      100.0,
						Category:    "Food",
						Description: "Lunch",
						Date:        "2025-12-01",
					},
				},
			},
			mockSetup: func(m *MockLedgerGatewayService) {
				m.On("TransactionBulkAdd", mock.Anything, mock.Anything).Return(&model.TransactionBulkAddResponse{
					Accepted: 0,
					Rejected: 1,
					Errors:   map[int64]string{0: "budget exceeded"},
				}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"Accepted":0,"Rejected":1,"Errors":{"0":"budget exceeded"}}`,
		},
		{
			name: "service error",
			requestBody: model.TransactionBulkAdd{
				Transactions: []model.TrasnactionAdd{
					{
						Amount:      100.0,
						Category:    "Food",
						Description: "Lunch",
						Date:        "2025-12-01",
					},
				},
			},
			mockSetup: func(m *MockLedgerGatewayService) {
				m.On("TransactionBulkAdd", mock.Anything, mock.Anything).Return((*model.TransactionBulkAddResponse)(nil), assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"assert.AnError general error for testing"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockLedgerGatewayService{}
			tt.mockSetup(mockService)

			handler := NewLedgerHandler(mockService)

			body, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest(http.MethodPost, "/transactions/bulk", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = req

			handler.TransactionBulkAdd(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedBody != "" {
				assert.JSONEq(t, tt.expectedBody, w.Body.String())
			}
			mockService.AssertExpectations(t)
		})
	}
}
