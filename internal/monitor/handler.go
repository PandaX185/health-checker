package monitor

import (
	"health-checker/internal/middleware"
	"net/http"
	"strconv"

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
	rg.GET("", h.ListServices)
	rg.GET("/:serviceId/health-checks", h.GetHealthChecks)
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

// ListServices godoc
//
//	 @Security BearerAuth
//		@Summary		List all registered services
//		@Description	Retrieve a list of all services registered for health monitoring
//		@Tags			services
//		@Produce		json
//		@Success		200		{array}		Service	"List of registered services"
//		@Failure		500		{object}	map[string]string	"Internal server error"
//		@Router			/services [get]
func (h *Handler) ListServices(ctx *gin.Context) {
	services, err := h.service.ListServices(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, services)
}

// GetHealthChecks godoc
//
//	 @Security BearerAuth
//		@Summary		Get health checks for a service
//		@Description	Retrieve health check records for a specific service by its ID
//		@Tags			services
//		@Produce		json
//		@Param			serviceId	path		int	true	"Service ID"
//		@Param			page		query		int	false	"Page number"	default(1)
//		@Param			limit		query		int	false	"Number of records per page"	default(10)
//		@Success		200			{array}		HealthCheck	"List of health checks"
//		@Failure		400			{object}	map[string]string	"Bad request"
//		@Failure		500			{object}	map[string]string	"Internal server error"
//		@Router			/services/{serviceId}/health-checks [get]
func (h *Handler) GetHealthChecks(ctx *gin.Context) {
	serviceID, ok := ctx.Params.Get("serviceId")
	if !ok {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid serviceId"})
		return
	}

	page, ok := ctx.GetQuery("page")
	if !ok {
		page = "1"
	}

	limit, ok := ctx.GetQuery("limit")
	if !ok {
		limit = "10"
	}

	serviceIDInt, err := strconv.Atoi(serviceID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "serviceId must be an integer"})
		return
	}
	pageInt, err := strconv.Atoi(page)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "page must be an integer"})
		return
	}
	limitInt, err := strconv.Atoi(limit)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "limit must be an integer"})
		return
	}

	checks, err := h.service.GetHealthChecksByServiceID(ctx.Request.Context(), serviceIDInt, pageInt, limitInt)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, checks)
}
