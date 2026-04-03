package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/user/app/internal/service"
	"github.com/user/app/pkg/response"
)

// RegisterRequest is the payload for user registration.
type RegisterRequest struct {
	Name     string `json:"name" binding:"required,min=2,max=100"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

// LoginRequest is the payload for user login.
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// RefreshRequest is the payload for token refresh.
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// AuthHandler handles authentication HTTP requests.
type AuthHandler struct {
	authService *service.AuthService
}

// NewAuthHandler creates a new AuthHandler instance.
func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// Register handles POST /api/auth/register
func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	user, accessToken, refreshToken, err := h.authService.Register(req.Name, req.Email, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrUserExists) {
			response.Error(c, http.StatusConflict, "User with this email already exists")
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to register user")
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"user":          user,
	})
}

// Login handles POST /api/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	user, accessToken, refreshToken, err := h.authService.Login(req.Email, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			response.Error(c, http.StatusUnauthorized, "Invalid email or password")
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to login")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"user":          user,
	})
}

// RefreshToken handles POST /api/auth/refresh
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	accessToken, refreshToken, err := h.authService.RefreshToken(req.RefreshToken)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "Invalid or expired refresh token")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
}

// Me handles GET /api/auth/me -- returns the authenticated user.
func (h *AuthHandler) Me(c *gin.Context) {
	val, exists := c.Get("userID")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "User not found in context")
		return
	}

	userID, ok := val.(uint)
	if !ok {
		response.Error(c, http.StatusInternalServerError, "Invalid user ID in context")
		return
	}

	user, err := h.authService.GetUser(userID)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			response.Error(c, http.StatusNotFound, "User not found")
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to get user")
		return
	}

	response.Success(c, http.StatusOK, user)
}
