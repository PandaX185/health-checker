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
		CreatedAt:     time.Now(),
	}
	return s.repo.Create(ctx, service)
}
