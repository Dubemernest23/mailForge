package models

import (
	"database/sql"
	"time"

	"github.com/uptrace/bun"
)

type Campaign struct {
	bun.BaseModel `bun:"table:campaigns,alias:c"`

	ID          uint64        `bun:"id,pk,autoincrement"`
	PublicID    string        `bun:"public_id,notnull"`
	UserID      uint64        `bun:"user_id,notnull"`
	ListID      sql.NullInt64 `bun:"list_id"`
	Name        string        `bun:"name,notnull"`
	Subject     string        `bun:"subject,notnull"`
	PreviewText string        `bun:"preview_text"`
	HtmlBody    string        `bun:"html_body"`
	PlainBody   string        `bun:"plain_body"`
	Status      string        `bun:"status,notnull"`
	ScheduledAt sql.NullTime  `bun:"scheduled_at"`
	SentAt      sql.NullTime  `bun:"sent_at"`
	CreatedAt   time.Time     `bun:"created_at,notnull"`
	UpdatedAt   time.Time     `bun:"updated_at,notnull"`

	User *User `bun:"rel:belongs-to,join:user_id=id"`
	List *List `bun:"rel:belongs-to,join:list_id=id"`
}