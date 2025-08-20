package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dinesh-man/ecommerce-order-processing-system/order-service/handler"
	"github.com/dinesh-man/ecommerce-order-processing-system/order-service/service"
	"github.com/dinesh-man/ecommerce-order-processing-system/pkg/mongodb"
	"github.com/dinesh-man/ecommerce-order-processing-system/pkg/redis-stream"
)

func main() {

	collectionName := os.Getenv("COLLECTION_NAME")
	if collectionName == "" {
		log.Fatal("Collection name not specified")
	}

	inventoryServiceURL := os.Getenv("INVENTORY_SERVICE_URL")
	if inventoryServiceURL == "" {
		log.Fatal("Inventory-service URL not specified")
	}

	mongodb.InitMongoDB()
	rdb, sk := redis_stream.InitRedis()

	orderService := service.NewOrderService(collectionName, inventoryServiceURL, rdb, sk)
	orderHandler := handler.NewOrderHandler(orderService)

	// Create and order or get order by /order?id=123
	http.HandleFunc("/order", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			orderHandler.CreateOrderHandler(w, r)
		case http.MethodGet:
			orderHandler.GetOrderHandler(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// list orders with optional /orders?status=
	http.HandleFunc("/orders", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			orderHandler.ListOrdersHandler(w, r)
			return
		}
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	})

	// Cancel order by /order/cancel?id=
	http.HandleFunc("/order/cancel", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			orderHandler.CancelOrderHandler(w, r)
			return
		}
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	})

	server := &http.Server{Addr: ":8080"}

	go func() {
		log.Println("Order service running on :8080")
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	// Listen for termination signals
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	<-stop // Block until signal is received
	log.Println("Shutdown signal received. Cleaning up...")

	// Gracefully shutdown HTTP server
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("HTTP server shutdown error: %v", err)
	}

	// Disconnect MongoDB to release resources acquired for connection pooling
	mongodb.DisconnectMongo()
	// Close Redis connection
	redis_stream.CloseRedis()

	log.Println("Order service shutdown complete.")
}
