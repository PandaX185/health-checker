package monitor

import (
	"context"
	"time"

	"go.uber.org/zap"
)

type MonitoringService struct {
	repo Repository
	log  *zap.Logger
}

func NewService(repo Repository, log *zap.Logger) *MonitoringService {
	return &MonitoringService{repo: repo, log: log}
}

func (s *MonitoringService) Register(ctx context.Context, dto RegisterServiceDTO) error {
	service := Service{
		Name:          dto.Name,
		URL:           dto.URL,
		CheckInterval: dto.CheckInterval,
		NextRunAt:     time.Now().Local().Add(time.Second * time.Duration(dto.CheckInterval)),
	}
	return s.repo.Create(ctx, service)
}

func (s *MonitoringService) ListServices(ctx context.Context) ([]Service, error) {
	return s.repo.ListServices(ctx)
}

func (s *MonitoringService) ClaimDueServices(ctx context.Context) ([]Service, error) {
	return s.repo.ClaimDueServices(ctx)
}
