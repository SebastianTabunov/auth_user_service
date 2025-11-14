package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
)

var db *sql.DB

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

func (c *DatabaseConfig) GetConnectionString() string {
	return "postgres://" + c.User + ":" + c.Password + "@" + c.Host + ":" + c.Port + "/" + c.DBName + "?sslmode=" + c.SSLMode
}

func NewConnection(cfg DatabaseConfig) (*sql.DB, error) {
	dsn := cfg.GetConnectionString()

	var err error
	db, err = sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	// Настройки пула соединений
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Проверка подключения
	if err := db.Ping(); err != nil {
		return nil, err
	}

	log.Println("✅ PostgreSQL connected successfully")
	return db, nil
}

func Ping() error {
	if db == nil {
		return fmt.Errorf("database not initialized")
	}
	return db.Ping()
}

func GetDB() *sql.DB {
	return db
}
