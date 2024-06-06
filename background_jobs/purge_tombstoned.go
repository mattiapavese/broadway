package background

import (
	"context"
	"fmt"
	"time"

	redisclient "github.com/mattiapavese/broadway/redis_client"
	handlers "github.com/mattiapavese/broadway/router_handlers"
)

func PurgeTombstoned() {
	rdb := redisclient.GetRedisClient("localhost", "6379", "")
	ctx := context.Background()

	for {
		tombstonedConnections, _ := rdb.SMembers(ctx, handlers.TOMBSTONED).Result()

		for _, conn := range tombstonedConnections {

			fmt.Println("Purging.")

			linkedChannels, _ := rdb.SMembers(ctx, conn+":link").Result()
			for _, chann := range linkedChannels {
				rdb.SRem(ctx, chann, conn).Result()
			}

			host, _ := rdb.Get(ctx, conn).Result()

			rdb.SRem(ctx, host, conn).Result()

			rdb.Del(ctx, conn+":link", conn).Result()

			rdb.SRem(ctx, handlers.TOMBSTONED, conn).Result()

		}

		time.Sleep(5 * time.Second)
	}

}
