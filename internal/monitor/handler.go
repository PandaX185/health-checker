package monitor

import (
	"health-checker/internal/middleware"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *MonitoringService
}

func NewHandler(service *MonitoringService) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.Use(middleware.AuthMiddleware())
	rg.POST("", h.RegisterService)
}

// RegisterService godoc
//
//	 @Security BearerAuth
//		@Summary		Register a new service
//		@Description	Register a new service for health monitoring
//		@Tags			services
//		@Accept			json
//		@Produce		json
//		@Param			service	body		RegisterServiceDTO	true	"Service data"
//		@Success		200		{object}	map[string]string	"Service registered successfully"
//		@Failure		400		{object}	map[string]string	"Bad request"
//		@Failure		500		{object}	map[string]string	"Internal server error"
//		@Router			/services [post]
func (h *Handler) RegisterService(ctx *gin.Context) {
	var body RegisterServiceDTO
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.Register(ctx.Request.Context(), body); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Service registered successfully"})
}
