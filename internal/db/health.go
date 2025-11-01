package db

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"
)

func (ps *postgresService) Health() map[string]string {
	stats := make(map[string]string)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := ps.pool.Ping(ctx); err != nil {
		stats["STATUS"] = fmt.Sprintf("db down: %v\n", err)
		log.Printf("ERROR: postgresService health check failed: %v", err)
		return stats
	}

	postgresStats := ps.pool.Stat()

	stats["acquire_count"] = strconv.FormatInt(postgresStats.AcquireCount(), 10)
	stats["acquire_duration"] = postgresStats.AcquireDuration().String()
	stats["acquired_conns"] = strconv.FormatInt(int64(postgresStats.AcquiredConns()), 10)
	stats["canceled_acquire_count"] = strconv.FormatInt(postgresStats.CanceledAcquireCount(), 10)
	stats["constructing_conns"] = strconv.FormatInt(int64(postgresStats.ConstructingConns()), 10)
	stats["empty_acquire_count"] = strconv.FormatInt(postgresStats.EmptyAcquireCount(), 10)
	stats["empty_acquire_wait_time"] = postgresStats.EmptyAcquireWaitTime().String()
	stats["idle_conns"] = strconv.FormatInt(int64(postgresStats.IdleConns()), 10)
	stats["max_conns"] = strconv.FormatInt(int64(postgresStats.MaxConns()), 10)
	stats["max_idle_destroy_count"] = strconv.FormatInt(postgresStats.MaxIdleDestroyCount(), 10)
	stats["max_lifetime_destroy_count"] = strconv.FormatInt(postgresStats.MaxLifetimeDestroyCount(), 10)
	stats["new_conns_count"] = strconv.FormatInt(postgresStats.NewConnsCount(), 10)
	stats["total_conns"] = strconv.FormatInt(int64(postgresStats.TotalConns()), 10)

	if postgresStats.AcquireDuration() > 500*time.Millisecond {
		stats["WARN"] = "PERFORMANCE ALERT: Connection acquisition is taking too long. Consider increasing pool size or optimizing queries."
	}

	if postgresStats.CanceledAcquireCount() > 50 {
		stats["WARN"] = "STABILITY ALERT: Some connection acquisition requests were canceled. This usually happens when the pool is full or requests timeout."
	}

	if postgresStats.IdleConns() == 0 {
		stats["WARN"] = "STABILITY ALERT: No idle connections available. New requests may be blocked. Consider increasing pool size or tuning the connection pool."
	}

	if postgresStats.MaxConns() > 50 {
		stats["WARN"] = "RESOURCE ALERT: Connection pool has reached its maximum capacity. New requests may be delayed or fail."
	}

	stats["STATUS"] = "It's helathy"

	return stats
}
