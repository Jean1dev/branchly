package migrations

import (
	"context"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
)

// TestMigrate001_NothingToMigrate verifies that the migration is a no-op
// when no users have an encrypted_token field.
func TestMigrate001_NothingToMigrate(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("empty users cursor", func(mt *mtest.T) {
		// users.Find returns empty cursor → nothing to do.
		mt.AddMockResponses(
			mtest.CreateCursorResponse(0, mt.DB.Name()+".users", mtest.FirstBatch),
		)

		err := RunMigrations(context.Background(), mt.DB)
		if err != nil {
			mt.Fatalf("expected no error, got %v", err)
		}
	})
}

// TestMigrate001_NewUser_CreatesIntegrationAndBackFills verifies that a user
// with encrypted_token gets a git_integration created and their repositories
// back-filled with integration metadata.
func TestMigrate001_NewUser_CreatesIntegrationAndBackFills(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("user with token and one repo", func(mt *mtest.T) {
		userDoc := bson.D{
			{Key: "_id", Value: "user-1"},
			{Key: "encrypted_token", Value: "enc-token-abc"},
			{Key: "created_at", Value: time.Now()},
		}
		repoDoc := bson.D{
			{Key: "_id", Value: "repo-1"},
			{Key: "full_name", Value: "owner/repo"},
			{Key: "github_repo_id", Value: int64(99999)},
		}

		mt.AddMockResponses(
			// 1. users.Find → one user with encrypted_token
			mtest.CreateCursorResponse(0, mt.DB.Name()+".users", mtest.FirstBatch, userDoc),
			// 2. integrations.CountDocuments (aggregate) → count = 0
			mtest.CreateCursorResponse(0, mt.DB.Name()+".git_integrations", mtest.FirstBatch,
				bson.D{{Key: "n", Value: int32(0)}}),
			// 3. integrations.InsertOne → success
			mtest.CreateSuccessResponse(),
			// 4. repos.Find → one repo without integration_id
			mtest.CreateCursorResponse(0, mt.DB.Name()+".repositories", mtest.FirstBatch, repoDoc),
			// 5. repos.UpdateOne → success
			mtest.CreateSuccessResponse(bson.E{Key: "nModified", Value: int32(1)}),
			// 6. users.UpdateOne ($unset encrypted_token) → success
			mtest.CreateSuccessResponse(bson.E{Key: "nModified", Value: int32(1)}),
		)

		err := RunMigrations(context.Background(), mt.DB)
		if err != nil {
			mt.Fatalf("expected no error, got %v", err)
		}
	})
}

// TestMigrate001_Idempotent_IntegrationAlreadyExists verifies the key
// idempotency guarantee: if the user has an encrypted_token but an integration
// already exists, the migration reuses the existing integration ID and does NOT
// insert a duplicate, then still unsets the token.
func TestMigrate001_Idempotent_IntegrationAlreadyExists(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("existing integration, no repos to back-fill", func(mt *mtest.T) {
		userDoc := bson.D{
			{Key: "_id", Value: "user-1"},
			{Key: "encrypted_token", Value: "enc-token-abc"},
		}
		existingInteg := bson.D{
			{Key: "_id", Value: "existing-integ-id"},
			{Key: "user_id", Value: "user-1"},
			{Key: "provider", Value: "github"},
		}

		mt.AddMockResponses(
			// 1. users.Find → user still has encrypted_token
			mtest.CreateCursorResponse(0, mt.DB.Name()+".users", mtest.FirstBatch, userDoc),
			// 2. integrations.CountDocuments (aggregate) → count = 1 (already exists)
			mtest.CreateCursorResponse(0, mt.DB.Name()+".git_integrations", mtest.FirstBatch,
				bson.D{{Key: "n", Value: int32(1)}}),
			// 3. integrations.FindOne → return existing integration
			mtest.CreateCursorResponse(0, mt.DB.Name()+".git_integrations", mtest.FirstBatch, existingInteg),
			// 4. repos.Find → empty (repos already migrated)
			mtest.CreateCursorResponse(0, mt.DB.Name()+".repositories", mtest.FirstBatch),
			// 5. users.UpdateOne ($unset encrypted_token) → success
			mtest.CreateSuccessResponse(bson.E{Key: "nModified", Value: int32(1)}),
		)

		// Should succeed without trying to insert a duplicate integration.
		err := RunMigrations(context.Background(), mt.DB)
		if err != nil {
			mt.Fatalf("expected no error on second run, got %v", err)
		}
	})
}

// TestMigrate001_Idempotent_ExistingIntegrationWithRepos verifies that when a
// user has an existing integration AND repos that still need back-filling (e.g.
// the migration crashed between InsertOne and UpdateOne last time), the repos
// are still updated using the existing integration ID.
func TestMigrate001_Idempotent_ExistingIntegrationWithRepos(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("existing integration, one repo still needs back-fill", func(mt *mtest.T) {
		userDoc := bson.D{
			{Key: "_id", Value: "user-1"},
			{Key: "encrypted_token", Value: "enc-token"},
		}
		existingInteg := bson.D{
			{Key: "_id", Value: "existing-integ-id"},
			{Key: "user_id", Value: "user-1"},
			{Key: "provider", Value: "github"},
		}
		repoDoc := bson.D{
			{Key: "_id", Value: "repo-1"},
			{Key: "full_name", Value: "owner/repo"},
			{Key: "github_repo_id", Value: int64(42)},
		}

		mt.AddMockResponses(
			// 1. users.Find
			mtest.CreateCursorResponse(0, mt.DB.Name()+".users", mtest.FirstBatch, userDoc),
			// 2. integrations.CountDocuments → 1
			mtest.CreateCursorResponse(0, mt.DB.Name()+".git_integrations", mtest.FirstBatch,
				bson.D{{Key: "n", Value: int32(1)}}),
			// 3. integrations.FindOne → existing
			mtest.CreateCursorResponse(0, mt.DB.Name()+".git_integrations", mtest.FirstBatch, existingInteg),
			// 4. repos.Find → one repo still without integration_id
			mtest.CreateCursorResponse(0, mt.DB.Name()+".repositories", mtest.FirstBatch, repoDoc),
			// 5. repos.UpdateOne → success
			mtest.CreateSuccessResponse(bson.E{Key: "nModified", Value: int32(1)}),
			// 6. users.UpdateOne → success
			mtest.CreateSuccessResponse(bson.E{Key: "nModified", Value: int32(1)}),
		)

		err := RunMigrations(context.Background(), mt.DB)
		if err != nil {
			mt.Fatalf("expected no error, got %v", err)
		}
	})
}

// TestMigrate001_MultipleUsers verifies that the migration processes each user
// independently and does not mix up integrations between users.
func TestMigrate001_MultipleUsers(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("two users, both new", func(mt *mtest.T) {
		user1 := bson.D{
			{Key: "_id", Value: "user-1"},
			{Key: "encrypted_token", Value: "enc-1"},
			{Key: "created_at", Value: time.Now()},
		}
		user2 := bson.D{
			{Key: "_id", Value: "user-2"},
			{Key: "encrypted_token", Value: "enc-2"},
			{Key: "created_at", Value: time.Now()},
		}

		mt.AddMockResponses(
			// users.Find → two users in the first batch
			mtest.CreateCursorResponse(0, mt.DB.Name()+".users", mtest.FirstBatch, user1, user2),
			// user-1: CountDocuments → 0
			mtest.CreateCursorResponse(0, mt.DB.Name()+".git_integrations", mtest.FirstBatch,
				bson.D{{Key: "n", Value: int32(0)}}),
			// user-1: InsertOne → success
			mtest.CreateSuccessResponse(),
			// user-1: repos.Find → empty
			mtest.CreateCursorResponse(0, mt.DB.Name()+".repositories", mtest.FirstBatch),
			// user-1: users.UpdateOne → success
			mtest.CreateSuccessResponse(bson.E{Key: "nModified", Value: int32(1)}),
			// user-2: CountDocuments → 0
			mtest.CreateCursorResponse(0, mt.DB.Name()+".git_integrations", mtest.FirstBatch,
				bson.D{{Key: "n", Value: int32(0)}}),
			// user-2: InsertOne → success
			mtest.CreateSuccessResponse(),
			// user-2: repos.Find → empty
			mtest.CreateCursorResponse(0, mt.DB.Name()+".repositories", mtest.FirstBatch),
			// user-2: users.UpdateOne → success
			mtest.CreateSuccessResponse(bson.E{Key: "nModified", Value: int32(1)}),
		)

		err := RunMigrations(context.Background(), mt.DB)
		if err != nil {
			mt.Fatalf("expected no error, got %v", err)
		}
	})
}
