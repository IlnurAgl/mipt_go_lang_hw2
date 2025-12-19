package handlers

import (
	"encoding/json"
	"gateway/internal/api"
	"ledger"
	"net/http"
)

type BudgetHandlers struct {
	service ledger.LedgerService
}

func NewBudgetHandlers(s ledger.LedgerService) *BudgetHandlers {
	return &BudgetHandlers{service: s}
}

func (t *BudgetHandlers) budgetGetHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("charset", "UTF-8")
	budgets := make([]api.BudgetResponse, 0)
	response, err := t.service.GetBudgets(r.Context())
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

func (t *BudgetHandlers) budgetPostHandler(w http.ResponseWriter, r *http.Request) {
	var budget api.CreateBudgetRequest
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("charset", "UTF-8")
	err := json.NewDecoder(r.Body).Decode(&budget)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, err := w.Write([]byte("{\"error\": \"invalid json\"}"))
		if err != nil {
			return
		}
		return
	}
	response, err := api.CreateBudget(t.service, budget, r.Context())
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

func (t *BudgetHandlers) BudgetHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		t.budgetGetHandler(w, r)
		return
	}
	if r.Method == "POST" {
		t.budgetPostHandler(w, r)
		return
	}
	w.WriteHeader(http.StatusMethodNotAllowed)
}
