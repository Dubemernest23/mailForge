package config

import (
	"strings"
	"testing"
)

func TestNewInitConfigUsesExplicitDatabaseDSN(t *testing.T) {
	t.Setenv("DB_DSN", "root:secret@tcp(localhost:3306)/custom_db?parseTime=true")

	cfg := NewInitConfig()

	if cfg.Database.DSN != "root:secret@tcp(localhost:3306)/custom_db?parseTime=true" {
		t.Fatalf("expected explicit DB_DSN to be used, got %q", cfg.Database.DSN)
	}
}

func TestNewInitConfigBuildsDatabaseDSNFromParts(t *testing.T) {
	t.Setenv("DB_DSN", "")
	t.Setenv("DB_HOST", "db.local")
	t.Setenv("DB_PORT", "3307")
	t.Setenv("DB_USER", "tester")
	t.Setenv("DB_PASSWORD", "secret")
	t.Setenv("DB_NAME", "mailforge_test")
	t.Setenv("DB_CHARSET", "utf8mb4")

	cfg := NewInitConfig()

	expectedPrefix := "tester:secret@tcp(db.local:3307)/mailforge_test?"
	if !strings.HasPrefix(cfg.Database.DSN, expectedPrefix) {
		t.Fatalf("expected DSN prefix %q, got %q", expectedPrefix, cfg.Database.DSN)
	}
	if !strings.Contains(cfg.Database.DSN, "charset=utf8mb4") {
		t.Fatalf("expected DSN to include charset, got %q", cfg.Database.DSN)
	}
	if !strings.Contains(cfg.Database.DSN, "parseTime=true") {
		t.Fatalf("expected DSN to include parseTime, got %q", cfg.Database.DSN)
	}
}

func TestNewInitConfigFallsBackForInvalidIntegerValues(t *testing.T) {
	t.Setenv("DB_DSN", "")
	t.Setenv("DB_PORT", "not-a-number")
	t.Setenv("SMTP_PORT", "also-not-a-number")

	cfg := NewInitConfig()

	if cfg.DB.Port != 3306 {
		t.Fatalf("expected fallback DB port 3306, got %d", cfg.DB.Port)
	}
	if cfg.Email.SmtpPort != 587 {
		t.Fatalf("expected fallback SMTP port 587, got %d", cfg.Email.SmtpPort)
	}
}
