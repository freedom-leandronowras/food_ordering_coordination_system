package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"food_ordering_coordination_system/internal/domain"
	httpapi "food_ordering_coordination_system/internal/http"
)

var (
	serverlessInitOnce sync.Once
	serverlessRouter   http.Handler
	serverlessInitErr  error
)

// Handler is the Vercel entrypoint for the Go API.
func Handler(w http.ResponseWriter, r *http.Request) {
	serverlessInitOnce.Do(initializeServerlessRouter)

	if serverlessInitErr != nil {
		http.Error(w, serverlessInitErr.Error(), http.StatusInternalServerError)
		return
	}

	proxiedRequest := cloneRequestWithPath(r, resolveAPIPath(r))
	serverlessRouter.ServeHTTP(w, proxiedRequest)
}

func initializeServerlessRouter() {
	cfg, err := loadServerlessConfig()
	if err != nil {
		serverlessInitErr = err
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	repo, _, err := ConnectMongoRepository(ctx, cfg)
	if err != nil {
		serverlessInitErr = fmt.Errorf("connect mongo repository: %w", err)
		return
	}

	aggregator := buildAggregator(cfg.VendorURLs)
	service := domain.NewFoodOrderingService(repo, repo, repo)
	serverlessRouter = withCORS(httpapi.NewFoodOrderingRouter(service, aggregator))
}

func loadServerlessConfig() (AppConfig, error) {
	cfg := AppConfig{
		Port:       "8080",
		VendorURLs: strings.TrimSpace(os.Getenv("VENDOR_URLS")),
		DB: DBConfig{
			URI:      strings.TrimSpace(os.Getenv("MONGODB_URI")),
			Database: strings.TrimSpace(os.Getenv("MONGODB_DATABASE")),
		},
	}

	var missing []string
	if cfg.DB.URI == "" {
		missing = append(missing, "MONGODB_URI")
	}
	if cfg.DB.Database == "" {
		missing = append(missing, "MONGODB_DATABASE")
	}
	if len(missing) > 0 {
		return AppConfig{}, errors.New("missing required environment variables: " + strings.Join(missing, ", "))
	}

	if cfg.VendorURLs != "" {
		if err := validateVendorURLs(cfg.VendorURLs); err != nil {
			return AppConfig{}, err
		}
	}

	return cfg, nil
}

func resolveAPIPath(r *http.Request) string {
	capturedPath := strings.TrimSpace(r.URL.Query().Get("path"))
	if capturedPath != "" {
		return ensureAPIPrefix(capturedPath)
	}

	if strings.HasPrefix(r.URL.Path, "/go-api") {
		suffix := strings.TrimPrefix(r.URL.Path, "/go-api")
		if suffix == "" {
			suffix = "/"
		}
		return ensureAPIPrefix(suffix)
	}

	return ensureAPIPrefix(r.URL.Path)
}

func ensureAPIPrefix(path string) string {
	cleaned := "/" + strings.TrimLeft(path, "/")
	if cleaned == "/" {
		return "/api"
	}
	if strings.HasPrefix(cleaned, "/api") {
		return cleaned
	}
	return "/api" + cleaned
}

func cloneRequestWithPath(r *http.Request, path string) *http.Request {
	cloned := r.Clone(r.Context())

	parsedURL := *r.URL
	parsedURL.Path = path
	parsedURL.RawPath = path
	parsedURL.RawQuery = dropPathQueryParam(r.URL.Query()).Encode()
	cloned.URL = &parsedURL

	return cloned
}

func dropPathQueryParam(values url.Values) url.Values {
	next := make(url.Values, len(values))
	for key, original := range values {
		if key == "path" {
			continue
		}
		copySlice := make([]string, len(original))
		copy(copySlice, original)
		next[key] = copySlice
	}
	return next
}
