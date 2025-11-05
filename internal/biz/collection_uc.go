package biz

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type CollectionUsecase struct {
	repo CollectionRepo
}

func NewCollectionUsecase(repo CollectionRepo) *CollectionUsecase {
	return &CollectionUsecase{repo: repo}
}

func (uc *CollectionUsecase) CreateCollection(ctx context.Context, url string, origin string) (*Collection, error) {
	if url == "" {
		return nil, ErrInvaildArgument.WithMessage("url cannot be empty")
	}
	if origin == "" {
		return nil, ErrInvaildArgument.WithMessage("origin cannot be empty")

	}
	col := &Collection{
		ID:        uuid.NewString(),
		URL:       url,
		Origin:    origin,
		CreatedAt: time.Now(),
	}

	if uc == nil || uc.repo == nil {
		return nil, ErrInvaildArgument.WithMessage("repository not configured")
	}

	if err := uc.repo.CreateCollection(ctx, col); err != nil {
		return nil, err
	}

	return col, nil
}
func (uc *CollectionUsecase) GetByTimeRange(ctx context.Context, days int) ([]*Collection, error) {
	if days <= 0 || days >= 15 {
		return nil, ErrInvaildArgument.WithMessage("too long ago")
	}
	end := time.Now()
	start := end.AddDate(0, 0, -days)

	cols, err := uc.repo.GetByTimeRange(ctx, start, end)
	if err != nil {
		return nil, err
	}
	return cols, nil

}
func (uc *CollectionUsecase) GetByOrigin(ctx context.Context, origin string) ([]*Collection, error) {
	if origin == "" {
		return nil, ErrInvaildArgument.WithMessage("the origin you want to search can't be empty")
	}
	cols, err := uc.repo.GetByOrigin(ctx, origin)
	if err != nil {
		return nil, err
	}
	return cols, err
}
