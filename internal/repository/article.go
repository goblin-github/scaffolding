package repository

import (
	"context"

	"gorm.io/gorm"

	"scaffolding/internal/model"
)

// ArticleRepository 是项目中唯一允许写 GORM 代码的地方。
type ArticleRepository struct {
	db *gorm.DB
}

func NewArticleRepository(db *gorm.DB) *ArticleRepository {
	return &ArticleRepository{db: db}
}

func (r *ArticleRepository) Create(ctx context.Context, article *model.Article) error {
	return r.db.WithContext(ctx).Create(article).Error
}

func (r *ArticleRepository) FindByID(ctx context.Context, id uint) (*model.Article, error) {
	var article model.Article
	err := r.db.WithContext(ctx).First(&article, id).Error
	if err != nil {
		return nil, err
	}
	return &article, nil
}

func (r *ArticleRepository) List(ctx context.Context, offset, limit int) ([]model.Article, int64, error) {
	var articles []model.Article
	var total int64

	if err := r.db.WithContext(ctx).Model(&model.Article{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := r.db.WithContext(ctx).Offset(offset).Limit(limit).Order("id DESC").Find(&articles).Error
	return articles, total, err
}
