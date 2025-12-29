package services

type RegisterServiceDTO struct {
	Name          string `json:"name" binding:"required" example:"My Service"`
	URL           string `json:"url" binding:"required,url" example:"https://example.com"`
	CheckInterval int    `json:"check_interval" binding:"required,min=1" example:"60"`
}
