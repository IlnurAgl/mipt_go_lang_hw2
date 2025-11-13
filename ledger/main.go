package ledger

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"ledger/internal/cache"
	"ledger/internal/db"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

var mu sync.Mutex

type Validatable interface {
	Validate() error
}

type Transaction struct {
	ID          int64
	Amount      float64
	Category    string
	Description string
	Date        string
}

func (tx *Transaction) Validate() error {
	if tx.Amount <= 0 {
		return errors.New("invalid amount")
	}
	if tx.Category == "" {
		return errors.New("invalid category")
	}
	return nil
}

var transactionId int64
var dbConn *sql.DB
var cacheConn *redis.Client

type Budget struct {
	Category string  `json:"category"`
	Limit    float64 `json:"limit"`
}

func (budget *Budget) Validate() error {
	if budget.Limit <= 0 {
		return errors.New("invalid limit")
	}
	if budget.Category == "" {
		return errors.New("invalid category")
	}
	return nil
}

func AddTransaction(transaction Transaction) (int64, error) {
	err := transaction.Validate()
	if err != nil {
		println(err.Error())
		return 0, err
	}
	var limit float64
	err = dbConn.QueryRow("SELECT limit_amount FROM budgets WHERE category=$1", transaction.Category).Scan(&limit)
	if err != nil {
		return 0, err
	}
	if limit == 0 {
		return 0, errors.New("invalid limit")
	}
	var totalAmount float64
	err = dbConn.QueryRow("SELECT COALESCE(SUM(amount),0) FROM expenses WHERE category=$1", transaction.Category).Scan(&totalAmount)
	if err != nil {
		return 0, err
	}
	if totalAmount+transaction.Amount > limit {
		return 0, errors.New("budget exceeded")
	}
	var newID int64
	err = dbConn.QueryRow("INSERT INTO expenses(amount, category, description, date) VALUES($1,$2,$3,$4) RETURNING id", transaction.Amount, transaction.Category, transaction.Description, transaction.Date).Scan(&newID)
	if err != nil {
		return 0, err
	}
	return newID, nil
}

func ListTransactions() ([]Transaction, error) {
	rows, err := dbConn.Query("SELECT id, amount, category, description, date FROM expenses ORDER BY date DESC, id DESC")
	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			println(err.Error())
		}
	}(rows)
	var dbTransactions []Transaction
	for rows.Next() {
		var t Transaction
		if err := rows.Scan(&t.ID, &t.Amount, &t.Category, &t.Description, &t.Date); err != nil {
			return dbTransactions, err
		}
		dbTransactions = append(dbTransactions, t)
	}
	if err = rows.Err(); err != nil {
		return dbTransactions, err
	}
	return dbTransactions, nil
}

func SetBudget(b Budget) error {
	err := b.Validate()
	if err != nil {
		return err
	}
	_, err = dbConn.Exec("INSERT INTO budgets(category, limit_amount) VALUES($1,$2) ON CONFLICT(category) DO UPDATE SET limit_amount =EXCLUDED.limit_amount", b.Category, b.Limit)
	if err != nil {
		return err
	}
	return nil
}

func ListBudgets() (map[string]Budget, error) {
	rows, err := dbConn.Query("SELECT category, limit_amount FROM budgets ORDER BY category")
	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			println(err.Error())
		}
	}(rows)
	dbBudgets := make(map[string]Budget)
	for rows.Next() {
		var b Budget
		if err := rows.Scan(&b.Category, &b.Limit); err != nil {
			return dbBudgets, err
		}
		dbBudgets[b.Category] = b
	}
	if err = rows.Err(); err != nil {
		return dbBudgets, err
	}
	return dbBudgets, nil
}

func GetReportSummary(from string, to string) (map[string]float64, error) {
	ctx := context.Background()
	result, err := cacheConn.Get(ctx, fmt.Sprintf("report:summary:%s:%s", from, to)).Result()
	if err != nil || result == "" {
		rows, err := dbConn.Query("SELECT category, SUM(amount) FROM expenses where date >= $1 and date <= $2 group by category", from, to)
		if err != nil {
			return nil, err
		}
		defer func(rows *sql.Rows) {
			err := rows.Close()
			if err != nil {
				println(err.Error())
			}
		}(rows)
		categories := make(map[string]float64)
		for rows.Next() {
			var category string
			var amount float64
			if err := rows.Scan(&category, &amount); err != nil {
				return categories, err
			}
			categories[category] = amount
		}
		jsonBytes, err := json.Marshal(categories)
		if err != nil {
			return categories, nil
		}
		_, err = cacheConn.Set(ctx, fmt.Sprintf("report:summary:%s:%s", from, to), string(jsonBytes), time.Second*30).Result()
		if err != nil {
			return categories, nil
		}
		return categories, nil
	}
	var res map[string]float64
	err = json.Unmarshal([]byte(result), &res)
	println("from cache")
	if err != nil {
		return nil, err
	}
	return res, nil
}

func InitConnection() error {
	if dbConn == nil {
		var err error
		dbConn, err = db.Connect()
		if err != nil {
			return err
		}
	}
	if cacheConn == nil {
		println("cache conn nil")
		var err error
		cacheConn, err = cache.Connect()
		if err != nil {
			return err
		}
		return nil
	}
	return nil
}
