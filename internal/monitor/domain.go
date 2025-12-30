package monitor

import "time"

type Service struct {
	ID            int       `json:"id" db:"id"`
	Name          string    `json:"name" db:"name"`
	URL           string    `json:"url" db:"url"`
	CheckInterval int       `json:"check_interval" db:"check_interval"`
	NextRunAt     time.Time `json:"next_run_at" db:"next_run_at"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
}

type RegisterServiceDTO struct {
	Name          string `json:"name" binding:"required" example:"My Service"`
	URL           string `json:"url" binding:"required,url" example:"https://example.com"`
	CheckInterval int    `json:"check_interval" binding:"required,min=1" example:"60"`
}

type HealthCheck struct {
	ID        int       `json:"id" db:"id"`
	ServiceID int       `json:"service_id" db:"service_id"`
	Status    string    `json:"status" db:"status"`
	Latency   int       `json:"latency" db:"latency"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}
