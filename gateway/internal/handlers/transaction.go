package handlers

import (
	"encoding/json"
	"gateway/internal/api"
	"ledger"
	"net/http"
)

func TransactionHandler(service ledger.LedgerService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
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
			response, err := api.CreateTransaction(service, transaction, r.Context())
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
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("charset", "UTF-8")
			err = json.NewEncoder(w).Encode(response)
			if err != nil {
				return
			}
			return
		}
		if r.Method == "GET" {
			trs := make([]api.TransactionResponse, 0)
			dbTrs, err := service.ListTransactions(r.Context())
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("charset", "UTF-8")
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
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("charset", "UTF-8")
			err = json.NewEncoder(w).Encode(trs)
			if err != nil {
				return
			}
		}
	}
}

func TransactionBulkHandler(service ledger.LedgerService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
			result, _ := api.BulkCreateTransaction(service, transactions, r.Context())
			w.WriteHeader(http.StatusOK)
			err = json.NewEncoder(w).Encode(result)
			if err != nil {
				return
			}
			return
		}
	}
}
