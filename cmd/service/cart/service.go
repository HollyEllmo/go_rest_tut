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
		productIDs[i] = item.ProductID
	}
	return productIDs, nil
}

func (h *Handler) createOrder(ps []types.Product, items []types.CartItem, userID int) (int, float64, error) {
	productMap := make(map[int]types.Product)
	for _, product := range ps {
		productMap[product.ID] = product
	}

	// check if all products are actually in stock
	if err := checkIfCartIsInStock(items, productMap); err != nil {
		return 0, 0, err
	}

	// calculate the total price
	totalPrice := calculateTotalPrice(items, productMap)

	// reduce quantity of products in our db
	for _, item := range items {
		product := productMap[item.ProductID]
		product.Quantity -= item.Quantity

		h.productStore.UpdateProduct(product)
	}

	// create order
	orderID, err := h.store.CreateOrder(types.Order{
		UserID:  userID,
		Total:   totalPrice,
		Status:  "pending",
		Address: "default address", // This should be replaced with actual address handling
	})

	if err != nil {
		return 0, 0, fmt.Errorf("failed to create order: %w", err)
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

func checkIfCartIsInStock(cartItems []types.CartItem, productMap map[int]types.Product) error {
	if len(cartItems) == 0 {
		return fmt.Errorf("cart is empty")
	}

	for _, item := range cartItems {
		product, ok := productMap[item.ProductID]
		if !ok {
			return fmt.Errorf("product with ID %d not found", item.ProductID)
		}
		if product.Quantity < item.Quantity {
			return fmt.Errorf("not enough stock for product %s (ID: %d), requested: %d, available: %d",
				product.Name, product.ID, item.Quantity, product.Quantity)
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
