package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/dinesh-man/ecommerce-order-processing-system/order-service/service"
	"github.com/dinesh-man/ecommerce-order-processing-system/pkg/models"
)

// OrderHandler handles HTTP requests for orders
type OrderHandler struct {
	service *service.OrderService
}

type ErrorResponse struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}

func NewOrderHandler(s *service.OrderService) *OrderHandler {
	return &OrderHandler{service: s}
}

func writeJSONError(w http.ResponseWriter, message string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(ErrorResponse{
		Message: message,
		Code:    code,
	})
}

// CreateOrderHandler handles POST /order
func (h *OrderHandler) CreateOrderHandler(w http.ResponseWriter, r *http.Request) {

	log.Printf("Received %s request for %s from %s", r.Method, r.URL.Path, r.RemoteAddr)

	if r.Method != http.MethodPost {
		writeJSONError(w, fmt.Sprintf("Method not allowed: %s", r.Method), http.StatusMethodNotAllowed)
		return
	}

	var order models.Order
	if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
		writeJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if len(order.Items) == 0 {
		writeJSONError(w, "Order must contain at least one item", http.StatusBadRequest)
		return
	}

	createdOrder, err := h.service.CreateOrder(order)
	if err != nil {
		writeJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdOrder)
}

// GetOrderHandler handles GET /order/?id=123
func (h *OrderHandler) GetOrderHandler(w http.ResponseWriter, r *http.Request) {

	log.Printf("Received %s request for %s from %s", r.Method, r.URL.Path, r.RemoteAddr)

	id := r.URL.Query().Get("id")

	if id == "" {
		writeJSONError(w, "Missing order id", http.StatusBadRequest)
		return
	}

	order, err := h.service.GetOrderByID(id)
	if err != nil {
		writeJSONError(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(order)
}

// ListOrdersHandler handles GET /orders?status=PENDING
func (h *OrderHandler) ListOrdersHandler(w http.ResponseWriter, r *http.Request) {

	log.Printf("Received %s request for %s from %s", r.Method, r.URL.Path, r.RemoteAddr)

	status := r.URL.Query().Get("status")
	cursor := r.URL.Query().Get("cursor")

	// Default pageSize to 10 if pagesize is not provided
	pageSize := int64(10)
	if ps := r.URL.Query().Get("pageSize"); ps != "" {
		if parsed, err := strconv.ParseInt(ps, 10, 64); err == nil && parsed > 0 {
			pageSize = parsed
		}
	}

	orders, nextCursor, err := h.service.ListOrders(status, cursor, pageSize)
	if err != nil {
		writeJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"orders":      orders,
		"next_cursor": nextCursor,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// CancelOrderHandler handles DELETE /order/cancel?id=123
func (h *OrderHandler) CancelOrderHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received %s request for %s from %s", r.Method, r.URL.Path, r.RemoteAddr)

	id := r.URL.Query().Get("id")

	if id == "" {
		writeJSONError(w, "Missing order id", http.StatusBadRequest)
		return
	}

	err := h.service.CancelOrder(id)
	if err != nil {
		writeJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message":"Order cancelled successfully"}`))
}
