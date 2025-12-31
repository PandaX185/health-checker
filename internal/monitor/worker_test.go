package monitor

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

func TestNewWorker(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{})
	mockRepo := new(MockRepository)
	mockEventBus := new(MockEventBus)
	logger := zap.NewNop()

	worker := NewWorker(rdb, mockRepo, logger, mockEventBus)

	assert.NotNil(t, worker)
	assert.Equal(t, rdb, worker.rdb)
	assert.Equal(t, mockRepo, worker.repo)
	assert.Equal(t, logger, worker.log)
	assert.Equal(t, mockEventBus, worker.eventBus)
	assert.Equal(t, HealthCheckStream, worker.stream)
	assert.Equal(t, HealthCheckGroup, worker.group)
	assert.NotNil(t, worker.httpClient)
}

func TestProcessJob_Success_UP(t *testing.T) {
	// Create test HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	mockRepo := new(MockRepository)
	mockEventBus := new(MockEventBus)

	worker := &Worker{
		repo:       mockRepo,
		eventBus:   mockEventBus,
		log:        zap.NewNop(),
		httpClient: &http.Client{Timeout: 5 * time.Second},
	}

	ctx := context.Background()
	service := map[string]interface{}{
		"service_id": "1",
		"url":        server.URL,
	}

	mockRepo.On("GetLatestHealthCheck", mock.Anything, 1).Return(nil, nil)
	mockRepo.On("CreateHealthCheck", mock.Anything, mock.MatchedBy(func(check HealthCheck) bool {
		return check.ServiceID == 1 && check.Status == "UP" && check.Latency >= 0
	})).Return(nil)

	err := worker.processJob(ctx, service)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockEventBus.AssertNotCalled(t, "Publish")
}

func TestProcessJob_Success_DOWN(t *testing.T) {
	// Create test HTTP server that returns 500
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	mockRepo := new(MockRepository)
	mockEventBus := new(MockEventBus)

	worker := &Worker{
		repo:       mockRepo,
		eventBus:   mockEventBus,
		log:        zap.NewNop(),
		httpClient: &http.Client{Timeout: 5 * time.Second},
	}

	ctx := context.Background()
	service := map[string]interface{}{
		"service_id": "1",
		"url":        server.URL,
	}

	mockRepo.On("GetLatestHealthCheck", mock.Anything, 1).Return(nil, nil)
	mockRepo.On("CreateHealthCheck", mock.Anything, mock.MatchedBy(func(check HealthCheck) bool {
		return check.ServiceID == 1 && check.Status == "DOWN"
	})).Return(nil)

	err := worker.processJob(ctx, service)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestProcessJob_HTTPTimeout(t *testing.T) {
	// Create test HTTP server that delays response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(10 * time.Second) // Longer than worker timeout
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	mockRepo := new(MockRepository)
	mockEventBus := new(MockEventBus)

	worker := &Worker{
		repo:       mockRepo,
		eventBus:   mockEventBus,
		log:        zap.NewNop(),
		httpClient: &http.Client{Timeout: 100 * time.Millisecond},
	}

	ctx := context.Background()
	service := map[string]interface{}{
		"service_id": "1",
		"url":        server.URL,
	}

	mockRepo.On("GetLatestHealthCheck", mock.Anything, 1).Return(nil, nil)
	mockRepo.On("CreateHealthCheck", mock.Anything, mock.MatchedBy(func(check HealthCheck) bool {
		return check.ServiceID == 1 && check.Status == "DOWN" // Timeout should mark as DOWN
	})).Return(nil)

	err := worker.processJob(ctx, service)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestProcessJob_StatusChange_UPtoDOWN(t *testing.T) {
	// Create test HTTP server that returns error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	mockRepo := new(MockRepository)
	mockEventBus := new(MockEventBus)

	worker := &Worker{
		repo:       mockRepo,
		eventBus:   mockEventBus,
		log:        zap.NewNop(),
		httpClient: &http.Client{Timeout: 5 * time.Second},
	}

	ctx := context.Background()
	service := map[string]interface{}{
		"service_id": "1",
		"url":        server.URL,
	}

	previousCheck := &HealthCheck{
		ServiceID: 1,
		Status:    "UP",
		CreatedAt: time.Now().Add(-1 * time.Minute),
	}

	mockRepo.On("GetLatestHealthCheck", mock.Anything, 1).Return(previousCheck, nil)
	mockRepo.On("CreateHealthCheck", mock.Anything, mock.MatchedBy(func(check HealthCheck) bool {
		return check.ServiceID == 1 && check.Status == "DOWN"
	})).Return(nil)
	mockEventBus.On("Publish", mock.Anything, mock.MatchedBy(func(event StatusChangeEvent) bool {
		return event.ServiceID == 1 && event.OldStatus == "UP" && event.NewStatus == "DOWN"
	})).Return(nil)

	err := worker.processJob(ctx, service)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockEventBus.AssertExpectations(t)
}

func TestProcessJob_StatusChange_DOWNtoUP(t *testing.T) {
	// Create test HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	mockRepo := new(MockRepository)
	mockEventBus := new(MockEventBus)

	worker := &Worker{
		repo:       mockRepo,
		eventBus:   mockEventBus,
		log:        zap.NewNop(),
		httpClient: &http.Client{Timeout: 5 * time.Second},
	}

	ctx := context.Background()
	service := map[string]interface{}{
		"service_id": "1",
		"url":        server.URL,
	}

	previousCheck := &HealthCheck{
		ServiceID: 1,
		Status:    "DOWN",
		CreatedAt: time.Now().Add(-1 * time.Minute),
	}

	mockRepo.On("GetLatestHealthCheck", mock.Anything, 1).Return(previousCheck, nil)
	mockRepo.On("CreateHealthCheck", mock.Anything, mock.MatchedBy(func(check HealthCheck) bool {
		return check.ServiceID == 1 && check.Status == "UP"
	})).Return(nil)
	mockEventBus.On("Publish", mock.Anything, mock.MatchedBy(func(event StatusChangeEvent) bool {
		return event.ServiceID == 1 && event.OldStatus == "DOWN" && event.NewStatus == "UP"
	})).Return(nil)

	err := worker.processJob(ctx, service)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockEventBus.AssertExpectations(t)
}

func TestProcessJob_NoStatusChange(t *testing.T) {
	// Create test HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	mockRepo := new(MockRepository)
	mockEventBus := new(MockEventBus)

	worker := &Worker{
		repo:       mockRepo,
		eventBus:   mockEventBus,
		log:        zap.NewNop(),
		httpClient: &http.Client{Timeout: 5 * time.Second},
	}

	ctx := context.Background()
	service := map[string]interface{}{
		"service_id": "1",
		"url":        server.URL,
	}

	previousCheck := &HealthCheck{
		ServiceID: 1,
		Status:    "UP", // Same status
		CreatedAt: time.Now().Add(-1 * time.Minute),
	}

	mockRepo.On("GetLatestHealthCheck", mock.Anything, 1).Return(previousCheck, nil)
	mockRepo.On("CreateHealthCheck", mock.Anything, mock.MatchedBy(func(check HealthCheck) bool {
		return check.ServiceID == 1 && check.Status == "UP"
	})).Return(nil)

	err := worker.processJob(ctx, service)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockEventBus.AssertNotCalled(t, "Publish") // No status change
}

func TestProcessJob_InvalidServiceID(t *testing.T) {
	mockRepo := new(MockRepository)
	mockEventBus := new(MockEventBus)

	worker := &Worker{
		repo:       mockRepo,
		eventBus:   mockEventBus,
		log:        zap.NewNop(),
		httpClient: &http.Client{},
	}

	ctx := context.Background()
	service := map[string]interface{}{
		"service_id": "invalid",
		"url":        "http://example.com",
	}

	err := worker.processJob(ctx, service)

	assert.Error(t, err)
	mockRepo.AssertNotCalled(t, "GetLatestHealthCheck")
	mockRepo.AssertNotCalled(t, "CreateHealthCheck")
}

func TestProcessJob_InvalidURL(t *testing.T) {
	mockRepo := new(MockRepository)
	mockEventBus := new(MockEventBus)

	worker := &Worker{
		repo:       mockRepo,
		eventBus:   mockEventBus,
		log:        zap.NewNop(),
		httpClient: &http.Client{},
	}

	ctx := context.Background()
	service := map[string]interface{}{
		"service_id": "1",
		"url":        123, // Not a string
	}

	err := worker.processJob(ctx, service)

	assert.Error(t, err)
	mockRepo.AssertNotCalled(t, "GetLatestHealthCheck")
	mockRepo.AssertNotCalled(t, "CreateHealthCheck")
}

func TestProcessJob_CreateHealthCheckError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	mockRepo := new(MockRepository)
	mockEventBus := new(MockEventBus)

	worker := &Worker{
		repo:       mockRepo,
		eventBus:   mockEventBus,
		log:        zap.NewNop(),
		httpClient: &http.Client{Timeout: 5 * time.Second},
	}

	ctx := context.Background()
	service := map[string]interface{}{
		"service_id": "1",
		"url":        server.URL,
	}

	mockRepo.On("GetLatestHealthCheck", mock.Anything, 1).Return(nil, nil)
	mockRepo.On("CreateHealthCheck", mock.Anything, mock.Anything).
		Return(errors.New("database error"))

	err := worker.processJob(ctx, service)

	assert.Error(t, err)
	assert.Equal(t, "database error", err.Error())
	mockRepo.AssertExpectations(t)
}

func TestToInt(t *testing.T) {
	tests := []struct {
		name    string
		input   interface{}
		want    int
		wantErr bool
	}{
		{
			name:    "string number",
			input:   "42",
			want:    42,
			wantErr: false,
		},
		{
			name:    "byte slice",
			input:   []byte("123"),
			want:    123,
			wantErr: false,
		},
		{
			name:    "int",
			input:   99,
			want:    99,
			wantErr: false,
		},
		{
			name:    "int32",
			input:   int32(77),
			want:    77,
			wantErr: false,
		},
		{
			name:    "int64",
			input:   int64(88),
			want:    88,
			wantErr: false,
		},
		{
			name:    "invalid string",
			input:   "not_a_number",
			want:    0,
			wantErr: true,
		},
		{
			name:    "invalid type",
			input:   3.14,
			want:    0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := toInt(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestWorker_EnsureConsumerGroup(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx := context.Background()
	rdb := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	defer rdb.Close()

	mockRepo := new(MockRepository)
	mockEventBus := new(MockEventBus)
	logger := zap.NewNop()

	worker := NewWorker(rdb, mockRepo, logger, mockEventBus)

	// Should not panic or error
	worker.ensureConsumerGroup(ctx)
}

// MockEventBus is a mock implementation of EventBus
type MockEventBus struct {
	mock.Mock
}

func (m *MockEventBus) Publish(ctx context.Context, event Event) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockEventBus) Subscribe(eventType string, handler EventHandler) {
	m.Called(eventType, handler)
}
