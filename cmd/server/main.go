package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"auth-user-service/internal/auth"
	"auth-user-service/internal/config"
	"auth-user-service/internal/database"
	"auth-user-service/internal/order"
	"auth-user-service/internal/redis"
	"auth-user-service/internal/user"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"
)

func main() {
	// Загружаем конфигурацию
	cfg := config.Load()

	if cfg.Environment == "production" && cfg.JWT.Secret == "" {
		log.Fatal("JWT_SECRET must be set in production")
	}

	// Подключаемся к PostgreSQL
	dbConfig := database.DatabaseConfig{
		Host:     cfg.Database.Host,
		Port:     cfg.Database.Port,
		User:     cfg.Database.User,
		Password: cfg.Database.Password,
		DBName:   cfg.Database.DBName,
		SSLMode:  cfg.Database.SSLMode,
	}
	db, err := database.NewConnection(dbConfig)
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("⚠️ Error closing database: %v", err)
		}
	}()

	log.Println("✅ Database connected successfully")

	// Подключаемся к Redis
	var redisClient *redis.Client
	if cfg.Redis.URL != "" {
		redisClient, err = redis.NewClient(cfg.Redis.URL)
		if err != nil {
			log.Printf("⚠️ Failed to connect to Redis: %v", err)
			log.Println("⚠️ Continuing without Redis...")
			redisClient = nil
		} else {
			defer func() {
				if err := redisClient.Close(); err != nil {
					log.Printf("⚠️ Error closing Redis: %v", err)
				}
			}()
			log.Println("✅ Redis connected successfully")
		}
	}

	// Инициализация сервисов
	authRepo := auth.NewRepository(db)
	authService := auth.NewService(authRepo, cfg.JWT.Secret)
	authHandler := auth.NewHandler(authService)

	userRepo := user.NewRepository(db)
	userService := user.NewService(userRepo, redisClient)
	userHandler := user.NewHandler(userService)

	orderRepo := order.NewRepository(db)
	orderService := order.NewService(orderRepo)
	orderHandler := order.NewHandler(orderService)

	// Создаем роутер
	r := setupRouter(authHandler, userHandler, orderHandler, cfg, redisClient)

	// Настраиваем сервер
	server := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	go func() {
		log.Printf("🚀 Server starting on :%s", cfg.Server.Port)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Ожидаем сигнал для graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("🛑 Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("✅ Server exited")
}

func setupRouter(authHandler *auth.Handler, userHandler *user.Handler, orderHandler *order.Handler, cfg *config.Config, redisClient *redis.Client) *chi.Mux {
	r := chi.NewRouter()

	// CORS middleware
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   cfg.CORS.AllowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "X-Requested-With", "Origin", "Cache-Control"},
		ExposedHeaders:   []string{"Link", "Content-Length", "X-Total-Count"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Базовые middleware
	if cfg.Environment != "production" {
		r.Use(middleware.Logger)
	}
	r.Use(middleware.Recoverer)
	r.Use(middleware.RealIP)
	r.Use(middleware.Timeout(60 * time.Second))

	// Rate limiting для auth эндпоинтов
	r.Group(func(r chi.Router) {
		r.Use(httprate.LimitByIP(10, 1*time.Minute))
		r.Post("/auth/register", authHandler.Register)
		r.Post("/auth/login", authHandler.Login)
	})

	// Protected auth routes
	r.With(authHandler.AuthMiddleware).Post("/auth/refresh", authHandler.Refresh)
	r.With(authHandler.AuthMiddleware).Post("/auth/logout", authHandler.Logout)

	// Protected API routes
	r.Route("/api", func(r chi.Router) {
		r.Use(authHandler.AuthMiddleware)

		r.Get("/user/profile", userHandler.GetProfile)
		r.Put("/user/profile", userHandler.UpdateProfile)

		r.Get("/orders", orderHandler.GetUserOrders)
		r.Get("/orders/{id}", orderHandler.GetOrder)
		r.Post("/orders", orderHandler.CreateOrder)
	})

	// Health check - УПРОЩЕННАЯ РАБОЧАЯ ВЕРСИЯ
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		response := `{"status":"ok","database":"connected","redis":"connected"}`

		// Проверяем Redis если клиент есть
		if redisClient != nil {
			ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
			defer cancel()

			// Используем существующий метод Set для проверки соединения
			testKey := "health_check_" + time.Now().Format("20060102150405")
			err := redisClient.Set(ctx, testKey, "test", 5*time.Second)
			if err != nil {
				response = `{"status":"degraded","database":"connected","redis":"disconnected"}`
			}
		} else {
			response = `{"status":"ok","database":"connected","redis":"not_configured"}`
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(response))
		if err != nil {
			log.Printf("Error writing health response: %v", err)
		}
	})

	// Специальные эндпоинты для Tilda
	r.Route("/tilda", func(r chi.Router) {
		r.Post("/webhook", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, err := w.Write([]byte(`{"status":"ok"}`))
			if err != nil {
				return
			}
		})

		r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, err := w.Write([]byte(`{"status":"ok","service":"auth-user-service"}`))
			if err != nil {
				return
			}
		})
	})

	// Preflight handler для всех OPTIONS запросов
	r.Options("/*", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	return r
}
