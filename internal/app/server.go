package app

import (
	"health-checker/internal/middleware"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
)

func NewServer(logger *zap.Logger) *gin.Engine {
	r := gin.New(func(e *gin.Engine) {
		e.Use(middleware.LoggingMiddleware(logger))
	})

	r.GET("/swagger/*any", ginSwagger.WrapHandler(
		swaggerFiles.Handler,
		ginSwagger.URL("/swagger/doc.json"),
	))

	return r
}
