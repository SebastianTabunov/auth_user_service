package user

import (
	"context"
	"fmt"
	"time"
)

type Service interface {
	GetProfile(userID int) (*Profile, error)
	UpdateProfile(userID int, profile *Profile) error
}

type service struct {
	repo  Repository
	redis RedisClient
}

type RedisClient interface {
	Get(ctx context.Context, key string, dest interface{}) error
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Delete(ctx context.Context, key string) error
}

func NewService(repo Repository, redisClient RedisClient) Service {
	return &service{
		repo:  repo,
		redis: redisClient,
	}
}

func (s *service) GetProfile(userID int) (*Profile, error) {
	// Пробуем получить из кэша Redis
	if s.redis != nil {
		cacheKey := fmt.Sprintf("user_profile:%d", userID)
		var cachedProfile Profile

		ctx := context.Background()
		err := s.redis.Get(ctx, cacheKey, &cachedProfile)
		if err == nil {
			return &cachedProfile, nil
		}
	}

	// Не нашли в кэше, получаем из БД
	profile, err := s.repo.GetProfile(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get profile: %w", err)
	}

	// Сохраняем в кэш если профиль найден
	if s.redis != nil && profile != nil {
		cacheKey := fmt.Sprintf("user_profile:%d", userID)
		ctx := context.Background()
		if err := s.redis.Set(ctx, cacheKey, profile, 10*time.Minute); err != nil {
			// Логируем ошибку кэширования, но не возвращаем её
			fmt.Printf("Warning: failed to cache profile: %v\n", err)
		}
	}

	return profile, nil
}

func (s *service) UpdateProfile(userID int, profile *Profile) error {
	// Обновляем в БД
	err := s.repo.UpdateProfile(userID, profile)
	if err != nil {
		return fmt.Errorf("failed to update profile: %w", err)
	}

	// Инвалидируем кэш
	if s.redis != nil {
		cacheKey := fmt.Sprintf("user_profile:%d", userID)
		ctx := context.Background()
		if err := s.redis.Delete(ctx, cacheKey); err != nil {
			// Логируем ошибку, но не возвращаем её
			fmt.Printf("Warning: failed to invalidate cache: %v\n", err)
		}
	}

	return nil
}
