package monitor

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

const HealthCheckGroup = "health_checkers"

type Worker struct {
	rdb      *redis.Client
	repo     Repository
	log      *zap.Logger
	stream   string
	group    string
	consumer string

	httpClient *http.Client
}

func NewWorker(rdb *redis.Client, repo Repository, logger *zap.Logger) *Worker {
	return &Worker{
		rdb:        rdb,
		repo:       repo,
		log:        logger,
		stream:     HealthCheckStream,
		group:      HealthCheckGroup,
		consumer:   "worker_1",
		httpClient: &http.Client{},
	}
}

func (w *Worker) Run(ctx context.Context) {
	w.ensureConsumerGroup(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		msgs, err := w.rdb.XReadGroup(ctx, &redis.XReadGroupArgs{
			Group:    w.group,
			Streams:  []string{w.stream, ">"},
			Consumer: w.consumer,
			Count:    1,
			Block:    5 * time.Second,
		}).Result()
		if err != nil && err != redis.Nil {
			w.log.Error("failed to read from stream", zap.Error(err))
			continue
		}

		for _, msg := range msgs {
			for _, v := range msg.Messages {
				if err := w.processJob(ctx, v.Values); err != nil {
					w.log.Error("failed to process job", zap.Error(err))
					continue
				}

				_, err := w.rdb.XAck(ctx, w.stream, w.group, v.ID).Result()
				if err != nil {
					w.log.Error("failed to acknowledge message", zap.Error(err))
				}
			}
		}
	}
}

func (w *Worker) ensureConsumerGroup(ctx context.Context) {
	err := w.rdb.XGroupCreateMkStream(ctx, w.stream, w.group, "0").Err()
	if err != nil && !strings.Contains(err.Error(), "BUSYGROUP") {
		w.log.Fatal("failed to create consumer group", zap.Error(err))
	}
}

func (w *Worker) processJob(parentCtx context.Context, service map[string]interface{}) error {
	ctx, cancel := context.WithTimeout(parentCtx, 5*time.Second)
	defer cancel()

	start := time.Now().Local()

	serviceID, err := toInt(service["service_id"])
	if err != nil {
		return err
	}
	url, ok := service["url"].(string)
	if !ok {
		return errors.New("failed to parse url")
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}

	status := "DOWN"
	resp, err := w.httpClient.Do(req)
	lat := time.Since(start).Milliseconds()
	if err == nil && resp.StatusCode >= 200 && resp.StatusCode < 300 {
		status = "UP"
	}

	if resp != nil {
		defer resp.Body.Close()
	}

	check := HealthCheck{
		ServiceID: serviceID,
		Status:    status,
		CreatedAt: time.Now().Local(),
		Latency:   int(lat),
	}

	return w.repo.CreateHealthCheck(ctx, check)
}

func toInt(v interface{}) (int, error) {
	switch t := v.(type) {
	case string:
		return strconv.Atoi(t)
	case []byte:
		return strconv.Atoi(string(t))
	case int64:
		return int(t), nil
	case int32:
		return int(t), nil
	case int:
		return t, nil
	default:
		return 0, errors.New("failed to parse service id")
	}
}
