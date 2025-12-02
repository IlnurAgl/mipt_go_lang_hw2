package handlers

import (
	"encoding/json"
	"ledger"
	"net/http"
	"time"
)

func isValidDate(dateString string) bool {
	layout := "2006-01-02"
	_, err := time.Parse(layout, dateString)
	return err == nil
}

func ReportHandler(service ledger.LedgerService) http.HandlerFunc {
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
