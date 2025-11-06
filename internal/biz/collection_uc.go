package biz

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type CollectionUsecase struct {
	repo     CollectionRepo
	originex OriginExtractor
}

func NewCollectionUsecase(repo CollectionRepo, ex OriginExtractor) *CollectionUsecase {
	return &CollectionUsecase{repo: repo, originex: ex}
}

func (uc *CollectionUsecase) CreateCollection(ctx context.Context, url string) (*Collection, error) {
	if url == "" {
		return nil, ErrInvalidArgument.WithMessage("url cannot be empty")
	}
	origin, err := uc.originex.Extract(ctx, url)
	if err != nil {
		// 如果 Extract 返回了 biz.ErrInvalidArgument，就直接透传
		return nil, err
	}
	col := &Collection{
		ID:        uuid.NewString(),
		URL:       url,
		Origin:    origin,
		CreatedAt: time.Now(),
	}

	if uc == nil || uc.repo == nil {
		return nil, ErrInvalidArgument.WithMessage("repository not configured")
	}

	if err := uc.repo.CreateCollection(ctx, col); err != nil {
		return nil, err
	}

	return col, nil
}
func (uc *CollectionUsecase) GetByTimeRange(ctx context.Context, days int) ([]*Collection, error) {
	if days <= 0 || days >= 15 {
		return nil, ErrInvalidArgument.WithMessage("too long ago")
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
		return nil, ErrInvalidArgument.WithMessage("the origin you want to search can't be empty")
	}
	cols, err := uc.repo.GetByOrigin(ctx, origin)
	if err != nil {
		return nil, err
	}
	return cols, err
}

func (uc *CollectionUsecase) GetAllGroupedByOrigin(ctx context.Context) (map[string][]*Collection, error) {
	maps, err := uc.repo.GetAllGroupedByOrigin(ctx)
	if err != nil {
		return nil, err
	}
	return maps, err
}
