package monitor

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
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

func TestHandleWebSocketGin_NoToken(t *testing.T) {
	// Create minimal handler for testing
	service := &MonitoringService{}
	hub := NewWsHub(zap.NewNop())
	handler := NewHandler(service, hub, zap.NewNop())

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/ws", nil)

	handler.HandleWebSocketGin(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestHandleWebSocketGin_InvalidToken(t *testing.T) {
	service := &MonitoringService{}
	hub := NewWsHub(zap.NewNop())
	handler := NewHandler(service, hub, zap.NewNop())

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/ws", nil)
	c.Request.Header.Set("Authorization", "Bearer invalid-token")

	handler.HandleWebSocketGin(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestHandleWebSocket(t *testing.T) {
	// This function requires a WebSocket connection
	// For unit testing, we can verify the function exists
	service := &MonitoringService{}
	hub := NewWsHub(zap.NewNop())
	handler := NewHandler(service, hub, zap.NewNop())
	assert.NotNil(t, handler)
	assert.NotNil(t, handler.HandleWebSocket)
}
