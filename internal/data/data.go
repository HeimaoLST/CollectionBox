package data

import (
	"context"
	"time"

	"github/heimaolst/collectionbox/internal/biz"

	_ "github.com/mattn/go-sqlite3"
	"gorm.io/gorm"
)

type CollectionPO struct {
	ID        string
	CreatedAt time.Time
	URL       string `gorm:"uniqueIndex"`
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

func (repo *sqlRepo) UpdateCollection(ctx context.Context, c *biz.Collection) error {
	_, err := gorm.G[CollectionPO](repo.db).Where("url = ?", c.URL).Update(ctx, "create_at", time.Now())

	return err
}

func (repo *sqlRepo) GetByTimeRange(ctx context.Context, start time.Time, end time.Time, origin string) ([]*biz.Collection, error) {
	var pos []*CollectionPO
	var err error
	if origin == "" {
		err = repo.db.WithContext(ctx).
			Where("create_at BETWEEN ? AND ?", start, end).
			Find(&pos).Error
	} else {
		err = repo.db.WithContext(ctx).
			Where("create_at BETWEEN ? AND ?", start, end).
			Where("origin = ?", origin).
			Find(&pos).Error
	}

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

func (repo *sqlRepo) GetAllGroupedByOrigin(ctx context.Context) (map[string][]*biz.Collection, error) {
	var pos []*CollectionPO

	// TODO: I think it will cause OOM in future
	err := repo.db.WithContext(ctx).Order("origin").Find(&pos).Error
	if err != nil {
		return nil, biz.ErrInternalError.WithMessage(err.Error())
	}
	resultMap := make(map[string][]*biz.Collection)

	for _, po := range pos {
		// 4. 将 PO 转换为 DO
		do := po.toBiz() // (假设您有这个转换函数)

		// 5. 按 Origin 存入 map
		resultMap[do.Origin] = append(resultMap[do.Origin], do)
	}

	return resultMap, nil
}
