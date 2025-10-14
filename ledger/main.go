package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
)

type Transaction struct {
	ID          int64
	Amount      float64
	Category    string
	Description string
	Date        string
}

var transactions []Transaction

var budgets map[string]Budget

type Budget struct {
	Category string  `json:"category"`
	Limit    float64 `json:"limit"`
	Period   string  `json:"period"`
}

func AddTransaction(transaction Transaction) error {
	if transaction.Amount == 0 {
		return errors.New("invalid amount")
	}
	var sum float64 = 0
	for _, tr := range transactions {
		if tr.Category == transaction.Category {
			sum += tr.Amount
		}
	}
	if sum+transaction.Amount > budgets[transaction.Category].Limit {
		return errors.New("budget exceeded")
	}
	transactions = append(transactions, transaction)
	return nil
}

func ListTransactions() []Transaction {
	return transactions
}

func SetBudget(b Budget) {
	budgets[b.Category] = b
}

func LoadBudgets(r io.Reader) error {
	data, err := io.ReadAll(r)
	if err != nil {
		return fmt.Errorf("ошибка чтения данных: %w", err)
	}

	var budgetList []Budget
	if err := json.Unmarshal(data, &budgetList); err != nil {
		return fmt.Errorf("ошибка парсинга JSON: %w", err)
	}

	// Добавляем каждый бюджет через SetBudget
	for _, budget := range budgetList {
		SetBudget(budget)
	}
	return nil
}

func main() {
	file, err := os.Open("budgets.json")
	if err != nil {
		println(fmt.Errorf("ошибка открытия файла: %w", err))
		return
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			return
		}
	}(file)
	reader := bufio.NewReader(file)
	budgets = make(map[string]Budget)
	err = LoadBudgets(reader)
	if err != nil {
		return
	}
	SetBudget(Budget{Limit: 5000, Period: "monthly", Category: "еда"})
	SetBudget(Budget{Limit: 2000, Period: "monthly", Category: "транспорт"})

	err = AddTransaction(Transaction{ID: 1, Amount: 3000, Category: "еда", Description: "test", Date: "2025-10-07T23:59:59"})
	if err != nil {
		fmt.Println(err)
	}
	err = AddTransaction(Transaction{ID: 2, Amount: 4000, Category: "еда", Description: "test", Date: "2025-10-07T23:59:59"})
	if err != nil {
		fmt.Println(err)
	}
	err = AddTransaction(Transaction{ID: 3, Amount: 1500, Category: "транспорт", Description: "test", Date: "2025-10-07T23:59:59"})
	if err != nil {
		fmt.Println(err)
	}
	for _, transaction := range ListTransactions() {
		fmt.Println(transaction)
	}
}
