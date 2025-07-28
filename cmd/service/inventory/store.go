package inventory

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

// GetCurrentStock вычисляет текущий остаток товара на основе всех движений
func (s *Store) GetCurrentStock(productID int) (int, error) {
	query := `
		SELECT COALESCE(SUM(
			CASE WHEN movement_type = 'IN' THEN quantity 
			     ELSE -quantity 
			END
		), 0) as current_stock
		FROM inventory_movements 
		WHERE product_id = ?
	`
	
	var stock int
	err := s.db.QueryRow(query, productID).Scan(&stock)
	if err != nil {
		return 0, fmt.Errorf("failed to get current stock for product %d: %w", productID, err)
	}
	
	return stock, nil
}

// GetProductsWithStock получает остатки для нескольких товаров одним запросом
func (s *Store) GetProductsWithStock(productIDs []int) (map[int]int, error) {
	if len(productIDs) == 0 {
		return make(map[int]int), nil
	}
	
	// Создаём плейсхолдеры для IN clause
	placeholders := ""
	args := make([]any, len(productIDs))
	for i, id := range productIDs {
		if i > 0 {
			placeholders += ","
		}
		placeholders += "?"
		args[i] = id
	}
	
	query := fmt.Sprintf(`
		SELECT 
			product_id,
			COALESCE(SUM(
				CASE WHEN movement_type = 'IN' THEN quantity 
				     ELSE -quantity 
				END
			), 0) as current_stock
		FROM inventory_movements 
		WHERE product_id IN (%s)
		GROUP BY product_id
	`, placeholders)
	
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get stock for products: %w", err)
	}
	defer rows.Close()
	
	stockMap := make(map[int]int)
	
	// Инициализируем все продукты нулевым остатком
	for _, id := range productIDs {
		stockMap[id] = 0
	}
	
	// Обновляем фактическими остатками
	for rows.Next() {
		var productID, stock int
		if err := rows.Scan(&productID, &stock); err != nil {
			return nil, fmt.Errorf("failed to scan stock row: %w", err)
		}
		stockMap[productID] = stock
	}
	
	return stockMap, nil
}

// ReserveStock резервирует товар для заказа (атомарная операция)
func (s *Store) ReserveStock(productID, quantity int, orderID int) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()
	
	// Получаем текущий остаток с блокировкой
	var currentStock int
	err = tx.QueryRow(`
		SELECT COALESCE(SUM(
			CASE WHEN movement_type = 'IN' THEN quantity 
			     ELSE -quantity 
			END
		), 0)
		FROM inventory_movements 
		WHERE product_id = ?
		FOR UPDATE
	`, productID).Scan(&currentStock)
	
	if err != nil {
		return fmt.Errorf("failed to get current stock: %w", err)
	}
	
	// Проверяем достаточность товара
	if currentStock < quantity {
		return fmt.Errorf("insufficient stock for product %d: available %d, requested %d", 
			productID, currentStock, quantity)
	}
	
	// Создаём запись о резервировании
	_, err = tx.Exec(`
		INSERT INTO inventory_movements 
		(product_id, movement_type, quantity, reason, reference_id, reference_type)
		VALUES (?, 'OUT', ?, 'Reserved for order', ?, 'ORDER')
	`, productID, quantity, orderID)
	
	if err != nil {
		return fmt.Errorf("failed to reserve stock: %w", err)
	}
	
	return tx.Commit()
}

// ReleaseStock освобождает зарезервированный товар
func (s *Store) ReleaseStock(productID, quantity int, reason string) error {
	_, err := s.db.Exec(`
		INSERT INTO inventory_movements 
		(product_id, movement_type, quantity, reason)
		VALUES (?, 'IN', ?, ?)
	`, productID, quantity, reason)
	
	if err != nil {
		return fmt.Errorf("failed to release stock: %w", err)
	}
	
	return nil
}

// AddStock добавляет товар на склад
func (s *Store) AddStock(productID, quantity int, reason string, refType types.InventoryRefType, refID *int) error {
	_, err := s.db.Exec(`
		INSERT INTO inventory_movements 
		(product_id, movement_type, quantity, reason, reference_id, reference_type)
		VALUES (?, 'IN', ?, ?, ?, ?)
	`, productID, quantity, reason, refID, refType)
	
	if err != nil {
		return fmt.Errorf("failed to add stock: %w", err)
	}
	
	return nil
}

// GetStockHistory возвращает историю движений товара
func (s *Store) GetStockHistory(productID int, limit int) ([]types.InventoryMovement, error) {
	query := `
		SELECT id, product_id, movement_type, quantity, reason, 
		       reference_id, reference_type, created_at
		FROM inventory_movements 
		WHERE product_id = ?
		ORDER BY created_at DESC
		LIMIT ?
	`
	
	rows, err := s.db.Query(query, productID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get stock history: %w", err)
	}
	defer rows.Close()
	
	var movements []types.InventoryMovement
	for rows.Next() {
		var movement types.InventoryMovement
		var refID sql.NullInt64
		var refType sql.NullString
		
		err := rows.Scan(
			&movement.ID,
			&movement.ProductID,
			&movement.MovementType,
			&movement.Quantity,
			&movement.Reason,
			&refID,
			&refType,
			&movement.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan movement row: %w", err)
		}
		
		if refID.Valid {
			id := int(refID.Int64)
			movement.ReferenceID = &id
		}
		if refType.Valid {
			rt := types.InventoryRefType(refType.String)
			movement.ReferenceType = &rt
		}
		
		movements = append(movements, movement)
	}
	
	return movements, nil
}