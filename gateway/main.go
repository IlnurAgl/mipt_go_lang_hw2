package main

import (
	"context"
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

func contextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancelFn := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancelFn()
		r = r.WithContext(ctx)
		done := make(chan struct{})

		go func() {
			next.ServeHTTP(w, r)
			close(done)
		}()
		select {
		case <-done:
			return
		case <-ctx.Done():
			if ctx.Err() == context.DeadlineExceeded {
				w.WriteHeader(http.StatusGatewayTimeout)
				fmt.Printf("Request timed out")
				err := json.NewEncoder(w).Encode(map[string]string{"error": "Request timed out"})
				if err != nil {
					return
				}
			}
			return
		}
	})
}

func transactionHandler(service ledger.LedgerService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
			response, err := internal.CreateTransaction(service, transaction, r.Context())
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
			err = json.NewEncoder(w).Encode(trs)
			if err != nil {
				return
			}
		}
	}
}

func budgetHandler(service ledger.LedgerService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("charset", "UTF-8")
			budgets := make([]internal.BudgetResponse, 0)
			response, err := service.GetBudgets(r.Context())
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				err = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			}
			for _, budget := range response {
				budgets = append(budgets, internal.BudgetResponse{
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
			response, err := internal.CreateBudget(service, budget, r.Context())
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

type Report struct {
	From string `json:"from"`
	To   string `json:"to"`
}

func isValidDate(dateString string) bool {
	layout := "2006-01-02"
	_, err := time.Parse(layout, dateString)
	return err == nil
}

func reportHandler(service ledger.LedgerService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("charset", "UTF-8")
			queryParams := r.URL.Query()
			from := queryParams.Get("from")
			validFrom := isValidDate(from)
			to := queryParams.Get("to")
			validTo := isValidDate(to)
			if !validFrom || !validTo {
				w.WriteHeader(http.StatusBadRequest)
				var s string
				if !validFrom {
					s = "Invalid from"
				} else {
					s = "Invalid to"
				}
				err := json.NewEncoder(w).Encode(map[string]string{"error": s})
				if err != nil {
					return
				}
				return
			}
			summary, err := service.GetReportSummary(from, to, r.Context())
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				err = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
				return
			}
			w.WriteHeader(http.StatusOK)
			err = json.NewEncoder(w).Encode(summary)
			if err != nil {
				return
			}
			return
		}
	}
}

func main() {
	service, closeFn, err := ledger.NewLedgerService()
	defer closeFn()
	if err != nil {
		println(err.Error())
		return
	}

	r := mux.NewRouter()
	r.Use(loggingMiddleware)
	r.Use(contextMiddleware)
	r.HandleFunc("/ping", ping)
	r.HandleFunc("/api/transaction", transactionHandler(service))
	r.HandleFunc("/api/budget", budgetHandler(service))
	r.HandleFunc("/api/reports/summary", reportHandler(service))

	log.Fatal(http.ListenAndServe(":8080", r))
}
