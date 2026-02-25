package handler

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

	"handler/internal/domain"
	httpapi "handler/internal/http"
	"handler/internal/integration"
	"handler/internal/integration/adapters"
	persistence "handler/internal/persistance"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type dbConfig struct {
	uri      string
	database string
}

type serverlessConfig struct {
	vendorURLs string
	db         dbConfig
}

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

	repo, _, err := connectMongoRepository(ctx, cfg)
	if err != nil {
		serverlessInitErr = fmt.Errorf("connect mongo repository: %w", err)
		return
	}

	aggregator := buildAggregator(cfg.vendorURLs)
	service := domain.NewFoodOrderingService(repo, repo, repo)
	serverlessRouter = withCORS(httpapi.NewFoodOrderingRouter(service, aggregator))
}

func loadServerlessConfig() (serverlessConfig, error) {
	cfg := serverlessConfig{
		vendorURLs: strings.TrimSpace(os.Getenv("VENDOR_URLS")),
		db: dbConfig{
			uri:      strings.TrimSpace(os.Getenv("MONGODB_URI")),
			database: strings.TrimSpace(os.Getenv("MONGODB_DATABASE")),
		},
	}

	var missing []string
	if cfg.db.uri == "" {
		missing = append(missing, "MONGODB_URI")
	}
	if cfg.db.database == "" {
		missing = append(missing, "MONGODB_DATABASE")
	}
	if len(missing) > 0 {
		return serverlessConfig{}, errors.New("missing required environment variables: " + strings.Join(missing, ", "))
	}

	if cfg.vendorURLs != "" {
		if err := validateVendorURLs(cfg.vendorURLs); err != nil {
			return serverlessConfig{}, err
		}
	}

	return cfg, nil
}

func connectMongoRepository(ctx context.Context, cfg serverlessConfig) (*persistence.MongoRepository, *mongo.Client, error) {
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.db.uri))
	if err != nil {
		return nil, nil, err
	}

	repo := persistence.NewMongoRepository(client.Database(cfg.db.database))
	if err := repo.EnsureSchema(ctx); err != nil {
		_ = client.Disconnect(ctx)
		return nil, nil, err
	}

	return repo, client, nil
}

func buildAggregator(raw string) *integration.Aggregator {
	agg := integration.NewAggregator()
	if strings.TrimSpace(raw) == "" {
		return agg
	}

	for _, entry := range strings.Split(raw, ",") {
		parts := strings.SplitN(strings.TrimSpace(entry), "=", 3)
		if len(parts) != 3 {
			continue
		}
		id, name, baseURL := parts[0], parts[1], parts[2]
		agg.Register(adapters.NewJSONServerAdapter(id, name, baseURL, nil))
	}

	return agg
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

func validateVendorURLs(raw string) error {
	entries := strings.Split(raw, ",")
	for _, entry := range entries {
		parts := strings.SplitN(strings.TrimSpace(entry), "=", 3)
		if len(parts) != 3 {
			return errors.New("VENDOR_URLS entries must follow id=name=url format")
		}
		id := strings.TrimSpace(parts[0])
		name := strings.TrimSpace(parts[1])
		baseURL := strings.TrimSpace(parts[2])
		if id == "" || name == "" || baseURL == "" {
			return errors.New("VENDOR_URLS entries must include non-empty id, name, and url")
		}
		parsed, err := url.ParseRequestURI(baseURL)
		if err != nil || parsed.Scheme == "" || parsed.Host == "" {
			return fmt.Errorf("VENDOR_URLS entry has invalid url: %q", baseURL)
		}
	}
	return nil
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
