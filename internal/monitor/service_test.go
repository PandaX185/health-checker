package monitor

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestService_Register(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockRepo := new(MockRepository)
		service := NewService(mockRepo)

		dto := RegisterServiceDTO{
			Name:          "Test Service",
			URL:           "http://example.com",
			CheckInterval: 60,
		}

		// We use MatchedBy to verify the Service object passed to Create
		// We ignore CreatedAt exact time match
		mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(s Service) bool {
			return s.Name == dto.Name && s.URL == dto.URL && s.CheckInterval == dto.CheckInterval
		})).Return(nil)

		err := service.Register(context.Background(), dto)
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("RepoError", func(t *testing.T) {
		mockRepo := new(MockRepository)
		service := NewService(mockRepo)

		dto := RegisterServiceDTO{
			Name:          "Test Service",
			URL:           "http://example.com",
			CheckInterval: 60,
		}

		mockRepo.On("Create", mock.Anything, mock.Anything).Return(errors.New("db error"))

		err := service.Register(context.Background(), dto)
		assert.Error(t, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestService_ListServices(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockRepo := new(MockRepository)
		service := NewService(mockRepo)

		expectedServices := []Service{
			{ID: 1, Name: "S1", URL: "u1", CheckInterval: 10},
			{ID: 2, Name: "S2", URL: "u2", CheckInterval: 20},
		}

		mockRepo.On("ListServices", mock.Anything).Return(expectedServices, nil)

		services, err := service.ListServices(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, expectedServices, services)
		mockRepo.AssertExpectations(t)
	})

	t.Run("RepoError", func(t *testing.T) {
		mockRepo := new(MockRepository)
		service := NewService(mockRepo)

		mockRepo.On("ListServices", mock.Anything).Return([]Service{}, errors.New("db error"))

		_, err := service.ListServices(context.Background())
		assert.Error(t, err)
		mockRepo.AssertExpectations(t)
	})
}
