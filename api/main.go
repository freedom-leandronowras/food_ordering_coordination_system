package main

import (
	"context"
	"food_ordering_coordination_system/internal/domain"
	httpapi "food_ordering_coordination_system/internal/http"
	"log"
	"net/http"
	"time"
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

	service := domain.NewFoodOrderingService(repo, repo)
	router := httpapi.NewFoodOrderingRouter(
		service,
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
