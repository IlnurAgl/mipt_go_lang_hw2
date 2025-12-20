package domain

import "context"

type Summary struct {
	Categories  map[string]float64 `json:"categories"`
	CacheResult bool
}

type SummaryRepository interface {
	GetSummary(ctx context.Context, from string, to string) (*Summary, error)
}
