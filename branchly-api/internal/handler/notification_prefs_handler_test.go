package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/branchly/branchly-api/internal/domain"
	"github.com/branchly/branchly-api/internal/middleware"
	"github.com/gin-gonic/gin"
)

// ---- mock user repository ----

type mockUserRepo struct {
	user    *domain.User
	findErr error
	updated *domain.NotificationPreferences
	saveErr error
}

func (m *mockUserRepo) UpsertByProvider(_ context.Context, _ *domain.User) (*domain.User, error) {
	return m.user, nil
}

func (m *mockUserRepo) FindByID(_ context.Context, _ string) (*domain.User, error) {
	return m.user, m.findErr
}

func (m *mockUserRepo) UpdateNotificationPreferences(_ context.Context, _ string, prefs domain.NotificationPreferences) error {
	m.updated = &prefs
	return m.saveErr
}

// ---- helpers ----

func notifPrefsRouter(repo *mockUserRepo) *gin.Engine {
	gin.SetMode(gin.TestMode)
	h := NewNotificationPrefsHandler(repo)
	r := gin.New()
	r.GET("/settings/notification-preferences", func(c *gin.Context) {
		c.Set(middleware.ContextUserIDKey, "user-1")
	}, h.Get)
	r.PATCH("/settings/notification-preferences", func(c *gin.Context) {
		c.Set(middleware.ContextUserIDKey, "user-1")
	}, h.Patch)
	r.GET("/internal/users/:id/notification-preferences", h.GetInternal)
	return r
}

func aUserWithPrefs(enabled, completed, failed, prOpened bool) *domain.User {
	return &domain.User{
		ID:    "user-1",
		Email: "user@example.com",
		Name:  "Test User",
		NotificationPreferences: domain.NotificationPreferences{
			Email: domain.EmailNotificationPreferences{
				Enabled:        enabled,
				OnJobCompleted: completed,
				OnJobFailed:    failed,
				OnPROpened:     prOpened,
			},
		},
	}
}

func getPrefs(r *gin.Engine) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, "/settings/notification-preferences", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func patchPrefs(r *gin.Engine, body map[string]any) *httptest.ResponseRecorder {
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPatch, "/settings/notification-preferences", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func getInternalPrefs(r *gin.Engine, userID string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, "/internal/users/"+userID+"/notification-preferences", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

// ---- tests: Get ----

func TestGetNotifPrefs_ReturnsPreferences(t *testing.T) {
	repo := &mockUserRepo{user: aUserWithPrefs(true, true, false, true)}
	r := notifPrefsRouter(repo)
	w := getPrefs(r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d — body: %s", w.Code, w.Body.String())
	}
	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	data, ok := resp["data"].(map[string]any)
	if !ok {
		t.Fatalf("missing data key in response: %s", w.Body.String())
	}
	prefs, ok := data["notification_preferences"].(map[string]any)
	if !ok {
		t.Fatalf("missing notification_preferences in data: %s", w.Body.String())
	}
	email, ok := prefs["email"].(map[string]any)
	if !ok {
		t.Fatalf("missing email in notification_preferences: %s", w.Body.String())
	}
	if email["enabled"] != true {
		t.Errorf("expected enabled=true, got %v", email["enabled"])
	}
	if email["on_job_failed"] != false {
		t.Errorf("expected on_job_failed=false, got %v", email["on_job_failed"])
	}
}

func TestGetNotifPrefs_UserNotFound_Returns500(t *testing.T) {
	repo := &mockUserRepo{user: nil, findErr: nil}
	r := notifPrefsRouter(repo)
	w := getPrefs(r)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500 when user not found, got %d", w.Code)
	}
}

// ---- tests: Patch ----

func TestPatchNotifPrefs_MergesEnabledFlag(t *testing.T) {
	repo := &mockUserRepo{user: aUserWithPrefs(true, true, true, true)}
	r := notifPrefsRouter(repo)
	w := patchPrefs(r, map[string]any{"enabled": false})

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d — body: %s", w.Code, w.Body.String())
	}
	if repo.updated == nil {
		t.Fatal("expected UpdateNotificationPreferences to be called")
	}
	if repo.updated.Email.Enabled != false {
		t.Errorf("expected enabled=false after patch, got %v", repo.updated.Email.Enabled)
	}
	// Other fields should remain unchanged.
	if repo.updated.Email.OnJobCompleted != true {
		t.Errorf("expected on_job_completed unchanged (true), got %v", repo.updated.Email.OnJobCompleted)
	}
}

func TestPatchNotifPrefs_MergesIndividualFlag(t *testing.T) {
	repo := &mockUserRepo{user: aUserWithPrefs(true, true, true, true)}
	r := notifPrefsRouter(repo)
	w := patchPrefs(r, map[string]any{"on_job_failed": false})

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d — body: %s", w.Code, w.Body.String())
	}
	if repo.updated == nil {
		t.Fatal("expected UpdateNotificationPreferences to be called")
	}
	if repo.updated.Email.OnJobFailed != false {
		t.Errorf("expected on_job_failed=false after patch")
	}
	if repo.updated.Email.Enabled != true {
		t.Errorf("expected enabled still true after partial patch")
	}
}

func TestPatchNotifPrefs_EmptyBody_Returns400(t *testing.T) {
	repo := &mockUserRepo{user: aUserWithPrefs(true, true, true, true)}
	r := notifPrefsRouter(repo)

	req := httptest.NewRequest(http.MethodPatch, "/settings/notification-preferences", bytes.NewReader([]byte("not-json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid JSON body, got %d", w.Code)
	}
}

// ---- tests: GetInternal ----

func TestGetInternalNotifPrefs_ReturnsEmailAndPrefs(t *testing.T) {
	repo := &mockUserRepo{user: aUserWithPrefs(true, false, true, false)}
	r := notifPrefsRouter(repo)
	w := getInternalPrefs(r, "user-1")

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d — body: %s", w.Code, w.Body.String())
	}
	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	data, ok := resp["data"].(map[string]any)
	if !ok {
		t.Fatalf("missing data key: %s", w.Body.String())
	}
	if data["email"] != "user@example.com" {
		t.Errorf("expected email user@example.com, got %v", data["email"])
	}
	if data["name"] != "Test User" {
		t.Errorf("expected name Test User, got %v", data["name"])
	}
	prefs, ok := data["notification_preferences"].(map[string]any)
	if !ok {
		t.Fatalf("missing notification_preferences: %s", w.Body.String())
	}
	email := prefs["email"].(map[string]any)
	if email["on_job_completed"] != false {
		t.Errorf("expected on_job_completed=false, got %v", email["on_job_completed"])
	}
}

func TestGetInternalNotifPrefs_UserNotFound_Returns404(t *testing.T) {
	repo := &mockUserRepo{user: nil, findErr: nil}
	r := notifPrefsRouter(repo)
	w := getInternalPrefs(r, "unknown-user")

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404 for unknown user, got %d", w.Code)
	}
}
