package handlers

import (
	"encoding/json"
	"gateway/internal/api"
	"ledger"
	"net/http"
)

func BudgetHandler(service ledger.LedgerService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("charset", "UTF-8")
			budgets := make([]api.BudgetResponse, 0)
			response, err := service.GetBudgets(r.Context())
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				err = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			}
			for _, budget := range response {
				budgets = append(budgets, api.BudgetResponse{
					Category: budget.Category,
					Limit:    budget.Limit,
				})
			}
			w.WriteHeader(http.StatusOK)
			err = json.NewEncoder(w).Encode(budgets)
			if err != nil {
				return
			}
		}
		if r.Method == "POST" {
			var budget api.CreateBudgetRequest
			err := json.NewDecoder(r.Body).Decode(&budget)
			if err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("charset", "UTF-8")
				w.WriteHeader(http.StatusBadRequest)
				_, err := w.Write([]byte("{\"error\": \"invalid json\"}"))
				if err != nil {
					return
				}
				return
			}
			response, err := api.CreateBudget(service, budget, r.Context())
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("charset", "UTF-8")
			if err != nil && err.Error() == "invalid limit" {
				w.WriteHeader(http.StatusBadRequest)
				err = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
				if err != nil {
					return
				}
				return
			}
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				err = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
				if err != nil {
					return
				}
				return
			}
			w.WriteHeader(http.StatusCreated)
			err = json.NewEncoder(w).Encode(response)
			if err != nil {
				return
			}
			return
		}
	}
}
