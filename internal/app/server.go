package app

import (
	"health-checker/internal/middleware"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func NewServer(logger *zap.Logger) *gin.Engine {
	r := gin.New(func(e *gin.Engine) {
		e.Use(middleware.LoggingMiddleware(logger))
	})
	return r
}
