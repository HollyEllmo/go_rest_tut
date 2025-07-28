package order

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/HollyEllmo/go_rest_tut/cmd/service/auth"
	"github.com/HollyEllmo/go_rest_tut/cmd/types"
	"github.com/HollyEllmo/go_rest_tut/cmd/utils"
	"github.com/gorilla/mux"
)

type Handler struct {
	store     types.OrderStore
	userStore types.UserStore
}

func NewHandler(store types.OrderStore, userStore types.UserStore) *Handler {
	return &Handler{store: store, userStore: userStore}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	// All order routes require authentication
	router.HandleFunc("/orders", auth.WithJWTAuth(h.handleGetOrders, h.userStore)).Methods(http.MethodGet)
	router.HandleFunc("/orders/{id}", auth.WithJWTAuth(h.handleGetOrder, h.userStore)).Methods(http.MethodGet)
}

// GET /api/v1/orders - get orders for authenticated user with optional filters
func (h *Handler) handleGetOrders(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserIDFromContext(r.Context())
	
	// Parse query parameters
	filters := types.OrderFilters{
		Limit:  10, // Default limit
		Offset: 0,  // Default offset
	}
	
	// Parse status filter
	if status := r.URL.Query().Get("status"); status != "" {
		if status == "pending" || status == "completed" || status == "cancelled" {
			filters.Status = &status
		} else {
			utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid status. Must be one of: pending, completed, cancelled"))
			return
		}
	}
	
	// Parse limit
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit < 1 || limit > 100 {
			utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid limit. Must be between 1 and 100"))
			return
		}
		filters.Limit = limit
	}
	
	// Parse offset
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		offset, err := strconv.Atoi(offsetStr)
		if err != nil || offset < 0 {
			utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid offset. Must be >= 0"))
			return
		}
		filters.Offset = offset
	}
	
	// Parse date filters
	if fromDateStr := r.URL.Query().Get("fromDate"); fromDateStr != "" {
		fromDate, err := time.Parse("2006-01-02", fromDateStr)
		if err != nil {
			utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid fromDate format. Use YYYY-MM-DD"))
			return
		}
		filters.FromDate = &fromDate
	}
	
	if toDateStr := r.URL.Query().Get("toDate"); toDateStr != "" {
		toDate, err := time.Parse("2006-01-02", toDateStr)
		if err != nil {
			utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid toDate format. Use YYYY-MM-DD"))
			return
		}
		// Set to end of day
		toDate = toDate.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
		filters.ToDate = &toDate
	}
	
	// Get orders
	orders, err := h.store.GetUserOrders(userID, filters)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}
	
	// Get total count for pagination
	totalCount, err := h.store.GetOrdersCount(userID, filters)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}
	
	// Determine if there are more results
	hasMore := len(orders) == filters.Limit && (filters.Offset+len(orders)) < totalCount
	
	response := types.OrderListResponse{
		Orders:  orders,
		Total:   totalCount,
		HasMore: hasMore,
	}
	
	utils.WriteJSON(w, http.StatusOK, response)
}

// GET /api/v1/orders/{id} - get specific order details
func (h *Handler) handleGetOrder(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserIDFromContext(r.Context())
	
	vars := mux.Vars(r)
	orderID, err := strconv.Atoi(vars["id"])
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid order ID"))
		return
	}
	
	order, err := h.store.GetOrderByID(orderID, userID)
	if err != nil {
		if err.Error() == "order not found or not owned by user" {
			utils.WriteError(w, http.StatusNotFound, fmt.Errorf("order not found"))
			return
		}
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}
	
	utils.WriteJSON(w, http.StatusOK, order)
}