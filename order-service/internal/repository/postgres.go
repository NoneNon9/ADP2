package repository

import (
	"database/sql"
	"errors"

	"order-service/internal/domain"
)

type PostgresOrderRepository struct {
	db *sql.DB
}

func NewPostgresOrderRepository(db *sql.DB) *PostgresOrderRepository {
	return &PostgresOrderRepository{db: db}
}

func (r *PostgresOrderRepository) Save(order domain.Order) error {
	query := `
		INSERT INTO orders (id, customer_id, item_name, amount, status, idempotency_key, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	var idemKey sql.NullString
	if order.IdempotencyKey != "" {
		idemKey = sql.NullString{String: order.IdempotencyKey, Valid: true}
	}

	_, err := r.db.Exec(query,
		order.ID,
		order.CustomerID,
		order.ItemName,
		order.Amount,
		order.Status,
		idemKey,
		order.CreatedAt,
	)
	return err
}

func (r *PostgresOrderRepository) GetByID(id string) (domain.Order, error) {
	query := `SELECT id, customer_id, item_name, amount, status, idempotency_key, created_at FROM orders WHERE id = $1`
	row := r.db.QueryRow(query, id)

	var order domain.Order
	var idemKey sql.NullString

	err := row.Scan(
		&order.ID,
		&order.CustomerID,
		&order.ItemName,
		&order.Amount,
		&order.Status,
		&idemKey,
		&order.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return order, errors.New("order not found")
	}

	if idemKey.Valid {
		order.IdempotencyKey = idemKey.String
	}

	return order, err
}

func (r *PostgresOrderRepository) GetByIdempotencyKey(key string) (domain.Order, error) {
	query := `SELECT id, customer_id, item_name, amount, status, idempotency_key, created_at FROM orders WHERE idempotency_key = $1`
	row := r.db.QueryRow(query, key)

	var order domain.Order
	var idemKey sql.NullString

	err := row.Scan(
		&order.ID,
		&order.CustomerID,
		&order.ItemName,
		&order.Amount,
		&order.Status,
		&idemKey,
		&order.CreatedAt,
	)

	if idemKey.Valid {
		order.IdempotencyKey = idemKey.String
	}

	return order, err
}

func (r *PostgresOrderRepository) UpdateStatus(id string, status string) error {
	query := `UPDATE orders SET status = $1 WHERE id = $2`
	_, err := r.db.Exec(query, status, id)
	return err
}
