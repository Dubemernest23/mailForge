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
	JwtSecret string // we will change the jwt struct since we are moving to RS256
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
			JwtSecret: getENV("JWT_SECRET", "your_jwt_secret"),
			JwtExpiry: getENV("JWT_EXPIRY", "24h"),
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
