package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/dinesh-man/ecommerce-order-processing-system/inventory-service/service"
)

type InventoryHandler struct {
	service *service.InventoryService
}

func NewInventoryHandler(s *service.InventoryService) *InventoryHandler {
	return &InventoryHandler{service: s}
}

// GetAllProductsHandler handles GET /products
func (h *InventoryHandler) GetAllProductsHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received %s request for %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	products, err := h.service.GetAllProducts()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(products)
}

// GetProductByIdHandler handles GET /product?id=P001
func (h *InventoryHandler) GetProductByIdHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received %s request for %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := r.URL.Query().Get("id")
	product, err := h.service.GetProductByID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(product)
}
