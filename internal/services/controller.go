package services

import "github.com/gin-gonic/gin"

type ServicesController struct {
	service ServicesService
}

func NewServicesController() *ServicesController {
	return &ServicesController{service: *NewServicesService()}
}

// RegisterService godoc
//
//	@Summary		Register a new service
//	@Description	Register a new service for health monitoring
//	@Tags			services
//	@Accept			json
//	@Produce		json
//	@Param			service	body		RegisterServiceDTO	true	"Service data"
//	@Success		200		{object}	map[string]string	"Service registered successfully"
//	@Failure		400		{object}	map[string]string	"Bad request"
//	@Failure		500		{object}	map[string]string	"Internal server error"
//	@Router			/services [post]
func (c *ServicesController) RegisterService() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		body := RegisterServiceDTO{}
		if err := ctx.ShouldBindJSON(&body); err != nil {
			ctx.JSON(400, gin.H{"error": err.Error()})
			return
		}

		err := c.service.RegisterService(body)
		if err != nil {
			ctx.JSON(500, gin.H{"error": err.Error()})
			return
		}

		ctx.JSON(200, gin.H{"message": "Service registered successfully"})
	}
}
