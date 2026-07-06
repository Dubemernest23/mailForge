package models

import (
	"time"
	"database/sql"

	"github.com/uptrace/bun"
)

type Subscriber struct {
	bun.BaseModel `bun:"table:subscribers,alias:s"`

	ID        uint64    `bun:"id,pk,autoincrement"`
	PublicID  string    `bun:"public_id,notnull"`
	UserID    uint64    `bun:"user_id,notnull"`
	Name      sql.NullString    `bun:"name"`
	Email     string    `bun:"email,notnull"`
	Status    string    `bun:"status,notnull,default:'subscribed'"`
	CreatedAt time.Time `bun:"created_at,notnull"`
	UpdatedAt time.Time `bun:"updated_at,notnull"`

	User *User `bun:"rel:belongs-to,join:user_id=id"`
	Lists []List `bun:"m2m:list_subscribers,join:Subscriber=List"`
}
