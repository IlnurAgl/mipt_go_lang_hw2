package main

import (
	"errors"
	"fmt"
)

type Transaction struct {
	ID          int64
	Amount      float64
	Category    string
	Description string
	Date        string
}

var transactions []Transaction

func AddTransaction(transaction Transaction) error {
	if transaction.Amount == 0 {
		return errors.New("invalid amount")
	}
	transactions = append(transactions, transaction)
	return nil
}

func ListTransactions() []Transaction {
	return transactions
}

func main() {
	_ = AddTransaction(Transaction{ID: 1, Amount: 3, Category: "A", Description: "test", Date: "2025-10-07T23:59:59"})
	_ = AddTransaction(Transaction{ID: 2, Amount: 6, Category: "B", Description: "test", Date: "2025-10-07T23:59:59"})
	_ = AddTransaction(Transaction{ID: 3, Amount: 4, Category: "C", Description: "test", Date: "2025-10-07T23:59:59"})
	for _, transaction := range ListTransactions() {
		fmt.Println(transaction)
	}
}
