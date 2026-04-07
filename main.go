package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	_ "auction-service/docs"

	"auction-service/delivery/api"
	"auction-service/global"
	"auction-service/manager"
)

// @title			Auction Service API
// @version		1.0.0
// @description	Auction Service REST API
// @host			localhost:8080
// @BasePath		/
// @securityDefinitions.apikey	BearerAuth
// @in							header
// @name						Authorization
// @description				Type "Bearer" followed by a space and JWT token.
func main() {
	container := manager.NewContainer()

	router := api.NewRouter(container)

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
		if err := container.Close(); err != nil {
			log.Printf("container close error: %s", err)
		}
	}()

	wg.Wait()
	log.Println("server exited")
}
