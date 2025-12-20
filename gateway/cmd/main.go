package main

import (
	"context"
	"fmt"
	"gateway/internal/config"
	"gateway/internal/handler"
	authv1 "gateway/internal/pb/auth/v1"
	ledgerv1 "gateway/internal/pb/ledger/v1"
	"gateway/internal/server/httpserver"
	"gateway/internal/service"
	"log"
	"net/http"
	"os/signal"
	"strings"
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

func authMiddleware(authConn authv1.AuthServiceClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.URL.Path == "/ping" {
			c.Next()
			return
		}
		token := c.GetHeader("Authorization")
		if token == "" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Invalid auth token"})
			c.Abort()
			return
		}
		has := strings.HasPrefix(token, "Bearer")
		if !has {
			c.JSON(http.StatusForbidden, gin.H{"error": "Invalid auth token"})
			c.Abort()
			return
		}
		resp, err := authConn.ValidateToken(c, &authv1.ValidateTokenRequest{Token: token[7:]})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Auth server down"})
			c.Abort()
			return
		}
		if !resp.Valid {
			c.JSON(http.StatusForbidden, gin.H{"error": "Invalid auth token"})
			c.Abort()
			return
		}
		c.Next()
	}
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg := config.Load()
	log.Printf("Connecting to gRPC backend at %s", cfg.GRPC.LedgerAddress)
	conn, err := grpcClient(cfg.GRPC.LedgerAddress)
	if err != nil {
		log.Fatalf("failed to connect to gRPC backend: %v", err)
	}
	defer conn.Close()
	log.Printf("Connected to gRPC backend successfully")

	log.Printf("Connecting to gRPC auth backend at %s", cfg.GRPC.AuthAddress)
	authConn, err := grpcClient(cfg.GRPC.AuthAddress)
	if err != nil {
		log.Fatalf("failed to connect to gRPC auth backend: %v", err)
	}
	defer authConn.Close()
	log.Printf("Connected to gRPC auth backend successfully")

	ledgerService := service.NewLedgerGatewayService(ledgerv1.NewLedgerServiceClient(conn))
	ledgerHandler := handler.NewLedgerHandler(ledgerService)
	engine := gin.New()
	engine.Use(gin.Logger(), gin.Recovery(), loggingMiddleware(), timeoutMiddleware(), authMiddleware(authv1.NewAuthServiceClient(authConn)))
	engine.GET("/ping", ping)
	engine.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    http.StatusNotFound,
			"message": "Not found",
		})
	})
	api := engine.Group("/api")
	ledgerHandler.Register(api)

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
