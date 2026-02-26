package notification

import (
	"context"
	"fmt"
	"net/smtp"
	"strings"
)

type SMTPSender struct {
	addr      string
	auth      smtp.Auth
	fromEmail string
	fromName  string
}

func NewSMTPSender(host string, port int, username, password, fromEmail, fromName string) *SMTPSender {
	addr := fmt.Sprintf("%s:%d", strings.TrimSpace(host), port)

	var auth smtp.Auth
	if strings.TrimSpace(username) != "" {
		auth = smtp.PlainAuth("", username, password, strings.TrimSpace(host))
	}

	return &SMTPSender{
		addr:      addr,
		auth:      auth,
		fromEmail: strings.TrimSpace(fromEmail),
		fromName:  strings.TrimSpace(fromName),
	}
}

func (s *SMTPSender) Send(_ context.Context, to, subject, body string) error {
	to = strings.TrimSpace(to)
	if to == "" {
		return fmt.Errorf("email recipient is required")
	}

	from := s.fromEmail
	if s.fromName != "" {
		from = fmt.Sprintf("%s <%s>", s.fromName, s.fromEmail)
	}

	msg := buildMessage(from, to, subject, body)
	if err := smtp.SendMail(s.addr, s.auth, s.fromEmail, []string{to}, []byte(msg)); err != nil {
		return fmt.Errorf("send smtp email: %w", err)
	}
	return nil
}

func buildMessage(from, to, subject, body string) string {
	body = strings.ReplaceAll(body, "\r\n", "\n")
	body = strings.ReplaceAll(body, "\r", "\n")
	body = strings.ReplaceAll(body, "\n", "\r\n")

	return strings.Join([]string{
		"From: " + from,
		"To: " + to,
		"Subject: " + subject,
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=UTF-8",
		"",
		body,
	}, "\r\n")
}
