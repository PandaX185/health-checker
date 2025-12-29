package monitor

import "time"

type Service struct {
	ID            int       `json:"id" db:"id"`
	Name          string    `json:"name" db:"name"`
	URL           string    `json:"url" db:"url"`
	CheckInterval int       `json:"check_interval" db:"check_interval"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
}

type RegisterServiceDTO struct {
	Name          string `json:"name" binding:"required" example:"My Service"`
	URL           string `json:"url" binding:"required,url" example:"https://example.com"`
	CheckInterval int    `json:"check_interval" binding:"required,min=1" example:"60"`
}
