package main

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"

	persistence "handler/internal/persistance"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type DBConfig struct {
	URI      string
	Database string
}

type AppConfig struct {
	Port       string
	VendorURLs string
	DB         DBConfig
}

func LoadFromEnv() (AppConfig, error) {
	cfg := AppConfig{
		Port:       strings.TrimSpace(os.Getenv("PORT")),
		VendorURLs: strings.TrimSpace(os.Getenv("VENDOR_URLS")),
		DB: DBConfig{
			URI:      strings.TrimSpace(os.Getenv("MONGODB_URI")),
			Database: strings.TrimSpace(os.Getenv("MONGODB_DATABASE")),
		},
	}

	var missing []string
	if cfg.Port == "" {
		missing = append(missing, "PORT")
	}
	if cfg.DB.URI == "" {
		missing = append(missing, "MONGODB_URI")
	}
	if cfg.DB.Database == "" {
		missing = append(missing, "MONGODB_DATABASE")
	}
	if cfg.VendorURLs == "" {
		missing = append(missing, "VENDOR_URLS")
	}

	if len(missing) > 0 {
		return AppConfig{}, fmt.Errorf("missing required environment variables: %s", strings.Join(missing, ", "))
	}

	if err := validateVendorURLs(cfg.VendorURLs); err != nil {
		return AppConfig{}, err
	}

	return cfg, nil
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

func ConnectMongoRepository(ctx context.Context, cfg AppConfig) (*persistence.MongoRepository, *mongo.Client, error) {
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.DB.URI))
	if err != nil {
		return nil, nil, err
	}

	repo := persistence.NewMongoRepository(client.Database(cfg.DB.Database))
	if err := repo.EnsureSchema(ctx); err != nil {
		_ = client.Disconnect(ctx)
		return nil, nil, err
	}

	return repo, client, nil
}
