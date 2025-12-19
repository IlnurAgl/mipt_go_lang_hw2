package grpcserver

import (
	"context"
	"ledger/internal/domain"
	pb "ledger/internal/pb/ledger/v1"
	"ledger/internal/service"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type LedgerServer struct {
	pb.UnimplementedLedgerServiceServer
	ledgerService *service.LedgerServiceImpl
}

var _ pb.LedgerServiceServer = (*LedgerServer)(nil)

func NewLedgerServer(svc *service.LedgerServiceImpl) *LedgerServer {
	return &LedgerServer{ledgerService: svc}
}

func (s *LedgerServer) BudgetAdd(ctx context.Context, req *pb.BudgetAddRequest) (*emptypb.Empty, error) {
	if req.GetCategory() == "" {
		return nil, status.Error(codes.InvalidArgument, "category is required")
	}
	if req.GetLimit() == 0 {
		return nil, status.Error(codes.InvalidArgument, "limit is required")
	}
	budget := domain.Budget{
		Category: req.GetCategory(),
		Limit:    float64(req.GetLimit()),
	}
	err := s.ledgerService.BudgetAdd(ctx, &budget)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "set budget: %v", err)
	}
	return &emptypb.Empty{}, nil
}

func (s *LedgerServer) BudgetGet(ctx context.Context, req *pb.BudgetGetRequest) (*pb.BudgetGetResponse, error) {
	if req.GetCategory() == "" {
		return nil, status.Error(codes.InvalidArgument, "category is required")
	}
	resp, err := s.ledgerService.BudgetGet(ctx, req.GetCategory())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "get budget: %v", err)
	}
	return &pb.BudgetGetResponse{
		Category: resp.Category,
		Limit:    float32(resp.Limit),
	}, nil
}

func (s *LedgerServer) BudgetsList(ctx context.Context, _ *emptypb.Empty) (*pb.BudgetGetListResponse, error) {
	budgets, err := s.ledgerService.BudgetsList(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "get budgets: %v", err)
	}

	pbBudgets := make([]*pb.BudgetGetResponse, 0)
	for _, budget := range budgets {
		pbBudgets = append(pbBudgets, &pb.BudgetGetResponse{
			Category: budget.Category,
			Limit:    float32(budget.Limit),
		})
	}

	return &pb.BudgetGetListResponse{
		Budgets: pbBudgets,
	}, nil
}

func (s *LedgerServer) TransactionAdd(ctx context.Context, req *pb.TransactionAddRequest) (*pb.TransactionAddResponse, error) {
	if req.GetAmount() == 0 {
		return nil, status.Error(codes.InvalidArgument, "amount is required")
	}
	if req.GetCategory() == "" {
		return nil, status.Error(codes.InvalidArgument, "category is required")
	}
	if req.GetDate() == "" {
		return nil, status.Error(codes.InvalidArgument, "date is required")
	}
	if req.GetDescription() == "" {
		return nil, status.Error(codes.InvalidArgument, "description is required")
	}
	tr := domain.Transaction{
		Amount:      float64(req.GetAmount()),
		Category:    req.GetCategory(),
		Date:        req.GetDate(),
		Description: req.GetDescription(),
	}
	id, err := s.ledgerService.TransactionAdd(ctx, &tr)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "add transaction: %v", err)
	}
	return &pb.TransactionAddResponse{
		Id: id,
	}, nil
}

func (s *LedgerServer) TransactionGet(ctx context.Context, req *pb.TransactionGetRequest) (*pb.TransactionGetResponse, error) {
	if req.GetId() == 0 {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}
	tr, err := s.ledgerService.TransactionGet(ctx, req.GetId())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "get transaction: %v", err)
	}
	return &pb.TransactionGetResponse{
		Id:          tr.ID,
		Amount:      float32(tr.Amount),
		Category:    tr.Category,
		Date:        tr.Date,
		Description: tr.Description,
	}, nil
}

func (s *LedgerServer) TransactionList(ctx context.Context, req *emptypb.Empty) (*pb.TransactionGetListResponse, error) {
	trs, err := s.ledgerService.TransactionsList(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list transactions: %v", err)
	}
	pbTrs := make([]*pb.TransactionGetResponse, len(trs))
	for i, tr := range trs {
		pbTrs[i] = &pb.TransactionGetResponse{
			Id:          tr.ID,
			Amount:      float32(tr.Amount),
			Category:    tr.Category,
			Date:        tr.Date,
			Description: tr.Description,
		}
	}

	return &pb.TransactionGetListResponse{
		Transactions: pbTrs,
	}, nil
}
