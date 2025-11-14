package config

import (
	"os"
	"strings"
)

type Config struct {
	Environment string
	Server      ServerConfig
	Database    DatabaseConfig
	Redis       RedisConfig
	JWT         JWTConfig
	CORS        CORSConfig
}

type ServerConfig struct {
	Port string
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type RedisConfig struct {
	URL string
}

type JWTConfig struct {
	Secret string
}

type CORSConfig struct {
	AllowedOrigins []string
}

func Load() *Config {
	return &Config{
		Environment: getEnv("APP_ENV", getEnv("ENVIRONMENT", "development")),
		Server: ServerConfig{
			Port: getEnv("PORT", "8080"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "user"),
			Password: getEnv("DB_PASSWORD", "password"),
			DBName:   getEnv("DB_NAME", "auth_service"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		Redis: RedisConfig{
			URL: getEnv("REDIS_URL", ""),
		},
		JWT: JWTConfig{
			Secret: getEnv("JWT_SECRET", ""),
		},
		CORS: CORSConfig{
			AllowedOrigins: getCORSAllowedOrigins(),
		},
	}
}

func (c *DatabaseConfig) GetConnectionString() string {
	return "postgres://" + c.User + ":" + c.Password + "@" + c.Host + ":" + c.Port + "/" + c.DBName + "?sslmode=" + c.SSLMode
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getCORSAllowedOrigins() []string {
	// Allow all origins by default
	corsOrigins := getEnv("CORS_ALLOWED_ORIGINS", "")
	if corsOrigins != "" {
		return strings.Split(corsOrigins, ",")
	}
	
	// Allow all origins in all environments
	return []string{"*"}
}
