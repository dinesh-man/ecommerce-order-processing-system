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

	"github.com/dinesh-man/ecommerce-order-processing-system/inventory-service/handler"
	"github.com/dinesh-man/ecommerce-order-processing-system/inventory-service/service"
	"github.com/dinesh-man/ecommerce-order-processing-system/pkg/mongodb"
)

func main() {

	collectionName := os.Getenv("COLLECTION_NAME")
	if collectionName == "" {
		log.Fatal("Collection name not specified")
	}

	mongodb.InitMongoDB()

	inventoryService := service.NewInventoryService(collectionName)
	inventoryHandler := handler.NewInventoryHandler(inventoryService)

	http.HandleFunc("/products", inventoryHandler.GetAllProductsHandler)
	http.HandleFunc("/product", inventoryHandler.GetProductByIdHandler)

	server := &http.Server{Addr: ":8080"}

	go func() {
		log.Println("Inventory service running on :8080")
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

	log.Println("Inventory service shutdown complete.")
}
