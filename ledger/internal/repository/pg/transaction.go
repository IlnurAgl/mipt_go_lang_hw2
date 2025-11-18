package pg

import (
	"database/sql"
	"ledger/internal/domain"
)

type TransactionPgRepository struct {
	db *sql.DB
}

func NewTransactionPgRepository(db *sql.DB) *TransactionPgRepository {
	return &TransactionPgRepository{
		db: db,
	}
}

func (r *TransactionPgRepository) GetAmountTransactionByCategory(category string) (float64, error) {
	var totalAmount float64
	err := r.db.QueryRow("SELECT COALESCE(SUM(amount),0) FROM expenses WHERE category=$1", category).Scan(&totalAmount)
	if err != nil {
		return 0, err
	}
	return totalAmount, nil
}

func (r *TransactionPgRepository) AddTransaction(transaction *domain.Transaction) (int64, error) {
	var newID int64
	err := r.db.QueryRow("INSERT INTO expenses(amount, category, description, date) VALUES($1,$2,$3,$4) RETURNING id", transaction.Amount, transaction.Category, transaction.Description, transaction.Date).Scan(&newID)
	if err != nil {
		return 0, err
	}
	return newID, nil
}

func (r *TransactionPgRepository) ListTransactions() ([]domain.Transaction, error) {
	rows, err := r.db.Query("SELECT id, amount, category, description, date FROM expenses ORDER BY date DESC, id DESC")
	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			println(err.Error())
		}
	}(rows)
	var dbTransactions []domain.Transaction
	for rows.Next() {
		var t domain.Transaction
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
