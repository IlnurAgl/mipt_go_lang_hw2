package main

import (
	"context"
	"encoding/json"
	"fmt"
	"gateway/internal/handlers"
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
	r.HandleFunc("/api/transaction", handlers.TransactionHandler(service))
	r.HandleFunc("/api/budget", handlers.BudgetHandler(service))
	r.HandleFunc("/api/reports/summary", handlers.ReportHandler(service))
	r.HandleFunc("/api/transactions/bulk", handlers.TransactionBulkHandler(service))

	log.Fatal(http.ListenAndServe(":8080", r))
}
