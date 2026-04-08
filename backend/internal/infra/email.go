// Package infra provides infrastructure adapters including email delivery.
//
// email.go implements a configurable email sender with:
//   - SMTP transport for dev/staging (net/smtp)
//   - No-op mode when SMTP is not configured (logs instead of sending)
//   - HTML template wrapper for notification emails
//
// Configurable via environment variables:
//   - SMTP_HOST: SMTP server hostname (required to enable email)
//   - SMTP_PORT: SMTP server port (default: 587)
//   - SMTP_USER: SMTP authentication username
//   - SMTP_PASS: SMTP authentication password
//   - SMTP_FROM: sender address (default: noreply@ezistock.vn)
//
// Requirements: 42.7
package infra

import (
	"fmt"
	"log/slog"
	"net/smtp"
	"os"
	"strings"
)

// EmailSender sends notification emails.
type EmailSender struct {
	host    string
	port    string
	user    string
	pass    string
	from    string
	enabled bool
	logger  *slog.Logger
}

// NewEmailSender creates an EmailSender from environment variables.
// When SMTP_HOST is empty the sender operates in no-op mode, logging
// each send attempt instead of delivering mail.
func NewEmailSender(logger *slog.Logger) *EmailSender {
	if logger == nil {
		logger = slog.Default()
	}

	host := os.Getenv("SMTP_HOST")
	port := os.Getenv("SMTP_PORT")
	if port == "" {
		port = "587"
	}
	from := os.Getenv("SMTP_FROM")
	if from == "" {
		from = "noreply@ezistock.vn"
	}

	enabled := host != ""
	if enabled {
		logger.Info("email sender initialized", "host", host, "port", port, "from", from)
	} else {
		logger.Info("email sender running in no-op mode (SMTP_HOST not set)")
	}

	return &EmailSender{
		host:    host,
		port:    port,
		user:    os.Getenv("SMTP_USER"),
		pass:    os.Getenv("SMTP_PASS"),
		from:    from,
		enabled: enabled,
		logger:  logger,
	}
}

// Send delivers an email with the given subject and HTML body to the
// recipient. In no-op mode the email is logged but not sent.
func (e *EmailSender) Send(to, subject, htmlBody string) error {
	if !e.enabled {
		e.logger.Info("email send (no-op)",
			"to", to,
			"subject", subject,
			"bodyLen", len(htmlBody),
		)
		return nil
	}

	msg := buildMIME(e.from, to, subject, htmlBody)

	addr := e.host + ":" + e.port
	var auth smtp.Auth
	if e.user != "" {
		auth = smtp.PlainAuth("", e.user, e.pass, e.host)
	}

	if err := smtp.SendMail(addr, auth, e.from, []string{to}, []byte(msg)); err != nil {
		return fmt.Errorf("smtp send to %s: %w", to, err)
	}

	e.logger.Debug("email sent", "to", to, "subject", subject)
	return nil
}

// Enabled reports whether the sender is configured to deliver real emails.
func (e *EmailSender) Enabled() bool {
	return e.enabled
}

// buildMIME constructs a minimal MIME message with HTML content type.
func buildMIME(from, to, subject, htmlBody string) string {
	var b strings.Builder
	b.WriteString("From: " + from + "\r\n")
	b.WriteString("To: " + to + "\r\n")
	b.WriteString("Subject: " + subject + "\r\n")
	b.WriteString("MIME-Version: 1.0\r\n")
	b.WriteString("Content-Type: text/html; charset=\"UTF-8\"\r\n")
	b.WriteString("\r\n")
	b.WriteString(htmlBody)
	return b.String()
}

// WrapNotificationHTML wraps content in a branded HTML email template
// suitable for EziStock notification emails (price alerts, investment
// ideas, mission triggers, portfolio summaries).
func WrapNotificationHTML(title, bodyContent string) string {
	return `<!DOCTYPE html>
<html lang="vi">
<head><meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1.0"></head>
<body style="margin:0;padding:0;background:#f4f4f7;font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,sans-serif">
<table width="100%" cellpadding="0" cellspacing="0" style="background:#f4f4f7;padding:24px 0">
<tr><td align="center">
<table width="600" cellpadding="0" cellspacing="0" style="background:#ffffff;border-radius:8px;overflow:hidden">
<tr><td style="background:#1a73e8;padding:20px 24px">
  <span style="color:#ffffff;font-size:20px;font-weight:600">EziStock</span>
</td></tr>
<tr><td style="padding:24px">
  <h2 style="margin:0 0 16px;color:#1a1a2e;font-size:18px">` + title + `</h2>
  <div style="color:#333;font-size:14px;line-height:1.6">` + bodyContent + `</div>
</td></tr>
<tr><td style="padding:16px 24px;background:#f9f9fb;color:#888;font-size:12px;text-align:center">
  EziStock &mdash; Vietnamese Stock Market Intelligence
</td></tr>
</table>
</td></tr>
</table>
</body>
</html>`
}
