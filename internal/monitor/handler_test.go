package monitor

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) Create(ctx context.Context, service Service) error {
	args := m.Called(ctx, service)
	return args.Error(0)
}

func (m *MockRepository) ListServices(ctx context.Context) ([]Service, error) {
	args := m.Called(ctx)
	return args.Get(0).([]Service), args.Error(1)
}

func (m *MockRepository) CreateHealthCheck(ctx context.Context, check HealthCheck) error {
	args := m.Called(ctx, check)
	return args.Error(0)
}

func (m *MockRepository) ClaimDueServices(ctx context.Context) ([]Service, error) {
	args := m.Called(ctx)
	return args.Get(0).([]Service), args.Error(1)
}

func (m *MockRepository) GetHealthChecksByServiceID(ctx context.Context, serviceID, page, limit int) ([]HealthCheck, error) {
	args := m.Called(ctx, serviceID, page, limit)
	return args.Get(0).([]HealthCheck), args.Error(1)
}

func (m *MockRepository) GetLatestHealthCheck(ctx context.Context, serviceID int) (*HealthCheck, error) {
	args := m.Called(ctx, serviceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*HealthCheck), args.Error(1)
}

func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.Default()
}

func TestRegisterService(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockRepo := new(MockRepository)
		service := NewService(mockRepo, zap.L())
		hub := NewWsHub(zap.L())
		handler := NewHandler(service, hub, zap.NewNop())

		mockRepo.On("Create", mock.Anything, mock.AnythingOfType("monitor.Service")).Return(nil)

		r := setupRouter()
		r.POST("/services", handler.RegisterService)

		dto := RegisterServiceDTO{
			Name:          "Test Service",
			URL:           "http://example.com",
			CheckInterval: 60,
		}
		jsonValue, _ := json.Marshal(dto)
		req, _ := http.NewRequest("POST", "/services", bytes.NewBuffer(jsonValue))
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockRepo.AssertExpectations(t)
	})

	t.Run("BadRequest", func(t *testing.T) {
		mockRepo := new(MockRepository)
		service := NewService(mockRepo, zap.L())
		hub := NewWsHub(zap.L())
		handler := NewHandler(service, hub, zap.NewNop())

		r := setupRouter()
		r.POST("/services", handler.RegisterService)

		req, _ := http.NewRequest("POST", "/services", bytes.NewBuffer([]byte("invalid json")))
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("InternalServerError", func(t *testing.T) {
		mockRepo := new(MockRepository)
		service := NewService(mockRepo, zap.L())
		hub := NewWsHub(zap.L())
		handler := NewHandler(service, hub, zap.NewNop())

		mockRepo.On("Create", mock.Anything, mock.AnythingOfType("monitor.Service")).Return(errors.New("db error"))

		r := setupRouter()
		r.POST("/services", handler.RegisterService)

		dto := RegisterServiceDTO{
			Name:          "Test Service",
			URL:           "http://example.com",
			CheckInterval: 60,
		}
		jsonValue, _ := json.Marshal(dto)
		req, _ := http.NewRequest("POST", "/services", bytes.NewBuffer(jsonValue))
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockRepo.AssertExpectations(t)
	})
}

func TestListServices(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockRepo := new(MockRepository)
		service := NewService(mockRepo, zap.L())
		hub := NewWsHub(zap.L())
		handler := NewHandler(service, hub, zap.NewNop())

		expectedServices := []Service{
			{ID: 1, Name: "Service 1", URL: "http://s1.com", CheckInterval: 60},
			{ID: 2, Name: "Service 2", URL: "http://s2.com", CheckInterval: 120},
		}
		mockRepo.On("ListServices", mock.Anything).Return(expectedServices, nil)

		r := setupRouter()
		r.GET("/services", handler.ListServices)

		req, _ := http.NewRequest("GET", "/services", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response []Service
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Len(t, response, 2)
		assert.Equal(t, expectedServices[0].Name, response[0].Name)

		mockRepo.AssertExpectations(t)
	})

	t.Run("InternalServerError", func(t *testing.T) {
		mockRepo := new(MockRepository)
		service := NewService(mockRepo, zap.L())
		hub := NewWsHub(zap.L())
		handler := NewHandler(service, hub, zap.NewNop())

		mockRepo.On("ListServices", mock.Anything).Return([]Service{}, errors.New("db error"))

		r := setupRouter()
		r.GET("/services", handler.ListServices)

		req, _ := http.NewRequest("GET", "/services", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockRepo.AssertExpectations(t)
	})
}

func TestGetHealthChecks(t *testing.T) {
	t.Run("Success_WithDefaults", func(t *testing.T) {
		mockRepo := new(MockRepository)
		service := NewService(mockRepo, zap.L())
		hub := NewWsHub(zap.L())
		handler := NewHandler(service, hub, zap.NewNop())

		expectedChecks := []HealthCheck{
			{ID: 1, ServiceID: 1, Status: "UP", Latency: 100},
		}
		mockRepo.On("GetHealthChecksByServiceID", mock.Anything, 1, 1, 10).Return(expectedChecks, nil)

		r := setupRouter()
		r.GET("/services/:serviceId/health-checks", handler.GetHealthChecks)

		req, _ := http.NewRequest("GET", "/services/1/health-checks", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Success_WithParams", func(t *testing.T) {
		mockRepo := new(MockRepository)
		service := NewService(mockRepo, zap.L())
		hub := NewWsHub(zap.L())
		handler := NewHandler(service, hub, zap.NewNop())

		expectedChecks := []HealthCheck{
			{ID: 1, ServiceID: 1, Status: "UP", Latency: 100},
		}
		mockRepo.On("GetHealthChecksByServiceID", mock.Anything, 1, 2, 20).Return(expectedChecks, nil)

		r := setupRouter()
		r.GET("/services/:serviceId/health-checks", handler.GetHealthChecks)

		req, _ := http.NewRequest("GET", "/services/1/health-checks?page=2&limit=20", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockRepo.AssertExpectations(t)
	})

	t.Run("InvalidServiceID", func(t *testing.T) {
		mockRepo := new(MockRepository)
		service := NewService(mockRepo, zap.L())
		hub := NewWsHub(zap.L())
		handler := NewHandler(service, hub, zap.NewNop())

		r := setupRouter()
		r.GET("/services/:serviceId/health-checks", handler.GetHealthChecks)

		req, _ := http.NewRequest("GET", "/services/invalid/health-checks", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("InvalidPage", func(t *testing.T) {
		mockRepo := new(MockRepository)
		service := NewService(mockRepo, zap.L())
		hub := NewWsHub(zap.L())
		handler := NewHandler(service, hub, zap.NewNop())

		r := setupRouter()
		r.GET("/services/:serviceId/health-checks", handler.GetHealthChecks)

		req, _ := http.NewRequest("GET", "/services/1/health-checks?page=invalid", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("InvalidLimit", func(t *testing.T) {
		mockRepo := new(MockRepository)
		service := NewService(mockRepo, zap.L())
		hub := NewWsHub(zap.L())
		handler := NewHandler(service, hub, zap.NewNop())

		r := setupRouter()
		r.GET("/services/:serviceId/health-checks", handler.GetHealthChecks)

		req, _ := http.NewRequest("GET", "/services/1/health-checks?limit=invalid", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("InternalServerError", func(t *testing.T) {
		mockRepo := new(MockRepository)
		service := NewService(mockRepo, zap.L())
		hub := NewWsHub(zap.L())
		handler := NewHandler(service, hub, zap.NewNop())

		mockRepo.On("GetHealthChecksByServiceID", mock.Anything, 1, 1, 10).Return([]HealthCheck{}, errors.New("db error"))

		r := setupRouter()
		r.GET("/services/:serviceId/health-checks", handler.GetHealthChecks)

		req, _ := http.NewRequest("GET", "/services/1/health-checks", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockRepo.AssertExpectations(t)
	})
}
