// Command discovery-worker materializes discovery event metrics. Run it as a
// short-lived scheduled job in production, or set DISCOVERY_WORKER_INTERVAL
// to keep a local worker running.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/hacuba/listings/internal/listings"
)

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("discovery-worker: DATABASE_URL is required; SQLite is development-only and has no aggregate worker")
	}
	batchSize, err := envInt("DISCOVERY_BATCH_SIZE", 250, 1, 1000)
	if err != nil {
		log.Fatal("discovery-worker: ", err)
	}
	store, err := listings.NewPostgresStore(context.Background(), dsn)
	if err != nil {
		log.Fatal("discovery-worker: connect: ", err)
	}
	defer store.Close()

	if os.Getenv("DISCOVERY_WORKER_INTERVAL") == "" {
		runUntilCaughtUp(store, batchSize)
		return
	}
	interval, err := time.ParseDuration(os.Getenv("DISCOVERY_WORKER_INTERVAL"))
	if err != nil || interval < time.Second {
		log.Fatal("discovery-worker: DISCOVERY_WORKER_INTERVAL must be at least 1s")
	}
	for {
		runUntilCaughtUp(store, batchSize)
		time.Sleep(interval)
	}
}

func runUntilCaughtUp(store *listings.PostgresStore, batchSize int) {
	total := 0
	for {
		processed, err := store.AggregateDiscoveryEvents(context.Background(), batchSize)
		if err != nil {
			log.Fatal("discovery-worker: aggregate: ", err)
		}
		total += processed
		if processed < batchSize {
			log.Printf("discovery-worker: aggregated %d events", total)
			return
		}
	}
}

func envInt(name string, fallback, min, max int) (int, error) {
	value := os.Getenv(name)
	if value == "" {
		return fallback, nil
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < min || parsed > max {
		return 0, fmt.Errorf("%s must be an integer from %d through %d", name, min, max)
	}
	return parsed, nil
}
