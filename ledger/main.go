package ledger

import (
	"ledger/internal/db"
	"ledger/internal/domain"
	"ledger/internal/repository/pg"
	"ledger/internal/service"
)

type LedgerService interface {
	SetBudget(category string, limit float64) error
	GetBudgets() (map[string]domain.Budget, error)
	AddTransaction(Amount float64, Category string, Description string, Date string) (int64, error)
	ListTransactions() ([]domain.Transaction, error)
}

func NewLedgerService() (LedgerService, func(), error) {
	dbConn, err := db.Connect()
	if err != nil {
		return nil, nil, err
	}
	closeFunc := func() {
		err := dbConn.Close()
		if err != nil {
			return
		}
	}
	return &service.LedgerServiceImpl{
			pg.NewBudgetPgRepository(dbConn),
			pg.NewTransactionPgRepository(dbConn),
		},
		closeFunc,
		nil
}
