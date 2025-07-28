package address

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/HollyEllmo/go_rest_tut/cmd/config"
	"github.com/HollyEllmo/go_rest_tut/cmd/db"
	"github.com/HollyEllmo/go_rest_tut/cmd/types"
	"github.com/go-sql-driver/mysql"
	_ "github.com/go-sql-driver/mysql"
)

var testDB *sql.DB
var addressStore *Store

func TestMain(m *testing.M) {
	cfg := config.Envs
	
	// Connect to test database
	testDBName := "go_rest_tut_test"
	var err error
	testDB, err = db.NewMySQLStorage(mysql.Config{
		User:                 cfg.DBUser,
		Passwd:               cfg.DBPassword,
		Net:                  "tcp",
		Addr:                 cfg.DBAddress,
		DBName:               testDBName,
		AllowNativePasswords: true,
		ParseTime:            true,
	})
	if err != nil {
		log.Fatalf("Failed to connect to test database: %v", err)
	}

	// Create test database if it doesn't exist
	setupTestDB(cfg, testDBName)
	
	// Run migrations on test database
	runTestMigrations()
	
	addressStore = NewStore(testDB)
	
	// Run tests
	code := m.Run()
	
	// Cleanup
	cleanupTestDB()
	testDB.Close()
	
	os.Exit(code)
}

func setupTestDB(cfg config.Config, testDBName string) {
	// Connect without database to create test database
	mainDB, err := db.NewMySQLStorage(mysql.Config{
		User:                 cfg.DBUser,
		Passwd:               cfg.DBPassword,
		Net:                  "tcp",
		Addr:                 cfg.DBAddress,
		DBName:               "",
		AllowNativePasswords: true,
		ParseTime:            true,
	})
	if err != nil {
		log.Fatalf("Failed to connect to main database: %v", err)
	}
	defer mainDB.Close()

	_, err = mainDB.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s", testDBName))
	if err != nil {
		log.Fatalf("Failed to create test database: %v", err)
	}
}

func runTestMigrations() {
	// Create users table for foreign key constraint
	usersTableSQL := `
		CREATE TABLE IF NOT EXISTS users (
			id INT AUTO_INCREMENT PRIMARY KEY,
			firstName VARCHAR(100) NOT NULL,
			lastName VARCHAR(100) NOT NULL,
			email VARCHAR(100) NOT NULL UNIQUE,
			password VARCHAR(255) NOT NULL,
			createdAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`
	
	// Create user_addresses table
	addressesTableSQL := `
		CREATE TABLE IF NOT EXISTS user_addresses (
			id INT AUTO_INCREMENT PRIMARY KEY,
			user_id INT NOT NULL,
			title VARCHAR(100) NOT NULL,
			first_name VARCHAR(100) NOT NULL,
			last_name VARCHAR(100) NOT NULL,
			company VARCHAR(100),
			address_line_1 VARCHAR(255) NOT NULL,
			address_line_2 VARCHAR(255),
			city VARCHAR(100) NOT NULL,
			state_province VARCHAR(100) NOT NULL,
			postal_code VARCHAR(20) NOT NULL,
			country VARCHAR(100) NOT NULL,
			phone VARCHAR(20),
			is_default BOOLEAN DEFAULT FALSE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
			INDEX idx_user_addresses_user_id (user_id),
			INDEX idx_user_addresses_default (user_id, is_default)
		)
	`
	
	if _, err := testDB.Exec(usersTableSQL); err != nil {
		log.Fatalf("Failed to create users table: %v", err)
	}
	
	if _, err := testDB.Exec(addressesTableSQL); err != nil {
		log.Fatalf("Failed to create user_addresses table: %v", err)
	}
}

func cleanupTestDB() {
	testDB.Exec("DROP TABLE IF EXISTS user_addresses")
	testDB.Exec("DROP TABLE IF EXISTS users")
}

func setupTestUser() int {
	_, err := testDB.Exec(`
		INSERT INTO users (firstName, lastName, email, password) 
		VALUES ('Test', 'User', 'test@example.com', 'hashedpassword')
	`)
	if err != nil {
		log.Fatalf("Failed to create test user: %v", err)
	}
	
	var userID int
	err = testDB.QueryRow("SELECT id FROM users WHERE email = 'test@example.com'").Scan(&userID)
	if err != nil {
		log.Fatalf("Failed to get test user ID: %v", err)
	}
	
	return userID
}

func cleanupTestData() {
	testDB.Exec("DELETE FROM user_addresses")
	testDB.Exec("DELETE FROM users")
}

func TestAddressStore_CreateAddress(t *testing.T) {
	defer cleanupTestData()
	userID := setupTestUser()
	
	payload := types.CreateAddressPayload{
		Title:         "Home",
		FirstName:     "John",
		LastName:      "Doe",
		AddressLine1:  "123 Main St",
		City:          "New York",
		StateProvince: "NY",
		PostalCode:    "10001",
		Country:       "USA",
		IsDefault:     true,
	}
	
	address, err := addressStore.CreateAddress(userID, payload)
	if err != nil {
		t.Fatalf("Failed to create address: %v", err)
	}
	
	if address.ID == 0 {
		t.Error("Expected address ID to be set")
	}
	if address.UserID != userID {
		t.Errorf("Expected user ID %d, got %d", userID, address.UserID)
	}
	if address.Title != payload.Title {
		t.Errorf("Expected title %s, got %s", payload.Title, address.Title)
	}
	if !address.IsDefault {
		t.Error("Expected address to be default")
	}
}

func TestAddressStore_GetUserAddresses(t *testing.T) {
	defer cleanupTestData()
	userID := setupTestUser()
	
	// Create two addresses
	payload1 := types.CreateAddressPayload{
		Title:         "Home",
		FirstName:     "John",
		LastName:      "Doe",
		AddressLine1:  "123 Main St",
		City:          "New York",
		StateProvince: "NY",
		PostalCode:    "10001",
		Country:       "USA",
		IsDefault:     true,
	}
	
	payload2 := types.CreateAddressPayload{
		Title:         "Work",
		FirstName:     "John",
		LastName:      "Doe",
		AddressLine1:  "456 Office Blvd",
		City:          "New York",
		StateProvince: "NY",
		PostalCode:    "10002",
		Country:       "USA",
		IsDefault:     false,
	}
	
	_, err := addressStore.CreateAddress(userID, payload1)
	if err != nil {
		t.Fatalf("Failed to create first address: %v", err)
	}
	
	_, err = addressStore.CreateAddress(userID, payload2)
	if err != nil {
		t.Fatalf("Failed to create second address: %v", err)
	}
	
	addresses, err := addressStore.GetUserAddresses(userID)
	if err != nil {
		t.Fatalf("Failed to get user addresses: %v", err)
	}
	
	if len(addresses) != 2 {
		t.Errorf("Expected 2 addresses, got %d", len(addresses))
	}
	
	// Check that default address comes first
	if !addresses[0].IsDefault {
		t.Error("Expected first address to be default")
	}
}

func TestAddressStore_GetAddressByID(t *testing.T) {
	defer cleanupTestData()
	userID := setupTestUser()
	
	payload := types.CreateAddressPayload{
		Title:         "Home",
		FirstName:     "John",
		LastName:      "Doe",
		AddressLine1:  "123 Main St",
		City:          "New York",
		StateProvince: "NY",
		PostalCode:    "10001",
		Country:       "USA",
		IsDefault:     true,
	}
	
	createdAddress, err := addressStore.CreateAddress(userID, payload)
	if err != nil {
		t.Fatalf("Failed to create address: %v", err)
	}
	
	// Test getting existing address
	address, err := addressStore.GetAddressByID(createdAddress.ID, userID)
	if err != nil {
		t.Fatalf("Failed to get address by ID: %v", err)
	}
	
	if address.ID != createdAddress.ID {
		t.Errorf("Expected address ID %d, got %d", createdAddress.ID, address.ID)
	}
	
	// Test getting non-existent address
	_, err = addressStore.GetAddressByID(9999, userID)
	if err == nil {
		t.Error("Expected error when getting non-existent address")
	}
}

func TestAddressStore_UpdateAddress(t *testing.T) {
	defer cleanupTestData()
	userID := setupTestUser()
	
	payload := types.CreateAddressPayload{
		Title:         "Home",
		FirstName:     "John",
		LastName:      "Doe",
		AddressLine1:  "123 Main St",
		City:          "New York",
		StateProvince: "NY",
		PostalCode:    "10001",
		Country:       "USA",
		IsDefault:     false,
	}
	
	createdAddress, err := addressStore.CreateAddress(userID, payload)
	if err != nil {
		t.Fatalf("Failed to create address: %v", err)
	}
	
	// Update address
	newTitle := "Updated Home"
	newCity := "Boston"
	updatePayload := types.UpdateAddressPayload{
		Title: &newTitle,
		City:  &newCity,
	}
	
	updatedAddress, err := addressStore.UpdateAddress(createdAddress.ID, userID, updatePayload)
	if err != nil {
		t.Fatalf("Failed to update address: %v", err)
	}
	
	if updatedAddress.Title != newTitle {
		t.Errorf("Expected title %s, got %s", newTitle, updatedAddress.Title)
	}
	if updatedAddress.City != newCity {
		t.Errorf("Expected city %s, got %s", newCity, updatedAddress.City)
	}
	// Other fields should remain unchanged
	if updatedAddress.FirstName != payload.FirstName {
		t.Errorf("Expected first name to remain %s, got %s", payload.FirstName, updatedAddress.FirstName)
	}
}

func TestAddressStore_DeleteAddress(t *testing.T) {
	defer cleanupTestData()
	userID := setupTestUser()
	
	payload := types.CreateAddressPayload{
		Title:         "Home",
		FirstName:     "John",
		LastName:      "Doe",
		AddressLine1:  "123 Main St",
		City:          "New York",
		StateProvince: "NY",
		PostalCode:    "10001",
		Country:       "USA",
		IsDefault:     false,
	}
	
	createdAddress, err := addressStore.CreateAddress(userID, payload)
	if err != nil {
		t.Fatalf("Failed to create address: %v", err)
	}
	
	// Delete address
	err = addressStore.DeleteAddress(createdAddress.ID, userID)
	if err != nil {
		t.Fatalf("Failed to delete address: %v", err)
	}
	
	// Verify address is deleted
	_, err = addressStore.GetAddressByID(createdAddress.ID, userID)
	if err == nil {
		t.Error("Expected error when getting deleted address")
	}
}

func TestAddressStore_DefaultAddressLogic(t *testing.T) {
	defer cleanupTestData()
	userID := setupTestUser()
	
	// Create first address as default
	payload1 := types.CreateAddressPayload{
		Title:         "Home",
		FirstName:     "John",
		LastName:      "Doe",
		AddressLine1:  "123 Main St",
		City:          "New York",
		StateProvince: "NY",
		PostalCode:    "10001",
		Country:       "USA",
		IsDefault:     true,
	}
	
	address1, err := addressStore.CreateAddress(userID, payload1)
	if err != nil {
		t.Fatalf("Failed to create first address: %v", err)
	}
	
	// Create second address as default (should unset first)
	payload2 := types.CreateAddressPayload{
		Title:         "Work",
		FirstName:     "John",
		LastName:      "Doe",
		AddressLine1:  "456 Office Blvd",
		City:          "New York",
		StateProvince: "NY",
		PostalCode:    "10002",
		Country:       "USA",
		IsDefault:     true,
	}
	
	address2, err := addressStore.CreateAddress(userID, payload2)
	if err != nil {
		t.Fatalf("Failed to create second address: %v", err)
	}
	
	// Check that only the second address is default
	defaultAddress, err := addressStore.GetDefaultAddress(userID)
	if err != nil {
		t.Fatalf("Failed to get default address: %v", err)
	}
	
	if defaultAddress.ID != address2.ID {
		t.Errorf("Expected default address ID %d, got %d", address2.ID, defaultAddress.ID)
	}
	
	// Check that first address is no longer default
	firstAddress, err := addressStore.GetAddressByID(address1.ID, userID)
	if err != nil {
		t.Fatalf("Failed to get first address: %v", err)
	}
	
	if firstAddress.IsDefault {
		t.Error("Expected first address to no longer be default")
	}
}

func TestAddressStore_SetDefaultAddress(t *testing.T) {
	defer cleanupTestData()
	userID := setupTestUser()
	
	// Create two non-default addresses
	payload1 := types.CreateAddressPayload{
		Title:         "Home",
		FirstName:     "John",
		LastName:      "Doe",
		AddressLine1:  "123 Main St",
		City:          "New York",
		StateProvince: "NY",
		PostalCode:    "10001",
		Country:       "USA",
		IsDefault:     false,
	}
	
	payload2 := types.CreateAddressPayload{
		Title:         "Work",
		FirstName:     "John",
		LastName:      "Doe",
		AddressLine1:  "456 Office Blvd",
		City:          "New York",
		StateProvince: "NY",
		PostalCode:    "10002",
		Country:       "USA",
		IsDefault:     false,
	}
	
	address1, err := addressStore.CreateAddress(userID, payload1)
	if err != nil {
		t.Fatalf("Failed to create first address: %v", err)
	}
	
	address2, err := addressStore.CreateAddress(userID, payload2)
	if err != nil {
		t.Fatalf("Failed to create second address: %v", err)
	}
	
	// Set second address as default
	err = addressStore.SetDefaultAddress(address2.ID, userID)
	if err != nil {
		t.Fatalf("Failed to set default address: %v", err)
	}
	
	// Verify default address
	defaultAddress, err := addressStore.GetDefaultAddress(userID)
	if err != nil {
		t.Fatalf("Failed to get default address: %v", err)
	}
	
	if defaultAddress.ID != address2.ID {
		t.Errorf("Expected default address ID %d, got %d", address2.ID, defaultAddress.ID)
	}
	
	// Set first address as default
	err = addressStore.SetDefaultAddress(address1.ID, userID)
	if err != nil {
		t.Fatalf("Failed to set new default address: %v", err)
	}
	
	// Verify new default address
	defaultAddress, err = addressStore.GetDefaultAddress(userID)
	if err != nil {
		t.Fatalf("Failed to get new default address: %v", err)
	}
	
	if defaultAddress.ID != address1.ID {
		t.Errorf("Expected default address ID %d, got %d", address1.ID, defaultAddress.ID)
	}
}

func TestAddressStore_GetDefaultAddress_NoDefault(t *testing.T) {
	defer cleanupTestData()
	userID := setupTestUser()
	
	// Create non-default address
	payload := types.CreateAddressPayload{
		Title:         "Home",
		FirstName:     "John",
		LastName:      "Doe",
		AddressLine1:  "123 Main St",
		City:          "New York",
		StateProvince: "NY",
		PostalCode:    "10001",
		Country:       "USA",
		IsDefault:     false,
	}
	
	_, err := addressStore.CreateAddress(userID, payload)
	if err != nil {
		t.Fatalf("Failed to create address: %v", err)
	}
	
	// Try to get default address when none exists
	_, err = addressStore.GetDefaultAddress(userID)
	if err == nil {
		t.Error("Expected error when no default address exists")
	}
	
	if err.Error() != "address not found" {
		t.Errorf("Expected 'address not found' error message, got: %v", err)
	}
}