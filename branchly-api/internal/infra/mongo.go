package infra

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

func ConnectMongo(ctx context.Context, uri string) (*mongo.Client, error) {
	connectCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	client, err := mongo.Connect(connectCtx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, fmt.Errorf("infra/mongo: connect: %w", err)
	}
	pingCtx, pingCancel := context.WithTimeout(ctx, 5*time.Second)
	defer pingCancel()
	if err := client.Ping(pingCtx, readpref.Primary()); err != nil {
		_ = client.Disconnect(context.Background())
		return nil, fmt.Errorf("infra/mongo: ping: %w", err)
	}
	return client, nil
}

func EnsureIndexes(ctx context.Context, db *mongo.Database) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	users := db.Collection("users")
	_, err := users.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{Key: "provider", Value: 1},
			{Key: "provider_id", Value: 1},
		},
		Options: options.Index().SetUnique(true).SetName("uniq_provider_provider_id"),
	})
	if err != nil {
		return fmt.Errorf("infra/mongo: users index: %w", err)
	}

	repos := db.Collection("repositories")
	_, err = repos.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "user_id", Value: 1}},
		Options: options.Index().SetName("idx_repositories_user_id"),
	})
	if err != nil {
		return fmt.Errorf("infra/mongo: repositories user_id index: %w", err)
	}
	_, err = repos.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{Key: "user_id", Value: 1},
			{Key: "github_repo_id", Value: 1},
		},
		Options: options.Index().SetUnique(true).SetName("uniq_user_github_repo"),
	})
	if err != nil {
		return fmt.Errorf("infra/mongo: repositories uniq index: %w", err)
	}

	jobs := db.Collection("jobs")
	for _, model := range []mongo.IndexModel{
		{Keys: bson.D{{Key: "user_id", Value: 1}}, Options: options.Index().SetName("idx_jobs_user_id")},
		{Keys: bson.D{{Key: "repository_id", Value: 1}}, Options: options.Index().SetName("idx_jobs_repository_id")},
		{Keys: bson.D{{Key: "status", Value: 1}}, Options: options.Index().SetName("idx_jobs_status")},
	} {
		if _, err := jobs.Indexes().CreateOne(ctx, model); err != nil {
			return fmt.Errorf("infra/mongo: jobs index: %w", err)
		}
	}

	slog.Info("mongo indexes ensured")
	return nil
}
