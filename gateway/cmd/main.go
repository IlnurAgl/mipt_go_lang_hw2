package main

import (
	"context"
	"fmt"
	"gateway/internal/config"
	"gateway/internal/server/httpserver"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/timeout"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func ping(c *gin.Context) {
	c.String(http.StatusOK, "pong")
}

func loggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		fmt.Printf("Request received: %s %s\n", c.Request.Method, c.Request.URL.Path)
		t := time.Now()
		c.Next()
		fmt.Printf("Request handled: %s %s, time: %s\n", c.Request.Method, c.Request.URL.Path, time.Since(t))
	}
}

func timeoutResponse(c *gin.Context) {
	c.String(http.StatusRequestTimeout, "timeout")
}

func timeoutMiddleware() gin.HandlerFunc {
	return timeout.New(
		timeout.WithTimeout(2*time.Second),
		timeout.WithResponse(timeoutResponse),
	)
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg := config.Load()
	log.Printf("Connecting to gRPC backend at %s", cfg.GRPC.Address)
	conn, err := grpcClient(cfg.GRPC.Address)
	if err != nil {
		log.Fatalf("failed to connect to gRPC backend: %v", err)
	}
	defer conn.Close()
	log.Printf("Connected to gRPC backend successfully")

	// service, closeDbFn, closeRedisFn, err := ledger.NewLedgerService()
	// defer closeDbFn()
	// defer closeRedisFn()
	// if err != nil {
	// 	println(err.Error())
	// 	return
	// }
	// trs := handlers.NewTransactionHandlers(service)
	// budgetHandlers := handlers.NewBudgetHandlers(service)

	// r := mux.NewRouter()
	// r.Use(loggingMiddleware)
	// r.Use(contextMiddleware)
	// r.HandleFunc("/ping", ping)
	// r.HandleFunc("/api/transaction", trs.TransactionHandler)
	// r.HandleFunc("/api/budget", budgetHandlers.BudgetHandler)
	// r.HandleFunc("/api/reports/summary", handlers.ReportHandler(service))
	// r.HandleFunc("/api/transactions/bulk", trs.TransactionBulkHandler)

	engine := gin.New()
	engine.Use(gin.Logger(), gin.Recovery(), loggingMiddleware(), timeoutMiddleware())
	engine.GET("/ping", ping)

	server := httpserver.New(cfg.HTTP, engine)

	go func() {
		log.Printf("HTTP server listening on %s", cfg.HTTP.Address)
		if err := server.Start(); err != nil {
			log.Fatalf("http server stopped: %v", err)
		}
	}()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.HTTP.ShutdownTimeout)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("graceful shutdown failed: %v", err)
	}
}

func grpcClient(address string) (*grpc.ClientConn, error) {
	conn, err := grpc.NewClient(
		address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}
	return conn, nil
}
