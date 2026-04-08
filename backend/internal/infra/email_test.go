package infra

import (
	"os"
	"strings"
	"testing"
)

func TestNewEmailSender_NoopWhenUnconfigured(t *testing.T) {
	os.Unsetenv("SMTP_HOST")
	s := NewEmailSender(nil)
	if s.Enabled() {
		t.Fatal("expected no-op mode when SMTP_HOST is unset")
	}
	// Send should succeed silently in no-op mode.
	if err := s.Send("user@example.com", "Test", "<p>hello</p>"); err != nil {
		t.Fatalf("no-op send returned error: %v", err)
	}
}

func TestNewEmailSender_EnabledWhenConfigured(t *testing.T) {
	t.Setenv("SMTP_HOST", "smtp.example.com")
	t.Setenv("SMTP_PORT", "465")
	t.Setenv("SMTP_FROM", "alerts@ezistock.vn")
	s := NewEmailSender(nil)
	if !s.Enabled() {
		t.Fatal("expected enabled when SMTP_HOST is set")
	}
	if s.from != "alerts@ezistock.vn" {
		t.Fatalf("expected from=alerts@ezistock.vn, got %s", s.from)
	}
	if s.port != "465" {
		t.Fatalf("expected port=465, got %s", s.port)
	}
}

func TestNewEmailSender_Defaults(t *testing.T) {
	t.Setenv("SMTP_HOST", "localhost")
	os.Unsetenv("SMTP_PORT")
	os.Unsetenv("SMTP_FROM")
	s := NewEmailSender(nil)
	if s.port != "587" {
		t.Fatalf("expected default port 587, got %s", s.port)
	}
	if s.from != "noreply@ezistock.vn" {
		t.Fatalf("expected default from noreply@ezistock.vn, got %s", s.from)
	}
}

func TestBuildMIME(t *testing.T) {
	msg := buildMIME("from@x.com", "to@y.com", "Hello", "<p>body</p>")
	if !strings.Contains(msg, "From: from@x.com\r\n") {
		t.Error("missing From header")
	}
	if !strings.Contains(msg, "To: to@y.com\r\n") {
		t.Error("missing To header")
	}
	if !strings.Contains(msg, "Subject: Hello\r\n") {
		t.Error("missing Subject header")
	}
	if !strings.Contains(msg, "Content-Type: text/html; charset=\"UTF-8\"") {
		t.Error("missing Content-Type header")
	}
	if !strings.Contains(msg, "<p>body</p>") {
		t.Error("missing body content")
	}
}

func TestWrapNotificationHTML(t *testing.T) {
	html := WrapNotificationHTML("Price Alert", "<p>SSI crossed 30,000 VND</p>")
	if !strings.Contains(html, "Price Alert") {
		t.Error("title not in output")
	}
	if !strings.Contains(html, "SSI crossed 30,000 VND") {
		t.Error("body content not in output")
	}
	if !strings.Contains(html, "EziStock") {
		t.Error("brand name not in output")
	}
	if !strings.HasPrefix(html, "<!DOCTYPE html>") {
		t.Error("expected valid HTML doctype")
	}
}
