package main

import (
	"context"
	"errors"
	"ledger/internal/cache"
	"ledger/internal/db"
	"ledger/internal/grpcserver"
	"ledger/internal/repository/pg"
	"ledger/internal/service"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	pb "ledger/internal/pb/ledger/v1"

	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

func main() {
	println("Start app")
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	grpcPort := os.Getenv("GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "9090"
	}
	grpcAddr := ":" + grpcPort
	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		log.Fatalf("listen gRPC: %v", err)
	}

	grpcSrv := grpc.NewServer()

	healthSrv := health.NewServer()
	grpc_health_v1.RegisterHealthServer(grpcSrv, healthSrv)

	println("Try connect to db")
	dbConn, err := db.Connect()
	if err != nil {
		println("Db connect error: %v", err.Error())
		healthSrv.SetServingStatus("", grpc_health_v1.HealthCheckResponse_NOT_SERVING)
		return
	}
	defer dbConn.Close()
	println("Db connected")

	redisConn, err := cache.Connect()
	if err != nil {
		println("Redis conntect error: %v", err.Error())
		healthSrv.SetServingStatus("", grpc_health_v1.HealthCheckResponse_NOT_SERVING)
		return
	}
	defer redisConn.Close()

	ledgerService := service.NewLedgerService(
		pg.NewBudgetPgRepository(dbConn),
		pg.NewTransactionPgRepository(dbConn),
		pg.NewSummaryPgRepository(dbConn, redisConn),
	)
	pb.RegisterLedgerServiceServer(grpcSrv, grpcserver.NewLedgerServer(ledgerService))

	reflection.Register(grpcSrv)

	healthSrv.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)

	g, gctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		log.Printf("gRPC server listening on %s", grpcAddr)
		if err := grpcSrv.Serve(lis); err != nil {
			return err
		}
		return nil
	})

	g.Go(func() error {
		<-gctx.Done()
		grpcSrv.GracefulStop()
		return nil
	})

	if err := g.Wait(); err != nil && !errors.Is(err, context.Canceled) {
		log.Fatalf("server error: %v", err)
	}
}
