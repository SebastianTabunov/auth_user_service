package user

import (
	"database/sql"
	"time"
)

type Repository interface {
	GetProfile(userID int) (*Profile, error)
	UpdateProfile(userID int, profile *Profile) error
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

type Profile struct {
	ID        int       `json:"id"`
	Email     string    `json:"email"`
	FirstName string    `json:"first_name,omitempty"`
	LastName  string    `json:"last_name,omitempty"`
	Phone     string    `json:"phone,omitempty"`
	Address   string    `json:"address,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type UpdateProfileRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Phone     string `json:"phone"`
	Address   string `json:"address"`
}

func (r *repository) GetProfile(userID int) (*Profile, error) {
	var profile Profile
	err := r.db.QueryRow(
		`SELECT u.id, u.email, COALESCE(u.first_name, ''), COALESCE(u.last_name, ''),
		 COALESCE(p.phone, ''), COALESCE(p.address, ''), u.created_at, COALESCE(p.updated_at, u.created_at)
		 FROM users u 
		 LEFT JOIN user_profiles p ON u.id = p.id 
		 WHERE u.id = $1`,
		userID,
	).Scan(
		&profile.ID, &profile.Email, &profile.FirstName, &profile.LastName,
		&profile.Phone, &profile.Address, &profile.CreatedAt, &profile.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &profile, nil
}

func (r *repository) UpdateProfile(userID int, profile *Profile) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Обновляем first_name и last_name в таблице users
	_, err = tx.Exec(
		`UPDATE users 
		 SET first_name = $1, last_name = $2, updated_at = NOW()
		 WHERE id = $3`,
		profile.FirstName, profile.LastName, userID,
	)
	if err != nil {
		return err
	}

	// Проверяем, существует ли профиль в user_profiles
	var exists bool
	err = tx.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM user_profiles WHERE id = $1)",
		userID,
	).Scan(&exists)

	if err != nil {
		return err
	}

	if exists {
		// Обновляем существующий профиль
		_, err = tx.Exec(
			`UPDATE user_profiles
			 SET phone = $1, address = $2, updated_at = NOW()
			 WHERE id = $3`,
			profile.Phone, profile.Address, userID,
		)
	} else {
		// Создаем новый профиль
		_, err = tx.Exec(
			`INSERT INTO user_profiles (id, phone, address)
			 VALUES ($1, $2, $3)`,
			userID, profile.Phone, profile.Address,
		)
	}

	if err != nil {
		return err
	}

	return tx.Commit()
}
