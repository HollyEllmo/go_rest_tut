package address

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

// GetUserAddresses retrieves all addresses for a user
func (s *Store) GetUserAddresses(userID int) ([]types.UserAddress, error) {
	query := `
		SELECT id, user_id, title, first_name, last_name, company, 
		       address_line_1, address_line_2, city, state_province, 
		       postal_code, country, phone, is_default, created_at, updated_at
		FROM user_addresses 
		WHERE user_id = ? 
		ORDER BY is_default DESC, created_at DESC
	`

	rows, err := s.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user addresses: %w", err)
	}
	defer rows.Close()

	var addresses []types.UserAddress
	for rows.Next() {
		address, err := s.scanRowIntoAddress(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan address row: %w", err)
		}
		addresses = append(addresses, *address)
	}

	return addresses, nil
}

// GetAddressByID retrieves a specific address for a user
func (s *Store) GetAddressByID(addressID, userID int) (*types.UserAddress, error) {
	query := `
		SELECT id, user_id, title, first_name, last_name, company, 
		       address_line_1, address_line_2, city, state_province, 
		       postal_code, country, phone, is_default, created_at, updated_at
		FROM user_addresses 
		WHERE id = ? AND user_id = ?
	`

	row := s.db.QueryRow(query, addressID, userID)
	return s.scanRowIntoAddress(row)
}

// CreateAddress creates a new address for a user
func (s *Store) CreateAddress(userID int, payload types.CreateAddressPayload) (*types.UserAddress, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// If this is set as default, unset other default addresses
	if payload.IsDefault {
		err = s.unsetDefaultAddresses(tx, userID)
		if err != nil {
			return nil, fmt.Errorf("failed to unset default addresses: %w", err)
		}
	}

	// Insert new address
	query := `
		INSERT INTO user_addresses 
		(user_id, title, first_name, last_name, company, address_line_1, address_line_2, 
		 city, state_province, postal_code, country, phone, is_default)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	result, err := tx.Exec(query, userID, payload.Title, payload.FirstName, payload.LastName,
		payload.Company, payload.AddressLine1, payload.AddressLine2, payload.City,
		payload.StateProvince, payload.PostalCode, payload.Country, payload.Phone, payload.IsDefault)

	if err != nil {
		return nil, fmt.Errorf("failed to create address: %w", err)
	}

	addressID, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get address ID: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Return the created address
	return s.GetAddressByID(int(addressID), userID)
}

// UpdateAddress updates an existing address
func (s *Store) UpdateAddress(addressID, userID int, payload types.UpdateAddressPayload) (*types.UserAddress, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// If setting as default, unset other default addresses
	if payload.IsDefault != nil && *payload.IsDefault {
		err = s.unsetDefaultAddresses(tx, userID)
		if err != nil {
			return nil, fmt.Errorf("failed to unset default addresses: %w", err)
		}
	}

	// Build dynamic update query
	setParts := []string{}
	args := []any{}

	if payload.Title != nil {
		setParts = append(setParts, "title = ?")
		args = append(args, *payload.Title)
	}
	if payload.FirstName != nil {
		setParts = append(setParts, "first_name = ?")
		args = append(args, *payload.FirstName)
	}
	if payload.LastName != nil {
		setParts = append(setParts, "last_name = ?")
		args = append(args, *payload.LastName)
	}
	if payload.Company != nil {
		setParts = append(setParts, "company = ?")
		args = append(args, *payload.Company)
	}
	if payload.AddressLine1 != nil {
		setParts = append(setParts, "address_line_1 = ?")
		args = append(args, *payload.AddressLine1)
	}
	if payload.AddressLine2 != nil {
		setParts = append(setParts, "address_line_2 = ?")
		args = append(args, *payload.AddressLine2)
	}
	if payload.City != nil {
		setParts = append(setParts, "city = ?")
		args = append(args, *payload.City)
	}
	if payload.StateProvince != nil {
		setParts = append(setParts, "state_province = ?")
		args = append(args, *payload.StateProvince)
	}
	if payload.PostalCode != nil {
		setParts = append(setParts, "postal_code = ?")
		args = append(args, *payload.PostalCode)
	}
	if payload.Country != nil {
		setParts = append(setParts, "country = ?")
		args = append(args, *payload.Country)
	}
	if payload.Phone != nil {
		setParts = append(setParts, "phone = ?")
		args = append(args, *payload.Phone)
	}
	if payload.IsDefault != nil {
		setParts = append(setParts, "is_default = ?")
		args = append(args, *payload.IsDefault)
	}

	if len(setParts) == 0 {
		return nil, fmt.Errorf("no fields to update")
	}

	// Add updated_at
	setParts = append(setParts, "updated_at = CURRENT_TIMESTAMP")

	// Add WHERE clause parameters
	args = append(args, addressID, userID)

	query := fmt.Sprintf(`
		UPDATE user_addresses 
		SET %s
		WHERE id = ? AND user_id = ?
	`, fmt.Sprintf("%s", setParts[0]))

	for i := 1; i < len(setParts); i++ {
		query = fmt.Sprintf(`
			UPDATE user_addresses 
			SET %s
			WHERE id = ? AND user_id = ?
		`, fmt.Sprintf("%s, %s", setParts[0], setParts[i]))
	}

	// Rebuild query properly
	setClause := ""
	for i, part := range setParts {
		if i > 0 {
			setClause += ", "
		}
		setClause += part
	}

	query = fmt.Sprintf(`
		UPDATE user_addresses 
		SET %s
		WHERE id = ? AND user_id = ?
	`, setClause)

	result, err := tx.Exec(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to update address: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return nil, fmt.Errorf("address not found or not owned by user")
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Return updated address
	return s.GetAddressByID(addressID, userID)
}

// DeleteAddress deletes an address
func (s *Store) DeleteAddress(addressID, userID int) error {
	query := "DELETE FROM user_addresses WHERE id = ? AND user_id = ?"
	result, err := s.db.Exec(query, addressID, userID)
	if err != nil {
		return fmt.Errorf("failed to delete address: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("address not found or not owned by user")
	}

	return nil
}

// GetDefaultAddress gets the default address for a user
func (s *Store) GetDefaultAddress(userID int) (*types.UserAddress, error) {
	query := `
		SELECT id, user_id, title, first_name, last_name, company, 
		       address_line_1, address_line_2, city, state_province, 
		       postal_code, country, phone, is_default, created_at, updated_at
		FROM user_addresses 
		WHERE user_id = ? AND is_default = TRUE
		LIMIT 1
	`

	row := s.db.QueryRow(query, userID)
	address, err := s.scanRowIntoAddress(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no default address found for user")
		}
		return nil, err
	}
	return address, nil
}

// SetDefaultAddress sets an address as default
func (s *Store) SetDefaultAddress(addressID, userID int) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// First, unset all default addresses for this user
	err = s.unsetDefaultAddresses(tx, userID)
	if err != nil {
		return fmt.Errorf("failed to unset default addresses: %w", err)
	}

	// Then set the specified address as default
	query := "UPDATE user_addresses SET is_default = TRUE WHERE id = ? AND user_id = ?"
	result, err := tx.Exec(query, addressID, userID)
	if err != nil {
		return fmt.Errorf("failed to set default address: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("address not found or not owned by user")
	}

	return tx.Commit()
}

// Helper function to scan database row into UserAddress struct
func (s *Store) scanRowIntoAddress(scanner interface {
	Scan(dest ...any) error
}) (*types.UserAddress, error) {
	var address types.UserAddress
	var company, addressLine2, phone sql.NullString

	err := scanner.Scan(
		&address.ID,
		&address.UserID,
		&address.Title,
		&address.FirstName,
		&address.LastName,
		&company,
		&address.AddressLine1,
		&addressLine2,
		&address.City,
		&address.StateProvince,
		&address.PostalCode,
		&address.Country,
		&phone,
		&address.IsDefault,
		&address.CreatedAt,
		&address.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("address not found")
		}
		return nil, err
	}

	// Handle nullable fields
	if company.Valid {
		address.Company = &company.String
	}
	if addressLine2.Valid {
		address.AddressLine2 = &addressLine2.String
	}
	if phone.Valid {
		address.Phone = &phone.String
	}

	return &address, nil
}

// Helper function to unset all default addresses for a user (within transaction)
func (s *Store) unsetDefaultAddresses(tx *sql.Tx, userID int) error {
	query := "UPDATE user_addresses SET is_default = FALSE WHERE user_id = ? AND is_default = TRUE"
	_, err := tx.Exec(query, userID)
	return err
}