package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) Create(ctx context.Context, user User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockRepository) GetUserByUsername(ctx context.Context, username string) (User, error) {
	args := m.Called(ctx, username)
	return args.Get(0).(User), args.Error(1)
}

func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.Default()
}

func TestRegisterUser(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockRepo := new(MockRepository)
		service := NewService(mockRepo, zap.L())
		handler := NewHandler(service)

		mockRepo.On("Create", mock.Anything, mock.AnythingOfType("auth.User")).Return(nil)

		r := setupRouter()
		r.POST("/auth/register", handler.RegisterUser)

		dto := RegisterUserDTO{
			Username: "testuser",
			Password: "password123",
		}
		jsonValue, _ := json.Marshal(dto)
		req, _ := http.NewRequest("POST", "/auth/register", bytes.NewBuffer(jsonValue))
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		mockRepo.AssertExpectations(t)
	})

	t.Run("BadRequest", func(t *testing.T) {
		mockRepo := new(MockRepository)
		service := NewService(mockRepo, zap.L())
		handler := NewHandler(service)

		r := setupRouter()
		r.POST("/auth/register", handler.RegisterUser)

		req, _ := http.NewRequest("POST", "/auth/register", bytes.NewBuffer([]byte("invalid json")))
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("InternalServerError", func(t *testing.T) {
		mockRepo := new(MockRepository)
		service := NewService(mockRepo, zap.L())
		handler := NewHandler(service)

		mockRepo.On("Create", mock.Anything, mock.AnythingOfType("auth.User")).Return(errors.New("db error"))

		r := setupRouter()
		r.POST("/auth/register", handler.RegisterUser)

		dto := RegisterUserDTO{
			Username: "testuser",
			Password: "password123",
		}
		jsonValue, _ := json.Marshal(dto)
		req, _ := http.NewRequest("POST", "/auth/register", bytes.NewBuffer(jsonValue))
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockRepo.AssertExpectations(t)
	})
}

func TestLoginUser(t *testing.T) {
	os.Setenv("JWT_SECRET", "testsecret")

	t.Run("Success", func(t *testing.T) {
		mockRepo := new(MockRepository)
		service := NewService(mockRepo, zap.L())
		handler := NewHandler(service)

		password := "password123"
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		user := User{
			ID:       1,
			Username: "testuser",
			Password: string(hashedPassword),
		}

		mockRepo.On("GetUserByUsername", mock.Anything, "testuser").Return(user, nil)

		r := setupRouter()
		r.POST("/auth/login", handler.LoginUser)

		dto := LoginUserDTO{
			Username: "testuser",
			Password: password,
		}
		jsonValue, _ := json.Marshal(dto)
		req, _ := http.NewRequest("POST", "/auth/login", bytes.NewBuffer(jsonValue))
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]string
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.NotEmpty(t, response["access_token"])

		mockRepo.AssertExpectations(t)
	})

	t.Run("InvalidCredentials", func(t *testing.T) {
		mockRepo := new(MockRepository)
		service := NewService(mockRepo, zap.L())
		handler := NewHandler(service)

		password := "password123"
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		user := User{
			ID:       1,
			Username: "testuser",
			Password: string(hashedPassword),
		}

		mockRepo.On("GetUserByUsername", mock.Anything, "testuser").Return(user, nil)

		r := setupRouter()
		r.POST("/auth/login", handler.LoginUser)

		dto := LoginUserDTO{
			Username: "testuser",
			Password: "wrongpassword",
		}
		jsonValue, _ := json.Marshal(dto)
		req, _ := http.NewRequest("POST", "/auth/login", bytes.NewBuffer(jsonValue))
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockRepo.AssertExpectations(t)
	})

	t.Run("UserNotFound", func(t *testing.T) {
		mockRepo := new(MockRepository)
		service := NewService(mockRepo, zap.L())
		handler := NewHandler(service)

		mockRepo.On("GetUserByUsername", mock.Anything, "unknown").Return(User{}, errors.New("user not found"))

		r := setupRouter()
		r.POST("/auth/login", handler.LoginUser)

		dto := LoginUserDTO{
			Username: "unknown",
			Password: "password123",
		}
		jsonValue, _ := json.Marshal(dto)
		req, _ := http.NewRequest("POST", "/auth/login", bytes.NewBuffer(jsonValue))
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockRepo.AssertExpectations(t)
	})
}

func TestRegisterRoutes(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewService(mockRepo, zap.L())
	handler := NewHandler(service)

	// Setup mock expectations for register
	mockRepo.On("Create", mock.Anything, mock.AnythingOfType("auth.User")).Return(nil).Maybe()
	// Setup mock expectations for login
	mockRepo.On("GetUserByUsername", mock.Anything, mock.AnythingOfType("string")).Return(User{}, errors.New("user not found")).Maybe()

	r := gin.New()
	group := r.Group("/auth")
	handler.RegisterRoutes(group)

	// Test that routes are registered by making requests
	dto := RegisterUserDTO{
		Username: "testuser",
		Password: "password123",
	}
	jsonValue, _ := json.Marshal(dto)

	// Test register route
	req, _ := http.NewRequest("POST", "/auth/register", bytes.NewBuffer(jsonValue))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusCreated || w.Code == http.StatusBadRequest || w.Code == http.StatusInternalServerError, "Register route should be registered")

	// Test login route
	loginDto := LoginUserDTO{
		Username: "testuser",
		Password: "password123",
	}
	jsonValue, _ = json.Marshal(loginDto)
	req, _ = http.NewRequest("POST", "/auth/login", bytes.NewBuffer(jsonValue))
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusBadRequest || w.Code == http.StatusInternalServerError || w.Code == http.StatusUnauthorized, "Login route should be registered")
}
