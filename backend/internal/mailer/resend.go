package mailer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"yardpass/internal/config"
)

const defaultResendAPIURL = "https://api.resend.com/emails"

type ResendMailer struct {
	apiKey string
	from   string
	apiURL string
	client *http.Client
}

func NewResendMailer(cfg *config.Config) *ResendMailer {
	return &ResendMailer{
		apiKey: cfg.Resend.APIKey,
		from:   cfg.SMTP.From,
		apiURL: defaultResendAPIURL,
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

func (m *ResendMailer) Send(to, subject, body string) error {
	if m.apiKey == "" {
		return fmt.Errorf("resend is not configured: set RESEND_API_KEY")
	}
	if m.from == "" {
		return fmt.Errorf("resend is not configured: set SMTP_FROM (e.g. onboarding@resend.dev)")
	}

	payload, err := json.Marshal(map[string]any{
		"from":    m.from,
		"to":      []string{to},
		"subject": subject,
		"text":    body,
	})
	if err != nil {
		return fmt.Errorf("marshal resend payload: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, m.apiURL, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("create resend request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+m.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := m.client.Do(req)
	if err != nil {
		return fmt.Errorf("resend request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}

	return fmt.Errorf("resend API %d: %s", resp.StatusCode, bytes.TrimSpace(respBody))
}
