package main

import (
	"auth/internal/db"
	"auth/internal/grpcserver"
	"auth/internal/repository/pg"
	"auth/internal/service"
	"context"
	"errors"
	"log"
	"net"
	"os"

	pb "auth/internal/pb/auth/v1"

	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

func main() {
	println("Start auth service")

	grpcPort := os.Getenv("GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "50052"
	}
	grpcAddr := ":" + grpcPort

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		println("need JWT_TOKEN")
		return
	}

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

	userRepo := pg.NewUserPgRepository(dbConn)
	authService := service.NewAuthService(userRepo, jwtSecret)
	authServer := grpcserver.NewAuthServer(authService)

	pb.RegisterAuthServiceServer(grpcSrv, authServer)

	reflection.Register(grpcSrv)

	healthSrv.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)

	g, gctx := errgroup.WithContext(context.Background())

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
