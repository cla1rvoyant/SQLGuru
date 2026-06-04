package service

import (
	"sqlguru/internal/domain"
	"sqlguru/internal/repository"
)

type ArticleService struct {
	articles repository.ArticleRepository
}

func NewArticleService(a repository.ArticleRepository) *ArticleService {
	return &ArticleService{articles: a}
}

func (s *ArticleService) GetArticle(id string) (*domain.Article, error) {
	return s.articles.GetByID(id)
}

func (s *ArticleService) GetAllArticles() ([]domain.Article, error) {
	return s.articles.GetAll()
}
