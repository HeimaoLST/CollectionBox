package biz

import (
	"context"
	"strings"
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

// CreateCollectionsFromText extracts all URL:Origin pairs from the input text
// and persists each as a Collection. Returns all successfully created Collections.
func (uc *CollectionUsecase) CreateCollectionsFromText(ctx context.Context, text string) ([]*Collection, error) {
	if strings.TrimSpace(text) == "" {
		return nil, ErrInvalidArgument.WithMessage("url cannot be empty")
	}
	if uc == nil || uc.repo == nil || uc.originex == nil {
		return nil, ErrInvalidArgument.WithMessage("repository or origin extractor not configured")
	}

	pairs, err := uc.originex.ExtractAll(ctx, text)
	if err != nil {
		return nil, err
	}
	created := make([]*Collection, 0, len(pairs))
	for _, p := range pairs {
		col := &Collection{
			ID:        uuid.NewString(),
			URL:       p.URL,
			Origin:    p.Origin,
			CreatedAt: time.Now(),
		}
		if err := uc.repo.CreateCollection(ctx, col); err != nil {
			// stop on first persist error and return progress + error
			return created, err
		}
		created = append(created, col)
	}
	return created, nil
}
func (uc *CollectionUsecase) GetByTimeRange(ctx context.Context, start, end time.Time, origin string) ([]*Collection, error) {
	// if days <= 0 || days >= 15 {
	// 	return nil, ErrInvalidArgument.WithMessage("too long ago")
	// }
	// end := time.Now()
	// start := end.AddDate(0, 0, -days)
	if end.After(time.Now()) {
		return nil, ErrInvalidArgument.WithMessage("can't get the url from future")
	}
	if !end.After(start) {
		return nil, ErrInvalidArgument.WithMessage("start time can't be after end time")
	}
	// limit the range to 15 days
	if end.Sub(start) > 15*24*time.Hour {
		return nil, ErrInvalidArgument.WithMessage("time range can't be longer than 15 days")
	}
	cols, err := uc.repo.GetByTimeRange(ctx, start, end, origin)
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
