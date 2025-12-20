package service

import (
	"auth/internal/domain"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type AuthService struct {
	userRepo    domain.UserRepository
	jwtSecret   []byte
	tokenExpiry time.Duration
}

type Claims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

func NewAuthService(userRepo domain.UserRepository, jwtSecret string) *AuthService {
	return &AuthService{
		userRepo:    userRepo,
		jwtSecret:   []byte(jwtSecret),
		tokenExpiry: 24 * time.Hour,
	}
}

func (s *AuthService) Login(login, password string) (string, error) {
	user, err := s.userRepo.GetUserByLogin(login)
	if err != nil {
		return "", errors.New("invalid credentials")
	}

	if !user.CheckPassword(password) {
		return "", errors.New("invalid credentials")
	}

	claims := Claims{
		UserID: user.ID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.tokenExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (s *AuthService) ValidateToken(tokenString string) (string, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return s.jwtSecret, nil
	})

	if err != nil {
		return "", err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims.UserID, nil
	}

	return "", errors.New("invalid token")
}

func (s *AuthService) Register(login, password string) error {
	user := &domain.User{
		Login:    login,
		Password: password,
	}

	if err := user.HashPassword(); err != nil {
		return err
	}

	return s.userRepo.CreateUser(user)
}
