package main

import (
	"context"
	"log"
	"net/http"
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

	aggregator := buildAggregator(cfg.VendorURLs)

	service := domain.NewFoodOrderingService(repo, repo, repo)
	authenticator := httpapi.NewAuthenticator(cfg.JWTSigningKey)
	authController := httpapi.NewAuthController(repo, authenticator, cfg.AuthTokenTTL, cfg.AuthAllowSelfAssignRoles)
	router := httpapi.NewFoodOrderingRouterWithAuth(service, aggregator, authController, cfg.JWTSigningKey)

	log.Printf(
		"api listening on :%s",
		cfg.Port,
	)
	if err := http.ListenAndServe(
		":"+cfg.Port, withCORS(router),
	); err != nil {
		log.Fatal(err)
	}
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func buildAggregator(raw string) *integration.Aggregator {
	agg := integration.NewAggregator()
	if strings.TrimSpace(raw) == "" {
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
