package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rezect/url-shortener/internal/analytics"
	"github.com/rezect/url-shortener/internal/config"
	"github.com/rezect/url-shortener/internal/handler"
	"github.com/rezect/url-shortener/internal/models"
	"github.com/rezect/url-shortener/internal/repository"
	"github.com/rezect/url-shortener/internal/service"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	dbPool, err := pgxpool.New(context.Background(), cfg.DBString())
	if err != nil {
		log.Fatal(err)
	}
	defer dbPool.Close()

	linkRepo := repository.NewLinkRepository(context.Background(), cfg.DBString(), dbPool)
	clickRepo := repository.NewClickRepository(context.Background(), cfg.DBString(), dbPool)

	ch := make(chan models.Click, 1000)
	queue := analytics.NewQueue(
		ch,
		clickRepo,
		cfg.Analytics.BatchSize,
		time.Duration(cfg.Analytics.FlushInterval)*time.Second,
	)
	queue.StartWorkers(cfg.Analytics.Workers)

	linkService := service.NewService(linkRepo, clickRepo)

	handler := handler.NewHandler(linkService, queue, cfg.Server.BaseURL)

	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      handler.GetMux(),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("Starting server on port: %v\n", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Error while starting server: %v\n", err)
		}
	}()

	<-sigCh

	log.Print("Server shutdown...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err = srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Error while shutdown server: %v\n", err)
	}

	handler.Stop()

	log.Printf("Сервер успешно остановлен\n")
}

