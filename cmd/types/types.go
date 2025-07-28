package types

import "time"

type UserStore interface {
	GetUserByEmail(email string) (*User, error)
	GetUserByID(id int) (*User, error)
	CreateUser(user User) error
}

type ProductStore interface {
	GetProducts() ([]Product, error)
	GetProductsByIDs(ps []int) ([]Product, error)
	CreateProduct(product *Product) error
	UpdateProduct(*Product) error
}

type OrderStore interface {
	CreateOrder(Order) (int, error)
	CreateOrderItem(OrderItem) error
}

type Order struct {
	ID        int       `json:"id"`
	UserID    int       `json:"userId"`
	Total     float64   `json:"total"`
	Status    string    `json:"status"`
	Address   string    `json:"address"`
	CreatedAt time.Time `json:"createdAt"`
}

type OrderItem struct {
	ID        int     `json:"id"`
	OrderID   int     `json:"orderId"`
	ProductID int     `json:"productId"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price"`
}

type Product struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Image       string    `json:"image"`
	Price       float64   `json:"price"`
	CreatedAt   time.Time `json:"createdAt"`
}

type CreateProductPayload struct {
	Name        string  `json:"name" validate:"required"`
	Description string  `json:"description" validate:"required"`
	Image       string  `json:"image" validate:"required"`
	Price       float64 `json:"price" validate:"required,gt=0"`
}

type User struct {
	ID        int       `json:"id"`
	FirstName string    `json:"firstName"`
	LastName  string    `json:"lastName"`
	Email     string    `json:"email"`
	Password  string    `json:"password"`
	CreatedAt time.Time `json:"createdAt"`
}

type RegisterUserPayload struct {
	FirstName string `json:"firstName" validate:"required"`
	LastName  string `json:"lastName" validate:"required"`
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required,min=3,max=100"`
}

type LoginUserPayload struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type CartItem struct {
	ProductID int `json:"productId"`
	Quantity  int `json:"quantity"`
}

type CartCheckoutPayload struct {
	Items []CartItem `json:"items" validate:"required"`
}

// Inventory Movement types
type InventoryMovement struct {
	ID            int                   `json:"id"`
	ProductID     int                   `json:"productId"`
	MovementType  InventoryMovementType `json:"movementType"`
	Quantity      int                   `json:"quantity"`
	Reason        string                `json:"reason"`
	ReferenceID   *int                  `json:"referenceId,omitempty"`
	ReferenceType *InventoryRefType     `json:"referenceType,omitempty"`
	CreatedAt     time.Time             `json:"createdAt"`
}

type InventoryMovementType string

const (
	MovementTypeIn  InventoryMovementType = "IN"
	MovementTypeOut InventoryMovementType = "OUT"
)

type InventoryRefType string

const (
	RefTypeOrder      InventoryRefType = "ORDER"
	RefTypeRestock    InventoryRefType = "RESTOCK"
	RefTypeAdjustment InventoryRefType = "ADJUSTMENT"
	RefTypeReturn     InventoryRefType = "RETURN"
)

// Inventory Store interface
type InventoryStore interface {
	GetCurrentStock(productID int) (int, error)
	GetProductsWithStock(productIDs []int) (map[int]int, error)
	ReserveStock(productID, quantity int, orderID int) error
	ReleaseStock(productID, quantity int, reason string) error
	AddStock(productID, quantity int, reason string, refType InventoryRefType, refID *int) error
	GetStockHistory(productID int, limit int) ([]InventoryMovement, error)
}
