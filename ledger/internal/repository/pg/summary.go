package pg

import (
	"context"
	"database/sql"
	"fmt"
	"ledger/internal/domain"
	"sync"
	"time"
)

type SummaryPgRepository struct {
	db *sql.DB
}

func NewSummaryPgRepository(db *sql.DB) *SummaryPgRepository {
	return &SummaryPgRepository{
		db: db,
	}
}

func (r *SummaryPgRepository) GetSummary(from string, to string, ctx context.Context) (*domain.Summary, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT distinct category FROM expenses WHERE date BETWEEN $1 AND $2", from, to)
	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			println(err.Error())
		}
	}(rows)
	var categories []string
	for rows.Next() {
		var c string
		if err := rows.Scan(&c); err != nil {
			println(err.Error())
			return nil, err
		}
		categories = append(categories, c)
	}
	if err = rows.Err(); err != nil {
		println(err.Error())
		return nil, err
	}
	var wg sync.WaitGroup
	result := make(map[string]float64)
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
	done := make(chan bool)

	go func() {
		for {
			select {
			case <-done:
				fmt.Println("Ticker stopped")
				return
			case t := <-ticker.C:
				fmt.Println("Tick at", t)
			}
		}
	}()
	for _, category := range categories {
		wg.Go(func() {
			var amount float64
			r.db.QueryRowContext(ctx, "SELECT sum(amount) FROM expenses WHERE category = $1 AND date BETWEEN $2 AND $3", category, from, to).Scan(&amount)
			result[category] = amount
		})
	}
	wg.Wait()
	done <- true
	return &domain.Summary{Categories: result}, nil
}
