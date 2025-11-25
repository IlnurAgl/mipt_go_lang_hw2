package domain

import "context"

type Summary struct {
	Categories map[string]float64 `json:"categories"`
}

type SummaryRepository interface {
	GetSummary(from string, to string, ctx context.Context) (*Summary, error)
}
