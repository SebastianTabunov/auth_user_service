package auth

import (
	"database/sql"
	"errors"
	"time"
)

// Repository интерфейс
type Repository interface {
	CreateUser(email, passwordHash, firstName, lastName string) (int, error)
	GetUserByEmail(email string) (*User, error)
	GetUserByID(id int) (*User, error)
	UserExists(email string) (bool, error)
	SaveRefreshToken(userID int, token string, expiresAt time.Time) error
	GetUserByRefreshToken(token string) (*User, error)
	DeleteRefreshToken(token string) error
}

// User представляет пользователя системы
type User struct {
	ID           int       `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	FirstName    string    `json:"first_name,omitempty"`
	LastName     string    `json:"last_name,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// RegisterRequest структура для регистрации
type RegisterRequest struct {
	Email     string `json:"email"`
	Password  string `json:"password"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

// LoginRequest структура для входа
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Validate валидирует структуру запроса
func (r *RegisterRequest) Validate() error {
	// Простая валидация
	if r.Email == "" || r.Password == "" || r.FirstName == "" || r.LastName == "" {
		return errors.New("all fields are required")
	}
	return nil
}

func (r *LoginRequest) Validate() error {
	if r.Email == "" || r.Password == "" {
		return errors.New("email and password are required")
	}
	return nil
}

// PostgreSQL реализация репозитория
type postgresRepository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &postgresRepository{db: db}
}

func (r *postgresRepository) CreateUser(email, passwordHash, firstName, lastName string) (int, error) {
	var id int
	err := r.db.QueryRow(
		"INSERT INTO users (email, password_hash, first_name, last_name) VALUES ($1, $2, $3, $4) RETURNING id",
		email, passwordHash, firstName, lastName,
	).Scan(&id)

	if err != nil {
		return 0, err
	}

	return id, nil
}

func (r *postgresRepository) GetUserByEmail(email string) (*User, error) {
	var user User
	err := r.db.QueryRow(
		"SELECT id, email, password_hash, first_name, last_name, created_at, updated_at FROM users WHERE email = $1",
		email,
	).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.FirstName, &user.LastName, &user.CreatedAt, &user.UpdatedAt)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("user not found")
	}
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *postgresRepository) GetUserByID(id int) (*User, error) {
	var user User
	err := r.db.QueryRow(
		"SELECT id, email, password_hash, first_name, last_name, created_at, updated_at FROM users WHERE id = $1",
		id,
	).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.FirstName, &user.LastName, &user.CreatedAt, &user.UpdatedAt)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("user not found")
	}
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *postgresRepository) UserExists(email string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)",
		email,
	).Scan(&exists)
	return exists, err
}

func (r *postgresRepository) SaveRefreshToken(userID int, token string, expiresAt time.Time) error {
	_, err := r.db.Exec(
		"INSERT INTO auth_tokens (user_id, token, expires_at) VALUES ($1, $2, $3)",
		userID, token, expiresAt,
	)
	return err
}

func (r *postgresRepository) GetUserByRefreshToken(token string) (*User, error) {
	var user User
	err := r.db.QueryRow(
		`SELECT u.id, u.email, u.password_hash, u.first_name, u.last_name, u.created_at, u.updated_at 
		 FROM users u 
		 JOIN auth_tokens t ON u.id = t.user_id 
		 WHERE t.token = $1 AND t.expires_at > $2`,
		token, time.Now(),
	).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.FirstName, &user.LastName, &user.CreatedAt, &user.UpdatedAt)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("invalid or expired refresh token")
	}
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *postgresRepository) DeleteRefreshToken(token string) error {
	_, err := r.db.Exec(
		"DELETE FROM auth_tokens WHERE token = $1",
		token,
	)
	return err
}
