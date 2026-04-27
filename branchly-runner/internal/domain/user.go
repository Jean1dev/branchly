package domain

import "context"

type EmailNotificationPreferences struct {
	Enabled        bool `bson:"enabled"`
	OnJobCompleted bool `bson:"on_job_completed"`
	OnJobFailed    bool `bson:"on_job_failed"`
	OnPROpened     bool `bson:"on_pr_opened"`
}

type NotificationPreferences struct {
	Email EmailNotificationPreferences `bson:"email"`
}

type User struct {
	ID                      string                  `bson:"_id"`
	Email                   string                  `bson:"email"`
	Name                    string                  `bson:"name"`
	NotificationPreferences NotificationPreferences `bson:"notification_preferences"`
}

type UserRepository interface {
	FindByID(ctx context.Context, id string) (*User, error)
}
