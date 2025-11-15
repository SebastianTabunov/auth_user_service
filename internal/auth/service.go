package auth

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type Service interface {
	Register(email, password, firstName, lastName string) (*User, error)
	Login(email, password string) (*User, error)
	GenerateToken(userID int, email string) (string, error)
	ValidateToken(tokenString string) (int, string, error)
	GetUserByID(userID int) (*User, error)
	RefreshToken(refreshToken string) (string, error)
	Logout(refreshToken string) error
}

type service struct {
	repo      Repository
	jwtSecret string
}

func NewService(repo Repository, jwtSecret string) Service {
	if jwtSecret == "" {
		panic("JWT secret is required")
	}
	return &service{
		repo:      repo,
		jwtSecret: jwtSecret,
	}
}

func (s *service) Register(email, password, firstName, lastName string) (*User, error) {
	// Проверяем существует ли пользователь
	exists, err := s.repo.UserExists(email)
	if err != nil {
		return nil, fmt.Errorf("failed to check user existence: %w", err)
	}
	if exists {
		return nil, errors.New("user already exists")
	}

	// Хэшируем пароль
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Создаем пользователя
	userID, err := s.repo.CreateUser(email, string(hashedPassword), firstName, lastName)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Получаем созданного пользователя
	user, err := s.repo.GetUserByID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get created user: %w", err)
	}

	return user, nil
}

func (s *service) Login(email, password string) (*User, error) {
	user, err := s.repo.GetUserByEmail(email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("invalid credentials")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Проверяем пароль
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	return user, nil
}

func (s *service) GenerateToken(userID int, email string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"email":   email,
		"exp":     time.Now().Add(time.Hour * 24 * 7).Unix(), // 7 дней
		"iat":     time.Now().Unix(),
		"type":    "access",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}

func (s *service) ValidateToken(tokenString string) (int, string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		return 0, "", err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userIDFloat, ok := claims["user_id"].(float64)
		if !ok {
			return 0, "", errors.New("invalid token: user_id not found")
		}

		email, ok := claims["email"].(string)
		if !ok {
			return 0, "", errors.New("invalid token: email not found")
		}

		return int(userIDFloat), email, nil
	}

	return 0, "", errors.New("invalid token")
}

func (s *service) GetUserByID(userID int) (*User, error) {
	return s.repo.GetUserByID(userID)
}

func (s *service) RefreshToken(refreshToken string) (string, error) {
	user, err := s.repo.GetUserByRefreshToken(refreshToken)
	if err != nil {
		return "", err
	}

	// Генерируем новый access token
	newToken, err := s.GenerateToken(user.ID, user.Email)
	if err != nil {
		return "", fmt.Errorf("failed to generate new token: %w", err)
	}

	return newToken, nil
}

func (s *service) Logout(refreshToken string) error {
	return s.repo.DeleteRefreshToken(refreshToken)
}
