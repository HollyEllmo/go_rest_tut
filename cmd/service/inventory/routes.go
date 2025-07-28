package inventory

import (
	"net/http"
	"strconv"

	"github.com/HollyEllmo/go_rest_tut/cmd/service/auth"
	"github.com/HollyEllmo/go_rest_tut/cmd/types"
	"github.com/HollyEllmo/go_rest_tut/cmd/utils"
	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
)

type Handler struct {
	store     types.InventoryStore
	userStore types.UserStore
}

func NewHandler(store types.InventoryStore, userStore types.UserStore) *Handler {
	return &Handler{store: store, userStore: userStore}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	// Admin only routes for inventory management
	router.HandleFunc("/inventory/{productId}/stock", auth.WithJWTAuth(h.handleGetStock, h.userStore)).Methods(http.MethodGet)
	router.HandleFunc("/inventory/{productId}/history", auth.WithJWTAuth(h.handleGetHistory, h.userStore)).Methods(http.MethodGet)
	router.HandleFunc("/inventory/{productId}/add", auth.WithJWTAuth(h.handleAddStock, h.userStore)).Methods(http.MethodPost)
}

func (h *Handler) handleGetStock(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	productID, err := strconv.Atoi(vars["productId"])
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	stock, err := h.store.GetCurrentStock(productID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]any{
		"product_id":    productID,
		"current_stock": stock,
	})
}

func (h *Handler) handleGetHistory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	productID, err := strconv.Atoi(vars["productId"])
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	limit := 50 // default limit
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	history, err := h.store.GetStockHistory(productID, limit)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]any{
		"product_id": productID,
		"history":    history,
	})
}

type AddStockPayload struct {
	Quantity int    `json:"quantity" validate:"required,gt=0"`
	Reason   string `json:"reason" validate:"required"`
}

func (h *Handler) handleAddStock(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	productID, err := strconv.Atoi(vars["productId"])
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	var payload AddStockPayload
	if err := utils.ParseJSON(r, &payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	if err := utils.Validate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		utils.WriteError(w, http.StatusBadRequest, errors)
		return
	}

	err = h.store.AddStock(productID, payload.Quantity, payload.Reason, types.RefTypeRestock, nil)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	newStock, err := h.store.GetCurrentStock(productID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]any{
		"product_id":    productID,
		"added":         payload.Quantity,
		"current_stock": newStock,
		"reason":        payload.Reason,
	})
}