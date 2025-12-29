package services

import (
	"context"
	"health-checker/internal/config"
	"time"

	"github.com/jackc/pgx/v5"
)

type ServicesService struct {
	db *pgx.Conn
}

func NewServicesService() *ServicesService {
	return &ServicesService{db: config.GetDatabase()}
}

func (s *ServicesService) RegisterService(dto RegisterServiceDTO) error {
	query := `
		INSERT INTO services (name, url, check_interval)
		VALUES ($1, $2, $3);
	`

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	_, err := s.db.Exec(ctx, query, dto.Name, dto.URL, dto.CheckInterval)
	return err
}
