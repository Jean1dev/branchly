package domain

import (
	"context"
	"time"
)

type EmailNotificationPreferences struct {
	Enabled       bool `bson:"enabled"        json:"enabled"`
	OnJobCompleted bool `bson:"on_job_completed" json:"on_job_completed"`
	OnJobFailed   bool `bson:"on_job_failed"  json:"on_job_failed"`
	OnPROpened    bool `bson:"on_pr_opened"   json:"on_pr_opened"`
}

type NotificationPreferences struct {
	Email EmailNotificationPreferences `bson:"email" json:"email"`
}

func DefaultNotificationPreferences() NotificationPreferences {
	return NotificationPreferences{
		Email: EmailNotificationPreferences{
			Enabled:        true,
			OnJobCompleted: true,
			OnJobFailed:    true,
			OnPROpened:     true,
		},
	}
}

type User struct {
	ID                      string                  `bson:"_id"`
	Provider                string                  `bson:"provider"`
	ProviderID              string                  `bson:"provider_id"`
	Email                   string                  `bson:"email"`
	Name                    string                  `bson:"name"`
	AvatarURL               string                  `bson:"avatar_url"`
	EncryptedToken          string                  `bson:"encrypted_token"`
	NotificationPreferences NotificationPreferences `bson:"notification_preferences"`
	CreatedAt               time.Time               `bson:"created_at"`
	UpdatedAt               time.Time               `bson:"updated_at"`
}

type UserRepository interface {
	UpsertByProvider(ctx context.Context, u *User) (*User, error)
	FindByID(ctx context.Context, id string) (*User, error)
	UpdateNotificationPreferences(ctx context.Context, id string, prefs NotificationPreferences) error
}
