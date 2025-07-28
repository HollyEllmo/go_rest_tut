package cart

import (
	"fmt"

	"github.com/HollyEllmo/go_rest_tut/cmd/types"
)

func getCartItemsIDs(items []types.CartItem) ([]int, error) {
	productIDs := make([]int, 0, len(items))
	for i, item := range items {
		if item.Quantity <= 0 {
			return nil, fmt.Errorf("invalid quantity for item %d: %d", i, item.Quantity)
		}
		productIDs = append(productIDs, item.ProductID)
	}
	return productIDs, nil
}

func (h *Handler) createOrder(ps []types.Product, items []types.CartItem, userID int) (int, float64, error) {
	productMap := make(map[int]types.Product)
	for _, product := range ps {
		productMap[product.ID] = product
	}

	// check if all products are actually in stock
	if err := h.checkIfCartIsInStock(items, productMap); err != nil {
		return 0, 0, err
	}

	// calculate the total price
	totalPrice := calculateTotalPrice(items, productMap)

	// create order first
	orderID, err := h.store.CreateOrder(types.Order{
		UserID:  userID,
		Total:   totalPrice,
		Status:  "pending",
		Address: "default address", // This should be replaced with actual address handling
	})

	if err != nil {
		return 0, 0, fmt.Errorf("failed to create order: %w", err)
	}

	// atomically reserve stock for all items
	for _, item := range items {
		err := h.inventoryStore.ReserveStock(item.ProductID, item.Quantity, orderID)
		if err != nil {
			// If any reservation fails, we need to rollback previous reservations
			// and cancel the order (in a real system you'd want proper saga pattern)
			return 0, 0, fmt.Errorf("failed to reserve stock for product %d: %w", item.ProductID, err)
		}
	}

	// create order items
	for _, item := range items {
		h.store.CreateOrderItem(types.OrderItem{
			OrderID:   orderID,
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			Price:     productMap[item.ProductID].Price,
		})
	}

	return orderID, totalPrice, nil
}

func (h *Handler) checkIfCartIsInStock(cartItems []types.CartItem, productMap map[int]types.Product) error {
	if len(cartItems) == 0 {
		return fmt.Errorf("cart is empty")
	}

	// Get product IDs for stock check
	productIDs := make([]int, len(cartItems))
	for i, item := range cartItems {
		productIDs[i] = item.ProductID
	}

	// Get current stock levels from inventory
	stockMap, err := h.inventoryStore.GetProductsWithStock(productIDs)
	if err != nil {
		return fmt.Errorf("failed to check inventory: %w", err)
	}

	for _, item := range cartItems {
		product, ok := productMap[item.ProductID]
		if !ok {
			return fmt.Errorf("product with ID %d not found", item.ProductID)
		}
		
		availableStock := stockMap[item.ProductID]
		if availableStock < item.Quantity {
			return fmt.Errorf("not enough stock for product %s (ID: %d), requested: %d, available: %d",
				product.Name, product.ID, item.Quantity, availableStock)
		}
	}
	return nil
}

func calculateTotalPrice(cartItems []types.CartItem, products map[int]types.Product) float64 {
	var total float64

	for _, item := range cartItems {
		product := products[item.ProductID]
		total += product.Price * float64(item.Quantity)
	}
	return total
}
