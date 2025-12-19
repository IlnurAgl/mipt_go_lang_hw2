package handlers

import (
	"encoding/json"
	"gateway/internal/api"
	"ledger"
	"net/http"
)

type TransactionHandlers struct {
	service ledger.LedgerService
}

func NewTransactionHandlers(s ledger.LedgerService) *TransactionHandlers {
	return &TransactionHandlers{service: s}
}

func (t *TransactionHandlers) transactionsGetHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("charset", "UTF-8")
	trs := make([]api.TransactionResponse, 0)
	dbTrs, err := t.service.ListTransactions(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		err = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		if err != nil {
			return
		}
		return
	}
	for _, tr := range dbTrs {
		trs = append(trs, api.TransactionResponse{
			ID:          tr.ID,
			Amount:      tr.Amount,
			Date:        tr.Date,
			Category:    tr.Category,
			Description: tr.Description,
		})
	}

	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(trs)
	if err != nil {
		return
	}
}

func (t *TransactionHandlers) transactionsPostHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("charset", "UTF-8")
	var transaction api.CreateTransactionRequest
	err := json.NewDecoder(r.Body).Decode(&transaction)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		err = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		if err != nil {
			return
		}
		return
	}
	response, err := api.CreateTransaction(t.service, transaction, r.Context())
	if err != nil {
		switch err.Error() {
		case "invalid transaction":
			w.WriteHeader(http.StatusBadRequest)
			err = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			if err != nil {
				return
			}
			return
		case "budget exceeded":
			w.WriteHeader(http.StatusConflict)
			err := json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			if err != nil {
				return
			}
			return
		default:
			w.WriteHeader(http.StatusInternalServerError)
			err = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			if err != nil {
				return
			}
			return
		}
	}

	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		return
	}
}

func (t *TransactionHandlers) TransactionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		t.transactionsPostHandler(w, r)
	}
	if r.Method == "GET" {
		t.transactionsGetHandler(w, r)
	}
	w.WriteHeader(http.StatusMethodNotAllowed)
}

func (t *TransactionHandlers) TransactionBulkHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("charset", "UTF-8")
		var transactions []api.CreateTransactionRequest
		err := json.NewDecoder(r.Body).Decode(&transactions)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			err = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			if err != nil {
				return
			}
			return
		}
		result, _ := api.BulkCreateTransaction(t.service, transactions, r.Context())
		w.WriteHeader(http.StatusOK)
		err = json.NewEncoder(w).Encode(result)
		if err != nil {
			return
		}
		return
	}
}
