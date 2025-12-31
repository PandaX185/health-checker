package database

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRdbInstance(t *testing.T) {
	// Just verify the singleton instance is initialized
	assert.NotNil(t, RdbInstance)
}
