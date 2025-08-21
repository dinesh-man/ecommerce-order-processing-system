package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dinesh-man/ecommerce-order-processing-system/pkg/redis-stream"
	"github.com/dinesh-man/ecommerce-order-processing-system/queue-service/queue"
)

type Response struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}

func writeJSONResponse(w http.ResponseWriter, message string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(Response{
		Message: message,
		Code:    code,
	})
}

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
		writeJSONResponse(w, "OK", http.StatusOK)
	})

	http.HandleFunc("/queue/size", func(w http.ResponseWriter, r *http.Request) {
		res, err := queue.QueueLength(streamKey)
		if err != nil {
			writeJSONResponse(w, fmt.Sprintf("Failed to get queue length: %s", err.Error()), http.StatusInternalServerError)
			return
		}
		writeJSONResponse(w, fmt.Sprintf("Current queue size is %d", res), http.StatusOK)
	})

	server := &http.Server{Addr: ":8080"}

	go func() {
		log.Println("Queue service running on :8080")
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

	// Close Redis connection
	redis_stream.CloseRedis()

	log.Println("Order service shutdown complete.")
}
