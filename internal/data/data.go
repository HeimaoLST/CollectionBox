package data

import (
	"context"
	"github/heimaolst/collectionbox/internal/biz"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"gorm.io/gorm"
)

type CollectionPO struct {
	ID        string
	CreatedAt time.Time
	URL       string
	Origin    string
}

type sqlRepo struct {
	db *gorm.DB
}

func fromBiz(do *biz.Collection) *CollectionPO {
	return &CollectionPO{
		ID:        do.ID,
		CreatedAt: do.CreatedAt,
		URL:       do.URL,
		Origin:    do.Origin,
	}
}
func (po *CollectionPO) toBiz() *biz.Collection {
	return &biz.Collection{
		ID:        po.ID,
		CreatedAt: po.CreatedAt,
		URL:       po.URL,
		Origin:    po.Origin,
	}
}
func NewSQLRepo(db *gorm.DB) biz.CollectionRepo {
	db.AutoMigrate(&CollectionPO{})
	return &sqlRepo{db: db}
}

func (repo *sqlRepo) CreateCollection(ctx context.Context, c *biz.Collection) error {
	po := fromBiz(c)
	err := repo.db.WithContext(ctx).Create(po).Error
	return err
}

func (repo *sqlRepo) GetByTimeRange(ctx context.Context, start time.Time, end time.Time) ([]*biz.Collection, error) {
	var pos []*CollectionPO

	err := repo.db.WithContext(ctx).
		Where("create_at BETWEEN ? AND ?", start, end).
		Find(&pos).Error
	if err != nil {
		return nil, biz.ErrInternalError.WithMessage(err.Error())
	}
	results := make([]*biz.Collection, 0, len(pos))
	for _, po := range pos {
		results = append(results, po.toBiz())
	}
	return results, nil
}

func (repo *sqlRepo) GetByOrigin(ctx context.Context, origin string) ([]*biz.Collection, error) {
	var pos []*CollectionPO

	err := repo.db.WithContext(ctx).
		Where("origin = ?", origin).
		Find(&pos).Error
	if err != nil {
		return nil, biz.ErrInternalError.WithMessage(err.Error())
	}
	results := make([]*biz.Collection, 0, len(pos))
	for _, po := range pos {
		results = append(results, po.toBiz())
	}
	return results, nil
}
