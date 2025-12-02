package ledger

import (
	"context"
	"ledger/internal/db"
	"ledger/internal/domain"
	"ledger/internal/repository/pg"
	"ledger/internal/service"
)

type Transaction = domain.Transaction

type LedgerService interface {
	SetBudget(category string, limit float64, ctx context.Context) error
	GetBudgets(ctx context.Context) (map[string]domain.Budget, error)
	AddTransaction(Amount float64, Category string, Description string, Date string, ctx context.Context) (int64, error)
	ListTransactions(ctx context.Context) ([]domain.Transaction, error)
	GetReportSummary(from string, to string, ctx context.Context) (*domain.Summary, error)
	BulkAddTransactions(transactions []domain.Transaction, numWorkers int, ctx context.Context) (*service.BulkResult, error)
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
			pg.NewSummaryPgRepository(dbConn),
		},
		closeFunc,
		nil
}
