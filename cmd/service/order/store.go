package cart

import (
	"database/sql"

	"github.com/HollyEllmo/go_rest_tut/cmd/types"
)

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) CreateOrder(order types.Order) (int, error) {
	rew, err := s.db.Exec(
		"INSERT INTO orders (userId, total, status, address) VALUES (?, ?, ?, ?)",
		order.UserID,
		order.Total,
		order.Status,
		order.Address,
	)
	if err != nil {
		return 0, err
	}
	id, err := rew.LastInsertId()
	if err != nil {
		return 0, err
	}
	return int(id), nil
}

func (s *Store) CreateOrderItem(orderItem types.OrderItem) error {
	_, err := s.db.Exec(
		"INSERT INTO order_items (orderId, productId, quantity, price) VALUES (?, ?, ?, ?)",
		orderItem.OrderID,
		orderItem.ProductID,
		orderItem.Quantity,
		orderItem.Price,
	)
	if err != nil {
		return err
	}
	return nil
}
