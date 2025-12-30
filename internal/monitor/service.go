package monitor

import (
	"context"
	"time"
)

type MonitoringService struct {
	repo Repository
}

func NewService(repo Repository) *MonitoringService {
	return &MonitoringService{repo: repo}
}

func (s *MonitoringService) Register(ctx context.Context, dto RegisterServiceDTO) error {
	service := Service{
		Name:          dto.Name,
		URL:           dto.URL,
		CheckInterval: dto.CheckInterval,
		NextRunAt:     time.Now().Add(time.Second * time.Duration(dto.CheckInterval)),
		CreatedAt:     time.Now(),
	}
	return s.repo.Create(ctx, service)
}

func (s *MonitoringService) ListServices(ctx context.Context) ([]Service, error) {
	return s.repo.ListServices(ctx)
}

func (s *MonitoringService) ListDueServices(ctx context.Context) ([]Service, error) {
	return s.repo.ListDueServices(ctx)
}

func (s *MonitoringService) UpdateNextRunAt(ctx context.Context, serviceID int, nextRunAt time.Time) error {
	return s.repo.UpdateNextRunAt(ctx, serviceID, nextRunAt)
}
