package pg

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"ledger/internal/domain"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

type SummaryPgRepository struct {
	db    *sql.DB
	cache *redis.Client
}

func NewSummaryPgRepository(db *sql.DB, cache *redis.Client) *SummaryPgRepository {
	return &SummaryPgRepository{
		db:    db,
		cache: cache,
	}
}

func (r *SummaryPgRepository) GetSummary(ctx context.Context, from string, to string) (*domain.Summary, error) {
	key := "report:summary:" + from + ":" + to
	val, err := r.cache.Get(ctx, key).Result()
	if err == nil {
		println("Get result from cache")
		var result map[string]float64
		if err := json.Unmarshal([]byte(val), &result); err == nil {
			return &domain.Summary{Categories: result, CacheResult: true}, nil
		}
	}
	println("Get result from db")
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
	data, _ := json.Marshal(result)
	print("save result to cache")
	r.cache.Set(ctx, key, data, 30*time.Second)
	return &domain.Summary{Categories: result, CacheResult: false}, nil
}
