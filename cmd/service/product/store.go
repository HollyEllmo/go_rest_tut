package product

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/HollyEllmo/go_rest_tut/cmd/types"
)

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) GetProducts() ([]types.Product, error) {
	rows, err := s.db.Query("SELECT * FROM products")
	if err != nil {
		return nil, err
	}

	products := make([]types.Product, 0)
	for rows.Next() {
		p, err := scanRowsIntoProduct(rows)
		if err != nil {
			return nil, err
		}
		products = append(products, *p)
	}
	return products, nil
}

func scanRowsIntoProduct(rows *sql.Rows) (*types.Product, error) {
	var p types.Product
	err := rows.Scan(
		&p.ID,
		&p.Name,
		&p.Description,
		&p.Price,
		&p.CreatedAt,
	)
	if err != nil {
		return &types.Product{}, err
	}
	return &p, nil
}

func (s *Store) CreateProduct(product *types.Product) error {
	_, err := s.db.Exec(
		"INSERT INTO products (name, description, price) VALUES (?, ?, ?)",
		product.Name,
		product.Description,
		product.Price,
	)
	if err != nil {
		return err
	}
	return nil
}

func (s *Store) GetProductsByIDs(productIDs []int) ([]types.Product, error) {
	placeholders := strings.Repeat("?,", len(productIDs)-1) + "?"
	query := fmt.Sprintf(`SELECT * FROM products WHERE id IN (%s)`, placeholders)

	// Convert Product IDs to interface slice
	args := make([]interface{}, len(productIDs))
	for i, id := range productIDs {
		args[i] = id
	}
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}

	products := []types.Product{}
	for rows.Next() {
		p, err := scanRowsIntoProduct(rows)
		if err != nil {
			return nil, err
		}
		products = append(products, *p)
	}
	return products, nil
}

func (s *Store) UpdateProduct(product types.Product) error {
	_, err := s.db.Exec(
		"UPDATE products SET name = ?, description = ?, price = ? WHERE id = ?",
		product.Name,
		product.Description,
		product.Price,
		product.ID,
	)
	if err != nil {
		return err
	}
	return nil
}
