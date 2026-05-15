package mailer

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"yardpass/internal/config"
)

func TestResendMailer_Send_success(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-key" {
			t.Fatalf("Authorization = %q", got)
		}
		body, _ := io.ReadAll(r.Body)
		if !strings.Contains(string(body), `"to":["user@example.com"]`) {
			t.Fatalf("body = %s", body)
		}
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	m := NewResendMailer(&config.Config{
		Resend: config.ResendConfig{APIKey: "test-key"},
		SMTP:   config.SMTPConfig{From: "onboarding@resend.dev"},
	})
	m.apiURL = srv.URL
	m.client = srv.Client()

	if err := m.Send("user@example.com", "subject", "hello"); err != nil {
		t.Fatalf("Send() error = %v", err)
	}
}

func TestResendMailer_Send_apiError(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = w.Write([]byte(`{"message":"invalid from"}`))
	}))
	t.Cleanup(srv.Close)

	m := NewResendMailer(&config.Config{
		Resend: config.ResendConfig{APIKey: "test-key"},
		SMTP:   config.SMTPConfig{From: "bad"},
	})
	m.apiURL = srv.URL
	m.client = srv.Client()

	if err := m.Send("user@example.com", "s", "b"); err == nil {
		t.Fatal("expected error")
	}
}

func TestNewEmailSender_prefersResend(t *testing.T) {
	t.Parallel()

	s := NewEmailSender(&config.Config{
		Resend: config.ResendConfig{APIKey: "re_test"},
		SMTP:   config.SMTPConfig{Host: "smtp.example.com", From: "a@b.c"},
	})
	if _, ok := s.(*ResendMailer); !ok {
		t.Fatalf("got %T, want *ResendMailer", s)
	}
}
