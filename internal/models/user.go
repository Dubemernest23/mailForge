package models

import (
	"database/sql"
	"time"

	"github.com/uptrace/bun"
)

type User struct {
	bun.BaseModel `bun:"table:users,alias:u"`

	ID                          uint64       `bun:"id,pk,autoincrement"`
	PublicId                    string       `bun:"public_id,notnull"`
	Username 					string 		 `bun:"username,notnull"`
	Email                       string       `bun:"email,notnull"`
	PasswordHash                string       `bun:"password_hash,notnull"`
	Role                        string       `bun:"role,notnull,default:'user'"`
	Status                      string       `bun:"status,notnull,default:'active'"`
	FailedLoginAttempts         uint32          `bun:"failed_login_attempts,notnull,default:0"`
	PasswordResetToken 			sql.NullString `bun:"password_reset_token"`
	PasswordResetTokenExpiresAt sql.NullTime `bun:"password_reset_expires_at,null"`
	LastLoginAt                 sql.NullTime `bun:"last_login_at,null"`
	CreatedAt                   time.Time    `bun:"created_at,notnull"`
	UpdatedAt                   time.Time    `bun:"updated_at,notnull"`
}
