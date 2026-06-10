package service

import (
	"context"

	"scaffolding/internal/errcode"
	"scaffolding/internal/model"
	"scaffolding/internal/repository"
)

// ArticleService 接收 context.Context（不是 *gin.Context），
// 这样同一个 service 可以被 HTTP handler 和 future worker 共用。
type ArticleService struct {
	repo *repository.ArticleRepository
}

func NewArticleService(repo *repository.ArticleRepository) *ArticleService {
	return &ArticleService{repo: repo}
}

func (s *ArticleService) Create(ctx context.Context, title, content string, tags []string) (*model.Article, error) {
	if len(tags) > 10 {
		return nil, errcode.ErrTooManyTags
	}

	article := &model.Article{
		Title:   title,
		Content: content,
		Tags:    tags,
	}
	if err := s.repo.Create(ctx, article); err != nil {
		return nil, errcode.ErrDatabase
	}
	return article, nil
}

func (s *ArticleService) Get(ctx context.Context, id uint) (*model.Article, error) {
	article, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, errcode.ErrNotFound
	}
	return article, nil
}
