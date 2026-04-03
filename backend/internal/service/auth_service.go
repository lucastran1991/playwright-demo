package service

import (
	"errors"

	"github.com/user/app/internal/model"
	"github.com/user/app/internal/repository"
	"github.com/user/app/pkg/token"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	ErrUserExists       = errors.New("user with this email already exists")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrUserNotFound     = errors.New("user not found")
)

// AuthService handles authentication business logic.
type AuthService struct {
	userRepo  *repository.UserRepository
	jwtSecret string
}

// NewAuthService creates a new AuthService instance.
func NewAuthService(repo *repository.UserRepository, secret string) *AuthService {
	return &AuthService{userRepo: repo, jwtSecret: secret}
}

// Register creates a new user with hashed password and returns the user.
func (s *AuthService) Register(name, email, password string) (*model.User, string, string, error) {
	// Check if user already exists
	if _, err := s.userRepo.FindByEmail(email); err == nil {
		return nil, "", "", ErrUserExists
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", "", err
	}

	user := &model.User{
		Name:     name,
		Email:    email,
		Password: string(hashedPassword),
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, "", "", err
	}

	accessToken, err := token.GenerateAccessToken(user.ID, user.Email, s.jwtSecret)
	if err != nil {
		return nil, "", "", err
	}

	refreshToken, err := token.GenerateRefreshToken(user.ID, user.Email, s.jwtSecret)
	if err != nil {
		return nil, "", "", err
	}

	return user, accessToken, refreshToken, nil
}

// Login validates credentials and returns JWT tokens.
func (s *AuthService) Login(email, password string) (*model.User, string, string, error) {
	user, err := s.userRepo.FindByEmail(email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, "", "", ErrInvalidCredentials
		}
		return nil, "", "", err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, "", "", ErrInvalidCredentials
	}

	accessToken, err := token.GenerateAccessToken(user.ID, user.Email, s.jwtSecret)
	if err != nil {
		return nil, "", "", err
	}

	refreshToken, err := token.GenerateRefreshToken(user.ID, user.Email, s.jwtSecret)
	if err != nil {
		return nil, "", "", err
	}

	return user, accessToken, refreshToken, nil
}

// RefreshToken validates a refresh token and issues a new token pair.
func (s *AuthService) RefreshToken(refreshTokenStr string) (string, string, error) {
	claims, err := token.ValidateToken(refreshTokenStr, s.jwtSecret)
	if err != nil {
		return "", "", err
	}

	if claims.TokenType != "refresh" {
		return "", "", errors.New("invalid token type: expected refresh token")
	}

	accessToken, err := token.GenerateAccessToken(claims.UserID, claims.Email, s.jwtSecret)
	if err != nil {
		return "", "", err
	}

	newRefreshToken, err := token.GenerateRefreshToken(claims.UserID, claims.Email, s.jwtSecret)
	if err != nil {
		return "", "", err
	}

	return accessToken, newRefreshToken, nil
}

// GetUser returns a user by ID.
func (s *AuthService) GetUser(id uint) (*model.User, error) {
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}
