package models

import (
	"time"

	"github.com/uptrace/bun"
)

type ListSubscriber struct {
	bun.BaseModel `bun:"table:list_subscribers,alias:ls"`

	ListID       uint64    `bun:"list_id,pk"`
	SubscriberID uint64    `bun:"subscriber_id,pk"`
	Status       string    `bun:"status,notnull"`
	CreatedAt    time.Time `bun:"created_at,notnull"`
	UpdatedAt    time.Time `bun:"updated_at,notnull"`

	List       *List       `bun:"rel:belongs-to,join:list_id=id"`
	Subscriber *Subscriber `bun:"rel:belongs-to,join:subscriber_id=id"`
}
