package manager

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/segmentio/kafka-go"
)

type TargetRepository interface {
	GetActiveTargets(ctx context.Context) ([]Target, error)
}

type Scheduler struct {
	repo        TargetRepository
	kafkaWriter *kafka.Writer
	logger      *slog.Logger
	tickRate    time.Duration
}

func NewScheduler(repo TargetRepository, kw *kafka.Writer, logger *slog.Logger, tickRate time.Duration) *Scheduler {
	return &Scheduler{
		repo:        repo,
		kafkaWriter: kw,
		logger:      logger,
		tickRate:    tickRate,
	}
}

func (s *Scheduler) Run(ctx context.Context) {
	ticker := time.NewTicker(s.tickRate)

	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("Scheduler stopping due to context cancellation")
			return
		case <-ticker.C:
			s.processTick(ctx)
		}
	}
}

func (s *Scheduler) processTick(ctx context.Context) {
	targets, err := s.repo.GetActiveTargets(ctx)
	if err != nil {
		s.logger.Error("Failed to get active targets", "error", err)
		return
	}

	for _, t := range targets {
		task := PingTask{
			TargetID:       t.ID,
			URL:            t.URL,
			TimeoutSeconds: t.TimeoutSeconds,
		}

		payload, _ := json.Marshal(task)

		err := s.kafkaWriter.WriteMessages(ctx, kafka.Message{
			Key:   []byte(t.ID),
			Value: payload,
		})

		if err != nil {
			s.logger.Error("Failed to publish task", "target_id", t.ID, "error", err)
		} else {
			s.logger.Debug("Task published", "target_id", t.ID, "url", t.URL)
		}
	}
}
