package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"uptime-monitor/internal/worker"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	slog.SetDefault(logger)

	kafkaBrokers := []string{getEnv("KAFKA_BROKERS", "localhost:9092")}

	nunWokers := 50

	pool := worker.NewPool(kafkaBrokers, nunWokers, logger)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool.Run(ctx)
}

// unique function env
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
