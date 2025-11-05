package main

import (
	"encoding/json"
	"fmt"
	"gateway/internal"
	"io"
	"ledger"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

func ping(w http.ResponseWriter, r *http.Request) {
	_, err := io.WriteString(w, "pong")
	if err != nil {
		return
	}
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("Request received: %s %s\n", r.Method, r.URL.Path)
		t := time.Now()
		next.ServeHTTP(w, r)
		fmt.Printf("Request handled: %s %s, time: %s\n", r.Method, r.URL.Path, time.Since(t))
	})
}

func transactionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		var transaction internal.CreateTransactionRequest
		err := json.NewDecoder(r.Body).Decode(&transaction)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			err = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			if err != nil {
				return
			}
			return
		}
		response, err := internal.CreateTransaction(transaction)
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
		trs := make([]internal.TransactionResponse, 0)
		for _, tr := range ledger.ListTransactions() {
			trs = append(trs, internal.TransactionResponse{
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
		err := json.NewEncoder(w).Encode(trs)
		if err != nil {
			return
		}
	}
}

func budgetHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		budgets := make([]internal.BudgetResponse, 0)
		for _, budget := range ledger.ListBudgets() {
			budgets = append(budgets, internal.BudgetResponse{
				Category: budget.Category,
				Limit:    budget.Limit,
			})
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("charset", "UTF-8")
		err := json.NewEncoder(w).Encode(budgets)
		if err != nil {
			return
		}
	}
	if r.Method == "POST" {
		var budget internal.CreateBudgetRequest
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
		response, err := internal.CreateBudget(budget)
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

func main() {
	r := mux.NewRouter()
	r.Use(loggingMiddleware)
	r.HandleFunc("/ping", ping)
	r.HandleFunc("/api/transaction", transactionHandler)
	r.HandleFunc("/api/budget", budgetHandler)

	log.Fatal(http.ListenAndServe(":8080", r))
}
