package auth

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type UserHandler struct {
	service *UserService
	logger  *zap.Logger
}

func NewHandler(service *UserService, logger *zap.Logger) *UserHandler {
	return &UserHandler{
		service: service,
		logger:  logger,
	}
}

func (h *UserHandler) RegisterRoutes(r *gin.RouterGroup) {
	r.POST("/register", h.RegisterUser)
	r.POST("/login", h.LoginUser)
}

// RegisterUser godoc
//
//	@Summary		Register a new user
//	@Description	Register a new user with username and password
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			user	body		RegisterUserDTO		true	"User registration data"
//	@Success		201		{object}	map[string]string	"User registered successfully"
//	@Failure		400		{object}	map[string]string	"Bad request"
//	@Failure		500		{object}	map[string]string	"Internal server error"
//	@Router			/auth/register [post]
func (h *UserHandler) RegisterUser(c *gin.Context) {
	var body RegisterUserDTO
	if err := c.ShouldBindJSON(&body); err != nil {
		h.logger.Error("failed to bind json", zap.Error(err))
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.RegisterUser(c.Request.Context(), body); err != nil {
		h.logger.Error("failed to register user", zap.Error(err))
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(201, gin.H{"message": "User registered successfully"})
}

// LoginUser godoc
//
//	@Summary		Login a user
//	@Description	Authenticate a user and return a JWT token
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			user	body		LoginUserDTO		true	"User login data"
//	@Success		200		{object}	map[string]string	"Access token"
//	@Failure		400		{object}	map[string]string	"Bad request"
//	@Failure		500		{object}	map[string]string	"Internal server error"
//	@Router			/auth/login [post]
func (h *UserHandler) LoginUser(c *gin.Context) {
	var body LoginUserDTO
	if err := c.ShouldBindJSON(&body); err != nil {
		h.logger.Error("failed to bind json", zap.Error(err))
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	token, err := h.service.Login(c.Request.Context(), body)
	if err != nil {
		h.logger.Error("failed to login", zap.Error(err))
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"access_token": token})
}
