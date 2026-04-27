package notifier

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

type EmailAPISender struct {
	baseURL   string
	fromEmail string
	client    *http.Client
	once      sync.Once
	onceErr   error
}

func NewEmailAPISender(baseURL, fromEmail string, timeout time.Duration) *EmailAPISender {
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	from := strings.TrimSpace(fromEmail)
	if from == "" {
		from = "noreply@branchly.com"
	}
	return &EmailAPISender{
		baseURL:   strings.TrimRight(strings.TrimSpace(baseURL), "/"),
		fromEmail: from,
		client:    &http.Client{Timeout: timeout},
	}
}

func (s *EmailAPISender) Send(ctx context.Context, event, to string, data JobNotifData) error {
	subject, slug, ok := eventSubjectAndSlug(event, data)
	if !ok {
		return nil
	}

	if err := s.ensureTemplates(ctx); err != nil {
		return err
	}

	body := map[string]any{
		"templateSlug": slug,
		"to":           to,
		"subject":      subject,
		"variables":    templateVars(data),
	}

	return s.doJSON(ctx, http.MethodPost, s.baseURL+"/v2/email/send", body, func(status int, response []byte) error {
		if status >= 200 && status < 300 {
			return nil
		}
		return fmt.Errorf("emailapi send failed (%d): %s", status, strings.TrimSpace(string(response)))
	})
}

func (s *EmailAPISender) ensureTemplates(ctx context.Context) error {
	s.once.Do(func() {
		s.onceErr = s.registerTemplates(ctx)
	})
	return s.onceErr
}

func (s *EmailAPISender) RegisterTemplates(ctx context.Context) error {
	return s.registerTemplates(ctx)
}

func (s *EmailAPISender) registerTemplates(ctx context.Context) error {
	templates, err := templateDefinitions()
	if err != nil {
		return err
	}
	for _, t := range templates {
		if err := s.upsertTemplate(ctx, t); err != nil {
			return err
		}
	}
	return nil
}

func (s *EmailAPISender) upsertTemplate(ctx context.Context, t TemplateDefinition) error {
	basePayload := map[string]any{
		"name":         t.Name,
		"fromEmail":    s.fromEmail,
		"htmlTemplate": t.HTMLTemplate,
	}

	putURL := s.baseURL + "/v2/email/templates/" + t.Slug
	putErr := s.doJSON(ctx, http.MethodPut, putURL, basePayload, func(status int, response []byte) error {
		if status >= 200 && status < 300 {
			return nil
		}
		if status == http.StatusNotFound {
			return errTemplateNotFound
		}
		return fmt.Errorf("emailapi upsert template %s failed (%d): %s", t.Slug, status, strings.TrimSpace(string(response)))
	})
	if putErr == nil {
		return nil
	}
	if putErr != errTemplateNotFound {
		return putErr
	}

	postPayload := map[string]any{
		"slug":         t.Slug,
		"name":         t.Name,
		"fromEmail":    s.fromEmail,
		"htmlTemplate": t.HTMLTemplate,
	}
	postURL := s.baseURL + "/v2/email/templates"
	return s.doJSON(ctx, http.MethodPost, postURL, postPayload, func(status int, response []byte) error {
		if status >= 200 && status < 300 {
			return nil
		}
		return fmt.Errorf("emailapi create template %s failed (%d): %s", t.Slug, status, strings.TrimSpace(string(response)))
	})
}

var errTemplateNotFound = errors.New("template not found")

func (s *EmailAPISender) doJSON(
	ctx context.Context,
	method string,
	url string,
	payload any,
	validate func(status int, response []byte) error,
) error {
	raw, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(raw))
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("request %s %s: %w", method, url, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	if err := validate(resp.StatusCode, body); err != nil {
		return err
	}
	return nil
}
