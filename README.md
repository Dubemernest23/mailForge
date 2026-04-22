# mailForge

A lightweight email campaign API built in Go using:

- `chi` router for HTTP routing
- `uber/fx` for application lifecycle and dependency injection
- `bun` ORM for database access
- Authentication support for protected endpoints

## Features

- Create, update, and manage email campaigns
- Send campaign-related requests through a REST API
- Secure endpoints with authentication
- Structured, modular Go architecture with dependency injection

## Getting Started

1. Install Go (1.22+ recommended).
2. Set up your database and configure connection settings in environment variables.
3. Copy or rename `.env.example` to `.env` and update values as needed.
4. Run the application:

```bash
go run ./cmd/mailforge
```

## Environment Variables

Copy `.env.example` to `.env` and configure the following values for your environment:

```env
# Application
APP_ENV=development
APP_PORT=8080
APP_NAME=MailForge

# MySQL Database
DB_HOST=localhost
DB_PORT=3306
DB_USER=root
DB_PASSWORD=yourpassword
DB_NAME=mailforge_db
DB_CHARSET=utf8mb4

# JWT Authentication
JWT_SECRET=supersecretkeychangethisinproduction
JWT_EXPIRY_HOURS=24

# Email Configuration (SMTP)
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=your-email@gmail.com
SMTP_PASSWORD=your-app-password
SMTP_FROM=noreply@mailforge.com

# Optional - For production email provider (Resend, SendGrid, etc.)
EMAIL_PROVIDER=smtp
RESEND_API_KEY=re_xxxxxxxxxxxxxxxxxxxx
```

## API Overview

The API exposes campaign management routes and authentication routes. Example endpoints may include:

- `POST /login`
- `GET /campaigns`
- `POST /campaigns`
- `PUT /campaigns/{id}`
- `DELETE /campaigns/{id}`

## Project Structure

```text
mailforge/
├── cmd/
│   └── api/
│       └── main.go                 # Entry point + FX bootstrap
├── internal/
│   ├── config/
│   │   └── config.go               # Config struct + loading
│   ├── di/
│   │   └── container.go            # Uber FX dependency injection
│   ├── domain/
│   │   ├── model/                  # Core business entities
│   │   │   ├── subscriber.go
│   │   │   ├── campaign.go
│   │   │   └── email_log.go
│   │   └── repository/             # Repository interfaces
│   │       ├── subscriber_repo.go
│   │       └── campaign_repo.go
│   ├── dto/                        # Request & Response DTOs
│   │   ├── subscriber.go
│   │   └── campaign.go
│   ├── handler/                    # HTTP Handlers (Chi)
│   │   ├── subscriber_handler.go
│   │   └── campaign_handler.go
│   ├── repository/                 # Bun implementations
│   │   ├── subscriber_repository.go
│   │   └── campaign_repository.go
│   ├── service/                    # Business logic
│   │   ├── subscriber_service.go
│   │   └── campaign_service.go
│   ├── routes/
│   │   └── routes.go               # Chi route definitions
│   ├── middleware/
│   │   └── logging.go
│   └── utils/                      # Shared utilities
├── pkg/
│   └── logger/
│       └── logger.go
├── migrations/                     # Database migrations (Bun)
├── scripts/                        # Useful scripts
├── .env.example
├── .gitignore
├── go.mod
├── go.sum
└── README.md
```

## Notes
This README is intentionally simple and designed as a starting point. 