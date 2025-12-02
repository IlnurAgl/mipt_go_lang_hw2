package service

import (
	"context"
	"errors"
	"fmt"
	"ledger/internal/domain"
	"sync"
)

type LedgerServiceImpl struct {
	BudgetRepository      domain.BudgetRepository
	TransactionRepository domain.TransactionRepository
	SummaryRepository     domain.SummaryRepository
}

func (l *LedgerServiceImpl) SetBudget(category string, limit float64, ctx context.Context) error {
	budget := domain.Budget{
		Category: category,
		Limit:    limit,
	}
	err := budget.Validate()
	if err != nil {
		return err
	}
	err = l.BudgetRepository.SetBudget(&budget, ctx)
	if err != nil {
		return err
	}
	return nil
}

func (l *LedgerServiceImpl) GetBudgets(ctx context.Context) (map[string]domain.Budget, error) {
	budgets, err := l.BudgetRepository.GetBudgets(ctx)
	if err != nil {
		return nil, err
	}
	budgetMap := make(map[string]domain.Budget)
	for _, budget := range budgets {
		budgetMap[budget.Category] = budget
	}
	return budgetMap, nil
}

func (l *LedgerServiceImpl) AddTransaction(Amount float64, Category string, Description string, Date string, ctx context.Context) (int64, error) {
	transaction := domain.Transaction{
		Amount:      Amount,
		Category:    Category,
		Description: Description,
		Date:        Date,
	}
	err := transaction.Validate()
	if err != nil {
		return 0, errors.New("invalid transaction")
	}
	budget, err := l.BudgetRepository.GetBudget(transaction.Category, ctx)
	if err != nil {
		return 0, err
	}
	amount, err := l.TransactionRepository.GetAmountTransactionByCategory(transaction.Category, ctx)
	if err != nil {
		return 0, err
	}
	if amount+transaction.Amount > budget.Limit {
		return 0, errors.New("budget exceeded")
	}
	id, err := l.TransactionRepository.AddTransaction(&transaction, ctx)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (l *LedgerServiceImpl) ListTransactions(ctx context.Context) ([]domain.Transaction, error) {
	transactions, err := l.TransactionRepository.ListTransactions(ctx)
	if err != nil {
		return nil, err
	}
	return transactions, nil
}

func (l *LedgerServiceImpl) GetReportSummary(from string, to string, ctx context.Context) (*domain.Summary, error) {
	return l.SummaryRepository.GetSummary(from, to, ctx)
}

type Result struct {
	success  bool
	errorStr string
	index    int
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
			budget, err := r.BudgetRepository.GetBudget(transaction.Category, ctx)
			if err != nil {
				if err.Error() == "sql: no rows in result set" {
					results <- Result{success: false, errorStr: "no budget category", index: job.index}
					break
				}
				results <- Result{success: false, errorStr: err.Error(), index: job.index}
				break
			}
			amount, err := r.TransactionRepository.GetAmountTransactionByCategory(transaction.Category, ctx)
			if err != nil {
				results <- Result{success: false, errorStr: err.Error(), index: job.index}
				break
			}
			if amount+transaction.Amount > budget.Limit {
				results <- Result{success: false, errorStr: "budget exceeded", index: job.index}
				break
			}
			_, err = r.TransactionRepository.AddTransaction(&transaction, ctx)
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

type BulkResult struct {
	Accepted int64
	Rejected int64
	Errors   map[int]string
	m        sync.Mutex
}

type WorkerJob struct {
	index       int
	transaction domain.Transaction
}

func (r *LedgerServiceImpl) BulkAddTransactions(transactions []domain.Transaction, numWorkers int, ctx context.Context) (*BulkResult, error) {
	jobs := make(chan WorkerJob, len(transactions))
	results := make(chan Result, len(transactions))
	var wg sync.WaitGroup
	for w := range numWorkers {
		wg.Add(1)
		go r.worker(w, jobs, results, ctx, &wg)
	}

	go func() {
		for j := 0; j < len(transactions); j++ {
			jobs <- WorkerJob{index: j, transaction: transactions[j]}
		}
		close(jobs)
	}()

	go func() {
		wg.Wait()
		close(results)
	}()
	res := BulkResult{Accepted: 0, Rejected: 0}
	res.Errors = make(map[int]string)

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
