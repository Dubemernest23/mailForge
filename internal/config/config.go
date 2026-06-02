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
	Charset  string
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
	db := DBConfig{
		Host:     getENV("DB_HOST", "localhost"),
		Port:     getEnvInt("DB_PORT", 3306),
		User:     getENV("DB_USER", "root"),
		Password: getENV("DB_PASSWORD", ""),
		Name:     getENV("DB_NAME", "mailforge_db"),
		Charset:  getENV("DB_CHARSET", "utf8mb4"),
	}

	return &Config{
		Server: ServerConfig{
			AppEnv:  getENV("APP_ENV", "development"),
			AppPort: getENV("APP_PORT", "3010"),
			AppName: getENV("APP_NAME", "MailForge"),
		},
		Database: DatabaseConfig{
			DSN: getDatabaseDSN(db),
		},
		Email: EmailConfig{
			SmtpHost:     getENV("SMTP_HOST", "smtp.gmail.com"),
			SmtpPort:     getEnvInt("SMTP_PORT", 587),
			SmtpUser:     getENV("SMTP_USER", "your-email@gmail.com"),
			SmtpPassword: getENV("SMTP_PASSWORD", "your-app-password"),
			SmtpFrom:     getENV("SMTP_FROM", "noreply@mailforge.com"),
		},
		Jwt: JwtConfig{
			JwtSecret: getENV("JWT_SECRET", "your_jwt_secret"),
			JwtExpiry: getENV("JWT_EXPIRY", "24h"),
		},
		DB: db,
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

func getDatabaseDSN(db DBConfig) string {
	if dsn := getENV("DB_DSN", ""); dsn != "" {
		return dsn
	}

	cfg := mysql.Config{
		User:                 db.User,
		Passwd:               db.Password,
		Net:                  "tcp",
		Addr:                 fmt.Sprintf("%s:%d", db.Host, db.Port),
		DBName:               db.Name,
		ParseTime:            true,
		AllowNativePasswords: true,
		Params: map[string]string{
			"charset": db.Charset,
		},
	}

	return cfg.FormatDSN()
}
