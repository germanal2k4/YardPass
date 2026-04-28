package mailer

import (
	"fmt"
	"net/smtp"
	"strings"

	"yardpass/internal/config"
)

type SMTPMailer struct {
	host     string
	port     string
	username string
	password string
	from     string
}

func NewSMTPMailer(cfg *config.Config) *SMTPMailer {
	return &SMTPMailer{
		host:     cfg.SMTP.Host,
		port:     cfg.SMTP.Port,
		username: cfg.SMTP.Username,
		password: cfg.SMTP.Password,
		from:     cfg.SMTP.From,
	}
}

func (m *SMTPMailer) Send(to, subject, body string) error {
	if m.host == "" || m.from == "" {
		return fmt.Errorf("smtp is not configured")
	}

	addr := fmt.Sprintf("%s:%s", m.host, m.port)
	message := strings.Join([]string{
		fmt.Sprintf("From: %s", m.from),
		fmt.Sprintf("To: %s", to),
		fmt.Sprintf("Subject: %s", subject),
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=UTF-8",
		"",
		body,
	}, "\r\n")

	var auth smtp.Auth
	if m.username != "" {
		auth = smtp.PlainAuth("", m.username, m.password, m.host)
	}

	return smtp.SendMail(addr, auth, m.from, []string{to}, []byte(message))
}
