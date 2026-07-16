package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/go-sql-driver/mysql"
)

// Config is the configuration struct for the application.
// Config, EmailConfig, JwtConfig, DatabaseConfig, and ServerConfig are all used to hold
//  the configuration values for the application.

type Config struct {
	Server ServerConfig
	Jwt    JwtConfig
	Email  EmailConfig
	DB     DBConfig
}

type EmailConfig struct {
	SmtpHost     string
	SmtpPort     int
	SmtpUser     string
	SmtpPassword string
	SmtpFrom     string
}

type JwtConfig struct {
	PrivateKeyPath string
	PublicKeyPath  string
	AccessExpiry   string // maps to JWT_ACCESS_EXPIRY
	RefreshExpiry  string // maps to JWT_REFRESH_EXPIRY
}

type DBConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Name     string
	Charset  string
}

func (d DBConfig) DSN() string {

	cfg := mysql.Config{
		User:                 d.User,
		Passwd:               d.Password,
		Net:                  "tcp",
		Addr:                 fmt.Sprintf("%s:%d", d.Host, d.Port),
		DBName:               d.Name,
		ParseTime:            true,
		AllowNativePasswords: true,
		Params: map[string]string{
			"charset": d.Charset,
		},
	}
	return cfg.FormatDSN()
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
		Email: EmailConfig{
			SmtpHost:     getENV("SMTP_HOST", "localhost"), // was smtp.gmail.com
			SmtpPort:     getEnvInt("SMTP_PORT", 1025),     // was 587
			SmtpUser:     getENV("SMTP_USER", ""),          // was placeholder
			SmtpPassword: getENV("SMTP_PASSWORD", ""),      // was placeholder
			SmtpFrom:     getENV("SMTP_FROM", "noreply@mailforge.com"),
		},
		Jwt: JwtConfig{
			AccessExpiry:   getENV("JWT_ACCESS_EXPIRY", "1h"),
			RefreshExpiry:  getENV("JWT_REFRESH_EXPIRY", "7d"),
			PrivateKeyPath: getENV("JWT_PRIVATE_KEY_PATH", ""),
			PublicKeyPath:  getENV("JWT_PUBLIC_KEY_PATH", ""),
		},
		DB: DBConfig{
			Host:     getENV("DB_HOST", "localhost"),
			Port:     getEnvInt("DB_PORT", 3306),
			User:     getENV("DB_USER", "root"),
			Password: getENV("DB_PASSWORD", ""),
			Name:     getENV("DB_NAME", "mailforge_db"),
			Charset:  getENV("DB_CHARSET", "utf8mb4"),
		},
	}
}

func getENV(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	val := getENV(key, "")
	if val == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(val)
	if err != nil {
		return fallback
	}

	return parsed
}
