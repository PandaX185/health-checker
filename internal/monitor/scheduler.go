package monitor

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

const HealthCheckStream = "health_checks"

type Scheduler struct {
	rdb          *redis.Client
	repo         Repository
	log          *zap.Logger
	ticker       *time.Ticker
	tickInterval int32
}

func NewScheduler(rdb *redis.Client, repo Repository, tickInterval int32, logger *zap.Logger) *Scheduler {
	return &Scheduler{
		rdb:          rdb,
		repo:         repo,
		log:          logger,
		tickInterval: tickInterval,
		ticker:       time.NewTicker(time.Duration(tickInterval) * time.Second),
	}
}

func (s *Scheduler) Start(ctx context.Context) {
	s.log.Info("Scheduler started", zap.Int32("tick_interval_seconds", s.tickInterval))
	for {
		select {
		case <-s.ticker.C:
			dueServices, err := s.repo.ClaimDueServices(ctx)
			if err != nil {
				s.log.Error("failed to list due services", zap.Error(err))
				continue
			}

			for _, service := range dueServices {
				err = s.Enqueue(ctx, service)
				if err != nil {
					s.log.Error("failed to enqueue service",
						zap.Int("service_id", service.ID),
						zap.Error(err),
					)
					continue
				}
			}

		case <-ctx.Done():
			s.ticker.Stop()
			return
		}

	}
}

func (s *Scheduler) Enqueue(ctx context.Context, service Service) error {
	if err := s.rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: HealthCheckStream,
		Values: map[string]interface{}{
			"service_id": service.ID,
			"url":        service.URL,
		},
	}).Err(); err != nil {
		return err
	}

	return nil
}
