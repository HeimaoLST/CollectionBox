package biz

import "context"

type OriginExtractor interface {
	Extract(ctx context.Context, rawURL string) (string, error)
}

