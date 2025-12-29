package auth

import (
	"context"
	"os"
	"time"

	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	repo Repository
}

func NewService(repo Repository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) RegisterUser(ctx context.Context, user RegisterUserDTO) error {
	hashedPassword, err := hashPassword(user.Password)
	if err != nil {
		return err
	}
	u := User{
		Username: user.Username,
		Password: hashedPassword,
	}

	return s.repo.Create(ctx, u)
}

func (s *UserService) Login(ctx context.Context, body LoginUserDTO) (string, error) {
	user, err := s.repo.GetUserByUsername(ctx, body.Username)
	if err != nil {
		return "", err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(body.Password))
	if err != nil {
		return "", err
	}

	return generateJwtToken(user)
}

func generateJwtToken(user User) (string, error) {
	key := os.Getenv("JWT_SECRET")
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"exp":      jwt.TimeFunc().Add(time.Hour * 24).Unix(),
	})

	return token.SignedString([]byte(key))
}

func hashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(hashedPassword), nil
}
