package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/dinesh-man/ecommerce-order-processing-system/queue-service/queue"
)

func main() {
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		log.Fatal("REDIS_ADDR not specified")
	}

	streamKey := os.Getenv("STREAM_KEY")
	if streamKey == "" {
		log.Fatal("STREAM_KEY not specified")
	}

	// init Redis
	err := queue.InitRedis(redisAddr)
	if err != nil {
		log.Fatalf("Failed to start Redis: %v", err)
	}

	log.Println("Redis Queue service is running on port 6379")

	http.HandleFunc("/queue/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	http.HandleFunc("/queue/size", func(w http.ResponseWriter, r *http.Request) {
		res, err := queue.QueueLength(streamKey)
		if err != nil {
			http.Error(w, "Failed to get queue length: "+err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write([]byte(fmt.Sprintf("%d", res)))
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}
