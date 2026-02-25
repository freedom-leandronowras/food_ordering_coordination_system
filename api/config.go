package main

import (
	"context"
	"errors"
	"os"

	persistence "food_ordering_coordination_system/internal/persistance"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type DBConfig struct {
	URI      string
	Database string
}

type AppConfig struct {
	Port string
	DB   DBConfig
}

func LoadFromEnv() (AppConfig, error) {
	cfg := AppConfig{
		Port: os.Getenv("PORT"),
		DB: DBConfig{
			URI:      os.Getenv("MONGODB_URI"),
			Database: os.Getenv("MONGODB_DATABASE"),
		},
	}

	if cfg.Port == "" {
		return AppConfig{}, errors.New("PORT is required")
	}
	if cfg.DB.URI == "" {
		return AppConfig{}, errors.New("MONGODB_URI is required")
	}
	if cfg.DB.Database == "" {
		return AppConfig{}, errors.New("MONGODB_DATABASE is required")
	}
	return cfg, nil
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
