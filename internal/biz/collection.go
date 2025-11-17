package biz

import (
	"context"
	"time"
)

type Collection struct {
	ID        string
	CreatedAt time.Time
	URL       string
	Origin    string
}

type CollectionRepo interface {
	CreateCollection(ctx context.Context, collection *Collection) error
	UpdateCollection(ctx context.Context, collection *Collection) error
	GetByTimeRange(ctx context.Context, start time.Time, end time.Time, origin string) ([]*Collection, error)
	GetByOrigin(ctx context.Context, origin string) ([]*Collection, error)
	GetAllGroupedByOrigin(context.Context) (map[string][]*Collection, error)
}
