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

	"github.com/dinesh-man/ecommerce-order-processing-system/order-processor/processor"
	"github.com/dinesh-man/ecommerce-order-processing-system/pkg/mongodb"
	"github.com/dinesh-man/ecommerce-order-processing-system/pkg/redis-stream"
	"github.com/google/uuid"
)

func main() {

	collectionName := os.Getenv("COLLECTION_NAME")
	if collectionName == "" {
		log.Fatal("COLLECTION_NAME not specified")
	}

	jobRunIntervalMints := os.Getenv("JOB_RUN_INTERVAL_MINUTES")
	if jobRunIntervalMints == "" {
		log.Fatal("JOB_RUN_INTERVAL_MINUTES not specified")
	}

	consumerGroup := os.Getenv("CONSUMER_GROUP")
	if consumerGroup == "" {
		log.Fatal("CONSUMER_GROUP not specified")
	}

	duration, err := time.ParseDuration(jobRunIntervalMints)
	if err != nil {
		log.Fatal("invalid job run interval specified")
	}

	// check duration is at least 1 minute
	// this is to prevent overwhelming the system with too many job runs
	if duration < time.Minute {
		log.Fatal("JOB_RUN_INTERVAL_MINUTES must be at least 1 minute")
	}

	mongodb.InitMongoDB()

	rdb, sk := redis_stream.InitRedis()

	// Start background job
	go processor.RunJob(rdb, sk, consumerGroup, uuid.NewString(), collectionName, duration)

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("OK")) })

	server := &http.Server{Addr: ":8080"}

	go func() {
		log.Println("Order processor running on :8080")
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

	log.Println("Order processor shutdown complete.")
}
