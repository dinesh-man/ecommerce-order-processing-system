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

func RunJob(redisClient *redis.Client, streamKey string, group string, consumerID string, collectionName string, jobRunIntervalMints time.Duration) {
	ticker := time.NewTicker(jobRunIntervalMints)
	defer ticker.Stop()

	for {
		log.Println("Cron job triggered...")
		ctx := context.Background()
		processOrders(ctx, redisClient, streamKey, group, consumerID, collectionName)
		<-ticker.C
	}
}

// Updates PENDING orders to PROCESSING every 5 minutes
func processOrders(ctx context.Context, rdb *redis.Client, streamKey string, group string, consumerID, collectionName string) {

	// Check if pending messages exist in the stream
	reclaimedMsgs := ReclaimStuckMessages(ctx, rdb, streamKey, group, consumerID)
	if len(reclaimedMsgs) > 0 {
		log.Printf("Reprocessing %d stuck messages from the stream", len(reclaimedMsgs))
		digestMessages(ctx, rdb, reclaimedMsgs, streamKey, group, collectionName)
	}

	//Read Redis consumer group for new messages
	res, err := rdb.XReadGroup(ctx, &redis.XReadGroupArgs{
		Group:    group,
		Consumer: consumerID,
		Streams:  []string{streamKey, ">"},
		Block:    5 * time.Second,
	}).Result()

	if err != nil && err != redis.Nil {
		log.Printf("Redis read error: %v", err)
		return
	}

	var newMsgs []redis.XMessage

	for _, stream := range res {
		for _, msg := range stream.Messages {
			claimed := rdb.XClaim(ctx, &redis.XClaimArgs{
				Stream:   streamKey,
				Group:    group,
				Consumer: consumerID,
				Messages: []string{msg.ID},
			}).Val()
			newMsgs = append(newMsgs, claimed...)
		}
	}

	if len(newMsgs) == 0 {
		log.Println("No new messages to process")
		return
	}

	digestMessages(ctx, rdb, newMsgs, streamKey, group, collectionName)
}

func digestMessages(ctx context.Context, rdb *redis.Client, messages []redis.XMessage, streamKey string, group string, collectionName string) {

	var pendingOrderIDs []primitive.ObjectID

	for _, msg := range messages {
		orderIDStr, _ := msg.Values["order_id"].(string)
		orderID, _ := primitive.ObjectIDFromHex(orderIDStr)
		pendingOrderIDs = append(pendingOrderIDs, orderID)

		log.Printf("Processing order: %s", orderIDStr)
	}

	//Update order status PENDING -> PROCESSING in bulk
	filter := bson.M{"_id": bson.M{"$in": pendingOrderIDs}, "status": "PENDING"}
	update := bson.M{"$set": bson.M{"status": "PROCESSING", "updated_at": time.Now()}}
	collection := mongodb.GetCollection(collectionName)

	result, err := collection.UpdateMany(ctx, filter, update)
	if err != nil {
		log.Printf("Error updating orders: %v", err)
		return
	}

	log.Printf("Bulk update completed. Orders updated to PROCESSING: %d", result.ModifiedCount)

	//Clean up stream entries only after DB update succeeds
	for _, msg := range messages {
		if err := rdb.XAck(ctx, streamKey, group, msg.ID).Err(); err != nil {
			log.Printf("Failed to ACK stream entry %s: %v", msg.ID, err)
			continue
		}
		if err := rdb.XDel(ctx, streamKey, msg.ID).Err(); err != nil {
			log.Printf("Failed to delete stream entry %s: %v", msg.ID, err)
		} else {
			log.Printf("Deleted stream entry: %s", msg.ID)
		}
	}
}
