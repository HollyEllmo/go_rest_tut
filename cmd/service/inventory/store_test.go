package inventory

import (
	"database/sql"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/HollyEllmo/go_rest_tut/cmd/config"
	"github.com/HollyEllmo/go_rest_tut/cmd/db"
	"github.com/HollyEllmo/go_rest_tut/cmd/types"
	mysqlCfg "github.com/go-sql-driver/mysql"
)

// Test database setup
func setupTestDB(t *testing.T) *sql.DB {
	// Load config
	config.Envs = config.Config{
		DBUser:     "root",
		DBPassword: "password",
		DBAddress:  "127.0.0.1:3306",
		DBName:     "ecom_test",
	}

	// Create test database connection
	testDB, err := db.NewMySQLStorage(mysqlCfg.Config{
		User:                 config.Envs.DBUser,
		Passwd:               config.Envs.DBPassword,
		Addr:                 config.Envs.DBAddress,
		DBName:               config.Envs.DBName,
		Net:                  "tcp",
		AllowNativePasswords: true,
		ParseTime:            true,
	})

	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Clean up any existing test data
	tables := []string{"inventory_movements", "order_items", "orders", "products", "users"}
	for _, table := range tables {
		_, err := testDB.Exec(fmt.Sprintf("DELETE FROM %s", table))
		if err != nil {
			t.Logf("Warning: Failed to clean table %s: %v", table, err)
		}
	}

	return testDB
}

// Helper function to setup test data
func setupTestData(t *testing.T, db *sql.DB) (int, int) {
	// Create a test product with unique name
	productName := fmt.Sprintf("Test Product %d", time.Now().UnixNano())
	result, err := db.Exec("INSERT INTO products (name, description, image, price) VALUES (?, ?, ?, ?)",
		productName, "Test Description", "test.jpg", 99.99)
	if err != nil {
		t.Fatalf("Failed to create test product: %v", err)
	}

	productID, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("Failed to get product ID: %v", err)
	}

	// Create a test user with unique email
	userEmail := fmt.Sprintf("test%d@example.com", time.Now().UnixNano())
	result, err = db.Exec("INSERT INTO users (firstName, lastName, email, password) VALUES (?, ?, ?, ?)",
		"Test", "User", userEmail, "hashedpassword")
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	userID, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("Failed to get user ID: %v", err)
	}

	return int(productID), int(userID)
}

// Helper function to add initial stock
func addInitialStock(t *testing.T, store *Store, productID, quantity int) {
	err := store.AddStock(productID, quantity, "Initial test stock", types.RefTypeRestock, nil)
	if err != nil {
		t.Fatalf("Failed to add initial stock: %v", err)
	}
}

// Helper function to create an order
func createTestOrder(t *testing.T, db *sql.DB, userID int, total float64) int {
	result, err := db.Exec("INSERT INTO orders (userID, total, status, address) VALUES (?, ?, ?, ?)",
		userID, total, "pending", "Test Address")
	if err != nil {
		t.Fatalf("Failed to create test order: %v", err)
	}

	orderID, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("Failed to get order ID: %v", err)
	}

	return int(orderID)
}

// Test concurrent stock reservation - this is the main test for race conditions
func TestConcurrentStockReservation(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	store := NewStore(db)
	productID, userID := setupTestData(t, db)

	// Add initial stock of 10 items
	initialStock := 10
	addInitialStock(t, store, productID, initialStock)

	// Verify initial stock
	stock, err := store.GetCurrentStock(productID)
	if err != nil {
		t.Fatalf("Failed to get initial stock: %v", err)
	}
	if stock != initialStock {
		t.Fatalf("Expected initial stock %d, got %d", initialStock, stock)
	}

	// Number of concurrent goroutines trying to reserve stock
	numGoroutines := 20
	itemsPerReservation := 1
	expectedSuccessfulReservations := initialStock / itemsPerReservation

	var wg sync.WaitGroup
	var mu sync.Mutex
	successfulReservations := 0
	failedReservations := 0
	results := make([]error, numGoroutines)

	// Launch concurrent goroutines trying to reserve stock
	for i := range numGoroutines {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()

			// Create order for this goroutine
			orderID := createTestOrder(t, db, userID, 99.99)

			// Try to reserve stock
			err := store.ReserveStock(productID, itemsPerReservation, orderID)
			results[goroutineID] = err

			// Track results thread-safely
			mu.Lock()
			if err != nil {
				failedReservations++
			} else {
				successfulReservations++
			}
			mu.Unlock()

			// Add some randomness to timing
			time.Sleep(time.Duration(goroutineID%5) * time.Millisecond)
		}(i)
	}

	// Wait for all goroutines to complete
	wg.Wait()

	// Verify results
	t.Logf("Successful reservations: %d", successfulReservations)
	t.Logf("Failed reservations: %d", failedReservations)
	t.Logf("Expected successful: %d", expectedSuccessfulReservations)

	// Check that we got exactly the expected number of successful reservations
	if successfulReservations != expectedSuccessfulReservations {
		t.Errorf("Expected %d successful reservations, got %d", expectedSuccessfulReservations, successfulReservations)
	}

	// Check that the remaining failed as expected
	expectedFailedReservations := numGoroutines - expectedSuccessfulReservations
	if failedReservations != expectedFailedReservations {
		t.Errorf("Expected %d failed reservations, got %d", expectedFailedReservations, failedReservations)
	}

	// Verify final stock is exactly 0
	finalStock, err := store.GetCurrentStock(productID)
	if err != nil {
		t.Fatalf("Failed to get final stock: %v", err)
	}
	if finalStock != 0 {
		t.Errorf("Expected final stock to be 0, got %d", finalStock)
	}

	// Verify that successful reservations have proper error messages
	insufficientStockErrors := 0
	for i, result := range results {
		if result != nil {
			t.Logf("Goroutine %d failed with: %v", i, result)
			if fmt.Sprintf("%v", result) == fmt.Sprintf("insufficient stock for product %d: available 0, requested %d", productID, itemsPerReservation) ||
				fmt.Sprintf("%v", result) == fmt.Sprintf("insufficient stock for product %d: available 1, requested %d", productID, itemsPerReservation) ||
				fmt.Sprintf("%v", result) == fmt.Sprintf("insufficient stock for product %d: available 2, requested %d", productID, itemsPerReservation) ||
				fmt.Sprintf("%v", result) == fmt.Sprintf("insufficient stock for product %d: available 3, requested %d", productID, itemsPerReservation) ||
				fmt.Sprintf("%v", result) == fmt.Sprintf("insufficient stock for product %d: available 4, requested %d", productID, itemsPerReservation) ||
				fmt.Sprintf("%v", result) == fmt.Sprintf("insufficient stock for product %d: available 5, requested %d", productID, itemsPerReservation) ||
				fmt.Sprintf("%v", result) == fmt.Sprintf("insufficient stock for product %d: available 6, requested %d", productID, itemsPerReservation) ||
				fmt.Sprintf("%v", result) == fmt.Sprintf("insufficient stock for product %d: available 7, requested %d", productID, itemsPerReservation) ||
				fmt.Sprintf("%v", result) == fmt.Sprintf("insufficient stock for product %d: available 8, requested %d", productID, itemsPerReservation) ||
				fmt.Sprintf("%v", result) == fmt.Sprintf("insufficient stock for product %d: available 9, requested %d", productID, itemsPerReservation) {
				insufficientStockErrors++
			}
		}
	}

	t.Logf("Insufficient stock errors: %d", insufficientStockErrors)
}

// Test overselling prevention - trying to buy more than available
func TestOversellePrevention(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	store := NewStore(db)
	productID, userID := setupTestData(t, db)

	// Add initial stock of 5 items
	initialStock := 5
	addInitialStock(t, store, productID, initialStock)

	// Try to reserve 6 items (more than available)
	orderID := createTestOrder(t, db, userID, 599.94)
	err := store.ReserveStock(productID, 6, orderID)

	// Should fail with insufficient stock error
	if err == nil {
		t.Fatal("Expected error when trying to reserve more stock than available")
	}

	expectedError := fmt.Sprintf("insufficient stock for product %d: available 5, requested 6", productID)
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%v'", expectedError, err)
	}

	// Verify stock wasn't changed
	finalStock, err := store.GetCurrentStock(productID)
	if err != nil {
		t.Fatalf("Failed to get final stock: %v", err)
	}
	if finalStock != initialStock {
		t.Errorf("Expected stock to remain %d, got %d", initialStock, finalStock)
	}
}

// Test high-load concurrent scenario with multiple products
func TestHighLoadConcurrentReservations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping high-load test in short mode")
	}

	db := setupTestDB(t)
	defer db.Close()

	store := NewStore(db)
	_, userID := setupTestData(t, db)

	// Create 3 products with different stock levels
	products := []struct {
		name  string
		stock int
	}{
		{"High Stock Product", 100},
		{"Medium Stock Product", 50},
		{"Low Stock Product", 10},
	}

	productIDs := make([]int, len(products))
	for i, product := range products {
		result, err := db.Exec("INSERT INTO products (name, description, image, price) VALUES (?, ?, ?, ?)",
			product.name, "Test Description", "test.jpg", 99.99)
		if err != nil {
			t.Fatalf("Failed to create product: %v", err)
		}

		productID, err := result.LastInsertId()
		if err != nil {
			t.Fatalf("Failed to get product ID: %v", err)
		}

		productIDs[i] = int(productID)
		addInitialStock(t, store, int(productID), product.stock)
	}

	// Launch 200 concurrent goroutines
	numGoroutines := 200
	var wg sync.WaitGroup
	var mu sync.Mutex
	totalSuccessful := 0
	totalFailed := 0

	for i := range numGoroutines {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()

			// Each goroutine randomly picks a product and tries to reserve 1-3 items
			productIndex := goroutineID % len(productIDs)
			productID := productIDs[productIndex]
			quantity := (goroutineID % 3) + 1 // 1, 2, or 3 items

			orderID := createTestOrder(t, db, userID, float64(quantity)*99.99)
			err := store.ReserveStock(productID, quantity, orderID)

			mu.Lock()
			if err != nil {
				totalFailed++
			} else {
				totalSuccessful++
			}
			mu.Unlock()
		}(i)
	}

	wg.Wait()

	t.Logf("High-load test results:")
	t.Logf("Total successful reservations: %d", totalSuccessful)
	t.Logf("Total failed reservations: %d", totalFailed)
	t.Logf("Total goroutines: %d", numGoroutines)

	// Verify that no stock went negative
	for i, productID := range productIDs {
		stock, err := store.GetCurrentStock(productID)
		if err != nil {
			t.Fatalf("Failed to get stock for product %d: %v", productID, err)
		}
		if stock < 0 {
			t.Errorf("Product %d (%s) has negative stock: %d", productID, products[i].name, stock)
		}
		t.Logf("Product %d final stock: %d (started with %d)", productID, stock, products[i].stock)
	}
}