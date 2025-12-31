package monitor

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Unit tests for repository operations (mocked database)

type MockDB struct {
	mock.Mock
}

func TestPostgresRepository_Create(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		if testing.Short() {
			t.Skip("Requires database")
		}
		// Covered in integration_test.go
	})
}

func TestPostgresRepository_ListServices(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		if testing.Short() {
			t.Skip("Requires database")
		}
		// Covered in integration_test.go
	})
}

func TestPostgresRepository_ClaimDueServices(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		if testing.Short() {
			t.Skip("Requires database")
		}
		// Covered in integration_test.go
	})
}

func TestPostgresRepository_CreateHealthCheck(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		if testing.Short() {
			t.Skip("Requires database")
		}
		// Covered in integration_test.go
	})
}

func TestPostgresRepository_GetHealthChecksByServiceID(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		if testing.Short() {
			t.Skip("Requires database")
		}
		// Covered in integration_test.go
	})
}

func TestPostgresRepository_GetLatestHealthCheck(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		if testing.Short() {
			t.Skip("Requires database")
		}
		// Covered in integration_test.go
	})
}

func TestMonitoringService_ClaimDueServices(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewService(mockRepo, nil)

	ctx := context.Background()
	expectedServices := []Service{
		{ID: 1, Name: "Service 1"},
	}

	mockRepo.On("ClaimDueServices", ctx).Return(expectedServices, nil)

	result, err := service.ClaimDueServices(ctx)
	assert.NoError(t, err)
	assert.Equal(t, expectedServices, result)
	mockRepo.AssertExpectations(t)
}

func TestMonitoringService_ClaimDueServices_Error(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewService(mockRepo, nil)

	ctx := context.Background()

	mockRepo.On("ClaimDueServices", ctx).Return([]Service{}, errors.New("db error"))

	result, err := service.ClaimDueServices(ctx)
	assert.Error(t, err)
	assert.Empty(t, result)
	mockRepo.AssertExpectations(t)
}

func TestService_Register_WithScheduler(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewService(mockRepo, nil)

	ctx := context.Background()
	dto := RegisterServiceDTO{
		Name:          "Test Service",
		URL:           "http://test.com",
		CheckInterval: 60,
	}

	mockRepo.On("Create", ctx, mock.MatchedBy(func(s Service) bool {
		return s.Name == dto.Name &&
			s.URL == dto.URL &&
			s.CheckInterval == dto.CheckInterval &&
			!s.NextRunAt.IsZero()
	})).Return(nil)

	err := service.Register(ctx, dto)
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestService_GetHealthChecksByServiceID_WithPagination(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewService(mockRepo, nil)

	ctx := context.Background()
	serviceID := 1
	page := 2
	limit := 20

	expectedChecks := []HealthCheck{
		{ID: 1, ServiceID: serviceID, Status: "UP", Latency: 100, CreatedAt: time.Now()},
		{ID: 2, ServiceID: serviceID, Status: "DOWN", Latency: 200, CreatedAt: time.Now()},
	}

	mockRepo.On("GetHealthChecksByServiceID", ctx, serviceID, page, limit).
		Return(expectedChecks, nil)

	result, err := service.GetHealthChecksByServiceID(ctx, serviceID, page, limit)
	assert.NoError(t, err)
	assert.Equal(t, expectedChecks, result)
	assert.Len(t, result, 2)
	mockRepo.AssertExpectations(t)
}

func TestService_GetHealthChecksByServiceID_Empty(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewService(mockRepo, nil)

	ctx := context.Background()
	serviceID := 999
	page := 1
	limit := 10

	mockRepo.On("GetHealthChecksByServiceID", ctx, serviceID, page, limit).
		Return([]HealthCheck{}, nil)

	result, err := service.GetHealthChecksByServiceID(ctx, serviceID, page, limit)
	assert.NoError(t, err)
	assert.Empty(t, result)
	mockRepo.AssertExpectations(t)
}

func TestService_GetHealthChecksByServiceID_Error(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewService(mockRepo, nil)

	ctx := context.Background()
	serviceID := 1
	page := 1
	limit := 10

	mockRepo.On("GetHealthChecksByServiceID", ctx, serviceID, page, limit).
		Return([]HealthCheck{}, errors.New("database connection error"))

	result, err := service.GetHealthChecksByServiceID(ctx, serviceID, page, limit)
	assert.Error(t, err)
	assert.Empty(t, result)
	assert.EqualError(t, err, "database connection error")
	mockRepo.AssertExpectations(t)
}
