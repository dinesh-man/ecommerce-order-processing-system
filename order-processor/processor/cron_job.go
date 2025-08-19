package processor

import (
	"context"
	"log"
	"time"

	"github.com/dinesh-man/ecommerce-order-processing-system/pkg/mongodb"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func RunJob(redisClient *redis.Client, streamKey string, collectionName string) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		log.Println("Cron job triggered...")
		processOrders(redisClient, streamKey, collectionName)
		<-ticker.C
	}
}

// Updates PENDING orders to PROCESSING every 5 minutes
func processOrders(rdb *redis.Client, streamKey string, collectionName string) {
	ctx := context.Background()

	//Read Redis queue
	res, err := rdb.XRead(ctx, &redis.XReadArgs{
		Streams: []string{streamKey, "0"},
		Block:   5 * time.Second,
	}).Result()

	if err != nil && err != redis.Nil {
		log.Printf("Redis read error: %v", err)
		return
	}

	var pendingOrderIDs []primitive.ObjectID
	var processedMessages []redis.XMessage

	for _, stream := range res {
		for _, msg := range stream.Messages {
			orderIDStr, ok := msg.Values["order_id"].(string)
			if !ok {
				log.Printf("Invalid order_id format: %v", msg.Values["order_id"])
				continue
			}

			orderID, err := primitive.ObjectIDFromHex(orderIDStr)
			if err != nil {
				log.Printf("Invalid ObjectID: %s, skipping", orderIDStr)
				continue
			}

			pendingOrderIDs = append(pendingOrderIDs, orderID)
			processedMessages = append(processedMessages, msg)
			log.Printf("Processing order %s in cron job", orderID)
		}
	}

	if len(pendingOrderIDs) == 0 {
		log.Println("No orders to process in the queue")
		return
	}

	//Update order status PENDING -> PROCESSING in bulk
	filter := bson.M{"_id": bson.M{"$in": pendingOrderIDs}, "status": "PENDING"}
	update := bson.M{"$set": bson.M{"status": "PROCESSING"}}
	collection := mongodb.GetCollection(collectionName)

	result, err := collection.UpdateMany(ctx, filter, update)
	if err != nil {
		log.Printf("Error updating orders: %v", err)
		return
	}

	log.Printf("Bulk update completed. Orders updated to PROCESSING: %d", result.ModifiedCount)

	//Clean up stream entries only after DB update succeeds
	for _, msg := range processedMessages {
		if err := rdb.XDel(ctx, streamKey, msg.ID).Err(); err != nil {
			log.Printf("Failed to delete stream entry %s: %v", msg.ID, err)
		} else {
			log.Printf("Deleted stream entry: %s", msg.ID)
		}
	}
}
