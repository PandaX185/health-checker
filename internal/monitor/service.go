package monitor

import (
	"context"
	"time"
)

type ServiceUseCase struct {
	repo Repository
}

func NewService(repo Repository) *ServiceUseCase {
	return &ServiceUseCase{repo: repo}
}

func (s *ServiceUseCase) Register(ctx context.Context, dto RegisterServiceDTO) error {
	service := Service{
		Name:          dto.Name,
		URL:           dto.URL,
		CheckInterval: dto.CheckInterval,
		CreatedAt:     time.Now(),
	}
	return s.repo.Create(ctx, service)
}
