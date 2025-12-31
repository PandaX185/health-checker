package monitor

import (
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestHandler_RegisterRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := &Handler{}
	router := gin.New()
	rg := router.Group("/api/v1/services")

	// Should not panic
	assert.NotPanics(t, func() {
		handler.RegisterRoutes(rg)
	})

	// Verify routes were registered
	routes := router.Routes()
	assert.NotEmpty(t, routes)
}
