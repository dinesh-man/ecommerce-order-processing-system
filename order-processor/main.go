package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/dinesh-man/ecommerce-order-processing-system/order-processor/processor"
	"github.com/dinesh-man/ecommerce-order-processing-system/pkg/mongodb"
	"github.com/dinesh-man/ecommerce-order-processing-system/pkg/redis-stream"
)

func main() {

	collectionName := os.Getenv("COLLECTION_NAME")
	if collectionName == "" {
		log.Fatal("Collection name not specified")
	}

	mongodb.InitMongoDB()

	rdb, sk := redis_stream.InitRedis()

	// Start background job
	go processor.RunJob(rdb, sk, collectionName)

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("OK")) })

	fmt.Println("Order Processor running on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
