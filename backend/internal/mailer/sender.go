package mailer

import (
	"yardpass/internal/config"
	"yardpass/internal/service"
)

// NewEmailSender uses Resend HTTPS API when RESEND_API_KEY is set (Railway Hobby blocks SMTP).
// Otherwise falls back to SMTP (local MailHog, Pro plan, etc.).
func NewEmailSender(cfg *config.Config) service.EmailSender {
	if cfg.Resend.APIKey != "" {
		return NewResendMailer(cfg)
	}
	return NewSMTPMailer(cfg)
}
