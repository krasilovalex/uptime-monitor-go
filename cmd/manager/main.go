package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"uptime-monitor/internal/manager"
	"uptime-monitor/internal/postgres"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/segmentio/kafka-go"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	slog.SetDefault(logger)

	kafkaBrokers := []string{getEnv("KAFKA_BROKERS", "localhost:9092")}
	serverAddr := getEnv("SERVER_ADDR", ":8080")

	dbURL := getEnv("DATABASE_URL", "postgres://uptime_user:supersecretpassword@localhost:5434/uptime_monitor?sslmode=disable")

	ctx := context.Background()
	dbPool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		logger.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer dbPool.Close()

	if err := dbPool.Ping(ctx); err != nil {
		logger.Error("Database is not responding", "error", err)
		os.Exit(1)
	}

	logger.Info("Connected to PostgreSQL")

	kafkaWriter := &kafka.Writer{
		Addr:                   kafka.TCP(kafkaBrokers...),
		Topic:                  "ping_tasks",
		Balancer:               &kafka.LeastBytes{},
		AllowAutoTopicCreation: true,
	}
	defer kafkaWriter.Close()

	repo := postgres.NewTargetRepository(dbPool)
	scheduler := manager.NewScheduler(repo, kafkaWriter, logger, 10*time.Second) // tick 10 secconds
	httpHandler := manager.NewHandler(repo)

	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	router.Route("/api/v1/targets", func(r chi.Router) {
		r.Post("/", httpHandler.CreateTarget)
		r.Delete("/{id}", httpHandler.DeleteTarget)
	})

	srv := &http.Server{
		Addr:    serverAddr,
		Handler: router,
	}

	gracefulCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go scheduler.Run(gracefulCtx)

	go func() {
		logger.Info("Starting Manager API Server", "addr", serverAddr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("HTTP Server failed", "error", err)
			os.Exit(1)
		}
	}()

	<-gracefulCtx.Done()
	logger.Info("Shutting down gracefully, press Ctrl+C again to force")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("HTTP Server forced to shutdown", "error", err)
	}

	logger.Info("Manager Service successfully stopped")

}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}

	return fallback
}
