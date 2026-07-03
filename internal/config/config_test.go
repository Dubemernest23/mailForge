package config

import (
	"strings"
	"testing"
)

func TestNewInitConfigBuildsDatabaseDSNFromParts(t *testing.T) {
	t.Setenv("DB_HOST", "db.local")
	t.Setenv("DB_PORT", "3307")
	t.Setenv("DB_USER", "tester")
	t.Setenv("DB_PASSWORD", "secret")
	t.Setenv("DB_NAME", "mailforge_test")
	t.Setenv("DB_CHARSET", "utf8mb4")

	cfg := NewInitConfig()

	expectedPrefix := "tester:secret@tcp(db.local:3307)/mailforge_test?"
	if !strings.HasPrefix(cfg.DB.DSN(), expectedPrefix) {
		t.Fatalf("expected DSN prefix %q, got %q", expectedPrefix, cfg.DB.DSN())
	}
	if !strings.Contains(cfg.DB.DSN(), "charset=utf8mb4") {
		t.Fatalf("expected DSN to include charset, got %q", cfg.DB.DSN())
	}
	if !strings.Contains(cfg.DB.DSN(), "parseTime=true") {
		t.Fatalf("expected DSN to include parseTime, got %q", cfg.DB.DSN())
	}
}

func TestNewInitConfigFallsBackForInvalidIntegerValues(t *testing.T) {
	t.Setenv("DB_PORT", "not-a-number")
	t.Setenv("SMTP_PORT", "also-not-a-number")

	cfg := NewInitConfig()

	if cfg.DB.Port != 3306 {
		t.Fatalf("expected fallback DB port 3306, got %d", cfg.DB.Port)
	}
	if cfg.Email.SmtpPort != 1025 {
		t.Fatalf("expected fallback SMTP port 1025, got %d", cfg.Email.SmtpPort)
	}
}
