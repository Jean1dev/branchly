package domain

import (
	"context"
	"time"
)

type User struct {
	ID             string    `bson:"_id"`
	Provider       string    `bson:"provider"`
	ProviderID     string    `bson:"provider_id"`
	Email          string    `bson:"email"`
	Name           string    `bson:"name"`
	AvatarURL      string    `bson:"avatar_url"`
	EncryptedToken string    `bson:"encrypted_token"`
	CreatedAt      time.Time `bson:"created_at"`
	UpdatedAt      time.Time `bson:"updated_at"`
}

type UserRepository interface {
	UpsertByProvider(ctx context.Context, u *User) (*User, error)
	FindByID(ctx context.Context, id string) (*User, error)
}
