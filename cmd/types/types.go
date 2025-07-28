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
	Items     []CartItem `json:"items" validate:"required"`
	AddressID *int       `json:"addressId,omitempty"` // Optional: use specific address, if nil use default
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

// User Address types
type UserAddress struct {
	ID            int       `json:"id"`
	UserID        int       `json:"userId"`
	Title         string    `json:"title"`
	FirstName     string    `json:"firstName"`
	LastName      string    `json:"lastName"`
	Company       *string   `json:"company,omitempty"`
	AddressLine1  string    `json:"addressLine1"`
	AddressLine2  *string   `json:"addressLine2,omitempty"`
	City          string    `json:"city"`
	StateProvince string    `json:"stateProvince"`
	PostalCode    string    `json:"postalCode"`
	Country       string    `json:"country"`
	Phone         *string   `json:"phone,omitempty"`
	IsDefault     bool      `json:"isDefault"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

type CreateAddressPayload struct {
	Title         string  `json:"title" validate:"required,max=100"`
	FirstName     string  `json:"firstName" validate:"required,max=100"`
	LastName      string  `json:"lastName" validate:"required,max=100"`
	Company       *string `json:"company,omitempty" validate:"omitempty,max=100"`
	AddressLine1  string  `json:"addressLine1" validate:"required,max=255"`
	AddressLine2  *string `json:"addressLine2,omitempty" validate:"omitempty,max=255"`
	City          string  `json:"city" validate:"required,max=100"`
	StateProvince string  `json:"stateProvince" validate:"required,max=100"`
	PostalCode    string  `json:"postalCode" validate:"required,max=20"`
	Country       string  `json:"country" validate:"required,max=100"`
	Phone         *string `json:"phone,omitempty" validate:"omitempty,max=20"`
	IsDefault     bool    `json:"isDefault"`
}

type UpdateAddressPayload struct {
	Title         *string `json:"title,omitempty" validate:"omitempty,max=100"`
	FirstName     *string `json:"firstName,omitempty" validate:"omitempty,max=100"`
	LastName      *string `json:"lastName,omitempty" validate:"omitempty,max=100"`
	Company       *string `json:"company,omitempty" validate:"omitempty,max=100"`
	AddressLine1  *string `json:"addressLine1,omitempty" validate:"omitempty,max=255"`
	AddressLine2  *string `json:"addressLine2,omitempty" validate:"omitempty,max=255"`
	City          *string `json:"city,omitempty" validate:"omitempty,max=100"`
	StateProvince *string `json:"stateProvince,omitempty" validate:"omitempty,max=100"`
	PostalCode    *string `json:"postalCode,omitempty" validate:"omitempty,max=20"`
	Country       *string `json:"country,omitempty" validate:"omitempty,max=100"`
	Phone         *string `json:"phone,omitempty" validate:"omitempty,max=20"`
	IsDefault     *bool   `json:"isDefault,omitempty"`
}

// Address Store interface
type AddressStore interface {
	GetUserAddresses(userID int) ([]UserAddress, error)
	GetAddressByID(addressID, userID int) (*UserAddress, error)
	CreateAddress(userID int, payload CreateAddressPayload) (*UserAddress, error)
	UpdateAddress(addressID, userID int, payload UpdateAddressPayload) (*UserAddress, error)
	DeleteAddress(addressID, userID int) error
	GetDefaultAddress(userID int) (*UserAddress, error)
	SetDefaultAddress(addressID, userID int) error
}
