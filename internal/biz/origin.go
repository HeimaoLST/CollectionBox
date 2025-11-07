package biz

import "context"

// URLOriPair represents a single extracted URL and its mapped Origin.
// This is what outer layers (usecase/service) consume to create Collections.
type URLOriPair struct {
	URL    string
	Origin string
}

// OriginExtractor now returns all URL:Origin pairs discovered in input text.
// Implementations should perform lightweight URL normalization and host->origin mapping.
// They SHOULD de-duplicate identical URL+Origin pairs.
type OriginExtractor interface {
	ExtractAll(ctx context.Context, rawText string) ([]URLOriPair, error)
}
