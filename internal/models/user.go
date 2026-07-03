package models

import (
	"database/sql"
	"time"

	"github.com/uptrace/bun"
)

type User struct {
	bun.BaseModel `bun:"table:users, alias:u"`

	ID                          uint64       `bun:"id,pk,autoincreament"`
	PublicId                    string       `bun:"public_id,notnull"`
	Userame                     string       `bun:"name,notnull"`
	Email                       string       `bun:"email,notnull"`
	PasswordHash                string       `bun:"password_hash,notnull"`
	Role                        string       `bun:"role,notnull,default:'user'"`
	Status                      string       `bun:"status,notnull,default:'active'"`
	FailedLoginAttempts         int          `bun:"failed_login_attempts,notnull,default:0"`
	PasswordRestToken           string       `bun:"password_reset_token,null"`
	PasswordResetTokenExpiresAt sql.NullTime `bun:"password_reset_expires_at,null"`
	LastLoginAt                 sql.NullTime `bun:"last_login_at,null"`
	CreatedAt                   time.Time    `bun:"created_at,notnull"`
	UpdatedAt                   time.Time    `bun:"updated_at,notnull"`
}
