package main

import (
	"bytes"
	"encoding/json"
	"gateway/internal"
	"io"
	"ledger"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBudgetHandler(t *testing.T) {
	bodyReader := bytes.NewBuffer([]byte("{\"limit\": 1000, \"category\": \"test\"}"))
	req := httptest.NewRequest(http.MethodPost, "/api/budget", bodyReader)
	w := httptest.NewRecorder()
	budgetHandler(w, req)
	res := w.Result()
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	}(res.Body)
	assert.Equal(t, http.StatusCreated, res.StatusCode)
	assert.Equal(t, "application/json", res.Header.Get("Content-Type"))
	assert.Equal(t, "UTF-8", res.Header.Get("charset"))
	data, err := io.ReadAll(res.Body)
	if err != nil {
		t.Errorf("expected error to be nil got %v", err)
	}
	assert.Equal(t, "{\"Category\":\"test\",\"Limit\":1000}\n", string(data))
}

func TestInvalidJsonBudgetHandler(t *testing.T) {
	bodyReader := bytes.NewBuffer([]byte("test"))
	req := httptest.NewRequest(http.MethodPost, "/api/budget", bodyReader)
	w := httptest.NewRecorder()
	budgetHandler(w, req)
	res := w.Result()
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	}(res.Body)
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
	assert.Equal(t, "application/json", res.Header.Get("Content-Type"))
	assert.Equal(t, "UTF-8", res.Header.Get("charset"))
	data, err := io.ReadAll(res.Body)
	if err != nil {
		t.Errorf("expected error to be nil got %v", err)
	}
	assert.Equal(t, "{\"error\": \"invalid json\"}", string(data))
}

func TestInvalidLimitBidgetHandler(t *testing.T) {
	bodyReader := bytes.NewBuffer([]byte("{\"limit\": -1000, \"category\": \"test\"}"))
	req := httptest.NewRequest(http.MethodPost, "/api/budget", bodyReader)
	w := httptest.NewRecorder()
	budgetHandler(w, req)
	res := w.Result()
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	}(res.Body)
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
	data, err := io.ReadAll(res.Body)
	if err != nil {
		t.Errorf("expected error to be nil got %v", err)
	}
	assert.Equal(t, "{\"error\":\"invalid limit\"}\n", string(data))
}

func TestIntegration(t *testing.T) {
	t.Parallel()
	bodyReader := bytes.NewBuffer([]byte("{\"limit\": 1000, \"category\": \"test\"}"))
	req := httptest.NewRequest(http.MethodPost, "/api/budget", bodyReader)
	w := httptest.NewRecorder()
	budgetHandler(w, req)
	res := w.Result()
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	}(res.Body)
	assert.Equal(t, http.StatusCreated, res.StatusCode)
	assert.Equal(t, "application/json", res.Header.Get("Content-Type"))
	assert.Equal(t, "UTF-8", res.Header.Get("charset"))
	tests := []struct {
		name         string
		tr           internal.CreateTransactionRequest
		status       int
		message      string
		transactions []internal.TransactionResponse
	}{
		{
			name: "ok",
			tr: internal.CreateTransactionRequest{
				Category:    "test",
				Amount:      1000,
				Description: "test",
				Date:        "2025-11-05",
			},
			status:  http.StatusCreated,
			message: "{\"id\":1,\"amount\":1000,\"category\":\"test\",\"description\":\"test\",\"date\":\"2025-11-05\"}\n",
			transactions: []internal.TransactionResponse{
				{
					ID:          1,
					Amount:      1000,
					Category:    "test",
					Description: "test",
					Date:        "2025-11-05",
				},
			},
		},
		{
			name: "budger exceeded",
			tr: internal.CreateTransactionRequest{
				Category:    "test",
				Amount:      2000,
				Description: "test",
				Date:        "2025-11-05",
			},
			status:  http.StatusConflict,
			message: "{\"error\":\"budget exceeded\"}\n",
		},
		{
			name: "invalid json",
			tr: internal.CreateTransactionRequest{
				Category:    "test",
				Description: "test",
				Date:        "2025-11-05",
			},
			status:  http.StatusBadRequest,
			message: "{\"error\":\"invalid transaction\"}\n",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			bodyReader := bytes.NewBuffer([]byte{})
			err := json.NewEncoder(bodyReader).Encode(test.tr)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			req := httptest.NewRequest(http.MethodPost, "/api/transaction", bodyReader)
			w := httptest.NewRecorder()
			transactionHandler(w, req)
			result := w.Result()
			defer func(Body io.ReadCloser) {
				err := Body.Close()
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}(result.Body)
			assert.Equal(t, result.StatusCode, test.status)
			resData, err := io.ReadAll(result.Body)
			if err != nil {
				t.Errorf("expected error to be nil got %v", err)
			}
			assert.Equal(t, test.message, string(resData))
			if test.status == http.StatusCreated {
				req := httptest.NewRequest(http.MethodGet, "/api/transaction", nil)
				w := httptest.NewRecorder()
				transactionHandler(w, req)
				result := w.Result()
				defer func(Body io.ReadCloser) {
					err := Body.Close()
					if err != nil {
						t.Errorf("Unexpected error: %v", err)
					}
				}(result.Body)
				var transactions []internal.TransactionResponse
				err = json.NewDecoder(result.Body).Decode(&transactions)
				if err != nil {
					t.Errorf("expected error to be nil got %v", err)
				}
				assert.Equal(t, test.transactions, transactions)
			}
			ledger.Reset()
		})
	}
}
