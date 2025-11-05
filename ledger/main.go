package ledger

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"
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

var transactions []Transaction

var budgets map[string]Budget
var transactionId int64

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

func AddTransaction(transaction Transaction) error {
	err := transaction.Validate()
	if err != nil {
		println(err.Error())
		return err
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

func SetBudget(b Budget) error {
	if budgets == nil {
		budgets = make(map[string]Budget)
	}
	err := b.Validate()
	if err != nil {
		return err
	}
	budgets[b.Category] = b
	return nil
}

func ListBudgets() map[string]Budget {
	return budgets
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
		err := SetBudget(budget)
		if err != nil {
			return err
		}
	}
	return nil
}

func CheckValid(v Validatable) error {
	return v.Validate()
}

func Reset() {
	transactions = []Transaction{}
}

func GetTransactionId() int64 {
	mu.Lock()
	transactionId++
	mu.Unlock()
	return transactionId
}

func ReadBudget() {
	transactionId = 0
	file, err := os.Open("ledger/budgets.json")
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
		println("Some error")
		return
	}
}

//t := Transaction{ID: 1, Amount: 3000}
//err = CheckValid(&t)
//if err != nil {
//	println(err.Error())
//}
//b := Budget{Limit: 0}
//err = CheckValid(&b)
//if err != nil {
//	println(err.Error())
//}

//err = AddTransaction(Transaction{ID: 1, Amount: 3000, Category: "еда", Description: "test", Date: "2025-10-07T23:59:59"})
//if err != nil {
//	fmt.Println(err)
//}
//err = AddTransaction(Transaction{ID: 2, Amount: 4000, Category: "еда", Description: "test", Date: "2025-10-07T23:59:59"})
//if err != nil {
//	fmt.Println(err)
//}
//err = AddTransaction(Transaction{ID: 3, Amount: 1500, Category: "транспорт", Description: "test", Date: "2025-10-07T23:59:59"})
//if err != nil {
//	fmt.Println(err)
//}
//for _, transaction := range ListTransactions() {
//	fmt.Println(transaction)
//}
//}
