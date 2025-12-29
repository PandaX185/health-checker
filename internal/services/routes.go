package services

import "github.com/gin-gonic/gin"

func RegisterServicesRoutes(rg *gin.RouterGroup, controller *ServicesController) {
	rg.POST("", controller.RegisterService())
}
