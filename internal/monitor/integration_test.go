package monitor

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// Integration tests require PostgreSQL and Redis to be running
// Run with: docker-compose up -d postgres redis

func TestIntegration_SchedulerToWorkerFlow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Setup test database tables
	SetupTestDatabase(t)

	ctx := context.Background()

	// Setup database
	dbURL := getTestDatabaseURL()
	pool, err := pgxpool.New(ctx, dbURL)
	require.NoError(t, err)
	defer pool.Close()

	// Clean up database
	_, err = pool.Exec(ctx, "TRUNCATE TABLE health_checks, services, users RESTART IDENTITY CASCADE")
	require.NoError(t, err)

	// Setup Redis
	redisURL := getTestRedisURL()
	rdb := redis.NewClient(&redis.Options{Addr: redisURL})
	defer rdb.Close()

	streamName := "health_checks_test_" + t.Name()

	// Clean up Redis: delete stream and consumer group completely
	rdb.Del(ctx, streamName)
	// XGroupDestroy will fail if group doesn't exist, that's ok
	rdb.XGroupDestroy(ctx, streamName, HealthCheckGroup)

	// Setup logger
	logger := zap.NewNop()

	// Setup event bus (using InMemoryEventBus)
	eventBus := NewInMemoryEventBus(logger)

	// Setup repository
	repo := NewRepository(pool)

	// Create a test HTTP server
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer testServer.Close()

	// Create a test service
	service := Service{
		Name:          "test-service-integration",
		URL:           testServer.URL,
		CheckInterval: 10,
		NextRunAt:     time.Now().Local(),
		CreatedAt:     time.Now().Local(),
	}

	err = repo.Create(ctx, service)
	require.NoError(t, err)

	// Get the created service
	services, err := repo.ListServices(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, services)
	createdService := &services[0]
	defer cleanupService(t, ctx, repo, createdService.ID)

	// Setup scheduler
	scheduler := NewScheduler(rdb, repo, 5, logger)
	scheduler.stream = streamName

	// Setup worker and ensure consumer group exists (created at "$" to read new messages)
	worker := NewWorker(rdb, repo, logger, eventBus)
	worker.stream = streamName
	worker.ensureConsumerGroup(ctx)

	// NOW Enqueue the service (after consumer group is created at "$")
	err = scheduler.Enqueue(ctx, *createdService)
	require.NoError(t, err)

	// Verify message was added to stream
	streamLen, err := rdb.XLen(ctx, streamName).Result()
	require.NoError(t, err)
	t.Logf("Stream length after enqueue: %d", streamLen)
	require.Greater(t, streamLen, int64(0), "Stream should have at least one message")

	// Process one message
	// Use ">" to read new undelivered messages
	msgs, err := worker.rdb.XReadGroup(ctx, &redis.XReadGroupArgs{
		Group:    HealthCheckGroup,
		Streams:  []string{streamName, ">"},
		Consumer: "test_consumer",
		Count:    1,
		Block:    5 * time.Second,
	}).Result()
	if err != nil {
		t.Logf("XReadGroup error: %v", err)
	}
	require.NoError(t, err, "XReadGroup should not return an error")
	require.Len(t, msgs, 1, "Should receive 1 stream")
	require.Len(t, msgs[0].Messages, 1, "Should receive 1 message")

	// Process the job
	err = worker.processJob(ctx, msgs[0].Messages[0].Values)
	require.NoError(t, err)

	// Verify health check was created
	checks, err := repo.GetHealthChecksByServiceID(ctx, createdService.ID, 1, 10)
	require.NoError(t, err)
	assert.Len(t, checks, 1)
	assert.Equal(t, "UP", checks[0].Status)
	assert.Equal(t, createdService.ID, checks[0].ServiceID)
}

func TestIntegration_StatusChangeEvent(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Setup test database tables
	SetupTestDatabase(t)

	ctx := context.Background()

	// Setup database
	dbURL := getTestDatabaseURL()
	pool, err := pgxpool.New(ctx, dbURL)
	require.NoError(t, err)
	defer pool.Close()

	// Clean up database
	_, err = pool.Exec(ctx, "TRUNCATE TABLE health_checks, services, users RESTART IDENTITY CASCADE")
	require.NoError(t, err)

	// Setup Redis
	redisURL := getTestRedisURL()
	rdb := redis.NewClient(&redis.Options{Addr: redisURL})
	defer rdb.Close()

	// Clean up Redis stream
	rdb.Del(ctx, HealthCheckStream)

	// Setup logger
	logger := zap.NewNop()

	// Setup event bus (using InMemoryEventBus)
	eventBus := NewInMemoryEventBus(logger)

	// Channel to capture events
	eventChan := make(chan StatusChangeEvent, 1)
	eventBus.Subscribe("StatusChange", func(ctx context.Context, event Event) {
		if sce, ok := event.(StatusChangeEvent); ok {
			eventChan <- sce
		}
	})

	// Setup repository
	repo := NewRepository(pool)

	// Create first test HTTP server (returns 200)
	testServerOK := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer testServerOK.Close()

	// Create a test service
	service := Service{
		Name:          "test-service-status-change",
		URL:           testServerOK.URL,
		CheckInterval: 10,
		NextRunAt:     time.Now().Local(),
		CreatedAt:     time.Now().Local(),
	}

	err = repo.Create(ctx, service)
	require.NoError(t, err)

	// Get the created service
	services, err := repo.ListServices(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, services)
	createdService := &services[0]
	defer cleanupService(t, ctx, repo, createdService.ID)

	// Setup worker
	worker := NewWorker(rdb, repo, logger, eventBus)

	// First check - should be UP
	jobData := map[string]interface{}{
		"service_id": createdService.ID,
		"url":        testServerOK.URL,
	}
	err = worker.processJob(ctx, jobData)
	require.NoError(t, err)

	// Verify first check
	checks, err := repo.GetHealthChecksByServiceID(ctx, createdService.ID, 1, 10)
	require.NoError(t, err)
	assert.Equal(t, "UP", checks[0].Status)

	// Change URL to failing server
	testServerFail := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer testServerFail.Close()

	// Second check - should be DOWN and trigger event
	jobData["url"] = testServerFail.URL
	err = worker.processJob(ctx, jobData)
	require.NoError(t, err)

	// Wait for event
	select {
	case event := <-eventChan:
		assert.Equal(t, createdService.ID, event.ServiceID)
		assert.Equal(t, "UP", event.OldStatus)
		assert.Equal(t, "DOWN", event.NewStatus)
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for status change event")
	}

	// Verify second check
	checks, err = repo.GetHealthChecksByServiceID(ctx, createdService.ID, 1, 10)
	require.NoError(t, err)
	assert.Len(t, checks, 2)
	assert.Equal(t, "DOWN", checks[0].Status)
}

func TestIntegration_RepositoryOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Setup test database tables
	SetupTestDatabase(t)

	ctx := context.Background()

	// Setup database
	dbURL := getTestDatabaseURL()
	pool, err := pgxpool.New(ctx, dbURL)
	require.NoError(t, err)
	defer pool.Close()

	// Clean up database
	_, err = pool.Exec(ctx, "TRUNCATE TABLE health_checks, services, users RESTART IDENTITY CASCADE")
	require.NoError(t, err)

	repo := NewRepository(pool)

	// Test Create
	service := Service{
		Name:          "test-repo-service",
		URL:           "http://example.com",
		CheckInterval: 30,
		NextRunAt:     time.Now().Local(),
		CreatedAt:     time.Now().Local(),
	}

	err = repo.Create(ctx, service)
	require.NoError(t, err)

	// Test ListServices
	services, err := repo.ListServices(ctx)
	require.NoError(t, err)
	assert.NotEmpty(t, services)
	created := services[0]
	defer cleanupService(t, ctx, repo, created.ID)

	// Test CreateHealthCheck
	check := HealthCheck{
		ServiceID: created.ID,
		Status:    "UP",
		Latency:   100,
		CreatedAt: time.Now().Local(),
	}
	err = repo.CreateHealthCheck(ctx, check)
	require.NoError(t, err)

	// Test GetHealthChecksByServiceID
	checks, err := repo.GetHealthChecksByServiceID(ctx, created.ID, 1, 10)
	require.NoError(t, err)
	assert.Len(t, checks, 1)
	assert.Equal(t, "UP", checks[0].Status)

	// Test GetLatestHealthCheck
	latest, err := repo.GetLatestHealthCheck(ctx, created.ID)
	require.NoError(t, err)
	assert.NotNil(t, latest)
	assert.Equal(t, "UP", latest.Status)

	// Create another check
	check2 := HealthCheck{
		ServiceID: created.ID,
		Status:    "DOWN",
		Latency:   200,
		CreatedAt: time.Now().Local(),
	}
	err = repo.CreateHealthCheck(ctx, check2)
	require.NoError(t, err)

	// Verify latest is now DOWN
	latest, err = repo.GetLatestHealthCheck(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, "DOWN", latest.Status)

	// Test ClaimDueServices
	claimed, err := repo.ClaimDueServices(ctx)
	require.NoError(t, err)
	t.Logf("Claimed %d services", len(claimed))
}

func TestIntegration_RedisStreamOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx := context.Background()

	// Setup Redis
	redisURL := getTestRedisURL()
	rdb := redis.NewClient(&redis.Options{Addr: redisURL})
	defer rdb.Close()

	streamName := "test_stream"
	groupName := "test_group"

	// Clean up
	rdb.Del(ctx, streamName)

	// Create consumer group
	err := rdb.XGroupCreateMkStream(ctx, streamName, groupName, "0").Err()
	require.NoError(t, err)

	// Add message to stream
	_, err = rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: streamName,
		Values: map[string]interface{}{
			"service_id": "123",
			"url":        "http://example.com",
		},
	}).Result()
	require.NoError(t, err)

	// Read from consumer group
	msgs, err := rdb.XReadGroup(ctx, &redis.XReadGroupArgs{
		Group:    groupName,
		Streams:  []string{streamName, ">"},
		Consumer: "test_consumer",
		Count:    1,
		Block:    time.Second,
	}).Result()
	require.NoError(t, err)
	assert.Len(t, msgs, 1)
	assert.Len(t, msgs[0].Messages, 1)

	msg := msgs[0].Messages[0]
	assert.Equal(t, "123", msg.Values["service_id"])
	assert.Equal(t, "http://example.com", msg.Values["url"])

	// Acknowledge message
	acked, err := rdb.XAck(ctx, streamName, groupName, msg.ID).Result()
	require.NoError(t, err)
	assert.Equal(t, int64(1), acked)

	// Clean up
	rdb.Del(ctx, streamName)
}

func getTestDatabaseURL() string {
	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:root@localhost:5432/health_checker?sslmode=disable"
	}
	return dbURL
}

func getTestRedisURL() string {
	redisURL := os.Getenv("TEST_REDIS_URL")
	if redisURL == "" {
		redisURL = "localhost:6379"
	}
	return redisURL
}

func cleanupService(t *testing.T, ctx context.Context, repo Repository, serviceID int) {
	// Clean up through direct SQL since we don't have DeleteService in Repository
	t.Logf("Note: Service cleanup would need direct database access (serviceID: %d)", serviceID)
}
