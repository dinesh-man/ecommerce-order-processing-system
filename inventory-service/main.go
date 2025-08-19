package main

import (
	"log"
	"net/http"
	"os"

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

	log.Println("Inventory service running on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
