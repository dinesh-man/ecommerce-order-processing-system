package queue

import (
	"context"
	"log"

	"github.com/redis/go-redis/v9"
)

var (
	ctx = context.Background()
	rdb *redis.Client
)

// Iinitializes Redis connection
func InitRedis(addr string) error {
	rdb = redis.NewClient(&redis.Options{
		Addr: addr,
	})

	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		log.Printf("Redis ping failed: %v", err)
		return err
	}

	log.Println("Redis ping successful")
	return nil
}

func QueueLength(streamKey string) (int64, error) {
	return rdb.LLen(ctx, streamKey).Result()
}
