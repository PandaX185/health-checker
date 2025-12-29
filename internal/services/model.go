package services

import "time"

type Service struct {
	ID            int       `json:"id" db:"id"`
	Name          string    `json:"name" db:"name"`
	URL           string    `json:"url" db:"url"`
	CheckInterval int       `json:"check_interval" db:"check_interval"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
}
