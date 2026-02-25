package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"food_ordering_coordination_system/internal/domain"
	httpapi "food_ordering_coordination_system/internal/http"
	"food_ordering_coordination_system/internal/integration"
	"food_ordering_coordination_system/internal/integration/adapters"
)

func main() {
	cfg, err := LoadFromEnv()
	if err != nil {
		log.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(
		context.Background(),
		15*time.Second,
	)
	defer cancel()

	repo, client, err := ConnectMongoRepository(ctx, cfg)
	if err != nil {
		log.Fatalf(
			"connect mongo repository: %v",
			err,
		)
	}
	defer func() {
		disconnectCtx, disconnectCancel := context.WithTimeout(
			context.Background(),
			5*time.Second,
		)
		defer disconnectCancel()
		_ = client.Disconnect(
			disconnectCtx,
		)
	}()

	aggregator := buildAggregator()

	service := domain.NewFoodOrderingService(repo, repo, repo)
	router := httpapi.NewFoodOrderingRouter(
		service, aggregator,
	)

	log.Printf(
		"api listening on :%s",
		cfg.Port,
	)
	if err := http.ListenAndServe(
		":"+cfg.Port, router,
	); err != nil {
		log.Fatal(err)
	}
}

// buildAggregator creates the fan-in/fan-out aggregator and registers vendor
// adapters from environment configuration.
//
// Set VENDOR_URLS to a comma-separated list of id=name=url triples:
//
//	VENDOR_URLS=pizza=Pizza Place=http://localhost:4001,sushi=Sushi Bar=http://localhost:4002
//
// When VENDOR_URLS is empty the aggregator starts with no adapters and the
// menu/vendor endpoints return empty arrays.
func buildAggregator() *integration.Aggregator {
	agg := integration.NewAggregator()

	raw := os.Getenv("VENDOR_URLS")
	if raw == "" {
		return agg
	}

	for _, entry := range strings.Split(raw, ",") {
		parts := strings.SplitN(strings.TrimSpace(entry), "=", 3)
		if len(parts) != 3 {
			log.Printf("skip malformed VENDOR_URLS entry: %q", entry)
			continue
		}
		id, name, url := parts[0], parts[1], parts[2]
		agg.Register(adapters.NewJSONServerAdapter(id, name, url, nil))
		log.Printf("registered vendor adapter: %s (%s) -> %s", id, name, url)
	}

	return agg
}
