package order

import (
	"database/sql"
	"fmt"

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

// GetUserOrders retrieves all orders for a user with filtering and pagination
func (s *Store) GetUserOrders(userID int, filters types.OrderFilters) ([]types.OrderWithItems, error) {
	query := `
		SELECT DISTINCT o.id, o.userId, o.total, o.status, o.address, o.createdAt
		FROM orders o
		WHERE o.userId = ?
	`
	args := []interface{}{userID}

	// Add status filter if provided
	if filters.Status != nil {
		query += " AND o.status = ?"
		args = append(args, *filters.Status)
	}

	// Add date range filters if provided
	if filters.FromDate != nil {
		query += " AND o.createdAt >= ?"
		args = append(args, *filters.FromDate)
	}

	if filters.ToDate != nil {
		query += " AND o.createdAt <= ?"
		args = append(args, *filters.ToDate)
	}

	// Add ordering and pagination
	query += " ORDER BY o.createdAt DESC"

	if filters.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, filters.Limit)
	}

	if filters.Offset > 0 {
		query += " OFFSET ?"
		args = append(args, filters.Offset)
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get user orders: %w", err)
	}
	defer rows.Close()

	var orders []types.OrderWithItems
	for rows.Next() {
		var order types.OrderWithItems
		err := rows.Scan(
			&order.ID,
			&order.UserID,
			&order.Total,
			&order.Status,
			&order.Address,
			&order.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan order: %w", err)
		}

		// Get items for this order
		items, err := s.getOrderItems(order.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get order items for order %d: %w", order.ID, err)
		}
		order.Items = items

		orders = append(orders, order)
	}

	return orders, nil
}

// GetOrderByID retrieves a specific order by ID for a user
func (s *Store) GetOrderByID(orderID, userID int) (*types.OrderWithItems, error) {
	query := `
		SELECT o.id, o.userId, o.total, o.status, o.address, o.createdAt
		FROM orders o
		WHERE o.id = ? AND o.userId = ?
	`

	var order types.OrderWithItems
	err := s.db.QueryRow(query, orderID, userID).Scan(
		&order.ID,
		&order.UserID,
		&order.Total,
		&order.Status,
		&order.Address,
		&order.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("order not found or not owned by user")
		}
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	// Get items for this order
	items, err := s.getOrderItems(order.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get order items: %w", err)
	}
	order.Items = items

	return &order, nil
}

// GetOrdersCount returns the total count of orders for a user with filters
func (s *Store) GetOrdersCount(userID int, filters types.OrderFilters) (int, error) {
	query := "SELECT COUNT(*) FROM orders WHERE userId = ?"
	args := []interface{}{userID}

	// Add status filter if provided
	if filters.Status != nil {
		query += " AND status = ?"
		args = append(args, *filters.Status)
	}

	// Add date range filters if provided
	if filters.FromDate != nil {
		query += " AND createdAt >= ?"
		args = append(args, *filters.FromDate)
	}

	if filters.ToDate != nil {
		query += " AND createdAt <= ?"
		args = append(args, *filters.ToDate)
	}

	var count int
	err := s.db.QueryRow(query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get orders count: %w", err)
	}

	return count, nil
}

// getOrderItems retrieves all items for a specific order with product details
func (s *Store) getOrderItems(orderID int) ([]types.OrderItemWithProduct, error) {
	query := `
		SELECT 
			oi.id, oi.orderId, oi.productId, oi.quantity, oi.price,
			p.name, p.image
		FROM order_items oi
		JOIN products p ON oi.productId = p.id
		WHERE oi.orderId = ?
		ORDER BY oi.id
	`

	rows, err := s.db.Query(query, orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to query order items: %w", err)
	}
	defer rows.Close()

	var items []types.OrderItemWithProduct
	for rows.Next() {
		var item types.OrderItemWithProduct
		err := rows.Scan(
			&item.ID,
			&item.OrderID,
			&item.ProductID,
			&item.Quantity,
			&item.Price,
			&item.ProductName,
			&item.ProductImage,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan order item: %w", err)
		}
		items = append(items, item)
	}

	return items, nil
}
