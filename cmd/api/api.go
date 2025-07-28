package api

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/HollyEllmo/go_rest_tut/cmd/service/cart"
	"github.com/HollyEllmo/go_rest_tut/cmd/service/inventory"
	"github.com/HollyEllmo/go_rest_tut/cmd/service/order"
	"github.com/HollyEllmo/go_rest_tut/cmd/service/product"
	"github.com/HollyEllmo/go_rest_tut/cmd/service/user"
	"github.com/gorilla/mux"
)

type APIServer struct {
	addr string
	db   *sql.DB
}

func NewAPIServer(addr string, db *sql.DB) *APIServer {
	return &APIServer{
		addr: addr,
		db:   db,
	}
}

func (s *APIServer) Run() error {
	router := mux.NewRouter()
	subrouter := router.PathPrefix("/api/v1").Subrouter()

	userStore := user.NewStore(s.db)
	userHandler := user.NewHandler(userStore)
	userHandler.RegisterRoutes(subrouter)

	productStore := product.NewStore(s.db)
	productHandler := product.NewHandler(productStore)
	productHandler.RegisterRoutes(subrouter)

	orderStore := order.NewStore(s.db)
	inventoryStore := inventory.NewStore(s.db)
	cartHandler := cart.NewHandler(orderStore, productStore, userStore, inventoryStore)
	cartHandler.RegisterRoutes(subrouter)

	inventoryHandler := inventory.NewHandler(inventoryStore, userStore)
	inventoryHandler.RegisterRoutes(subrouter)

	log.Println("Listening on", s.addr)

	return http.ListenAndServe(s.addr, router)
}
