package main

import (
	"log"
	"net/http"
	"os"

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

	// cancel order by /order/cancel?id=
	http.HandleFunc("/order/cancel", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			orderHandler.CancelOrderHandler(w, r)
			return
		}
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	})

	log.Println("Order service running on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
