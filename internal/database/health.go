package database

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
		stats["status"] = "down"
		stats["error"] = fmt.Sprintf("db down: %v\n", err)

		log.Printf("db down: %v", err)

		return stats
	}

	stats["status"] = "up"
	stats["message"] = "It's healthy"

	poolStats := ps.pool.Stat()

	stats["acquire_count"] = strconv.FormatInt(poolStats.AcquireCount(), 10)
	stats["acquire_duration"] = poolStats.AcquireDuration().String()
	stats["acquired_conns"] = strconv.FormatInt(int64(poolStats.AcquiredConns()), 10)
	stats["canceled_acquire_count"] = strconv.FormatInt(poolStats.CanceledAcquireCount(), 10)
	stats["constructing_conns"] = strconv.FormatInt(int64(poolStats.ConstructingConns()), 10)
	stats["empty_acquire_count"] = strconv.FormatInt(poolStats.EmptyAcquireCount(), 10)
	stats["empty_acquire_wait_time"] = poolStats.EmptyAcquireWaitTime().String()
	stats["idle_conns"] = strconv.FormatInt(int64(poolStats.IdleConns()), 10)
	stats["max_conns"] = strconv.FormatInt(int64(poolStats.MaxConns()), 10)
	stats["max_idle_destroy_count"] = strconv.FormatInt(poolStats.MaxIdleDestroyCount(), 10)
	stats["max_lifetime_destroy_count"] = strconv.FormatInt(poolStats.MaxLifetimeDestroyCount(), 10)
	stats["new_conns_count"] = strconv.FormatInt(poolStats.NewConnsCount(), 10)
	stats["total_conns"] = strconv.FormatInt(int64(poolStats.TotalConns()), 10)

	if poolStats.TotalConns() > 10 {
		stats["warning"] = "Total number of database connections is higher than normal. This could indicate a potential connection leak or increased load."
	}

	if poolStats.CanceledAcquireCount() > 5 {
		stats["warning"] = "Frequent cancellations while acquiring connections. This may signal contention or slow queries."
	}

	if ps.pool.Config().MaxConns > 20 {
		stats["warning"] = "Maximum allowed database connections is set high. Ensure the database can handle this load."
	}

	return stats
}
