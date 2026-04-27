package notifier

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestEmailAPISender_SendRegistersTemplatesAndSends(t *testing.T) {
	var mu sync.Mutex
	putCalls := 0
	postTemplateCalls := 0
	sendCalls := 0
	receivedSendBody := map[string]any{}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()

		switch {
		case r.Method == http.MethodPut && strings.HasPrefix(r.URL.Path, "/v2/email/templates/"):
			putCalls++
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"error":"not found"}`))
		case r.Method == http.MethodPost && r.URL.Path == "/v2/email/templates":
			postTemplateCalls++
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"ok":true}`))
		case r.Method == http.MethodPost && r.URL.Path == "/v2/email/send":
			sendCalls++
			if err := json.NewDecoder(r.Body).Decode(&receivedSendBody); err != nil {
				t.Fatalf("decode send body: %v", err)
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"ok":true}`))
		default:
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
	}))
	defer srv.Close()

	sender := NewEmailAPISender(srv.URL, "noreply@branchly.com", 2*time.Second)
	err := sender.Send(context.Background(), "job_completed", "user@example.com", testData())
	if err != nil {
		t.Fatalf("send failed: %v", err)
	}

	if putCalls != 3 {
		t.Fatalf("expected 3 PUT template calls, got %d", putCalls)
	}
	if postTemplateCalls != 3 {
		t.Fatalf("expected 3 POST template calls, got %d", postTemplateCalls)
	}
	if sendCalls != 1 {
		t.Fatalf("expected 1 send call, got %d", sendCalls)
	}
	if receivedSendBody["templateSlug"] != TemplateSlugJobCompleted {
		t.Fatalf("unexpected templateSlug: %v", receivedSendBody["templateSlug"])
	}
	if receivedSendBody["to"] != "user@example.com" {
		t.Fatalf("unexpected to: %v", receivedSendBody["to"])
	}
}

func TestEmailAPISender_SendSkipsUnknownEvent(t *testing.T) {
	sender := NewEmailAPISender("http://example.com", "noreply@branchly.com", 2*time.Second)
	err := sender.Send(context.Background(), "unknown_event", "user@example.com", testData())
	if err != nil {
		t.Fatalf("expected nil for unknown event, got %v", err)
	}
}

func TestEmailAPISender_SendReturnsErrorOnSendFailure(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPut && strings.HasPrefix(r.URL.Path, "/v2/email/templates/"):
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodPost && r.URL.Path == "/v2/email/send":
			w.WriteHeader(http.StatusBadGateway)
			_, _ = w.Write([]byte("upstream error"))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	sender := NewEmailAPISender(srv.URL, "noreply@branchly.com", 2*time.Second)
	err := sender.Send(context.Background(), "job_failed", "user@example.com", testData())
	if err == nil {
		t.Fatalf("expected send error, got nil")
	}
}
