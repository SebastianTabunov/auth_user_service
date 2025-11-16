package order

import (
	"database/sql"
	"time"
)

type Repository interface {
	GetOrder(orderID, userID int) (*Order, error)
	CreateOrder(order *Order) (int, error)
	GetUserOrders(userID int) ([]Order, error)
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

type Order struct {
	ID          int       `json:"id"`
	UserID      int       `json:"user_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Price       float64   `json:"price"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CreateOrderRequest struct {
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
}

func (r *repository) GetOrder(orderID, userID int) (*Order, error) {
	var order Order
	err := r.db.QueryRow(
		`SELECT id, user_id, title, description, price, status, created_at, updated_at 
		 FROM orders 
		 WHERE id = $1 AND user_id = $2`,
		orderID, userID,
	).Scan(
		&order.ID, &order.UserID, &order.Title, &order.Description,
		&order.Price, &order.Status, &order.CreatedAt, &order.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &order, nil
}

func (r *repository) CreateOrder(order *Order) (int, error) {
	var id int
	err := r.db.QueryRow(
		`INSERT INTO orders (user_id, title, description, price, status) 
		 VALUES ($1, $2, $3, $4, $5) 
		 RETURNING id, created_at, updated_at`,
		order.UserID, order.Title, order.Description, order.Price, "pending",
	).Scan(&id, &order.CreatedAt, &order.UpdatedAt)

	return id, err
}

func (r *repository) GetUserOrders(userID int) ([]Order, error) {
	rows, err := r.db.Query(
		`SELECT id, user_id, title, description, price, status, created_at, updated_at 
		 FROM orders 
		 WHERE user_id = $1 
		 ORDER BY created_at DESC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []Order
	for rows.Next() {
		var order Order
		err := rows.Scan(
			&order.ID, &order.UserID, &order.Title, &order.Description,
			&order.Price, &order.Status, &order.CreatedAt, &order.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return orders, nil
}
