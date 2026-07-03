package models

import (
	"time"

	"github.com/uptrace/bun"
)

type ListSubscriber struct {
	bun.BaseModel `bun:"table:list_subscriber,alias:ls"`

	ListID       uint64    `bun:"list_id,notnull"`
	SubscriberID uint64    `bun:"subscriber_id,notnull"`
	Status       string    `bun:"status,notnull,default:'subscribed'"`
	CreatedAt    time.Time `bun:"created_at,notnull"`
	UpdatedAt    time.Time `bun:"updated_At,notnull"`

	List       *List       `bun:"rel:belong-to,join:list_id=id"`
	Subscriber *Subscriber `bun:"rel:belong-to,join:subscriber_id=id"`
}
