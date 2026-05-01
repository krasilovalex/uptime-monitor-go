package worker

import (
	"context"
	"encoding/json"
	"log/slog"
	"sync"

	"github.com/segmentio/kafka-go"
)

type Pool struct {
	reader     *kafka.Reader
	writer     *kafka.Writer
	pinger     *Pinger
	numWorkers int
	logger     *slog.Logger
}

func NewPool(brokers []string, numWorkers int, logger *slog.Logger) *Pool {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: brokers,
		Topic:   "ping_tasks",
		GroupID: "worker_pool_group",
	})

	writer := &kafka.Writer{
		Addr:                   kafka.TCP(brokers...),
		Topic:                  "ping_results",
		Balancer:               &kafka.LeastBytes{},
		AllowAutoTopicCreation: true,
	}

	return &Pool{
		reader:     reader,
		writer:     writer,
		pinger:     NewPinger(),
		numWorkers: numWorkers,
		logger:     logger,
	}
}

func (p *Pool) Run(ctx context.Context) {
	tasks := make(chan PingTask, p.numWorkers*2)
	results := make(chan PingResult, p.numWorkers*2)

	var wg sync.WaitGroup

	wg.Add(1)
	go p.resultPublisher(ctx, &wg, results)

	for i := 0; i < p.numWorkers; i++ {
		wg.Add(1)
		go p.worker(ctx, &wg, tasks, results)
	}

	p.logger.Info("Worker Pool started", "workers", p.numWorkers)
	for {
		msg, err := p.reader.ReadMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				break
			}
			p.logger.Error("Failed to read message", "error", err)
			continue
		}

		var task PingTask
		if err := json.Unmarshal(msg.Value, &task); err != nil {
			p.logger.Error("Failed to parse task", "error", err)
			continue
		}

		select {
		case <-ctx.Done():
			goto SHUTDOWN
		case tasks <- task:
		}
	}

SHUTDOWN:
	p.logger.Info("Shutting down worker pool...")
	close(tasks)
	wg.Wait()
	p.reader.Close()
	p.writer.Close()
	p.logger.Info("Worker pool stopped")
}

func (p *Pool) worker(ctx context.Context, wg *sync.WaitGroup, tasks <-chan PingTask, results chan<- PingResult) {
	defer wg.Done()
	for task := range tasks {
		res := p.pinger.Ping(ctx, task)
		results <- res
	}
}

func (p *Pool) resultPublisher(ctx context.Context, wg *sync.WaitGroup, results <-chan PingResult) {
	defer wg.Done()
	for res := range results {
		payload, _ := json.Marshal(res)
		err := p.writer.WriteMessages(ctx, kafka.Message{
			Key:   []byte(res.TargetID),
			Value: payload,
		})

		if err != nil {
			p.logger.Error("Failed to write result", "error", err)
		} else {
			p.logger.Debug("Result published", "target_id", res.TargetID, "is_up", res.IsUp)
		}
	}
}
