package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	_ "auction-service/docs"

	"auction-service/delivery/api"
	"auction-service/global"
	"auction-service/infrastructure"
	"auction-service/manager"

	"github.com/hibiken/asynq"
)

//	@title						Auction Service API
//	@version					1.0.0
//	@description				Auction Service REST API
//	@host						localhost:8080
//	@BasePath					/
//	@securityDefinitions.apikey	BearerAuth
//	@in							header
//	@name						Authorization
//	@description				Type "Bearer" followed by a space and JWT token.
func main() {
	container := manager.NewContainer()

	// Cleanup orphaned tmp files older than 1 hour every 30 minutes
	go func() {
		ticker := time.NewTicker(30 * time.Minute)
		defer ticker.Stop()
		tmpDir := fmt.Sprintf("%s/tmp", global.GetConfig().StorageDir)
		for range ticker.C {
			_ = filepath.Walk(tmpDir, func(path string, info os.FileInfo, err error) error {
				if err != nil || info.IsDir() {
					return nil
				}
				if time.Since(info.ModTime()) > time.Hour {
					if removeErr := os.Remove(path); removeErr == nil {
						log.Printf("cleaned up orphaned tmp file: %s", path)
					}
				}
				return nil
			})
		}
	}()

	router := api.NewRouter(container)

	// ── Auction lifecycle worker (Redis-backed, ~500 ms scheduling precision) ──
	asynqSrv := infrastructure.NewAsynqServer(global.GetConfig().Redis)
	mux := asynq.NewServeMux()
	mux.HandleFunc(infrastructure.TypeAuctionStart, func(ctx context.Context, t *asynq.Task) error {
		var p infrastructure.AuctionTaskPayload
		if err := json.Unmarshal(t.Payload(), &p); err != nil {
			return fmt.Errorf("auction:start invalid payload: %w", err)
		}
		return container.UseCaseManager().AuctionUseCase().HandleStartAuction(ctx, p.AuctionId)
	})
	mux.HandleFunc(infrastructure.TypeAuctionClose, func(ctx context.Context, t *asynq.Task) error {
		var p infrastructure.AuctionTaskPayload
		if err := json.Unmarshal(t.Payload(), &p); err != nil {
			return fmt.Errorf("auction:close invalid payload: %w", err)
		}
		if err := container.UseCaseManager().AuctionUseCase().HandleCloseAuction(ctx, p.AuctionId); err != nil {
			return err
		}
		// Create the initial payment for the winner (no-op if no bids).
		return container.UseCaseManager().PaymentUseCase().CreateInitialPaymentForWinner(ctx, p.AuctionId)
	})
	mux.HandleFunc(infrastructure.TypePaymentExpiry, func(ctx context.Context, t *asynq.Task) error {
		var p infrastructure.PaymentTaskPayload
		if err := json.Unmarshal(t.Payload(), &p); err != nil {
			return fmt.Errorf("payment:expire invalid payload: %w", err)
		}
		return container.UseCaseManager().PaymentUseCase().HandlePaymentExpiry(ctx, p.PaymentId)
	})
	if err := asynqSrv.Start(mux); err != nil {
		log.Fatalf("asynq worker start error: %v", err)
	}

	// Re-enqueue tasks for any SCHEDULED/ON_GOING auctions that lost their
	// Redis tasks during a restart. Overdue ON_GOING auctions are closed
	// directly; returns IDs that need payment initialisation.
	startupClosedIds := container.UseCaseManager().AuctionUseCase().EnqueueScheduledTasks(context.Background())
	for _, auctionId := range startupClosedIds {
		if err := container.UseCaseManager().PaymentUseCase().CreateInitialPaymentForWinner(context.Background(), auctionId); err != nil {
			log.Printf("[startup] CreateInitialPaymentForWinner for %s failed: %v", auctionId, err)
		}
	}

	// Process any payments that expired while the server was down.
	if err := container.UseCaseManager().PaymentUseCase().RecoverExpiredPayments(context.Background()); err != nil {
		log.Printf("[startup] RecoverExpiredPayments error: %v", err)
	}

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", global.GetConfig().Port),
		Handler: router.Handler(),
	}

	go func() {
		log.Printf("server starting on port %d", global.GetConfig().Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("http listen error: %s\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	<-quit

	log.Println("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := srv.Shutdown(ctx); err != nil {
			log.Printf("server shutdown error: %s", err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		asynqSrv.Shutdown()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := container.Close(); err != nil {
			log.Printf("container close error: %s", err)
		}
	}()

	wg.Wait()
	log.Println("server exited")
}
