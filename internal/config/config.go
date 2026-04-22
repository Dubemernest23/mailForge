package config

import (
	"os"
	"strconv"
)

// Config is the configuration struct for the application.
// Config, EmailConfig, JwtConfig, DatabaseConfig, and ServerConfig are all used to hold the configuration values for the application.

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Jwt      JwtConfig
	Email    EmailConfig
	DB       DBConfig
}

type EmailConfig struct {
	SmtpHost     string
	SmtpPort     int
	SmtpUser     string
	SmtpPassword string
	SmtpFrom     string
}

type JwtConfig struct {
	JwtSecret string
	JwtExpiry string
}

type DBConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Name     string
}

type DatabaseConfig struct {
	DSN string
}

type ServerConfig struct {
	AppEnv  string
	AppPort string
	AppName string
}

func NewInitConfig() *Config {
	return &Config{
		Server: ServerConfig{
			AppEnv:  getENV("APP_ENV", "development"),
			AppPort: getENV("APP_PORT", "3010"),
			AppName: getENV("APP_NAME", "MailForge"),
		},
		Database: DatabaseConfig{
			DSN: getENV("DB_DSN", ""),
		},
		Email: EmailConfig{
			SmtpHost: getENV("SMTP_HOST", "smtp.gmail.com"),
			SmtpPort: func() int {
				port, err := strconv.Atoi(getENV("SMTP_PORT", "587"))
				if err != nil {
					return 587
				}
				return port
			}(),
			SmtpUser:     getENV("SMTP_USER", "your-email@gmail.com"),
			SmtpPassword: getENV("SMTP_PASSWORD", "your-app-password"),
			SmtpFrom:     getENV("SMTP_FROM", "noreply@mailforge.com"),
		},
		Jwt: JwtConfig{
			JwtSecret: getENV("JWT_SECRET", "your_jwt_secret"),
			JwtExpiry: getENV("JWT_EXPIRY", "24h"),
		},
		DB: DBConfig{
			Host:     getENV("DB_HOST", "localhost"),
			Port:     func() int { port, _ := strconv.Atoi(getENV("DB_PORT", "3306")); return port }(),
			User:     getENV("DB_USER", "root"),
			Password: getENV("DB_PASSWORD", ""),
			Name:     getENV("DB_NAME", "mailforge_db"),
		},
	}
}

func getENV(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
