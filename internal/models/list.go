package models

import (
	"time"

	"github.com/uptrace/bun"
)

type List struct {
	bun.BaseModel `bun:"table:lists, alias:l"`

	ID          uint64    `bun:"id,pk,autoincrement"`
	PublicID    string    `bun:"public_id,notnull"`
	UserID      uint64    `bun:"user_id,notnull"`
	Name        string    `bun:"name,notnull"`
	Description string    `bun:"description"`
	Status      string    `bun:"status,notnull,default:'active'"`
	CreatedAt   time.Time `bun:"created_at,notnull"`
	UpdatedAt   time.Time `bun:"updated_at,notnull"`

	User *User `bun:"rel:belongs-to,join:user_id=id"`
}
