package auth

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

func TestService_RegisterUser(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockRepo := new(MockRepository)
		service := NewService(mockRepo, zap.L())

		dto := RegisterUserDTO{
			Username: "testuser",
			Password: "password123",
		}

		mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(u User) bool {
			// Verify username matches and password is NOT the plain text one
			return u.Username == dto.Username && u.Password != dto.Password
		})).Return(nil)

		err := service.RegisterUser(context.Background(), dto)
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("RepoError", func(t *testing.T) {
		mockRepo := new(MockRepository)
		service := NewService(mockRepo, zap.L())

		dto := RegisterUserDTO{
			Username: "testuser",
			Password: "password123",
		}

		mockRepo.On("Create", mock.Anything, mock.Anything).Return(errors.New("db error"))

		err := service.RegisterUser(context.Background(), dto)
		assert.Error(t, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestService_Login(t *testing.T) {
	os.Setenv("JWT_SECRET", "testsecret")

	t.Run("Success", func(t *testing.T) {
		mockRepo := new(MockRepository)
		service := NewService(mockRepo, zap.L())

		password := "password123"
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

		user := User{
			ID:       1,
			Username: "testuser",
			Password: string(hashedPassword),
		}

		mockRepo.On("GetUserByUsername", mock.Anything, "testuser").Return(user, nil)

		dto := LoginUserDTO{
			Username: "testuser",
			Password: password,
		}

		token, err := service.Login(context.Background(), dto)
		assert.NoError(t, err)
		assert.NotEmpty(t, token)
		mockRepo.AssertExpectations(t)
	})

	t.Run("UserNotFound", func(t *testing.T) {
		mockRepo := new(MockRepository)
		service := NewService(mockRepo, zap.L())

		mockRepo.On("GetUserByUsername", mock.Anything, "unknown").Return(User{}, errors.New("user not found"))

		dto := LoginUserDTO{
			Username: "unknown",
			Password: "password123",
		}

		token, err := service.Login(context.Background(), dto)
		assert.Error(t, err)
		assert.Empty(t, token)
		mockRepo.AssertExpectations(t)
	})

	t.Run("WrongPassword", func(t *testing.T) {
		mockRepo := new(MockRepository)
		service := NewService(mockRepo, zap.L())

		password := "password123"
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

		user := User{
			ID:       1,
			Username: "testuser",
			Password: string(hashedPassword),
		}

		mockRepo.On("GetUserByUsername", mock.Anything, "testuser").Return(user, nil)

		dto := LoginUserDTO{
			Username: "testuser",
			Password: "wrongpassword",
		}

		token, err := service.Login(context.Background(), dto)
		assert.Error(t, err)
		assert.Empty(t, token)
		mockRepo.AssertExpectations(t)
	})
}
