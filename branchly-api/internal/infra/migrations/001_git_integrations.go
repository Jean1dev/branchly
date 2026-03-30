package migrations

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// RunMigrations runs all pending migrations idempotently.
// Must be called before the server starts accepting requests.
func RunMigrations(ctx context.Context, db *mongo.Database) error {
	if err := migrate001GitIntegrations(ctx, db); err != nil {
		return fmt.Errorf("migrations: 001_git_integrations: %w", err)
	}
	return nil
}

// migrate001GitIntegrations moves GitHub OAuth tokens from the users collection
// into the git_integrations collection and back-fills repositories with provider
// metadata.
func migrate001GitIntegrations(ctx context.Context, db *mongo.Database) error {
	users := db.Collection("users")
	integrations := db.Collection("git_integrations")
	repos := db.Collection("repositories")

	// Find all users that still have an encrypted_token (not yet migrated).
	cur, err := users.Find(ctx, bson.M{"encrypted_token": bson.M{"$exists": true, "$ne": ""}})
	if err != nil {
		return fmt.Errorf("find users with token: %w", err)
	}
	defer cur.Close(ctx)

	type userDoc struct {
		ID             string    `bson:"_id"`
		EncryptedToken string    `bson:"encrypted_token"`
		CreatedAt      time.Time `bson:"created_at"`
	}

	migrated := 0
	for cur.Next(ctx) {
		var u userDoc
		if err := cur.Decode(&u); err != nil {
			return fmt.Errorf("decode user: %w", err)
		}

		// Idempotency: skip if integration already exists for this user/GitHub.
		count, err := integrations.CountDocuments(ctx, bson.M{
			"user_id":  u.ID,
			"provider": "github",
		})
		if err != nil {
			return fmt.Errorf("count integrations for user %s: %w", u.ID, err)
		}

		var integrationID string
		if count == 0 {
			// Create new git_integration for this user's GitHub token.
			integrationID = uuid.New().String()
			connectedAt := u.CreatedAt
			if connectedAt.IsZero() {
				connectedAt = time.Now().UTC()
			}
			_, err = integrations.InsertOne(ctx, bson.M{
				"_id":             integrationID,
				"user_id":         u.ID,
				"provider":        "github",
				"encrypted_token": u.EncryptedToken,
				"token_type":      "oauth",
				"scopes":          []string{"repo"},
				"connected_at":    connectedAt,
			})
			if err != nil {
				return fmt.Errorf("insert integration for user %s: %w", u.ID, err)
			}
			slog.Info("migration 001: created git_integration", "user_id", u.ID)
		} else {
			// Fetch existing integration ID for repo back-fill.
			var existing struct {
				ID string `bson:"_id"`
			}
			if err := integrations.FindOne(ctx, bson.M{
				"user_id":  u.ID,
				"provider": "github",
			}).Decode(&existing); err != nil {
				return fmt.Errorf("find existing integration for user %s: %w", u.ID, err)
			}
			integrationID = existing.ID
		}

		// Back-fill repositories that lack integration_id.
		repoCur, err := repos.Find(ctx, bson.M{
			"user_id":        u.ID,
			"integration_id": bson.M{"$exists": false},
		})
		if err != nil {
			return fmt.Errorf("find repos for user %s: %w", u.ID, err)
		}

		type repoDoc struct {
			ID           string `bson:"_id"`
			FullName     string `bson:"full_name"`
			GithubRepoID int64  `bson:"github_repo_id"`
		}

		for repoCur.Next(ctx) {
			var r repoDoc
			if err := repoCur.Decode(&r); err != nil {
				repoCur.Close(ctx)
				return fmt.Errorf("decode repo: %w", err)
			}
			cloneURL := "https://github.com/" + r.FullName + ".git"
			externalID := fmt.Sprintf("%d", r.GithubRepoID)
			_, err = repos.UpdateOne(ctx,
				bson.M{"_id": r.ID},
				bson.M{"$set": bson.M{
					"integration_id": integrationID,
					"provider":       "github",
					"external_id":    externalID,
					"clone_url":      cloneURL,
				}},
				options.Update(),
			)
			if err != nil {
				repoCur.Close(ctx)
				return fmt.Errorf("update repo %s: %w", r.ID, err)
			}
		}
		if err := repoCur.Err(); err != nil {
			return fmt.Errorf("repos cursor for user %s: %w", u.ID, err)
		}
		repoCur.Close(ctx)

		// Remove encrypted_token from user document.
		_, err = users.UpdateOne(ctx,
			bson.M{"_id": u.ID},
			bson.M{"$unset": bson.M{"encrypted_token": ""}},
		)
		if err != nil {
			return fmt.Errorf("unset token for user %s: %w", u.ID, err)
		}

		migrated++
	}

	if err := cur.Err(); err != nil {
		return fmt.Errorf("users cursor: %w", err)
	}

	if migrated > 0 {
		slog.Info("migration 001: complete", "users_migrated", migrated)
	} else {
		slog.Info("migration 001: nothing to migrate (already up to date)")
	}
	return nil
}
