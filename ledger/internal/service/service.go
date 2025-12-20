package service

import (
	"context"
	"errors"
	"fmt"
	"ledger/internal/domain"
	"sync"
)

type LedgerService interface {
	BudgetAdd(ctx context.Context, budget *domain.Budget) error
	BudgetGet(ctx context.Context, category string) (*domain.Budget, error)
	BudgetsList(ctx context.Context) (map[string]domain.Budget, error)
	TransactionAdd(ctx context.Context, transaction *domain.Transaction) (int64, error)
	TransactionGet(ctx context.Context, id int64) (*domain.Transaction, error)
	TransactionsList(ctx context.Context) ([]domain.Transaction, error)
	BulkAddTransactions(ctx context.Context, transactions []domain.Transaction, numWorkers int) (*BulkTransactionResult, error)
	GetReportSummary(ctx context.Context, from string, to string) (*domain.Summary, error)
}

type LedgerServiceImpl struct {
	budgetRepository      domain.BudgetRepository
	transactionRepository domain.TransactionRepository
	summaryRepository     domain.SummaryRepository
}

var _ LedgerService = (*LedgerServiceImpl)(nil)

func NewLedgerService(budgetRepository domain.BudgetRepository, transactionsRepository domain.TransactionRepository, summaryRepository domain.SummaryRepository) *LedgerServiceImpl {
	return &LedgerServiceImpl{
		budgetRepository:      budgetRepository,
		transactionRepository: transactionsRepository,
		summaryRepository:     summaryRepository,
	}
}

func (l *LedgerServiceImpl) BudgetAdd(ctx context.Context, budget *domain.Budget) error {
	err := budget.Validate()
	if err != nil {
		return err
	}
	err = l.budgetRepository.SetBudget(budget, ctx)
	if err != nil {
		return err
	}
	return nil
}

func (l *LedgerServiceImpl) BudgetGet(ctx context.Context, category string) (*domain.Budget, error) {
	budget, err := l.budgetRepository.GetBudget(category, ctx)
	if err != nil {
		return nil, err
	}
	return budget, nil
}

func (l *LedgerServiceImpl) BudgetsList(ctx context.Context) (map[string]domain.Budget, error) {
	budgets, err := l.budgetRepository.GetBudgets(ctx)
	if err != nil {
		return nil, err
	}
	budgetMap := make(map[string]domain.Budget)
	for _, budget := range budgets {
		budgetMap[budget.Category] = budget
	}
	return budgetMap, nil
}

func (l *LedgerServiceImpl) TransactionAdd(ctx context.Context, transaction *domain.Transaction) (int64, error) {
	err := transaction.Validate()
	if err != nil {
		return 0, errors.New("invalid transaction")
	}
	budget, err := l.budgetRepository.GetBudget(transaction.Category, ctx)
	if err != nil {
		return 0, err
	}
	amount, err := l.transactionRepository.GetAmountTransactionByCategoryAndMonth(ctx, transaction.Category, transaction.Date)
	if err != nil {
		return 0, err
	}
	if amount+transaction.Amount > budget.Limit {
		return 0, errors.New("budget exceeded")
	}
	id, err := l.transactionRepository.AddTransaction(transaction, ctx)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (l *LedgerServiceImpl) TransactionGet(ctx context.Context, id int64) (*domain.Transaction, error) {
	transaction, err := l.transactionRepository.GetTransaction(id, ctx)
	if err != nil {
		return nil, err
	}
	return transaction, err
}

func (l *LedgerServiceImpl) TransactionsList(ctx context.Context) ([]domain.Transaction, error) {
	transactions, err := l.transactionRepository.ListTransactions(ctx)
	if err != nil {
		return nil, err
	}
	return transactions, nil
}

func (l *LedgerServiceImpl) GetReportSummary(ctx context.Context, from string, to string) (*domain.Summary, error) {
	return l.summaryRepository.GetSummary(ctx, from, to)
}

type Result struct {
	success  bool
	errorStr string
	index    int64
}

func (r *LedgerServiceImpl) worker(id int, jobs <-chan WorkerJob, results chan<- Result, ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case job, ok := <-jobs:
			if !ok {
				fmt.Printf("Worker %d exiting because jobs channel closed\n", id)
				return
			}
			transaction := job.transaction
			err := transaction.Validate()
			if err != nil {
				results <- Result{success: false, errorStr: "invalid transaction", index: job.index}
				break
			}
			budget, err := r.budgetRepository.GetBudget(transaction.Category, ctx)
			if err != nil {
				if err.Error() == "sql: no rows in result set" {
					results <- Result{success: false, errorStr: "no budget category", index: job.index}
					break
				}
				results <- Result{success: false, errorStr: err.Error(), index: job.index}
				break
			}
			amount, err := r.transactionRepository.GetAmountTransactionByCategory(transaction.Category, ctx)
			if err != nil {
				results <- Result{success: false, errorStr: err.Error(), index: job.index}
				break
			}
			if amount+transaction.Amount > budget.Limit {
				results <- Result{success: false, errorStr: "budget exceeded", index: job.index}
				break
			}
			_, err = r.transactionRepository.AddTransaction(&transaction, ctx)
			if err != nil {
				results <- Result{success: false, errorStr: err.Error(), index: job.index}
				break
			}
			results <- Result{success: true, errorStr: "", index: job.index}
		case <-ctx.Done():
			fmt.Printf("Worker %d exiting due to context cancellation\n", id)
			return
		}
	}
}

type BulkTransactionResult struct {
	Accepted int64
	Rejected int64
	Errors   map[int64]string
	m        sync.Mutex
}

type WorkerJob struct {
	index       int64
	transaction domain.Transaction
}

func (r *LedgerServiceImpl) BulkAddTransactions(ctx context.Context, transactions []domain.Transaction, numWorkers int) (*BulkTransactionResult, error) {
	jobs := make(chan WorkerJob, len(transactions))
	results := make(chan Result, len(transactions))
	var wg sync.WaitGroup
	for w := range numWorkers {
		wg.Add(1)
		go r.worker(w, jobs, results, ctx, &wg)
	}

	go func() {
		for j := 0; j < len(transactions); j++ {
			jobs <- WorkerJob{index: int64(j), transaction: transactions[j]}
		}
		close(jobs)
	}()

	go func() {
		wg.Wait()
		close(results)
	}()
	res := BulkTransactionResult{Accepted: 0, Rejected: 0}
	res.Errors = make(map[int64]string)

	for result := range results {
		res.m.Lock()
		if result.success {
			res.Accepted++
		} else {
			res.Rejected++
			res.Errors[result.index] = result.errorStr
		}
		res.m.Unlock()
	}

	return &res, nil
}
