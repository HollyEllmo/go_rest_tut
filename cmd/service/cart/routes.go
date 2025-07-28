package cart

import (
	"net/http"

	"github.com/HollyEllmo/go_rest_tut/cmd/service/auth"
	"github.com/HollyEllmo/go_rest_tut/cmd/types"
	"github.com/HollyEllmo/go_rest_tut/cmd/utils"
	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
)

type Handler struct {
	store          types.OrderStore
	productStore   types.ProductStore
	userStore      types.UserStore
	inventoryStore types.InventoryStore
}

func NewHandler(store types.OrderStore, productStore types.ProductStore, userStore types.UserStore, inventoryStore types.InventoryStore) *Handler {
	return &Handler{store: store, productStore: productStore, userStore: userStore, inventoryStore: inventoryStore}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/cart/checkout", auth.WithJWTAuth(h.handleCheckout, h.userStore)).Methods(http.MethodPost)
}

func (h *Handler) handleCheckout(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserIDFromContext(r.Context())

	var cart types.CartCheckoutPayload
	if err := utils.ParseJSON(r, &cart); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	if err := utils.Validate.Struct(cart); err != nil {
		errors := err.(validator.ValidationErrors)
		utils.WriteError(w, http.StatusBadRequest, errors)
		return
	}

	productIDs, err := getCartItemsIDs(cart.Items)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	// get products from the store
	ps, err := h.productStore.GetProductsByIDs(productIDs)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	orderID, totalPrice, err := h.createOrder(ps, cart.Items, userID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusCreated, map[string]any{
		"total_price": totalPrice,
		"order_id":    orderID,
	})
}
