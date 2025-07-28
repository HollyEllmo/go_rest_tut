package address

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
	store     types.AddressStore
	userStore types.UserStore
}

func NewHandler(store types.AddressStore, userStore types.UserStore) *Handler {
	return &Handler{store: store, userStore: userStore}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	// All address routes require authentication
	router.HandleFunc("/addresses", auth.WithJWTAuth(h.handleGetAddresses, h.userStore)).Methods(http.MethodGet)
	router.HandleFunc("/addresses", auth.WithJWTAuth(h.handleCreateAddress, h.userStore)).Methods(http.MethodPost)
	router.HandleFunc("/addresses/{id}", auth.WithJWTAuth(h.handleGetAddress, h.userStore)).Methods(http.MethodGet)
	router.HandleFunc("/addresses/{id}", auth.WithJWTAuth(h.handleUpdateAddress, h.userStore)).Methods(http.MethodPut)
	router.HandleFunc("/addresses/{id}", auth.WithJWTAuth(h.handleDeleteAddress, h.userStore)).Methods(http.MethodDelete)
	router.HandleFunc("/addresses/{id}/default", auth.WithJWTAuth(h.handleSetDefaultAddress, h.userStore)).Methods(http.MethodPost)
}

// GET /api/v1/addresses - get all addresses for authenticated user
func (h *Handler) handleGetAddresses(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserIDFromContext(r.Context())

	addresses, err := h.store.GetUserAddresses(userID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]any{
		"addresses": addresses,
		"count":     len(addresses),
	})
}

// POST /api/v1/addresses - create new address
func (h *Handler) handleCreateAddress(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserIDFromContext(r.Context())

	var payload types.CreateAddressPayload
	if err := utils.ParseJSON(r, &payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	if err := utils.Validate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		utils.WriteError(w, http.StatusBadRequest, errors)
		return
	}

	address, err := h.store.CreateAddress(userID, payload)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusCreated, address)
}

// GET /api/v1/addresses/{id} - get specific address
func (h *Handler) handleGetAddress(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserIDFromContext(r.Context())
	
	vars := mux.Vars(r)
	addressID, err := strconv.Atoi(vars["id"])
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	address, err := h.store.GetAddressByID(addressID, userID)
	if err != nil {
		if err.Error() == "address not found" {
			utils.WriteError(w, http.StatusNotFound, err)
			return
		}
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, address)
}

// PUT /api/v1/addresses/{id} - update address
func (h *Handler) handleUpdateAddress(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserIDFromContext(r.Context())
	
	vars := mux.Vars(r)
	addressID, err := strconv.Atoi(vars["id"])
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	var payload types.UpdateAddressPayload
	if err := utils.ParseJSON(r, &payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	if err := utils.Validate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		utils.WriteError(w, http.StatusBadRequest, errors)
		return
	}

	address, err := h.store.UpdateAddress(addressID, userID, payload)
	if err != nil {
		if err.Error() == "address not found or not owned by user" {
			utils.WriteError(w, http.StatusNotFound, err)
			return
		}
		if err.Error() == "no fields to update" {
			utils.WriteError(w, http.StatusBadRequest, err)
			return
		}
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, address)
}

// DELETE /api/v1/addresses/{id} - delete address
func (h *Handler) handleDeleteAddress(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserIDFromContext(r.Context())
	
	vars := mux.Vars(r)
	addressID, err := strconv.Atoi(vars["id"])
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	err = h.store.DeleteAddress(addressID, userID)
	if err != nil {
		if err.Error() == "address not found or not owned by user" {
			utils.WriteError(w, http.StatusNotFound, err)
			return
		}
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]string{
		"message": "Address deleted successfully",
	})
}

// POST /api/v1/addresses/{id}/default - set address as default
func (h *Handler) handleSetDefaultAddress(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserIDFromContext(r.Context())
	
	vars := mux.Vars(r)
	addressID, err := strconv.Atoi(vars["id"])
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	err = h.store.SetDefaultAddress(addressID, userID)
	if err != nil {
		if err.Error() == "address not found or not owned by user" {
			utils.WriteError(w, http.StatusNotFound, err)
			return
		}
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]string{
		"message": "Default address updated successfully",
	})
}