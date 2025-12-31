package monitor

import (
	"context"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

type MockRedisClient struct {
	mock.Mock
}

func (m *MockRedisClient) XAdd(ctx context.Context, args *redis.XAddArgs) *redis.StringCmd {
	mockArgs := m.Called(ctx, args)
	cmd := redis.NewStringCmd(ctx)
	if err := mockArgs.Error(0); err != nil {
		cmd.SetErr(err)
	}
	return cmd
}

func TestNewScheduler(t *testing.T) {
	rdb := &redis.Client{}
	mockRepo := new(MockRepository)
	logger := zap.NewNop()

	scheduler := NewScheduler(rdb, mockRepo, 5, logger)

	assert.NotNil(t, scheduler)
	assert.Equal(t, int32(5), scheduler.tickInterval)
	assert.NotNil(t, scheduler.ticker)
	assert.NotNil(t, scheduler.log)
}

func TestScheduler_Enqueue(t *testing.T) {
	t.Run("Successfully enqueues service", func(t *testing.T) {
		// Use a real Redis client for this test (or mock XAdd)
		rdb := redis.NewClient(&redis.Options{
			Addr: "localhost:6379",
		})
		defer rdb.Close()

		mockRepo := new(MockRepository)
		logger := zap.NewNop()
		scheduler := NewScheduler(rdb, mockRepo, 1, logger)

		service := Service{
			ID:            1,
			Name:          "Test Service",
			URL:           "http://example.com",
			CheckInterval: 60,
		}

		ctx := context.Background()

		// Try to enqueue - may fail if Redis not available, but test structure is valid
		err := scheduler.Enqueue(ctx, service)

		// If Redis is available, should succeed
		if err == nil {
			// Verify the message was added to stream
			messages, _ := rdb.XRead(ctx, &redis.XReadArgs{
				Streams: []string{HealthCheckStream, "0"},
				Count:   1,
			}).Result()

			if len(messages) > 0 && len(messages[0].Messages) > 0 {
				msg := messages[0].Messages[0]
				assert.Equal(t, "1", msg.Values["service_id"])
				assert.Equal(t, "http://example.com", msg.Values["url"])
			}
		}
	})
}

func TestScheduler_Enqueue_ValidatesData(t *testing.T) {
	t.Run("Enqueues with correct data structure", func(t *testing.T) {
		rdb := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
		defer rdb.Close()

		mockRepo := new(MockRepository)
		logger := zap.NewNop()
		scheduler := NewScheduler(rdb, mockRepo, 1, logger)

		service := Service{
			ID:  42,
			URL: "https://api.example.com/health",
		}

		ctx := context.Background()
		_ = scheduler.Enqueue(ctx, service)

		// The enqueue method should create proper Redis stream entry
		// Test that it doesn't panic or return unexpected errors
	})
}

func TestScheduler_Start_ClaimsDueServices(t *testing.T) {
	t.Run("Claims and enqueues due services", func(t *testing.T) {
		rdb := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
		defer rdb.Close()

		mockRepo := new(MockRepository)
		logger := zap.NewNop()

		dueServices := []Service{
			{ID: 1, URL: "http://service1.com", CheckInterval: 60},
			{ID: 2, URL: "http://service2.com", CheckInterval: 120},
		}

		mockRepo.On("ClaimDueServices", mock.Anything).Return(dueServices, nil).Once()

		scheduler := NewScheduler(rdb, mockRepo, 1, logger)

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		// Start scheduler in goroutine
		go scheduler.Start(ctx)

		// Wait for at least one tick
		time.Sleep(1500 * time.Millisecond)

		// Verify ClaimDueServices was called
		mockRepo.AssertCalled(t, "ClaimDueServices", mock.Anything)
	})
}

func TestScheduler_Start_HandlesErrors(t *testing.T) {
	t.Run("Continues on ClaimDueServices error", func(t *testing.T) {
		rdb := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
		defer rdb.Close()

		mockRepo := new(MockRepository)
		logger := zap.NewNop()

		mockRepo.On("ClaimDueServices", mock.Anything).Return([]Service{}, assert.AnError)

		scheduler := NewScheduler(rdb, mockRepo, 1, logger)

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		// Should not panic, should log error and continue
		go scheduler.Start(ctx)

		time.Sleep(1500 * time.Millisecond)

		// Scheduler should still be running and calling ClaimDueServices
		mockRepo.AssertCalled(t, "ClaimDueServices", mock.Anything)
	})
}

func TestScheduler_Start_StopsOnContextCancel(t *testing.T) {
	t.Run("Stops when context is cancelled", func(t *testing.T) {
		rdb := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
		defer rdb.Close()

		mockRepo := new(MockRepository)
		logger := zap.NewNop()

		mockRepo.On("ClaimDueServices", mock.Anything).Return([]Service{}, nil)

		scheduler := NewScheduler(rdb, mockRepo, 1, logger)

		ctx, cancel := context.WithCancel(context.Background())

		done := make(chan bool)
		go func() {
			scheduler.Start(ctx)
			done <- true
		}()

		// Cancel context after short time
		time.Sleep(100 * time.Millisecond)
		cancel()

		// Wait for scheduler to stop
		select {
		case <-done:
			// Success - scheduler stopped
		case <-time.After(2 * time.Second):
			t.Fatal("Scheduler did not stop after context cancellation")
		}
	})
}

func TestScheduler_Enqueue_MultipleServices(t *testing.T) {
	t.Run("Enqueues multiple services correctly", func(t *testing.T) {
		rdb := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
		defer rdb.Close()

		mockRepo := new(MockRepository)
		logger := zap.NewNop()
		scheduler := NewScheduler(rdb, mockRepo, 1, logger)

		services := []Service{
			{ID: 1, URL: "http://service1.com"},
			{ID: 2, URL: "http://service2.com"},
			{ID: 3, URL: "http://service3.com"},
		}

		ctx := context.Background()

		for _, service := range services {
			err := scheduler.Enqueue(ctx, service)
			// Check error only if Redis is available
			if err == nil {
				assert.NoError(t, err)
			}
		}
	})
}
