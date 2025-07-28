package order

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/HollyEllmo/go_rest_tut/cmd/config"
	"github.com/HollyEllmo/go_rest_tut/cmd/db"
	"github.com/HollyEllmo/go_rest_tut/cmd/types"
	"github.com/go-sql-driver/mysql"
	_ "github.com/go-sql-driver/mysql"
)

var testDB *sql.DB
var orderStore *Store

func TestMain(m *testing.M) {
	cfg := config.Envs
	
	// Connect to test database
	testDBName := "go_rest_tut_order_test"
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
	
	orderStore = NewStore(testDB)
	
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
	// Create users table
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
	
	// Create products table
	productsTableSQL := `
		CREATE TABLE IF NOT EXISTS products (
			id INT AUTO_INCREMENT PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			description TEXT,
			image VARCHAR(255) NOT NULL,
			price DECIMAL(10,2) NOT NULL,
			createdAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`
	
	// Create orders table
	ordersTableSQL := `
		CREATE TABLE IF NOT EXISTS orders (
			id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
			userId INT UNSIGNED NOT NULL,
			total DECIMAL(10,2) NOT NULL,
			status ENUM('pending','completed','cancelled') NOT NULL DEFAULT 'pending',
			address TEXT NOT NULL,
			createdAt TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			
			KEY idx_orders_user_id (userId),
			KEY idx_orders_status (status),
			KEY idx_orders_created_at (createdAt)
		)
	`
	
	// Create order_items table
	orderItemsTableSQL := `
		CREATE TABLE IF NOT EXISTS order_items (
			id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
			orderId INT UNSIGNED NOT NULL,
			productId INT UNSIGNED NOT NULL,
			quantity INT NOT NULL,
			price DECIMAL(10,2) NOT NULL,
			
			KEY idx_order_items_order_id (orderId),
			KEY idx_order_items_product_id (productId)
		)
	`
	
	tables := []string{usersTableSQL, productsTableSQL, ordersTableSQL, orderItemsTableSQL}
	
	for _, tableSQL := range tables {
		if _, err := testDB.Exec(tableSQL); err != nil {
			log.Fatalf("Failed to create table: %v", err)
		}
	}
}

func cleanupTestDB() {
	testDB.Exec("DROP TABLE IF EXISTS order_items")
	testDB.Exec("DROP TABLE IF EXISTS orders")
	testDB.Exec("DROP TABLE IF EXISTS products")
	testDB.Exec("DROP TABLE IF EXISTS users")
}

func setupTestData() (int, int, int) {
	// Create test user
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
	
	// Create test product
	_, err = testDB.Exec(`
		INSERT INTO products (name, description, image, price) 
		VALUES ('Test Product', 'A test product', 'test.jpg', 99.99)
	`)
	if err != nil {
		log.Fatalf("Failed to create test product: %v", err)
	}
	
	var productID int
	err = testDB.QueryRow("SELECT id FROM products WHERE name = 'Test Product'").Scan(&productID)
	if err != nil {
		log.Fatalf("Failed to get test product ID: %v", err)
	}
	
	// Create test order
	order := types.Order{
		UserID:  userID,
		Total:   199.98,
		Status:  "completed",
		Address: "123 Test St, Test City, TC 12345",
	}
	
	orderID, err := orderStore.CreateOrder(order)
	if err != nil {
		log.Fatalf("Failed to create test order: %v", err)
	}
	
	// Create test order items
	orderItem := types.OrderItem{
		OrderID:   orderID,
		ProductID: productID,
		Quantity:  2,
		Price:     99.99,
	}
	
	err = orderStore.CreateOrderItem(orderItem)
	if err != nil {
		log.Fatalf("Failed to create test order item: %v", err)
	}
	
	return userID, productID, orderID
}

func cleanupTestData() {
	testDB.Exec("DELETE FROM order_items")
	testDB.Exec("DELETE FROM orders")
	testDB.Exec("DELETE FROM products")
	testDB.Exec("DELETE FROM users")
}

func TestOrderStore_GetUserOrders(t *testing.T) {
	defer cleanupTestData()
	userID, _, _ := setupTestData()
	
	// Test getting all orders
	filters := types.OrderFilters{
		Limit:  10,
		Offset: 0,
	}
	
	orders, err := orderStore.GetUserOrders(userID, filters)
	if err != nil {
		t.Fatalf("Failed to get user orders: %v", err)
	}
	
	if len(orders) != 1 {
		t.Errorf("Expected 1 order, got %d", len(orders))
	}
	
	order := orders[0]
	if order.UserID != userID {
		t.Errorf("Expected user ID %d, got %d", userID, order.UserID)
	}
	
	if order.Status != "completed" {
		t.Errorf("Expected status 'completed', got '%s'", order.Status)
	}
	
	if len(order.Items) != 1 {
		t.Errorf("Expected 1 order item, got %d", len(order.Items))
	}
	
	item := order.Items[0]
	if item.ProductName != "Test Product" {
		t.Errorf("Expected product name 'Test Product', got '%s'", item.ProductName)
	}
	
	if item.Quantity != 2 {
		t.Errorf("Expected quantity 2, got %d", item.Quantity)
	}
}

func TestOrderStore_GetUserOrdersWithStatusFilter(t *testing.T) {
	defer cleanupTestData()
	userID, productID, _ := setupTestData()
	
	// Create a pending order
	pendingOrder := types.Order{
		UserID:  userID,
		Total:   49.99,
		Status:  "pending",
		Address: "456 Another St",
	}
	
	pendingOrderID, err := orderStore.CreateOrder(pendingOrder)
	if err != nil {
		t.Fatalf("Failed to create pending order: %v", err)
	}
	
	// Add item to pending order
	err = orderStore.CreateOrderItem(types.OrderItem{
		OrderID:   pendingOrderID,
		ProductID: productID,
		Quantity:  1,
		Price:     49.99,
	})
	if err != nil {
		t.Fatalf("Failed to create pending order item: %v", err)
	}
	
	// Test filtering by completed status
	completedStatus := "completed"
	filters := types.OrderFilters{
		Status: &completedStatus,
		Limit:  10,
		Offset: 0,
	}
	
	orders, err := orderStore.GetUserOrders(userID, filters)
	if err != nil {
		t.Fatalf("Failed to get completed orders: %v", err)
	}
	
	if len(orders) != 1 {
		t.Errorf("Expected 1 completed order, got %d", len(orders))
	}
	
	if orders[0].Status != "completed" {
		t.Errorf("Expected completed status, got '%s'", orders[0].Status)
	}
	
	// Test filtering by pending status
	pendingStatus := "pending"
	filters.Status = &pendingStatus
	
	orders, err = orderStore.GetUserOrders(userID, filters)
	if err != nil {
		t.Fatalf("Failed to get pending orders: %v", err)
	}
	
	if len(orders) != 1 {
		t.Errorf("Expected 1 pending order, got %d", len(orders))
	}
	
	if orders[0].Status != "pending" {
		t.Errorf("Expected pending status, got '%s'", orders[0].Status)
	}
}

func TestOrderStore_GetUserOrdersWithPagination(t *testing.T) {
	defer cleanupTestData()
	userID, productID, _ := setupTestData()
	
	// Create multiple orders
	for i := 0; i < 3; i++ {
		order := types.Order{
			UserID:  userID,
			Total:   float64(100 + i*10),
			Status:  "completed",
			Address: fmt.Sprintf("Address %d", i),
		}
		
		orderID, err := orderStore.CreateOrder(order)
		if err != nil {
			t.Fatalf("Failed to create order %d: %v", i, err)
		}
		
		err = orderStore.CreateOrderItem(types.OrderItem{
			OrderID:   orderID,
			ProductID: productID,
			Quantity:  1,
			Price:     float64(100 + i*10),
		})
		if err != nil {
			t.Fatalf("Failed to create order item %d: %v", i, err)
		}
		
		// Add slight delay to ensure different timestamps
		time.Sleep(10 * time.Millisecond)
	}
	
	// Test pagination: limit 2, offset 0
	filters := types.OrderFilters{
		Limit:  2,
		Offset: 0,
	}
	
	orders, err := orderStore.GetUserOrders(userID, filters)
	if err != nil {
		t.Fatalf("Failed to get orders with pagination: %v", err)
	}
	
	if len(orders) != 2 {
		t.Errorf("Expected 2 orders in first page, got %d", len(orders))
	}
	
	// Test second page: limit 2, offset 2
	filters.Offset = 2
	
	orders, err = orderStore.GetUserOrders(userID, filters)
	if err != nil {
		t.Fatalf("Failed to get second page: %v", err)
	}
	
	if len(orders) != 2 {
		t.Errorf("Expected 2 orders in second page, got %d", len(orders))
	}
}

func TestOrderStore_GetOrderByID(t *testing.T) {
	defer cleanupTestData()
	userID, _, orderID := setupTestData()
	
	// Test getting existing order
	order, err := orderStore.GetOrderByID(orderID, userID)
	if err != nil {
		t.Fatalf("Failed to get order by ID: %v", err)
	}
	
	if order.ID != orderID {
		t.Errorf("Expected order ID %d, got %d", orderID, order.ID)
	}
	
	if order.UserID != userID {
		t.Errorf("Expected user ID %d, got %d", userID, order.UserID)
	}
	
	if len(order.Items) != 1 {
		t.Errorf("Expected 1 order item, got %d", len(order.Items))
	}
	
	// Test getting non-existent order
	_, err = orderStore.GetOrderByID(9999, userID)
	if err == nil {
		t.Error("Expected error when getting non-existent order")
	}
	
	if err.Error() != "order not found or not owned by user" {
		t.Errorf("Expected specific error message, got: %v", err)
	}
	
	// Test getting order from different user
	_, err = orderStore.GetOrderByID(orderID, userID+1)
	if err == nil {
		t.Error("Expected error when getting order from different user")
	}
}

func TestOrderStore_GetOrdersCount(t *testing.T) {
	defer cleanupTestData()
	userID, productID, _ := setupTestData()
	
	// Create additional orders with different statuses
	pendingOrder := types.Order{
		UserID:  userID,
		Total:   49.99,
		Status:  "pending",
		Address: "Pending Address",
	}
	
	pendingOrderID, err := orderStore.CreateOrder(pendingOrder)
	if err != nil {
		t.Fatalf("Failed to create pending order: %v", err)
	}
	
	err = orderStore.CreateOrderItem(types.OrderItem{
		OrderID:   pendingOrderID,
		ProductID: productID,
		Quantity:  1,
		Price:     49.99,
	})
	if err != nil {
		t.Fatalf("Failed to create pending order item: %v", err)
	}
	
	// Test count without filters
	filters := types.OrderFilters{}
	count, err := orderStore.GetOrdersCount(userID, filters)
	if err != nil {
		t.Fatalf("Failed to get orders count: %v", err)
	}
	
	if count != 2 {
		t.Errorf("Expected count 2, got %d", count)
	}
	
	// Test count with status filter
	completedStatus := "completed"
	filters.Status = &completedStatus
	
	count, err = orderStore.GetOrdersCount(userID, filters)
	if err != nil {
		t.Fatalf("Failed to get completed orders count: %v", err)
	}
	
	if count != 1 {
		t.Errorf("Expected completed count 1, got %d", count)
	}
}

func TestOrderStore_GetUserOrdersWithDateFilter(t *testing.T) {
	defer cleanupTestData()
	userID, productID, _ := setupTestData()
	
	// Create an order from yesterday
	yesterday := time.Now().AddDate(0, 0, -1)
	
	// We need to manually insert with specific date since CreateOrder uses CURRENT_TIMESTAMP
	_, err := testDB.Exec(`
		INSERT INTO orders (userId, total, status, address, createdAt) 
		VALUES (?, ?, ?, ?, ?)
	`, userID, 75.50, "completed", "Yesterday Address", yesterday)
	if err != nil {
		t.Fatalf("Failed to create yesterday order: %v", err)
	}
	
	var yesterdayOrderID int
	err = testDB.QueryRow("SELECT id FROM orders WHERE address = 'Yesterday Address'").Scan(&yesterdayOrderID)
	if err != nil {
		t.Fatalf("Failed to get yesterday order ID: %v", err)
	}
	
	err = orderStore.CreateOrderItem(types.OrderItem{
		OrderID:   yesterdayOrderID,
		ProductID: productID,
		Quantity:  1,
		Price:     75.50,
	})
	if err != nil {
		t.Fatalf("Failed to create yesterday order item: %v", err)
	}
	
	// Test filtering from today
	today := time.Now().Truncate(24 * time.Hour)
	filters := types.OrderFilters{
		FromDate: &today,
		Limit:    10,
		Offset:   0,
	}
	
	orders, err := orderStore.GetUserOrders(userID, filters)
	if err != nil {
		t.Fatalf("Failed to get today's orders: %v", err)
	}
	
	if len(orders) != 1 {
		t.Errorf("Expected 1 order from today, got %d", len(orders))
	}
	
	// Test filtering up to yesterday (end of day)
	filters.FromDate = nil
	yesterdayEnd := yesterday.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
	filters.ToDate = &yesterdayEnd
	
	orders, err = orderStore.GetUserOrders(userID, filters)
	if err != nil {
		t.Fatalf("Failed to get yesterday's orders: %v", err)
	}
	
	if len(orders) != 1 {
		t.Errorf("Expected 1 order until yesterday, got %d", len(orders))
	}
}