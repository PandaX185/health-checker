package app

import (
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestNewServer(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zap.NewNop()

	engine := NewServer(logger)

	assert.NotNil(t, engine)
}
