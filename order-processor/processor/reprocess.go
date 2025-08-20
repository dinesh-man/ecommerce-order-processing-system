package processor

import (
	"context"

	"github.com/redis/go-redis/v9"
)

// ReclaimStuckMessages reprocess messages that are stuck in the pending state due to consumer failures
func ReclaimStuckMessages(ctx context.Context, rdb *redis.Client, streamKey, group, consumerID string) []redis.XMessage {

	pending := rdb.XPendingExt(ctx, &redis.XPendingExtArgs{
		Stream: streamKey,
		Group:  group,
		Start:  "-",
		End:    "+",
		Count:  10,
	}).Val()

	var reclaimedMsgs []redis.XMessage

	for _, entry := range pending {
		claimed := rdb.XClaim(ctx, &redis.XClaimArgs{
			Stream:   streamKey,
			Group:    group,
			Consumer: consumerID,
			Messages: []string{entry.ID},
		}).Val()
		reclaimedMsgs = append(reclaimedMsgs, claimed...)
	}

	return reclaimedMsgs
}
