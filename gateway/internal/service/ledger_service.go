package service

import (
	"context"
	"gateway/internal/model"
	ledgerv1 "gateway/internal/pb/ledger/v1"
)

type LedgerGatewayService interface {
	BudgetAdd(ctx context.Context, req model.BudgetAdd) error
	BudgetGet(ctx context.Context, req model.BudgetGet) (*model.BudgetGetResponse, error)
	BudgetList(ctx context.Context) ([]model.BudgetGetResponse, error)
	TransactionAdd(ctx context.Context, req model.TrasnactionAdd) (*model.TransactionAddResponse, error)
	TransactionGet(ctx context.Context, req model.TransactionGet) (*model.TransactionGetResponse, error)
	TransactionList(ctx context.Context) ([]model.TransactionGetResponse, error)
	TransactionBulkAdd(ctx context.Context, req model.TransactionBulkAdd) (*model.TransactionBulkAddResponse, error)
	ReportSummary(ctx context.Context, req model.ReportSummary) (*model.ReportSummaryResponse, error)
}

type ledgerGatewayService struct {
	client ledgerv1.LedgerServiceClient
}

func NewLedgerGatewayService(client ledgerv1.LedgerServiceClient) LedgerGatewayService {
	if client == nil {
		panic("budget gateway service requires gRPC client")
	}
	return &ledgerGatewayService{client: client}
}

func (l *ledgerGatewayService) BudgetAdd(ctx context.Context, req model.BudgetAdd) error {
	_, err := l.client.BudgetAdd(ctx, &ledgerv1.BudgetAddRequest{
		Category: req.Category,
		Limit:    float32(req.Limit),
	})
	return err
}

func (l *ledgerGatewayService) BudgetGet(ctx context.Context, req model.BudgetGet) (*model.BudgetGetResponse, error) {
	resp, err := l.client.BudgetGet(ctx, &ledgerv1.BudgetGetRequest{
		Category: req.Category,
	})
	if err != nil {
		return nil, err
	}
	return &model.BudgetGetResponse{
		Category: resp.Category,
		Limit:    float64(resp.Limit),
	}, nil
}

func (l *ledgerGatewayService) BudgetList(ctx context.Context) ([]model.BudgetGetResponse, error) {
	resp, err := l.client.BudgetsList(ctx, nil)
	if err != nil {
		return nil, err
	}
	out := make([]model.BudgetGetResponse, 0, len(resp.GetBudgets()))
	for _, r := range resp.GetBudgets() {
		if r == nil {
			continue
		}
		out = append(out, model.BudgetGetResponse{
			Category: r.Category,
			Limit:    float64(r.Limit),
		})
	}
	return out, nil
}

func (l *ledgerGatewayService) TransactionAdd(ctx context.Context, req model.TrasnactionAdd) (*model.TransactionAddResponse, error) {
	resp, err := l.client.TransactionAdd(ctx, &ledgerv1.TransactionAddRequest{
		Amount:      float32(req.Amount),
		Category:    req.Category,
		Description: req.Description,
		Date:        req.Date,
	})
	if err != nil {
		return nil, err
	}
	return &model.TransactionAddResponse{
		Id: resp.GetId(),
	}, nil
}

func (l *ledgerGatewayService) TransactionGet(ctx context.Context, req model.TransactionGet) (*model.TransactionGetResponse, error) {
	resp, err := l.client.TransactionGet(ctx, &ledgerv1.TransactionGetRequest{
		Id: req.Id,
	})
	if err != nil {
		return nil, err
	}
	return &model.TransactionGetResponse{
		Id:          resp.GetId(),
		Amount:      float64(resp.GetAmount()),
		Category:    resp.GetCategory(),
		Description: resp.GetDescription(),
		Date:        resp.GetDate(),
	}, nil
}

func (l *ledgerGatewayService) TransactionList(ctx context.Context) ([]model.TransactionGetResponse, error) {
	resp, err := l.client.TransactionList(ctx, nil)
	if err != nil {
		return nil, err
	}
	out := make([]model.TransactionGetResponse, 0, len(resp.GetTransactions()))
	for _, tr := range resp.GetTransactions() {
		if tr == nil {
			continue
		}
		out = append(out, model.TransactionGetResponse{
			Id:          tr.GetId(),
			Amount:      float64(tr.GetAmount()),
			Category:    tr.GetCategory(),
			Description: tr.GetDescription(),
			Date:        tr.GetDate(),
		})
	}
	return out, nil
}

func (l *ledgerGatewayService) ReportSummary(ctx context.Context, req model.ReportSummary) (*model.ReportSummaryResponse, error) {
	resp, err := l.client.ReportSummary(ctx, &ledgerv1.SummaryRequest{From: req.From, To: req.To})
	if err != nil {
		return nil, err
	}
	return &model.ReportSummaryResponse{Report: resp.Report, CacheResult: resp.CacheResult}, nil
}

func (l *ledgerGatewayService) TransactionBulkAdd(ctx context.Context, req model.TransactionBulkAdd) (*model.TransactionBulkAddResponse, error) {
	trs := make([]*ledgerv1.TransactionAddRequest, 0, len(req.Transactions))
	for _, transaction := range req.Transactions {
		trs = append(trs, &ledgerv1.TransactionAddRequest{
			Amount:      float32(transaction.Amount),
			Category:    transaction.Category,
			Date:        transaction.Date,
			Description: transaction.Description,
		})
	}
	resp, err := l.client.BulkAddTransactions(ctx, &ledgerv1.TransactionBulkAddRequest{Transactions: trs})
	if err != nil {
		return nil, err
	}
	return &model.TransactionBulkAddResponse{
		Accepted: resp.Accepted,
		Rejected: resp.Rejected,
		Errors:   resp.Errors,
	}, nil
}
